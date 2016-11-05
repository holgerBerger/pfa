package pfalib

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// List returns list of all files in archive
func List(reader io.Reader) *[]FileSection {
	list := make([]FileSection, 0, 1024)

	var (
		sectionheader    SectionHeader
		fileheader       FileSection
		filebodyheader   FilebodySection
		filefooterheader FileFooter
	)

	for {

		// read section header to determine which header to read next
		err := binary.Read(reader, binary.BigEndian, &sectionheader)
		if err != nil {
			break
		}

		// file
		if sectionheader.Type == int16(fileE) {
			fileheaderbuffer := make([]byte, sectionheader.HeaderSize)
			_, err := reader.Read(fileheaderbuffer)
			if err != nil {
				panic(err)
			}
			err = json.Unmarshal(fileheaderbuffer, &fileheader)
			if err != nil {
				panic(err)
			}
			list = append(list, fileheader)
		}

		// file body
		if sectionheader.Type == int16(filebodyE) {
			err := binary.Read(reader, binary.BigEndian, &filebodyheader)
			if err != nil {
				panic(err)
			}
			bodybuffer := make([]byte, filebodyheader.Bodysize)
			_, err = reader.Read(bodybuffer)
			if err != nil {
				panic(err)
			}
		}

		// file end
		if sectionheader.Type == int16(filefooterE) {
			err := binary.Read(reader, binary.BigEndian, &filefooterheader)
			if err != nil {
				panic(err)
			}
		}

		// directory
		if sectionheader.Type == int16(directoryE) {
			fmt.Fprintln(os.Stderr, "directory not yet supported")
			// FIXME
		}

		// softlink
		if sectionheader.Type == int16(softlinkE) {
			fmt.Fprintln(os.Stderr, "softlink not yet supported")
			// FIXME
		}

	}

	return &list
}
