package testserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	flinkartifactv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-artifact/v1"
)

var artifactVersions = &[]flinkartifactv1.ArtifactV1FlinkArtifactVersion{
	{
		Version:      "1.0.0",
		ReleaseNotes: flinkartifactv1.PtrString("Initial release"),
		IsBeta:       flinkartifactv1.PtrBool(false),
		UploadSource: flinkartifactv1.ArtifactV1FlinkArtifactVersionUploadSourceOneOf{
			ArtifactV1UploadSourcePresignedUrl: &flinkartifactv1.ArtifactV1UploadSourcePresignedUrl{
				Location: flinkartifactv1.PtrString("PRESIGNED_URL_LOCATION"),
				UploadId: flinkartifactv1.PtrString("u-123"),
			},
		},
	},
}

// Handler for: "/artifact/v1/flink-artifacts"
func handleFlinkArtifactPlugins(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var decodeRespone flinkartifactv1.ArtifactV1FlinkArtifact
			require.NoError(t, json.NewDecoder(r.Body).Decode(&decodeRespone))
			var plugin flinkartifactv1.ArtifactV1FlinkArtifact
			switch strings.ToLower(decodeRespone.GetRuntimeLanguage()) {
			case "java", "":
				plugin = flinkartifactv1.ArtifactV1FlinkArtifact{
					Id:            flinkartifactv1.PtrString("cfa-123456"),
					Cloud:         flinkartifactv1.PtrString("AWS"),
					Region:        flinkartifactv1.PtrString("us-west-2"),
					Environment:   flinkartifactv1.PtrString("env-123456"),
					DisplayName:   flinkartifactv1.PtrString("my-flink-artifact"),
					Class:         flinkartifactv1.PtrString("io.confluent.flink.example1.test"),
					ContentFormat: flinkartifactv1.PtrString("JAR"),
					Versions:      artifactVersions,
				}
			case "python":
				plugin = flinkartifactv1.ArtifactV1FlinkArtifact{
					Id:            flinkartifactv1.PtrString("cfa-789012"),
					Cloud:         flinkartifactv1.PtrString("AWS"),
					Region:        flinkartifactv1.PtrString("us-east-1"),
					Environment:   flinkartifactv1.PtrString("env-789012"),
					DisplayName:   flinkartifactv1.PtrString("my-flink-python-artifact"),
					Class:         flinkartifactv1.PtrString("io.confluent.flink.example2.test"),
					ContentFormat: flinkartifactv1.PtrString("ZIP"),
					Versions:      artifactVersions,
				}
			}
			err := json.NewEncoder(w).Encode(plugin)
			require.NoError(t, err)
		case http.MethodGet:
			plugin1 := flinkartifactv1.ArtifactV1FlinkArtifact{
				Id:            flinkartifactv1.PtrString("cfa-123456"),
				Cloud:         flinkartifactv1.PtrString("AWS"),
				Region:        flinkartifactv1.PtrString("us-west-2"),
				Environment:   flinkartifactv1.PtrString("env-123456"),
				DisplayName:   flinkartifactv1.PtrString("my-flink-artifact"),
				Class:         flinkartifactv1.PtrString("io.confluent.flink.example1.test"),
				ContentFormat: flinkartifactv1.PtrString("JAR"),
				Versions:      artifactVersions,
			}
			plugin2 := flinkartifactv1.ArtifactV1FlinkArtifact{
				Id:            flinkartifactv1.PtrString("cfa-789012"),
				Cloud:         flinkartifactv1.PtrString("AWS"),
				Region:        flinkartifactv1.PtrString("us-east-1"),
				Environment:   flinkartifactv1.PtrString("env-789012"),
				DisplayName:   flinkartifactv1.PtrString("my-flink-python-artifact"),
				Class:         flinkartifactv1.PtrString("io.confluent.flink.example2.test"),
				ContentFormat: flinkartifactv1.PtrString("ZIP"),
				Versions:      artifactVersions,
			}
			err := json.NewEncoder(w).Encode(flinkartifactv1.ArtifactV1FlinkArtifactList{Data: []flinkartifactv1.ArtifactV1FlinkArtifact{plugin1, plugin2}})
			require.NoError(t, err)
		}
	}
}

