# Task Summary: Azure & GCP Ingress Gateway/Access Point CLI Support (PR #3283)

## Context

PR #3283 adds Azure Private Link and GCP Private Service Connect (PSC) ingress gateway and access point support to the `confluent` CLI, alongside existing AWS ingress support. The branch is `ipl-support-azure-gcp-accesspoint`.

---

## Files Changed

### 1. `internal/network/command_gateway.go`

**Add new constants** (alphabetically in the `const` block):
```go
gcpIngressPrivateServiceConnect = "GcpIngressPrivateServiceConnect"
azureIngressPrivateLink         = "AzureIngressPrivateLink"
gcpEgressPrivateLink            = "GcpEgressPrivateLink"
gcpIngressPrivateLink           = "GcpIngressPrivateLink"
```

**Update `createGatewayTypes`** to include GCP type:
```go
createGatewayTypes = []string{"egress-privatelink", "ingress-privatelink", "private-network-interface", "ingress-private-service-connect"}
```

**Update `listGatewayTypes`** (drop `gcp-*-privatelink` aliases, keep only PSC forms):
```go
listGatewayTypes = []string{"aws-egress-privatelink", "aws-ingress-privatelink", "azure-egress-privatelink", "azure-ingress-privatelink", "gcp-egress-private-service-connect", "gcp-ingress-private-service-connect"}
```

**Update `gatewayTypeMap`** — GCP filter values must use `gcpEgressPrivateLink`/`gcpIngressPrivateLink` (i.e., `"GcpEgressPrivateLink"` / `"GcpIngressPrivateLink"`), NOT the display-type constants. The API list filter uses different strings than the describe output:
```go
gatewayTypeMap = map[string]string{
    "aws-egress-privatelink":              awsEgressPrivateLink,
    "aws-ingress-privatelink":             awsIngressPrivateLink,
    "azure-egress-privatelink":            azureEgressPrivateLink,
    "azure-ingress-privatelink":           azureIngressPrivateLink,
    "gcp-egress-private-service-connect":  gcpEgressPrivateLink,
    "gcp-ingress-private-service-connect": gcpIngressPrivateLink,
}
```

**Update `gatewayOut` struct** — rename `VpcEndpointServiceName` human label to include `AWS` prefix:
```go
VpcEndpointServiceName string `human:"AWS VPC Endpoint Service Name,omitempty" serialized:"aws_vpc_endpoint_service_name,omitempty"`
```

Add new output fields:
```go
AzurePrivateLinkServiceAlias              string   `human:"Azure Private Link Service Alias,omitempty" ...`
AzurePrivateLinkServiceResourceId         string   `human:"Azure Private Link Resource ID,omitempty" ...`
GcpPrivateServiceConnectServiceAttachment string   `human:"GCP PSC Service Attachment,omitempty" ...`
```

**Update `addRegionFlagGateway`** description:
```go
cmd.Flags().String("region", "", "AWS, Azure, or GCP region of the gateway.")
```

**Add `getGatewayCloud`** handling for new types and update `getGatewayType` / `printGatewayTable` for `azureIngressPrivateLink` and `gcpIngressPrivateServiceConnect`.

---

### 2. `internal/network/command_gateway_create.go`

Add Azure ingress and GCP ingress cases to the switch:
```go
case pcloud.Azure:
    // existing egress-privatelink case ...
    } else if gatewayType == "ingress-privatelink" {
        createGateway.Spec.Config = &networkinggatewayv1.NetworkingV1GatewaySpecConfigOneOf{
            NetworkingV1AzureIngressPrivateLinkGatewaySpec: &networkinggatewayv1.NetworkingV1AzureIngressPrivateLinkGatewaySpec{
                Kind: "AzureIngressPrivateLinkGatewaySpec", Region: region,
            },
        }
    }
case pcloud.Gcp:
    if gatewayType == "ingress-private-service-connect" {
        createGateway.Spec.Config = &networkinggatewayv1.NetworkingV1GatewaySpecConfigOneOf{
            NetworkingV1GcpIngressPrivateServiceConnectGatewaySpec: &networkinggatewayv1.NetworkingV1GcpIngressPrivateServiceConnectGatewaySpec{
                Kind: "GcpIngressPrivateServiceConnectGatewaySpec", Region: region,
            },
        }
    }
```

