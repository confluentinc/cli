package mock

import (
	"fmt"
	"os"
	"path/filepath"
)

var count = 0

type ConfluentPlatform struct {
	ConfluentHome    string
	ConfluentCurrent string
}

func NewConfluentPlatform() *ConfluentPlatform {
	return new(ConfluentPlatform)
}

func (cp *ConfluentPlatform) NewConfluentHome() error {
	dir, err := newDir()
	cp.ConfluentHome = dir
	if err != nil {
		return err
	}

	return os.Setenv("CONFLUENT_HOME", dir)
}

func (cp *ConfluentPlatform) NewConfluentCurrent() error {
	dir, err := newDir()
	cp.ConfluentCurrent = dir
	if err != nil {
		return err
	}

	return os.Setenv("CONFLUENT_CURRENT", dir)
}

func newDir() (string, error) {
	path := filepath.Join(os.TempDir(), fmt.Sprintf("confluent.test.%06d", count))
	count++

	return path, os.Mkdir(path, 0777)
}

func (cp *ConfluentPlatform) AddFileToConfluentHome(file string) error {
	path := filepath.Join(cp.ConfluentHome, file)

	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return err
	}
	if _, err := os.Create(path); err != nil {
		return err
	}

	return nil
}

func (cp *ConfluentPlatform) TearDown() {
	os.RemoveAll(cp.ConfluentHome)
	os.RemoveAll(cp.ConfluentCurrent)
}
