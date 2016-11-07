package main

/*

	scan directory tree and create list of all files
	to be archived

*/

import (
	"os"
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
	s.scangroup.Add(1)
	s.scannerChannel <- dir
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

			s.FileMutex.Lock()
			for _, i := range direntries {
				s.Files = append(s.Files, pfalib.DirEntry{Path: dir, File: i})
			}
			s.FileMutex.Unlock()

			for _, entry := range direntries {
				if entry.IsDir() {
					s.scangroup.Add(1)
					// this could block
					go func(name string) {
						s.scannerChannel <- name
					}(dir + "/" + entry.Name())
				} else {
					totalsize += entry.Size()
				}
			}

			f.Close()
		}
		s.scangroup.Done()
	}
	s.sizeChannel <- totalsize
}
