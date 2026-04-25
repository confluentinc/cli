package schemaregistry

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"
)

func TestRegisterSchemaWithAuth_ForwardsNormalizeFlag(t *testing.T) {
	tests := []struct {
		name          string
		normalize     bool
		expectedQuery string
	}{
		{
			name:          "normalize false",
			normalize:     false,
			expectedQuery: "false",
		},
		{
			name:          "normalize true",
			normalize:     true,
			expectedQuery: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var captured url.Values

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				captured = r.URL.Query()
				w.Header().Set("Content-Type", "application/vnd.schemaregistry.v1+json")
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(`{"id": 1}`))
				require.NoError(t, err)
			}))
			defer server.Close()

			schemaPath := filepath.Join(t.TempDir(), "employee.avsc")
			require.NoError(t, os.WriteFile(schemaPath, []byte(`{"type":"string"}`), 0644))

			srConfig := srsdk.NewConfiguration()
			srConfig.Servers = srsdk.ServerConfigurations{{URL: server.URL}}
			client := NewClientWithApiKey(srConfig, srsdk.BasicAuth{UserName: "u", Password: "p"})

			cfg := &RegisterSchemaConfigs{
				Subject:    "employee-value",
				SchemaType: "AVRO",
				SchemaPath: schemaPath,
				Normalize:  tt.normalize,
			}

			id, err := RegisterSchemaWithAuth(cfg, client)
			require.NoError(t, err)
			assert.Equal(t, int32(1), id)
			assert.Equal(t, tt.expectedQuery, captured.Get("normalize"))
		})
	}
}
