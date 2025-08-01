package testserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"sort"
	"time"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	apikeysv2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"
	byokv1 "github.com/confluentinc/ccloud-sdk-go-v2/byok/v1"
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
)

type ListMeta interface {
	SetNext(string)
}

type ResourceList[T any] interface {
	SetData(T)
}

type ErrorJson struct {
	Message string `json:"message"`
}

type ApiKeyListV2 []apikeysv2.IamV2ApiKey

// Len is part of sort.Interface.
func (d ApiKeyListV2) Len() int {
	return len(d)
}

// Swap is part of sort.Interface.
func (d ApiKeyListV2) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

// Less is part of sort.Interface. We use Key as the value to sort by
func (d ApiKeyListV2) Less(i, j int) bool {
	return *d[i].Id < *d[j].Id
}

func fillKeyStoreV2() {
	keyStoreV2["MYKEY1"] = &apikeysv2.IamV2ApiKey{
		Id: apikeysv2.PtrString("MYKEY1"),
		Spec: &apikeysv2.IamV2ApiKeySpec{
			Resource: &apikeysv2.ObjectReference{
				Id:         "lkc-bob",
				ApiVersion: apikeysv2.PtrString("cmk/v2"),
				Kind:       apikeysv2.PtrString("Cluster"),
			},
			Owner:       &apikeysv2.ObjectReference{Id: "u11"},
			Description: apikeysv2.PtrString("Example description"),
		},
	}

	keyStoreV2["MYKEY2"] = &apikeysv2.IamV2ApiKey{
		Id: apikeysv2.PtrString("MYKEY2"),
		Spec: &apikeysv2.IamV2ApiKeySpec{
			Resource: &apikeysv2.ObjectReference{
				Id:         "lkc-abc",
				ApiVersion: apikeysv2.PtrString("cmk/v2"),
				Kind:       apikeysv2.PtrString("Cluster"),
			},
			Owner:       &apikeysv2.ObjectReference{Id: "u-17"},
			Description: apikeysv2.PtrString(""),
		},
	}

	keyStoreV2["MULTICLUSTERKEY1"] = &apikeysv2.IamV2ApiKey{
		Id: apikeysv2.PtrString("MULTICLUSTERKEY1"),
		Spec: &apikeysv2.IamV2ApiKeySpec{
			Resource: &apikeysv2.ObjectReference{
				Id:         "lkc-abc",
				ApiVersion: apikeysv2.PtrString("cmk/v2"),
				Kind:       apikeysv2.PtrString("Cluster"),
			},
			Resources: &[]apikeysv2.ObjectReference{
				{
					Id:         "lkc-abc",
					ApiVersion: apikeysv2.PtrString("cmk/v2"),
					Kind:       apikeysv2.PtrString("Cluster"),
				},
				{
					Id:   "lsrc-1234",
					Kind: apikeysv2.PtrString("SchemaRegistry"),
				},
			},
			Owner:       &apikeysv2.ObjectReference{Id: "u-44ddd"},
			Description: apikeysv2.PtrString("works for two clusters"),
		},
	}

	keyStoreV2["MULTICLUSTERKEY2"] = &apikeysv2.IamV2ApiKey{
		Id: apikeysv2.PtrString("MULTICLUSTERKEY2"),
		Spec: &apikeysv2.IamV2ApiKeySpec{
			Resource: &apikeysv2.ObjectReference{
				Id:         "lkc-abc",
				ApiVersion: apikeysv2.PtrString("cmk/v2"),
				Kind:       apikeysv2.PtrString("Cluster"),
			},
			Resources: &[]apikeysv2.ObjectReference{
				{
					Id:         "lkc-abc",
					ApiVersion: apikeysv2.PtrString("cmk/v2"),
					Kind:       apikeysv2.PtrString("Cluster"),
				},
				{
					Id:   "lsrc-abc123",
					Kind: apikeysv2.PtrString("SchemaRegistry"),
				},
			},
			Owner:       &apikeysv2.ObjectReference{Id: "u-44ddd"},
			Description: apikeysv2.PtrString("works for two clusters but on a different sr cluster"),
		},
	}

	keyStoreV2["MULTICLUSTERKEY3"] = &apikeysv2.IamV2ApiKey{
		Id: apikeysv2.PtrString("MULTICLUSTERKEY3"),
		Spec: &apikeysv2.IamV2ApiKeySpec{
			Resource: &apikeysv2.ObjectReference{
				Id:         "lkc-abc",
				ApiVersion: apikeysv2.PtrString("cmk/v2"),
				Kind:       apikeysv2.PtrString("Cluster"),
			},
			Resources: &[]apikeysv2.ObjectReference{
				{
					Id:         "lkc-abc",
					ApiVersion: apikeysv2.PtrString("cmk/v2"),
					Kind:       apikeysv2.PtrString("Cluster"),
				},
				{
					Id:   "lsrc-1234",
					Kind: apikeysv2.PtrString("SchemaRegistry"),
				},
			},
			Owner:       &apikeysv2.ObjectReference{Id: "sa-12345"},
			Description: apikeysv2.PtrString("works for two clusters and owned by service account"),
		},
	}

	keyStoreV2["UIAPIKEY100"] = &apikeysv2.IamV2ApiKey{
		Id: apikeysv2.PtrString("UIAPIKEY100"),
		Spec: &apikeysv2.IamV2ApiKeySpec{
			Resource: &apikeysv2.ObjectReference{
				Id:         "lkc-cool1",
				ApiVersion: apikeysv2.PtrString("cmk/v2"),
				Kind:       apikeysv2.PtrString("Cluster"),
			},
			Owner:       &apikeysv2.ObjectReference{Id: "u-22bbb"},
			Description: apikeysv2.PtrString(""),
		},
	}
	keyStoreV2["UIAPIKEY101"] = &apikeysv2.IamV2ApiKey{
		Id: apikeysv2.PtrString("UIAPIKEY101"),
		Spec: &apikeysv2.IamV2ApiKeySpec{
			Resource: &apikeysv2.ObjectReference{
				Id:         "lkc-other1",
				ApiVersion: apikeysv2.PtrString("cmk/v2"),
				Kind:       apikeysv2.PtrString("Cluster"),
			},
			Owner:       &apikeysv2.ObjectReference{Id: "u-22bbb"},
			Description: apikeysv2.PtrString(""),
		},
	}
	keyStoreV2["UIAPIKEY102"] = &apikeysv2.IamV2ApiKey{
		Id: apikeysv2.PtrString("UIAPIKEY102"),
		Spec: &apikeysv2.IamV2ApiKeySpec{
			Resource:    &apikeysv2.ObjectReference{Id: "lksqlc-ksql1", Kind: apikeysv2.PtrString("ksqlDB")},
			Owner:       &apikeysv2.ObjectReference{Id: "u-22bbb"},
			Description: apikeysv2.PtrString(""),
		},
	}
	keyStoreV2["UIAPIKEY103"] = &apikeysv2.IamV2ApiKey{
		Id: apikeysv2.PtrString("UIAPIKEY103"),
		Spec: &apikeysv2.IamV2ApiKeySpec{
			Resource: &apikeysv2.ObjectReference{
				Id:         "lkc-cool1",
				ApiVersion: apikeysv2.PtrString("cmk/v2"),
				Kind:       apikeysv2.PtrString("Cluster"),
			},
			Owner:       &apikeysv2.ObjectReference{Id: "u-22bbb"},
			Description: apikeysv2.PtrString(""),
		},
	}
	keyStoreV2["SERVICEACCOUNTKEY1"] = &apikeysv2.IamV2ApiKey{
		Id: apikeysv2.PtrString("SERVICEACCOUNTKEY1"),
		Spec: &apikeysv2.IamV2ApiKeySpec{
			Resource: &apikeysv2.ObjectReference{
				Id:         "lkc-bob",
				ApiVersion: apikeysv2.PtrString("cmk/v2"),
				Kind:       apikeysv2.PtrString("Cluster"),
			},
			Owner:       &apikeysv2.ObjectReference{Id: serviceAccountResourceId},
			Description: apikeysv2.PtrString(""),
		},
	}
	keyStoreV2["DEACTIVATEDUSERKEY"] = &apikeysv2.IamV2ApiKey{
		Id: apikeysv2.PtrString("DEACTIVATEDUSERKEY"),
		Spec: &apikeysv2.IamV2ApiKeySpec{
			Resource: &apikeysv2.ObjectReference{
				Id:         "lkc-bob",
				ApiVersion: apikeysv2.PtrString("cmk/v2"),
				Kind:       apikeysv2.PtrString("Cluster"),
			},
			Owner:       &apikeysv2.ObjectReference{Id: deactivatedUserResourceId},
			Description: apikeysv2.PtrString(""),
		},
	}
	for _, k := range keyStoreV2 {
		k.Metadata = &apikeysv2.ObjectMeta{CreatedAt: keyTime}
	}
}

