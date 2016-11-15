package main

/*

	scan directory tree and create list of all files
	to be archived

*/

import (
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/holgerBerger/pfa/pfalib"
)

// Scanner scans all directories and builds a tree
type Scanner struct {
	scangroup      sync.WaitGroup
	scannerChannel chan string
	sizeChannel    chan int64
	TotalSize      int64
	FileMutex      sync.RWMutex
	Files          []pfalib.DirEntry
}

// NewScanner creates a scanner, one scanner runs several go-routines
func NewScanner() *Scanner {
	var scanner Scanner
	scanner.scannerChannel = make(chan string, 10) // FIXME why does 1 not work??
	scanner.sizeChannel = make(chan int64, 1)
	scanner.Files = make([]pfalib.DirEntry, 0, 1000)
	return &scanner
}

// AddDir adds a directory to be scanned
func (s *Scanner) AddDir(dir string) {
	cleaned := path.Clean(dir)

	// exclude directories starting with ..
	if len(cleaned) >= 2 && cleaned[:2] == ".." {
		fmt.Fprintln(os.Stderr, "ommited ", dir)
		return
	}

	/* FIXME

	// we try to avoid adding output file into output (=recursion),
	// by checking if output file might be below input directory
	if path.IsAbs(cleaned) {
		cwd, _ := os.Getwd()
		if len(cleaned) <= len(cwd) && cwd[:len(cleaned)] == cleaned {
			fmt.Fprintln(os.Stderr, "ommited ", dir)
			return
		}
	}

	*/

	s.scangroup.Add(1)
	s.scannerChannel <- cleaned

}

// StartScan starts nr go-routines, and scans all directories added using AddDir before
func (s *Scanner) StartScan(nr int) {
	for i := 0; i < nr; i++ {
		go s.Scanner()
	}

	// wait until all work is done
	s.scangroup.Wait()
	// then terminate workers
	close(s.scannerChannel)

	for i := 0; i < nr; i++ {
		s.TotalSize += <-s.sizeChannel
	}
}

// Scanner is the worker go-routine to do the work
func (s *Scanner) Scanner() {
	var totalsize int64
	for dir := range s.scannerChannel {
		f, err := os.Open(dir)
		if err == nil {
			direntries, _ := f.Readdir(0)

			for _, entry := range direntries {
				if entry.IsDir() {
					//s.scangroup.Add(1)
					s.FileMutex.Lock()
					s.Files = append(s.Files, pfalib.DirEntry{Path: dir, File: entry})
					s.FileMutex.Unlock()
					// this could block
					/*
						go func(name string) {
							s.scannerChannel <- name
						}(path.Join(dir, entry.Name()))
					*/
					//s.scannerChannel <- path.Join(dir, entry.Name())
				} else {
					totalsize += entry.Size()
				}
			}

			// append after handling directories, to make sure directories are created first
			s.FileMutex.Lock()
			for _, entry := range direntries {
				if !entry.IsDir() {
					s.Files = append(s.Files, pfalib.DirEntry{Path: dir, File: entry})
				} else {
					// this could block,
					// moved here, so we descend after local files
					s.scangroup.Add(1)
					go func(name string) {
						s.scannerChannel <- name
					}(path.Join(dir, entry.Name()))
				}
			}
			s.FileMutex.Unlock()

			f.Close()
		}
		s.scangroup.Done()
	}
	s.sizeChannel <- totalsize
}
