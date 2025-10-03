package testserver

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	piv2 "github.com/confluentinc/ccloud-sdk-go-v2/provider-integration/v2"
)

const (
	AzureIntegrationConfig = "AzureIntegrationConfig"
	GcpIntegrationConfig   = "GcpIntegrationConfig"
)

// Handler for "/pim/v2/integrations/{id}"
func handleProviderIntegrationV2(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		switch r.Method {
		case http.MethodGet:
			handleProviderIntegrationV2Get(t, id)(w, r)
		case http.MethodDelete:
			handleProviderIntegrationV2Delete(t)(w, r)
		case http.MethodPatch:
			handleProviderIntegrationV2Update(t)(w, r)
		}
	}
}

// Handler for "/pim/v2/integrations"
func handleProviderIntegrationsV2(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handleProviderIntegrationV2Create(t)(w, r)
		case http.MethodGet:
			handleProviderIntegrationV2List(t)(w, r)
		}
	}
}

// Handler for "/pim/v2/integrations/{id}/validate"
func handleProviderIntegrationV2Validate(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		switch id {
		case "pi-123456":
			// Azure validation success
			w.WriteHeader(http.StatusOK)
		case "pi-789012":
			// GCP validation success
			w.WriteHeader(http.StatusOK)
		default:
			// Validation failure
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error_code": "400",
				"message":    "Cloud provider setup incomplete",
			})
		}
	}
}

func handleProviderIntegrationV2Get(t *testing.T, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch id {
		case "pi-123456":
			mockResponse := piv2.PimV2Integration{
				Id:          piv2.PtrString(id),
				DisplayName: piv2.PtrString("azure-test"),
				Provider:    piv2.PtrString("azure"),
				Environment: &piv2.ObjectReference{Id: "env-596"},
				Status:      piv2.PtrString("DRAFT"),
				Config: &piv2.PimV2IntegrationConfigOneOf{
					PimV2AzureIntegrationConfig: &piv2.PimV2AzureIntegrationConfig{
						Kind:                      AzureIntegrationConfig,
						CustomerAzureTenantId:     piv2.PtrString("00000000-0000-0000-0000-000000000000"),
						ConfluentMultiTenantAppId: piv2.PtrString("app-123456789"),
					},
				},
			}
			err := json.NewEncoder(w).Encode(mockResponse)
			require.NoError(t, err)
		case "pi-789012":
			mockResponse := piv2.PimV2Integration{
				Id:          piv2.PtrString(id),
				DisplayName: piv2.PtrString("gcp-test"),
				Provider:    piv2.PtrString("gcp"),
				Environment: &piv2.ObjectReference{Id: "env-596"},
				Status:      piv2.PtrString("DRAFT"),
				Config: &piv2.PimV2IntegrationConfigOneOf{
					PimV2GcpIntegrationConfig: &piv2.PimV2GcpIntegrationConfig{
						Kind:                         GcpIntegrationConfig,
						CustomerGoogleServiceAccount: piv2.PtrString("my-service-account@my-project.iam.gserviceaccount.com"),
						GoogleServiceAccount:         piv2.PtrString("confluent-sa-123456789@gcp-sa-cloud.iam.gserviceaccount.com"),
					},
				},
			}
			err := json.NewEncoder(w).Encode(mockResponse)
			require.NoError(t, err)
		case "pi-invalid":
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func handleProviderIntegrationV2Delete(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleProviderIntegrationV2List(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mockResponse := piv2.PimV2IntegrationList{
			Data: []piv2.PimV2Integration{
				{
					Id:          piv2.PtrString("pi-123456"),
					DisplayName: piv2.PtrString("azure-test"),
					Provider:    piv2.PtrString("azure"),
					Environment: &piv2.ObjectReference{Id: "env-596"},
					Status:      piv2.PtrString("DRAFT"),
					Config: &piv2.PimV2IntegrationConfigOneOf{
						PimV2AzureIntegrationConfig: &piv2.PimV2AzureIntegrationConfig{
							Kind:                      AzureIntegrationConfig,
							CustomerAzureTenantId:     piv2.PtrString("00000000-0000-0000-0000-000000000000"),
							ConfluentMultiTenantAppId: piv2.PtrString("app-123456789"),
						},
					},
				},
				{
					Id:          piv2.PtrString("pi-789012"),
					DisplayName: piv2.PtrString("gcp-test"),
					Provider:    piv2.PtrString("gcp"),
					Environment: &piv2.ObjectReference{Id: "env-596"},
					Status:      piv2.PtrString("DRAFT"),
					Config: &piv2.PimV2IntegrationConfigOneOf{
						PimV2GcpIntegrationConfig: &piv2.PimV2GcpIntegrationConfig{
							Kind:                         GcpIntegrationConfig,
							CustomerGoogleServiceAccount: piv2.PtrString("my-service-account@my-project.iam.gserviceaccount.com"),
							GoogleServiceAccount:         piv2.PtrString("confluent-sa-123456789@gcp-sa-cloud.iam.gserviceaccount.com"),
						},
					},
				},
			},
		}
		err := json.NewEncoder(w).Encode(mockResponse)
		require.NoError(t, err)
	}
}

func handleProviderIntegrationV2Create(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request piv2.PimV2Integration
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))

		provider := strings.ToLower(request.GetProvider())
		var id string
		displayName := request.GetDisplayName()
		
		switch provider {
		case "azure":
			if displayName == "azure-test" {
				id = "pi-123456"
			} else {
				id = "pi-azure-new"
			}
		case "gcp":
			if displayName == "gcp-test" {
				id = "pi-789012"
			} else {
				id = "pi-gcp-new"
			}
		default:
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error_code": "400",
				"message":    "Invalid provider. Must be 'azure' or 'gcp'.",
			})
			return
		}

		mockResponse := piv2.PimV2Integration{
			Id:          piv2.PtrString(id),
			DisplayName: request.DisplayName,
			Provider:    piv2.PtrString(provider),
			Environment: request.Environment,
			Status:      piv2.PtrString("DRAFT"),
		}

		err := json.NewEncoder(w).Encode(mockResponse)
		require.NoError(t, err)
	}
}

