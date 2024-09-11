package update

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const checksums = `a393  confluent_darwin_amd64
fefb  confluent_linux_amd64
921a  confluent_windows_amd64.exe`

func TestFindChecksum(t *testing.T) {
	checksum, err := findChecksum(checksums, "confluent_darwin_amd64")
	require.NoError(t, err)
	require.Equal(t, []byte{0xa3, 0x93}, checksum)
}

func TestFindChecksum_DoesNotExist(t *testing.T) {
	_, err := findChecksum(checksums, "does_not_exist")
	require.Error(t, err)
}
