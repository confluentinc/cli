package s3

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/require"

	pio "github.com/confluentinc/cli/v3/pkg/io"
	pversion "github.com/confluentinc/cli/v3/pkg/version"
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
		w.WriteHeader(http.StatusInternalServerError)
	})
	return httptest.NewServer(mux)
}

func TestPublicRepo_GetAvailableBinaryVersions(t *testing.T) {
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
			fields: fields{Endpoint: NewMockPublicS3(ListVersionsPublicFixture, "/", "prefix=confluent-cli/", req).URL},
			args:   args{name: "confluent"},
			want:   makeVersions("0.47.0", "0.48.0"),
		},
		{
			name:    "excludes files that don't match our naming standards",
			fields:  fields{Endpoint: NewMockPublicS3(ListVersionsPublicFixtureInvalidNames, "/", "prefix=confluent-cli/", req).URL},
			args:    args{name: "confluent"},
			wantErr: true,
		},
		{
			name:   "excludes other binaries in the same bucket/path",
			fields: fields{Endpoint: NewMockPublicS3(ListVersionsPublicFixtureOtherBinaries, "/", "prefix=confluent-cli/", req).URL},
			args:   args{name: "confluent"},
			want:   makeVersions("0.42.0"),
		},
		{
			name:   "excludes binaries with dirty or SNAPSHOT versions",
			fields: fields{Endpoint: NewMockPublicS3(ListVersionsPublicFixtureDirtyVersions, "/", "prefix=confluent-cli/", req).URL},
			args:   args{name: "confluent"},
			want:   makeVersions("0.44.0"),
		},
		{
			name:   "sorts by version",
			fields: fields{Endpoint: NewMockPublicS3(ListVersionsPublicFixtureUnsortedVersions, "/", "prefix=confluent-cli/", req).URL},
			args:   args{name: "confluent"},
			want:   makeVersions("0.42.0", "0.42.1", "0.43.0"),
		},
		{
			name:    "errors when non-semver version found",
			fields:  fields{Endpoint: NewMockPublicS3(ListVersionsPublicFixtureNonSemver, "/", "prefix=confluent-cli/", req).URL},
			args:    args{name: "confluent"},
			wantErr: true,
		},
		{
			name:    "errors when S3 returns non-200 response",
			fields:  fields{Endpoint: NewMockPublicS3Error().URL},
			args:    args{name: "confluent"},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Need to inject these so tests pass in different environments (e.g., CI)
			goos := "darwin"
			goarch := "amd64"
			r := NewPublicRepo(&PublicRepoParams{
				S3BinPrefixFmt: "%s-cli",
			})
			r.endpoint = test.fields.Endpoint
			r.goos = goos
			r.goarch = goarch

			got, err := r.GetAvailableBinaryVersions(test.args.name)
			if (err != nil) != test.wantErr {
				t.Errorf("PublicRepo.GetAvailableBinaryVersions() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("PublicRepo.GetAvailableBinaryVersions() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestPublicRepo_GetLatestMajorAndMinorVersion(t *testing.T) {
	req := require.New(t)

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
			fields:    fields{Endpoint: NewMockPublicS3(ListVersionsPublicFixture, "/", "prefix=confluent-cli/", req).URL},
			args:      args{name: "confluent"},
			wantMinor: makeVersion("0.48.0"),
		},
		{
			name:      "sorts by version",
			fields:    fields{Endpoint: NewMockPublicS3(ListVersionsPublicFixtureUnsortedVersions, "/", "prefix=confluent-cli/", req).URL},
			args:      args{name: "confluent"},
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
			fields:    fields{Endpoint: NewMockPublicS3(ListVersionsPublicFixtureMajorAndMinor, "/", "prefix=confluent-cli/", req).URL},
			args:      args{name: "confluent"},
			wantMajor: makeVersion("1.0.0"),
			wantMinor: makeVersion("0.1.0"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Need to inject these so tests pass in different environments (e.g., CI)
			goos := "darwin"
			goarch := "amd64"
			r := NewPublicRepo(&PublicRepoParams{
				S3BinPrefixFmt: "%s-cli",
			})
			r.endpoint = test.fields.Endpoint
			r.goos = goos
			r.goarch = goarch

			v, _ := version.NewVersion("v0.0.0")
			latestMajorVersion, latestMinorVersion, err := r.GetLatestMajorAndMinorVersion(test.args.name, v)
			if (err != nil) != test.wantErr {
				t.Errorf("PublicRepo.GetLatestMajorAndMinorVersion() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if !reflect.DeepEqual(latestMajorVersion, test.wantMajor) {
				t.Errorf("PublicRepo.GetLatestMajorAndMinorVersion() majorVersion = %v, want %v", latestMajorVersion, test.wantMajor)
			}
			if !reflect.DeepEqual(latestMinorVersion, test.wantMinor) {
				t.Errorf("PublicRepo.GetLatestMajorAndMinorVersion() minorVersion = %v, want %v", latestMinorVersion, test.wantMinor)
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
				Endpoint: NewMockPublicS3(ListReleaseNotesVersionsPublicFixture, "/", "prefix=confluent-cli/release-notes/", req).URL,
			},
			want: makeVersions("0.0.0", "0.1.0"),
		},
		{
			name: "sorts by version",
			fields: fields{
				Endpoint: NewMockPublicS3(ListReleaseNotesVersionsPublicFixtureUnsortedVersions, "/", "prefix=confluent-cli/release-notes/", req).URL,
			},
			want: makeVersions("0.42.0", "0.42.1", "0.43.0"),
		},
		{
			name: "invalid file names",
			fields: fields{
				Endpoint: NewMockPublicS3(ListReleaseNotesVersionsInvalidFiles, "/", "prefix=confluent-cli/release-notes/", req).URL,
			},
			wantErr: true,
		},
		{
			name: "include only valid file names",
			fields: fields{
				Endpoint: NewMockPublicS3(ListReleaseNotesVersionsExcludeInvalidFiles, "/", "prefix=confluent-cli/release-notes/", req).URL,
			},
			want: makeVersions("0.47.0"),
		},
		{
			name: "ignore non-semver version",
			fields: fields{
				Endpoint: NewMockPublicS3(ListReleaseNotesVersionsPublicFixtureNonSemver, "/", "prefix=confluent-cli/release-notes/", req).URL,
			},
			wantErr: true,
		},
		{
			name: "errors when S3 returns non-200 response",
			fields: fields{
				Endpoint: NewMockPublicS3Error().URL,
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := NewPublicRepo(&PublicRepoParams{
				S3BinBucket:             test.fields.S3BinBucket,
				S3BinRegion:             test.fields.S3BinRegion,
				S3ReleaseNotesPrefixFmt: "%s-cli/release-notes",
			})
			r.endpoint = test.fields.Endpoint

			got, err := r.GetAvailableReleaseNotesVersions(pversion.CLIName)
			if (err != nil) != test.wantErr {
				t.Errorf("PublicRepo.GetAvailableReleaseNotesVersions() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("PublicRepo.GetAvailableReleaseNotesVersions() = %v, want %v", got, test.want)
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
		Endpoint string
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
				Endpoint: NewMockPublicS3(ListReleaseNotesVersionsPublicFixture, "/", "prefix=confluent-cli/release-notes/", req).URL,
			},
			want: makeVersions("0.1.0"),
		},
		{
			name: "sorts by version",
			fields: fields{
				Endpoint: NewMockPublicS3(ListReleaseNotesVersionsPublicFixtureUnsortedVersions, "/", "prefix=confluent-cli/release-notes/", req).URL,
			},
			want: makeVersions("0.42.0", "0.42.1", "0.43.0"),
		},
		{
			name: "invalid file names",
			fields: fields{
				Endpoint: NewMockPublicS3(ListReleaseNotesVersionsInvalidFiles, "/", "prefix=confluent-cli/release-notes/", req).URL,
			},
			wantErr: true,
		},
		{
			name: "include only valid file names",
			fields: fields{
				Endpoint: NewMockPublicS3(ListReleaseNotesVersionsExcludeInvalidFiles, "/", "prefix=confluent-cli/release-notes/", req).URL,
			},
			want: makeVersions("0.47.0"),
		},
		{
			name: "ignore non-semver version",
			fields: fields{
				Endpoint: NewMockPublicS3(ListReleaseNotesVersionsPublicFixtureNonSemver, "/", "prefix=confluent-cli/release-notes/", req).URL,
			},
			wantErr: true,
		},
		{
			name: "errors when S3 returns non-200 response",
			fields: fields{
				Endpoint: NewMockPublicS3Error().URL,
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := NewPublicRepo(&PublicRepoParams{
				S3ReleaseNotesPrefixFmt: "%s-cli/release-notes",
			})
			r.endpoint = test.fields.Endpoint

			got, err := r.GetLatestReleaseNotesVersions(pversion.CLIName, currentVersion)
			if test.wantErr {
				req.Error(err)
			} else {
				req.NoError(err)
			}

			req.Equal(test.want, got)
		})
	}
}

func TestPublicRepo_DownloadVersion(t *testing.T) {
	req := require.New(t)

	downloadDir, err := os.MkdirTemp("", "cli-test5-")
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
		wantBytes int
		wantErr   bool
	}{
		{
			name:    "should err if unable to download",
			fields:  fields{Endpoint: NewMockPublicS3Error().URL},
			wantErr: true,
		},
		{
			name:   "should download version",
			fields: fields{Endpoint: NewMockPublicS3(ListVersionsPublicFixture, "/confluent-cli/0.47.0/confluent_0.47.0_darwin_amd64", "", req).URL},
			args: args{
				name:        "confluent",
				version:     "0.47.0",
				downloadDir: downloadDir,
			},
			wantPath:  "confluent-v0.47.0-darwin-amd64",
			wantBytes: 3921,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Need to inject these so tests pass in different environments (e.g., CI)
			goos := "darwin"
			goarch := "amd64"
			r := NewPublicRepo(&PublicRepoParams{S3BinPrefixFmt: "%s-cli"})
			r.endpoint = test.fields.Endpoint
			r.goos = goos
			r.goarch = goarch
			if test.fields.FileSystem != nil {
				r.fs = test.fields.FileSystem
			}

			payload, err := r.DownloadVersion(test.args.name, test.args.version)
			if (err != nil) != test.wantErr {
				t.Errorf("PublicRepo.DownloadVersion() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if len(payload) != test.wantBytes {
				t.Errorf("PublicRepo.DownloadVersion() downloadedBytes = %v, wantBytes %v", len(payload), test.wantBytes)
			}
		})
	}
}

func TestPublicRepo_DownloadReleaseNotes(t *testing.T) {
	req := require.New(t)

	downloadDir, err := os.MkdirTemp("", "cli-test5-")
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
				Endpoint: NewMockPublicS3(ReleaseNotesFileV0470, "/confluent-cli/release-notes/0.47.0/release-notes.rst", "", req).URL,
			},
			args: args{
				name:    "confluent",
				version: "0.47.0",
			},
			want: ReleaseNotesFileV0470,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := NewPublicRepo(&PublicRepoParams{
				S3BinBucket:             test.fields.S3BinBucket,
				S3BinRegion:             test.fields.S3BinRegion,
				S3ReleaseNotesPrefixFmt: "/%s-cli/release-notes",
			})
			r.endpoint = test.fields.Endpoint

			releaseNotes, err := r.DownloadReleaseNotes(test.args.name, test.args.version)
			if (err != nil) != test.wantErr {
				t.Errorf("PublicRepo.DownloadVersion() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if !strings.HasSuffix(releaseNotes, test.want) {
				t.Errorf("PublicRepo.DownloadVersion() download = %v, want %v", releaseNotes, test.want)
			}
		})
	}
}
