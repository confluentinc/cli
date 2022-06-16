package testserver

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	productv1 "github.com/confluentinc/cc-structs/kafka/product/core/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	apikeysv2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	mdsv2 "github.com/confluentinc/ccloud-sdk-go-v2/mds/v2"
	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
)

var (
	resourceNotFoundErrMsg      = `{"error":{"code":403,"message":"resource not found","nested_errors":{},"details":[],"stack":null},"cluster":null}`
	serviceAccountInvalidErrMsg = `{"errors":[{"status":"403","detail":"service account is not valid"}]}`
)

type ApiKeyList []*schedv1.ApiKey

// Len is part of sort.Interface.
func (d ApiKeyList) Len() int {
	return len(d)
}

// Swap is part of sort.Interface.
func (d ApiKeyList) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

// Less is part of sort.Interface. We use Key as the value to sort by
func (d ApiKeyList) Less(i, j int) bool {
	return d[i].Key < d[j].Key
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

func fillKeyStore() {
	keyStore[keyIndex] = &schedv1.ApiKey{
		Id:     keyIndex,
		Key:    "MYKEY1",
		Secret: "MYSECRET1",
		LogicalClusters: []*schedv1.ApiKey_Cluster{
			{Id: "lkc-bob", Type: "kafka"},
		},
		UserId:         1,
		UserResourceId: "u11",
	}
	keyIndex += 1
	keyStore[keyIndex] = &schedv1.ApiKey{
		Id:     keyIndex,
		Key:    "MYKEY2",
		Secret: "MYSECRET2",
		LogicalClusters: []*schedv1.ApiKey_Cluster{
			{Id: "lkc-abc", Type: "kafka"},
		},
		UserId:         2,
		UserResourceId: "u-17",
	}
	keyIndex += 1
	keyStore[100] = &schedv1.ApiKey{
		Id:     keyIndex,
		Key:    "UIAPIKEY100",
		Secret: "UIAPISECRET100",
		LogicalClusters: []*schedv1.ApiKey_Cluster{
			{Id: "lkc-cool1", Type: "kafka"},
		},
		UserId:         4,
		UserResourceId: "u-22bbb",
	}
	keyStore[101] = &schedv1.ApiKey{
		Id:     keyIndex,
		Key:    "UIAPIKEY101",
		Secret: "UIAPISECRET101",
		LogicalClusters: []*schedv1.ApiKey_Cluster{
			{Id: "lkc-other1", Type: "kafka"},
		},
		UserId:         4,
		UserResourceId: "u-22bbb",
	}
	keyStore[102] = &schedv1.ApiKey{
		Id:     keyIndex,
		Key:    "UIAPIKEY102",
		Secret: "UIAPISECRET102",
		LogicalClusters: []*schedv1.ApiKey_Cluster{
			{Id: "lksqlc-ksql1", Type: "ksql"},
		},
		UserId:         4,
		UserResourceId: "u-22bbb",
	}
	keyStore[103] = &schedv1.ApiKey{
		Id:     keyIndex,
		Key:    "UIAPIKEY103",
		Secret: "UIAPISECRET103",
		LogicalClusters: []*schedv1.ApiKey_Cluster{
			{Id: "lkc-cool1", Type: "kafka"},
		},
		UserId:         4,
		UserResourceId: "u-22bbb",
	}
	keyStore[200] = &schedv1.ApiKey{
		Id:     keyIndex,
		Key:    "SERVICEACCOUNTKEY1",
		Secret: "SERVICEACCOUNTSECRET1",
		LogicalClusters: []*schedv1.ApiKey_Cluster{
			{Id: "lkc-bob", Type: "kafka"},
		},
		UserId:         serviceAccountID,
		UserResourceId: serviceAccountResourceID,
	}
	keyStore[201] = &schedv1.ApiKey{
		Id:     keyIndex,
		Key:    "DEACTIVATEDUSERKEY",
		Secret: "DEACTIVATEDUSERSECRET",
		LogicalClusters: []*schedv1.ApiKey_Cluster{
			{Id: "lkc-bob", Type: "kafka"},
		},
		UserId:         deactivatedUserID,
		UserResourceId: deactivatedResourceID,
	}
	for _, k := range keyStore {
		k.Created = keyTimestamp
	}
}

func fillKeyStoreV2() {
	keyStoreV2["MYKEY1"] = &apikeysv2.IamV2ApiKey{
		Id: apikeysv2.PtrString("MYKEY1"),
		Spec: &apikeysv2.IamV2ApiKeySpec{
			Resource:    &apikeysv2.ObjectReference{Id: "lkc-bob", Kind: apikeysv2.PtrString("Cluster")},
			Owner:       &apikeysv2.ObjectReference{Id: "u11"},
			Description: apikeysv2.PtrString(""),
		},
	}

	keyStoreV2["MYKEY2"] = &apikeysv2.IamV2ApiKey{
		Id: apikeysv2.PtrString("MYKEY2"),
		Spec: &apikeysv2.IamV2ApiKeySpec{
			Resource:    &apikeysv2.ObjectReference{Id: "lkc-abc", Kind: apikeysv2.PtrString("Cluster")},
			Owner:       &apikeysv2.ObjectReference{Id: "u-17"},
			Description: apikeysv2.PtrString(""),
		},
	}

	keyStoreV2["UIAPIKEY100"] = &apikeysv2.IamV2ApiKey{
		Id: apikeysv2.PtrString("UIAPIKEY100"),
		Spec: &apikeysv2.IamV2ApiKeySpec{
			Resource:    &apikeysv2.ObjectReference{Id: "lkc-cool1", Kind: apikeysv2.PtrString("Cluster")},
			Owner:       &apikeysv2.ObjectReference{Id: "u-22bbb"},
			Description: apikeysv2.PtrString(""),
		},
	}
	keyStoreV2["UIAPIKEY101"] = &apikeysv2.IamV2ApiKey{
		Id: apikeysv2.PtrString("UIAPIKEY101"),
		Spec: &apikeysv2.IamV2ApiKeySpec{
			Resource:    &apikeysv2.ObjectReference{Id: "lkc-other1", Kind: apikeysv2.PtrString("Cluster")},
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
			Resource:    &apikeysv2.ObjectReference{Id: "lkc-cool1", Kind: apikeysv2.PtrString("Cluster")},
			Owner:       &apikeysv2.ObjectReference{Id: "u-22bbb"},
			Description: apikeysv2.PtrString(""),
		},
	}
	keyStoreV2["SERVICEACCOUNTKEY1"] = &apikeysv2.IamV2ApiKey{
		Id: apikeysv2.PtrString("SERVICEACCOUNTKEY1"),
		Spec: &apikeysv2.IamV2ApiKeySpec{
			Resource:    &apikeysv2.ObjectReference{Id: "lkc-bob", Kind: apikeysv2.PtrString("Cluster")},
			Owner:       &apikeysv2.ObjectReference{Id: serviceAccountResourceID},
			Description: apikeysv2.PtrString(""),
		},
	}
	keyStoreV2["DEACTIVATEDUSERKEY"] = &apikeysv2.IamV2ApiKey{
		Id: apikeysv2.PtrString("DEACTIVATEDUSERKEY"),
		Spec: &apikeysv2.IamV2ApiKeySpec{
			Resource:    &apikeysv2.ObjectReference{Id: "lkc-bob", Kind: apikeysv2.PtrString("Cluster")},
			Owner:       &apikeysv2.ObjectReference{Id: deactivatedResourceID},
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
		uidFilter := (uid == "") || (uid == key.Spec.Owner.Id)
		clusterFilter := (resourceId == "") || (resourceId == key.Spec.Resource.Id)
		if uidFilter && clusterFilter {
			apiKeys = append(apiKeys, *key)
		}
	}
	sort.Sort(ApiKeyListV2(apiKeys))
	return &apikeysv2.IamV2ApiKeyList{Data: apiKeys}
}

func getV2ApiKey(apiKey *schedv1.ApiKey) *apikeysv2.IamV2ApiKey {
	return &apikeysv2.IamV2ApiKey{
		Id: apikeysv2.PtrString(apiKey.Key),
		Spec: &apikeysv2.IamV2ApiKeySpec{
			Owner:       &apikeysv2.ObjectReference{Id: apiKey.UserResourceId},
			Secret:      apikeysv2.PtrString(fmt.Sprintf("MYSECRET%d", keyIndex)),
			Resource:    &apikeysv2.ObjectReference{Id: apiKey.LogicalClusters[0].Id, Kind: apikeysv2.PtrString(resourceTypeToKind[apiKey.LogicalClusters[0].Type])},
			Description: apikeysv2.PtrString(apiKey.Description),
		},
		Metadata: &apikeysv2.ObjectMeta{CreatedAt: keyTime},
	}
}

func writeServiceAccountInvalidError(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusForbidden)
	_, err := io.WriteString(w, serviceAccountInvalidErrMsg)
	return err
}

func writeResourceNotFoundError(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusForbidden)
	_, err := io.WriteString(w, resourceNotFoundErrMsg)
	return err
}

