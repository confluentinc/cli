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

var awsJavaArtifact = flinkartifactv1.ArtifactV1FlinkArtifact{
	Id:                flinkartifactv1.PtrString("cfa-123456"),
	Cloud:             flinkartifactv1.PtrString("AWS"),
	Region:            flinkartifactv1.PtrString("us-west-2"),
	Environment:       flinkartifactv1.PtrString("env-123456"),
	DisplayName:       flinkartifactv1.PtrString("my-flink-artifact"),
	ContentFormat:     flinkartifactv1.PtrString("JAR"),
	Description:       flinkartifactv1.PtrString("CliArtifactTest"),
	DocumentationLink: flinkartifactv1.PtrString("https://docs.confluent.io"),
	Versions:          artifactVersions,
}

var awsPythonArtifact = flinkartifactv1.ArtifactV1FlinkArtifact{
	Id:            flinkartifactv1.PtrString("cfa-789012"),
	Cloud:         flinkartifactv1.PtrString("AWS"),
	Region:        flinkartifactv1.PtrString("us-east-1"),
	Environment:   flinkartifactv1.PtrString("env-789012"),
	DisplayName:   flinkartifactv1.PtrString("my-flink-python-artifact"),
	ContentFormat: flinkartifactv1.PtrString("ZIP"),
	Description: flinkartifactv1.PtrString("This is a longer description example to verify the" +
		" output of the CLI is not affected and remains readable"),
	DocumentationLink: flinkartifactv1.PtrString("https://docs.confluent.io"),
	Versions:          artifactVersions,
}

var azureJavaArtifact = flinkartifactv1.ArtifactV1FlinkArtifact{
	Id:                flinkartifactv1.PtrString("cfa-123456"),
	Cloud:             flinkartifactv1.PtrString("AZURE"),
	Region:            flinkartifactv1.PtrString("centralus"),
	Environment:       flinkartifactv1.PtrString("env-123456"),
	DisplayName:       flinkartifactv1.PtrString("my-flink-artifact"),
	ContentFormat:     flinkartifactv1.PtrString("JAR"),
	Description:       flinkartifactv1.PtrString("CliArtifactTest"),
	DocumentationLink: flinkartifactv1.PtrString("https://docs.confluent.io"),
	Versions:          artifactVersions,
}

var azurePythonArtifact = flinkartifactv1.ArtifactV1FlinkArtifact{
	Id:            flinkartifactv1.PtrString("cfa-789012"),
	Cloud:         flinkartifactv1.PtrString("AZURE"),
	Region:        flinkartifactv1.PtrString("centralus"),
	Environment:   flinkartifactv1.PtrString("env-789012"),
	DisplayName:   flinkartifactv1.PtrString("my-flink-python-artifact"),
	ContentFormat: flinkartifactv1.PtrString("ZIP"),
	Description: flinkartifactv1.PtrString("This is a longer description example to verify the" +
		" output of the CLI is not affected and remains readable"),
	DocumentationLink: flinkartifactv1.PtrString("https://docs.confluent.io"),
	Versions:          artifactVersions,
}

var gcpJavaArtifact = flinkartifactv1.ArtifactV1FlinkArtifact{
	Id:                flinkartifactv1.PtrString("cfa-123456"),
	Cloud:             flinkartifactv1.PtrString("GCP"),
	Region:            flinkartifactv1.PtrString("us-central1"),
	Environment:       flinkartifactv1.PtrString("env-123456"),
	DisplayName:       flinkartifactv1.PtrString("my-flink-artifact"),
	ContentFormat:     flinkartifactv1.PtrString("JAR"),
	Description:       flinkartifactv1.PtrString("CliArtifactTest"),
	DocumentationLink: flinkartifactv1.PtrString("https://docs.confluent.io"),
	Versions:          artifactVersions,
}

var gcpPythonArtifact = flinkartifactv1.ArtifactV1FlinkArtifact{
	Id:            flinkartifactv1.PtrString("cfa-789012"),
	Cloud:         flinkartifactv1.PtrString("GCP"),
	Region:        flinkartifactv1.PtrString("us-central1"),
	Environment:   flinkartifactv1.PtrString("env-789012"),
	DisplayName:   flinkartifactv1.PtrString("my-flink-python-artifact"),
	ContentFormat: flinkartifactv1.PtrString("ZIP"),
	Description: flinkartifactv1.PtrString("This is a longer description example to verify the" +
		" output of the CLI is not affected and remains readable"),
	DocumentationLink: flinkartifactv1.PtrString("https://docs.confluent.io"),
	Versions:          artifactVersions,
}

