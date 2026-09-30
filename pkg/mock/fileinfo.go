package mock

import (
	"io/fs"
	"os"
	"time"
)

type FileInfo struct {
	NameVal string
	ModeVal os.FileMode
}

func (f *FileInfo) Type() fs.FileMode {
	panic("implement me")
}

func (f *FileInfo) Info() (fs.FileInfo, error) {
	return &FileInfo{NameVal: f.NameVal, ModeVal: f.ModeVal}, nil
}

func (f *FileInfo) Name() string {
	return f.NameVal
}

func (f *FileInfo) Size() int64 {
	panic("implement me")
}

func (f *FileInfo) Mode() os.FileMode {
	return f.ModeVal
}

func (f *FileInfo) ModTime() time.Time {
	panic("implement me")
}

func (f *FileInfo) IsDir() bool {
	panic("implement me")
}

func (f *FileInfo) Sys() any {
	panic("implement me")
}
