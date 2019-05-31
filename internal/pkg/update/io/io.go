//go:generate mocker --prefix "" --dst ../mock/filesystem.go --pkg mock --selfpkg github.com/confluentinc/cli io.go FileSystem
package io

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/mattn/go-isatty"
)

// FileSystem interface wraps IO so that we can mock it in our unit tests
type FileSystem interface {
	// os
	Open(name string) (File, error)
	Stat(name string) (os.FileInfo, error)
	Create(name string) (File, error)
	Chtimes(name string, atime time.Time, mtime time.Time) error
	Chmod(name string, mode os.FileMode) error
	Remove(name string) error
	RemoveAll(path string) error
	// ioutil
	TempDir(dir, prefix string) (name string, err error)
	// io
	Copy(dst io.Writer, src io.Reader) (written int64, err error)
	Move(src string, dst string) error
	// bufio
	NewBufferedReader(rd io.Reader) Reader
	// isatty
	IsTerminal(fd uintptr) bool
}

// File interface is used by FileSystem interface to enable mocking in unit tests
type File interface {
	io.Closer
	io.Reader
	io.ReaderAt
	io.Writer
	io.WriterAt
	io.Seeker
	Stat() (os.FileInfo, error)
	Fd() uintptr
}

// Reader reads buffered strings
type Reader interface {
	ReadString(delim byte) (string, error)
}

// RealFileSystem implements fileSystem using the local disk.
type RealFileSystem struct{}

var _ FileSystem = (*RealFileSystem)(nil)

func (*RealFileSystem) Open(name string) (File, error)                   { return os.Open(name) }
func (*RealFileSystem) Stat(name string) (os.FileInfo, error)            { return os.Stat(name) }
func (*RealFileSystem) Create(name string) (File, error)                 { return os.Create(name) }
func (*RealFileSystem) Chtimes(n string, a time.Time, m time.Time) error { return os.Chtimes(n, a, m) }
func (*RealFileSystem) Chmod(name string, mode os.FileMode) error        { return os.Chmod(name, mode) }
func (*RealFileSystem) Remove(name string) error                         { return os.Remove(name) }
func (*RealFileSystem) RemoveAll(path string) error                      { return os.RemoveAll(path) }
func (*RealFileSystem) TempDir(dir, prefix string) (string, error)       { return ioutil.TempDir(dir, prefix) }
func (*RealFileSystem) Copy(dst io.Writer, src io.Reader) (int64, error) { return io.Copy(dst, src) }
func (*RealFileSystem) Move(src string, dst string) error                { return os.Rename(src, dst) }
func (*RealFileSystem) NewBufferedReader(rd io.Reader) Reader            { return bufio.NewReader(rd) }
func (*RealFileSystem) IsTerminal(fd uintptr) bool                       { return isatty.IsTerminal(fd) }
