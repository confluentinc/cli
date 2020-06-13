package shell

import (
	"testing"

	goprompt "github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/config"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/shell/completer"
	"github.com/confluentinc/cli/internal/pkg/shell/prompt"
)

func Test_livePrefixFunc(t *testing.T) {
	type args struct {
		cliPrompt *prompt.ShellPrompt
	}
	tests := []struct {
		name          string
		args          args
		wantPrefix    string
		wantUsePrefix bool
	}{
		{
			name: "succeed setting prefix",
			args: args{
				cliPrompt: prompt.NewShellPrompt(
					&cobra.Command{},
					completer.NewCobraCompleter(&cobra.Command{}),
					&v3.Config{
						BaseConfig: &config.BaseConfig{
							Params: &config.Params{CLIName: "testingtesting"},
						}},
					goprompt.OptionParser(new(mockConsoleParser)),
				)},
			wantPrefix:    "testingtesting > ",
			wantUsePrefix: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix, usePrefix := livePrefixFunc(tt.args.cliPrompt)()
			require.Equal(t, tt.wantPrefix, prefix)
			require.Equal(t, tt.wantUsePrefix, usePrefix)
		})
	}
}

type mockConsoleParser struct {
}

func (c mockConsoleParser) Setup() error {
	return nil
}

func (c mockConsoleParser) TearDown() error {
	return nil
}

func (c mockConsoleParser) GetKey(b []byte) goprompt.Key {
	return 0
}

func (c mockConsoleParser) GetWinSize() *goprompt.WinSize {
	return nil
}

func (c mockConsoleParser) Read() ([]byte, error) {
	return nil, nil
}
