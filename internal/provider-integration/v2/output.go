// Copyright 2024 Confluent Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v2

import "fmt"

// providerIntegrationOut represents the output struct for provider integrations
// Used for both list and detailed views with omitempty fields
type providerIntegrationOut struct {
	Id          string          `human:"ID" serialized:"id"`
	DisplayName string          `human:"Name" serialized:"display_name"`
	Provider    string          `human:"Provider" serialized:"provider"`
	Environment string          `human:"Environment" serialized:"environment"`
	Status      string          `human:"Status" serialized:"status"`
	Usages      []string        `human:"Usages" serialized:"usages,omitempty"`
	AzureConfig *azureConfigOut `human:"Azure Configuration" serialized:"azure_config,omitempty"`
	GcpConfig   *gcpConfigOut   `human:"GCP Configuration" serialized:"gcp_config,omitempty"`
}

type azureConfigOut struct {
	CustomerTenantId          string `human:"Customer Azure Tenant ID" serialized:"customer_azure_tenant_id"`
	ConfluentMultiTenantAppId string `human:"Confluent Multi-Tenant App ID" serialized:"confluent_multi_tenant_app_id"`
}

func (a *azureConfigOut) String() string {
	return fmt.Sprintf("Customer Azure Tenant ID: %s\nConfluent Multi-Tenant App ID: %s", a.CustomerTenantId, a.ConfluentMultiTenantAppId)
}

type gcpConfigOut struct {
	CustomerServiceAccount string `human:"Customer Google Service Account" serialized:"customer_google_service_account"`
	GoogleServiceAccount   string `human:"Google Service Account" serialized:"google_service_account"`
}

func (g *gcpConfigOut) String() string {
	return fmt.Sprintf("Customer Google Service Account: %s\nGoogle Service Account: %s", g.CustomerServiceAccount, g.GoogleServiceAccount)
}
