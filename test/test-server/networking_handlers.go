package testserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	networkingipv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-ip/v1"
	networkingprivatelinkv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-privatelink/v1"
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
			handleNetworkingTransitGatewayAttachmentList(t)(w, r)
		case http.MethodPost:
			handleNetworkingTransitGatewayAttachmentCreate(t)(w, r)
		}
	}
}

// Handler for "/networking/v1/private-link-accesses/{id}"
func handleNetworkingPrivateLinkAccess(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		switch r.Method {
		case http.MethodGet:
			handleNetworkingPrivateLinkAccessGet(t, id)(w, r)
		case http.MethodPatch:
			handleNetworkingPrivateLinkAccessUpdate(t, id)(w, r)
		case http.MethodDelete:
			handleNetworkingPrivateLinkAccessDelete(t, id)(w, r)
		}
	}
}

// Handler for "/networking/v1/private-link-accesses"
func handleNetworkingPrivateLinkAccesses(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleNetworkingPrivateLinkAccessList(t)(w, r)
		case http.MethodPost:
			handleNetworkingPrivateLinkAccessCreate(t)(w, r)
		}
	}
}

// Handler for "/networking/v1/private-link-attachments/{id}"
func handleNetworkingPrivateLinkAttachment(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		switch r.Method {
		case http.MethodGet:
			handleNetworkingPrivateLinkAttachmentGet(t, id)(w, r)
		case http.MethodPatch:
			handleNetworkingPrivateLinkAttachmentUpdate(t, id)(w, r)
		case http.MethodDelete:
			handleNetworkingPrivateLinkAttachmentDelete(t, id)(w, r)
		}
	}
}

// Handler for "/networking/v1/private-link-attachments"
func handleNetworkingPrivateLinkAttachments(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleNetworkingPrivateLinkAttachmentList(t)(w, r)
		case http.MethodPost:
			handleNetworkingPrivateLinkAttachmentCreate(t)(w, r)
		}
	}
}

// Handler for "/networking/v1/private-link-attachment-connections/{id}"
func handleNetworkingPrivateLinkAttachmentConnection(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		switch r.Method {
		case http.MethodGet:
			handleNetworkingPrivateLinkAttachmentConnectionGet(t, id)(w, r)
		case http.MethodPatch:
			handleNetworkingPrivateLinkAttachmentConnectionUpdate(t, id)(w, r)
		case http.MethodDelete:
			handleNetworkingPrivateLinkAttachmentConnectionDelete(t, id)(w, r)
		}
	}
}

// Handler for "/networking/v1/private-link-attachment-connections"
func handleNetworkingPrivateLinkAttachmentConnections(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleNetworkingPrivateLinkAttachmentConnectionList(t)(w, r)
		case http.MethodPost:
			handleNetworkingPrivateLinkAttachmentConnectionCreate(t)(w, r)
		}
	}
}

// Handler for "/networking/v1/ip-addresses"
func handleNetworkingIpAddresses(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleNetworkingIpAddressList(t)(w, r)
		}
	}
}

// Handler for "/networking/v1/network-link-service/{id}"
func handleNetworkingNetworkLinkService(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		switch r.Method {
		case http.MethodGet:
			handleNetworkingNetworkLinkServiceGet(t, id)(w, r)
		case http.MethodDelete:
			handleNetworkingNetworkLinkServiceDelete(t, id)(w, r)
		case http.MethodPatch:
			handleNetworkingNetworkLinkServiceUpdate(t, id)(w, r)
		}
	}
}

// Handler for "/networking/v1/network-link-services"
func handleNetworkingNetworkLinkServices(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleNetworkingNetworkLinkServiceList(t)(w, r)
		case http.MethodPost:
			handleNetworkingNetworkLinkServiceCreate(t)(w, r)
		}
	}
}

// Handler for "/networking/v1/network-link-endpoint/{id}"
func handleNetworkingNetworkLinkEndpoint(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		switch r.Method {
		case http.MethodGet:
			handleNetworkingNetworkLinkEndpointGet(t, id)(w, r)
		case http.MethodDelete:
			handleNetworkingNetworkLinkEndpointDelete(t, id)(w, r)
		case http.MethodPatch:
			handleNetworkingNetworkLinkEndpointUpdate(t, id)(w, r)
		}
	}
}

// Handler for "/networking/v1/network-link-endpoints"
func handleNetworkingNetworkLinkEndpoints(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleNetworkingNetworkLinkEndpointList(t)(w, r)
		case http.MethodPost:
			handleNetworkingNetworkLinkEndpointCreate(t)(w, r)
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
			network := getAwsNetwork("n-abcde1", "prod-aws-us-east1", "READY", []string{"TRANSITGATEWAY", "PEERING"})
			err := json.NewEncoder(w).Encode(network)
			require.NoError(t, err)
		case "n-abcde2":
			network := getGcpNetwork("n-abcde2", "prod-gcp-us-central1", "READY", []string{"PEERING"})
			err := json.NewEncoder(w).Encode(network)
			require.NoError(t, err)
		case "n-abcde3":
			network := getAzureNetwork("n-abcde3", "prod-azure-eastus2", "READY", []string{"PEERING"})
			err := json.NewEncoder(w).Encode(network)
			require.NoError(t, err)
		case "n-abcde4":
			network := getAwsNetwork("n-abcde4", "prod-aws-us-east1", "PROVISIONING", []string{"TRANSITGATEWAY", "PEERING"})
			err := json.NewEncoder(w).Encode(network)
			require.NoError(t, err)
		case "n-abcde5":
			network := getGcpNetwork("n-abcde5", "prod-gcp-us-central1", "PROVISIONING", []string{"PEERING"})
			err := json.NewEncoder(w).Encode(network)
			require.NoError(t, err)
		case "n-abcde6":
			network := getAzureNetwork("n-abcde6", "prod-azure-eastus2", "PROVISIONING", []string{"PEERING"})
			err := json.NewEncoder(w).Encode(network)
			require.NoError(t, err)
		case "n-abcde7":
			network := getAwsNetwork("n-abcde7", "prod-aws-us-east1", "READY", []string{"PRIVATELINK"})
			err := json.NewEncoder(w).Encode(network)
			require.NoError(t, err)
		case "n-abcde8":
			network := getGcpNetwork("n-abcde8", "prod-gcp-us-central1", "READY", []string{"PRIVATELINK"})
			err := json.NewEncoder(w).Encode(network)
			require.NoError(t, err)
		case "n-abcde9":
			network := getAzureNetwork("n-abcde9", "prod-azure-eastus2", "READY", []string{"PRIVATELINK"})
			err := json.NewEncoder(w).Encode(network)
			require.NoError(t, err)
		case "n-abcde10":
			network := getAwsNetwork("n-abcde10", "prod-aws-us-east1", "PROVISIONING", []string{"PRIVATELINK"})
			err := json.NewEncoder(w).Encode(network)
			require.NoError(t, err)
		case "n-abcde11":
			network := getGcpNetwork("n-abcde11", "stag-gcp-us-central1", "PROVISIONING", []string{"PRIVATELINK"})
			err := json.NewEncoder(w).Encode(network)
			require.NoError(t, err)
		case "n-abcde12":
			network := getAzureNetwork("n-abcde12", "dev-azure-eastus2", "PROVISIONING", []string{"PRIVATELINK"})
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
			network := getAwsNetwork("n-abcde1", "new-prod-aws-us-east1", "READY", []string{"TRANSITGATEWAY", "PEERING"})
			err := json.NewEncoder(w).Encode(network)
			require.NoError(t, err)
		}
	}
}

func handleNetworkingNetworkList(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		name := q["spec.display_name"]
		cloud := q["spec.cloud"]
		region := q["spec.region"]
		cidr := q["spec.cidr"]
		phase := q["status.phase"]
		connection := q["spec.connection_types"]

		networkList := getNetworkList(name, cloud, region, cidr, phase, connection)
		err := json.NewEncoder(w).Encode(networkList)
		require.NoError(t, err)
	}
}