// Handler for: "/artifact/v1/flink-artifacts"
func handleFlinkArtifacts(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var decodeRespone flinkartifactv1.ArtifactV1FlinkArtifact
			require.NoError(t, json.NewDecoder(r.Body).Decode(&decodeRespone))
			var artifact flinkartifactv1.ArtifactV1FlinkArtifact
			switch strings.ToLower(decodeRespone.GetCloud()) {
			case "gcp":
				switch strings.ToLower(decodeRespone.GetRuntimeLanguage()) {
				case "java", "":
					artifact = gcpJavaArtifact
				case "python":
					artifact = gcpPythonArtifact
				}
			case "azure":
				switch strings.ToLower(decodeRespone.GetRuntimeLanguage()) {
				case "java", "":
					artifact = azureJavaArtifact
				case "python":
					artifact = azurePythonArtifact
				}
			case "aws":
				switch strings.ToLower(decodeRespone.GetRuntimeLanguage()) {
				case "java", "":
					artifact = awsJavaArtifact
				case "python":
					artifact = awsPythonArtifact
				}
			}
			err := json.NewEncoder(w).Encode(artifact)
			require.NoError(t, err)
		case http.MethodGet:
			var artifact1 flinkartifactv1.ArtifactV1FlinkArtifact
			var artifact2 flinkartifactv1.ArtifactV1FlinkArtifact
			switch strings.ToLower(r.URL.Query().Get("cloud")) {
			case "gcp":
				artifact1 = gcpJavaArtifact
				artifact2 = gcpPythonArtifact
			case "azure":
				artifact1 = azureJavaArtifact
				artifact2 = azurePythonArtifact
			case "aws":
				artifact1 = awsJavaArtifact
				artifact2 = awsPythonArtifact
			}
			artifactList := &flinkartifactv1.ArtifactV1FlinkArtifactList{Data: []flinkartifactv1.ArtifactV1FlinkArtifact{artifact1, artifact2}}
			setPageToken(artifactList, &artifactList.Metadata, r.URL)
			err := json.NewEncoder(w).Encode(artifactList)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/artifact/v1/flink-artifacts/{id}"
func handleFlinkArtifactId(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			vars := mux.Vars(r)
			id := vars["id"]
			var artifact flinkartifactv1.ArtifactV1FlinkArtifact
			if strings.ToLower(r.URL.Query().Get("cloud")) == "azure" {
				artifact = azureJavaArtifact
			} else if strings.ToLower(r.URL.Query().Get("cloud")) == "gcp" {
				artifact = gcpJavaArtifact
			} else if id == "cfa-123456" {
				artifact = flinkartifactv1.ArtifactV1FlinkArtifact{
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
				artifact = flinkartifactv1.ArtifactV1FlinkArtifact{
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
				artifact = flinkartifactv1.ArtifactV1FlinkArtifact{
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
			err := json.NewEncoder(w).Encode(artifact)
			require.NoError(t, err)
		case http.MethodPatch:
			artifact := flinkartifactv1.ArtifactV1FlinkArtifact{
				Id:          flinkartifactv1.PtrString("cfa-123456"),
				DisplayName: flinkartifactv1.PtrString("CliArtifactTestUpdate"),
			}
			err := json.NewEncoder(w).Encode(artifact)
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
		} else if r.Method == http.MethodPut {
			uploadUrl := flinkartifactv1.ArtifactV1PresignedUrl{
				ContentFormat: flinkartifactv1.PtrString("ZIP"),
				Cloud:         flinkartifactv1.PtrString("AZURE"),
				Region:        flinkartifactv1.PtrString("centralus"),
				UploadId:      flinkartifactv1.PtrString("e53bb2e8-8de3-49fa-9fb1-4e3fd9a16b66"),
				UploadUrl:     flinkartifactv1.PtrString(fmt.Sprintf("%s/connect/v1/dummy-presigned-url", TestV2CloudUrl.String())),
			}
			err := json.NewEncoder(w).Encode(uploadUrl)
			require.NoError(t, err)
		}
	}
}
