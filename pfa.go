package main

import (
	"fmt"
	"os"

	flags "github.com/jessevdk/go-flags"
)

var opts struct {
	Create      bool   `long:"create" short:"c" description:"create archive"`
	List        bool   `long:"list" short:"l" description:"list archive"`
	Extract     bool   `long:"extract" short:"e" description:"extract archive"`
	Scanners    int    `long:"scanners" short:"s" default:"32" description:"number of threads scanning directories"`
	Blocksize   int32  `long:"blocksize" short:"b" default:"1024" description:"blocksize in KiB"`
	Readers     int    `long:"readers" short:"r" default:"32" description:"number of reading threads"`
	Files       int    `long:"files" short:"f" default:"1" description:"number of output files"`
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
		if len(opts.Output) == 0 {
			fmt.Fprintln(os.Stderr, "create mode requires output file!")
			os.Exit(1)
		}
		if opts.Files > 1 {
			createMultiple2(args, opts.Files)
		} else {
			create(args)
		}
	} else if opts.Extract {
		if len(opts.Input) == 0 {
			fmt.Fprintln(os.Stderr, "extract mode requires inut file!")
			os.Exit(1)
		}
		// TODO
	} else if opts.List {
		list()
	} else {
		fmt.Fprintln(os.Stderr, "create, extract or list has to be chosen.")
	}

}