func getNetworkList(filterName, filterCloud, filterRegion, filterCidr, filterPhase, filterConnection []string) networkingv1.NetworkingV1NetworkList {
	gcpNetwork := getGcpNetwork("n-abcde1", "prod-gcp-us-central1", "READY", []string{"PEERING"})
	azureNetwork := getAzureNetwork("n-abcde2", "prod-azure-eastus2", "READY", []string{"PRIVATELINK"})

	// Same cloud, sort by region
	awsNetwork := getAwsNetwork("n-abcde3", "prod-aws-us-east1", "READY", []string{"TRANSITGATEWAY", "PEERING"})
	awsNetwork2 := getAwsNetwork("n-abcde4", "prod-aws-us-east1", "READY", []string{"TRANSITGATEWAY", "PEERING"})
	awsNetwork2.Spec.SetRegion("us-west-2")

	// Same cloud, region, sort by created_at
	awsNetwork3 := getAwsNetwork("n-abcde5", "", "READY", []string{"TRANSITGATEWAY", "PEERING"})
	awsNetwork.Metadata = &networkingv1.ObjectMeta{CreatedAt: networkingv1.PtrTime(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))}
	awsNetwork3.Metadata = &networkingv1.ObjectMeta{CreatedAt: networkingv1.PtrTime(time.Date(2023, 1, 1, 0, 0, 0, 1, time.UTC))}

	networkList := networkingv1.NetworkingV1NetworkList{
		Data: []networkingv1.NetworkingV1Network{
			gcpNetwork, azureNetwork, awsNetwork, awsNetwork2, awsNetwork3,
		},
	}
	networkList.Data = filterNetworkList(networkList.Data, filterName, filterCloud, filterRegion, filterCidr, filterPhase, filterConnection)

	return networkList
}

func filterNetworkList(networkList []networkingv1.NetworkingV1Network, name, cloud, region, cidr, phase, connection []string) []networkingv1.NetworkingV1Network {
	var filteredNetworkList []networkingv1.NetworkingV1Network
	for _, networkSpec := range networkList {
		if (slices.Contains(name, networkSpec.Spec.GetDisplayName()) || name == nil) &&
			(slices.Contains(cloud, networkSpec.Spec.GetCloud()) || cloud == nil) &&
			(slices.Contains(region, networkSpec.Spec.GetRegion()) || region == nil) &&
			(slices.Contains(cidr, networkSpec.Spec.GetCidr()) || cidr == nil) &&
			(slices.Contains(phase, networkSpec.Status.GetPhase()) || phase == nil) &&
			(containsFilter(connection, networkSpec.Status.GetActiveConnectionTypes().Items) || connection == nil) {
			filteredNetworkList = append(filteredNetworkList, networkSpec)
		}
	}
	return filteredNetworkList
}

func containsFilter(filter, resource []string) bool {
	overlap := make(map[string]bool)

	for _, val := range filter {
		overlap[strings.ToUpper(val)] = true
	}

	for _, val := range resource {
		if overlap[val] {
			return true
		}
	}

	return false
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

			if slices.Contains(connectionTypes, "PRIVATELINK") {
				network.Status.DnsDomain = networkingv1.PtrString("")
				network.Status.ZonalSubdomains = &map[string]string{}
			}

			switch *body.Spec.Cloud {
			case "AWS":
				network.Status.Cloud = &networkingv1.NetworkingV1NetworkStatusCloudOneOf{
					NetworkingV1AwsNetwork: &networkingv1.NetworkingV1AwsNetwork{
						Kind: "AwsNetwork",
					},
				}
				if slices.Contains(connectionTypes, "PRIVATELINK") {
					network.Status.Cloud.NetworkingV1AwsNetwork.PrivateLinkEndpointService = networkingv1.PtrString("")
				}
			}

			err = json.NewEncoder(w).Encode(network)
			require.NoError(t, err)
		}
	}
}

func getAwsNetwork(id, name, phase string, connectionTypes []string) networkingv1.NetworkingV1Network {
	network := networkingv1.NetworkingV1Network{
		Id: networkingv1.PtrString(id),
		Spec: &networkingv1.NetworkingV1NetworkSpec{
			Environment: &networkingv1.ObjectReference{Id: "env-00000"},
			DisplayName: networkingv1.PtrString(name),
			Cloud:       networkingv1.PtrString("AWS"),
			Region:      networkingv1.PtrString("us-east-1"),
			Cidr:        networkingv1.PtrString("10.200.0.0/16"),
			Zones:       &[]string{"use1-az1", "use1-az2", "use1-az3"},
		},
		Status: &networkingv1.NetworkingV1NetworkStatus{
			Phase:                    phase,
			SupportedConnectionTypes: networkingv1.NetworkingV1SupportedConnectionTypes{Items: connectionTypes},
			ActiveConnectionTypes:    networkingv1.NetworkingV1ConnectionTypes{Items: []string{}},
			Cloud: &networkingv1.NetworkingV1NetworkStatusCloudOneOf{
				NetworkingV1AwsNetwork: &networkingv1.NetworkingV1AwsNetwork{
					Kind: "AwsNetwork",
				},
			},
		},
	}

	if slices.Contains(connectionTypes, "PRIVATELINK") {
		network.Spec.DnsConfig = &networkingv1.NetworkingV1DnsConfig{Resolution: "PRIVATE"}
		network.Status.DnsDomain = networkingv1.PtrString("")
		network.Status.ZonalSubdomains = &map[string]string{}
		network.Status.Cloud.NetworkingV1AwsNetwork.PrivateLinkEndpointService = networkingv1.PtrString("")
		if phase == "READY" {
			network.Status.DnsDomain = networkingv1.PtrString("abcdef1gh2i.us-east-1.aws.devel.cpdev.cloud")
			network.Status.ZonalSubdomains = &map[string]string{
				"use1-az1": "use1-az1.abcdef1gh2i.us-east-1.aws.devel.cpdev.cloud",
				"use1-az2": "use1-az2.abcdef1gh2i.us-east-1.aws.devel.cpdev.cloud",
				"use1-az3": "use1-az3.abcdef1gh2i.us-east-1.aws.devel.cpdev.cloud",
			}
			network.Status.Cloud.NetworkingV1AwsNetwork.PrivateLinkEndpointService = networkingv1.PtrString("com.amazonaws.vpce.us-east-2.vpce-svc-01234567890abcdef")
		}
	}

	if phase == "READY" {
		network.Status.ActiveConnectionTypes = networkingv1.NetworkingV1ConnectionTypes{Items: connectionTypes}
		network.Status.Cloud.NetworkingV1AwsNetwork.Vpc = "vpc-00000000000000000"
		network.Status.Cloud.NetworkingV1AwsNetwork.Account = "000000000000"
	}

	return network
}

