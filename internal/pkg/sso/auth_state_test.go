package sso

import (
	"github.com/confluentinc/cli/internal/pkg/config"
	"testing"
	"github.com/stretchr/testify/require"
)

func TestState(t *testing.T) {
	config := &config.Config{AuthURL: "https://devel.cpdev.cloud"}
	state, err := NewState(config)
	require.NoError(t, err)
	require.Equal(t, state.SSOProviderDomain, "login.confluent-dev.io")


}
