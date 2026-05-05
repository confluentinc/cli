# GCP IngressPrivateServiceConnect & Azure IngressPrivateLink CLI Testing

Corresponding PR: #3283 Add Azure and GCP ingress access point and gateway support

## Prerequisites

```
dist/confluent_darwin_arm64_v8.0/confluent login --url stag.cpdev.cloud --organization 94b47aa0-ed3a-44da-9460-6cec0c710e84 --save
```

---

## Test Checklist

### Gateways

- [x] **1. Create Azure ingress gateway**
```
$ dist/confluent_darwin_arm64_v8.0/confluent network gateway create test-azure-ingress --cloud azure --region centralus --type ingress-privatelink
+-------------+-------------------------+
| ID          | gw-stgc6dj0d2           |
| Name        | test-azure-ingress      |
| Environment | env-stgcznd2qz          |
| Region      | centralus               |
| Type        | AzureIngressPrivateLink |
| Phase       | PROVISIONING            |
+-------------+-------------------------+
```
Expected: gateway created, `Type: AzureIngressPrivateLink`, phase `PROVISIONING` ✅

- [x] **2. Create GCP ingress gateway**
```
$ dist/confluent_darwin_arm64_v8.0/confluent network gateway create test-gcp-ingress --cloud gcp --region us-central1 --type ingress-private-service-connect
+-------------+---------------------------------+
| ID          | gw-stgco5rq5v                   |
| Name        | test-gcp-ingress                |
| Environment | env-stgcznd2qz                  |
| Region      | us-central1                     |
| Type        | GcpIngressPrivateServiceConnect |
| Phase       | PROVISIONING                    |
+-------------+---------------------------------+
```
Expected: gateway created, `Type: GcpIngressPrivateServiceConnect`, phase `PROVISIONING` ✅

- [x] **3. Describe Azure gateway**
```
$ dist/confluent_darwin_arm64_v8.0/confluent network gateway describe gw-stgc6dj0d2
+--------------------------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| ID                             | gw-stgc6dj0d2                                                                                                                                                      |
| Name                           | test-azure-ingress                                                                                                                                                 |
| Environment                    | env-stgcznd2qz                                                                                                                                                     |
| Region                         | centralus                                                                                                                                                          |
| Type                           | AzureIngressPrivateLink                                                                                                                                            |
| Azure Private Link Service     | plattg-stgc56xrj6-privatelink.54b70a2e-b341-41df-bb6c-fbb01bf7c481.centralus.azure.privatelinkservice                                                              |
| Alias                          |                                                                                                                                                                    |
| Azure Private Link Resource ID | /subscriptions/3c0d5cd7-137e-48bb-80ba-d2b7d177986f/resourceGroups/plattg-stgc56xrj6/providers/Microsoft.Network/privateLinkServices/plattg-stgc56xrj6-privatelink |
| Phase                          | CREATED                                                                                                                                                            |
+--------------------------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------+
```
Expected: `Azure Private Link Service Alias` and `Azure Private Link Resource ID` populated ✅

- [x] **4. Describe GCP gateway**
```
$ dist/confluent_darwin_arm64_v8.0/confluent network gateway describe gw-stgco5rq5v
+----------------------------+---------------------------------------------------------------------------------------------------+
| ID                         | gw-stgco5rq5v                                                                                     |
| Name                       | test-gcp-ingress                                                                                  |
| Environment                | env-stgcznd2qz                                                                                    |
| Region                     | us-central1                                                                                       |
| Type                       | GcpIngressPrivateServiceConnect                                                                   |
| GCP PSC Service Attachment | projects/traffic-stag/regions/us-central1/serviceAttachments/plattg-stgcwplnlg-service-attachment |
| Phase                      | CREATED                                                                                           |
+----------------------------+---------------------------------------------------------------------------------------------------+
```
Expected: `GCP PSC Service Attachment` populated ✅

