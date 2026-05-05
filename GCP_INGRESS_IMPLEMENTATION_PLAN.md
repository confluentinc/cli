# GCP Ingress Private Link Implementation Plan

## Overview
Add GCP support for ingress private link endpoints, following the pattern established by AWS ingress implementation (commit 31b97e2).

## Prerequisites
- [ ] Verify SDK has `NetworkingV1GcpIngressPrivateServiceConnectEndpoint` types
- [ ] Check if SDK version needs to be updated for GCP ingress support
- [ ] Confirm GCP ingress gateway types exist in `networking-gateway` SDK

---

## Step 1: SDK Version Check
**Goal:** Ensure SDK has the necessary GCP ingress types

**Tasks:**
1. Check current SDK version in `go.mod`
2. Search for GCP ingress types in SDK:
   - `NetworkingV1GcpIngressPrivateServiceConnectEndpoint`
   - `NetworkingV1GcpIngressPrivateServiceConnectEndpointStatus`
   - `NetworkingV1GcpIngressPrivateServiceConnectGatewaySpec`
3. Update SDK versions if needed (like AWS did: networking-access-point v0.5.0 → v0.8.0)

**Files to check:**
- `go.mod`
- SDK documentation/types

---

## Step 2: Update Ingress Endpoint Output Struct
**File:** `internal/network/command_access_point_private_link_ingress_endpoint.go`

**Changes:**
Add GCP-specific fields to `ingressEndpointOut` struct (around line 17-26):

```go
type ingressEndpointOut struct {
    Id                        string `human:"ID" serialized:"id"`
    Name                      string `human:"Name,omitempty" serialized:"name,omitempty"`
    Environment               string `human:"Environment" serialized:"environment"`
    Gateway                   string `human:"Gateway" serialized:"gateway"`
    Phase                     string `human:"Phase" serialized:"phase"`

    // AWS fields
    AwsVpcEndpointId          string `human:"AWS VPC Endpoint ID,omitempty" serialized:"aws_vpc_endpoint_id,omitempty"`
    AwsVpcEndpointServiceName string `human:"AWS VPC Endpoint Service Name,omitempty" serialized:"aws_vpc_endpoint_service_name,omitempty"`
    DnsDomain                 string `human:"DNS Domain,omitempty" serialized:"dns_domain,omitempty"`

    // GCP fields (ADD THESE - based on egress pattern)
    GcpPrivateServiceConnectServiceAttachment string `human:"GCP Private Service Connect Service Attachment,omitempty" serialized:"gcp_private_service_connect_service_attachment,omitempty"`
    GcpPrivateServiceConnectConnectionId      string `human:"GCP Private Service Connect Connection ID,omitempty" serialized:"gcp_private_service_connect_connection_id,omitempty"`
    GcpPrivateServiceConnectEndpointIpAddress string `human:"GCP Private Service Connect Endpoint IP Address,omitempty" serialized:"gcp_private_service_connect_endpoint_ip_address,omitempty"`
}
```

**Note:** Field names should match what the GCP ingress SDK types provide.

---

## Step 3: Update Create Command
**File:** `internal/network/command_access_point_private_link_ingress_endpoint_create.go`

### 3.1 Update Command Flags (around line 29-39)
```go
pcmd.AddCloudFlag(cmd)  // Change from AddCloudAwsFlag to support all clouds
cmd.Flags().String("vpc-endpoint-id", "", "ID of an AWS VPC endpoint.")
cmd.Flags().String("psc-connection-id", "", "ID of a GCP Private Service Connect connection.") // ADD THIS

cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
cobra.CheckErr(cmd.MarkFlagRequired("gateway"))
// Make vpc-endpoint-id conditional, not always required
```

### 3.2 Update Examples (around line 22-27)
Add GCP example:
```go
examples.Example{
    Text: "Create a GCP Private Service Connect ingress endpoint.",
    Code: "confluent network access-point private-link ingress-endpoint create --cloud gcp --gateway gw-123456 --psc-connection-id <connection-id>",
},
```

