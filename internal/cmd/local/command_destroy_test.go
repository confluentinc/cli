package local

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/mock"
)

func TestDestroy(t *testing.T) {
	req := require.New(t)

	cp := mock.NewConfluentPlatform()
	defer cp.TearDown()
	req.NoError(cp.NewConfluentHome())
	req.NoError(cp.NewConfluentCurrent())

	out, err := mockLocalCommand("destroy")
	req.NoError(err)

	for service := range services {
		req.Contains(out, fmt.Sprintf("%s is [DOWN]\n", service))
	}
	req.Contains(out, fmt.Sprintf("Deleting: %s\n", cp.ConfluentCurrent))

	req.NoDirExists(cp.ConfluentCurrent)
	req.NoFileExists(cp.TrackingFile)
}