**Validate nil Config** (catch unsupported cloud+type combinations):
```go
if createGateway.Spec.Config == nil {
    return fmt.Errorf("type %q is not supported for --cloud %s", gatewayType, strings.ToLower(cloud))
}
```

**Validate `--zones` flag** (only valid for `private-network-interface`):
```go
if len(zones) > 0 && gatewayType != "private-network-interface" {
    return fmt.Errorf("flag \"--zones\" is only valid for --type private-network-interface")
}
```

**Change cloud flag** from `AddCloudAwsAzureFlag` to `AddCloudFlag` (to allow GCP).

---

### 3. `internal/network/command_gateway_list.go`

Add region handling for new types:
```go
if gatewayType == azureIngressPrivateLink {
    out.Region = gateway.Spec.Config.NetworkingV1AzureIngressPrivateLinkGatewaySpec.GetRegion()
}
if gatewayType == gcpIngressPrivateServiceConnect {
    out.Region = gateway.Spec.Config.NetworkingV1GcpIngressPrivateServiceConnectGatewaySpec.GetRegion()
}
```

Also fix pre-existing inconsistency — AWS type comparisons must use constants (not string literals):
```go
case pcloud.Aws:
    if gatewayType == awsEgressPrivateLink { ... }
    else if gatewayType == awsIngressPrivateLink { ... }
    else if gatewayType == awsPrivateNetworkInterface { ... }
```

Add status fields for new types in the switch:
```go
case pcloud.Azure:
    } else if gatewayType == azureIngressPrivateLink {
        out.AzurePrivateLinkServiceAlias = gateway.Status.CloudGateway.NetworkingV1AzureIngressPrivateLinkGatewayStatus.GetPrivateLinkServiceAlias()
        out.AzurePrivateLinkServiceResourceId = gateway.Status.CloudGateway.NetworkingV1AzureIngressPrivateLinkGatewayStatus.GetPrivateLinkServiceResourceId()
case pcloud.Gcp:
    } else if gatewayType == gcpIngressPrivateServiceConnect {
        out.GcpPrivateServiceConnectServiceAttachment = gateway.Status.CloudGateway.NetworkingV1GcpIngressPrivateServiceConnectGatewayStatus.GetPrivateServiceConnectServiceAttachment()
```

---

### 4. `internal/network/command_access_point_private_link_ingress_endpoint.go`

**Add new fields to `ingressEndpointOut`**:
```go
AzurePrivateLinkServiceAlias              string `human:"Azure Private Link Service Alias,omitempty" ...`
AzurePrivateLinkServiceResourceId         string `human:"Azure Private Link Service Resource ID,omitempty" ...`
AzurePrivateEndpointResourceId            string `human:"Azure Private Endpoint Resource ID,omitempty" ...`
GcpPrivateServiceConnectServiceAttachment string `human:"GCP PSC Service Attachment,omitempty" ...`
GcpPrivateServiceConnectConnectionId      string `human:"GCP PSC Connection ID,omitempty" ...`
ErrorMessage                              string `human:"Error Message,omitempty" serialized:"error_message,omitempty"`
```

**Populate new fields** from status in `printPrivateLinkIngressEndpointTable` and the list function:
```go
out.ErrorMessage = ingressEndpoint.Status.GetErrorMessage()

if ingressEndpoint.Status.Config != nil && ingressEndpoint.Status.Config.NetworkingV1AzureIngressPrivateLinkEndpointStatus != nil {
    out.AzurePrivateLinkServiceAlias = ...GetPrivateLinkServiceAlias()
    out.AzurePrivateEndpointResourceId = ...GetPrivateEndpointResourceId()
    out.DnsDomain = ...GetDnsDomain()
}
if ingressEndpoint.Status.Config != nil && ingressEndpoint.Status.Config.NetworkingV1GcpIngressPrivateServiceConnectEndpointStatus != nil {
    out.GcpPrivateServiceConnectServiceAttachment = ...GetPrivateServiceConnectServiceAttachment()
    out.GcpPrivateServiceConnectConnectionId = ...GetPrivateServiceConnectConnectionId()
    out.DnsDomain = ...GetDnsDomain()
}
```

**Update autocomplete** to include Azure and GCP ingress endpoints:
```go
ingressEndpoints := slices.DeleteFunc(accessPoints, func(ap networkingaccesspointv1.NetworkingV1AccessPoint) bool {
    return ap.Spec.GetConfig().NetworkingV1AwsIngressPrivateLinkEndpoint == nil &&
        ap.Spec.GetConfig().NetworkingV1AzureIngressPrivateLinkEndpoint == nil &&
        ap.Spec.GetConfig().NetworkingV1GcpIngressPrivateServiceConnectEndpoint == nil
})
```

