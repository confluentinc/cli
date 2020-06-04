package mock

import (
	"fmt"
	"os"
	"path/filepath"
)

// Ensure that each CONFLUENT_HOME directory has a unique name to satisfy windows tests
var count = 0

type ConfluentHome struct {
	dir string
}

func NewConfluentHomeMock() *ConfluentHome {
	count++
	id := fmt.Sprintf("confluent%d", count)

	return &ConfluentHome{
		dir: filepath.Join(os.TempDir(), id),
	}
}

func (c ConfluentHome) Setup() error {
	return os.Setenv("CONFLUENT_HOME", c.dir)
}

func (c ConfluentHome) AddFile(file string) error {
	path := filepath.Join(c.dir, file)

	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return err
	}

	if _, err := os.Create(path); err != nil {
		return err
	}

	return nil
}

func (c ConfluentHome) TearDown() error {
	return os.RemoveAll(c.dir)
}
