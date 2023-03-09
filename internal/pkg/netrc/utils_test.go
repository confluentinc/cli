package netrc

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseNetrcMachineName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		machineName string
		want        *MachineContextInfo
	}{
		{
			name:        "ccloud username password login",
			machineName: "confluent-cli:ccloud-username-password:login-csreesangkom@confluent.io-https://confluent.cloud",
			want: &MachineContextInfo{
				CredentialType: ccloudUsernamePasswordString,
				Username:       "csreesangkom@confluent.io",
				URL:            "https://confluent.cloud",
				CaCertPath:     "",
			},
		},
		{
			name:        "confluent username password login no ca-cert-path",
			machineName: "confluent-cli:mds-username-password:login-alice-http://localhost:8090",
			want: &MachineContextInfo{
				CredentialType: mdsUsernamePasswordString,
				Username:       "alice",
				URL:            "http://localhost:8090",
				CaCertPath:     "",
			},
		},
		{
			name:        "confluent username password login with ca-cert-path",
			machineName: "confluent-cli:mds-username-password:login-alice-http://localhost:8090?cacertpath=path",
			want: &MachineContextInfo{
				CredentialType: mdsUsernamePasswordString,
				Username:       "alice",
				URL:            "http://localhost:8090",
				CaCertPath:     "path",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseNetrcMachineName(tt.machineName)
			require.NoError(t, err)
			compareMachineContextInfo(t, tt.want, got)
		})
	}
}

func compareMachineContextInfo(t *testing.T, expected, actual *MachineContextInfo) {
	var mismatchValues []string
	if expected.CredentialType != actual.CredentialType {
		mismatchValues = append(mismatchValues, "CredentialType")
	}
	if expected.Username != actual.Username {
		mismatchValues = append(mismatchValues, "Username")
	}
	if expected.URL != actual.URL {
		mismatchValues = append(mismatchValues, "URL")
	}
	if expected.CaCertPath != actual.CaCertPath {
		mismatchValues = append(mismatchValues, "CaCertPath")
	}

	if len(mismatchValues) > 0 {
		failedMessageFormat := "MachineContextInfo fields mistmatch: %s\n" +
			"expected: %+v,\n" +
			"got: %+v"

		mismatchFields := strings.Join(mismatchValues, ", ")

		require.Failf(t, "Fail", failedMessageFormat, mismatchFields, expected, actual)
	}
}
