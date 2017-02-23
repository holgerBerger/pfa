package main

// FIXME all this is just to create archived so far!!

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/holgerBerger/pfa/pfalib"
)

// LocalProxy is the local endpoint of a proxy to a remote node
type LocalProxy struct {
	node   string
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

// NewLocalProxy creates the local endpoint, and starts the proxy, so it also creates the
// remote end of the proxy
func NewLocalProxy(node string, index int, filename string, blocksize int32, numreaders int, compression pfalib.CompressionType) LocalProxy {
	var err error
	proxy := LocalProxy{node, nil, nil, nil}
	// FIXME hard coded creation!!
	proxy.cmd = exec.Command("/usr/bin/ssh", node, "/tmp/pfa", "-c", "-o", fmt.Sprintf("%s.%d", opts.Output, index), "-b",
		strconv.Itoa(int(opts.Blocksize)), "-r", strconv.Itoa(int(opts.Readers)), "-p", opts.Compression)
	proxy.stdin, err = proxy.cmd.StdinPipe()
	if err != nil {
		panic(err)
	}
	proxy.stdout, err = proxy.cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	proxy.cmd.Start()
	return proxy
}

// AppendFile sends path+name to remote side
func (l LocalProxy) AppendFile(name pfalib.DirEntry) {
	l.stdin.Write([]byte(name.Path + "/" + name.File.Name() + "\n"))
}

// Close closes the connection, and waits for remote side to finish
func (l LocalProxy) Close() (int64, time.Duration, int64, int64) {
	l.stdin.Close()
	l.stdout.Close()
	err := l.cmd.Wait()
	if err != nil {
		panic(err)
	}
	return 0, 0, 0, 0
}

////////////////////////////////

type RemoteProxy struct {
}

func NewRemoteProxy() RemoteProxy {
	proxy := RemoteProxy{}

	// determine compression method
	var compressionmethod pfalib.CompressionType

	compressionmethod = pfalib.NoneC

	if opts.Compression == "snappy" {
		compressionmethod = pfalib.SnappyC
	} else if opts.Compression == "zstd" {
		compressionmethod = pfalib.ZstandardC
	} else if opts.Compression == "none" {
		compressionmethod = pfalib.NoneC
	} else {
		fmt.Fprintln(os.Stderr, "unknown compression method, not compressing.")
	}

	// create outfile
	outfile, err := os.Create(opts.Output)
	if err != nil {
		panic("could not open outfile!")
	}
	boutfile := bufio.NewWriterSize(outfile, int(opts.Blocksize*1024))

	// create archive writer
	archiver := pfalib.NewArchiveWriter(boutfile, opts.Blocksize*1024, opts.Readers, compressionmethod)

	stdin := bufio.NewReader(os.Stdin)

	// append all files
	for {
		filename, err := stdin.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
		}
		tmpstat, err := os.Stat(string(filename))
		f := pfalib.DirEntry{Path: string(filename), File: tmpstat}
		// fmt.Println("adding", f.Path, f.File.Name())
		if f.Path != "" {
			archiver.AppendFile(f)
		} else {
			// ignore markers
		}
	}

	// finalize archive
	//files, timediff, bytes, cbytes := archiver.Close()
	_, _, _, _ = archiver.Close()
	boutfile.Flush()
	outfile.Close()

	return proxy
}
