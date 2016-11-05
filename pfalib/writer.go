package pfalib

import (
	"fmt"
	"io"
	"os"
	"sync"

	capnp "zombiezen.com/go/capnproto2"
)

// DirEntry is used to pass file information into archive writer
type DirEntry struct {
	Path string
	File os.FileInfo
}

// ArchiveWriter is the archive streaming object
type ArchiveWriter struct {
	writer        io.Writer       // stream to write to
	blocksize     int             // reading blocksize
	numreaders    int             // number of parallel readers
	appendchannel chan DirEntry   // channel to send files through for appending
	workgroup     *sync.WaitGroup // waitgroup for readers
	writerlock    *sync.Mutex     // lock to protect writer
	nextid        int64           // next fileid to be written
	idlock        *sync.Mutex     // mutex to protect nextid
}

// NewArchiveWriter creates a new archive object,
// writing to "writer", which can be a file or a size limited
// multifile container or a multistream container
// reading with "blocksize" with "numreaders" reading goroutines
func NewArchiveWriter(writer io.Writer, blocksize int, numreaders int) *ArchiveWriter {
	archivewriter := ArchiveWriter{writer, blocksize, numreaders, make(chan DirEntry, 1), new(sync.WaitGroup), new(sync.Mutex), 0, new(sync.Mutex)}
	for i := 0; i < numreaders; i++ {
		go archivewriter.readWorker()
	}
	return &archivewriter
}

// AppendFile appends a file into the stream
func (w *ArchiveWriter) AppendFile(name DirEntry) {
	w.appendchannel <- name
}

// Close finishes writing to the archive, returning number of written files
func (w *ArchiveWriter) Close() int64 {
	close(w.appendchannel)
	w.workgroup.Wait()
	return w.nextid
}

///// private functions /////////

// readWorker runs in parallel and processes ibnput objects, supports
// files and directories
func (w *ArchiveWriter) readWorker() {
	w.workgroup.Add(1)
	for f := range w.appendchannel {
		if f.File.IsDir() {
			// TODO directory handling
			fmt.Fprint(os.Stderr, "file <", f.Path+"/"+f.File.Name(), "> is of unsupported type.\n")
		} else if f.File.Mode().IsRegular() {
			w.readFile(f)
		} else {
			fmt.Fprint(os.Stderr, "file <", f.Path+"/"+f.File.Name(), "> is of unsupported type.\n")
		}
	}
	w.workgroup.Done()
}

// readFile reads a file and pushes it into archive
func (w *ArchiveWriter) readFile(file DirEntry) {
	buffer := make([]byte, w.blocksize)
	f, err := os.Open(file.Path)
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
				w.writeFileFragment(fileid, buffer[:n])
			}
		} // file read loop
	} else {
		fmt.Fprint(os.Stderr, "could not open file <", file.Path, "> for reading!\n")
	}
	f.Close()
}

// writeFileHeader writes header to archive and returns unique id for the file
func (w *ArchiveWriter) writeFileHeader(file DirEntry) int64 {
	w.idlock.Lock()
	id := w.nextid
	w.nextid++
	w.idlock.Unlock()

	// create header
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		panic(err)
	}

	fh, err := NewRootFileEntry(seg)
	if err != nil {
		panic(err)
	}
	fh.SetName(file.Path /* + "/" + file.File.Name() */)
	fh.SetMode(int64(file.File.Mode().Perm()))

	// TODO fill all headers

	fh.SetSize(file.File.Size())
	fh.SetId(id)

	// write header
	w.writerlock.Lock()
	err = capnp.NewEncoder(w.writer).Encode(msg)
	w.writerlock.Unlock()
	if err != nil {
		panic(err)
	}

	return id
}

// writeFileFragment writes part of a file to archive
func (w *ArchiveWriter) writeFileFragment(fileid int64, buffer []byte) {

	// create header
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		panic(err)
	}

	fs, err := NewRootFileSegment(seg)
	if err != nil {
		panic(err)
	}
	fs.SetId(fileid)
	fs.SetSize(int64(len(buffer)))

	// write header
	w.writerlock.Lock()
	err = capnp.NewEncoder(w.writer).Encode(msg)
	if err != nil {
		panic(err)
	}

	// write data
	w.writer.Write(buffer)
	w.writerlock.Unlock()
}
