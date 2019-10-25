package cmd_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/update/mock"
	"github.com/confluentinc/cli/internal/pkg/version"
)

func TestPreRun_Anonymous_SetLoggingLevel(t *testing.T) {
	type fields struct {
		Logger  *log.Logger
		Command string
	}
	tests := []struct {
		name   string
		fields fields
		want   log.Level
	}{
		{
			name: "default logging level",
			fields: fields{
				Logger:  log.New(),
				Command: "help",
			},
			want: log.ERROR,
		},
		{
			name: "warn logging level",
			fields: fields{
				Logger:  log.New(),
				Command: "help -v",
			},
			want: log.WARN,
		},
		{
			name: "info logging level",
			fields: fields{
				Logger:  log.New(),
				Command: "help -vv",
			},
			want: log.INFO,
		},
		{
			name: "debug logging level",
			fields: fields{
				Logger:  log.New(),
				Command: "help -vvv",
			},
			want: log.DEBUG,
		},
		{
			name: "trace logging level",
			fields: fields{
				Logger:  log.New(),
				Command: "help -vvvv",
			},
			want: log.TRACE,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			require.NoError(t, cfg.Load())

			ver := version.NewVersion("ccloud", "Confluent Cloud CLI", "https://confluent.cloud; support@confluent.io", "1.2.3", "abc1234", "01/23/45", "CI")

			r := &pcmd.PreRun{
				Version: ver.Version,
				Logger:  tt.fields.Logger,
				Config:  cfg,
				UpdateClient: &mock.Client{
					CheckForUpdatesFunc: func(n, v string, f bool) (bool, string, error) {
						return false, "", nil
					},
				},
			}

			root := &cobra.Command{Run: func(cmd *cobra.Command, args []string) {}}
			root.Flags().CountP("verbose", "v", "Increase verbosity")

			args := strings.Split(tt.fields.Command, " ")
			_, err := pcmd.ExecuteCommand(root, args...)
			require.NoError(t, err)

			err = r.Anonymous()(root, args)
			require.NoError(t, err)

			got := tt.fields.Logger.GetLevel()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PreRun.HasAPIKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPreRun_HasAPIKey_SetupLoggingAndCheckForUpdates(t *testing.T) {
	cfg := &config.Config{}
	require.NoError(t, cfg.Load())

	ver := version.NewVersion("ccloud", "Confluent Cloud CLI", "https://confluent.cloud; support@confluent.io", "1.2.3", "abc1234", "01/23/45", "CI")

	calledAnonymous := false
	r := &pcmd.PreRun{
		Version: ver.Version,
		Logger:  log.New(),
		Config:  cfg,
		UpdateClient: &mock.Client{
			CheckForUpdatesFunc: func(n, v string, f bool) (bool, string, error) {
				calledAnonymous = true
				return false, "", nil
			},
		},
	}

	root := &cobra.Command{Run: func(cmd *cobra.Command, args []string) {}}
	root.Flags().CountP("verbose", "v", "Increase verbosity")

	args := strings.Split("help", " ")
	_, err := pcmd.ExecuteCommand(root, args...)
	require.NoError(t, err)

	err = r.Anonymous()(root, args)
	require.NoError(t, err)

	if !calledAnonymous {
		t.Errorf("PreRun.HasAPIKey() didn't call the Anonymous() helper to set logging level and updates")
	}
}
