package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/holgerBerger/pfa/pfalib"
	flags "github.com/jessevdk/go-flags"
)

var opts struct {
	Create      bool   `long:"create" short:"c" description:"create archive"`
	List        bool   `long:"list" short:"l" description:"list archive"`
	Extract     bool   `long:"extract" short:"e" description:"extract archive"`
	Scanners    int    `long:"scanners" short:"s" default:"32" description:"number of threads scanning directories"`
	Blocksize   int32  `long:"blocksize" short:"b" default:"1024" description:"blocksize in KiB"`
	Readers     int    `long:"readers" short:"r" default:"32" description:"number of reading threads"`
	Output      string `long:"output" short:"o" description:"file name of output archive in create mode"`
	Input       string `long:"input" short:"i" description:"file name of input archive in list and extract mode"`
	Compression string `long:"compression" short:"p" default:"none" description:"compression, one of <none>, <zstd> or <snappy>"`
}

func main() {
	args, err := flags.Parse(&opts)

	if err != nil {
		//fmt.Println(err)
		os.Exit(1)
	}

	if opts.Create {
		create(args)
	}

	if opts.List {
		list()
	}

}

// create outfile file
func create(args []string) {
	// we scan all files beforehand, to get an idea how big the tree is
	// this wastes some time for large trees, but we scan fast...
	scanner := NewScanner()
	for _, dir := range args {
		scanner.AddDir(dir)
	}

	// start the scanner, this blocks until scanning is done
	scanstart := time.Now()
	scanner.StartScan(opts.Scanners)
	fmt.Printf("scanned %d files in %1.1f seconds, %1.2f files/s.\n",
		len(scanner.Files),
		time.Since(scanstart).Seconds(),
		float64(len(scanner.Files))/time.Since(scanstart).Seconds(),
	)

	// determine compression method
	var compressionmethod pfalib.CompressionType

	compressionmethod = pfalib.NoneC

	if opts.Compression == "snappy" {
		compressionmethod = pfalib.SnappyC
	}
	if opts.Compression == "zstd" {
		compressionmethod = pfalib.ZstandardC
	}

	// create outfile
	outfile, err := os.Create(opts.Output)
	if err != nil {
		panic("could not open outfile!")
	}

	boutfile := bufio.NewWriterSize(outfile, int(opts.Blocksize*1024))

	// create archive write
	archiver := pfalib.NewArchiveWriter(boutfile, opts.Blocksize*1024, opts.Readers, compressionmethod)

	// append all files
	for _, f := range scanner.Files {
		archiver.AppendFile(f)
	}

	// finalize archive
	files, timediff, bytes, cbytes := archiver.Close()
	boutfile.Flush()
	outfile.Close()

	fmt.Printf("written %d files in %1.1f seconds, written with %1.2f MB/s.\n",
		files, timediff.Seconds(), (float64(cbytes)/(timediff.Seconds()))/(1024*1024))
	if compressionmethod != pfalib.NoneC {
		fmt.Printf("%f%% compression.\n", float64(cbytes)/float64(bytes)*100.0)
	}

}

// list input file
func list() {
	infile, err := os.Open(opts.Input)
	if err != nil {
		panic("could not open infile!")
	}

	for _, file := range *pfalib.List(infile) {
		if file.FileID == 0 {
			fmt.Printf("          %s\n", file.File.Dirname)
		} else {
			fmt.Printf("%9d %s\n", file.Filesize, file.File.Dirname)
		}
	}

	infile.Close()
}
