package s3

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/require"
)

func TestNewPrefixedKey(t *testing.T) {
	type args struct {
		prefix        string
		sep           string
		prefixVersion bool
	}
	tests := []struct {
		name    string
		args    args
		want    *PrefixedKey
		wantErr bool
	}{
		{
			name: "should return an error if sep is empty",
			args: args{
				prefix:        "pre",
				sep:           "",
				prefixVersion: false,
			},
			wantErr: true,
		},
		{
			name: "should return an error if sep is space",
			args: args{
				prefix:        "pre",
				sep:           " ",
				prefixVersion: false,
			},
			wantErr: true,
		},
		{
			name: "should return a PrefixedKey",
			args: args{
				prefix:        "",
				sep:           "_",
				prefixVersion: false,
			},
			want: &PrefixedKey{
				Prefix:        "",
				PrefixVersion: false,
				Separator:     "_",
				goos:          runtime.GOOS,
				goarch:        runtime.GOARCH,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPrefixedKey(tt.args.prefix, tt.args.sep, tt.args.prefixVersion)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPrefixedKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPrefixedKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrefixedKey_ParseVersion(t *testing.T) {
	req := require.New(t)

	makeVersion := func(v string) *version.Version {
		ver, err := version.NewSemver(v)
		req.NoError(err)
		return ver
	}

	type fields struct {
		Prefix    string
		Separator string
		Versioned bool
		goos      string
		goarch    string
	}
	type args struct {
		key  string
		name string
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
				Prefix:    "pre",
				Versioned: true,
			},
			args: args{
				key:  "pre/0.23.0/fancy_0.23.0_darwin_amd64",
				name: "fancy",
			},
			wantMatch: true,
			wantVer:   makeVersion("0.23.0"),
		},
		{
			name: "should support configurable separators",
			fields: fields{
				Prefix:    "pre",
				Separator: "-",
				Versioned: true,
			},
			args: args{
				key:  "pre/0.23.0/fancy-0.23.0-darwin-amd64",
				name: "fancy",
			},
			wantMatch: true,
			wantVer:   makeVersion("0.23.0"),
		},
		{
			name: "should support v-prefixed versions",
			fields: fields{
				Prefix:    "pre",
				Versioned: true,
			},
			args: args{
				key:  "pre/v0.23.0/fancy_v0.23.0_darwin_amd64",
				name: "fancy",
			},
			wantMatch: true,
			wantVer:   makeVersion("v0.23.0"),
		},
		{
			name: "should not match if versions are different",
			fields: fields{
				Prefix:    "pre",
				Versioned: true,
			},
			args: args{
				key:  "pre/0.23.0/fancy_0.24.0_darwin_amd64",
				name: "fancy",
			},
			wantMatch: false,
		},
		{
			name: "will not match if prefix contains the separator (TODO)",
			fields: fields{
				Prefix:    "my-pre",
				Separator: "-",
				Versioned: true,
			},
			args: args{
				key:  "my-pre/0.23.0/fancy-0.23.0-darwin-amd64",
				name: "fancy",
			},
			wantMatch: false,
		},
		{
			name: "will not match if name contains the separator (TODO)",
			fields: fields{
				Prefix:    "pre",
				Separator: "-",
				Versioned: true,
			},
			args: args{
				key:  "pre/0.23.0/fancy-cli-0.23.0-darwin-amd64",
				name: "fancy-cli",
			},
			wantMatch: false,
		},
		{
			name: "should support parsing without version prefix",
			fields: fields{
				Prefix:    "pre",
				Versioned: false,
			},
			args: args{
				key:  "pre/fancy_0.23.0_darwin_amd64",
				name: "fancy",
			},
			wantMatch: true,
			wantVer:   makeVersion("0.23.0"),
		},
		{
			name:   "should support empty prefix",
			fields: fields{},
			args: args{
				key: "fancy_0.23.0_darwin_amd64",
			},
			wantMatch: true,
			wantVer:   makeVersion("0.23.0"),
		},
		{
			name: "should require .exe for windows binaries",
			fields: fields{
				Prefix:    "pre",
				Versioned: true,
				goos:      "windows",
				goarch:    "386",
			},
			args: args{
				key: "pre/0.23.0/fancy_0.23.0_windows_386.exe",
			},
			wantMatch: true,
			wantVer:   makeVersion("0.23.0"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fields.Separator == "" {
				tt.fields.Separator = "_"
			}
			if tt.fields.goos == "" {
				tt.fields.goos = "darwin"
			}
			if tt.fields.goarch == "" {
				tt.fields.goarch = "amd64"
			}
			p, err := NewPrefixedKey(tt.fields.Prefix, tt.fields.Separator, tt.fields.Versioned)
			req.NoError(err)
			// Need to inject these so tests pass in different environments (e.g., CI)
			p.goos = tt.fields.goos
			p.goarch = tt.fields.goarch
			match, ver, err := p.ParseVersion(tt.args.key, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("PrefixedKey.ParseVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if match != tt.wantMatch {
				t.Errorf("PrefixedKey.ParseVersion() match = %v, wantMatch %v", match, tt.wantMatch)
			}
			if !reflect.DeepEqual(ver, tt.wantVer) {
				t.Errorf("PrefixedKey.ParseVersion() ver = %v, wantVer %v", ver, tt.wantVer)
			}
		})
	}
}

func TestPrefixedKey_URLFor(t *testing.T) {
	req := require.New(t)
	type fields struct {
		Prefix          string
		VersionPrefixed bool
		OS              string
		Arch            string
	}
	type args struct {
		name    string
		version string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "with version prefix",
			fields: fields{
				Prefix:          "my-pre",
				VersionPrefixed: true,
			},
			args: args{
				name:    "fancy-cli",
				version: "v1.23.0",
			},
			want: "my-pre/v1.23.0/fancy-cli_v1.23.0_darwin_amd64",
		},
		{
			name: "without version prefix",
			fields: fields{
				Prefix:          "my-pre",
				VersionPrefixed: false,
			},
			args: args{
				name:    "fancy-cli",
				version: "v1.23.0",
			},
			want: "my-pre/fancy-cli_v1.23.0_darwin_amd64",
		},
		{
			name:   "with empty prefix and no version prefix",
			fields: fields{},
			args: args{
				name:    "fancy-cli",
				version: "0.1.2",
			},
			want: "fancy-cli_0.1.2_darwin_amd64",
		},
		{
			name: "with empty prefix and with version prefix",
			fields: fields{
				VersionPrefixed: true,
			},
			args: args{
				name:    "fancy-cli",
				version: "0.1.2",
			},
			want: "0.1.2/fancy-cli_0.1.2_darwin_amd64",
		},
		{
			name: "should include file extension for windows (with prefix and version-prefix)",
			fields: fields{
				Prefix:          "pre",
				VersionPrefixed: true,
				OS:              "windows",
				Arch:            "386",
			},
			args: args{
				name:    "fancy-cli",
				version: "0.1.2",
			},
			want: "pre/0.1.2/fancy-cli_0.1.2_windows_386.exe",
		},
		{
			name: "should include file extension for windows (without prefix or version-prefix)",
			fields: fields{
				OS:   "windows",
				Arch: "386",
			},
			args: args{
				name:    "fancy-cli",
				version: "0.1.2",
			},
			want: "fancy-cli_0.1.2_windows_386.exe",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fields.OS == "" {
				tt.fields.OS = "darwin"
				tt.fields.Arch = "amd64"
			}
			p, err := NewPrefixedKey(tt.fields.Prefix, "_", tt.fields.VersionPrefixed)
			req.NoError(err)
			// Need to inject these so tests pass in different environments (e.g., CI)
			p.goos = tt.fields.OS
			p.goarch = tt.fields.Arch
			if got := p.URLFor(tt.args.name, tt.args.version); got != tt.want {
				t.Errorf("PrefixedKey.URLFor() = %v, want %v", got, tt.want)
			}
		})
	}
}
