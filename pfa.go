package main

import (
	"os"

	"github.com/holgerBerger/pfa/pfalib"
	flags "github.com/jessevdk/go-flags"
)

var opts struct {
	Create      bool   `long:"create" short:"c" description:"create archive"`
	List        bool   `long:"list" short:"l" description:"list archive"`
	Scanners    int    `long:"scanners" short:"s" default:"8" description:"number of threads scanning directories"`
	Blocksize   int32  `long:"blocksize" short:"b" default:"1024" description:"blocksize in KiB"`
	Writers     int    `long:"writers" short:"w" default:"4" description:"number of writing processes, equals streams in archive."`
	Readers     int    `long:"readers" short:"r" default:"8" description:"number of reading threads"`
	Archivesize int64  `long:"filesize" short:"f" default:"16" description:"filesize in GiB of one archive file"`
	Output      string `long:"output" short:"o" description:"file name of output archive in create mode"`
}

func main() {
	args, err := flags.Parse(&opts)

	if err != nil {
		//fmt.Println(err)
		os.Exit(1)
	}

	if opts.Create {

		// we scan all files beforehand, to get an idea how big the tree is
		// this wastes some time for large trees, but we scan fast...
		scanner := NewScanner()
		for _, dir := range args {
			scanner.AddDir(dir)
		}

		// start the scanner, this blocks until scanning is done
		scanner.StartScan(opts.Scanners)

		outfile, err := os.Create(opts.Output)
		if err != nil {
			panic("could not open outfile!")
		}

		// create archive write
		archiver := pfalib.NewArchiveWriter(outfile, opts.Blocksize, opts.Readers)

		// append all files
		for _, f := range scanner.Files {
			archiver.AppendFile(f)
		}

		// finalize archive
		archiver.Close()

	}
}
