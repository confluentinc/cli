package testserver

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/v4/pkg/cpstructs"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

const archivePath = "test/fixtures/input/connect/test-plugin.zip"

var pluginManifestMap = map[string]*cpstructs.Manifest{
	"/api/plugins/confluentinc/bad-md5": {
		Name:    "integration-test-plugin",
		Title:   "Integration Test Plugin",
		Version: "0.1.0",
		Owner: cpstructs.Owner{
			Username: "confluentinc",
			Name:     "Confluent, Inc.",
		},
		Licenses: []cpstructs.License{
			{
				Name: "Apache License 2.0",
				Url:  "https://www.apache.org/licenses/LICENSE-2.0",
			},
		},
	},
	"/api/plugins/confluentinc/bad-sha1": {
		Name:    "integration-test-plugin",
		Title:   "Integration Test Plugin",
		Version: "0.1.0",
		Owner: cpstructs.Owner{
			Username: "confluentinc",
			Name:     "Confluent, Inc.",
		},
		Licenses: []cpstructs.License{
			{
				Name: "Apache License 2.0",
				Url:  "https://www.apache.org/licenses/LICENSE-2.0",
			},
		},
	},
	"/api/plugins/confluentinc/integration-test-plugin": {
		Name:    "integration-test-plugin",
		Title:   "Integration Test Plugin",
		Version: "0.1.0",
		Owner: cpstructs.Owner{
			Username: "confluentinc",
			Name:     "Confluent, Inc.",
		},
		Licenses: []cpstructs.License{
			{
				Name: "Apache License 2.0",
				Url:  "https://www.apache.org/licenses/LICENSE-2.0",
			},
		},
	},
	"/api/plugins/confluentinc/integration-test-plugin/versions/0.1.0": {
		Name:    "integration-test-plugin",
		Title:   "Integration Test Plugin",
		Version: "0.1.0",
		Owner: cpstructs.Owner{
			Username: "confluentinc",
			Name:     "Confluent, Inc.",
		},
		Licenses: []cpstructs.License{
			{
				Name: "Apache License 2.0",
				Url:  "https://www.apache.org/licenses/LICENSE-2.0",
			},
		},
	},
	"/api/plugins/confluentinc/integration-test-plugin/versions/0.0.5": {
		Name:    "integration-test-plugin",
		Title:   "Integration Test Plugin",
		Version: "0.0.5",
		Owner: cpstructs.Owner{
			Username: "confluentinc",
			Name:     "Confluent, Inc.",
		},
		Licenses: []cpstructs.License{
			{
				Name: "Apache License 2.0",
				Url:  "https://www.apache.org/licenses/LICENSE-2.0",
			},
		},
	},
	"/api/plugins/confluentinc/integration-test-plugin/versions/0.0.4": {
		Name:    "integration-test-plugin",
		Title:   "Integration Test Plugin",
		Version: "0.0.4",
		Owner: cpstructs.Owner{
			Username: "confluentinc",
			Name:     "Confluent, Inc.",
		},
		Licenses: []cpstructs.License{
			{
				Name: "Apache License 2.0",
				Url:  "https://www.apache.org/licenses/LICENSE-2.0",
			},
		},
		EndOfLifeAt: "2025-06-01T00:00:00Z",
	},
}

func newManifest(urlPath string) *cpstructs.Manifest {
	manifest, ok := pluginManifestMap[urlPath]
	if !ok {
		return nil
	}
	return manifest
}

// Handler for: "/api/plugins/{owner}/{id}"
func handleHubPlugin(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		responseManifest := newManifest(r.URL.Path)
		if responseManifest == nil {
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
			return
		}

		switch r.Method {
		case http.MethodGet:
			if utils.DoesPathExist(archivePath) {
				archive, err := os.ReadFile(archivePath)
				require.NoError(t, err)

				responseManifest.Archive = cpstructs.Archive{
					Url:  fmt.Sprintf("%s/api/plugins/confluentinc/integration-test-plugin/versions/0.1.0/confluentinc-integration-test-plugin.zip", TestHubUrl.String()),
					Md5:  fmt.Sprintf("%x", md5.Sum(archive)),
					Sha1: fmt.Sprintf("%x", sha1.Sum(archive)),
				}
				id := mux.Vars(r)["id"]
				if id == "bad-md5" {
					responseManifest.Archive.Md5 = "12345"
				}
				if id == "bad-sha1" {
					responseManifest.Archive.Sha1 = "12345"
				}
			}
			err := json.NewEncoder(w).Encode(responseManifest)
			require.NoError(t, err)
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

// Handler for: "/api/plugins/{owner}/{id}/versions/{version}"
func handleHubPluginVersion(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		responseManifest := newManifest(r.URL.Path)
		if responseManifest == nil {
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
			return
		}

		switch r.Method {
		case http.MethodGet:
			if utils.DoesPathExist(archivePath) {
				archive, err := os.ReadFile(archivePath)
				require.NoError(t, err)

				responseManifest.Archive = cpstructs.Archive{
					Url:  fmt.Sprintf("%s/api/plugins/confluentinc/integration-test-plugin/versions/0.1.0/confluentinc-integration-test-plugin.zip", TestHubUrl.String()),
					Md5:  fmt.Sprintf("%x", md5.Sum(archive)),
					Sha1: fmt.Sprintf("%x", sha1.Sum(archive)),
				}
			}
			err := json.NewEncoder(w).Encode(responseManifest)
			require.NoError(t, err)
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

// Handler for: "/api/plugins/{owner}/{id}/versions/{version}/{archive}"
func handleHubPluginArchive(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		if vars["owner"] != "confluentinc" || vars["id"] != "integration-test-plugin" || vars["version"] != "0.1.0" || vars["archive"] != "confluentinc-integration-test-plugin.zip" {
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
			return
		}

		switch r.Method {
		case http.MethodGet:
			archive, err := os.ReadFile(archivePath)
			require.NoError(t, err)
			_, err = w.Write(archive)
			require.NoError(t, err)
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}
}
