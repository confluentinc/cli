package update

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/update/mock"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name   string
		params *ClientParams
		want   *ClientParams
	}{
		{
			name:   "should set default values (interval=24h, clock=real clock)",
			params: &ClientParams{},
			want: &ClientParams{
				CheckInterval: 24 * time.Hour,
				Clock:         clockwork.NewRealClock(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewClient(tt.params); !reflect.DeepEqual(got.ClientParams, tt.want) {
				t.Errorf("NewClient() = %v, want %#v", got.ClientParams, tt.want)
			}
		})
	}
}

func TestCheckForUpdates(t *testing.T) {
	tmpCheckFile1, err := ioutil.TempFile("", "cli-test1-*")
	require.NoError(t, err)
	defer os.Remove(tmpCheckFile1.Name())

	// we don't need to cross compile for tests
	u, err := user.Current()
	require.NoError(t, err)
	tmpCheckFile2Handle, err := ioutil.TempFile(u.HomeDir, "cli-test2-*")
	// replace the user homedir with ~ to test expansion by our own code
	tmpCheckFile2 := strings.Replace(tmpCheckFile2Handle.Name(), u.HomeDir, "~", 1)
	defer os.Remove(tmpCheckFile2Handle.Name())

	require.NoError(t, err)
	type args struct {
		name           string
		currentVersion string
		forceCheck     bool
	}
	tests := []struct {
		name                string
		client              *client
		args                args
		wantUpdateAvailable bool
		wantLatestVersion   string
		wantErr             bool
	}{
		{
			name: "should err if currentVersion isn't semver",
			client: NewClient(&ClientParams{
				Repository: &mock.Repository{},
				Logger:     log.New(),
			}),
			args: args{
				name:           "my-cli",
				currentVersion: "gobbledegook",
			},
			wantUpdateAvailable: false,
			wantLatestVersion:   "gobbledegook",
			wantErr:             true,
		},
		{
			name: "should err if can't get versions",
			client: NewClient(&ClientParams{
				Repository: &mock.Repository{
					GetAvailableVersionsFunc: func(name string) (version.Collection, error) {
						return nil, errors.New("zap")
					},
				},
				Logger: log.New(),
			}),
			args: args{
				name:           "my-cli",
				currentVersion: "v1.2.3",
			},
			wantUpdateAvailable: false,
			wantLatestVersion:   "v1.2.3",
			wantErr:             true,
		},
		{
			name: "should return the most recent version",
			client: NewClient(&ClientParams{
				Repository: &mock.Repository{
					GetAvailableVersionsFunc: func(name string) (version.Collection, error) {
						v1, _ := version.NewSemver("v1")
						v2, _ := version.NewSemver("v2")
						v3, _ := version.NewSemver("v3")
						return version.Collection{
							v1, v2, v3,
						}, nil
					},
				},
				Logger: log.New(),
			}),
			args: args{
				name:           "my-cli",
				currentVersion: "v1.2.3",
			},
			wantUpdateAvailable: true,
			wantLatestVersion:   "v3",
			wantErr:             false,
		},
		{
			name: "should not check again if checked recently",
			client: NewClient(&ClientParams{
				Repository: &mock.Repository{
					GetAvailableVersionsFunc: func(name string) (version.Collection, error) {
						require.Fail(t, "Shouldn't be called")
						return nil, errors.New("whoops")
					},
				},
				Logger: log.New(),
				// This check file was created by the TmpFile process, modtime is current, so should skip check
				CheckFile: tmpCheckFile1.Name(),
			}),
			args: args{
				name:           "my-cli",
				currentVersion: "v1.2.3",
			},
			wantUpdateAvailable: false,
			wantLatestVersion:   "v1.2.3",
			wantErr:             false,
		},
		{
			name: "should respect forceCheck even if you checked recently",
			client: NewClient(&ClientParams{
				Repository: &mock.Repository{
					GetAvailableVersionsFunc: func(name string) (version.Collection, error) {
						v1, _ := version.NewSemver("v1")
						v2, _ := version.NewSemver("v2")
						v3, _ := version.NewSemver("v3")
						return version.Collection{
							v1, v2, v3,
						}, nil
					},
				},
				Logger: log.New(),
				// This check file was created by the TmpFile process, modtime is current, so should skip check
				CheckFile: tmpCheckFile1.Name(),
			}),
			args: args{
				name:           "my-cli",
				currentVersion: "v1.2.3",
				forceCheck:     true,
			},
			wantUpdateAvailable: true,
			wantLatestVersion:   "v3",
			wantErr:             false,
		},
		{
			name: "should err if you can't create the CheckFile",
			client: NewClient(&ClientParams{
				Repository: &mock.Repository{
					GetAvailableVersionsFunc: func(name string) (version.Collection, error) {
						v1, _ := version.NewSemver("v1")
						return version.Collection{v1}, nil
					},
				},
				Logger: log.New(),
				// This file doesn't exist but you won't have permission to create it
				CheckFile: "/sbin/cant-write-here",
			}),
			args: args{
				name:           "my-cli",
				currentVersion: "v1.2.3",
			},
			wantUpdateAvailable: false,
			wantLatestVersion:   "v1.2.3",
			wantErr:             true,
		},
		{
			name: "should err if you can't touch the CheckFile",
			client: NewClient(&ClientParams{
				Repository: &mock.Repository{
					GetAvailableVersionsFunc: func(name string) (version.Collection, error) {
						v1, _ := version.NewSemver("v1")
						return version.Collection{v1}, nil
					},
				},
				Logger: log.New(),
				// This file doesn't exist but you won't have permission to touch it
				CheckFile: "/sbin/ping",
			}),
			args: args{
				name:           "my-cli",
				currentVersion: "v1.2.3",
			},
			wantUpdateAvailable: false,
			wantLatestVersion:   "v1.2.3",
			wantErr:             true,
		},
		{
			name: "should support files in your homedir",
			client: NewClient(&ClientParams{
				Repository: &mock.Repository{
					GetAvailableVersionsFunc: func(name string) (version.Collection, error) {
						require.Fail(t, "Shouldn't be called")
						return nil, errors.New("whoops")
					},
				},
				Logger: log.New(),
				// This check file name has ~ in the path
				CheckFile: tmpCheckFile2,
			}),
			args: args{
				name:           "my-cli",
				currentVersion: "v1.2.3",
			},
			wantUpdateAvailable: false,
			wantLatestVersion:   "v1.2.3",
			wantErr:             false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUpdateAvailable, gotLatestVersion, err := tt.client.CheckForUpdates(tt.args.name, tt.args.currentVersion, tt.args.forceCheck)
			if (err != nil) != tt.wantErr {
				t.Errorf("client.CheckForUpdates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUpdateAvailable != tt.wantUpdateAvailable {
				t.Errorf("client.CheckForUpdates() gotUpdateAvailable = %v, want %v", gotUpdateAvailable, tt.wantUpdateAvailable)
			}
			if gotLatestVersion != tt.wantLatestVersion {
				t.Errorf("client.CheckForUpdates() gotLatestVersion = %v, want %v", gotLatestVersion, tt.wantLatestVersion)
			}
		})
	}
}

