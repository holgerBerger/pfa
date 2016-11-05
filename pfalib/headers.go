package pfalib

// ArchiveHeader is at begin of file to identify
type ArchiveHeader struct {
	Magic   uint64 // PFA1PFA1
	Version uint16
	Ctime   uint64 // epoch of file creation
}

// SectionHeader is at start of section and identifies following header
type SectionHeader struct {
	Magic      uint32 // PFA1
	Type       uint16 // type of section
	HeaderSize uint16 // size of following JSON header
}

type sectionType uint16

const (
	fileE sectionType = iota
	directoryE
	softlinkE
	filebodyE
	filefooterE
)

type compressionType uint16

const (
	none compressionType = iota
	zstandard
	zlib
	snappy
	lzo
)

// DirectorySection represents a directory
type DirectorySection struct {
	Dirname string // filename in UTF-8
	UID     uint32 // owners uid
	GID     uint32 // owners gid
	Owner   string // username
	Group   string // groupname
	Mtime   uint64 // timestamp modify
	Ctime   uint64 // timestamp creation
	Atime   uint64 // timestamp access
	Mode    uint64 // file permissions
}

// FileSection is a file header
type FileSection struct {
	File        DirectorySection
	Filesize    uint64 // size in bytes FIXME needed? could be used for preallocation in extraction
	FileID      uint64 // unique ID for file, used in body
	Compression uint16 // type of compression
}

// FilebodySection is a part of a file
type FilebodySection struct {
	FileID   uint64 // unique id of this file within this stream
	Bodysize uint64 // size of the following payload
}

// FileFooter marks end of a file
type FileFooter struct {
	FileIO uint64
	CRC    uint64
}

// SoftLinkSection represents a softline
type SoftLinkSection struct {
	File       DirectorySection
	Targetname string
}
