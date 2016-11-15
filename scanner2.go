package main

import "github.com/holgerBerger/pfa/pfalib"

type DirNode struct {
	parent   *DirNode
	children []*DirNode
	self     pfalib.DirEntry
	files    pfalib.DirEntry
}
