package context

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/mock"
)

func TestParseStringFlag(t *testing.T) {
	data := "data"

	cmd := &cobra.Command{}
	cmd.Flags().String("flag", "", "")

	c := &command{
		CLICommand: &pcmd.CLICommand{Command: cmd},
		resolver: &pcmd.FlagResolverImpl{
			Out: new(bytes.Buffer),
			Prompt: &mock.Prompt{
				IsPipeFunc:   func() (bool, error) { return false, nil },
				ReadLineFunc: func() (string, error) { return data, nil },
			},
		},
	}

	out, err := c.parseStringFlag(cmd, "flag", "Flag: ", false)
	require.NoError(t, err)
	require.Equal(t, data, out)
}

func TestParseStringFlag_ErrEmpty(t *testing.T) {
	data := "    "

	cmd := &cobra.Command{}
	cmd.Flags().String("flag", "", "")

	c := &command{
		CLICommand: &pcmd.CLICommand{Command: cmd},
		resolver: &pcmd.FlagResolverImpl{
			Out: new(bytes.Buffer),
			Prompt: &mock.Prompt{
				IsPipeFunc:   func() (bool, error) { return false, nil },
				ReadLineFunc: func() (string, error) { return data, nil },
			},
		},
	}

	_, err := c.parseStringFlag(cmd, "flag", "Flag: ", false)
	require.Error(t, err)
	require.Equal(t, fmt.Sprintf(errors.CannotBeEmptyErrorMsg, "flag"), err.Error())
}
