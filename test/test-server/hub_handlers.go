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

	"github.com/confluentinc/cli/internal/pkg/ccstructs"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

const archivePath = "test/fixtures/input/connect/test-plugin.zip"

func newManifest() *ccstructs.Manifest {
	return &ccstructs.Manifest{
		Name:    "integration-test-plugin",
		Title:   "Integration Test Plugin",
		Version: "0.1.0",
		Owner: ccstructs.Owner{
			Username: "confluentinc",
			Name:     "Confluent, Inc.",
		},
		Licenses: []ccstructs.License{
			{
				Name: "Example License 1.0",
				Url:  "url-to-license-goes-here",
			},
		},
	}
}

// Handler for: "/api/plugins/{owner}/{id}"
func handleHubPlugin(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		if id == "dne-connector" || vars["owner"] != "confluentinc" {
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
			return
		}

		switch r.Method {
		case http.MethodGet:
			responseManifest := newManifest()
			if utils.DoesPathExist(archivePath) {
				archive, err := os.ReadFile(archivePath)
				require.NoError(t, err)

				responseManifest.Archive = ccstructs.Archive{
					Url:  fmt.Sprintf("%s/api/plugins/confluentinc/integration-test-plugin/versions/0.1.0/confluentinc-integration-test-plugin.zip", TestHubUrl.String()),
					Md5:  fmt.Sprintf("%x", md5.Sum(archive)),
					Sha1: fmt.Sprintf("%x", sha1.Sum(archive)),
				}
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
		vars := mux.Vars(r)
		if vars["id"] == "dne-connector" || vars["owner"] != "confluentinc" || vars["version"] != "0.0.5" {
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
			return
		}

		switch r.Method {
		case http.MethodGet:
			responseManifest := newManifest()
			responseManifest.Version = "0.0.5"
			if utils.DoesPathExist(archivePath) {
				archive, err := os.ReadFile(archivePath)
				require.NoError(t, err)

				responseManifest.Archive = ccstructs.Archive{
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