func apiKeysFilterV2(url *url.URL) *apikeysv2.IamV2ApiKeyList {
	var apiKeys []apikeysv2.IamV2ApiKey
	q := url.Query()
	uid := q.Get("spec.owner")
	resourceId := q.Get("spec.resource")

	for _, key := range keyStoreV2 {
		uidFilter := uid == "" || uid == key.Spec.Owner.GetId()
		clusterFilter := resourceId == "" || containsResourceId(key, resourceId)
		if uidFilter && clusterFilter {
			apiKeys = append(apiKeys, *key)
		}
	}
	sort.Sort(ApiKeyListV2(apiKeys))
	return &apikeysv2.IamV2ApiKeyList{Data: apiKeys}
}

func containsResourceId(key *apikeysv2.IamV2ApiKey, resourceId string) bool {
	if len(key.Spec.GetResources()) == 0 {
		return key.Spec.Resource.Id == resourceId
	}

	return slices.ContainsFunc(key.Spec.GetResources(), func(o apikeysv2.ObjectReference) bool {
		return o.GetId() == resourceId
	})
}

func fillByokStoreV1() map[string]*byokv1.ByokV1Key {
	byokStoreV1 := map[string]*byokv1.ByokV1Key{}

	byokStoreV1["cck-001"] = &byokv1.ByokV1Key{
		Id:       byokv1.PtrString("cck-001"),
		Metadata: &byokv1.ObjectMeta{CreatedAt: byokv1.PtrTime(time.Date(2022, time.November, 12, 8, 24, 0, 0, time.UTC))},
		Key: &byokv1.ByokV1KeyKeyOneOf{
			ByokV1AwsKey: &byokv1.ByokV1AwsKey{
				KeyArn: "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012",
				Kind:   "AwsKey",
				Roles: &[]string{
					"arn:aws:iam::123456789012:role/role1",
					"arn:aws:iam::123456789012:role/role2",
				},
			},
		},
		Provider: byokv1.PtrString("AWS"),
		State:    byokv1.PtrString("IN_USE"),
	}

	byokStoreV1["cck-002"] = &byokv1.ByokV1Key{
		Id:       byokv1.PtrString("cck-002"),
		Metadata: &byokv1.ObjectMeta{CreatedAt: byokv1.PtrTime(time.Date(2022, time.November, 7, 5, 30, 0, 0, time.UTC))},
		Key: &byokv1.ByokV1KeyKeyOneOf{
			ByokV1AwsKey: &byokv1.ByokV1AwsKey{
				KeyArn: "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012",
				Kind:   "AwsKey",
				Roles: &[]string{
					"arn:aws:iam::123456789012:role/role1",
					"arn:aws:iam::123456789012:role/role2",
				},
			},
		},
		Provider: byokv1.PtrString("AWS"),
		State:    byokv1.PtrString("AVAILABLE"),
	}

	byokStoreV1["cck-003"] = &byokv1.ByokV1Key{
		Id:       byokv1.PtrString("cck-003"),
		Metadata: &byokv1.ObjectMeta{CreatedAt: byokv1.PtrTime(time.Date(2023, time.January, 1, 12, 0, 30, 0, time.UTC))},
		Key: &byokv1.ByokV1KeyKeyOneOf{
			ByokV1AzureKey: &byokv1.ByokV1AzureKey{
				ApplicationId: byokv1.PtrString("00000000-0000-0000-0000-000000000000"),
				KeyId:         "https://a-vault.vault.azure.net/keys/a-key",
				KeyVaultId:    "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/a-resourcegroups/providers/Microsoft.KeyVault/vaults/a-vault",
				Kind:          "AzureKey",
				TenantId:      "00000000-0000-0000-0000-000000000000",
			},
		},
		Provider: byokv1.PtrString("Azure"),
		State:    byokv1.PtrString("AVAILABLE"),
	}

	byokStoreV1["cck-004"] = &byokv1.ByokV1Key{
		Id:       byokv1.PtrString("cck-004"),
		Metadata: &byokv1.ObjectMeta{CreatedAt: byokv1.PtrTime(time.Date(2023, time.January, 1, 13, 0, 30, 0, time.UTC))},
		Key: &byokv1.ByokV1KeyKeyOneOf{
			ByokV1GcpKey: &byokv1.ByokV1GcpKey{
				KeyId: "projects/exampleproject/locations/us-central1/keyRings/testkeyring/cryptoKeys/testbyokkey/cryptoKeyVersions/3",
				Kind:  "GcpKey",
			},
		},
		Provider: byokv1.PtrString("GCP"),
		State:    byokv1.PtrString("AVAILABLE"),
	}

	return byokStoreV1
}

