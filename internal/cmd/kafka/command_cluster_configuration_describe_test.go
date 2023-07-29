package kafka

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCatchConfigurationNotFound(t *testing.T) {
	err := catchConfigurationNotFound(fmt.Errorf("404 Not Found"), "configuration.dne")
	require.Error(t, err)
	require.Equal(t, `configuration "configuration.dne" not found`, err.Error())
}

func TestCatchConfigurationNotFound_Nil(t *testing.T) {
	err := catchConfigurationNotFound(nil, "configuration.dne")
	require.NoError(t, err)
}

func TestCatchConfigurationNotFound_NoMatch(t *testing.T) {
	err := catchConfigurationNotFound(fmt.Errorf("204 No Content"), "configuration.dne")
	require.Error(t, err)
	require.Equal(t, `204 No Content`, err.Error())
}
