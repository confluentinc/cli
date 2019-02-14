package shared

import (
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestConfig_Load(t *testing.T) {
	type args struct {
		contents string
	}
	tests := []struct {
		name    string
		args    *args
		want    *Config
		wantErr bool
		file    string
	}{
		{
			name: "should load auth token from file",
			args: &args{
				contents: "{\"auth_token\": \"abc123\"}",
			},
			want: &Config{
				AuthToken:   "abc123",
				Platforms:   map[string]*Platform{},
				Credentials: map[string]*Credential{},
				Contexts:    map[string]*Context{},
			},
			file: "/tmp/TestConfig_Load.json",
		},
		{
			name: "should load auth url from file",
			args: &args{
				contents: "{\"auth_url\": \"https://stag.cpdev.cloud\"}",
			},
			want: &Config{
				AuthURL:     "https://stag.cpdev.cloud",
				Platforms:   map[string]*Platform{},
				Credentials: map[string]*Credential{},
				Contexts:    map[string]*Context{},
			},
			file: "/tmp/TestConfig_Load.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewConfig()
			c.Filename = tt.file
			err := ioutil.WriteFile(tt.file, []byte(tt.args.contents), 0644)
			if err != nil {
				t.Errorf("unable to test config to file: %v", err)
			}
			if err := c.Load(); (err != nil) != tt.wantErr {
				t.Errorf("Config.Load() error = %v, wantErr %v", err, tt.wantErr)
			}
			c.Filename = "" // only for testing
			if !reflect.DeepEqual(c, tt.want) {
				t.Errorf("Config.Load() = %v, want %v", c, tt.want)
			}
			os.Remove(tt.file)
		})
	}
}

func TestConfig_Save(t *testing.T) {
	type args struct {
		url   string
		token string
	}
	tests := []struct {
		name    string
		args    *args
		want    string
		wantErr bool
		file    string
	}{
		{
			name: "save auth token to file",
			args: &args{
				token: "abc123",
			},
			want: "\"auth_token\": \"abc123\"",
			file: "/tmp/TestConfig_Save.json",
		},
		{
			name: "save auth url to file",
			args: &args{
				url: "https://stag.cpdev.cloud",
			},
			want: "\"auth_url\": \"https://stag.cpdev.cloud\"",
			file: "/tmp/TestConfig_Save.json",
		},
		{
			name: "create parent config dirs",
			args: &args{
				token: "abc123",
			},
			want: "\"auth_token\": \"abc123\"",
			file: "/tmp/xyz987/TestConfig_Save.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{Filename: tt.file, AuthToken: tt.args.token, AuthURL: tt.args.url}
			if err := c.Save(); (err != nil) != tt.wantErr {
				t.Errorf("Config.Save() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, _ := ioutil.ReadFile(tt.file)
			if !strings.Contains(string(got), tt.want) {
				t.Errorf("Config.Save() = %v, want contains %v", string(got), tt.want)
			}
			fd, _ := os.Stat(tt.file)
			if fd.Mode() != 0600 {
				t.Errorf("Config.Save() file should only be readable by user")
			}
			os.RemoveAll("/tmp/xyz987")
		})
	}
}
