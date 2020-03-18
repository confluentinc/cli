package auth

import (
	"testing"
)

func TestNetRCReader(t *testing.T) {
	username := "existing-user"
	password := "existing-password"
	tests := []struct {
		name    string
		want    []string
		contextName string
		wantErr bool
		file    string
	}{
		{
			name: "Context exist",
			want: []string{username, password},
			contextName: "existing-context",
			file: "test_files/netrc",
		},
		{
			name: "No file error",
			contextName: "existing-context",
			wantErr: true,
			file: "test_files/not-netrc",
		},
		{
			name: "Context doesn't exist",
			contextName: "non-existing-context",
			wantErr: true,
			file: "test_files/netrc",
		},
		{
			name: "Context exist with no password",
			want: []string{username, ""},
			contextName: "no-password-context",
			file: "test_files/netrc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			netrcHandler := netrcHandler{fileName:tt.file}
			var username, password string
			var err error
			if username, password, err = netrcHandler.getNetrcCredentials(tt.contextName); (err != nil) != tt.wantErr {
				t.Errorf("getNetrcCredentials error = %+v, wantErr %+v", err, tt.wantErr)
			}
			if len(tt.want) != 0 && !t.Failed() && username != tt.want[0] {
				t.Errorf("getNetrcCredenials username = %+v, want %+v", username, tt.want[0])
			}
			if len(tt.want) == 2 && !t.Failed() && password != tt.want[1] {
				t.Errorf("getNetrcCredenials password = %+v, want %+v", password, tt.want[1])
			}
		})
	}
}