- [x] **5. Filter by Azure ingress type**
```
$ dist/confluent_darwin_arm64_v8.0/confluent network gateway list --types azure-ingress-privatelink --output yaml
- id: gw-stgc1q5k9v
  name: test-azure-ingress-gateway-updated
  environment: env-stgcznd2qz
  region: centralus
  type: AzureIngressPrivateLink
  azure_private_link_service_alias: plattg-stgc56xrj6-privatelink.54b70a2e-b341-41df-bb6c-fbb01bf7c481.centralus.azure.privatelinkservice
  azure_private_link_service_resource_id: /subscriptions/3c0d5cd7-137e-48bb-80ba-d2b7d177986f/resourceGroups/plattg-stgc56xrj6/providers/Microsoft.Network/privateLinkServices/plattg-stgc56xrj6-privatelink
  phase: READY
- id: gw-stgc6dj0d2
  name: test-azure-ingress
  environment: env-stgcznd2qz
  region: centralus
  type: AzureIngressPrivateLink
  azure_private_link_service_alias: plattg-stgc56xrj6-privatelink.54b70a2e-b341-41df-bb6c-fbb01bf7c481.centralus.azure.privatelinkservice
  azure_private_link_service_resource_id: /subscriptions/3c0d5cd7-137e-48bb-80ba-d2b7d177986f/resourceGroups/plattg-stgc56xrj6/providers/Microsoft.Network/privateLinkServices/plattg-stgc56xrj6-privatelink
  phase: CREATED
```
Expected: only Azure ingress gateways returned ✅

- [x] **6. Filter by GCP ingress type**
```
$ dist/confluent_darwin_arm64_v8.0/confluent network gateway list --types gcp-ingress-private-service-connect --output yaml
- id: gw-stgc6ed3l2
  name: test-gcp-ingress-gateway-updated
  environment: env-stgcznd2qz
  region: us-central1
  type: GcpIngressPrivateServiceConnect
  gcp_private_service_connect_service_attachment: projects/traffic-stag/regions/us-central1/serviceAttachments/plattg-stgcwplnlg-service-attachment
  phase: READY
- id: gw-stgco5rq5v
  name: test-gcp-ingress
  environment: env-stgcznd2qz
  region: us-central1
  type: GcpIngressPrivateServiceConnect
  gcp_private_service_connect_service_attachment: projects/traffic-stag/regions/us-central1/serviceAttachments/plattg-stgcwplnlg-service-attachment
  phase: CREATED
```
Expected: only GCP ingress gateways returned ✅

- [x] **7. Filter by both types**
```
$ dist/confluent_darwin_arm64_v8.0/confluent network gateway list --types azure-ingress-privatelink,gcp-ingress-private-service-connect --output yaml
- id: gw-stgc1q5k9v
  name: test-azure-ingress-gateway-updated
  environment: env-stgcznd2qz
  region: centralus
  type: AzureIngressPrivateLink
  azure_private_link_service_alias: plattg-stgc56xrj6-privatelink.54b70a2e-b341-41df-bb6c-fbb01bf7c481.centralus.azure.privatelinkservice
  azure_private_link_service_resource_id: /subscriptions/3c0d5cd7-137e-48bb-80ba-d2b7d177986f/resourceGroups/plattg-stgc56xrj6/providers/Microsoft.Network/privateLinkServices/plattg-stgc56xrj6-privatelink
  phase: READY
- id: gw-stgc6dj0d2
  name: test-azure-ingress
  environment: env-stgcznd2qz
  region: centralus
  type: AzureIngressPrivateLink
  azure_private_link_service_alias: plattg-stgc56xrj6-privatelink.54b70a2e-b341-41df-bb6c-fbb01bf7c481.centralus.azure.privatelinkservice
  azure_private_link_service_resource_id: /subscriptions/3c0d5cd7-137e-48bb-80ba-d2b7d177986f/resourceGroups/plattg-stgc56xrj6/providers/Microsoft.Network/privateLinkServices/plattg-stgc56xrj6-privatelink
  phase: CREATED
- id: gw-stgc6ed3l2
  name: test-gcp-ingress-gateway-updated
  environment: env-stgcznd2qz
  region: us-central1
  type: GcpIngressPrivateServiceConnect
  gcp_private_service_connect_service_attachment: projects/traffic-stag/regions/us-central1/serviceAttachments/plattg-stgcwplnlg-service-attachment
  phase: READY
- id: gw-stgco5rq5v
  name: test-gcp-ingress
  environment: env-stgcznd2qz
  region: us-central1
  type: GcpIngressPrivateServiceConnect
  gcp_private_service_connect_service_attachment: projects/traffic-stag/regions/us-central1/serviceAttachments/plattg-stgcwplnlg-service-attachment
  phase: CREATED
```
Expected: both Azure and GCP ingress gateways returned ✅

---

### Access Points

