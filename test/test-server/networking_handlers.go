package testserver

import (
	"encoding/json"
	"net/http"
	"slices"
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

// Handler for: "/networking/v1/networks"
func handleNetworkingNetworks(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleNetworkingNetworkList(t)(w, r)
		case http.MethodPost:
			handleNetworkingNetworkCreate(t)(w, r)
		}
	}
}

// Handler for "/networking/v1/peerings/{id}"
func handleNetworkingPeering(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		switch r.Method {
		case http.MethodGet:
			handleNetworkingPeeringGet(t, id)(w, r)
		case http.MethodPatch:
			handleNetworkingPeeringUpdate(t, id)(w, r)
		case http.MethodDelete:
			handleNetworkingPeeringDelete(t, id)(w, r)
		}
	}
}

// Handler for "/networking/v1/peerings"
func handleNetworkingPeerings(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleNetworkingPeeringList(t)(w, r)
		case http.MethodPost:
			handleNetworkingPeeringCreate(t)(w, r)
		}
	}
}

// Handler for "/networking/v1/transit-gateway-attachments/{id}"
func handleNetworkingTransitGatewayAttachment(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		switch r.Method {
		case http.MethodGet:
			handleNetworkingTransitGatewayAttachmentGet(t, id)(w, r)
		case http.MethodPatch:
			handleNetworkingTransitGatewayAttachmentUpdate(t, id)(w, r)
		case http.MethodDelete:
			handleNetworkingTransitGatewayAttachmentDelete(t, id)(w, r)
		}
	}
}

// Handler for "/networking/v1/transit-gateway-attachments"
func handleNetworkingTransitGatewayAttachments(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleNetworkingTransitGatewayAttachmentsList(t)(w, r)
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
			network := getAwsNetwork("n-abcde1", "prod-aws-us-east1", "READY")
			err := json.NewEncoder(w).Encode(network)
			require.NoError(t, err)
		case "n-abcde2":
			network := getGcpNetwork("n-abcde2", "prod-gcp-us-central1", "READY")
			err := json.NewEncoder(w).Encode(network)
			require.NoError(t, err)
		case "n-abcde3":
			network := getAzureNetwork("n-abcde3", "prod-azure-eastus2", "READY")
			err := json.NewEncoder(w).Encode(network)
			require.NoError(t, err)
		case "n-abcde4":
			network := getAwsNetwork("n-abcde4", "prod-aws-us-east1", "PROVISIONING")
			err := json.NewEncoder(w).Encode(network)
			require.NoError(t, err)
		case "n-abcde5":
			network := getGcpNetwork("n-abcde5", "prod-gcp-us-central1", "PROVISIONING")
			err := json.NewEncoder(w).Encode(network)
			require.NoError(t, err)
		case "n-abcde6":
			network := getAzureNetwork("n-abcde6", "prod-azure-eastus2", "PROVISIONING")
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
		case "n-abcde1", "n-abcde2":
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
			network := getAwsNetwork("n-abcde1", "new-prod-aws-us-east1", "READY")
			err := json.NewEncoder(w).Encode(network)
			require.NoError(t, err)
		}
	}
}

func handleNetworkingNetworkList(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		awsNetwork := getAwsNetwork("n-abcde1", "prod-aws-us-east1", "READY")
		gcpNetwork := getGcpNetwork("n-abcde2", "prod-gcp-us-central1", "READY")
		azureNetwork := getAzureNetwork("n-abcde3", "prod-azure-eastus2", "READY")

		pageToken := r.URL.Query().Get("page_token")
		var networkList networkingv1.NetworkingV1NetworkList
		switch pageToken {
		case "azure":
			networkList = networkingv1.NetworkingV1NetworkList{
				Data:     []networkingv1.NetworkingV1Network{azureNetwork},
				Metadata: networkingv1.ListMeta{},
			}
		case "gcp":
			networkList = networkingv1.NetworkingV1NetworkList{
				Data:     []networkingv1.NetworkingV1Network{gcpNetwork},
				Metadata: networkingv1.ListMeta{Next: *networkingv1.NewNullableString(networkingv1.PtrString("/networking/v1/networks?environment=a-595&page_size=1&page_token=azure"))},
			}
		default:
			networkList = networkingv1.NetworkingV1NetworkList{
				Data:     []networkingv1.NetworkingV1Network{awsNetwork},
				Metadata: networkingv1.ListMeta{Next: *networkingv1.NewNullableString(networkingv1.PtrString("/networking/v1/networks?environment=a-595&page_size=1&page_token=gcp"))},
			}
		}

		err := json.NewEncoder(w).Encode(networkList)
		require.NoError(t, err)
	}
}

