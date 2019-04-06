package s3

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/log"
)

func NewMockPublicS3(prefix, response string, req *require.Assertions) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		req.Equal("/", r.URL.Path)
		req.Equal(1, len(r.URL.Query()))
		req.Equal(prefix+"/", r.URL.Query()["prefix"][0])
		_, _ = io.WriteString(w, response)
	})
	return httptest.NewServer(mux)
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
		Logger      *log.Logger
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
				Logger:   logger,
				Endpoint: NewMockPublicS3("ccloud-cli", ListVersionsPublicFixture, req).URL,
			},
			args: args{
				name: "ccloud",
			},
			want: makeVersions("0.47.0", "0.48.0"),
		},
		{
			name: "excludes files that don't match our naming standards",
			fields: fields{
				Logger:   logger,
				Endpoint: NewMockPublicS3("ccloud-cli", ListVersionsPublicFixtureInvalidNames, req).URL,
			},
			args: args{
				name: "confluent",
			},
			wantErr: true,
		},
		{
			name: "excludes files that aren't prefixed correctly",
			fields: fields{
				Logger:      logger,
				Endpoint:    NewMockPublicS3("confluent", ListVersionsPublicFixtureInvalidPrefix, req).URL,
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
				Logger:   logger,
				Endpoint: NewMockPublicS3("ccloud-cli", ListVersionsPublicFixtureOtherBinaries, req).URL,
			},
			args: args{
				name: "ccloud",
			},
			want: makeVersions("0.42.0"),
		},
		{
			name: "excludes binaries with dirty or SNAPSHOT versions",
			fields: fields{
				Logger:   logger,
				Endpoint: NewMockPublicS3("ccloud-cli", ListVersionsPublicFixtureDirtyVersions, req).URL,
			},
			args: args{
				name: "confluent",
			},
			want: makeVersions("0.44.0"),
		},
		{
			name: "sorts by version",
			fields: fields{
				Logger:   logger,
				Endpoint: NewMockPublicS3("ccloud-cli", ListVersionsPublicFixtureUnsortedVersions, req).URL,
			},
			args: args{
				name: "confluent",
			},
			want: makeVersions("0.42.0", "0.43.0", "0.44.0"),
		},
		{
			name: "errors when no version available",
			fields: fields{
				Logger:   logger,
				Endpoint: NewMockPublicS3("ccloud-cli", ListVersionsPublicFixture, req).URL,
			},
			args: args{
				name: "confluent",
			},
			wantErr: true,
		},
		{
			name: "errors when non-semver version found",
			fields: fields{
				Logger:   logger,
				Endpoint: NewMockPublicS3("ccloud-cli", ListVersionsPublicFixtureNonSemver, req).URL,
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
				S3KeyParser: &VersionPrefixedKeyParser{
					Prefix: tt.fields.S3BinPrefix,
					Name:   tt.args.name,
					OS:     goos,
					Arch:   goarch,
				},
				Logger: tt.fields.Logger,
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
				t.Errorf("PublicRepo.GetAvailableVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}
