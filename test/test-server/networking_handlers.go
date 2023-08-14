package testserver

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

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
				Metadata: &networkingv1.ObjectMeta{
					Self:         "https://api.confluent.cloud/networking/v1/networks/n-abcde1",
					ResourceName: networkingv1.PtrString("crn://confluent.cloud/organization=9bb441c4-edef-46ac-8a41-c49e44a3fd9a/environment=env-00000/network=n-abcde1"),
					CreatedAt:    networkingv1.PtrTime(time.Date(2023, time.February, 24, 0, 0, 0, 0, time.UTC)),
					UpdatedAt:    networkingv1.PtrTime(time.Date(2023, time.February, 24, 0, 0, 0, 0, time.UTC)),
					DeletedAt:    networkingv1.PtrTime(time.Date(2023, time.February, 24, 0, 0, 0, 0, time.UTC)),
				},
				Spec: &networkingv1.NetworkingV1NetworkSpec{
					Environment:     &networkingv1.ObjectReference{Id: "env-00000"},
					DisplayName:     networkingv1.PtrString("prod-aws-us-east1"),
					Cloud:           networkingv1.PtrString("AWS"),
					Region:          networkingv1.PtrString("us-east-1"),
					ConnectionTypes: &networkingv1.NetworkingV1ConnectionTypes{Items: []string{"PRIVATELINK"}},
					Cidr:            networkingv1.PtrString("10.200.0.0/16"),
					Zones:           &[]string{"use1-az1", "use1-az2", "use1-az3"},
					ZonesInfo: &networkingv1.NetworkingV1ZonesInfo{Items: []networkingv1.NetworkingV1ZoneInfo{
						{ZoneId: networkingv1.PtrString("use1-az1"), Cidr: networkingv1.PtrString("10.20.0.0/27")},
						{ZoneId: networkingv1.PtrString("use1-az2"), Cidr: networkingv1.PtrString("10.20.0.0/27")},
						{ZoneId: networkingv1.PtrString("use1-az3"), Cidr: networkingv1.PtrString("10.20.0.0/27")},
					}},
					DnsConfig:    &networkingv1.NetworkingV1DnsConfig{Resolution: "CHASED_PRIVATE"},
					ReservedCidr: networkingv1.PtrString("172.20.255.0/24"),
				},
				Status: &networkingv1.NetworkingV1NetworkStatus{
					Phase:                    "READY",
					SupportedConnectionTypes: networkingv1.NetworkingV1SupportedConnectionTypes{Items: []string{"PRIVATELINK"}},
					ActiveConnectionTypes:    networkingv1.NetworkingV1ConnectionTypes{Items: []string{"PRIVATELINK"}},
				},
			}
			err := json.NewEncoder(w).Encode(network)
			require.NoError(t, err)
		}
	}
}
