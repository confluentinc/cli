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
	req.NoError(cp.NewConfluentCurrentDir())

	out, err := mockLocalCommand("current")
	req.NoError(err)
	req.Contains(out, cp.ConfluentCurrent)

	trackingFile := filepath.Join(cp.ConfluentCurrentDir, "confluent.current")
	req.FileExists(trackingFile)
}

func TestCurrentGetTrackedDir(t *testing.T) {
	req := require.New(t)

	cp := mock.NewConfluentPlatform()
	defer cp.TearDown()
	req.NoError(cp.NewConfluentCurrent())

	req.NoError(ioutil.WriteFile(cp.TrackingFile, []byte("test"), 0777))

	out, err := mockLocalCommand("current")
	req.NoError(err)
	req.Contains(out, "test")
}