func getGcpNetwork(id, name, phase string, connectionTypes []string) networkingv1.NetworkingV1Network {
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
			SupportedConnectionTypes: networkingv1.NetworkingV1SupportedConnectionTypes{Items: connectionTypes},
			ActiveConnectionTypes:    networkingv1.NetworkingV1ConnectionTypes{Items: []string{}},
			Cloud: &networkingv1.NetworkingV1NetworkStatusCloudOneOf{
				NetworkingV1GcpNetwork: &networkingv1.NetworkingV1GcpNetwork{
					Kind: "GcpNetwork",
				},
			},
		},
	}

	if slices.Contains(connectionTypes, "PRIVATELINK") {
		network.Spec.DnsConfig = &networkingv1.NetworkingV1DnsConfig{Resolution: "PRIVATE"}
		network.Status.DnsDomain = networkingv1.PtrString("")
		network.Status.ZonalSubdomains = &map[string]string{}
		network.Status.Cloud.NetworkingV1GcpNetwork.PrivateServiceConnectServiceAttachments = &map[string]string{}
		if phase == "READY" {
			network.Status.DnsDomain = networkingv1.PtrString("0123456789abcdef.us-central1.gcp.devel.cpdev.cloud")
			network.Status.ZonalSubdomains = &map[string]string{
				"us-central1-a": "us-central1-a.0123456789abcdef.us-central1.gcp.devel.cpdev.cloud",
				"us-central1-b": "us-central1-b.0123456789abcdef.us-central1.gcp.devel.cpdev.cloud",
				"us-central1-c": "us-central1-c.0123456789abcdef.us-central1.gcp.devel.cpdev.cloud",
			}
			network.Status.Cloud.NetworkingV1GcpNetwork.PrivateServiceConnectServiceAttachments = &map[string]string{
				"us-central1-a": "projects/gcp-project/regions/us-central1/serviceAttachments/gcp-vpc-service-attachment-us-central1-a",
				"us-central1-b": "projects/gcp-project/regions/us-central1/serviceAttachments/gcp-vpc-service-attachment-us-central1-b",
				"us-central1-c": "projects/gcp-project/regions/us-central1/serviceAttachments/gcp-vpc-service-attachment-us-central1-c",
			}
		}
	}

	if phase == "READY" {
		network.Status.ActiveConnectionTypes = networkingv1.NetworkingV1ConnectionTypes{Items: connectionTypes}
		network.Status.Cloud.NetworkingV1GcpNetwork.Project = "gcp-project"
		network.Status.Cloud.NetworkingV1GcpNetwork.VpcNetwork = "gcp-vpc"
	}

	return network
}

func getAzureNetwork(id, name, phase string, connectionTypes []string) networkingv1.NetworkingV1Network {
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
			SupportedConnectionTypes: networkingv1.NetworkingV1SupportedConnectionTypes{Items: connectionTypes},
			ActiveConnectionTypes:    networkingv1.NetworkingV1ConnectionTypes{Items: []string{}},
			Cloud: &networkingv1.NetworkingV1NetworkStatusCloudOneOf{
				NetworkingV1AzureNetwork: &networkingv1.NetworkingV1AzureNetwork{
					Kind: "AzureNetwork",
				},
			},
		},
	}

	if slices.Contains(connectionTypes, "PRIVATELINK") {
		network.Spec.DnsConfig = &networkingv1.NetworkingV1DnsConfig{Resolution: "PRIVATE"}
		network.Status.DnsDomain = networkingv1.PtrString("")
		network.Status.ZonalSubdomains = &map[string]string{}
		network.Status.Cloud.NetworkingV1AzureNetwork.PrivateLinkServiceAliases = &map[string]string{}
		network.Status.Cloud.NetworkingV1AzureNetwork.PrivateLinkServiceResourceIds = &map[string]string{}
		if phase == "READY" {
			network.Status.DnsDomain = networkingv1.PtrString("0123456789a.eastus2.azure.devel.cpdev.cloud")
			network.Status.ZonalSubdomains = &map[string]string{
				"1": "az1.0123456789a.eastus2.azure.devel.cpdev.cloud",
				"2": "az2.0123456789a.eastus2.azure.devel.cpdev.cloud",
				"3": "az3.0123456789a.eastus2.azure.devel.cpdev.cloud",
			}
			network.Status.Cloud.NetworkingV1AzureNetwork.PrivateLinkServiceAliases = &map[string]string{
				"1": "azure-vnet-privatelink-1.a0a0aa00-a000-0aa0-a00a-0aaa0000a00a.eastus2.azure.privatelinkservice",
				"2": "azure-vnet-privatelink-2.b0b0bb00-b000-0bb0-b00b-0bbb0000b00b.eastus2.azure.privatelinkservice",
				"3": "azure-vnet-privatelink-3.c0c0cc00-c000-0cc0-c00c-0ccc0000c00c.eastus2.azure.privatelinkservice",
			}
			network.Status.Cloud.NetworkingV1AzureNetwork.PrivateLinkServiceResourceIds = &map[string]string{
				"1": "/subscriptions/aa000000-a000-0a00-00aa-0000aaa0a0a0/resourceGroups/azure-vnet/providers/Microsoft.Network/privateLinkServices/azure-vnet-privatelink-1",
				"2": "/subscriptions/aa000000-a000-0a00-00aa-0000aaa0a0a0/resourceGroups/azure-vnet/providers/Microsoft.Network/privateLinkServices/azure-vnet-privatelink-2",
				"3": "/subscriptions/aa000000-a000-0a00-00aa-0000aaa0a0a0/resourceGroups/azure-vnet/providers/Microsoft.Network/privateLinkServices/azure-vnet-privatelink-3",
			}
		}
	}

	if phase == "READY" {
		network.Status.ActiveConnectionTypes = networkingv1.NetworkingV1ConnectionTypes{Items: connectionTypes}
		network.Status.Cloud.NetworkingV1AzureNetwork.Vnet = "azure-vnet"
		network.Status.Cloud.NetworkingV1AzureNetwork.Subscription = "aa000000-a000-0a00-00aa-0000aaa0a0a0"
	}

	return network
}

func handleNetworkingPeeringList(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		name := q["spec.display_name"]
		network := q["spec.network"]
		phase := q["status.phase"]

		peeringList := getPeeringList(name, network, phase)
		err := json.NewEncoder(w).Encode(peeringList)
		require.NoError(t, err)
	}
}

func getPeeringList(filterName, filterNetwork, filterPhase []string) networkingv1.NetworkingV1PeeringList {
	peeringList := networkingv1.NetworkingV1PeeringList{
		Data: []networkingv1.NetworkingV1Peering{
			getPeering("peer-111111", "aws-peering", "AWS"),
			getPeering("peer-111112", "gcp-peering", "GCP"),
			getPeering("peer-111113", "azure-peering", "AZURE"),
		},
	}
	peeringList.Data = filterPeeringList(peeringList.Data, filterName, filterNetwork, filterPhase)

	return peeringList
}

func filterPeeringList(peeringList []networkingv1.NetworkingV1Peering, name, network, phase []string) []networkingv1.NetworkingV1Peering {
	var filteredPeeringList []networkingv1.NetworkingV1Peering
	for _, peeringSpec := range peeringList {
		if (slices.Contains(name, peeringSpec.Spec.GetDisplayName()) || name == nil) &&
			(slices.Contains(network, peeringSpec.Spec.Network.GetId()) || network == nil) &&
			(slices.Contains(phase, peeringSpec.Status.GetPhase()) || phase == nil) {
			filteredPeeringList = append(filteredPeeringList, peeringSpec)
		}
	}
	return filteredPeeringList
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
			body := &networkingv1.NetworkingV1Peering{}
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

func handleNetworkingTransitGatewayAttachmentList(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		name := q["spec.display_name"]
		network := q["spec.network"]
		phase := q["status.phase"]

		tgwaList := getTransitGatewayAttachmentList(name, network, phase)
		err := json.NewEncoder(w).Encode(tgwaList)
		require.NoError(t, err)
	}
}

func getTransitGatewayAttachmentList(filterName, filterNetwork, filterPhase []string) networkingv1.NetworkingV1TransitGatewayAttachmentList {
	tgwaList := networkingv1.NetworkingV1TransitGatewayAttachmentList{
		Data: []networkingv1.NetworkingV1TransitGatewayAttachment{
			getTransitGatewayAttachment("tgwa-111111", "aws-tgwa1"),
			getTransitGatewayAttachment("tgwa-222222", "aws-tgwa2"),
			getTransitGatewayAttachment("tgwa-333333", "aws-tgwa3"),
		},
	}
	tgwaList.Data = filterTransitGatwayAttachmentList(tgwaList.Data, filterName, filterNetwork, filterPhase)

	return tgwaList
}

