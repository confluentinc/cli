package secret

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsPathValid(t *testing.T) {
	t.Run("IsPathValid: empty path returns false", func(t *testing.T) {
		req := require.New(t)
		valid := IsPathValid("")
		req.False(valid)
	})
}

func TestLoadPropertiesFile(t *testing.T) {
	t.Run("LoadPropertiesFile: empty path yields error", func(t *testing.T) {
		req := require.New(t)
		_, err := LoadPropertiesFile("")
		req.Error(err)
	})
}
