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
