package mock

import (
	"io"
	"os"
	"time"

	io2 "github.com/confluentinc/cli/internal/pkg/update/io"
)

// PassThroughFileSystem is useful for optionally mocking some methods
// We have to check whether Mock.<Name>Func is nil because our mocks panic if called with nil func
type PassThroughFileSystem struct {
	Mock *FileSystem
	FS   io2.FileSystem
}

func (c *PassThroughFileSystem) Open(name string) (io2.File, error) {
	if c.Mock.OpenFunc != nil {
		return c.Mock.Open(name)
	}
	return c.FS.Open(name)
}

func (c *PassThroughFileSystem) Stat(name string) (os.FileInfo, error) {
	if c.Mock.StatFunc != nil {
		return c.Mock.Stat(name)
	}
	return c.FS.Stat(name)
}

func (c *PassThroughFileSystem) Create(name string) (io2.File, error) {
	if c.Mock.CreateFunc != nil {
		return c.Mock.Create(name)
	}
	return c.FS.Create(name)
}

func (c *PassThroughFileSystem) Chtimes(n string, a time.Time, m time.Time) error {
	if c.Mock.ChtimesFunc != nil {
		return c.Mock.Chtimes(n, a, m)
	}
	return c.FS.Chtimes(n, a, m)
}

func (c *PassThroughFileSystem) Chmod(name string, mode os.FileMode) error {
	if c.Mock.ChmodFunc != nil {
		return c.Mock.Chmod(name, mode)
	}
	return c.FS.Chmod(name, mode)
}

func (c *PassThroughFileSystem) Remove(name string) error {
	if c.Mock.RemoveFunc != nil {
		return c.Mock.Remove(name)
	}
	return c.FS.Remove(name)
}

func (c *PassThroughFileSystem) RemoveAll(path string) error {
	if c.Mock.RemoveAllFunc != nil {
		return c.Mock.RemoveAllFunc(path)
	}
	return c.FS.RemoveAll(path)
}

func (c *PassThroughFileSystem) TempDir(dir, prefix string) (string, error) {
	if c.Mock.TempDirFunc != nil {
		return c.Mock.TempDir(dir, prefix)
	}
	return c.FS.TempDir(dir, prefix)
}

func (c *PassThroughFileSystem) Copy(dst io.Writer, src io.Reader) (int64, error) {
	if c.Mock.CopyFunc != nil {
		return c.Mock.Copy(dst, src)
	}
	return c.FS.Copy(dst, src)
}
