package pfalib

import "time"

type ArchiveWriterInterface interface {
	AppendFile(name DirEntry)
	Close() (int64, time.Duration, int64, int64)
}