func filterTransitGatwayAttachmentList(tgwaList []networkingv1.NetworkingV1TransitGatewayAttachment, name, network, phase []string) []networkingv1.NetworkingV1TransitGatewayAttachment {
	var filteredAttachmentList []networkingv1.NetworkingV1TransitGatewayAttachment
	for _, attachmentSpec := range tgwaList {
		if (slices.Contains(name, attachmentSpec.Spec.GetDisplayName()) || name == nil) &&
			(slices.Contains(network, attachmentSpec.Spec.Network.GetId()) || network == nil) &&
			(slices.Contains(phase, attachmentSpec.Status.GetPhase()) || phase == nil) {
			filteredAttachmentList = append(filteredAttachmentList, attachmentSpec)
		}
	}
	return filteredAttachmentList
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
			body := &networkingv1.NetworkingV1TransitGatewayAttachment{}
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

func handleNetworkingTransitGatewayAttachmentCreate(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := &networkingv1.NetworkingV1TransitGatewayAttachment{}
		err := json.NewDecoder(r.Body).Decode(body)
		require.NoError(t, err)

		networkId := body.Spec.Network.GetId()

		switch networkId {
		case "n-duplicate":
			w.WriteHeader(http.StatusConflict)
			err := writeErrorJson(w, "tgwa-123456 already exists between the network and the transit gateway")
			require.NoError(t, err)
		case "n-azure":
			w.WriteHeader(http.StatusBadRequest)
			err := writeErrorJson(w, "The provided network n-azure resides in AZURE which differs from AWS, the cloud specified for this resource.")
			require.NoError(t, err)
		default:
			attachment := networkingv1.NetworkingV1TransitGatewayAttachment{
				Id:   networkingv1.PtrString("tgwa-123456"),
				Kind: networkingv1.PtrString("TransitGatewayAttachment"),
				Spec: &networkingv1.NetworkingV1TransitGatewayAttachmentSpec{
					DisplayName: networkingv1.PtrString(body.Spec.GetDisplayName()),
					Environment: &networkingv1.ObjectReference{Id: body.Spec.Environment.GetId()},
					Network:     &networkingv1.ObjectReference{Id: body.Spec.Network.GetId()},
					Cloud: &networkingv1.NetworkingV1TransitGatewayAttachmentSpecCloudOneOf{
						NetworkingV1AwsTransitGatewayAttachment: &networkingv1.NetworkingV1AwsTransitGatewayAttachment{
							Kind:             "AwsTransitGatewayAttachment",
							RamShareArn:      body.Spec.Cloud.NetworkingV1AwsTransitGatewayAttachment.GetRamShareArn(),
							Routes:           body.Spec.Cloud.NetworkingV1AwsTransitGatewayAttachment.GetRoutes(),
							TransitGatewayId: body.Spec.Cloud.NetworkingV1AwsTransitGatewayAttachment.GetTransitGatewayId(),
						},
					},
				},
				Status: &networkingv1.NetworkingV1TransitGatewayAttachmentStatus{
					Phase: "PROVISIONING",
					Cloud: &networkingv1.NetworkingV1TransitGatewayAttachmentStatusCloudOneOf{
						NetworkingV1AwsTransitGatewayAttachmentStatus: &networkingv1.NetworkingV1AwsTransitGatewayAttachmentStatus{
							Kind:                       networkingv1.PtrString("AwsTransitGatewayAttachmentStatus"),
							TransitGatewayAttachmentId: "tgw-attach-00000000000000000",
						},
					},
				},
			}
			err = json.NewEncoder(w).Encode(attachment)
			require.NoError(t, err)
		}
	}
}

func handleNetworkingPrivateLinkAccessGet(t *testing.T, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch id {
		case "pla-invalid":
			w.WriteHeader(http.StatusNotFound)
			err := writeErrorJson(w, "The private-link-access pla-invalid was not found.")
			require.NoError(t, err)
		case "pla-111111":
			access := getPrivateLinkAccess("pla-111111", "aws-pla", "AWS")
			err := json.NewEncoder(w).Encode(access)
			require.NoError(t, err)
		case "pla-111112":
			access := getPrivateLinkAccess("pla-111112", "gcp-pla", "GCP")
			err := json.NewEncoder(w).Encode(access)
			require.NoError(t, err)
		case "pla-111113":
			access := getPrivateLinkAccess("pla-111113", "azure-pla", "AZURE")
			err := json.NewEncoder(w).Encode(access)
			require.NoError(t, err)
		}
	}
}

func getPrivateLinkAccess(id, name, cloud string) networkingv1.NetworkingV1PrivateLinkAccess {
	access := networkingv1.NetworkingV1PrivateLinkAccess{
		Id: networkingv1.PtrString(id),
		Spec: &networkingv1.NetworkingV1PrivateLinkAccessSpec{
			Cloud:       &networkingv1.NetworkingV1PrivateLinkAccessSpecCloudOneOf{},
			DisplayName: networkingv1.PtrString(name),
			Environment: &networkingv1.ObjectReference{Id: "env-00000"},
			Network:     &networkingv1.ObjectReference{Id: "n-abcde1"},
		},
		Status: &networkingv1.NetworkingV1PrivateLinkAccessStatus{
			Phase: "READY",
		},
	}

	switch cloud {
	case "AWS":
		access.Spec.Cloud.NetworkingV1AwsPrivateLinkAccess = &networkingv1.NetworkingV1AwsPrivateLinkAccess{
			Kind:    "AwsPrivateLinkAccess",
			Account: "012345678901",
		}
	case "GCP":
		access.Spec.Cloud.NetworkingV1GcpPrivateServiceConnectAccess = &networkingv1.NetworkingV1GcpPrivateServiceConnectAccess{
			Kind:    "GcpPrivateServiceConnectAccess",
			Project: "temp-gear-123456",
		}
	case "AZURE":
		access.Spec.Cloud.NetworkingV1AzurePrivateLinkAccess = &networkingv1.NetworkingV1AzurePrivateLinkAccess{
			Kind:         "AzurePrivateLinkAccess",
			Subscription: "1234abcd-12ab-34cd-1234-123456abcdef",
		}
	}

	return access
}

func handleNetworkingPrivateLinkAccessList(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		name := q["spec.display_name"]
		network := q["spec.network"]
		phase := q["status.phase"]

		plaList := getPrivateLinkAccessList(name, network, phase)
		err := json.NewEncoder(w).Encode(plaList)
		require.NoError(t, err)
	}
}

func getPrivateLinkAccessList(filterName, filterNetwork, filterPhase []string) networkingv1.NetworkingV1PrivateLinkAccessList {
	plaList := networkingv1.NetworkingV1PrivateLinkAccessList{
		Data: []networkingv1.NetworkingV1PrivateLinkAccess{
			getPrivateLinkAccess("pla-111111", "aws-pla", "AWS"),
			getPrivateLinkAccess("pla-111112", "gcp-pla", "GCP"),
			getPrivateLinkAccess("pla-111113", "azure-pla", "AZURE"),
		},
	}
	plaList.Data = filterAccessList(plaList.Data, filterName, filterNetwork, filterPhase)

	return plaList
}

func filterAccessList(accessList []networkingv1.NetworkingV1PrivateLinkAccess, name, network, phase []string) []networkingv1.NetworkingV1PrivateLinkAccess {
	var filteredAccessList []networkingv1.NetworkingV1PrivateLinkAccess
	for _, accessSpec := range accessList {
		if (slices.Contains(name, accessSpec.Spec.GetDisplayName()) || name == nil) &&
			(slices.Contains(network, accessSpec.Spec.Network.GetId()) || network == nil) &&
			(slices.Contains(phase, accessSpec.Status.GetPhase()) || phase == nil) {
			filteredAccessList = append(filteredAccessList, accessSpec)
		}
	}
	return filteredAccessList
}

func handleNetworkingPrivateLinkAccessUpdate(t *testing.T, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch id {
		case "pla-invalid":
			w.WriteHeader(http.StatusNotFound)
			err := writeErrorJson(w, "The private-link-access pla-invalid was not found.")
			require.NoError(t, err)
		case "pla-111111":
			body := &networkingv1.NetworkingV1PrivateLinkAccess{}
			err := json.NewDecoder(r.Body).Decode(body)
			require.NoError(t, err)

			attachment := getPrivateLinkAccess("pla-111111", body.Spec.GetDisplayName(), "AWS")
			err = json.NewEncoder(w).Encode(attachment)
			require.NoError(t, err)
		}
	}
}