// Handler for: "/artifact/v1/flink-artifacts/{id}"
func handleFlinkArtifactPluginsId(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			vars := mux.Vars(r)
			id := vars["id"]
			var plugin flinkartifactv1.ArtifactV1FlinkArtifact
			if id == "cfa-123456" {
				plugin = flinkartifactv1.ArtifactV1FlinkArtifact{
					Id:            flinkartifactv1.PtrString("cfa-123456"),
					Cloud:         flinkartifactv1.PtrString("AWS"),
					Region:        flinkartifactv1.PtrString("us-west-2"),
					Environment:   flinkartifactv1.PtrString("env-123456"),
					DisplayName:   flinkartifactv1.PtrString("my-flink-artifact"),
					Class:         flinkartifactv1.PtrString("io.confluent.flink.example1.test"),
					ContentFormat: flinkartifactv1.PtrString("JAR"),
					Versions:      artifactVersions,
				}
			} else if id == "cfa-789012" {
				plugin = flinkartifactv1.ArtifactV1FlinkArtifact{
					Id:            flinkartifactv1.PtrString("cfa-789012"),
					Cloud:         flinkartifactv1.PtrString("AWS"),
					Region:        flinkartifactv1.PtrString("us-east-1"),
					Environment:   flinkartifactv1.PtrString("env-789012"),
					DisplayName:   flinkartifactv1.PtrString("my-flink-python-artifact"),
					Description:   flinkartifactv1.PtrString("Flink custom artifact"),
					Class:         flinkartifactv1.PtrString("io.confluent.flink.example2.test"),
					ContentFormat: flinkartifactv1.PtrString("ZIP"),
					Versions:      artifactVersions,
				}
			} else {
				plugin = flinkartifactv1.ArtifactV1FlinkArtifact{
					Id:            flinkartifactv1.PtrString("cfa-789013"),
					Cloud:         flinkartifactv1.PtrString("AWS"),
					Region:        flinkartifactv1.PtrString("us-west-2"),
					Environment:   flinkartifactv1.PtrString("env-789013"),
					DisplayName:   flinkartifactv1.PtrString("CliArtifactTest"),
					Class:         flinkartifactv1.PtrString("io.confluent.flink.example3.test"),
					ContentFormat: flinkartifactv1.PtrString("JAR"),
					Versions:      artifactVersions,
				}
			}
			err := json.NewEncoder(w).Encode(plugin)
			require.NoError(t, err)
		case http.MethodPatch:
			plugin := flinkartifactv1.ArtifactV1FlinkArtifact{
				Id:          flinkartifactv1.PtrString("cfa-123456"),
				DisplayName: flinkartifactv1.PtrString("CliArtifactTestUpdate"),
			}
			err := json.NewEncoder(w).Encode(plugin)
			require.NoError(t, err)
		case http.MethodDelete:
			err := json.NewEncoder(w).Encode(flinkartifactv1.ArtifactV1FlinkArtifact{})
			require.NoError(t, err)
		}
	}
}

// Handler for: "/artifact/v1/presigned-upload-url"
func handleFlinkArtifactUploadUrl(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			uploadUrl := flinkartifactv1.ArtifactV1PresignedUrl{
				ContentFormat: flinkartifactv1.PtrString("ZIP"),
				Cloud:         flinkartifactv1.PtrString("AWS"),
				Region:        flinkartifactv1.PtrString("us-west-2"),
				UploadId:      flinkartifactv1.PtrString("e53bb2e8-8de3-49fa-9fb1-4e3fd9a16b66"),
				UploadUrl:     flinkartifactv1.PtrString(fmt.Sprintf("%s/connect/v1/dummy-presigned-url", TestV2CloudUrl.String())),
			}
			err := json.NewEncoder(w).Encode(uploadUrl)
			require.NoError(t, err)
		}
	}
}