- [x] **8. Create Azure ingress access point**
```
$ dist/confluent_darwin_arm64_v8.0/confluent network access-point private-link ingress-endpoint create \
  --cloud azure \
  --gateway gw-stgc6dj0d2 \
  --private-endpoint-resource-id /subscriptions/26812801-9a17-44c2-8398-a2e2ab4eb803/resourcegroups/richard-testing/providers/Microsoft.Network/privateEndpoints/testing-cli
+----------------------------------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| ID                                     | ap-stgcdnlrn4                                                                                                                                                      |
| Environment                            | env-stgcznd2qz                                                                                                                                                     |
| Gateway                                | gw-stgc6dj0d2                                                                                                                                                      |
| Phase                                  | PROVISIONING                                                                                                                                                       |
| Azure Private Link Service Alias       | plattg-stgc56xrj6-privatelink.54b70a2e-b341-41df-bb6c-fbb01bf7c481.centralus.azure.privatelinkservice                                                              |
| Azure Private Link Service Resource ID | /subscriptions/3c0d5cd7-137e-48bb-80ba-d2b7d177986f/resourceGroups/plattg-stgc56xrj6/providers/Microsoft.Network/privateLinkServices/plattg-stgc56xrj6-privatelink |
| Azure Private Endpoint Resource ID     | /subscriptions/26812801-9a17-44c2-8398-a2e2ab4eb803/resourcegroups/richard-testing/providers/Microsoft.Network/privateEndpoints/testing-cli                        |
| DNS Domain                             | apstgcdnlrn4.centralus.azure.accesspoint.stag.cpdev.cloud                                                                                                          |
+----------------------------------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------+

```
Expected: access point created, `Azure Private Endpoint Resource ID` and `DNS Domain` populated ✅

- [x] **9. Create GCP ingress access point**
```
$ dist/confluent_darwin_arm64_v8.0/confluent network access-point private-link ingress-endpoint create \
  --cloud gcp \
  --gateway gw-stgco5rq5v \
  --private-service-connect-connection-id 8345469524639756
+----------------------------+---------------------------------------------------------------------------------------------------+
| ID                         | ap-stgc4w0oo8                                                                                     |
| Environment                | env-stgcznd2qz                                                                                    |
| Gateway                    | gw-stgco5rq5v                                                                                     |
| Phase                      | PROVISIONING                                                                                      |
| GCP PSC Service Attachment | projects/traffic-stag/regions/us-central1/serviceAttachments/plattg-stgcwplnlg-service-attachment |
| GCP PSC Connection ID      | 8345469524639756                                                                                  |
| DNS Domain                 | apstgc4w0oo8.us-central1.gcp.accesspoint.stag.cpdev.cloud                                         |
+----------------------------+---------------------------------------------------------------------------------------------------+

```
Expected: access point created, `GCP PSC Connection ID` and `DNS Domain` populated ✅

- [x] **10. Verify cross-cloud flag rejection**
```
$ dist/confluent_darwin_arm64_v8.0/confluent network access-point private-link ingress-endpoint create \
  --cloud gcp \
  --gateway gw-stgco5rq5v \
  --private-endpoint-resource-id /subscriptions/26812801-9a17-44c2-8398-a2e2ab4eb803/resourcegroups/richard-testing/providers/Microsoft.Network/privateEndpoints/testing-cli
Error: "--vpc-endpoint-id" and "--private-endpoint-resource-id" are not valid for --cloud gcp; use "--private-service-connect-connection-id"
```
Expected: error `"--private-endpoint-resource-id" is not valid for --cloud gcp` ✅

- [x] **11. Update access point**
```
$ dist/confluent_darwin_arm64_v8.0/confluent network access-point private-link ingress-endpoint update ap-stgcdnlrn4 --name my-azure-ingress-ap
+----------------------------------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| ID                                     | ap-stgcdnlrn4                                                                                                                                                      |
| Name                                   | my-azure-ingress-ap                                                                                                                                                |
| Environment                            | env-stgcznd2qz                                                                                                                                                     |
| Gateway                                | gw-stgc6dj0d2                                                                                                                                                      |
| Phase                                  | READY                                                                                                                                                              |
| Azure Private Link Service Alias       | plattg-stgc56xrj6-privatelink.54b70a2e-b341-41df-bb6c-fbb01bf7c481.centralus.azure.privatelinkservice                                                              |
| Azure Private Link Service Resource ID | /subscriptions/3c0d5cd7-137e-48bb-80ba-d2b7d177986f/resourceGroups/plattg-stgc56xrj6/providers/Microsoft.Network/privateLinkServices/plattg-stgc56xrj6-privatelink |
| Azure Private Endpoint Resource ID     | /subscriptions/26812801-9a17-44c2-8398-a2e2ab4eb803/resourcegroups/richard-testing/providers/Microsoft.Network/privateEndpoints/testing-cli                        |
| DNS Domain                             | apstgcdnlrn4.centralus.azure.accesspoint.stag.cpdev.cloud                                                                                                          |
+----------------------------------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------+

$ dist/confluent_darwin_arm64_v8.0/confluent network access-point private-link ingress-endpoint update ap-stgc4w0oo8 --name my-gcp-ingress-ap
+----------------------------+---------------------------------------------------------------------------------------------------+
| ID                         | ap-stgc4w0oo8                                                                                     |
| Name                       | my-gcp-ingress-ap                                                                                 |
| Environment                | env-stgcznd2qz                                                                                    |
| Gateway                    | gw-stgco5rq5v                                                                                     |
| Phase                      | READY                                                                                             |
| GCP PSC Service Attachment | projects/traffic-stag/regions/us-central1/serviceAttachments/plattg-stgcwplnlg-service-attachment |
| GCP PSC Connection ID      | 8345469524639756                                                                                  |
| DNS Domain                 | apstgc4w0oo8.us-central1.gcp.accesspoint.stag.cpdev.cloud                                         |
+----------------------------+---------------------------------------------------------------------------------------------------+
```
Expected: name updated, other fields unchanged ✅

