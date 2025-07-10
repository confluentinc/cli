package testserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	ccpmv1 "github.com/confluentinc/ccloud-sdk-go-v2/ccpm/v1"
)

// CCPM Plugin handlers
func handleCCPMPlugins(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// List plugins
			cloud := r.URL.Query().Get("spec.cloud")
			environment := r.URL.Query().Get("environment")

			// Check if this is an error test case
			if environment == "env-error" && cloud == "" {
				// This is the error handling test case - return error
				http.Error(w, "API Error", http.StatusInternalServerError)
				return
			}

			plugins := []map[string]interface{}{
				{
					"api_version": "ccpm/v1",
					"kind":        "CustomConnectPlugin",
					"id":          "ccp-123456",
					"metadata": map[string]interface{}{
						"created_at": "2023-01-01T00:00:00Z",
						"updated_at": "2023-01-01T00:00:00Z",
					},
					"spec": map[string]interface{}{
						"display_name":     "CliPluginTest1",
						"description":      "Test plugin",
						"cloud":            "AWS",
						"runtime_language": "JAVA",
						"environment": map[string]interface{}{
							"id": environment,
						},
					},
				},
				{
					"api_version": "ccpm/v1",
					"kind":        "CustomConnectPlugin",
					"id":          "ccp-789012",
					"metadata": map[string]interface{}{
						"created_at": "2023-01-01T00:00:00Z",
						"updated_at": "2023-01-01T00:00:00Z",
					},
					"spec": map[string]interface{}{
						"display_name":     "CliPluginTest2",
						"description":      "Test plugin",
						"cloud":            "GCP",
						"runtime_language": "JAVA",
						"environment": map[string]interface{}{
							"id": environment,
						},
					},
				},
			}

			// Filter by cloud if specified
			if cloud != "" {
				filtered := []map[string]interface{}{}
				cloudUpper := strings.ToUpper(cloud)
				for _, plugin := range plugins {
					pluginCloud := plugin["spec"].(map[string]interface{})["cloud"].(string)
					if pluginCloud == cloudUpper {
						filtered = append(filtered, plugin)
					}
				}
				plugins = filtered
			}

			response := map[string]interface{}{
				"data": plugins,
				"metadata": map[string]interface{}{
					"next": "",
				},
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)

		case http.MethodPost:
			// Create plugin
			var request map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			// Check if this is an error test case
			spec, ok := request["spec"].(map[string]interface{})
			if !ok {
				http.Error(w, "Missing spec", http.StatusBadRequest)
				return
			}

			environment, ok := spec["environment"].(map[string]interface{})
			if ok {
				if envId, ok := environment["id"].(string); ok && envId == "env-error" {
					// This is the error handling test case - return error
					http.Error(w, "API Error", http.StatusInternalServerError)
					return
				}
			}

			// Validate required fields
			displayName, ok := spec["display_name"].(string)
			if !ok || displayName == "" {
				http.Error(w, "Missing display_name", http.StatusBadRequest)
				return
			}

			cloud, ok := spec["cloud"].(string)
			if !ok || cloud == "" {
				http.Error(w, "Missing cloud", http.StatusBadRequest)
				return
			}
			if cloud != "AWS" && cloud != "GCP" &&
				cloud != "AZURE" && cloud != "aws" && cloud != "gcp" && cloud != "azure" {
				http.Error(w, "Invalid cloud", http.StatusBadRequest)
				return
			}

			response := map[string]interface{}{
				"api_version": "ccpm/v1",
				"kind":        "CustomConnectPlugin",
				"id":          "ccp-123456",
				"metadata": map[string]interface{}{
					"created_at": "2023-01-01T00:00:00Z",
					"updated_at": "2023-01-01T00:00:00Z",
				},
				"spec": map[string]interface{}{
					"display_name":     displayName,
					"description":      spec["description"],
					"cloud":            cloud,
					"runtime_language": "JAVA",
					"environment":      spec["environment"],
				},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func handleCCPMPluginId(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		pluginId := vars["id"]

		switch r.Method {
		case http.MethodGet:
			// Describe plugin
			environment := r.URL.Query().Get("environment")

			if pluginId == "invalid-id" {
				http.Error(w, "Plugin not found", http.StatusNotFound)
				return
			}

			// Check if this is an error test case
			if environment == "env-error" {
				// This is the error handling test case - return error
				http.Error(w, "API Error", http.StatusInternalServerError)
				return
			}

			response := map[string]interface{}{
				"api_version": "ccpm/v1",
				"kind":        "CustomConnectPlugin",
				"id":          pluginId,
				"metadata": map[string]interface{}{
					"created_at": "2023-01-01T00:00:00Z",
					"updated_at": "2023-01-01T00:00:00Z",
				},
				"spec": map[string]interface{}{
					"display_name":     "CliPluginTest",
					"description":      "Test plugin",
					"cloud":            "AWS",
					"runtime_language": "JAVA",
					"environment": map[string]interface{}{
						"id": environment,
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)

		case http.MethodPatch:
			// Update plugin
			if pluginId == "invalid-id" {
				http.Error(w, "Plugin not found", http.StatusNotFound)
				return
			}

			var request map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			// Extract environment from request body for PATCH
			var environment string
			if spec, ok := request["spec"].(map[string]interface{}); ok {
				if env, ok := spec["environment"].(map[string]interface{}); ok {
					if envId, ok := env["id"].(string); ok {
						environment = envId
					}
				}
			}

			// Check if this is an error test case
			if environment == "env-error" {
				// This is the error handling test case - return error
				http.Error(w, "API Error", http.StatusInternalServerError)
				return
			}

			response := map[string]interface{}{
				"api_version": "ccpm/v1",
				"kind":        "CustomConnectPlugin",
				"id":          pluginId,
				"metadata": map[string]interface{}{
					"created_at": "2023-01-01T00:00:00Z",
					"updated_at": "2023-01-01T00:00:00Z",
				},
				"spec": map[string]interface{}{
					"display_name":     "Updated Plugin Name",
					"description":      "Updated description",
					"cloud":            "AWS",
					"runtime_language": "JAVA",
					"environment": map[string]interface{}{
						"id": environment,
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)

		case http.MethodDelete:
			// Delete plugin
			environment := r.URL.Query().Get("environment")

			if pluginId == "invalid-id" {
				http.Error(w, "Plugin not found", http.StatusNotFound)
				return
			}

			// Check if this is an error test case
			if environment == "env-error" {
				// This is the error handling test case - return error
				http.Error(w, "API Error", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// CCPM Plugin Version handlers
func handleCCPMPluginVersions(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		pluginId := vars["plugin_id"]
		environment := r.URL.Query().Get("environment")

		switch r.Method {
		case http.MethodGet:
			// List versions
			if pluginId == "invalid-id" {
				http.Error(w, "Plugin not found", http.StatusNotFound)
				return
			}

			versions := []map[string]interface{}{
				{
					"api_version": "ccpm/v1",
					"kind":        "CustomConnectPluginVersion",
					"id":          "ver-123456",
					"metadata": map[string]interface{}{
						"created_at": "2023-01-01T00:00:00Z",
						"updated_at": "2023-01-01T00:00:00Z",
					},
					"spec": map[string]interface{}{
						"version":                     "1.0.0",
						"content_format":              "ZIP",
						"documentation_link":          "https://docs.confluent.io",
						"sensitive_config_properties": []string{"password", "secret"},
						"environment": map[string]interface{}{
							"id": environment,
						},
						"connector_classes": []map[string]interface{}{
							{
								"class_name": "io.confluent.kafka.connect.datagen.DatagenConnector",
								"type":       "SOURCE",
							},
						},
					},
					"status": map[string]interface{}{
						"phase": "READY",
					},
				},
				{
					"api_version": "ccpm/v1",
					"kind":        "CustomConnectPluginVersion",
					"id":          "ver-789012",
					"metadata": map[string]interface{}{
						"created_at": "2023-01-01T00:00:00Z",
						"updated_at": "2023-01-01T00:00:00Z",
					},
					"spec": map[string]interface{}{
						"version":                     "2.0.0",
						"content_format":              "JAR",
						"documentation_link":          "",
						"sensitive_config_properties": []string{},
						"environment": map[string]interface{}{
							"id": environment,
						},
						"connector_classes": []map[string]interface{}{
							{
								"class_name": "io.confluent.kafka.connect.datagen.DatagenConnector",
								"type":       "SOURCE",
							},
						},
					},
					"status": map[string]interface{}{
						"phase": "READY",
					},
				},
			}

			response := map[string]interface{}{
				"data": versions,
				"metadata": map[string]interface{}{
					"next": "",
				},
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)

		case http.MethodPost:
			// Create version
			if pluginId == "invalid-id" {
				http.Error(w, "Plugin not found", http.StatusNotFound)
				return
			}

			var request map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			// Validate required fields
			spec, ok := request["spec"].(map[string]interface{})
			if !ok {
				http.Error(w, "Missing spec", http.StatusBadRequest)
				return
			}

			version, ok := spec["version"].(string)
			if !ok || version == "" {
				http.Error(w, "Missing version", http.StatusBadRequest)
				return
			}
			// class
			connectorClasses, ok := spec["connector_classes"].([]interface{})
			if !ok || len(connectorClasses) == 0 {
				http.Error(w, "Missing connector_classes", http.StatusBadRequest)
				return
			}

			response := map[string]interface{}{
				"api_version": "ccpm/v1",
				"kind":        "CustomConnectPluginVersion",
				"id":          "ver-123456",
				"metadata": map[string]interface{}{
					"created_at": "2023-01-01T00:00:00Z",
					"updated_at": "2023-01-01T00:00:00Z",
				},
				"spec": map[string]interface{}{
					"version":                     version,
					"content_format":              "ZIP",
					"documentation_link":          spec["documentation_link"],
					"sensitive_config_properties": spec["sensitive_config_properties"],
					"environment":                 spec["environment"],
					"connector_classes":           connectorClasses,
				},
				"status": map[string]interface{}{
					"phase": "READY",
				},
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func handleCCPMPluginVersionId(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		versionId := vars["version_id"]
		environment := r.URL.Query().Get("environment")

		switch r.Method {
		case http.MethodGet:
			// Describe version
			if versionId == "invalid-version" {
				http.Error(w, "Version not found", http.StatusNotFound)
				return
			}

			response := map[string]interface{}{
				"api_version": "ccpm/v1",
				"kind":        "CustomConnectPluginVersion",
				"id":          versionId,
				"metadata": map[string]interface{}{
					"created_at": "2023-01-01T00:00:00Z",
					"updated_at": "2023-01-01T00:00:00Z",
				},
				"spec": map[string]interface{}{
					"version":                     "1.0.0",
					"content_format":              "ZIP",
					"documentation_link":          "https://docs.confluent.io",
					"sensitive_config_properties": []string{"password", "secret"},
					"environment": map[string]interface{}{
						"id": environment,
					},
					"connector_classes": []map[string]interface{}{
						{
							"class_name": "io.confluent.kafka.connect.datagen.DatagenConnector",
							"type":       "SOURCE",
						},
					},
				},
				"status": map[string]interface{}{
					"phase": "READY",
				},
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)

		case http.MethodDelete:
			// Delete version
			if versionId == "invalid-version" {
				http.Error(w, "Version not found", http.StatusNotFound)
				return
			}

			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// CCPM Presigned URL handler
func handleCCPMPresignedUrl(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			// Create presigned URL
			var request map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
			// get environment from request body
			environment, ok := request["environment"].(map[string]interface{})
			if !ok || environment == nil {
				http.Error(w, "Missing environment", http.StatusBadRequest)
				return
			}

			response := map[string]interface{}{
				"api_version":    ccpmv1.PtrString("ccpm/v1"),
				"kind":           ccpmv1.PtrString("PresignedUrl"),
				"content_format": ccpmv1.PtrString("ZIP"),
				"cloud":          ccpmv1.PtrString("AWS"),
				"upload_id":      ccpmv1.PtrString("e53bb2e8-8de3-49fa-9fb1-4e3fd9a16b66"),
				"upload_url":     ccpmv1.PtrString(fmt.Sprintf("%s/connect/v1/dummy-presigned-url", TestV2CloudUrl.String())),
				"upload_form_data": map[string]interface{}{
					"bucket":               "confluent-custom-connectors-stag-us-west-2",
					"key":                  "staging/custom-plugin/2f37f0b6-f8da-4e8b-bc5f-282ebb0511be/connect-e53bb2e8-8de3-49fa-9fb1-4e3fd9a16b66/plugin.zip",
					"policy":               "string",
					"x-amz-algorithm":      "AWS4-HMAC-SHA256",
					"x-amz-credential":     "string",
					"x-amz-date":           "20230725T013857Z",
					"x-amz-security-token": "string",
					"x-amz-signature":      "string",
				},
				"environment": environment,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func handleCCPMUploadFile(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			err := json.NewEncoder(w).Encode(ccpmv1.PtrString("Success"))
			require.NoError(t, err)
		}
	}
}
