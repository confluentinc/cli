package testserver

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"
)

// Handler for: "/networking/v1/networks/{id}"
func handleNetworkingNetwork(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		switch r.Method {
		case http.MethodGet:
			handleNetworkingNetworkGet(t, id)(w, r)
		case http.MethodDelete:
			handleNetworkingNetworkDelete(t, id)(w, r)
		case http.MethodPatch:
			handleNetworkingNetworkUpdate(t, id)(w, r)
		}
	}
}

func handleNetworkingNetworks(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleNetworkingNetworkList(t)(w, r)
		}
	}
}

func handleNetworkingNetworkGet(t *testing.T, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch id {
		case "n-invalid":
			w.WriteHeader(http.StatusNotFound)
			err := writeErrorJson(w, "The network n-invalid was not found.")
			require.NoError(t, err)
		case "n-abcde1":
			network := networkingv1.NetworkingV1Network{
				Id: networkingv1.PtrString("n-abcde1"),
				Spec: &networkingv1.NetworkingV1NetworkSpec{
					Environment: &networkingv1.ObjectReference{Id: "env-00000"},
					DisplayName: networkingv1.PtrString("prod-aws-us-east1"),
					Cloud:       networkingv1.PtrString("AWS"),
					Region:      networkingv1.PtrString("us-east-1"),
					Cidr:        networkingv1.PtrString("10.200.0.0/16"),
					Zones:       &[]string{"use1-az1", "use1-az2", "use1-az3"},
					DnsConfig:   &networkingv1.NetworkingV1DnsConfig{Resolution: "CHASED_PRIVATE"},
				},
				Status: &networkingv1.NetworkingV1NetworkStatus{
					Phase:                 "READY",
					ActiveConnectionTypes: networkingv1.NetworkingV1ConnectionTypes{Items: []string{"PRIVATELINK", "TRANSITGATEWAY"}},
				},
			}
			err := json.NewEncoder(w).Encode(network)
			require.NoError(t, err)
		}
	}
}

func handleNetworkingNetworkDelete(t *testing.T, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch id {
		case "n-invalid":
			w.WriteHeader(http.StatusNotFound)
			err := writeErrorJson(w, "The network n-invalid was not found.")
			require.NoError(t, err)
		case "n-dependency":
			w.WriteHeader(http.StatusConflict)
			err := writeErrorJson(w, "Network deletion not allowed due to existing dependencies. Please delete the following dependent resources before attempting to delete the network again: pla-1abcde")
			require.NoError(t, err)
		case "n-abcde1":
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func handleNetworkingNetworkUpdate(t *testing.T, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch id {
		case "n-invalid":
			w.WriteHeader(http.StatusNotFound)
			err := writeErrorJson(w, "The network n-invalid was not found.")
			require.NoError(t, err)
		case "n-abcde1":
			network := networkingv1.NetworkingV1Network{
				Id: networkingv1.PtrString("n-abcde1"),
				Spec: &networkingv1.NetworkingV1NetworkSpec{
					Environment: &networkingv1.ObjectReference{Id: "env-00000"},
					DisplayName: networkingv1.PtrString("new-prod-aws-us-east1"),
					Cloud:       networkingv1.PtrString("AWS"),
					Region:      networkingv1.PtrString("us-east-1"),
					Cidr:        networkingv1.PtrString("10.200.0.0/16"),
					Zones:       &[]string{"use1-az1", "use1-az2", "use1-az3"},
					DnsConfig:   &networkingv1.NetworkingV1DnsConfig{Resolution: "CHASED_PRIVATE"},
				},
				Status: &networkingv1.NetworkingV1NetworkStatus{
					Phase:                 "READY",
					ActiveConnectionTypes: networkingv1.NetworkingV1ConnectionTypes{Items: []string{"PRIVATELINK", "TRANSITGATEWAY"}},
				},
			}
			err := json.NewEncoder(w).Encode(network)
			require.NoError(t, err)
		}
	}
}

func handleNetworkingNetworkList(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		awsNetwork := networkingv1.NetworkingV1Network{
			Id: networkingv1.PtrString("n-abcde1"),
			Spec: &networkingv1.NetworkingV1NetworkSpec{
				Environment: &networkingv1.ObjectReference{Id: "env-00000"},
				DisplayName: networkingv1.PtrString("prod-aws-us-east1"),
				Cloud:       networkingv1.PtrString("AWS"),
				Region:      networkingv1.PtrString("us-east-1"),
				Cidr:        networkingv1.PtrString("10.200.0.0/16"),
				Zones:       &[]string{"use1-az1", "use1-az2", "use1-az3"},
				DnsConfig:   &networkingv1.NetworkingV1DnsConfig{Resolution: "CHASED_PRIVATE"},
			},
			Status: &networkingv1.NetworkingV1NetworkStatus{
				Phase:                 "READY",
				ActiveConnectionTypes: networkingv1.NetworkingV1ConnectionTypes{Items: []string{"PRIVATELINK", "TRANSITGATEWAY"}},
			},
		}
		gcpNetwork := networkingv1.NetworkingV1Network{
			Id: networkingv1.PtrString("n-abcde2"),
			Spec: &networkingv1.NetworkingV1NetworkSpec{
				Environment: &networkingv1.ObjectReference{Id: "env-00000"},
				DisplayName: networkingv1.PtrString("prod-gcp-us-central1"),
				Cloud:       networkingv1.PtrString("GCP"),
				Region:      networkingv1.PtrString("us-central1"),
				Cidr:        networkingv1.PtrString("10.1.0.0/16"),
				Zones:       &[]string{"us-central1-a", "us-central1-b", "us-central1-c"},
			},
			Status: &networkingv1.NetworkingV1NetworkStatus{
				Phase:                 "READY",
				ActiveConnectionTypes: networkingv1.NetworkingV1ConnectionTypes{Items: []string{"PRIVATELINK"}},
			},
		}
		azureNetwork := networkingv1.NetworkingV1Network{
			Id: networkingv1.PtrString("n-abcde3"),
			Spec: &networkingv1.NetworkingV1NetworkSpec{
				Environment: &networkingv1.ObjectReference{Id: "env-00000"},
				DisplayName: networkingv1.PtrString("prod-azure-eastus2"),
				Cloud:       networkingv1.PtrString("AZURE"),
				Region:      networkingv1.PtrString("eastus2"),
				Cidr:        networkingv1.PtrString("10.0.0.0/16"),
				Zones:       &[]string{"1", "2", "3"},
			},
			Status: &networkingv1.NetworkingV1NetworkStatus{
				Phase:                 "READY",
				ActiveConnectionTypes: networkingv1.NetworkingV1ConnectionTypes{Items: []string{}},
			},
		}

		networkList := networkingv1.NetworkingV1NetworkList{Data: []networkingv1.NetworkingV1Network{awsNetwork, gcpNetwork, azureNetwork}}
		err := json.NewEncoder(w).Encode(networkList)
		require.NoError(t, err)
	}
}