---

### 5. `internal/network/command_access_point_private_link_ingress_endpoint_create.go`

**New flags** (flag names must be ≤20 chars and ≤2 hyphens):
```go
pcmd.AddCloudFlag(cmd)
addGatewayFlag(cmd, c.AuthenticatedCLICommand)  // required flag BEFORE optional flags
cmd.Flags().String("vpc-endpoint-id", "", "ID of an AWS VPC endpoint; only valid with --cloud aws.")
cmd.Flags().String("endpoint-resource-id", "", "Resource ID of an Azure Private Endpoint; only valid with --cloud azure.")
cmd.Flags().String("psc-connection-id", "", "ID of a GCP Private Service Connect connection; only valid with --cloud gcp.")
```

**Mark mutually exclusive and required**:
```go
cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
cobra.CheckErr(cmd.MarkFlagRequired("gateway"))
cmd.MarkFlagsMutuallyExclusive("vpc-endpoint-id", "endpoint-resource-id", "psc-connection-id")
```

**Handle each cloud** in `createIngressEndpoint`:
```go
case pcloud.Azure:
    if endpointResourceId == "" {
        return fmt.Errorf("flag \"endpoint-resource-id\" is required for --cloud azure")
    }
    // set AzureIngressPrivateLinkEndpoint config
case pcloud.Gcp:
    if pscConnectionId == "" {
        return fmt.Errorf("flag \"psc-connection-id\" is required for --cloud gcp")
    }
    // set GcpIngressPrivateServiceConnectEndpoint config
```

---

### 6. `go.mod` / `go.sum`

Replace internal SDK replace directives with public versions:
- `networking-gateway`: `v0.5.0` → `v0.7.0`
- `networking-access-point`: `v0.8.0` → `v0.10.0`

Remove the `replace` block entirely:
```
# REMOVE:
replace (
    github.com/confluentinc/ccloud-sdk-go-v2/networking-access-point => github.com/confluentinc/ccloud-sdk-go-v2-internal/networking-access-point v0.13.0
    github.com/confluentinc/ccloud-sdk-go-v2/networking-gateway => github.com/confluentinc/ccloud-sdk-go-v2-internal/networking-gateway v0.13.0
)
```

Run `go mod tidy` after.

---

### 7. `cmd/lint/main.go`

Add `"psc"` to `vocabWords` slice alphabetically between `"protobuf"` and `"rbac"`:
```go
"protobuf",
"psc",
"rbac",
```

---

### 8. `test/test-server/networking_handlers.go`

- Add `gw-11111` (Azure ingress) and `gw-99999` (GCP ingress) to the gateway list handler
- Add `ap-11111` (Azure ingress) and `ap-22222` (GCP ingress) to the access point list handler
- Update `getGatewayTypeFromSpec` to return `"GcpEgressPrivateLink"` / `"GcpIngressPrivateLink"` for GCP PSC types (to match the API filter values)
- Add `handleNetworkingGatewayPost` cases for `AzureIngressPrivateLinkGatewaySpec` and `GcpIngressPrivateServiceConnectGatewaySpec`
- Add access point create/get handlers for Azure and GCP ingress types

---

## Golden Files to Regenerate

After making all code changes, run with `-update` to regenerate golden files:

```bash
GOCOVERDIR=$(mktemp -d) go test ./test/... -update -run "TestCLI/TestNetworkGateway$"
GOCOVERDIR=$(mktemp -d) go test ./test/... -update -run "TestCLI/TestNetworkGatewayList"
GOCOVERDIR=$(mktemp -d) go test ./test/... -update -run "TestCLI/TestNetworkGatewayDescribe"
GOCOVERDIR=$(mktemp -d) go test ./test/... -update -run "TestCLI/TestNetworkAccessPointPrivateLinkIngressEndpoint$"
GOCOVERDIR=$(mktemp -d) go test ./test/... -update -run "TestCLI/TestNetworkAccessPointPrivateLinkIngressEndpointList"
GOCOVERDIR=$(mktemp -d) go test ./test/... -update -run "TestCLI/TestNetworkDnsRecord_Autocomplete"
GOCOVERDIR=$(mktemp -d) go test ./test/... -update -run "TestCLI/TestHelp"
```