func handleNetworkingNetworkCreate(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := &networkingv1.NetworkingV1Network{}
		err := json.NewDecoder(r.Body).Decode(body)
		require.NoError(t, err)

		connectionTypes := body.Spec.ConnectionTypes.Items

		if slices.Contains(connectionTypes, "TRANSITGATEWAY") && (body.Spec.Cidr == nil && body.Spec.ZonesInfo == nil) {
			w.WriteHeader(http.StatusBadRequest)
			err := writeErrorJson(w, "A cidr must be provided when using TRANSITGATEWAY.")
			require.NoError(t, err)
		} else {
			network := &networkingv1.NetworkingV1Network{
				Id: networkingv1.PtrString("n-abcde1"),
				Spec: &networkingv1.NetworkingV1NetworkSpec{
					Environment: &networkingv1.ObjectReference{Id: body.Spec.Environment.Id},
					DisplayName: body.Spec.DisplayName,
					Cloud:       body.Spec.Cloud,
					Region:      body.Spec.Region,
				},
				Status: &networkingv1.NetworkingV1NetworkStatus{
					Phase:                    "PROVISIONING",
					SupportedConnectionTypes: networkingv1.NetworkingV1SupportedConnectionTypes{Items: connectionTypes},
					ActiveConnectionTypes:    networkingv1.NetworkingV1ConnectionTypes{Items: []string{}},
				},
			}

			if body.Spec.Zones != nil {
				network.Spec.SetZones(*body.Spec.Zones)
			}

			if body.Spec.DnsConfig != nil && body.Spec.DnsConfig.Resolution != "" {
				network.Spec.SetDnsConfig(networkingv1.NetworkingV1DnsConfig{Resolution: body.Spec.DnsConfig.Resolution})
			}

			if body.Spec.ZonesInfo != nil {
				network.Spec.SetZonesInfo(*body.Spec.ZonesInfo)
			}

			if body.Spec.Cidr != nil {
				network.Spec.SetCidr(*body.Spec.Cidr)
			}

			if body.Spec.ReservedCidr != nil {
				network.Spec.SetReservedCidr(*body.Spec.ReservedCidr)
			}

			err = json.NewEncoder(w).Encode(network)
			require.NoError(t, err)
		}
	}
}

func getAwsNetwork(id, name, phase string) networkingv1.NetworkingV1Network {
	network := networkingv1.NetworkingV1Network{
		Id: networkingv1.PtrString(id),
		Spec: &networkingv1.NetworkingV1NetworkSpec{
			Environment: &networkingv1.ObjectReference{Id: "env-00000"},
			DisplayName: networkingv1.PtrString(name),
			Cloud:       networkingv1.PtrString("AWS"),
			Region:      networkingv1.PtrString("us-east-1"),
			Cidr:        networkingv1.PtrString("10.200.0.0/16"),
			Zones:       &[]string{"use1-az1", "use1-az2", "use1-az3"},
			DnsConfig:   &networkingv1.NetworkingV1DnsConfig{Resolution: "CHASED_PRIVATE"},
		},
		Status: &networkingv1.NetworkingV1NetworkStatus{
			Phase:                    phase,
			SupportedConnectionTypes: networkingv1.NetworkingV1SupportedConnectionTypes{Items: []string{"PRIVATELINK", "TRANSITGATEWAY"}},
			ActiveConnectionTypes:    networkingv1.NetworkingV1ConnectionTypes{Items: []string{}},
		},
	}

	if phase == "READY" {
		network.Status.ActiveConnectionTypes = networkingv1.NetworkingV1ConnectionTypes{Items: []string{"PRIVATELINK", "TRANSITGATEWAY"}}
		network.Status.Cloud = &networkingv1.NetworkingV1NetworkStatusCloudOneOf{
			NetworkingV1AwsNetwork: &networkingv1.NetworkingV1AwsNetwork{
				Kind:    "AwsNetwork",
				Vpc:     "vpc-00000000000000000",
				Account: "000000000000",
			},
		}
	}

	return network
}