### 3.3 Update createIngressEndpoint function (around line 82-92)
Add GCP case to switch statement:
```go
switch cloud {
case pcloud.Aws:
    createIngressEndpoint.Spec.Config = &networkingaccesspointv1.NetworkingV1AccessPointSpecConfigOneOf{
        NetworkingV1AwsIngressPrivateLinkEndpoint: &networkingaccesspointv1.NetworkingV1AwsIngressPrivateLinkEndpoint{
            Kind:          "AwsIngressPrivateLinkEndpoint",
            VpcEndpointId: vpcEndpointId,
        },
    }
case pcloud.Gcp:  // ADD THIS CASE
    pscConnectionId, err := cmd.Flags().GetString("psc-connection-id")
    if err != nil {
        return err
    }
    createIngressEndpoint.Spec.Config = &networkingaccesspointv1.NetworkingV1AccessPointSpecConfigOneOf{
        NetworkingV1GcpIngressPrivateServiceConnectEndpoint: &networkingaccesspointv1.NetworkingV1GcpIngressPrivateServiceConnectEndpoint{
            Kind:                                    "GcpIngressPrivateServiceConnectEndpoint",
            PrivateServiceConnectConnectionId:       pscConnectionId,  // Verify field name in SDK
        },
    }
default:
    return fmt.Errorf("ingress endpoints are only supported for AWS and GCP")
}
```

### 3.4 Add Flag Validation
Add logic to require the right flag for each cloud:
```go
// After getting cloud flag
if cloud == pcloud.Aws {
    cobra.CheckErr(cmd.MarkFlagRequired("vpc-endpoint-id"))
} else if cloud == pcloud.Gcp {
    cobra.CheckErr(cmd.MarkFlagRequired("psc-connection-id"))
}
```

---

## Step 4: Update List Command
**File:** `internal/network/command_access_point_private_link_ingress_endpoint_list.go`

### 4.1 Update Filter Logic (around line 571-573)
Change from AWS-only filter to include GCP:
```go
for _, ingressEndpoint := range ingressEndpoints {
    if ingressEndpoint.Spec == nil {
        return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
    }

    // Filter to include both AWS and GCP ingress endpoints
    config := ingressEndpoint.Spec.GetConfig()
    if config.NetworkingV1AwsIngressPrivateLinkEndpoint == nil &&
       config.NetworkingV1GcpIngressPrivateServiceConnectEndpoint == nil {
        continue  // Skip if neither AWS nor GCP ingress endpoint
    }

    if ingressEndpoint.Status == nil {
        return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
    }

    // ... rest of the code
}
```

### 4.2 Add GCP Status Extraction (around line 586-592)
```go
// Existing AWS status extraction
if ingressEndpoint.Status.Config != nil && ingressEndpoint.Status.Config.NetworkingV1AwsIngressPrivateLinkEndpointStatus != nil {
    out.AwsVpcEndpointId = ingressEndpoint.Status.Config.NetworkingV1AwsIngressPrivateLinkEndpointStatus.GetVpcEndpointId()
    out.AwsVpcEndpointServiceName = ingressEndpoint.Status.Config.NetworkingV1AwsIngressPrivateLinkEndpointStatus.GetVpcEndpointServiceName()
    if ingressEndpoint.Status.Config.NetworkingV1AwsIngressPrivateLinkEndpointStatus.HasDnsDomain() {
        out.DnsDomain = ingressEndpoint.Status.Config.NetworkingV1AwsIngressPrivateLinkEndpointStatus.GetDnsDomain()
    }
}

// ADD GCP status extraction
if ingressEndpoint.Status.Config != nil && ingressEndpoint.Status.Config.NetworkingV1GcpIngressPrivateServiceConnectEndpointStatus != nil {
    out.GcpPrivateServiceConnectServiceAttachment = ingressEndpoint.Status.Config.NetworkingV1GcpIngressPrivateServiceConnectEndpointStatus.GetPrivateServiceConnectServiceAttachment()
    out.GcpPrivateServiceConnectConnectionId = ingressEndpoint.Status.Config.NetworkingV1GcpIngressPrivateServiceConnectEndpointStatus.GetPrivateServiceConnectConnectionId()
    out.GcpPrivateServiceConnectEndpointIpAddress = ingressEndpoint.Status.Config.NetworkingV1GcpIngressPrivateServiceConnectEndpointStatus.GetPrivateServiceConnectEndpointIpAddress()
}
```

**Note:** Verify actual method names in SDK (GetPrivateServiceConnectServiceAttachment, etc.)

---

## Step 5: Update Print Function
**File:** `internal/network/command_access_point_private_link_ingress_endpoint.go`