---

## New Integration Tests to Add

In `test/network_test.go`:

**Gateway create errors:**
```go
{args: "network gateway create my-gateway --cloud aws --type ingress-private-service-connect --region us-west-2", fixture: "network/gateway/create-invalid-cloud-type.golden", exitCode: 1},
```

**Gateway list filters:**
```go
{args: "network gateway list --types azure-ingress-privatelink", fixture: "network/gateway/list-filter-azure-ingress-type.golden"},
{args: "network gateway list --types gcp-ingress-private-service-connect", fixture: "network/gateway/list-filter-gcp-ingress-type.golden"},
```

**Access point create errors:**
```go
{args: "network access-point private-link ingress-endpoint create --cloud azure --gateway gw-11111 --vpc-endpoint-id vpce-1234567890abcdef0", fixture: "network/access-point/private-link/ingress-endpoint/create-wrong-cloud-flag.golden", exitCode: 1},
{args: "network access-point private-link ingress-endpoint create --cloud aws --gateway gw-88888 --vpc-endpoint-id vpce-1234567890abcdef0 --endpoint-resource-id /subscriptions/0000000/resourceGroups/rg/providers/Microsoft.Network/privateEndpoints/pe", fixture: "network/access-point/private-link/ingress-endpoint/create-two-cloud-flags.golden", exitCode: 1},
```

**Access point create happy path:**
```go
{args: "network access-point private-link ingress-endpoint create --cloud azure --gateway gw-11111 --endpoint-resource-id /subscriptions/0000000/resourceGroups/resourceGroupName/providers/Microsoft.Network/privateEndpoints/privateEndpointName", fixture: "network/access-point/private-link/ingress-endpoint/create-azure.golden"},
{args: "network access-point private-link ingress-endpoint create --cloud gcp --gateway gw-99999 --psc-connection-id 111111111111111111", fixture: "network/access-point/private-link/ingress-endpoint/create-gcp.golden"},
```

---

## Key Decisions / Gotchas

1. **GCP filter type mismatch**: The API list filter for GCP PSC gateways uses `"GcpIngressPrivateLink"` / `"GcpEgressPrivateLink"` — NOT `"GcpIngressPrivateServiceConnect"`. The describe output shows `GcpIngressPrivateServiceConnect` but the filter API uses a different string. Confirmed via staging testing.

2. **Flag naming constraints**: CLI lint enforces max 20 chars and max 2 hyphens per flag name. `--private-endpoint-resource-id` (28 chars, 3 hyphens) and `--private-service-connect-connection-id` (38 chars, 4 hyphens) violate this. Use `--endpoint-resource-id` and `--psc-connection-id` instead.

3. **Required flag ordering**: The CLI lint requires required flags to be registered before optional flags. `addGatewayFlag` must come before `cmd.Flags().String("vpc-endpoint-id", ...)`.

4. **`MarkFlagsMutuallyExclusive` replaces cross-cloud checks**: The per-cloud rejection checks (e.g. "flag X is not valid for --cloud aws") are unreachable because `MarkFlagsMutuallyExclusive` already prevents two flags from being set. Remove the cross-cloud rejection checks; keep only the required-flag checks.

5. **`--zones` for non-PNI types**: `--zones` is silently accepted but ignored for Azure and GCP ingress types. Add an explicit upfront error when `len(zones) > 0` and type is not `private-network-interface`.

6. **`VpcEndpointServiceName` serialized key change**: Renamed from `vpc_endpoint_service_name` to `aws_vpc_endpoint_service_name`. This is technically a breaking change to JSON output — be aware.

---

## Verification Commands

```bash
# Build
export PATH=$PATH:$(go env GOPATH)/bin
make build

# Lint
make lint-cli

# Integration tests
GOCOVERDIR=$(mktemp -d) make integration-test INTEGRATION_TEST_ARGS="-run TestCLI/TestNetworkGateway"
GOCOVERDIR=$(mktemp -d) make integration-test INTEGRATION_TEST_ARGS="-run TestCLI/TestNetworkGatewayList"
GOCOVERDIR=$(mktemp -d) make integration-test INTEGRATION_TEST_ARGS="-run TestCLI/TestNetworkAccessPointPrivateLinkIngressEndpoint"
GOCOVERDIR=$(mktemp -d) make integration-test INTEGRATION_TEST_ARGS="-run TestCLI/TestHelp"
```
