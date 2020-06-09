package local

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/mock"
)

func TestCurrentCreateAndTrackDir(t *testing.T) {
	req := require.New(t)

	cp := mock.NewConfluentPlatform()
	defer cp.TearDown()

	req.NoError(cp.NewConfluentCurrent())

	out, err := mockLocalCommand("current")
	req.NoError(err)
	req.Contains(out, cp.ConfluentCurrent)

	trackingFile := filepath.Join(cp.ConfluentCurrent, "confluent.current")
	req.FileExists(trackingFile)
}

func TestCurrentGetTrackedDir(t *testing.T) {
	req := require.New(t)

	cp := mock.NewConfluentPlatform()
	defer cp.TearDown()

	req.NoError(cp.NewConfluentCurrent())

	trackingFile := filepath.Join(cp.ConfluentCurrent, "confluent.current")
	req.NoError(ioutil.WriteFile(trackingFile, []byte("test"), 0777))

	out, err := mockLocalCommand("current")
	req.NoError(err)
	req.Contains(out, "test")
}