func getBaseDescribeCluster(id, name string) *schedv1.KafkaCluster {
	return &schedv1.KafkaCluster{
		Id:              id,
		Name:            name,
		Deployment:      &schedv1.Deployment{Sku: productv1.Sku_BASIC},
		NetworkIngress:  100,
		NetworkEgress:   100,
		Storage:         500,
		ServiceProvider: "aws",
		Region:          "us-west-2",
		Endpoint:        "SASL_SSL://kafka-endpoint",
		ApiEndpoint:     "http://kafka-api-url",
		RestEndpoint:    "http://kafka-rest-url",
	}
}

func getCmkBasicDescribeCluster(id string, name string) *cmkv2.CmkV2Cluster {
	return &cmkv2.CmkV2Cluster{
		Spec: &cmkv2.CmkV2ClusterSpec{
			DisplayName: cmkv2.PtrString(name),
			Cloud:       cmkv2.PtrString("aws"),
			Region:      cmkv2.PtrString("us-west-2"),
			Config: &cmkv2.CmkV2ClusterSpecConfigOneOf{
				CmkV2Basic: &cmkv2.CmkV2Basic{Kind: "Basic"},
			},
			KafkaBootstrapEndpoint: cmkv2.PtrString("SASL_SSL://kafka-endpoint"),
			HttpEndpoint:           cmkv2.PtrString("http://kafka-rest-url"),
			Availability:           cmkv2.PtrString("SINGLE_ZONE"),
		},
		Id: cmkv2.PtrString(id),
		Status: &cmkv2.CmkV2ClusterStatus{
			Phase: "PROVISIONED",
		},
	}
}

