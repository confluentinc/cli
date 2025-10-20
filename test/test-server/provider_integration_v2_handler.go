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
		var request piv2.PimV2IntegrationValidateRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))

		id := request.GetId()
		switch id {
		case "pi-not-configured":
			// Validation failure - setup incomplete
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"id":     "test-error-id",
						"status": "400",
						"code":   "bad_request",
						"detail": "Cloud provider setup incomplete",
						"source": map[string]interface{}{},
					},
				},
			})
		default:
			// Unknown integration
			w.WriteHeader(http.StatusNotFound)
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
				Status:      piv2.PtrString("CREATED"),
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
				Status:      piv2.PtrString("CREATED"),
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
		case "pi-not-configured":
			mockResponse := piv2.PimV2Integration{
				Id:          piv2.PtrString(id),
				DisplayName: piv2.PtrString("not-configured-test"),
				Provider:    piv2.PtrString("azure"),
				Environment: &piv2.ObjectReference{Id: "env-596"},
				Status:      piv2.PtrString("DRAFT"),
				Config: &piv2.PimV2IntegrationConfigOneOf{
					PimV2AzureIntegrationConfig: &piv2.PimV2AzureIntegrationConfig{
						Kind:                      AzureIntegrationConfig,
						ConfluentMultiTenantAppId: piv2.PtrString("app-not-configured"),
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
			} else if displayName == "atomic-test-invalid-azure" {
				id = "pi-atomic-azure"
			} else {
				id = "pi-azure-new"
			}
		case "gcp":
			if displayName == "gcp-test" {
				id = "pi-789012"
			} else if displayName == "atomic-test-invalid-gcp" {
				id = "pi-atomic-gcp"
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

		// Add Confluent-managed identity based on provider
		switch provider {
		case "azure":
			mockResponse.Config = &piv2.PimV2IntegrationConfigOneOf{
				PimV2AzureIntegrationConfig: &piv2.PimV2AzureIntegrationConfig{
					Kind:                      AzureIntegrationConfig,
					ConfluentMultiTenantAppId: piv2.PtrString("app-123456789"),
				},
			}
		case "gcp":
			mockResponse.Config = &piv2.PimV2IntegrationConfigOneOf{
				PimV2GcpIntegrationConfig: &piv2.PimV2GcpIntegrationConfig{
					Kind:                 GcpIntegrationConfig,
					GoogleServiceAccount: piv2.PtrString("confluent-sa-123456789@gcp-sa-cloud.iam.gserviceaccount.com"),
				},
			}
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

		// Handle invalid configuration test cases
		if request.Config != nil {
			if request.Config.PimV2AzureIntegrationConfig != nil {
				azureTenantId := request.Config.PimV2AzureIntegrationConfig.GetCustomerAzureTenantId()
				if azureTenantId == "invalid-uuid" || azureTenantId == "not-a-valid-uuid" {
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(map[string]interface{}{
						"errors": []map[string]interface{}{
							{
								"id":     "test-error-id",
								"status": "400",
								"code":   "bad_request",
								"detail": "invalid customer AZURE tenant id",
								"source": map[string]interface{}{},
							},
						},
					})
					return
				}
			}
			if request.Config.PimV2GcpIntegrationConfig != nil {
				gcpSA := request.Config.PimV2GcpIntegrationConfig.GetCustomerGoogleServiceAccount()
				if gcpSA == "invalid-format" {
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(map[string]interface{}{
						"errors": []map[string]interface{}{
							{
								"id":     "test-error-id",
								"status": "400",
								"code":   "bad_request",
								"detail": "invalid Google Service Account",
								"source": map[string]interface{}{},
							},
						},
					})
					return
				}
			}
		}

		// Handle non-existent integration
		if id == "pi-invalid" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var mockResponse piv2.PimV2Integration
		switch id {
		case "pi-123456":
			azureConfig := &piv2.PimV2AzureIntegrationConfig{
				Kind:                      AzureIntegrationConfig,
				ConfluentMultiTenantAppId: piv2.PtrString("app-123456789"),
			}
			// Add customer tenant ID if provided in request
			if request.Config != nil && request.Config.PimV2AzureIntegrationConfig != nil {
				azureConfig.CustomerAzureTenantId = request.Config.PimV2AzureIntegrationConfig.CustomerAzureTenantId
			}
			
			mockResponse = piv2.PimV2Integration{
				Id:          piv2.PtrString(id),
				DisplayName: piv2.PtrString("azure-test"),
				Provider:    piv2.PtrString("azure"),
				Environment: request.Environment,
				Status:      piv2.PtrString("CREATED"),
				Config: &piv2.PimV2IntegrationConfigOneOf{
					PimV2AzureIntegrationConfig: azureConfig,
				},
			}
		case "pi-789012":
			gcpConfig := &piv2.PimV2GcpIntegrationConfig{
				Kind:                 GcpIntegrationConfig,
				GoogleServiceAccount: piv2.PtrString("confluent-sa-123456789@gcp-sa-cloud.iam.gserviceaccount.com"),
			}
			// Add customer service account if provided in request
			if request.Config != nil && request.Config.PimV2GcpIntegrationConfig != nil {
				gcpConfig.CustomerGoogleServiceAccount = request.Config.PimV2GcpIntegrationConfig.CustomerGoogleServiceAccount
			}
			
			mockResponse = piv2.PimV2Integration{
				Id:          piv2.PtrString(id),
				DisplayName: piv2.PtrString("gcp-test"),
				Provider:    piv2.PtrString("gcp"),
				Environment: request.Environment,
				Status:      piv2.PtrString("CREATED"),
				Config: &piv2.PimV2IntegrationConfigOneOf{
					PimV2GcpIntegrationConfig: gcpConfig,
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
