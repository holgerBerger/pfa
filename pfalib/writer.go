package pfalib

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc64"
	"io"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/Datadog/zstd"
	"github.com/golang/snappy"
)

// DirEntry is used to pass file information into archive writer
type DirEntry struct {
	Path string
	File os.FileInfo
}

// ArchiveWriter is the archive streaming object
type ArchiveWriter struct {
	writer        io.Writer           // stream to write to
	blocksize     int32               // reading blocksize
	numreaders    int                 // number of parallel readers
	appendchannel chan DirEntry       // channel to send files through for appending
	workgroup     *sync.WaitGroup     // waitgroup for readers
	writerlock    *sync.Mutex         // lock to protect writer
	nextid        int64               // next fileid to be written
	idlock        *sync.Mutex         // mutex to protect nextid
	starttime     time.Time           // time of creation of writer
	byteswritten  int64               // bytes written to file
	cbyteswritten int64               // bytes written after compression
	compression   CompressionType     // type of compression
	crctable      *crc64.Table        // crc polynomial
	dircache      map[string]DirEntry // rememer directories already created
	dircachelock  *sync.RWMutex       // lock to protect dircache
}

// NewArchiveWriter creates a new archive object,
// writing to "writer", which can be a file or a size limited
// multifile container or a multistream container
// reading with "blocksize" with "numreaders" reading goroutines
func NewArchiveWriter(writer io.Writer, blocksize int32, numreaders int, compression CompressionType) *ArchiveWriter {
	archivewriter := ArchiveWriter{writer, blocksize, numreaders, make(chan DirEntry, 1), new(sync.WaitGroup),
		new(sync.Mutex), 1, new(sync.Mutex), time.Now(), 0, 0, compression, nil, make(map[string]DirEntry), new(sync.RWMutex)}
	for i := 0; i < numreaders; i++ {
		go archivewriter.readWorker()
	}
	archivewriter.crctable = crc64.MakeTable(crc64.ISO) // ise ISO polynomial
	return &archivewriter
}

// AppendFile appends a file into the stream
func (w *ArchiveWriter) AppendFile(name DirEntry) {
	w.appendchannel <- name
}

// Close finishes writing to the archive, returning number of written files
func (w *ArchiveWriter) Close() (int64, time.Duration, int64, int64) {
	close(w.appendchannel)
	w.workgroup.Wait()
	return w.nextid, time.Since(w.starttime), w.byteswritten, w.cbyteswritten
}

/************* private functions **************/

// checkPath checks recursive if path was already created
func (w *ArchiveWriter) checkPath(f string) {
	if f == "/" {
		return
	}
	w.checkPath(filepath.Dir(f))
	fmt.Println("checkPath<", f)
	w.dircachelock.RLock()
	_, ok := w.dircache[f]
	w.dircachelock.RUnlock()
	if !ok {
		fmt.Println("!! no idea about:", f)
	}
}

// readWorker runs in parallel and processes input objects, supports
// files and directories
func (w *ArchiveWriter) readWorker() {
	w.workgroup.Add(1)
	for f := range w.appendchannel {
		if f.File.IsDir() {
			w.dircachelock.RLock()
			_, ok := w.dircache[f.Path]
			w.dircachelock.RUnlock()
			if !ok {
				w.dircachelock.Lock()
				w.dircache[f.Path] = f
				w.dircachelock.Unlock()
			}
			w.readDir(f)
		} else if f.File.Mode().IsRegular() {
			w.dircachelock.RLock()
			_, ok := w.dircache[f.Path]
			w.dircachelock.RUnlock()
			if !ok {
				fmt.Println("path not yet made", f.Path, f.File.Name())
				w.checkPath(f.Path)
			}
			w.readFile(f)
		} else {
			fmt.Fprint(os.Stderr, "file <", path.Join(f.Path, f.File.Name()), "> is of unsupported type.\n")
		}
	}
	w.workgroup.Done()
}

// readDir adds a directory to archive
func (w *ArchiveWriter) readDir(file DirEntry) {
	w.writeDirHeader(file)
}

