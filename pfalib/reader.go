package pfalib

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// ArchiveReader is the archive reader object
type ArchiveReader struct {
	archives  []*os.File
	waitgroup *sync.WaitGroup
}

// NewReader creates a archive reader
func NewReader() *ArchiveReader {
	archivereader := ArchiveReader{nil, new(sync.WaitGroup)}
	return &archivereader
}

// AddFile adds a input file to extract from to the reader
func (r *ArchiveReader) AddFile(file *os.File) {
	go r.processFile(file)
	r.archives = append(r.archives, file)
}

// Finish processes all the added input files and extracts the data
func (r *ArchiveReader) Finish() {
	time.Sleep(100 * time.Millisecond)
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

	for {
		// read section header to determine which header to read next
		err := binary.Read(reader, binary.BigEndian, &sectionheader)
		if err != nil {
			break
		}

		switch sectionheader.Type {
		// file
		case uint16(fileE):
			fileheaderbuffer := make([]byte, sectionheader.HeaderSize)
			_, err := reader.Read(fileheaderbuffer)
			if err != nil {
				panic(err)
			}
			err = json.Unmarshal(fileheaderbuffer, &fileheader)
			if err != nil {
				panic(err)
			}
			fmt.Println("file:", fileheader.File.Dirname)
			//list = append(list, fileheader)

			// file body
		case uint16(filebodyE):
			err := binary.Read(reader, binary.BigEndian, &filebodyheader)
			if err != nil {
				panic(err)
			}
			bodybuffer := make([]byte, filebodyheader.Bodysize)
			_, err = reader.Read(bodybuffer)
			if err != nil {
				panic(err)
			}

			// file end
		case uint16(filefooterE):
			err := binary.Read(reader, binary.BigEndian, &filefooterheader)
			if err != nil {
				panic(err)
			}

			// directory
		case uint16(directoryE):
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

			// softlink
		case uint16(softlinkE):
			fmt.Fprintln(os.Stderr, "softlink not yet supported")
			// FIXME

		default:
			panic("unexpted type in section header." /* + sectionheader.Type */)

		} // switch
	} // for

	r.waitgroup.Done()
}