func handleNetworkingPrivateLinkAccessDelete(t *testing.T, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch id {
		case "pla-invalid":
			w.WriteHeader(http.StatusNotFound)
			err := writeErrorJson(w, "The private-link-access pla-invalid was not found.")
			require.NoError(t, err)
		case "pla-111111", "pla-222222":
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func handleNetworkingPrivateLinkAccessCreate(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := &networkingv1.NetworkingV1PrivateLinkAccess{}
		err := json.NewDecoder(r.Body).Decode(body)
		require.NoError(t, err)

		networkId := body.Spec.Network.GetId()

		switch networkId {
		case "n-duplicate":
			w.WriteHeader(http.StatusConflict)
			err := writeErrorJson(w, "Private link already exists.")
			require.NoError(t, err)
		case "n-azure":
			w.WriteHeader(http.StatusBadRequest)
			err := writeErrorJson(w, "The provided network n-azure resides in AZURE which differs from AWS, the cloud specified for this resource.")
			require.NoError(t, err)
		default:
			access := networkingv1.NetworkingV1PrivateLinkAccess{
				Id:   networkingv1.PtrString("pla-123456"),
				Kind: networkingv1.PtrString("PrivateLinkAccess"),
				Spec: &networkingv1.NetworkingV1PrivateLinkAccessSpec{
					Cloud:       body.Spec.Cloud,
					DisplayName: networkingv1.PtrString(body.Spec.GetDisplayName()),
					Environment: &networkingv1.ObjectReference{Id: body.Spec.Environment.GetId()},
					Network:     &networkingv1.ObjectReference{Id: body.Spec.Network.GetId()},
				},
				Status: &networkingv1.NetworkingV1PrivateLinkAccessStatus{
					Phase: "PROVISIONING",
				},
			}
			err = json.NewEncoder(w).Encode(access)
			require.NoError(t, err)
		}
	}
}

func handleNetworkingPrivateLinkAttachmentGet(t *testing.T, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch id {
		case "platt-invalid":
			w.WriteHeader(http.StatusNotFound)
			return
		case "platt-111111":
			attachment := getPrivateLinkAttachment("platt-111111", "aws-platt", "WAITING_FOR_CONNECTIONS")
			err := json.NewEncoder(w).Encode(attachment)
			require.NoError(t, err)
		case "platt-111112":
			attachment := getPrivateLinkAttachment("platt-111112", "aws-platt", "PROVISIONING")
			err := json.NewEncoder(w).Encode(attachment)
			require.NoError(t, err)
		}
	}
}

func getPrivateLinkAttachment(id, name, phase string) networkingprivatelinkv1.NetworkingV1PrivateLinkAttachment {
	attachment := networkingprivatelinkv1.NetworkingV1PrivateLinkAttachment{
		Id: networkingv1.PtrString(id),
		Spec: &networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentSpec{
			Cloud:       networkingprivatelinkv1.PtrString("AWS"),
			Region:      networkingprivatelinkv1.PtrString("us-west-2"),
			DisplayName: networkingprivatelinkv1.PtrString(name),
			Environment: &networkingprivatelinkv1.ObjectReference{Id: "env-00000"},
		},
		Status: &networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentStatus{
			Phase: phase,
		},
	}

	if phase != "PROVISIONING" {
		attachment.Status.Cloud = &networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentStatusCloudOneOf{
			NetworkingV1AwsPrivateLinkAttachmentStatus: &networkingprivatelinkv1.NetworkingV1AwsPrivateLinkAttachmentStatus{
				Kind: "AwsPrivateLinkAttachmentStatus",
				VpcEndpointService: networkingprivatelinkv1.NetworkingV1AwsVpcEndpointService{
					VpcEndpointServiceName: "com.amazonaws.vpce.us-west-2.vpce-svc-01234567890abcdef",
				},
			},
		}
	}

	return attachment
}

func handleNetworkingPrivateLinkAttachmentList(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		name := q["spec.display_name"]
		cloud := q["spec.cloud"]
		region := q["spec.region"]
		phase := q["status.phase"]

		plattList := getPrivateLinkAttachmentList(name, cloud, region, phase)
		err := json.NewEncoder(w).Encode(plattList)
		require.NoError(t, err)
	}
}

func getPrivateLinkAttachmentList(filterName, filterCloud, filterRegion, filterPhase []string) networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentList {
	plattList := networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentList{
		Data: []networkingprivatelinkv1.NetworkingV1PrivateLinkAttachment{
			getPrivateLinkAttachment("platt-111111", "aws-platt-1", "PROVISIONING"),
			getPrivateLinkAttachment("platt-111112", "aws-platt-2", "WAITING_FOR_CONNECTIONS"),
			getPrivateLinkAttachment("platt-111113", "aws-platt-3", "WAITING_FOR_CONNECTIONS"),
		},
	}
	plattList.Data = filterAttachmentList(plattList.Data, filterName, filterCloud, filterRegion, filterPhase)

	return plattList
}

func filterAttachmentList(attachmentList []networkingprivatelinkv1.NetworkingV1PrivateLinkAttachment, name, cloud, region, phase []string) []networkingprivatelinkv1.NetworkingV1PrivateLinkAttachment {
	var filteredAttachmentList []networkingprivatelinkv1.NetworkingV1PrivateLinkAttachment
	for _, attachmentSpec := range attachmentList {
		if (slices.Contains(name, attachmentSpec.Spec.GetDisplayName()) || name == nil) &&
			(slices.Contains(cloud, attachmentSpec.Spec.GetCloud()) || cloud == nil) &&
			(slices.Contains(region, attachmentSpec.Spec.GetRegion()) || region == nil) &&
			(slices.Contains(phase, attachmentSpec.Status.GetPhase()) || phase == nil) {
			filteredAttachmentList = append(filteredAttachmentList, attachmentSpec)
		}
	}
	return filteredAttachmentList
}

func handleNetworkingPrivateLinkAttachmentUpdate(t *testing.T, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch id {
		case "platt-invalid":
			w.WriteHeader(http.StatusNotFound)
			return
		case "platt-111111":
			body := &networkingprivatelinkv1.NetworkingV1PrivateLinkAttachment{}
			err := json.NewDecoder(r.Body).Decode(body)
			require.NoError(t, err)

			attachment := getPrivateLinkAttachment("platt-111111", body.Spec.GetDisplayName(), "WAITING_FOR_CONNECTIONS")
			err = json.NewEncoder(w).Encode(attachment)
			require.NoError(t, err)
		}
	}
}

func handleNetworkingPrivateLinkAttachmentDelete(_ *testing.T, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch id {
		case "platt-invalid":
			w.WriteHeader(http.StatusNotFound)
			return
		case "platt-111111", "platt-222222":
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func handleNetworkingPrivateLinkAttachmentCreate(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := &networkingprivatelinkv1.NetworkingV1PrivateLinkAttachment{}
		err := json.NewDecoder(r.Body).Decode(body)
		require.NoError(t, err)

		if body.Spec.DisplayName == nil {
			w.WriteHeader(http.StatusBadRequest)
			err := writeErrorJson(w, "The private link attachment name must be provided.")
			require.NoError(t, err)
			return
		}

		cloud := body.Spec.GetCloud()
		region := body.Spec.GetRegion()

		switch cloud {
		case "AWS":
			attachment := networkingprivatelinkv1.NetworkingV1PrivateLinkAttachment{
				Id:   networkingprivatelinkv1.PtrString("pla-123456"),
				Kind: networkingprivatelinkv1.PtrString("PrivateLinkAttachment"),
				Spec: &networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentSpec{
					Cloud:       networkingprivatelinkv1.PtrString(cloud),
					Region:      networkingprivatelinkv1.PtrString(region),
					DisplayName: networkingprivatelinkv1.PtrString(body.Spec.GetDisplayName()),
					Environment: &networkingprivatelinkv1.ObjectReference{Id: body.Spec.Environment.GetId()},
				},
				Status: &networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentStatus{
					Phase: "PROVISIONING",
				},
			}
			err = json.NewEncoder(w).Encode(attachment)
			require.NoError(t, err)
		default:
			w.WriteHeader(http.StatusBadRequest)
			err := writeErrorJson(w, fmt.Sprintf("The private link attachment region '%s' for the cloud provider '%s' is not supported.", region, strings.ToLower(cloud)))
			require.NoError(t, err)
		}
	}
}

func handleNetworkingPrivateLinkAttachmentConnectionGet(t *testing.T, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch id {
		case "plattc-invalid":
			w.WriteHeader(http.StatusNotFound)
			return
		case "plattc-111111":
			connection := getPrivateLinkAttachmentConnection("plattc-111111", "aws-plattc", "READY")
			err := json.NewEncoder(w).Encode(connection)
			require.NoError(t, err)
		case "plattc-111112":
			connection := getPrivateLinkAttachmentConnection("plattc-111112", "aws-plattc", "PROVISIONING")
			err := json.NewEncoder(w).Encode(connection)
			require.NoError(t, err)
		}
	}
}

