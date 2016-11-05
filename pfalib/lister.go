package pfalib

import (
	"fmt"
	"io"

	capnp "zombiezen.com/go/capnproto2"
)

// List returns list of all files in archive
func List(reader io.Reader) *[]FileEntry {
	list := make([]FileEntry, 0, 1024)

	decoder := capnp.NewDecoder(reader)

	for {
		msg, err := decoder.Decode()
		fmt.Println(err)
		filentry, err := ReadRootFileEntry(msg)
		if err == nil {
			fmt.Println(filentry.Name())
			list = append(list, filentry)
		} else {
			if err == io.EOF {
				break
			}
		}
	}

	return &list
}
