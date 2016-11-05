package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/holgerBerger/pfa/pfalib"
)

func TestNew(t *testing.T) {
	writer := bytes.NewBuffer(make([]byte, 0, 1024))
	archivewriter := pfalib.NewArchiveWriter(writer, 128, 8)

	fileinfo, err := os.Stat("testdata/a")
	if err != nil {
		fmt.Fprint(os.Stderr, "test setup is not working!\n")
		t.Fatal()
	}
	direntry := pfalib.DirEntry{Path: "testdata", File: fileinfo}
	archivewriter.AppendFile(direntry)
	archivewriter.AppendFile(direntry)

	fileinfo, err = os.Stat("testdata/b")
	if err != nil {
		fmt.Fprint(os.Stderr, "test setup is not working!\n")
		t.Fatal()
	}
	direntry = pfalib.DirEntry{Path: "testdata", File: fileinfo}
	archivewriter.AppendFile(direntry)
	archivewriter.AppendFile(direntry)

	if archivewriter.Close() != 4 {
		t.Error("unexpected number of files written.")
	}

	if !bytes.Contains(writer.Bytes(), []byte("file A")) {
		t.Error("archive does not contain file a")
	}
	if !bytes.Contains(writer.Bytes(), []byte("File B")) {
		t.Error("archive does not contain file b")
	}

	//fmt.Println(writer)
	reader := bytes.NewReader(writer.Bytes())
	l := *pfalib.List(reader)
	if len(l) != 4 {
		t.Error("wrong number of files read from archive")
	}
}