func getPrivateLinkAttachmentConnection(id, name, phase string) networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentConnection {
	connection := networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentConnection{
		Id: networkingv1.PtrString(id),
		Spec: &networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentConnectionSpec{
			PrivateLinkAttachment: &networkingprivatelinkv1.ObjectReference{Id: "platt-111111"},
			Cloud: &networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentConnectionSpecCloudOneOf{
				NetworkingV1AwsPrivateLinkAttachmentConnection: &networkingprivatelinkv1.NetworkingV1AwsPrivateLinkAttachmentConnection{
					Kind:          "AwsPrivateLinkAttachmentConnection",
					VpcEndpointId: "vpce-01234567890abcdef",
				},
			},
			DisplayName: networkingprivatelinkv1.PtrString(name),
			Environment: &networkingprivatelinkv1.ObjectReference{Id: "env-00000"},
		},
		Status: &networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentConnectionStatus{
			Phase: phase,
			Cloud: &networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentConnectionStatusCloudOneOf{
				NetworkingV1AwsPrivateLinkAttachmentConnectionStatus: &networkingprivatelinkv1.NetworkingV1AwsPrivateLinkAttachmentConnectionStatus{
					Kind:                   "AwsPrivateLinkAttachmentConnectionStatus",
					VpcEndpointId:          "vpce-01234567890abcdef",
					VpcEndpointServiceName: "",
				},
			},
		},
	}

	if phase != "PROVISIONING" {
		connection.Status.Cloud.NetworkingV1AwsPrivateLinkAttachmentConnectionStatus.VpcEndpointServiceName = "com.amazonaws.vpce.us-west-2.vpce-svc-01234567890abcdef"
	}

	return connection
}

func handleNetworkingPrivateLinkAttachmentConnectionList(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		connection1 := getPrivateLinkAttachmentConnection("plattc-111111", "aws-plattc-1", "PROVISIONING")
		connection2 := getPrivateLinkAttachmentConnection("plattc-111112", "aws-plattc-2", "READY")
		connection3 := getPrivateLinkAttachmentConnection("plattc-111113", "aws-plattc-3", "READY")

		privateLinkAttachment := r.URL.Query().Get("spec.private_link_attachment")
		switch privateLinkAttachment {
		case "platt-invalid":
			err := json.NewEncoder(w).Encode(networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentConnectionList{
				Data:     []networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentConnection{},
				Metadata: networkingprivatelinkv1.ListMeta{},
			})
			require.NoError(t, err)
		case "platt-111111":
			pageToken := r.URL.Query().Get("page_token")
			var connectionList networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentConnectionList
			switch pageToken {
			case "aws-plattc-3":
				connectionList = networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentConnectionList{
					Data:     []networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentConnection{connection3},
					Metadata: networkingprivatelinkv1.ListMeta{},
				}
			case "aws-plattc-2":
				connectionList = networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentConnectionList{
					Data:     []networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentConnection{connection2},
					Metadata: networkingprivatelinkv1.ListMeta{Next: *networkingprivatelinkv1.NewNullableString(networkingprivatelinkv1.PtrString("/networking/v1/private-link-attachments?environment=env-00000&page_size=1&page_token=aws-plattc-3"))},
				}
			default:
				connectionList = networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentConnectionList{
					Data:     []networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentConnection{connection1},
					Metadata: networkingprivatelinkv1.ListMeta{Next: *networkingprivatelinkv1.NewNullableString(networkingprivatelinkv1.PtrString("/networking/v1/private-link-attachments?environment=env-00000&page_size=1&page_token=aws-plattc-2"))},
				}
			}

			err := json.NewEncoder(w).Encode(connectionList)
			require.NoError(t, err)
		}
	}
}

func handleNetworkingPrivateLinkAttachmentConnectionUpdate(t *testing.T, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch id {
		case "plattc-invalid":
			w.WriteHeader(http.StatusNotFound)
			return
		case "plattc-111111":
			body := &networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentConnection{}
			err := json.NewDecoder(r.Body).Decode(body)
			require.NoError(t, err)

			connection := getPrivateLinkAttachmentConnection("plattc-111111", body.Spec.GetDisplayName(), "READY")
			err = json.NewEncoder(w).Encode(connection)
			require.NoError(t, err)
		}
	}
}

func handleNetworkingPrivateLinkAttachmentConnectionDelete(_ *testing.T, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch id {
		case "plattc-invalid":
			w.WriteHeader(http.StatusNotFound)
			return
		case "plattc-111111", "plattc-222222":
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func handleNetworkingPrivateLinkAttachmentConnectionCreate(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := &networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentConnection{}
		err := json.NewDecoder(r.Body).Decode(body)
		require.NoError(t, err)

		if body.Spec.DisplayName == nil {
			w.WriteHeader(http.StatusBadRequest)
			err := writeErrorJson(w, "The private link attachment connection name must be provided.")
			require.NoError(t, err)
			return
		}

		if body.Spec.Cloud == nil {
			w.WriteHeader(http.StatusBadRequest)
			err := writeErrorJson(w, "The request body is malformed (missing cloud).")
			require.NoError(t, err)
			return
		}

		name := body.Spec.GetDisplayName()

		switch body.Spec.Cloud.GetActualInstance().(type) {
		case *networkingprivatelinkv1.NetworkingV1AwsPrivateLinkAttachmentConnection:
			switch name {
			case "aws-plattc":
				connection := networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentConnection{
					Id:   networkingprivatelinkv1.PtrString("plattc-123456"),
					Kind: networkingprivatelinkv1.PtrString("PrivateLinkAttachmentConnection"),
					Spec: &networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentConnectionSpec{
						Cloud:                 body.Spec.Cloud,
						DisplayName:           body.Spec.DisplayName,
						Environment:           &networkingprivatelinkv1.ObjectReference{Id: body.Spec.Environment.GetId()},
						PrivateLinkAttachment: &networkingprivatelinkv1.ObjectReference{Id: body.Spec.PrivateLinkAttachment.GetId()},
					},
					Status: &networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentConnectionStatus{
						Phase: "PROVISIONING",
						Cloud: &networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentConnectionStatusCloudOneOf{
							NetworkingV1AwsPrivateLinkAttachmentConnectionStatus: &networkingprivatelinkv1.NetworkingV1AwsPrivateLinkAttachmentConnectionStatus{
								Kind:                   "AwsPrivateLinkAttachmentConnectionStatus",
								VpcEndpointId:          body.Spec.Cloud.NetworkingV1AwsPrivateLinkAttachmentConnection.GetVpcEndpointId(),
								VpcEndpointServiceName: "",
							},
						},
					},
				}
				err = json.NewEncoder(w).Encode(connection)
				require.NoError(t, err)
			case "aws-plattc-wrong-endpoint":
				w.WriteHeader(http.StatusBadRequest)
				err := writeErrorJson(w, "The AWS VPC Endpoint ID is invalid.")
				require.NoError(t, err)
			case "aws-plattc-invalid-platt":
				w.WriteHeader(http.StatusBadRequest)
				err := writeErrorJson(w, "The private link attachment was not found.")
				require.NoError(t, err)
			}
		case *networkingprivatelinkv1.NetworkingV1GcpPrivateLinkAttachmentConnection:
			switch name {
			case "gcp-plattc-wrong-platt-cloud":
				w.WriteHeader(http.StatusBadRequest)
				err := writeErrorJson(w, "The AWS spec for the private link attachment connection must be provided.")
				require.NoError(t, err)
			}
		}
	}
}

func handleNetworkingNetworkLinkServiceGet(t *testing.T, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch id {
		case "nls-invalid":
			w.WriteHeader(http.StatusNotFound)
			err := writeErrorJson(w, "The network link service nls-invalid was not found.")
			require.NoError(t, err)
		default:
			nls := getNetworkLinkService(id, "my-network-link-service")
			err := json.NewEncoder(w).Encode(nls)
			require.NoError(t, err)
		}
	}
}

