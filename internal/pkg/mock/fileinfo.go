package mock

import (
	"os"
	"time"
)

type FileInfo struct {
	NameVal  string
	IsDirVal bool
}

func (f *FileInfo) Name() string {
	return f.NameVal
}

func (f *FileInfo) Size() int64 {
	panic("implement me")
}

func (f *FileInfo) Mode() os.FileMode {
	panic("implement me")
}

func (f *FileInfo) ModTime() time.Time {
	panic("implement me")
}

func (f *FileInfo) IsDir() bool {
	return f.IsDirVal
}

func (f *FileInfo) Sys() interface{} {
	panic("implement me")
}
