package update

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/errors"
	pio "github.com/confluentinc/cli/internal/pkg/io"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/mock"
	updateMock "github.com/confluentinc/cli/internal/pkg/update/mock"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name   string
		params *ClientParams
		want   *client
	}{
		{
			name:   "should set default values (interval=24h, clock=real clock, fs=real fs, os=real os)",
			params: &ClientParams{},
			want: &client{
				ClientParams: &ClientParams{CheckInterval: 24 * time.Hour, OS: runtime.GOOS},
				clock:        clockwork.NewRealClock(),
				fs:           &pio.RealFileSystem{},
			},
		},
		{
			name: "should set provided values",
			params: &ClientParams{
				CheckInterval: 48 * time.Hour,
				OS:            "duckduckgoos",
				DisableCheck:  true,
			},
			want: &client{
				ClientParams: &ClientParams{
					CheckInterval: 48 * time.Hour,
					OS:            "duckduckgoos",
					DisableCheck:  true,
				},
				clock: clockwork.NewRealClock(),
				fs:    &pio.RealFileSystem{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewClient(tt.params); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewClient() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestCheckForUpdates(t *testing.T) {
	tmpCheckFile1, err := ioutil.TempFile("", "cli-test1-")
	require.NoError(t, err)
	defer os.Remove(tmpCheckFile1.Name())

	// we don't need to cross compile for tests
	u, err := user.Current()
	require.NoError(t, err)
	tmpCheckFile2Handle, err := ioutil.TempFile(u.HomeDir, "cli-test2-")
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
				Repository: &updateMock.Repository{},
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
				Repository: &updateMock.Repository{
					GetLatestBinaryVersionFunc: func(name string) (i *version.Version, e error) {
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
			name: "should return the new version",
			client: NewClient(&ClientParams{
				Repository: &updateMock.Repository{
					GetLatestBinaryVersionFunc: func(name string) (*version.Version, error) {
						v3, _ := version.NewSemver("v3")
						return v3, nil
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
				Repository: &updateMock.Repository{
					GetLatestBinaryVersionFunc: func(name string) (*version.Version, error) {
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
				Repository: &updateMock.Repository{
					GetLatestBinaryVersionFunc: func(name string) (*version.Version, error) {
						v3, _ := version.NewSemver("v3")
						return v3, nil
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
				Repository: &updateMock.Repository{
					GetLatestBinaryVersionFunc: func(name string) (*version.Version, error) {
						v1, _ := version.NewSemver("v2")
						return v1, nil
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
				Repository: &updateMock.Repository{
					GetLatestBinaryVersionFunc: func(name string) (*version.Version, error) {
						v1, _ := version.NewSemver("v2")
						return v1, nil
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
				Repository: &updateMock.Repository{
					GetLatestBinaryVersionFunc: func(name string) (*version.Version, error) {
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
		{
			name: "should not check if disabled",
			client: NewClient(&ClientParams{
				Repository: &updateMock.Repository{
					GetLatestBinaryVersionFunc: func(name string) (*version.Version, error) {
						require.Fail(t, "Shouldn't be called")
						return nil, errors.New("whoops")
					},
				},
				Logger:       log.New(),
				DisableCheck: true,
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
			name: "checks - error",
			client: NewClient(&ClientParams{
				Repository: &updateMock.Repository{
					GetLatestBinaryVersionFunc: func(name string) (*version.Version, error) {
						return nil, errors.New("whoops")
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
			name: "checks - success - update",
			client: NewClient(&ClientParams{
				Repository: &updateMock.Repository{
					GetLatestBinaryVersionFunc: func(name string) (*version.Version, error) {
						return version.Must(version.NewVersion("v1.2.4")), nil
					},
				},
				Logger: log.New(),
			}),
			args: args{
				name:           "my-cli",
				currentVersion: "v1.2.3",
			},
			wantUpdateAvailable: true,
			wantLatestVersion:   "v1.2.4",
			wantErr:             false,
		},
		{
			name: "checks - success - same version",
			client: NewClient(&ClientParams{
				Repository: &updateMock.Repository{
					GetLatestBinaryVersionFunc: func(name string) (*version.Version, error) {
						return version.Must(version.NewVersion("v1.2.4")), nil
					},
				},
				Logger: log.New(),
			}),
			args: args{
				name:           "my-cli",
				currentVersion: "v1.2.4",
			},
			wantUpdateAvailable: false,
			wantLatestVersion:   "v1.2.4",
			wantErr:             false,
		},
		{
			name: "checks - success - hyphen no update",
			client: NewClient(&ClientParams{
				Repository: &updateMock.Repository{
					GetLatestBinaryVersionFunc: func(name string) (*version.Version, error) {
						return version.Must(version.NewVersion("v0.238.0")), nil
					},
				},
				Logger: log.New(),
			}),
			args: args{
				name:           "my-cli",
				currentVersion: "v0.238.0-7-g5060ef4",
			},
			wantUpdateAvailable: false,
			wantLatestVersion:   "v0.238.0-7-g5060ef4",
			wantErr:             false,
		},
		{
			name: "checks - success - hyphen same version",
			client: NewClient(&ClientParams{
				Repository: &updateMock.Repository{
					GetLatestBinaryVersionFunc: func(name string) (*version.Version, error) {
						return version.Must(version.NewVersion("v0.238.0-7-g5060ef4")), nil
					},
				},
				Logger: log.New(),
			}),
			args: args{
				name:           "my-cli",
				currentVersion: "v0.238.0-7-g5060ef4",
			},
			wantUpdateAvailable: false,
			wantLatestVersion:   "v0.238.0-7-g5060ef4",
			wantErr:             false,
		},
		{
			name: "checks - success - hyphen update",
			client: NewClient(&ClientParams{
				Repository: &updateMock.Repository{
					GetLatestBinaryVersionFunc: func(name string) (*version.Version, error) {
						return version.Must(version.NewVersion("v0.238.0-7-g5060ef4")), nil
					},
				},
				Logger: log.New(),
			}),
			args: args{
				name:           "my-cli",
				currentVersion: "v0.238.0",
			},
			wantUpdateAvailable: true,
			wantLatestVersion:   "v0.238.0-7-g5060ef4",
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

	tmpDir, err := ioutil.TempDir("", "cli-test3-")
	req.NoError(err)
	defer os.RemoveAll(tmpDir)
	checkFile := filepath.FromSlash(fmt.Sprintf("%s/new-check-file", tmpDir))

	repo := &updateMock.Repository{
		GetLatestBinaryVersionFunc: func(name string) (*version.Version, error) {
			v3, _ := version.NewSemver("v3")
			return v3, nil
		},
	}
	clock := clockwork.NewFakeClockAt(time.Now())
	client := NewClient(&ClientParams{
		Repository: repo,
		Logger:     log.New(),
		CheckFile:  checkFile,
	})
	client.clock = clock

	// Should check and find update
	updateAvailable, latestVersion, err := client.CheckForUpdates("my-cli", "v1.2.3", false)
	req.NoError(err)
	req.True(updateAvailable)
	req.Equal("v3", latestVersion)
	req.True(repo.GetLatestBinaryVersionCalled())

	// Shouldn't check anymore for 24 hours
	for i := 0; i < 3; i++ {
		clock.Advance(8*time.Hour + -1*time.Second)
		repo.Reset()

		_, _, _ = client.CheckForUpdates("my-cli", "v1.2.3", false)
		req.False(repo.GetLatestBinaryVersionCalled())
	}

	// 5 days pass...
	clock.Advance(5 * 24 * time.Hour)

	// Should check and find update
	updateAvailable, latestVersion, err = client.CheckForUpdates("my-cli", "v1.2.3", false)
	req.NoError(err)
	req.True(updateAvailable)
	req.Equal("v3", latestVersion)
	req.True(repo.GetLatestBinaryVersionCalled())

	// Shouldn't check anymore for 24 hours
	for i := 0; i < 3; i++ {
		clock.Advance(8*time.Hour + -1*time.Second)
		repo.Reset()

		_, _, _ = client.CheckForUpdates("my-cli", "v1.2.3", false)
		req.False(repo.GetLatestBinaryVersionCalled())
	}

	// Finally we should check once more
	clock.Advance(3 * time.Second)
	repo.Reset()
	_, _, _ = client.CheckForUpdates("my-cli", "v1.2.3", false)
	req.True(repo.GetLatestBinaryVersionCalled())
}

func TestCheckForUpdates_NoCheckFileGiven(t *testing.T) {
	req := require.New(t)

	repo := &updateMock.Repository{
		GetLatestBinaryVersionFunc: func(name string) (*version.Version, error) {
			v3, _ := version.NewSemver("v3")
			return v3, nil
		},
	}
	client := NewClient(&ClientParams{
		Repository: repo,
		Logger:     log.New(),
	})
	client.clock = clockwork.NewFakeClockAt(time.Now())

	// Should check for updates every time if no CheckFile given to serve as the "last check" cache
	for i := 0; i < 3; i++ {
		updateAvailable, latestVersion, err := client.CheckForUpdates("my-cli", "v1.2.3", false)
		req.NoError(err)
		req.True(updateAvailable)
		req.Equal("v3", latestVersion)
		req.True(repo.GetLatestBinaryVersionCalled())
		repo.Reset()
	}
}

func TestGetLatestReleaseNotes(t *testing.T) {
	releaseNotesVersion := "1.0.0"
	releaseNotes := "nice release notes"
	tests := []struct {
		name             string
		client           *client
		wantVersion      string
		wantReleaseNotes string
		wantErr          bool
	}{
		{
			name: "success",
			client: NewClient(&ClientParams{
				Repository: &updateMock.Repository{
					GetLatestReleaseNotesVersionFunc: func() (*version.Version, error) {
						v, _ := version.NewSemver(releaseNotesVersion)
						return v, nil
					},
					DownloadReleaseNotesFunc: func(version string) (s string, e error) {
						return releaseNotes, nil
					},
				},
				Logger: log.New(),
			}),
			wantVersion:      releaseNotesVersion,
			wantReleaseNotes: releaseNotes,
			wantErr:          false,
		},
		{
			name: "error getting release notes version",
			client: NewClient(&ClientParams{
				Repository: &updateMock.Repository{
					GetLatestReleaseNotesVersionFunc: func() (*version.Version, error) {
						return nil, errors.New("whoops")
					},
					DownloadReleaseNotesFunc: func(version string) (s string, e error) {
						return "", nil
					},
				},
				Logger: log.New(),
			}),
			wantErr: true,
		},
		{
			name: "error downloading release notes",
			client: NewClient(&ClientParams{
				Repository: &updateMock.Repository{
					GetLatestReleaseNotesVersionFunc: func() (*version.Version, error) {
						v, _ := version.NewSemver("v1")
						return v, nil
					},
					DownloadReleaseNotesFunc: func(version string) (s string, e error) {
						return "", errors.New("whoops")
					},
				},
				Logger: log.New(),
			}),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotReleaseNotesVersion, gotReleaseNotes, err := tt.client.GetLatestReleaseNotes()
			if (err != nil) != tt.wantErr {
				t.Errorf("client.GetLatestReleaseNotes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotReleaseNotesVersion != tt.wantVersion {
				t.Errorf("client.GetLatestReleaseNotes() got releaseNotesVersion = %v, want %v", releaseNotesVersion, tt.wantVersion)
			}
			if gotReleaseNotes != tt.wantReleaseNotes {
				t.Errorf("client.GetLatestReleaseNotes() got releaseNotes = %v, want %v", gotReleaseNotes, tt.wantReleaseNotes)
			}
		})
	}
}

func TestUpdateBinary(t *testing.T) {
	req := require.New(t)

	binName := "fake_cli"

	installDir, err := ioutil.TempDir("", "cli-test4-")
	require.NoError(t, err)
	defer os.Remove(installDir)
	installedBin := filepath.FromSlash(fmt.Sprintf("%s/%s", installDir, binName))
	_ = ioutil.WriteFile(installedBin, []byte("old version"), os.ModePerm)

	downloadDir, err := ioutil.TempDir("", "cli-test5-")
	require.NoError(t, err)
	defer os.Remove(downloadDir)
	downloadedBin := filepath.FromSlash(fmt.Sprintf("%s/%s", downloadDir, binName))
	_ = ioutil.WriteFile(downloadedBin, []byte("new version"), os.ModePerm)

	clock := clockwork.NewFakeClockAt(time.Now())

	type args struct {
		name    string
		version string
		path    string
	}
	tests := []struct {
		name    string
		client  *client
		args    args
		wantErr bool
	}{
		{
			name: "can update application binary",
			client: &client{
				ClientParams: &ClientParams{
					Repository: &updateMock.Repository{
						DownloadVersionFunc: func(name, version, downloadDir string) (string, int64, error) {
							req.Equal(binName, name)
							req.Equal("v123.456.789", version)
							req.Contains(downloadDir, binName)
							clock.Advance(23 * time.Second)
							return downloadedBin, 16 * 1000 * 1000, nil
						},
					},
					Logger: log.New(),
				},
				clock: clock,
				fs:    &pio.RealFileSystem{},
			},
			args: args{
				name:    binName,
				version: "v123.456.789",
				path:    installedBin,
			},
		},
		{
			name: "err if unable to download package",
			client: &client{
				ClientParams: &ClientParams{
					Repository: &updateMock.Repository{
						DownloadVersionFunc: func(name, version, downloadDir string) (string, int64, error) {
							return "", 0, errors.New("out of disk!")
						},
					},
					Logger: log.New(),
				},
				clock: clock,
				fs:    &pio.RealFileSystem{},
			},
			args: args{
				name:    binName,
				version: "v1",
				path:    installedBin,
			},
			wantErr: true,
		},
		{
			name: "err if unable to copy binary",
			client: &client{
				ClientParams: &ClientParams{
					Repository: &updateMock.Repository{
						DownloadVersionFunc: func(name, version, downloadDir string) (string, int64, error) {
							req.Equal(binName, name)
							req.Equal("v1", version)
							req.Contains(downloadDir, binName)
							clock.Advance(23 * time.Second)
							return downloadedBin, 16 * 1000 * 1000, nil
						},
					},
					Logger: log.New(),
				},
				clock: clock,
				fs: &mock.PassThroughFileSystem{
					Mock: &mock.FileSystem{
						CopyFunc: func(dst io.Writer, src io.Reader) (i int64, e error) {
							return 0, errors.New("my dog ate my disks")
						},
					},
					FS: &pio.RealFileSystem{},
				},
			},
			args: args{
				name:    binName,
				version: "v1",
				path:    installedBin,
			},
			wantErr: true,
		},
		{
			name: "no attempt to mv binary (darwin)",
			client: &client{
				ClientParams: &ClientParams{
					Repository: &updateMock.Repository{
						DownloadVersionFunc: func(name, version, downloadDir string) (string, int64, error) {
							req.Equal(binName, name)
							req.Equal("v1", version)
							req.Contains(downloadDir, binName)
							clock.Advance(23 * time.Second)
							return downloadedBin, 16 * 1000 * 1000, nil
						},
					},
					Logger: log.New(),
					OS:     "darwin",
				},
				clock: clock,
				fs: &mock.PassThroughFileSystem{
					Mock: &mock.FileSystem{
						MoveFunc: func(src string, dst string) error {
							return errors.New("move func intentionally failed")
						},
					},
					FS: &pio.RealFileSystem{},
				},
			},
			args: args{
				name:    binName,
				version: "v1",
				path:    installedBin,
			},
			wantErr: false,
		},
		{
			name: "err if unable to mv binary (windows)",
			client: &client{
				ClientParams: &ClientParams{
					Repository: &updateMock.Repository{
						DownloadVersionFunc: func(name, version, downloadDir string) (string, int64, error) {
							req.Equal(binName, name)
							req.Equal("v1", version)
							req.Contains(downloadDir, binName)
							clock.Advance(23 * time.Second)
							return downloadedBin, 16 * 1000 * 1000, nil
						},
					},
					Logger: log.New(),
					OS:     "windows",
				},
				clock: clock,
				fs: &mock.PassThroughFileSystem{
					Mock: &mock.FileSystem{
						MoveFunc: func(src string, dst string) error {
							return errors.New("move func intentionally failed")
						},
					},
					FS: &pio.RealFileSystem{},
				},
			},
			args: args{
				name:    binName,
				version: "v1",
				path:    installedBin,
			},
			wantErr: true,
		},
		{
			name: "err if first mv succeeds, then copy fails, then second mv fails",
			client: &client{
				ClientParams: &ClientParams{
					Repository: &updateMock.Repository{
						DownloadVersionFunc: func(name, version, downloadDir string) (string, int64, error) {
							req.Equal(binName, name)
							req.Equal("v1", version)
							req.Contains(downloadDir, binName)
							clock.Advance(23 * time.Second)
							return downloadedBin, 16 * 1000 * 1000, nil
						},
					},
					Logger: log.New(),
				},
				clock: clock,
				fs: &mock.PassThroughFileSystem{
					Mock: &mock.FileSystem{
						MoveFunc: func(src string, dst string) error {
							if dst == installedBin { // this will be the case in the second mv call
								return errors.New("move func intentionally failed")
							}
							return nil
						},
						CopyFunc: func(dst io.Writer, src io.Reader) (i int64, e error) {
							return 0, errors.New("my dog ate my disks")
						},
					},
					FS: &pio.RealFileSystem{},
				},
			},
			args: args{
				name:    binName,
				version: "v1",
				path:    installedBin,
			},
			wantErr: true,
		},
		{
			name: "err if unable to chmod binary",
			client: &client{
				ClientParams: &ClientParams{
					Repository: &updateMock.Repository{
						DownloadVersionFunc: func(name, version, downloadDir string) (string, int64, error) {
							req.Equal(binName, name)
							req.Equal("v1", version)
							req.Contains(downloadDir, binName)
							clock.Advance(23 * time.Second)
							return downloadedBin, 16 * 1000 * 1000, nil
						},
					},
					Logger: log.New(),
				},
				clock: clock,
				fs: &mock.PassThroughFileSystem{
					Mock: &mock.FileSystem{
						ChmodFunc: func(name string, mode os.FileMode) error {
							return errors.New("my dog ate my disks")
						},
					},
					FS: &pio.RealFileSystem{},
				},
			},
			args: args{
				name:    binName,
				version: "v1",
				path:    installedBin,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.client.Out == nil {
				tt.client.Out = os.Stdout
			}
			if err := tt.client.UpdateBinary(tt.args.name, tt.args.version, tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("client.UpdateBinary() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPromptToDownload(t *testing.T) {
	req := require.New(t)

	clock := clockwork.NewFakeClockAt(time.Now())
	countRepeated := 0
	countNoConfirm := 0
	countNoPrompt := 0

	makeFS := func(terminal bool, input string) pio.FileSystem {
		return &mock.PassThroughFileSystem{
			Mock: &mock.FileSystem{
				IsTerminalFunc: func(fd uintptr) bool {
					return terminal
				},
				NewBufferedReaderFunc: func(rd io.Reader) pio.Reader {
					req.Equal(os.Stdin, rd)
					_, _ = fmt.Println() // to go to newline after test prompt
					return bytes.NewBuffer([]byte(input + "\n"))
				},
			},
			FS: &pio.RealFileSystem{},
		}
	}

	makeClient := func(fs pio.FileSystem) *client {
		client := NewClient(&ClientParams{
			Repository: &updateMock.Repository{},
			Logger:     log.New(),
		})
		client.clock = clock
		client.fs = fs
		return client
	}

	type args struct {
		name          string
		currVersion   string
		latestVersion string
		confirm       bool
	}

	basicArgs := args{
		name:          "my-cli",
		currVersion:   "v1.2.0",
		latestVersion: "v2.0.0",
		confirm:       true,
	}

	tests := []struct {
		name   string
		client *client
		args   args
		want   bool
	}{
		{
			name:   "should prompt interactively and return true for yes",
			client: makeClient(makeFS(true, "yes")),
			args:   basicArgs,
			want:   true,
		},
		{
			name:   "should prompt interactively and return true for y",
			client: makeClient(makeFS(true, "y")),
			args:   basicArgs,
			want:   true,
		},
		{
			name:   "should prompt interactively and return true for Y",
			client: makeClient(makeFS(true, "Y")),
			args:   basicArgs,
			want:   true,
		},
		{
			name:   "should prompt interactively and return false for no",
			client: makeClient(makeFS(true, "no")),
			args:   basicArgs,
			want:   false,
		},
		{
			name:   "should prompt interactively and return false for n",
			client: makeClient(makeFS(true, "n")),
			args:   basicArgs,
			want:   false,
		},
		{
			name:   "should prompt interactively and return false for N",
			client: makeClient(makeFS(true, "N")),
			args:   basicArgs,
			want:   false,
		},
		{
			name:   "should prompt interactively and ignore trailing whitespace",
			client: makeClient(makeFS(true, "y ")),
			args:   basicArgs,
			want:   true,
		},
		{
			name: "should prompt repeatedly until user enters yes/no",
			client: makeClient(&mock.PassThroughFileSystem{
				Mock: &mock.FileSystem{
					IsTerminalFunc: func(fd uintptr) bool {
						return true
					},
					NewBufferedReaderFunc: func(rd io.Reader) pio.Reader {
						req.Equal(os.Stdin, rd)
						_, _ = fmt.Println() // to go to newline after test prompt
						countRepeated++
						switch countRepeated {
						case 1:
							return bytes.NewBuffer([]byte("maybe"))
						case 2:
							return bytes.NewBuffer([]byte("youwish"))
						case 3:
							return bytes.NewBuffer([]byte("YES"))
						case 4:
							return bytes.NewBuffer([]byte("never"))
						case 5:
							return bytes.NewBuffer([]byte("no"))
						}
						return bytes.NewBuffer([]byte("n"))
					},
				},
				FS: &pio.RealFileSystem{},
			}),
			args: basicArgs,
			want: false,
		},
		{
			name: "should skip confirmation if not requested",
			client: makeClient(&mock.PassThroughFileSystem{
				Mock: &mock.FileSystem{
					IsTerminalFunc: func(fd uintptr) bool {
						return true
					},
					NewBufferedReaderFunc: func(rd io.Reader) pio.Reader {
						countNoConfirm++
						return bytes.NewBuffer([]byte("n"))
					},
				},
				FS: &pio.RealFileSystem{},
			}),
			args: args{
				name:          "my-cli",
				currVersion:   "v1.2.0",
				latestVersion: "v2.0.0",
				confirm:       false,
			},
			want: true,
		},
		{
			name: "should skip confirmation if not a TTY",
			client: makeClient(&mock.PassThroughFileSystem{
				Mock: &mock.FileSystem{
					IsTerminalFunc: func(fd uintptr) bool {
						return false
					},
					NewBufferedReaderFunc: func(rd io.Reader) pio.Reader {
						countNoPrompt++
						return bytes.NewBuffer([]byte("n"))
					},
				},
				FS: &pio.RealFileSystem{},
			}),
			args: args{
				name:          "my-cli",
				currVersion:   "v1.2.0",
				latestVersion: "v2.0.0",
				confirm:       false,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.client.Out == nil {
				tt.client.Out = os.Stdout
			}
			if got := tt.client.PromptToDownload(tt.args.name, tt.args.currVersion, tt.args.latestVersion, "", tt.args.confirm); got != tt.want {
				t.Errorf("client.PromptToDownload() = %v, want %v", got, tt.want)
			}
		})
	}
	req.Equal(5, countRepeated)
	req.Equal(0, countNoConfirm)
	req.Equal(0, countNoPrompt)
}