func writeErrorJson(w http.ResponseWriter, message string) error {
	errorJson, err := json.Marshal(ErrorJson{Message: message})
	if err != nil {
		return err
	}
	_, err = io.Writer.Write(w, errorJson)
	return err
}

func writeServiceAccountInvalidError(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusForbidden)
	return writeErrorJson(w, "service account is not valid")
}

func writeResourceNotFoundError(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusForbidden)
	return writeErrorJson(w, "resource not found")
}

func writeInvalidRoleNameError(w http.ResponseWriter, roleName string) error {
	w.WriteHeader(http.StatusBadRequest)
	return writeErrorJson(w, fmt.Sprintf("Invalid role name : %s", roleName))
}

func writeUserConflictError(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusConflict)
	return writeErrorJson(w, "This user already exists within the Organization")
}

func getCmkBasicDescribeCluster(id, name string) *cmkv2.CmkV2Cluster {
	return &cmkv2.CmkV2Cluster{
		Spec: &cmkv2.CmkV2ClusterSpec{
			DisplayName: cmkv2.PtrString(name),
			Cloud:       cmkv2.PtrString("aws"),
			Region:      cmkv2.PtrString("us-west-2"),
			Config: &cmkv2.CmkV2ClusterSpecConfigOneOf{
				CmkV2Basic: &cmkv2.CmkV2Basic{Kind: "Basic"},
			},
			KafkaBootstrapEndpoint: cmkv2.PtrString("SASL_SSL://kafka-endpoint"),
			HttpEndpoint:           cmkv2.PtrString(TestKafkaRestProxyUrl.String()),
			Availability:           cmkv2.PtrString("SINGLE_ZONE"),
			Endpoints: &cmkv2.ModelMap{
				"pni-abc123-basic": cmkv2.CmkV2Endpoints{
					KafkaBootstrapEndpoint: "SASL_SSL://pni-abc123-basic.kafka.us-west-2.aws.confluent.cloud:9092",
					HttpEndpoint:           "https://pni-abc123-basic.rest.us-west-2.aws.confluent.cloud",
					ConnectionType:         "PNI",
				},
				"privatelink-uvw456-basic": cmkv2.CmkV2Endpoints{
					KafkaBootstrapEndpoint: "SASL_SSL://pl-uvw456-basic.kafka.us-west-2.aws.confluent.cloud:9092",
					HttpEndpoint:           "https://pl-uvw456-basic.rest.us-west-2.aws.confluent.cloud",
					ConnectionType:         "PRIVATELINK",
				},
				"privatelink-xyz789-basic": cmkv2.CmkV2Endpoints{
					KafkaBootstrapEndpoint: "SASL_SSL://pl-xyz789-basic.kafka.us-west-2.aws.confluent.cloud:9092",
					HttpEndpoint:           "https://pl-xyz789-basic.rest.us-west-2.aws.confluent.cloud",
					ConnectionType:         "PRIVATELINK",
				},
				"public-0001-basic": cmkv2.CmkV2Endpoints{
					KafkaBootstrapEndpoint: "SASL_SSL://public-0001-basic.kafka.us-west-2.aws.confluent.cloud:9092",
					HttpEndpoint:           "https://public-0001-basic.rest.us-west-2.aws.confluent.cloud",
					ConnectionType:         "PUBLIC",
				},
			},
		},
		Id: cmkv2.PtrString(id),
		Status: &cmkv2.CmkV2ClusterStatus{
			Phase: "PROVISIONED",
		},
	}
}