// readFile reads a file and pushes it into archive
func (w *ArchiveWriter) readFile(file DirEntry) {
	buffer := make([]byte, w.blocksize)

	crc := crc64.New(w.crctable)

	f, err := os.Open(path.Join(file.Path, file.File.Name()))
	if err == nil {
		fileid := w.writeFileHeader(file)
		// read blocks and stream them into file
		for {
			n, err := f.Read(buffer)
			if err == io.EOF {
				break
			}
			if err != nil {
				panic(err) // bail out, we do not expect this
			}
			//fmt.Println("write fragment of", name, n, len(buffer), id)
			if n > 0 {
				crc.Write(buffer[:n])
				w.writeFileFragment(fileid, buffer[:n])
			}
		} // file read loop
		w.writeFileFooter(fileid, crc.Sum64())
		f.Close()
	} else {
		fmt.Fprint(os.Stderr, "could not open file <", file.Path, "> for reading!\n")
	}
}

func (w *ArchiveWriter) writeDirHeader(file DirEntry) {
	fmt.Println("writing dir header ", file.File.Name())
	fh, err := json.Marshal(DirectorySection{
		path.Join(file.Path, file.File.Name()),
		0, 0, "", "", 0, 0, 0,
		uint64(file.File.Mode().Perm()),
	})
	if err != nil {
		panic(err)
	}

	// write header
	w.writerlock.Lock()
	binary.Write(w.writer, binary.BigEndian, SectionHeader{uint32(0x46503141), uint16(directoryE), uint16(len(fh))})
	w.writer.Write(fh)
	w.writerlock.Unlock()
}

// writeFileHeader writes header to archive and returns unique id for the file
func (w *ArchiveWriter) writeFileHeader(file DirEntry) int64 {
	fmt.Println("writing file header ", file.File.Name())
	w.idlock.Lock()
	id := w.nextid
	w.nextid++
	w.byteswritten += file.File.Size()
	w.idlock.Unlock()

	fh, err := json.Marshal(FileSection{
		DirectorySection{
			path.Join(file.Path, file.File.Name()),
			0, 0, "", "", 0, 0, 0,
			uint64(file.File.Mode().Perm())},
		uint64(file.File.Size()),
		uint64(id),
		uint16(w.compression),
	})
	if err != nil {
		panic(err)
	}

	// write header
	w.writerlock.Lock()
	binary.Write(w.writer, binary.BigEndian, SectionHeader{uint32(0x46503141), uint16(fileE), uint16(len(fh))})
	w.writer.Write(fh)
	w.writerlock.Unlock()

	return id
}

// writeFileFooter writes footer at file end
func (w *ArchiveWriter) writeFileFooter(fileid int64, crc uint64) {
	w.writerlock.Lock()

	// write header
	binary.Write(w.writer, binary.BigEndian, SectionHeader{uint32(0x46503141), uint16(filefooterE), uint16(0)})
	binary.Write(w.writer, binary.BigEndian, FileFooter{uint64(fileid), crc})

	w.writerlock.Unlock()
}

// writeFileFragment writes part of a file to archive
func (w *ArchiveWriter) writeFileFragment(fileid int64, buffer []byte) {

	switch w.compression {
	case SnappyC:
		cbuffer := make([]byte, 2*w.blocksize)
		cbuffer = snappy.Encode(cbuffer, buffer)
		w.writerlock.Lock()
		// write header
		binary.Write(w.writer, binary.BigEndian, SectionHeader{uint32(0x46503141), uint16(filebodyE), uint16(0)})
		binary.Write(w.writer, binary.BigEndian, FilebodySection{uint64(fileid), uint64(len(cbuffer))})
		// write data
		w.cbyteswritten += int64(len(cbuffer))
		w.writer.Write(cbuffer)
	case ZstandardC:
		cbuffer := make([]byte, 2*w.blocksize)
		cbuffer, _ = zstd.Compress(cbuffer, buffer)
		w.writerlock.Lock()
		// write header
		binary.Write(w.writer, binary.BigEndian, SectionHeader{uint32(0x46503141), uint16(filebodyE), uint16(0)})
		binary.Write(w.writer, binary.BigEndian, FilebodySection{uint64(fileid), uint64(len(cbuffer))})
		// write data
		w.cbyteswritten += int64(len(cbuffer))
		w.writer.Write(cbuffer)
	case NoneC:
		w.writerlock.Lock()
		// write header
		binary.Write(w.writer, binary.BigEndian, SectionHeader{uint32(0x46503141), uint16(filebodyE), uint16(0)})
		binary.Write(w.writer, binary.BigEndian, FilebodySection{uint64(fileid), uint64(len(buffer))})
		// write data
		w.cbyteswritten += int64(len(buffer))
		w.writer.Write(buffer)
	default:
		panic("archive writer called with unsupported compression type.")
	}

	w.writerlock.Unlock()
}
