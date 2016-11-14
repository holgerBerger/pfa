package pfalib

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc64"
	"os"
	"runtime"
	"sync"

	"github.com/Datadog/zstd"
	"github.com/golang/snappy"
)

// ArchiveReader is the archive reader object
type ArchiveReader struct {
	archives  []*os.File
	waitgroup *sync.WaitGroup
	crctable  *crc64.Table
}

// NewReader creates a archive reader
func NewReader() *ArchiveReader {
	archivereader := ArchiveReader{nil, new(sync.WaitGroup), crc64.MakeTable(crc64.ISO)}
	archivereader.waitgroup.Add(1)
	return &archivereader
}

// AddFile adds a input file to extract from to the reader
func (r *ArchiveReader) AddFile(file *os.File) {
	go r.processFile(file)
	r.archives = append(r.archives, file)
}

// Finish processes all the added input files and extracts the data
func (r *ArchiveReader) Finish() {
	runtime.Gosched()
	r.waitgroup.Done()
	r.waitgroup.Wait()
	for _, f := range r.archives {
		f.Close()
	}
}

//////////// private methods ///////

func (r *ArchiveReader) processFile(reader *os.File) {
	r.waitgroup.Add(1)

	var (
		sectionheader    SectionHeader
		fileheader       FileSection
		filebodyheader   FilebodySection
		filefooterheader FileFooter
		directoryheader  DirectorySection
	)

	var fileworkers sync.WaitGroup
	fileidmap := make(map[uint64]chan []byte)
	crcmap := make(map[uint64]chan uint64)

	for {
		// read section header to determine which header to read next
		err := binary.Read(reader, binary.BigEndian, &sectionheader)
		if err != nil {
			break
		}

		switch sectionheader.Type {

		case uint16(fileE): // FILE --------------------------------------------
			fileheaderbuffer := make([]byte, sectionheader.HeaderSize)
			_, err := reader.Read(fileheaderbuffer)
			if err != nil {
				panic(err)
			}
			err = json.Unmarshal(fileheaderbuffer, &fileheader)
			if err != nil {
				panic(err)
			}
			// fmt.Println("file:", fileheader.File.Dirname, fileheader.FileID)
			// create channel to push data through
			datachan := make(chan []byte)
			fileidmap[fileheader.FileID] = datachan
			crcchan := make(chan uint64)
			crcmap[fileheader.FileID] = crcchan
			// create worker for each file, will get data through channel and channel will
			// get closed when file footer is read
			fileworkers.Add(1)
			go r.fileWorker(fileheader, datachan, &fileworkers, crcchan)

		case uint16(filebodyE): // FILE BODY -----------------------------------
			err := binary.Read(reader, binary.BigEndian, &filebodyheader)
			if err != nil {
				panic(err)
			}
			bodybuffer := make([]byte, filebodyheader.Bodysize)
			_, err = reader.Read(bodybuffer)
			if err != nil {
				panic(err)
			}
			// fmt.Println("bodysegment", filebodyheader.FileID)
			fileidmap[filebodyheader.FileID] <- bodybuffer

		case uint16(filefooterE): // FILE END -----------------------------------
			err := binary.Read(reader, binary.BigEndian, &filefooterheader)
			if err != nil {
				panic(err)
			}
			close(fileidmap[filebodyheader.FileID])
			delete(fileidmap, filebodyheader.FileID)
			crc := <-crcmap[filebodyheader.FileID]
			if crc != filefooterheader.CRC {
				fmt.Fprintln(os.Stderr, "Error: archive CRC mismatch!")
				// TODO add file name here
			}
			delete(crcmap, filebodyheader.FileID)

		case uint16(directoryE): // DIRECTORY -----------------------------------
			dirheaderbuffer := make([]byte, sectionheader.HeaderSize)
			_, err := reader.Read(dirheaderbuffer)
			if err != nil {
				panic(err)
			}
			err = json.Unmarshal(dirheaderbuffer, &directoryheader)
			if err != nil {
				panic(err)
			}
			//list = append(list, FileSection{directoryheader, 0, 0, 0})
			fmt.Println("dir:", directoryheader.Dirname)

		case uint16(softlinkE): // SOFTLINK ---------------------------------------
			fmt.Fprintln(os.Stderr, "softlink not yet supported")
			// FIXME

		default: // ERROR ---------------------------------------------------------
			panic("unexpted type in section header." /* + sectionheader.Type */)

		} // switch
	} // for

	fileworkers.Wait()
	r.waitgroup.Done()

	if len(crcmap) != 0 {
		fmt.Fprintln(os.Stderr, "Error: archive does not close all contained files!")
	}

}

//
func (r *ArchiveReader) fileWorker(file FileSection, datachan chan []byte, fileworker *sync.WaitGroup, crcchan chan uint64) {
	fmt.Println("starting worker", file.FileID, file.File.Dirname)

	// TODO create file

	crc := crc64.New(r.crctable)

	for data := range datachan {
		//fmt.Println("file:", len(data), file.FileID, file.Compression)
		switch file.Compression {
		case uint16(SnappyC):
			buffer, _ := snappy.Decode(nil, data)
			crc.Write(buffer)
			// TODO write to file
		case uint16(ZstandardC):
			buffer, _ := zstd.Decompress(nil, data)
			crc.Write(buffer)
			// TODO write to file
		case uint16(NoneC):
			crc.Write(data)
			// TODO write to file
		default:
			panic("unsupported compression type.")
		}
	}

	// TODO close file

	fmt.Println("ending worker", file.FileID)
	crcchan <- crc.Sum64()
	fileworker.Done()
}

func (r *ArchiveReader) dirWorker() {
	// TODO create directory
}
