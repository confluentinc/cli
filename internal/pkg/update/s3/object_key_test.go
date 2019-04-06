package s3

import (
	"reflect"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/require"
)

func TestVersionPrefixedKey_ParseVersion(t *testing.T) {
	req := require.New(t)

	makeVersion := func(v string) *version.Version {
		ver, err := version.NewSemver(v)
		req.NoError(err)
		return ver
	}

	type fields struct {
		Prefix    string
		Name      string
		Separator string
	}
	type args struct {
		key string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantMatch bool
		wantVer   *version.Version
		wantErr   bool
	}{
		{
			name: "should parse version from key",
			fields: fields{
				Prefix: "pre",
				Name:   "fancy",
			},
			args: args{
				key: "pre/0.23.0/fancy_0.23.0_darwin_amd64",
			},
			wantMatch: true,
			wantVer:   makeVersion("0.23.0"),
		},
		{
			name: "should support configurable separators",
			fields: fields{
				Prefix:    "pre",
				Name:      "fancy",
				Separator: "-",
			},
			args: args{
				key: "pre/0.23.0/fancy-0.23.0-darwin-amd64",
			},
			wantMatch: true,
			wantVer:   makeVersion("0.23.0"),
		},
		{
			name: "should support v-prefixed versions",
			fields: fields{
				Prefix: "pre",
				Name:   "fancy",
			},
			args: args{
				key: "pre/v0.23.0/fancy_v0.23.0_darwin_amd64",
			},
			wantMatch: true,
			wantVer:   makeVersion("v0.23.0"),
		},
		{
			name: "should not match if versions are different",
			fields: fields{
				Prefix: "pre",
				Name:   "fancy",
			},
			args: args{
				key: "pre/0.23.0/fancy_0.24.0_darwin_amd64",
			},
			wantMatch: false,
		},
		{
			name: "should not match if prefix contains the separator",
			fields: fields{
				Prefix:    "my-pre",
				Name:      "fancy",
				Separator: "-",
			},
			args: args{
				key: "my-pre/0.23.0/fancy-0.23.0-darwin-amd64",
			},
			wantMatch: false,
		},
		{
			name: "should not match if name contains the separator",
			fields: fields{
				Prefix:    "pre",
				Name:      "fancy-cli",
				Separator: "-",
			},
			args: args{
				key: "pre/0.23.0/fancy-cli-0.23.0-darwin-amd64",
			},
			wantMatch: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewVersionPrefixedKey(tt.fields.Prefix, tt.fields.Name, tt.fields.Separator)
			// Need to inject these so tests pass in different environments (e.g., CI)
			p.goos = "darwin"
			p.goarch = "amd64"
			match, ver, err := p.ParseVersion(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("VersionPrefixedKey.ParseVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if match != tt.wantMatch {
				t.Errorf("VersionPrefixedKey.ParseVersion() match = %v, wantMatch %v", match, tt.wantMatch)
			}
			if !reflect.DeepEqual(ver, tt.wantVer) {
				t.Errorf("VersionPrefixedKey.ParseVersion() ver = %v, wantVer %v", ver, tt.wantVer)
			}
		})
	}
}