### 5.1 Update printPrivateLinkIngressEndpointTable (around line 282-293)
Add GCP status extraction (same as Step 4.2 above):
```go
// After AWS status extraction, add:
if ingressEndpoint.Status.Config != nil && ingressEndpoint.Status.Config.NetworkingV1GcpIngressPrivateServiceConnectEndpointStatus != nil {
    out.GcpPrivateServiceConnectServiceAttachment = ingressEndpoint.Status.Config.NetworkingV1GcpIngressPrivateServiceConnectEndpointStatus.GetPrivateServiceConnectServiceAttachment()
    out.GcpPrivateServiceConnectConnectionId = ingressEndpoint.Status.Config.NetworkingV1GcpIngressPrivateServiceConnectEndpointStatus.GetPrivateServiceConnectConnectionId()
    out.GcpPrivateServiceConnectEndpointIpAddress = ingressEndpoint.Status.Config.NetworkingV1GcpIngressPrivateServiceConnectEndpointStatus.GetPrivateServiceConnectEndpointIpAddress()
}
```

---

## Step 6: Update Autocomplete
**File:** `internal/network/command_access_point_private_link_ingress_endpoint.go`

### 6.1 Update autocompleteIngressEndpoints (around line 255-257)
Change filter to include GCP:
```go
ingressEndpoints := slices.DeleteFunc(accessPoints, func(accessPoint networkingaccesspointv1.NetworkingV1AccessPoint) bool {
    config := accessPoint.Spec.GetConfig()
    return config.NetworkingV1AwsIngressPrivateLinkEndpoint == nil &&
           config.NetworkingV1GcpIngressPrivateServiceConnectEndpoint == nil
})
```

---

## Step 7: Gateway Support (If Needed)
**Files:**
- `internal/network/command_gateway.go`
- `internal/network/command_gateway_create.go`
- `internal/network/command_gateway_list.go`

### 7.1 Check if GCP Ingress Gateways Exist
Determine if GCP has ingress-specific gateway types or if existing GCP egress gateway handles both.

### 7.2 If GCP Ingress Gateways Are Separate:

**command_gateway.go:**
```go
const (
    awsEgressPrivateLink           = "AwsEgressPrivateLink"
    awsIngressPrivateLink          = "AwsIngressPrivateLink"
    gcpEgressPrivateServiceConnect = "GcpEgressPrivateServiceConnect"
    gcpIngressPrivateServiceConnect = "GcpIngressPrivateServiceConnect" // ADD THIS
    // ...
)

var (
    createGatewayTypes = []string{"egress-privatelink", "ingress-privatelink", "private-network-interface"}
    listGatewayTypes   = []string{
        "aws-egress-privatelink",
        "aws-ingress-privatelink",
        "azure-egress-privatelink",
        "gcp-egress-private-service-connect",
        "gcp-ingress-private-service-connect", // ADD THIS
    }
    gatewayTypeMap = map[string]string{
        "aws-egress-privatelink":              awsEgressPrivateLink,
        "aws-ingress-privatelink":             awsIngressPrivateLink,
        "azure-egress-privatelink":            azureEgressPrivateLink,
        "gcp-egress-private-service-connect":  gcpEgressPrivateServiceConnect,
        "gcp-ingress-private-service-connect": gcpIngressPrivateServiceConnect, // ADD THIS
    }
)
```

Add GCP ingress fields to `gatewayOut` struct and update:
- `getGatewayCloud()` to recognize GCP ingress gateways
- `getGatewayType()` to identify GCP ingress type
- `printGatewayTable()` to extract GCP ingress gateway details

**command_gateway_create.go:**
Add GCP ingress case to gateway creation (similar to AWS ingress pattern)

**command_gateway_list.go:**
Add GCP ingress handling in list output

---

## Step 8: Update Tests
**Files:**
- `test/network_test.go`
- `test/test-server/networking_handlers.go`
- Test fixtures in `test/fixtures/output/network/access-point/private-link/ingress-endpoint/`

