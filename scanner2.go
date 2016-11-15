package main

import (
	"os"
	"sync"

	"github.com/holgerBerger/pfa/pfalib"
)

type DirNode struct {
	parent   *DirNode
	children []*DirNode
	self     pfalib.DirEntry
	files    []pfalib.DirEntry
}

type Scanner2 struct {
	dirs      []RootT
	workgroup *sync.WaitGroup
}

type RootT struct {
	path string
	root *DirNode
}

func NewScanner2(dirs []string) *Scanner2 {
	scanner := Scanner2{make([]RootT, len(dirs)), new(sync.WaitGroup)}
	for i := range dirs {
		scanner.dirs[i] = RootT{dirs[i], nil}
		scanner.workgroup.Add(1)
		go scanner.worker(&scanner.dirs[i])
	}
	return &scanner
}

func (s *Scanner2) Finish() {
	s.workgroup.Wait()
}

//////////////////////////////////////////////
func (s *Scanner2) worker(dir *RootT) {
	f, err := os.Open(dir.path)
	if err == nil {
		direntries, _ := f.Readdir(0)
		for _, entry := range direntries {
			if entry.IsDir() {
				if dir.root.children == nil {
					dir.root.children = make([]*DirNode, 10)
				}
				this := new(DirNode)
				//this.self = pfalib.DirEntry{entry.Name(), entry}
				dir.root.children = append(dir.root.children, this)
			} else {

			}
		}
	}
	s.workgroup.Done()
}
