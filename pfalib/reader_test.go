package pfalib

import (
	"fmt"
	"os"
	"testing"
)

func TestReader(t *testing.T) {
	writer, err := os.Create("/tmp/pfa_list_test.pfa")
	if err != nil {
		fmt.Fprint(os.Stderr, "test setup is not working!\n")
		t.Fatal()
	}
	archivewriter := NewArchiveWriter(writer, 128, 8, ZstandardC)

	fileinfo, err := os.Stat("testdata/a")
	if err != nil {
		fmt.Fprint(os.Stderr, "test setup is not working!\n")
		t.Fatal()
	}
	direntry := DirEntry{Path: "testdata", File: fileinfo}
	archivewriter.AppendFile(direntry)

	fileinfo, err = os.Stat("testdata/b")
	if err != nil {
		fmt.Fprint(os.Stderr, "test setup is not working!\n")
		t.Fatal()
	}
	direntry = DirEntry{Path: "testdata", File: fileinfo}
	archivewriter.AppendFile(direntry)

	fileinfo, err = os.Stat("testdata/c")
	if err != nil {
		fmt.Fprint(os.Stderr, "test setup is not working!\n")
		t.Fatal()
	}
	direntry = DirEntry{Path: "testdata", File: fileinfo}
	archivewriter.AppendFile(direntry)

	files, _, _, _ := archivewriter.Close()
	if files != 3 {
		t.Error("unexpected number of files written.")
	}

	infile, err := os.Open("/tmp/pfa_list_test.pfa")
	if err != nil {
		fmt.Fprint(os.Stderr, "archive was not written!\n")
		t.Fatal()
	}

	reader := NewReader()
	reader.AddFile(infile)
	reader.Finish()

}