### 8.1 Add Test Cases (test/network_test.go)
Around line 1170-1175, add GCP test cases:
```go
{args: "network access-point private-link ingress-endpoint create --cloud gcp --gateway gw-12345 --psc-connection-id psc-abc123", fixture: "network/access-point/private-link/ingress-endpoint/create-gcp.golden"},
{args: "network access-point private-link ingress-endpoint create my-gcp-ingress --cloud gcp --gateway gw-12345 --psc-connection-id psc-abc123", fixture: "network/access-point/private-link/ingress-endpoint/create-gcp-name.golden"},
{args: "network access-point private-link ingress-endpoint describe ap-77777", fixture: "network/access-point/private-link/ingress-endpoint/describe-gcp.golden"},
{args: "network access-point private-link ingress-endpoint describe ap-77777 --output json", fixture: "network/access-point/private-link/ingress-endpoint/describe-gcp-json.golden"},
```

### 8.2 Add Mock Server Handlers (test/test-server/networking_handlers.go)
Create GCP ingress endpoint mock responses similar to AWS:
- Mock access point with GCP ingress config
- Mock status responses with GCP-specific fields

### 8.3 Create Golden Files
Create expected output fixtures:
- `create-gcp.golden`
- `create-gcp-name.golden`
- `describe-gcp.golden`
- `describe-gcp-json.golden`
- Update `list.golden` and `list-json.golden` to include GCP entries

### 8.4 Update Gateway Tests (if applicable)
Add GCP ingress gateway tests if implementing gateway support

---

## Step 9: Update Documentation/Examples

### 9.1 Update Help Text
Ensure command help shows both AWS and GCP are supported

### 9.2 Update Error Messages
Change "only supported for AWS" to "supported for AWS and GCP"

---

## Step 10: Validation & Testing

### 10.1 Pre-commit Checklist
- [ ] All tests pass: `go test ./...`
- [ ] Build succeeds: `go build`
- [ ] Linting passes
- [ ] Manual testing with test server

### 10.2 Integration Testing
- [ ] Test GCP ingress endpoint create
- [ ] Test GCP ingress endpoint list
- [ ] Test GCP ingress endpoint describe
- [ ] Test GCP ingress endpoint update
- [ ] Test GCP ingress endpoint delete
- [ ] Test mixed AWS/GCP ingress endpoint list
- [ ] Test autocomplete with GCP endpoints
- [ ] Test GCP ingress gateway (if applicable)

### 10.3 Edge Cases
- [ ] Test with invalid GCP connection IDs
- [ ] Test with missing required flags
- [ ] Test output formats (JSON, YAML)
- [ ] Test error handling for API failures

---

## Open Questions to Resolve

1. **SDK Field Names:** What are the exact field names in the GCP ingress SDK types?
   - PrivateServiceConnectConnectionId?
   - PrivateServiceConnectServiceAttachment?
   - Check SDK documentation

2. **Gateway Support:** Does GCP have separate ingress gateway types or does the existing GCP egress gateway handle both directions?

3. **Required Fields:** What fields are required for GCP ingress endpoint creation?
   - Just connection ID?
   - Any additional configuration?

4. **Status Fields:** What status information does GCP ingress return?
   - Connection ID, IP address, service attachment?
   - DNS/domain information?

5. **Feature Flag:** Should GCP ingress be behind the same feature flag as AWS ingress, or a separate one?

---

## Files Summary

### Files to Modify:
1. `go.mod` (if SDK update needed)
2. `internal/network/command_access_point_private_link_ingress_endpoint.go`
3. `internal/network/command_access_point_private_link_ingress_endpoint_create.go`
4. `internal/network/command_access_point_private_link_ingress_endpoint_list.go`
5. `internal/network/command_gateway.go` (if GCP ingress gateways exist)
6. `internal/network/command_gateway_create.go` (if GCP ingress gateways exist)
7. `internal/network/command_gateway_list.go` (if GCP ingress gateways exist)
8. `test/network_test.go`
9. `test/test-server/networking_handlers.go`

### Files to Create:
1. Test fixture golden files for GCP ingress endpoints

---

## Estimated Effort

- **Step 1 (SDK Check):** 30 minutes
- **Steps 2-6 (Core Implementation):** 2-3 hours
- **Step 7 (Gateway Support):** 1-2 hours (if needed)
- **Steps 8-9 (Tests & Docs):** 2-3 hours
- **Step 10 (Validation):** 1-2 hours

**Total:** 6-11 hours depending on gateway requirements

---

## References

- AWS Ingress Commit: `31b97e26297183ae99b24f71b6b554d2295a7bc7`
- AWS Implementation pattern in `command_access_point_private_link_ingress_endpoint_*.go`
- GCP Egress Implementation pattern in `command_access_point_private_link_egress_endpoint_*.go`