func getNetworkLinkService(id, name string) networkingv1.NetworkingV1NetworkLinkService {
	service := networkingv1.NetworkingV1NetworkLinkService{
		Id: networkingv1.PtrString(id),
		Spec: &networkingv1.NetworkingV1NetworkLinkServiceSpec{
			DisplayName: networkingv1.PtrString(name),
			Description: networkingv1.PtrString("example network link service"),
			Accept: &networkingv1.NetworkingV1NetworkLinkServiceAcceptPolicy{
				Networks:     &[]string{"n-abcde2", "n-abcde3"},
				Environments: &[]string{"env-11111", "env-22222"},
			},
			Environment: &networkingv1.GlobalObjectReference{Id: "env-00000"},
			Network:     &networkingv1.EnvScopedObjectReference{Id: "n-abcde1"},
		},
		Status: &networkingv1.NetworkingV1NetworkLinkServiceStatus{Phase: "READY"},
	}

	switch id {
	case "nls-111111":
		service.Spec.Accept = &networkingv1.NetworkingV1NetworkLinkServiceAcceptPolicy{
			Networks:     &[]string{"n-abcde2"},
			Environments: &[]string{"env-11111"},
		}
	case "nls-333333":
		service.Spec.Accept = &networkingv1.NetworkingV1NetworkLinkServiceAcceptPolicy{
			Networks:     &[]string{"n-abcde2", "n-abcde3", "n-abcde4"},
			Environments: &[]string{"env-11111", "env-22222", "env-33333"},
		}
	}

	return service
}

func handleNetworkingNetworkLinkServiceList(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		name := q["spec.display_name"]
		network := q["spec.network"]
		phase := q["status.phase"]

		nlsList := getNetworkLinkServiceList(name, network, phase)
		err := json.NewEncoder(w).Encode(nlsList)
		require.NoError(t, err)
	}
}

func getNetworkLinkServiceList(filterName, filterNetwork, filterPhase []string) networkingv1.NetworkingV1NetworkLinkServiceList {
	nlsList := networkingv1.NetworkingV1NetworkLinkServiceList{
		Data: []networkingv1.NetworkingV1NetworkLinkService{
			getNetworkLinkService("nls-111111", "my-network-link-service-1"),
			getNetworkLinkService("nls-222222", "my-network-link-service-2"),
			getNetworkLinkService("nls-333333", "my-network-link-service-3"),
		},
	}
	nlsList.Data = filterServiceList(nlsList.Data, filterName, filterNetwork, filterPhase)

	return nlsList
}

func filterServiceList(serviceList []networkingv1.NetworkingV1NetworkLinkService, name, network, phase []string) []networkingv1.NetworkingV1NetworkLinkService {
	var filteredServiceList []networkingv1.NetworkingV1NetworkLinkService
	for _, serviceSpec := range serviceList {
		if (slices.Contains(name, serviceSpec.Spec.GetDisplayName()) || name == nil) &&
			(slices.Contains(network, serviceSpec.Spec.Network.GetId()) || network == nil) &&
			(slices.Contains(phase, serviceSpec.Status.GetPhase()) || phase == nil) {
			filteredServiceList = append(filteredServiceList, serviceSpec)
		}
	}
	return filteredServiceList
}

func handleNetworkingNetworkLinkServiceDelete(_ *testing.T, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch id {
		case "nls-invalid":
			w.WriteHeader(http.StatusNotFound)
			return
		case "nls-111111", "nls-222222":
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func handleNetworkingNetworkLinkServiceCreate(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := &networkingv1.NetworkingV1NetworkLinkService{}
		err := json.NewDecoder(r.Body).Decode(body)
		require.NoError(t, err)

		name := body.Spec.GetDisplayName()
		switch name {
		case "nls-duplicate":
			w.WriteHeader(http.StatusConflict)
			err := writeErrorJson(w, "Exceeds the maximum number of network link service (1) supported per network.")
			require.NoError(t, err)
		default:
			service := networkingv1.NetworkingV1NetworkLinkService{
				Id: networkingv1.PtrString("nls-abcde1"),
				Spec: &networkingv1.NetworkingV1NetworkLinkServiceSpec{
					DisplayName: networkingv1.PtrString(body.Spec.GetDisplayName()),
					Description: networkingv1.PtrString(body.Spec.GetDescription()),
					Accept: &networkingv1.NetworkingV1NetworkLinkServiceAcceptPolicy{
						Networks:     body.Spec.Accept.Networks,
						Environments: body.Spec.Accept.Environments,
					},
					Environment: &networkingv1.GlobalObjectReference{Id: "env-00000"},
					Network:     body.Spec.Network,
				},
				Status: &networkingv1.NetworkingV1NetworkLinkServiceStatus{Phase: "READY"},
			}
			err = json.NewEncoder(w).Encode(service)
			require.NoError(t, err)
		}
	}
}

func handleNetworkingNetworkLinkServiceUpdate(t *testing.T, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch id {
		case "nls-invalid":
			w.WriteHeader(http.StatusNotFound)
			err := writeErrorJson(w, "The network link service nls-invalid was not found.")
			require.NoError(t, err)
		default:
			body := &networkingv1.NetworkingV1NetworkLinkService{}
			err := json.NewDecoder(r.Body).Decode(body)
			require.NoError(t, err)

			service := getNetworkLinkService("nls-111111", "my-network-link-service")
			if body.Spec.DisplayName != nil {
				service.Spec.SetDisplayName(body.Spec.GetDisplayName())
			}
			if body.Spec.Description != nil {
				service.Spec.SetDescription(body.Spec.GetDescription())
			}
			if body.Spec.Accept.GetNetworks() != nil {
				service.Spec.Accept.SetNetworks(body.Spec.Accept.GetNetworks())
			}
			if body.Spec.Accept.GetEnvironments() != nil {
				service.Spec.Accept.SetEnvironments(body.Spec.Accept.GetEnvironments())
			}
			err = json.NewEncoder(w).Encode(service)
			require.NoError(t, err)
		}
	}
}

func handleNetworkingNetworkLinkEndpointGet(t *testing.T, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch id {
		case "nle-invalid":
			w.WriteHeader(http.StatusNotFound)
			err := writeErrorJson(w, "The network link endpoint nle-invalid was not found.")
			require.NoError(t, err)
		case "nle-123456":
			nle := getNetworkLinkEndpoint("nle-123456", "my-network-link-endpoint")
			err := json.NewEncoder(w).Encode(nle)
			require.NoError(t, err)
		}
	}
}

