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
	"github.com/confluentinc/cli/internal/pkg/log"
	pio "github.com/confluentinc/cli/internal/pkg/update/io"
	"github.com/confluentinc/cli/internal/pkg/update/mock"
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

func NewTestVersionPrefixedKeyParser(prefix, name, goos, goarch string, req *require.Assertions) *PrefixedKey {
	p, err := NewPrefixedKey(prefix, "_", true)
	req.NoError(err)
	p.goos = goos
	p.goarch = goarch
	return p
}

func TestPublicRepo_GetAvailableVersions(t *testing.T) {
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
		S3BinBucket string
		S3BinRegion string
		S3BinPrefix string
		Endpoint    string
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
			name: "can get available versions for requested package and current os/arch",
			fields: fields{
				Endpoint: NewMockPublicS3(ListVersionsPublicFixture, "/", "prefix=ccloud-cli/", req).URL,
			},
			args: args{
				name: "ccloud",
			},
			want: makeVersions("0.47.0", "0.48.0"),
		},
		{
			name: "excludes files that don't match our naming standards",
			fields: fields{
				Endpoint: NewMockPublicS3(ListVersionsPublicFixtureInvalidNames, "/", "prefix=ccloud-cli/", req).URL,
			},
			args: args{
				name: "confluent",
			},
			wantErr: true,
		},
		{
			name: "excludes files that aren't prefixed correctly",
			fields: fields{
				Endpoint:    NewMockPublicS3(ListVersionsPublicFixtureInvalidPrefix, "/", "prefix=confluent/", req).URL,
				S3BinPrefix: "confluent",
			},
			args: args{
				name: "confluent",
			},
			wantErr: true,
		},
		{
			name: "excludes other binaries in the same bucket/path",
			fields: fields{
				Endpoint: NewMockPublicS3(ListVersionsPublicFixtureOtherBinaries, "/", "prefix=ccloud-cli/", req).URL,
			},
			args: args{
				name: "ccloud",
			},
			want: makeVersions("0.42.0"),
		},
		{
			name: "excludes binaries with dirty or SNAPSHOT versions",
			fields: fields{
				Endpoint: NewMockPublicS3(ListVersionsPublicFixtureDirtyVersions, "/", "prefix=ccloud-cli/", req).URL,
			},
			args: args{
				name: "confluent",
			},
			want: makeVersions("0.44.0"),
		},
		{
			name: "sorts by version",
			fields: fields{
				Endpoint: NewMockPublicS3(ListVersionsPublicFixtureUnsortedVersions, "/", "prefix=ccloud-cli/", req).URL,
			},
			args: args{
				name: "confluent",
			},
			want: makeVersions("0.42.0", "0.43.0", "0.44.0"),
		},
		{
			name: "errors when no version available",
			fields: fields{
				Endpoint: NewMockPublicS3(ListVersionsPublicFixture, "/", "prefix=ccloud-cli/", req).URL,
			},
			args: args{
				name: "confluent",
			},
			wantErr: true,
		},
		{
			name: "errors when non-semver version found",
			fields: fields{
				Endpoint: NewMockPublicS3(ListVersionsPublicFixtureNonSemver, "/", "prefix=ccloud-cli/", req).URL,
			},
			args: args{
				name: "confluent",
			},
			wantErr: true,
		},
		{
			name: "errors when S3 returns non-200 response",
			fields: fields{
				Endpoint: NewMockPublicS3Error().URL,
			},
			args: args{
				name: "confluent",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fields.S3BinPrefix == "" {
				tt.fields.S3BinPrefix = "ccloud-cli"
			}
			// Need to inject these so tests pass in different environments (e.g., CI)
			goos := "darwin"
			goarch := "amd64"
			r := NewPublicRepo(&PublicRepoParams{
				S3BinBucket: tt.fields.S3BinBucket,
				S3BinRegion: tt.fields.S3BinRegion,
				S3BinPrefix: tt.fields.S3BinPrefix,
				S3ObjectKey: NewTestVersionPrefixedKeyParser(tt.fields.S3BinPrefix, tt.args.name, goos, goarch, req),
				Logger:      logger,
			})
			r.endpoint = tt.fields.Endpoint
			r.goos = goos
			r.goarch = goarch

			got, err := r.GetAvailableVersions(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("PublicRepo.GetAvailableVersions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PublicRepo.GetAvailableVersions() = %v, wantPath %v", got, tt.want)
			}
		})
	}
}

func TestPublicRepo_DownloadVersion(t *testing.T) {
	req := require.New(t)

	downloadDir, err := ioutil.TempDir("", "cli-test5-")
	require.NoError(t, err)
	defer os.Remove(downloadDir)

	type fields struct {
		S3BinBucket string
		S3BinRegion string
		S3BinPrefix string
		Endpoint    string
		FileSystem  pio.FileSystem
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
				FileSystem: &mock.PassThroughFileSystem{
					Mock: &mock.FileSystem{
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
				FileSystem: &mock.PassThroughFileSystem{
					Mock: &mock.FileSystem{
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
			if tt.fields.S3BinPrefix == "" {
				tt.fields.S3BinPrefix = "ccloud-cli"
			}
			// Need to inject these so tests pass in different environments (e.g., CI)
			goos := "darwin"
			goarch := "amd64"
			r := NewPublicRepo(&PublicRepoParams{
				S3BinBucket: tt.fields.S3BinBucket,
				S3BinRegion: tt.fields.S3BinRegion,
				S3BinPrefix: tt.fields.S3BinPrefix,
				S3ObjectKey: NewTestVersionPrefixedKeyParser(tt.fields.S3BinPrefix, tt.args.name, goos, goarch, req),
				Logger:      log.New(),
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
				t.Errorf("PublicRepo.DownloadVersion() downloadedBytes = %v, wantPath %v", downloadedBytes, tt.wantBytes)
			}
		})
	}
}
