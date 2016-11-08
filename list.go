package main

import (
	"fmt"
	"os"

	"github.com/holgerBerger/pfa/pfalib"
)

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