func getNetworkLinkEndpoint(id, name string) networkingv1.NetworkingV1NetworkLinkEndpoint {
	endpoint := networkingv1.NetworkingV1NetworkLinkEndpoint{
		Id: networkingv1.PtrString(id),
		Spec: &networkingv1.NetworkingV1NetworkLinkEndpointSpec{
			DisplayName:        networkingv1.PtrString(name),
			Description:        networkingv1.PtrString("example network link endpoint"),
			Environment:        &networkingv1.GlobalObjectReference{Id: "env-00000"},
			Network:            &networkingv1.EnvScopedObjectReference{Id: "n-abcde1"},
			NetworkLinkService: &networkingv1.EnvScopedObjectReference{Id: "nls-123456"},
		},
		Status: &networkingv1.NetworkingV1NetworkLinkEndpointStatus{Phase: "READY"},
	}

	return endpoint
}

func handleNetworkingNetworkLinkEndpointList(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		name := q["spec.display_name"]
		network := q["spec.network"]
		phase := q["status.phase"]
		service := q["spec.network_link_service"]

		nleList := getNetworkLinkEndpointList(name, network, phase, service)
		err := json.NewEncoder(w).Encode(nleList)
		require.NoError(t, err)
	}
}

func getNetworkLinkEndpointList(filterName, filterNetwork, filterPhase, filterService []string) networkingv1.NetworkingV1NetworkLinkEndpointList {
	nleList := networkingv1.NetworkingV1NetworkLinkEndpointList{
		Data: []networkingv1.NetworkingV1NetworkLinkEndpoint{
			getNetworkLinkEndpoint("nle-111111", "my-network-link-endpoint-1"),
			getNetworkLinkEndpoint("nle-222222", "my-network-link-endpoint-2"),
			getNetworkLinkEndpoint("nle-333333", "my-network-link-endpoint-3"),
		},
	}
	nleList.Data = filterEndpointList(nleList.Data, filterName, filterNetwork, filterPhase, filterService)

	return nleList
}

func filterEndpointList(endpointList []networkingv1.NetworkingV1NetworkLinkEndpoint, name, network, phase, service []string) []networkingv1.NetworkingV1NetworkLinkEndpoint {
	var filteredEndpointList []networkingv1.NetworkingV1NetworkLinkEndpoint
	for _, endpointSpec := range endpointList {
		if (slices.Contains(name, endpointSpec.Spec.GetDisplayName()) || name == nil) &&
			(slices.Contains(network, endpointSpec.Spec.Network.GetId()) || network == nil) &&
			(slices.Contains(phase, endpointSpec.Status.GetPhase()) || phase == nil) &&
			(slices.Contains(service, endpointSpec.Spec.NetworkLinkService.GetId()) || service == nil) {
			filteredEndpointList = append(filteredEndpointList, endpointSpec)
		}
	}
	return filteredEndpointList
}

func handleNetworkingNetworkLinkEndpointDelete(_ *testing.T, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch id {
		case "nle-invalid":
			w.WriteHeader(http.StatusNotFound)
		case "nle-111111", "nle-222222":
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func handleNetworkingNetworkLinkEndpointCreate(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := &networkingv1.NetworkingV1NetworkLinkEndpoint{}
		err := json.NewDecoder(r.Body).Decode(body)
		require.NoError(t, err)

		switch body.Spec.GetDisplayName() {
		case "nle-duplicate":
			w.WriteHeader(http.StatusConflict)
			err := writeErrorJson(w, "Cannot have more than 1 active/in provision/pending accept/disconnected NetworkLinkEndpoint attached to the same NetworkLinkService in the same network.")
			require.NoError(t, err)
		case "nle-same-id":
			w.WriteHeader(http.StatusConflict)
			err := writeErrorJson(w, "Network ID for NetworkLinkEndpoint and NetworkLinkService cannot be the same.")
			require.NoError(t, err)
		default:
			endpoint := networkingv1.NetworkingV1NetworkLinkEndpoint{
				Id: networkingv1.PtrString("nle-abcde1"),
				Spec: &networkingv1.NetworkingV1NetworkLinkEndpointSpec{
					DisplayName:        networkingv1.PtrString(body.Spec.GetDisplayName()),
					Description:        networkingv1.PtrString(body.Spec.GetDescription()),
					Environment:        &networkingv1.GlobalObjectReference{Id: "env-00000"},
					Network:            body.Spec.Network,
					NetworkLinkService: &networkingv1.EnvScopedObjectReference{Id: body.Spec.NetworkLinkService.GetId()},
				},
				Status: &networkingv1.NetworkingV1NetworkLinkEndpointStatus{Phase: "PENDING_ACCEPT"},
			}
			err = json.NewEncoder(w).Encode(endpoint)
			require.NoError(t, err)
		}
	}
}

func handleNetworkingNetworkLinkEndpointUpdate(t *testing.T, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch id {
		case "nle-invalid":
			w.WriteHeader(http.StatusNotFound)
			err := writeErrorJson(w, "The network link endpoint nle-invalid was not found.")
			require.NoError(t, err)
		default:
			body := &networkingv1.NetworkingV1NetworkLinkEndpoint{}
			err := json.NewDecoder(r.Body).Decode(body)
			require.NoError(t, err)

			endpoint := getNetworkLinkEndpoint("nle-111111", "my-network-link-endpoint")
			if body.Spec.DisplayName != nil {
				endpoint.Spec.SetDisplayName(body.Spec.GetDisplayName())
			}
			if body.Spec.Description != nil {
				endpoint.Spec.SetDescription(body.Spec.GetDescription())
			}
			err = json.NewEncoder(w).Encode(endpoint)
			require.NoError(t, err)
		}
	}
}

func handleNetworkingIpAddressList(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		cloud := q["cloud"]
		region := q["region"]
		services := q["services"]
		addressType := q["address_type"]

		ipList := getIpAddressList(cloud, region, services, addressType)
		err := json.NewEncoder(w).Encode(ipList)
		require.NoError(t, err)
	}
}

func getIpAddressList(filterCloud, filterRegion, filterServices, filterAddressType []string) networkingipv1.NetworkingV1IpAddressList {
	ipList := networkingipv1.NetworkingV1IpAddressList{
		Data: []networkingipv1.NetworkingV1IpAddress{
			getIpAddress("10.200.0.0/28", "ANY", "global", []string{"EXTERNAL_OAUTH"}),
			getIpAddress("10.201.0.0/28", "AWS", "us-west-2", []string{"CONNECT"}),
			getIpAddress("10.201.0.0/28", "AWS", "us-east-1", []string{"CONNECT"}),
			getIpAddress("10.202.0.0/28", "AWS", "us-east-1", []string{"CONNECT"}),
			getIpAddress("10.202.0.0/28", "GCP", "us-central1", []string{"KAFKA"}),
			getIpAddress("10.203.0.0/28", "AZURE", "centralus", []string{"KAFKA", "CONNECT"}),
		},
	}
	ipList.Data = filterIpAddressList(ipList.Data, filterCloud, filterRegion, filterServices, filterAddressType)

	return ipList
}

func filterIpAddressList(ipList []networkingipv1.NetworkingV1IpAddress, cloud, region, services, addressType []string) []networkingipv1.NetworkingV1IpAddress {
	var filteredIpAddressList []networkingipv1.NetworkingV1IpAddress
	for _, ipSpec := range ipList {
		if (slices.Contains(cloud, ipSpec.GetCloud()) || cloud == nil) &&
			(slices.Contains(region, ipSpec.GetRegion()) || region == nil) &&
			(slices.Contains(addressType, ipSpec.GetAddressType()) || addressType == nil) &&
			(containsFilter(services, ipSpec.GetServices().Items) || services == nil) {
			filteredIpAddressList = append(filteredIpAddressList, ipSpec)
		}
	}
	return filteredIpAddressList
}

func getIpAddress(ipPrefix, cloud, region string, services []string) networkingipv1.NetworkingV1IpAddress {
	return networkingipv1.NetworkingV1IpAddress{
		Kind:        networkingipv1.PtrString("IpAddress"),
		AddressType: networkingipv1.PtrString("EGRESS"),
		IpPrefix:    networkingipv1.PtrString(ipPrefix),
		Cloud:       networkingipv1.PtrString(cloud),
		Region:      networkingipv1.PtrString(region),
		Services:    &networkingipv1.Set{Items: services},
	}
}
