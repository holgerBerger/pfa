package pfalib

// FileHeader is at begin of file to identify
type FileHeader struct {
	Magic   int64 // PFA1PFA1
	Version int16
	Ctime   int64 // epoch of file creation
}

// SectionHeader is at start of section and identifies following header
type SectionHeader struct {
	Magic      int32 // PFA1
	Type       int16 // type of section
	HeaderSize int32 // size of following JSON header
}

type sectionType uint16

const (
	file sectionType = iota
	directory
	softlink
	filebody
)

// DirectorySection
type DirectorySection struct {
	Dirname string // filename in UTF-8
	Uid     uint32 // owners uid
	Gid     uint32 // owners gid
	Owner   string // username
	Group   string // groupname
	Mtime   uint64 // timestamp modify
	Ctime   uint64 // timestamp creation
	Atime   uint64 // timestamp access
	Mode    uint64 // file permissions
}

type FileSection struct {
	File     DirectorySection
	Filesize uint64 // size in bytes FIXME needed? could be used for preallocation in extraction
	FileID   uint64 // unique ID for file, used in body
}

type FilebodySection struct {
	FileID   uint64 // unique id of this file within this stream
	Bodysize uint64 // size of the following payload
}

// SoftLinkSection represents a softline
type SoftLinkSection struct {
	File       DirectorySection
	Targetname string
}
