package testserver

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	pi "github.com/confluentinc/ccloud-sdk-go-v2/provider-integration/v1"
)

const (
	AwsIntegrationConfig = "AwsIntegrationConfig"
)

// Handler for "/pim/v1/integrations/{id}"
func handleProviderIntegration(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		switch r.Method {
		case http.MethodGet:
			handleProviderIntegrationGet(t, id)(w, r)
		case http.MethodDelete:
			handleProviderIntegrationDelete(t)(w, r)
		}
	}
}

// Handler for "/pim/v1/integrations"
func handleProviderIntegrations(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handleProviderIntegrationCreate(t)(w, r)
		case http.MethodGet:
			handleProviderIntegrationList(t)(w, r)
		}
	}
}

func handleProviderIntegrationGet(t *testing.T, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch id {
		case "cspi-42o61":
			mockResponse := pi.PimV1Integration{
				Id:          pi.PtrString(id),
				DisplayName: pi.PtrString("cdong-test2"),
				Provider:    pi.PtrString("aws"),
				Environment: &pi.GlobalObjectReference{Id: "env-9zgy77"},
				Config:      getPimV1IntegrationConfig("aws"),
			}
			err := json.NewEncoder(w).Encode(mockResponse)
			require.NoError(t, err)
		case "cspi-4xgwq":
			mockResponse := pi.PimV1Integration{
				Id:          pi.PtrString(id),
				DisplayName: pi.PtrString("cdong-test2"),
				Provider:    pi.PtrString("aws"),
				Environment: &pi.GlobalObjectReference{Id: "env-9zgy77"},
				Config:      getPimV1IntegrationConfig("aws"),
			}
			err := json.NewEncoder(w).Encode(mockResponse)
			require.NoError(t, err)
		case "cspi-invalid":
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func handleProviderIntegrationDelete(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(pi.PimV1Integration{})
		require.NoError(t, err)
	}
}

func handleProviderIntegrationList(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		configs := getPimV1IntegrationConfigList()
		require.Len(t, configs, 2, "expected 2 configs, but got a different length")
		mockResponse := []pi.PimV1Integration{
			{
				Id:          pi.PtrString("cspi-42o61"),
				DisplayName: pi.PtrString("cdong-test2"),
				Provider:    pi.PtrString("aws"),
				Environment: &pi.GlobalObjectReference{Id: "env-9zgy77"},
				Config:      configs[0],
				Usages:      &[]string{},
			},
			{
				Id:          pi.PtrString("cspi-4xgwq"),
				DisplayName: pi.PtrString("cdong-test"),
				Provider:    pi.PtrString("aws"),
				Environment: &pi.GlobalObjectReference{Id: "env-9zgy77"},
				Config:      configs[1],
				Usages:      &[]string{},
			},
		}
		result := pi.PimV1IntegrationList{Data: mockResponse}
		setPageToken(&result, &result.Metadata, r.URL)
		err := json.NewEncoder(w).Encode(result)
		require.NoError(t, err)
	}
}

func handleProviderIntegrationCreate(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var mockRequest pi.PimV1Integration
		require.NoError(t, json.NewDecoder(r.Body).Decode(&mockRequest))

		provider := mockRequest.GetProvider()
		mockResponse := &pi.PimV1Integration{
			Id:          pi.PtrString("cspi-42o61"),
			DisplayName: pi.PtrString("cdong-test2"),
			Provider:    pi.PtrString(provider),
			Environment: &pi.GlobalObjectReference{Id: "env-9zgy77"},
			Config:      getPimV1IntegrationConfig(provider),
		}

		err := json.NewEncoder(w).Encode(mockResponse)
		require.NoError(t, err)
	}
}

func getPimV1IntegrationConfig(provider string) *pi.PimV1IntegrationConfigOneOf {
	provider = strings.ToLower(provider)
	switch provider {
	case "aws":
		return &pi.PimV1IntegrationConfigOneOf{
			PimV1AwsIntegrationConfig: &pi.PimV1AwsIntegrationConfig{
				IamRoleArn:         pi.PtrString("arn:aws:iam::851725421142:role/cspi-42o61"),
				ExternalId:         pi.PtrString("999219d4-37f4-49ac-abfe-d2b6528fb21b"),
				CustomerIamRoleArn: pi.PtrString("arn:aws:iam::037803949979:role/tarun-iam-test-role"),
				Kind:               AwsIntegrationConfig,
			},
		}
	default:
		return &pi.PimV1IntegrationConfigOneOf{}
	}
}

func getPimV1IntegrationConfigList() []*pi.PimV1IntegrationConfigOneOf {
	return []*pi.PimV1IntegrationConfigOneOf{
		{
			PimV1AwsIntegrationConfig: &pi.PimV1AwsIntegrationConfig{
				IamRoleArn:         pi.PtrString("arn:aws:iam::851725421142:role/cspi-42o61"),
				ExternalId:         pi.PtrString("999219d4-37f4-49ac-abfe-d2b6528fb21b"),
				CustomerIamRoleArn: pi.PtrString("arn:aws:iam::037803949979:role/tarun-iam-test-role"),
				Kind:               AwsIntegrationConfig,
			},
		},
		{
			PimV1AwsIntegrationConfig: &pi.PimV1AwsIntegrationConfig{
				IamRoleArn:         pi.PtrString("arn:aws:iam::851725421142:role/cspi-4xgwq"),
				ExternalId:         pi.PtrString("b5f0a7e6-15e3-4879-aa01-881910554cd4"),
				CustomerIamRoleArn: pi.PtrString("arn:aws:iam::000000000000:role/my-test-aws-role"),
				Kind:               AwsIntegrationConfig,
			},
		},
	}
}