func handleProviderIntegrationV2Update(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		var request piv2.PimV2IntegrationUpdate
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))

		var mockResponse piv2.PimV2Integration
		switch id {
		case "pi-123456":
			mockResponse = piv2.PimV2Integration{
				Id:          piv2.PtrString(id),
				DisplayName: piv2.PtrString("azure-test"),
				Provider:    piv2.PtrString("azure"),
				Environment: request.Environment,
				Status:      piv2.PtrString("CREATED"),
				Config: &piv2.PimV2IntegrationConfigOneOf{
					PimV2AzureIntegrationConfig: &piv2.PimV2AzureIntegrationConfig{
						Kind:                      AzureIntegrationConfig,
						CustomerAzureTenantId:     request.Config.PimV2AzureIntegrationConfig.CustomerAzureTenantId,
						ConfluentMultiTenantAppId: piv2.PtrString("app-123456789"),
					},
				},
			}
		case "pi-789012":
			mockResponse = piv2.PimV2Integration{
				Id:          piv2.PtrString(id),
				DisplayName: piv2.PtrString("gcp-test"),
				Provider:    piv2.PtrString("gcp"),
				Environment: request.Environment,
				Status:      piv2.PtrString("CREATED"),
				Config: &piv2.PimV2IntegrationConfigOneOf{
					PimV2GcpIntegrationConfig: &piv2.PimV2GcpIntegrationConfig{
						Kind:                         GcpIntegrationConfig,
						CustomerGoogleServiceAccount: request.Config.PimV2GcpIntegrationConfig.CustomerGoogleServiceAccount,
						GoogleServiceAccount:         piv2.PtrString("confluent-sa-123456789@gcp-sa-cloud.iam.gserviceaccount.com"),
					},
				},
			}
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}

		err := json.NewEncoder(w).Encode(mockResponse)
		require.NoError(t, err)
	}
}