func getCmkDedicatedDescribeCluster(id string, name string, cku int32) *cmkv2.CmkV2Cluster {
	return &cmkv2.CmkV2Cluster{
		Spec: &cmkv2.CmkV2ClusterSpec{
			DisplayName: cmkv2.PtrString(name),
			Cloud:       cmkv2.PtrString("aws"),
			Region:      cmkv2.PtrString("us-west-2"),
			Config: &cmkv2.CmkV2ClusterSpecConfigOneOf{
				CmkV2Dedicated: &cmkv2.CmkV2Dedicated{Kind: "Dedicated", Cku: cku},
			},
			KafkaBootstrapEndpoint: cmkv2.PtrString("SASL_SSL://kafka-endpoint"),
			HttpEndpoint:           cmkv2.PtrString("http://kafka-rest-url"),
			Availability:           cmkv2.PtrString("SINGLE_ZONE"),
		},
		Id: cmkv2.PtrString(id),
		Status: &cmkv2.CmkV2ClusterStatus{
			Phase: "PROVISIONED",
			Cku:   cmkv2.PtrInt32(cku),
		},
	}
}

func buildUser(id int32, email, firstName, lastName, resourceId string) *orgv1.User {
	return &orgv1.User{
		Id:             id,
		Email:          email,
		FirstName:      firstName,
		LastName:       lastName,
		OrganizationId: 0,
		Deactivated:    false,
		Verified:       nil,
		ResourceId:     resourceId,
	}
}

func buildIamUser(email, name, resourceId string) iamv2.IamV2User {
	return iamv2.IamV2User{
		Email:    iamv2.PtrString(email),
		FullName: iamv2.PtrString(name),
		Id:       iamv2.PtrString(resourceId),
	}
}

func buildInvitation(id, email, resourceId, status string) *orgv1.Invitation {
	return &orgv1.Invitation{
		Id:             id,
		Email:          email,
		UserResourceId: resourceId,
		Status:         status,
	}
}

func buildRoleBinding(user, roleName, crn string) mdsv2.IamV2RoleBinding {
	return mdsv2.IamV2RoleBinding{
		Id:         mdsv2.PtrString("0"),
		Principal:  mdsv2.PtrString("User:" + user),
		RoleName:   mdsv2.PtrString(roleName),
		CrnPattern: mdsv2.PtrString(crn),
	}
}

func isRoleBindingMatch(rolebinding mdsv2.IamV2RoleBinding, principal, roleName, crnPattern string) bool {
	if !strings.Contains(*rolebinding.CrnPattern, strings.TrimSuffix(crnPattern, "/*")) {
		return false
	}
	if principal != "" && principal != *rolebinding.Principal {
		return false
	}
	if roleName != "" && roleName != *rolebinding.RoleName {
		return false
	}
	return true
}

func isValidEnvironmentId(environments []*orgv1.Account, reqEnvId string) *orgv1.Account {
	for _, env := range environments {
		if reqEnvId == env.Id {
			return env
		}
	}
	return nil
}

func isValidOrgEnvironmentId(environments []*orgv2.OrgV2Environment, reqEnvId string) *orgv2.OrgV2Environment {
	for _, env := range environments {
		if reqEnvId == *env.Id {
			return env
		}
	}
	return nil
}
