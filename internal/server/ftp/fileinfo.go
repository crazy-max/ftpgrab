package ftp

import (
	"os"
	"time"
)

type fileInfo struct {
	name  string
	size  int64
	mode  os.FileMode
	mtime time.Time
}

func (f *fileInfo) Name() string {
	return f.name
}

func (f *fileInfo) Size() int64 {
	return f.size
}

func (f *fileInfo) Mode() os.FileMode {
	return f.mode
}

func (f *fileInfo) ModTime() time.Time {
	return f.mtime
}

func (f *fileInfo) IsDir() bool {
	return f.mode.IsDir()
}

func (f *fileInfo) Sys() interface{} {
	return nil
}