func getCmkDedicatedDescribeCluster(id, name string, cku int32) *cmkv2.CmkV2Cluster {
	return &cmkv2.CmkV2Cluster{
		Spec: &cmkv2.CmkV2ClusterSpec{
			DisplayName: cmkv2.PtrString(name),
			Cloud:       cmkv2.PtrString("aws"),
			Region:      cmkv2.PtrString("us-west-2"),
			Config: &cmkv2.CmkV2ClusterSpecConfigOneOf{
				CmkV2Dedicated: &cmkv2.CmkV2Dedicated{Kind: "Dedicated", Cku: cku},
			},
			KafkaBootstrapEndpoint: cmkv2.PtrString("SASL_SSL://kafka-endpoint"),
			HttpEndpoint:           cmkv2.PtrString(TestKafkaRestProxyUrl.String()),
			Availability:           cmkv2.PtrString("SINGLE_ZONE"),
			Endpoints: &cmkv2.ModelMap{
				"pni-abc123-dedicated": cmkv2.CmkV2Endpoints{
					KafkaBootstrapEndpoint: "SASL_SSL://pni-abc123-dedicated.kafka.us-west-2.aws.confluent.cloud:9092",
					HttpEndpoint:           "https://pni-abc123-dedicated.rest.us-west-2.aws.confluent.cloud",
					ConnectionType:         "PNI",
				},
				"privatelink-uvw456-dedicated": cmkv2.CmkV2Endpoints{
					KafkaBootstrapEndpoint: "SASL_SSL://pl-uvw456-dedicated.kafka.us-west-2.aws.confluent.cloud:9092",
					HttpEndpoint:           "https://pl-uvw456-dedicated.rest.us-west-2.aws.confluent.cloud",
					ConnectionType:         "PRIVATELINK",
				},
				"privatelink-xyz789-dedicated": cmkv2.CmkV2Endpoints{
					KafkaBootstrapEndpoint: "SASL_SSL://pl-xyz789-dedicated.kafka.us-west-2.aws.confluent.cloud:9092",
					HttpEndpoint:           "https://pl-xyz789-dedicated.rest.us-west-2.aws.confluent.cloud",
					ConnectionType:         "PRIVATELINK",
				},
				"public-0001-dedicated": cmkv2.CmkV2Endpoints{
					KafkaBootstrapEndpoint: "SASL_SSL://public-0001-dedicated.kafka.us-west-2.aws.confluent.cloud:9092",
					HttpEndpoint:           "https://public-0001-dedicated.rest.us-west-2.aws.confluent.cloud",
					ConnectionType:         "PUBLIC",
				},
			},
		},
		Id: cmkv2.PtrString(id),
		Status: &cmkv2.CmkV2ClusterStatus{
			Phase: "PROVISIONED",
			Cku:   cmkv2.PtrInt32(cku),
		},
	}
}

