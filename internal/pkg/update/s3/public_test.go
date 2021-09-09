package s3

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/errors"
	pio "github.com/confluentinc/cli/internal/pkg/io"
	"github.com/confluentinc/cli/internal/pkg/log"
	pmock "github.com/confluentinc/cli/internal/pkg/mock"
)

func NewMockPublicS3(response, path, query string, req *require.Assertions) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		req.Equal(path, r.URL.Path)
		req.Equal(query, r.URL.RawQuery)
		_, _ = io.WriteString(w, response)
	})
	return httptest.NewServer(mux)
}

func NewMockPublicS3Error() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	})
	return httptest.NewServer(mux)
}

func TestPublicRepo_GetAvailableBinaryVersions(t *testing.T) {
	req := require.New(t)
	logger := log.New()

	makeVersions := func(versions ...string) version.Collection {
		col := version.Collection{}
		for _, v := range versions {
			ver, err := version.NewSemver(v)
			req.NoError(err)
			col = append(col, ver)
		}
		return col
	}

	type fields struct {
		Endpoint string
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    version.Collection
		wantErr bool
	}{
		{
			name:   "can get available versions for requested package and current os/arch",
			fields: fields{Endpoint: NewMockPublicS3(ListVersionsPublicFixture, "/", "prefix=ccloud-cli/", req).URL},
			args:   args{name: "ccloud"},
			want:   makeVersions("0.47.0", "0.48.0"),
		},
		{
			name:    "excludes files that don't match our naming standards",
			fields:  fields{Endpoint: NewMockPublicS3(ListVersionsPublicFixtureInvalidNames, "/", "prefix=ccloud-cli/", req).URL},
			args:    args{name: "ccloud"},
			wantErr: true,
		},
		{
			name:   "excludes other binaries in the same bucket/path",
			fields: fields{Endpoint: NewMockPublicS3(ListVersionsPublicFixtureOtherBinaries, "/", "prefix=ccloud-cli/", req).URL},
			args:   args{name: "ccloud"},
			want:   makeVersions("0.42.0"),
		},
		{
			name:   "excludes binaries with dirty or SNAPSHOT versions",
			fields: fields{Endpoint: NewMockPublicS3(ListVersionsPublicFixtureDirtyVersions, "/", "prefix=ccloud-cli/", req).URL},
			args:   args{name: "ccloud"},
			want:   makeVersions("0.44.0"),
		},
		{
			name:   "sorts by version",
			fields: fields{Endpoint: NewMockPublicS3(ListVersionsPublicFixtureUnsortedVersions, "/", "prefix=ccloud-cli/", req).URL},
			args:   args{name: "ccloud"},
			want:   makeVersions("0.42.0", "0.42.1", "0.43.0"),
		},
		{
			name:    "errors when non-semver version found",
			fields:  fields{Endpoint: NewMockPublicS3(ListVersionsPublicFixtureNonSemver, "/", "prefix=ccloud-cli/", req).URL},
			args:    args{name: "ccloud"},
			wantErr: true,
		},
		{
			name:    "errors when S3 returns non-200 response",
			fields:  fields{Endpoint: NewMockPublicS3Error().URL},
			args:    args{name: "confluent"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Need to inject these so tests pass in different environments (e.g., CI)
			goos := "darwin"
			goarch := "amd64"
			r := NewPublicRepo(&PublicRepoParams{
				S3BinPrefixFmt: "%s-cli",
				Logger:         logger,
			})
			r.endpoint = tt.fields.Endpoint
			r.goos = goos
			r.goarch = goarch

			got, err := r.GetAvailableBinaryVersions(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("PublicRepo.GetAvailableBinaryVersions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PublicRepo.GetAvailableBinaryVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPublicRepo_GetLatestMajorAndMinorVersion(t *testing.T) {
	req := require.New(t)
	logger := log.New()

	makeVersion := func(v string) *version.Version {
		ver, err := version.NewSemver(v)
		req.NoError(err)
		return ver
	}

	type fields struct {
		Endpoint string
	}
	type args struct {
		name string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantMajor *version.Version
		wantMinor *version.Version
		wantErr   bool
	}{
		{
			name:      "can get available versions for requested package and current os/arch",
			fields:    fields{Endpoint: NewMockPublicS3(ListVersionsPublicFixture, "/", "prefix=ccloud-cli/", req).URL},
			args:      args{name: "ccloud"},
			wantMinor: makeVersion("0.48.0"),
		},
		{
			name:      "sorts by version",
			fields:    fields{Endpoint: NewMockPublicS3(ListVersionsPublicFixtureUnsortedVersions, "/", "prefix=ccloud-cli/", req).URL},
			args:      args{name: "ccloud"},
			wantMinor: makeVersion("0.43.0"),
		},
		{
			name:    "errors when S3 returns non-200 response",
			fields:  fields{Endpoint: NewMockPublicS3Error().URL},
			args:    args{name: "confluent"},
			wantErr: true,
		},
		{
			name:      "different major and minor versions",
			fields:    fields{Endpoint: NewMockPublicS3(ListVersionsPublicFixtureMajorAndMinor, "/", "prefix=ccloud-cli/", req).URL},
			args:      args{name: "ccloud"},
			wantMajor: makeVersion("1.0.0"),
			wantMinor: makeVersion("0.1.0"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Need to inject these so tests pass in different environments (e.g., CI)
			goos := "darwin"
			goarch := "amd64"
			r := NewPublicRepo(&PublicRepoParams{
				S3BinPrefixFmt: "%s-cli",
				Logger:         logger,
			})
			r.endpoint = tt.fields.Endpoint
			r.goos = goos
			r.goarch = goarch

			v, _ := version.NewVersion("v0.0.0")
			latestMajorVersion, latestMinorVersion, err := r.GetLatestMajorAndMinorVersion(tt.args.name, v)
			if (err != nil) != tt.wantErr {
				t.Errorf("PublicRepo.GetLatestMajorAndMinorVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(latestMajorVersion, tt.wantMajor) {
				t.Errorf("PublicRepo.GetLatestMajorAndMinorVersion() majorVersion = %v, want %v", latestMajorVersion, tt.wantMajor)
			}
			if !reflect.DeepEqual(latestMinorVersion, tt.wantMinor) {
				t.Errorf("PublicRepo.GetLatestMajorAndMinorVersion() minorVersion = %v, want %v", latestMinorVersion, tt.wantMinor)
			}
		})
	}
}

func TestPublicRepo_GetAvailableReleaseNotesVersions(t *testing.T) {
	req := require.New(t)

	makeVersions := func(versions ...string) version.Collection {
		col := version.Collection{}
		for _, v := range versions {
			ver, err := version.NewSemver(v)
			req.NoError(err)
			col = append(col, ver)
		}
		return col
	}

	type fields struct {
		S3BinBucket string
		S3BinRegion string
		CLIName     string
		Endpoint    string
	}
	tests := []struct {
		name    string
		fields  fields
		want    version.Collection
		wantErr bool
	}{
		{
			name: "can get available versions for requested release notes",
			fields: fields{
				Endpoint: NewMockPublicS3(ListReleaseNotesVersionsPublicFixture, "/", "prefix=ccloud-cli/release-notes/", req).URL,
				CLIName:  "ccloud",
			},
			want: makeVersions("0.0.0", "0.1.0"),
		},
		{
			name: "sorts by version",
			fields: fields{
				Endpoint: NewMockPublicS3(ListReleaseNotesVersionsPublicFixtureUnsortedVersions, "/", "prefix=ccloud-cli/release-notes/", req).URL,
				CLIName:  "ccloud",
			},
			want: makeVersions("0.42.0", "0.42.1", "0.43.0"),
		},
		{
			name: "invalid file names",
			fields: fields{
				Endpoint: NewMockPublicS3(ListReleaseNotesVersionsInvalidFiles, "/", "prefix=ccloud-cli/release-notes/", req).URL,
				CLIName:  "ccloud",
			},
			wantErr: true,
		},
		{
			name: "include only valid file names",
			fields: fields{
				Endpoint: NewMockPublicS3(ListReleaseNotesVersionsExcludeInvalidFiles, "/", "prefix=ccloud-cli/release-notes/", req).URL,
				CLIName:  "ccloud",
			},
			want: makeVersions("0.47.0"),
		},
		{
			name: "error when no files available",
			fields: fields{
				Endpoint: NewMockPublicS3(ListReleaseNotesVersionsPublicFixture, "/", "prefix=confluent-cli/release-notes/", req).URL,
				CLIName:  "confluent",
			},
			wantErr: true,
		},
		{
			name: "ignore non-semver version",
			fields: fields{
				Endpoint: NewMockPublicS3(ListReleaseNotesVersionsPublicFixtureNonSemver, "/", "prefix=ccloud-cli/release-notes/", req).URL,
				CLIName:  "ccloud",
			},
			wantErr: true,
		},
		{
			name: "errors when S3 returns non-200 response",
			fields: fields{
				Endpoint: NewMockPublicS3Error().URL,
				CLIName:  "ccloud",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewPublicRepo(&PublicRepoParams{
				S3BinBucket:             tt.fields.S3BinBucket,
				S3BinRegion:             tt.fields.S3BinRegion,
				S3ReleaseNotesPrefixFmt: "%s-cli/release-notes",
				Logger:                  log.New(),
			})
			r.endpoint = tt.fields.Endpoint

			got, err := r.GetAvailableReleaseNotesVersions(tt.fields.CLIName)
			if (err != nil) != tt.wantErr {
				t.Errorf("PublicRepo.GetAvailableReleaseNotesVersions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PublicRepo.GetAvailableReleaseNotesVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPublicRepo_GetLatestReleaseNotesVersion(t *testing.T) {
	req := require.New(t)

	currentVersion := "v0.0.0"

	makeVersions := func(versions ...string) version.Collection {
		col := version.Collection{}
		for _, v := range versions {
			ver, err := version.NewSemver(v)
			req.NoError(err)
			col = append(col, ver)
		}
		return col
	}

	type fields struct {
		S3BinBucket string
		S3BinRegion string
		CLIName     string
		Endpoint    string
	}
	tests := []struct {
		name    string
		fields  fields
		want    version.Collection
		wantErr bool
	}{
		{
			name: "can get available versions for requested release notes",
			fields: fields{
				Endpoint: NewMockPublicS3(ListReleaseNotesVersionsPublicFixture, "/", "prefix=ccloud-cli/release-notes/", req).URL,
				CLIName:  "ccloud",
			},
			want: makeVersions("0.1.0"),
		},
		{
			name: "sorts by version",
			fields: fields{
				Endpoint: NewMockPublicS3(ListReleaseNotesVersionsPublicFixtureUnsortedVersions, "/", "prefix=ccloud-cli/release-notes/", req).URL,
				CLIName:  "ccloud",
			},
			want: makeVersions("0.42.0", "0.42.1", "0.43.0"),
		},
		{
			name: "invalid file names",
			fields: fields{
				Endpoint: NewMockPublicS3(ListReleaseNotesVersionsInvalidFiles, "/", "prefix=ccloud-cli/release-notes/", req).URL,
				CLIName:  "ccloud",
			},
			wantErr: true,
		},
		{
			name: "include only valid file names",
			fields: fields{
				Endpoint: NewMockPublicS3(ListReleaseNotesVersionsExcludeInvalidFiles, "/", "prefix=ccloud-cli/release-notes/", req).URL,
				CLIName:  "ccloud",
			},
			want: makeVersions("0.47.0"),
		},
		{
			name: "error when no files available",
			fields: fields{
				Endpoint: NewMockPublicS3(ListReleaseNotesVersionsPublicFixture, "/", "prefix=confluent-cli/release-notes/", req).URL,
				CLIName:  "confluent",
			},
			wantErr: true,
		},
		{
			name: "ignore non-semver version",
			fields: fields{
				Endpoint: NewMockPublicS3(ListReleaseNotesVersionsPublicFixtureNonSemver, "/", "prefix=ccloud-cli/release-notes/", req).URL,
				CLIName:  "ccloud",
			},
			wantErr: true,
		},
		{
			name: "errors when S3 returns non-200 response",
			fields: fields{
				Endpoint: NewMockPublicS3Error().URL,
				CLIName:  "ccloud",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewPublicRepo(&PublicRepoParams{
				S3BinBucket:             tt.fields.S3BinBucket,
				S3BinRegion:             tt.fields.S3BinRegion,
				S3ReleaseNotesPrefixFmt: "%s-cli/release-notes",
				Logger:                  log.New(),
			})
			r.endpoint = tt.fields.Endpoint

			got, err := r.GetLatestReleaseNotesVersions(tt.fields.CLIName, currentVersion)
			if tt.wantErr {
				req.Error(err)
			} else {
				req.NoError(err)
			}

			req.Equal(tt.want, got)
		})
	}
}

func TestPublicRepo_DownloadVersion(t *testing.T) {
	req := require.New(t)

	downloadDir, err := ioutil.TempDir("", "cli-test5-")
	require.NoError(t, err)
	defer os.Remove(downloadDir)

	type fields struct {
		Endpoint   string
		FileSystem pio.FileSystem
	}
	type args struct {
		name        string
		version     string
		downloadDir string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantPath  string
		wantBytes int64
		wantErr   bool
	}{
		{
			name: "should err if unable to download",
			fields: fields{
				Endpoint: NewMockPublicS3Error().URL,
			},
			wantErr: true,
		},
		{
			name: "should err if unable to open/create file at path",
			fields: fields{
				Endpoint: NewMockPublicS3(ListVersionsPublicFixture,
					"/ccloud-cli/0.47.0/ccloud_0.47.0_darwin_amd64", "", req).URL,
				FileSystem: &pmock.PassThroughFileSystem{
					Mock: &pmock.FileSystem{
						CopyFunc: func(dst io.Writer, src io.Reader) (i int64, e error) {
							return 0, errors.New("you no can do that")
						},
					},
					FS: &pio.RealFileSystem{},
				},
			},
			args: args{
				name:        "ccloud",
				version:     "0.47.0",
				downloadDir: downloadDir,
			},
			wantErr: true,
		},
		{
			name: "should err if unable to write/copy file to path",
			fields: fields{
				Endpoint: NewMockPublicS3(ListVersionsPublicFixture,
					"/ccloud-cli/0.47.0/ccloud_0.47.0_darwin_amd64", "", req).URL,
				FileSystem: &pmock.PassThroughFileSystem{
					Mock: &pmock.FileSystem{
						CreateFunc: func(name string) (pio.File, error) {
							return nil, errors.New("you no can do that")
						},
					},
					FS: &pio.RealFileSystem{},
				},
			},
			args: args{
				name:        "ccloud",
				version:     "0.47.0",
				downloadDir: downloadDir,
			},
			wantErr: true,
		},
		{
			name: "should download version",
			fields: fields{
				Endpoint: NewMockPublicS3(ListVersionsPublicFixture,
					"/ccloud-cli/0.47.0/ccloud_0.47.0_darwin_amd64", "", req).URL,
			},
			args: args{
				name:        "ccloud",
				version:     "0.47.0",
				downloadDir: downloadDir,
			},
			wantPath:  "ccloud-v0.47.0-darwin-amd64",
			wantBytes: 3840,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Need to inject these so tests pass in different environments (e.g., CI)
			goos := "darwin"
			goarch := "amd64"
			r := NewPublicRepo(&PublicRepoParams{
				S3BinPrefixFmt: "%s-cli",
				Logger:         log.New(),
			})
			r.endpoint = tt.fields.Endpoint
			r.goos = goos
			r.goarch = goarch
			if tt.fields.FileSystem != nil {
				r.fs = tt.fields.FileSystem
			}

			downloadPath, downloadedBytes, err := r.DownloadVersion(tt.args.name, tt.args.version, tt.args.downloadDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("PublicRepo.DownloadVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !strings.HasSuffix(downloadPath, tt.wantPath) {
				t.Errorf("PublicRepo.DownloadVersion() downloadPath = %v, wantPath %v", downloadPath, tt.wantPath)
			}
			if downloadedBytes != tt.wantBytes {
				t.Errorf("PublicRepo.DownloadVersion() downloadedBytes = %v, wantBytes %v", downloadedBytes, tt.wantBytes)
			}
		})
	}
}

func TestPublicRepo_DownloadReleaseNotes(t *testing.T) {
	req := require.New(t)

	downloadDir, err := ioutil.TempDir("", "cli-test5-")
	require.NoError(t, err)
	defer os.Remove(downloadDir)

	type fields struct {
		S3BinBucket string
		S3BinRegion string
		Endpoint    string
	}
	type args struct {
		name    string
		version string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    string
	}{
		{
			name: "should err if unable to download",
			fields: fields{
				Endpoint: NewMockPublicS3Error().URL,
			},
			wantErr: true,
		},
		{
			name: "should download release notes",
			fields: fields{
				Endpoint: NewMockPublicS3(ReleaseNotesFileV0470,
					"/ccloud-cli/release-notes/0.47.0/release-notes.rst", "", req).URL,
			},
			args: args{
				name:    "ccloud",
				version: "0.47.0",
			},
			want: ReleaseNotesFileV0470,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewPublicRepo(&PublicRepoParams{
				S3BinBucket:             tt.fields.S3BinBucket,
				S3BinRegion:             tt.fields.S3BinRegion,
				S3ReleaseNotesPrefixFmt: "/%s-cli/release-notes",
				Logger:                  log.New(),
			})
			r.endpoint = tt.fields.Endpoint

			releaseNotes, err := r.DownloadReleaseNotes(tt.args.name, tt.args.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("PublicRepo.DownloadVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !strings.HasSuffix(releaseNotes, tt.want) {
				t.Errorf("PublicRepo.DownloadVersion() download = %v, want %v", releaseNotes, tt.want)
			}
		})
	}
}