func TestCheckForUpdates_BehaviorOverTime(t *testing.T) {
	req := require.New(t)

	tmpDir, err := ioutil.TempDir("", "cli-test3-*")
	req.NoError(err)
	defer os.RemoveAll(tmpDir)
	checkFile := fmt.Sprintf("%s/new-check-file", tmpDir)

	repo := &mock.Repository{
		GetAvailableVersionsFunc: func(name string) (version.Collection, error) {
			v1, _ := version.NewSemver("v1")
			v2, _ := version.NewSemver("v2")
			v3, _ := version.NewSemver("v3")
			return version.Collection{
				v1, v2, v3,
			}, nil
		},
	}
	client := NewClient(&ClientParams{
		Repository: repo,
		Logger:     log.New(),
		CheckFile:  checkFile,
		Clock:      clockwork.NewFakeClockAt(time.Now()),
	})

	// Should check and find update
	updateAvailable, latestVersion, err := client.CheckForUpdates("my-cli", "v1.2.3", false)
	req.NoError(err)
	req.True(updateAvailable)
	req.Equal("v3", latestVersion)
	req.True(repo.GetAvailableVersionsCalled())

	// Shouldn't check anymore for 24 hours
	lastCheck := client.Clock.Now()
	for i := 0; i < 3; i++ {
		lastCheck = lastCheck.Add(8 * time.Hour).Add(-1 * time.Second)
		client.Clock = clockwork.NewFakeClockAt(lastCheck)
		repo.Reset()

		updateAvailable, latestVersion, err = client.CheckForUpdates("my-cli", "v1.2.3", false)
		req.False(repo.GetAvailableVersionsCalled())
	}

	// 5 days pass...
	client.Clock = clockwork.NewFakeClockAt(time.Now().Add(5 * 24 * time.Hour))

	// Should check and find update
	updateAvailable, latestVersion, err = client.CheckForUpdates("my-cli", "v1.2.3", false)
	req.NoError(err)
	req.True(updateAvailable)
	req.Equal("v3", latestVersion)
	req.True(repo.GetAvailableVersionsCalled())

	// Shouldn't check anymore for 24 hours
	for i := 0; i < 3; i++ {
		lastCheck = lastCheck.Add(8 * time.Hour).Add(-1 * time.Second)
		client.Clock = clockwork.NewFakeClockAt(lastCheck)
		repo.Reset()

		updateAvailable, latestVersion, err = client.CheckForUpdates("my-cli", "v1.2.3", false)
		req.False(repo.GetAvailableVersionsCalled())
	}
}

func TestCheckForUpdates_NoCheckFileGiven(t *testing.T) {
	req := require.New(t)

	repo := &mock.Repository{
		GetAvailableVersionsFunc: func(name string) (version.Collection, error) {
			v1, _ := version.NewSemver("v1")
			v2, _ := version.NewSemver("v2")
			v3, _ := version.NewSemver("v3")
			return version.Collection{
				v1, v2, v3,
			}, nil
		},
	}
	client := NewClient(&ClientParams{
		Repository: repo,
		Logger:     log.New(),
		Clock:      clockwork.NewFakeClockAt(time.Now()),
	})

	// Should check and find the update nonstop
	for i := 0; i < 3; i++ {
		updateAvailable, latestVersion, err := client.CheckForUpdates("my-cli", "v1.2.3", false)
		req.NoError(err)
		req.True(updateAvailable)
		req.Equal("v3", latestVersion)
		req.True(repo.GetAvailableVersionsCalled())
		repo.Reset()
	}
}