- [x] **12. Describe Azure access point**
```
$ dist/confluent_darwin_arm64_v8.0/confluent network access-point private-link ingress-endpoint describe ap-stgcdnlrn4
+----------------------------------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| ID                                     | ap-stgcdnlrn4                                                                                                                                                      |
| Name                                   | my-azure-ingress-ap                                                                                                                                                |
| Environment                            | env-stgcznd2qz                                                                                                                                                     |
| Gateway                                | gw-stgc6dj0d2                                                                                                                                                      |
| Phase                                  | READY                                                                                                                                                              |
| Azure Private Link Service Alias       | plattg-stgc56xrj6-privatelink.54b70a2e-b341-41df-bb6c-fbb01bf7c481.centralus.azure.privatelinkservice                                                              |
| Azure Private Link Service Resource ID | /subscriptions/3c0d5cd7-137e-48bb-80ba-d2b7d177986f/resourceGroups/plattg-stgc56xrj6/providers/Microsoft.Network/privateLinkServices/plattg-stgc56xrj6-privatelink |
| Azure Private Endpoint Resource ID     | /subscriptions/26812801-9a17-44c2-8398-a2e2ab4eb803/resourcegroups/richard-testing/providers/Microsoft.Network/privateEndpoints/testing-cli                        |
| DNS Domain                             | apstgcdnlrn4.centralus.azure.accesspoint.stag.cpdev.cloud                                                                                                          |
+----------------------------------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------+
```
Expected: `Azure Private Link Service Alias`, `Azure Private Endpoint Resource ID`, `DNS Domain` populated ✅

- [x] **13. Describe GCP access point**
```
$ dist/confluent_darwin_arm64_v8.0/confluent network access-point private-link ingress-endpoint describe ap-stgc4w0oo8
+----------------------------+---------------------------------------------------------------------------------------------------+
| ID                         | ap-stgc4w0oo8                                                                                     |
| Name                       | my-gcp-ingress-ap                                                                                 |
| Environment                | env-stgcznd2qz                                                                                    |
| Gateway                    | gw-stgco5rq5v                                                                                     |
| Phase                      | READY                                                                                             |
| GCP PSC Service Attachment | projects/traffic-stag/regions/us-central1/serviceAttachments/plattg-stgcwplnlg-service-attachment |
| GCP PSC Connection ID      | 8345469524639756                                                                                  |
| DNS Domain                 | apstgc4w0oo8.us-central1.gcp.accesspoint.stag.cpdev.cloud                                         |
+----------------------------+---------------------------------------------------------------------------------------------------+
```
Expected: `GCP PSC Service Attachment`, `GCP PSC Connection ID`, `DNS Domain` populated ✅