func getGcpNetwork(id, name, phase string) networkingv1.NetworkingV1Network {
	network := networkingv1.NetworkingV1Network{
		Id: networkingv1.PtrString(id),
		Spec: &networkingv1.NetworkingV1NetworkSpec{
			Environment: &networkingv1.ObjectReference{Id: "env-00000"},
			DisplayName: networkingv1.PtrString(name),
			Cloud:       networkingv1.PtrString("GCP"),
			Region:      networkingv1.PtrString("us-central1"),
			Cidr:        networkingv1.PtrString("10.1.0.0/16"),
			Zones:       &[]string{"us-central1-a", "us-central1-b", "us-central1-c"},
		},
		Status: &networkingv1.NetworkingV1NetworkStatus{
			Phase:                    phase,
			SupportedConnectionTypes: networkingv1.NetworkingV1SupportedConnectionTypes{Items: []string{"PRIVATELINK"}},
			ActiveConnectionTypes:    networkingv1.NetworkingV1ConnectionTypes{Items: []string{}},
		},
	}

	if phase == "READY" {
		network.Status.ActiveConnectionTypes = networkingv1.NetworkingV1ConnectionTypes{Items: []string{"PRIVATELINK"}}
		network.Status.Cloud = &networkingv1.NetworkingV1NetworkStatusCloudOneOf{
			NetworkingV1GcpNetwork: &networkingv1.NetworkingV1GcpNetwork{
				Kind:       "GcpNetwork",
				Project:    "gcp-project",
				VpcNetwork: "gcp-vpc",
			},
		}
	}

	return network
}

func getAzureNetwork(id, name, phase string) networkingv1.NetworkingV1Network {
	network := networkingv1.NetworkingV1Network{
		Id: networkingv1.PtrString(id),
		Spec: &networkingv1.NetworkingV1NetworkSpec{
			Environment: &networkingv1.ObjectReference{Id: "env-00000"},
			DisplayName: networkingv1.PtrString(name),
			Cloud:       networkingv1.PtrString("AZURE"),
			Region:      networkingv1.PtrString("eastus2"),
			Cidr:        networkingv1.PtrString("10.0.0.0/16"),
			Zones:       &[]string{"1", "2", "3"},
		},
		Status: &networkingv1.NetworkingV1NetworkStatus{
			Phase:                    phase,
			SupportedConnectionTypes: networkingv1.NetworkingV1SupportedConnectionTypes{Items: []string{"PEERING"}},
			ActiveConnectionTypes:    networkingv1.NetworkingV1ConnectionTypes{Items: []string{}},
		},
	}

	if phase == "READY" {
		network.Status.ActiveConnectionTypes = networkingv1.NetworkingV1ConnectionTypes{Items: []string{"PEERING"}}
		network.Status.Cloud = &networkingv1.NetworkingV1NetworkStatusCloudOneOf{
			NetworkingV1AzureNetwork: &networkingv1.NetworkingV1AzureNetwork{
				Kind:         "AzureNetwork",
				Vnet:         "azure-vnet",
				Subscription: "aa000000-a000-0a00-00aa-0000aaa0a0a0",
			},
		}
	}
	return network
}

func handleNetworkingPeeringList(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		awsPeering := getPeering("peer-111111", "aws-peering", "AWS")
		gcpPeering := getPeering("peer-111112", "gcp-peering", "GCP")
		azurePeering := getPeering("peer-111113", "azure-peering", "AZURE")

		pageToken := r.URL.Query().Get("page_token")
		var peeringList networkingv1.NetworkingV1PeeringList
		switch pageToken {
		case "azure":
			peeringList = networkingv1.NetworkingV1PeeringList{
				Data:     []networkingv1.NetworkingV1Peering{azurePeering},
				Metadata: networkingv1.ListMeta{},
			}
		case "gcp":
			peeringList = networkingv1.NetworkingV1PeeringList{
				Data:     []networkingv1.NetworkingV1Peering{gcpPeering},
				Metadata: networkingv1.ListMeta{Next: *networkingv1.NewNullableString(networkingv1.PtrString("/networking/v1/peerings?environment=env-00000&page_size=1&page_token=azure"))},
			}
		default:
			peeringList = networkingv1.NetworkingV1PeeringList{
				Data:     []networkingv1.NetworkingV1Peering{awsPeering},
				Metadata: networkingv1.ListMeta{Next: *networkingv1.NewNullableString(networkingv1.PtrString("/networking/v1/peerings?environment=env-00000&page_size=1&page_token=gcp"))},
			}
		}

		err := json.NewEncoder(w).Encode(peeringList)
		require.NoError(t, err)
	}
}

