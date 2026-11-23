package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertDateFormatToString(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		date := time.Date(2021, time.June, 16, 12, 0, 0, 0, time.UTC)
		timestamp := &types.Timestamp{Seconds: date.Unix()}
		actual := convertDateFormat(timestamp)
		expected := "06/16/2021"
		assert.Equal(t, expected, actual)
	})
	t.Run("fail", func(t *testing.T) {
		date := time.Date(2021, time.April, 16, 12, 0, 0, 0, time.UTC)
		timestamp := &types.Timestamp{Seconds: date.Unix()}
		actual := convertDateFormat(timestamp)
		expected := "06/16/2021"
		assert.NotEqual(t, expected, actual)
	})
	t.Run("fail, nil", func(t *testing.T) {
		actual := convertDateFormat(&types.Timestamp{})
		expected := "01/01/1970"
		assert.Equal(t, expected, actual)
	})
}

func TestGetCCloudTokens_ExpiredRefreshTokenWithoutPassword(t *testing.T) {
	tests := []struct {
		name         string
		password     string
		wantRequests int
	}{
		{name: "no password skips the login request", password: "", wantRequests: 1},
		{name: "password falls back to a login request", password: "pass", wantRequests: 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				w.WriteHeader(http.StatusUnauthorized)
			}))
			defer server.Close()

			credentials := &Credentials{Username: "user@example.com", Password: test.password, AuthRefreshToken: "expired"}
			_, _, err := NewAuthTokenHandler().GetCCloudTokens(NewCCloudClientFactory("test"), server.URL, credentials, false, "")
			require.Error(t, err)
			require.Equal(t, test.wantRequests, requests)
		})
	}
}
