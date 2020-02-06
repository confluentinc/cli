package cmd_test

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/jonboulle/clockwork"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"github.com/confluentinc/cli/internal/pkg/log"
	pmock "github.com/confluentinc/cli/internal/pkg/mock"
	"github.com/confluentinc/cli/internal/pkg/update/mock"
	cliMock "github.com/confluentinc/cli/mock"
)

var (
	expiredAuthTokenForDevCLoud = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJvcmdhbml6YXRpb25JZCI6MT" +
		"U5NCwidXNlcklkIjoxNTM3MiwiZXhwIjoxNTc0NzIwODgzLCJqdGkiOiJkMzFlYjc2OC0zNzIzLTQ4MTEtYjg3" +
		"Zi1lMTQ2YTQyYmMyMjciLCJpYXQiOjE1NzQ3MTcyODMsImlzcyI6IkNvbmZsdWVudCIsInN1YiI6IjE1MzcyIn" +
		"0.r9o6HEaacidXV899sjYDajCfVd_Tczyfk541jzidw8r0TRGz74RxL2UFK0aGyR4tNrJRSOJlYHSEBNMV7" +
		"J1sEzdGj_mYbvdAL8feH3Sj0uOf1BSKEdhOLsZbQRPn1TnUwUI0ujxjvY3V4l9unXjdRcNceQx1RcAIm8JEo" +
		"Vjpgsb5MRQWYHlTTEwJe5MVY-dZZEsq40YzlchmFi8LVYCxY3rtwEtINbFJx7K-0rW-GJWyek2zRMiUDtmXI" +
		"o8C87TmR9JfLAhLGYKYB-sMnX1FWQs1GSEf9CBGerhZ6T4wwTu_GCVEqg_kDZpGxx1V3nTr0K_lHN8QxFHoJA" +
		"ccbtRHKFuQZaXkJjhsq4i6q9OV-wgL_G7y003Z-hRiBvdBPoEqecXOfI6HKYbzfv9P9N2p0UnfPF2fZWitcmd" +
		"55IgHZ15zwDkFqixoV1hY_tG7dWtQNZIlPDabgm5UH0mc7GS2dh9Z5spZTvqH8xZ0SFF6T5-iFqpJjm6wkzMd6" +
		"1u9UuWTTTNG-Nr_8abS0cYfChZIXde3D1so2KhG4r6uAB1onlNWK4Gq2Lc9uT_r2tKcGDqyZWFPvVtAepr8duW" +
		"ts27QsDs7BvMnwSkUjGv6scSJZWX1fMZbXh7zd0Khg_13dWshAyE935n46T4S7VJm9JhZLEwUcoOPOhWmVcJn5xSJ-YQ"
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
			ver := pmock.NewVersionMock()
			r := &pcmd.PreRun{
				Version: ver,
				Logger:  tt.fields.Logger,
				UpdateClient: &mock.Client{
					CheckForUpdatesFunc: func(n, v string, f bool) (bool, string, error) {
						return false, "", nil
					},
				},
				FlagResolver: &pcmd.FlagResolverImpl{
					Prompt: &pcmd.RealPrompt{},
					Out:    os.Stdout,
				},
				Analytics: cliMock.NewDummyAnalyticsMock(),
			}

			root := &cobra.Command{Run: func(cmd *cobra.Command, args []string) {}}
			root.Flags().CountP("verbose", "v", "Increase verbosity")
			cfg := v2.New(nil)
			require.NoError(t, cfg.Load())
			rootCmd := pcmd.NewAnonymousCLICommand(root, cfg, r)

			args := strings.Split(tt.fields.Command, " ")
			_, err := pcmd.ExecuteCommand(rootCmd.Command, args...)
			require.NoError(t, err)

			got := tt.fields.Logger.GetLevel()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PreRun.HasAPIKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPreRun_HasAPIKey_SetupLoggingAndCheckForUpdates(t *testing.T) {
	cfg := v2.New(nil)
	require.NoError(t, cfg.Load())

	ver := pmock.NewVersionMock()

	calledAnonymous := false
	r := &pcmd.PreRun{
		Version: ver,
		Logger:  log.New(),
		UpdateClient: &mock.Client{
			CheckForUpdatesFunc: func(n, v string, f bool) (bool, string, error) {
				calledAnonymous = true
				return false, "", nil
			},
		},
		FlagResolver: &pcmd.FlagResolverImpl{
			Prompt: &pcmd.RealPrompt{},
			Out:    os.Stdout,
		},
		Analytics: cliMock.NewDummyAnalyticsMock(),
	}

	root := &cobra.Command{Run: func(cmd *cobra.Command, args []string) {}}
	root.Flags().CountP("verbose", "v", "Increase verbosity")
	rootCmd := pcmd.NewAnonymousCLICommand(root, cfg, r)
	args := strings.Split("help", " ")
	_, err := pcmd.ExecuteCommand(rootCmd.Command, args...)
	require.NoError(t, err)

	if !calledAnonymous {
		t.Errorf("PreRun.HasAPIKey() didn't call the Anonymous() helper to set logging level and updates")
	}
}

func TestPreRun_CallsAnalyticsTrackCommand(t *testing.T) {
	cfg := v2.New(nil)
	require.NoError(t, cfg.Load())

	ver := pmock.NewVersionMock()
	analyticsClient := cliMock.NewDummyAnalyticsMock()

	r := &pcmd.PreRun{
		Version: ver,
		Logger:  log.New(),
		UpdateClient: &mock.Client{
			CheckForUpdatesFunc: func(n, v string, f bool) (bool, string, error) {
				return false, "", nil
			},
		},
		FlagResolver: &pcmd.FlagResolverImpl{
			Prompt: &pcmd.RealPrompt{},
			Out:    os.Stdout,
		},
		Analytics: analyticsClient,
	}

	root := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {},
	}
	rootCmd := pcmd.NewAnonymousCLICommand(root, cfg, r)
	root.Flags().CountP("verbose", "v", "Increase verbosity")

	_, err := pcmd.ExecuteCommand(rootCmd.Command)
	require.NoError(t, err)

	require.True(t, analyticsClient.TrackCommandCalled())
}

func TestPreRun_TokenExpires(t *testing.T) {
	cfg := v2.AuthenticatedConfigMock()
	cfg.Context().State.AuthToken = expiredAuthTokenForDevCLoud

	ver := pmock.NewVersionMock()
	analyticsClient := cliMock.NewDummyAnalyticsMock()

	r := &pcmd.PreRun{
		Version: ver,
		Logger:  log.New(),
		UpdateClient: &mock.Client{
			CheckForUpdatesFunc: func(n, v string, f bool) (bool, string, error) {
				return false, "", nil
			},
		},
		FlagResolver: &pcmd.FlagResolverImpl{
			Prompt: &pcmd.RealPrompt{},
			Out:    os.Stdout,
		},
		Analytics: analyticsClient,
		Clock:     clockwork.NewRealClock(),
	}

	root := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {},
	}
	rootCmd := pcmd.NewAnonymousCLICommand(root, cfg, r)
	root.Flags().CountP("verbose", "v", "Increase verbosity")

	_, err := pcmd.ExecuteCommand(rootCmd.Command)
	require.NoError(t, err)

	// Check auth is nil for now, until there is a better to create a fake logged in user and check if it's logged out
	require.Nil(t, cfg.Context().State.Auth)
	require.True(t, analyticsClient.SessionTimedOutCalled())
}
