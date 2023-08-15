package mock

import (
	"io"
	"os"
	"time"

	pio "github.com/confluentinc/cli/v3/pkg/io"
)

// PassThroughFileSystem is useful for optionally mocking some methods
// We have to check whether Mock.<Name>Func is nil because our mocks panic-recovery if called with nil func
type PassThroughFileSystem struct {
	Mock *FileSystem
	FS   pio.FileSystem
}

var _ pio.FileSystem = (*PassThroughFileSystem)(nil)

func (c *PassThroughFileSystem) Open(name string) (pio.File, error) {
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

func (c *PassThroughFileSystem) Create(name string) (pio.File, error) {
	if c.Mock.CreateFunc != nil {
		return c.Mock.Create(name)
	}
	return c.FS.Create(name)
}

func (c *PassThroughFileSystem) Chtimes(n string, a, m time.Time) error {
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

func (c *PassThroughFileSystem) ReadDir(dirname string) ([]os.DirEntry, error) {
	if c.Mock.ReadDirFunc != nil {
		return c.Mock.ReadDirFunc(dirname)
	}
	return c.FS.ReadDir(dirname)
}

func (c *PassThroughFileSystem) MkdirTemp(dir, prefix string) (string, error) {
	if c.Mock.MkdirTempFunc != nil {
		return c.Mock.MkdirTemp(dir, prefix)
	}
	return c.FS.MkdirTemp(dir, prefix)
}

func (c *PassThroughFileSystem) Copy(dst io.Writer, src io.Reader) (int64, error) {
	if c.Mock.CopyFunc != nil {
		return c.Mock.Copy(dst, src)
	}
	return c.FS.Copy(dst, src)
}

func (c *PassThroughFileSystem) Move(src, dst string) error {
	if c.Mock.MoveFunc != nil {
		return c.Mock.Move(src, dst)
	}
	return c.FS.Move(src, dst)
}

func (c *PassThroughFileSystem) NewBufferedReader(rd io.Reader) pio.Reader {
	if c.Mock.NewBufferedReaderFunc != nil {
		return c.Mock.NewBufferedReader(rd)
	}
	return c.FS.NewBufferedReader(rd)
}

func (c *PassThroughFileSystem) IsTerminal(fd uintptr) bool {
	if c.Mock.IsTerminalFunc != nil {
		return c.Mock.IsTerminal(fd)
	}
	return c.FS.IsTerminal(fd)
}

func (c *PassThroughFileSystem) Glob(pattern string) ([]string, error) {
	if c.Mock.GlobFunc != nil {
		return c.Mock.Glob(pattern)
	}
	return c.FS.Glob(pattern)
}