func getPeering(id, name, cloud string) networkingv1.NetworkingV1Peering {
	peering := networkingv1.NetworkingV1Peering{
		Id: networkingv1.PtrString(id),
		Spec: &networkingv1.NetworkingV1PeeringSpec{
			Cloud:       &networkingv1.NetworkingV1PeeringSpecCloudOneOf{},
			DisplayName: networkingv1.PtrString(name),
			Environment: &networkingv1.ObjectReference{Id: "env-00000"},
			Network:     &networkingv1.ObjectReference{Id: "n-abcde1"},
		},
		Status: &networkingv1.NetworkingV1PeeringStatus{
			Phase: "READY",
		},
	}

	switch cloud {
	case "AWS":
		peering.Spec.Cloud.NetworkingV1AwsPeering = &networkingv1.NetworkingV1AwsPeering{
			Kind:           "AwsPeering",
			Account:        "000000000000",
			Vpc:            "vpc-00000000000000000",
			Routes:         []string{"10.108.16.0/21"},
			CustomerRegion: "us-east-1",
		}
	case "GCP":
		peering.Spec.Cloud.NetworkingV1GcpPeering = &networkingv1.NetworkingV1GcpPeering{
			Kind:       "GcpPeering",
			Project:    "p-1",
			VpcNetwork: "v-1",
		}
	case "AZURE":
		peering.Spec.Cloud.NetworkingV1AzurePeering = &networkingv1.NetworkingV1AzurePeering{
			Kind:           "AzurePeering",
			Tenant:         "t-1",
			Vnet:           "/subscriptions/s-1/resourceGroups/group-1/providers/Microsoft.Network/virtualNetworks/vnet-1",
			CustomerRegion: "centralus",
		}
	}
	return peering
}

func handleNetworkingPeeringGet(t *testing.T, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch id {
		case "peer-invalid":
			w.WriteHeader(http.StatusNotFound)
			err := writeErrorJson(w, "The peering peer-invalid was not found.")
			require.NoError(t, err)
		case "peer-111111":
			peering := getPeering("peer-111111", "aws-peering", "AWS")
			err := json.NewEncoder(w).Encode(peering)
			require.NoError(t, err)
		case "peer-111112":
			peering := getPeering("peer-111112", "gcp-peering", "GCP")
			err := json.NewEncoder(w).Encode(peering)
			require.NoError(t, err)
		case "peer-111113":
			peering := getPeering("peer-111113", "azure-peering", "AZURE")
			err := json.NewEncoder(w).Encode(peering)
			require.NoError(t, err)
		}
	}
}

func handleNetworkingPeeringUpdate(t *testing.T, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch id {
		case "peer-invalid":
			w.WriteHeader(http.StatusNotFound)
			err := writeErrorJson(w, "The peering peer-invalid was not found.")
			require.NoError(t, err)
		case "peer-111111":
			body := &networkingv1.NetworkingV1Network{}
			err := json.NewDecoder(r.Body).Decode(body)
			require.NoError(t, err)

			peering := getPeering("peer-111111", body.Spec.GetDisplayName(), "AWS")
			err = json.NewEncoder(w).Encode(peering)
			require.NoError(t, err)
		}
	}
}

func handleNetworkingPeeringDelete(t *testing.T, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch id {
		case "peer-invalid":
			w.WriteHeader(http.StatusNotFound)
			err := writeErrorJson(w, "The network peer-invalid was not found.")
			require.NoError(t, err)
		case "peer-111111", "peer-111112":
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func handleNetworkingPeeringCreate(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := &networkingv1.NetworkingV1Peering{}
		err := json.NewDecoder(r.Body).Decode(body)
		require.NoError(t, err)

		peering := networkingv1.NetworkingV1Peering{
			Id: networkingv1.PtrString("peer-111111"),
			Spec: &networkingv1.NetworkingV1PeeringSpec{
				Cloud:       body.Spec.Cloud,
				DisplayName: body.Spec.DisplayName,
				Environment: &networkingv1.ObjectReference{Id: body.Spec.Environment.GetId()},
				Network:     &networkingv1.ObjectReference{Id: body.Spec.Network.GetId()},
			},
			Status: &networkingv1.NetworkingV1PeeringStatus{
				Phase: "PROVISIONING",
			},
		}

		err = json.NewEncoder(w).Encode(peering)
		require.NoError(t, err)
	}
}

func handleNetworkingTransitGatewayAttachmentGet(t *testing.T, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch id {
		case "tgwa-invalid":
			w.WriteHeader(http.StatusNotFound)
			return
		case "tgwa-111111":
			attachment := getTransitGatewayAttachment("tgwa-111111", "aws-tgwa1")
			err := json.NewEncoder(w).Encode(attachment)
			require.NoError(t, err)
		}
	}
}