- [x] **14. List all access points**
```
$ dist/confluent_darwin_arm64_v8.0/confluent network access-point private-link ingress-endpoint list --output yaml
- id: ap-stgc4pwgn8
  name: test-azure-ingress-ap-updated
  environment: env-stgcznd2qz
  gateway: gw-stgc1q5k9v
  phase: READY
  azure_private_link_service_alias: plattg-stgc56xrj6-privatelink.54b70a2e-b341-41df-bb6c-fbb01bf7c481.centralus.azure.privatelinkservice
  azure_private_link_service_resource_id: /subscriptions/3c0d5cd7-137e-48bb-80ba-d2b7d177986f/resourceGroups/plattg-stgc56xrj6/providers/Microsoft.Network/privateLinkServices/plattg-stgc56xrj6-privatelink
  azure_private_endpoint_resource_id: /subscriptions/26812801-9a17-44c2-8398-a2e2ab4eb803/resourcegroups/richard-testing/providers/Microsoft.Network/privateEndpoints/testing
  dns_domain: apstgc4pwgn8.centralus.azure.accesspoint.stag.cpdev.cloud
- id: ap-stgc4w0oo8
  name: my-gcp-ingress-ap
  environment: env-stgcznd2qz
  gateway: gw-stgco5rq5v
  phase: READY
  gcp_private_service_connect_service_attachment: projects/traffic-stag/regions/us-central1/serviceAttachments/plattg-stgcwplnlg-service-attachment
  gcp_private_service_connect_connection_id: "8345469524639756"
  dns_domain: apstgc4w0oo8.us-central1.gcp.accesspoint.stag.cpdev.cloud
- id: ap-stgcdnlnn4
  name: test-gcp-ingress-ap-updated
  environment: env-stgcznd2qz
  gateway: gw-stgc6ed3l2
  phase: READY
  gcp_private_service_connect_service_attachment: projects/traffic-stag/regions/us-central1/serviceAttachments/plattg-stgcwplnlg-service-attachment
  gcp_private_service_connect_connection_id: "8345469524639755"
  dns_domain: apstgcdnlnn4.us-central1.gcp.accesspoint.stag.cpdev.cloud
- id: ap-stgcdnlrn4
  name: my-azure-ingress-ap
  environment: env-stgcznd2qz
  gateway: gw-stgc6dj0d2
  phase: READY
  azure_private_link_service_alias: plattg-stgc56xrj6-privatelink.54b70a2e-b341-41df-bb6c-fbb01bf7c481.centralus.azure.privatelinkservice
  azure_private_link_service_resource_id: /subscriptions/3c0d5cd7-137e-48bb-80ba-d2b7d177986f/resourceGroups/plattg-stgc56xrj6/providers/Microsoft.Network/privateLinkServices/plattg-stgc56xrj6-privatelink
  azure_private_endpoint_resource_id: /subscriptions/26812801-9a17-44c2-8398-a2e2ab4eb803/resourcegroups/richard-testing/providers/Microsoft.Network/privateEndpoints/testing-cli
  dns_domain: apstgcdnlrn4.centralus.azure.accesspoint.stag.cpdev.cloud
```
Expected: Azure and GCP access points listed with correct cloud-specific fields ✅

---

### Automated Tests

- [x] **15. Integration tests**
```
$ make integration-test INTEGRATION_TEST_ARGS="-run TestCLI/TestNetworkGateway"
ok  github.com/confluentinc/cli/v4/test  8.344s

$ make integration-test INTEGRATION_TEST_ARGS="-run TestCLI/TestNetworkAccessPointPrivateLinkIngressEndpoint"
ok  github.com/confluentinc/cli/v4/test  6.327s

$ make integration-test INTEGRATION_TEST_ARGS="-run TestCLI/TestHelp"
ok  github.com/confluentinc/cli/v4/test  135.274s
```
Expected: all pass ✅

---

## Results

| # | Test | Result | Notes |
|---|------|--------|-------|
| 1 | Create Azure gateway | ✅ Pass | |
| 2 | Create GCP gateway | ✅ Pass | |
| 3 | Describe Azure gateway | ✅ Pass | Phase: CREATED, Azure PL Service Alias and Resource ID populated |
| 4 | Describe GCP gateway | ✅ Pass | Phase: CREATED, GCP PSC Service Attachment populated |
| 5 | Filter Azure ingress type | ✅ Pass | |
| 6 | Filter GCP ingress type | ✅ Pass | Original map values restored after manual testing revealed API expects "GcpIngressPrivateLink" not "GcpIngressPrivateServiceConnect" |
| 7 | Filter both types | ✅ Pass | |
| 8 | Create Azure access point | ✅ Pass | ap-stgcdnlrn4, Phase: READY |
| 9 | Create GCP access point | ✅ Pass | ap-stgc4w0oo8, Phase: READY |
| 10 | Cross-cloud flag rejection | ✅ Pass | Fixed check order so cross-cloud message shows before required message |
| 11 | Update access point | ✅ Pass | |
| 12 | Describe Azure access point | ✅ Pass | |
| 13 | Describe GCP access point | ✅ Pass | |
| 14 | List access points | ✅ Pass | |
| 15 | Automated tests | ✅ Pass | |