func getCmkUnknownDescribeCluster(id, name string) *cmkv2.CmkV2Cluster {
	return &cmkv2.CmkV2Cluster{
		Spec: &cmkv2.CmkV2ClusterSpec{
			DisplayName:            cmkv2.PtrString(name),
			Cloud:                  cmkv2.PtrString("aws"),
			Region:                 cmkv2.PtrString("us-west-2"),
			KafkaBootstrapEndpoint: cmkv2.PtrString("SASL_SSL://kafka-endpoint"),
			HttpEndpoint:           cmkv2.PtrString(TestKafkaRestProxyUrl.String()),
			Availability:           cmkv2.PtrString("SINGLE_ZONE"),
			Endpoints: &cmkv2.ModelMap{
				"pni-abc123-unknown": cmkv2.CmkV2Endpoints{
					KafkaBootstrapEndpoint: "SASL_SSL://pni-abc123-unknown.kafka.us-west-2.aws.confluent.cloud:9092",
					HttpEndpoint:           "https://pni-abc123-unknown.rest.us-west-2.aws.confluent.cloud",
					ConnectionType:         "PNI",
				},
				"privatelink-uvw456-unknown": cmkv2.CmkV2Endpoints{
					KafkaBootstrapEndpoint: "SASL_SSL://pl-uvw456-unknown.kafka.us-west-2.aws.confluent.cloud:9092",
					HttpEndpoint:           "https://pl-uvw456-unknown.rest.us-west-2.aws.confluent.cloud",
					ConnectionType:         "PRIVATELINK",
				},
				"privatelink-xyz789-unknown": cmkv2.CmkV2Endpoints{
					KafkaBootstrapEndpoint: "SASL_SSL://pl-xyz789-unknown.kafka.us-west-2.aws.confluent.cloud:9092",
					HttpEndpoint:           "https://pl-xyz789-unknown.rest.us-west-2.aws.confluent.cloud",
					ConnectionType:         "PRIVATELINK",
				},
				"public-0001-unknown": cmkv2.CmkV2Endpoints{
					KafkaBootstrapEndpoint: "SASL_SSL://public-0001.kafka.us-west-2.aws.confluent.cloud:9092",
					HttpEndpoint:           "https://public-0001-unknown.rest.us-west-2.aws.confluent.cloud",
					ConnectionType:         "PUBLIC",
				},
			},
		},
		Id:     cmkv2.PtrString(id),
		Status: &cmkv2.CmkV2ClusterStatus{Phase: "PROVISIONED"},
	}
}

func buildUser(id int32, email, firstName, lastName, resourceId string) *ccloudv1.User {
	return &ccloudv1.User{
		Id:         id,
		Email:      email,
		FirstName:  firstName,
		LastName:   lastName,
		ResourceId: resourceId,
	}
}

// Set a page token for an empty second page
func setPageToken[T any](resourceList ResourceList[T], meta ListMeta, url *url.URL) {
	if url.Query().Get("page_token") == "" { // Initial response: set a page token
		nextPage := fmt.Sprintf("%s&%s", url.String(), "page_token=UvmDWOB1iwfAIBPj6EYb")
		meta.SetNext(nextPage)
	} else { // Second response: set no page token, and return an empty data list
		var zero T
		resourceList.SetData(zero)
	}
}
