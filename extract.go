package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/holgerBerger/pfa/pfalib"
)

// extract input file
func extract() {

	reader := pfalib.NewReader()

	infile, err := os.Open(opts.Input)
	if err == nil {
		reader.AddFile(infile)
	} else {
		files, err := filepath.Glob(opts.Input + ".*")
		if err != nil {
			panic("could not open input file " + opts.Input)
		}
		for _, f := range files {
			infile, err := os.Open(opts.Input)
			if err != nil {
				reader.AddFile(infile)
			} else {
				fmt.Fprintln(os.Stderr, "could not open inputfile", f)
			}
		}
	}

	reader.Finish()
}