func handleNetworkingTransitGatewayAttachmentsList(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		attachment1 := getTransitGatewayAttachment("tgwa-111111", "aws-tgwa1")
		attachment2 := getTransitGatewayAttachment("tgwa-222222", "aws-tgwa2")
		attachment3 := getTransitGatewayAttachment("tgwa-333333", "aws-tgwa3")

		pageToken := r.URL.Query().Get("page_token")
		var transitGatewayAttachmentList networkingv1.NetworkingV1TransitGatewayAttachmentList
		switch pageToken {
		case "aws-tgwa3":
			transitGatewayAttachmentList = networkingv1.NetworkingV1TransitGatewayAttachmentList{
				Data:     []networkingv1.NetworkingV1TransitGatewayAttachment{attachment3},
				Metadata: networkingv1.ListMeta{},
			}
		case "aws-tgwa2":
			transitGatewayAttachmentList = networkingv1.NetworkingV1TransitGatewayAttachmentList{
				Data:     []networkingv1.NetworkingV1TransitGatewayAttachment{attachment2},
				Metadata: networkingv1.ListMeta{Next: *networkingv1.NewNullableString(networkingv1.PtrString("/networking/v1/transit-gateway-attachments?environment=env-00000&page_size=1&page_token=aws-tgwa3"))},
			}
		default:
			transitGatewayAttachmentList = networkingv1.NetworkingV1TransitGatewayAttachmentList{
				Data:     []networkingv1.NetworkingV1TransitGatewayAttachment{attachment1},
				Metadata: networkingv1.ListMeta{Next: *networkingv1.NewNullableString(networkingv1.PtrString("/networking/v1/transit-gateway-attachments?environment=env-00000&page_size=1&page_token=aws-tgwa2"))},
			}
		}

		err := json.NewEncoder(w).Encode(transitGatewayAttachmentList)
		require.NoError(t, err)
	}
}

func getTransitGatewayAttachment(id, name string) networkingv1.NetworkingV1TransitGatewayAttachment {
	attachment := networkingv1.NetworkingV1TransitGatewayAttachment{
		Id:   networkingv1.PtrString(id),
		Kind: networkingv1.PtrString("TransitGatewayAttachment"),
		Spec: &networkingv1.NetworkingV1TransitGatewayAttachmentSpec{
			DisplayName: networkingv1.PtrString(name),
			Environment: &networkingv1.ObjectReference{Id: "env-00000"},
			Network:     &networkingv1.ObjectReference{Id: "n-abcde1"},
			Cloud: &networkingv1.NetworkingV1TransitGatewayAttachmentSpecCloudOneOf{
				NetworkingV1AwsTransitGatewayAttachment: &networkingv1.NetworkingV1AwsTransitGatewayAttachment{
					Kind:             "AwsTransitGatewayAttachment",
					RamShareArn:      "arn:aws:ram:us-east-1:000000000000:resource-share/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx",
					Routes:           []string{"10.0.0.0/16"},
					TransitGatewayId: "tgw-00000000000000000",
				},
			},
		},
		Status: &networkingv1.NetworkingV1TransitGatewayAttachmentStatus{
			Phase: "READY",
			Cloud: &networkingv1.NetworkingV1TransitGatewayAttachmentStatusCloudOneOf{
				NetworkingV1AwsTransitGatewayAttachmentStatus: &networkingv1.NetworkingV1AwsTransitGatewayAttachmentStatus{
					Kind:                       networkingv1.PtrString("AwsTransitGatewayAttachmentStatus"),
					TransitGatewayAttachmentId: "tgw-attach-00000000000000000",
				},
			},
		},
	}
	return attachment
}

func handleNetworkingTransitGatewayAttachmentUpdate(t *testing.T, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch id {
		case "tgwa-invalid":
			w.WriteHeader(http.StatusNotFound)
			return
		case "tgwa-111111":
			body := &networkingv1.NetworkingV1Network{}
			err := json.NewDecoder(r.Body).Decode(body)
			require.NoError(t, err)

			attachment := getTransitGatewayAttachment("tgwa-111111", body.Spec.GetDisplayName())
			err = json.NewEncoder(w).Encode(attachment)
			require.NoError(t, err)
		}
	}
}

func handleNetworkingTransitGatewayAttachmentDelete(_ *testing.T, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch id {
		case "tgw-invalid":
			w.WriteHeader(http.StatusNotFound)
			return
		case "tgwa-111111", "tgwa-222222":
			w.WriteHeader(http.StatusNoContent)
		}
	}
}
