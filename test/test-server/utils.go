package test_server

import (
	"io"
	"net/http"
	"net/url"
	"strconv"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	productv1 "github.com/confluentinc/cc-structs/kafka/product/core/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
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

func fillKeyStore() {
	keyStore[keyIndex] = &schedv1.ApiKey{
		Id:     keyIndex,
		Key:    "MYKEY1",
		Secret: "MYSECRET1",
		LogicalClusters: []*schedv1.ApiKey_Cluster{
			{Id: "lkc-bob", Type: "kafka"},
		},
		UserId: 12,
	}
	keyIndex += 1
	keyStore[keyIndex] = &schedv1.ApiKey{
		Id:     keyIndex,
		Key:    "MYKEY2",
		Secret: "MYSECRET2",
		LogicalClusters: []*schedv1.ApiKey_Cluster{
			{Id: "lkc-abc", Type: "kafka"},
		},
		UserId: 18,
	}
	keyIndex += 1
	keyStore[100] = &schedv1.ApiKey{
		Id:     keyIndex,
		Key:    "UIAPIKEY100",
		Secret: "UIAPISECRET100",
		LogicalClusters: []*schedv1.ApiKey_Cluster{
			{Id: "lkc-cool1", Type: "kafka"},
		},
		UserId: 25,
	}
	keyStore[101] = &schedv1.ApiKey{
		Id:     keyIndex,
		Key:    "UIAPIKEY101",
		Secret: "UIAPISECRET101",
		LogicalClusters: []*schedv1.ApiKey_Cluster{
			{Id: "lkc-other1", Type: "kafka"},
		},
		UserId: 25,
	}
	keyStore[102] = &schedv1.ApiKey{
		Id:     keyIndex,
		Key:    "UIAPIKEY102",
		Secret: "UIAPISECRET102",
		LogicalClusters: []*schedv1.ApiKey_Cluster{
			{Id: "lksqlc-ksql1", Type: "ksql"},
		},
		UserId: 25,
	}
	keyStore[103] = &schedv1.ApiKey{
		Id:     keyIndex,
		Key:    "UIAPIKEY103",
		Secret: "UIAPISECRET103",
		LogicalClusters: []*schedv1.ApiKey_Cluster{
			{Id: "lkc-cool1", Type: "kafka"},
		},
		UserId: 25,
	}
	keyStore[200] = &schedv1.ApiKey{
		Id:     keyIndex,
		Key:    "SERVICEACCOUNTKEY1",
		Secret: "SERVICEACCOUNTSECRET1",
		LogicalClusters: []*schedv1.ApiKey_Cluster{
			{Id: "lkc-bob", Type: "kafka"},
		},
		UserId: serviceAccountID,
	}
	keyStore[201] = &schedv1.ApiKey{
		Id:     keyIndex,
		Key:    "DEACTIVATEDUSERKEY",
		Secret: "DEACTIVATEDUSERSECRET",
		LogicalClusters: []*schedv1.ApiKey_Cluster{
			{Id: "lkc-bob", Type: "kafka"},
		},
		UserId: deactivatedUserID,
	}
	for _, k := range keyStore {
		k.Created = keyTimestamp
	}
}

func apiKeysFilter(url *url.URL) []*schedv1.ApiKey {
	var apiKeys []*schedv1.ApiKey
	q := url.Query()
	uid := q.Get("user_id")
	clusterIds := q["cluster_id"]

	for _, a := range keyStore {
		uidFilter := (uid == "0") || (uid == strconv.Itoa(int(a.UserId)))
		clusterFilter := (len(clusterIds) == 0) || func(clusterIds []string) bool {
			for _, c := range a.LogicalClusters {
				for _, clusterId := range clusterIds {
					if c.Id == clusterId {
						return true
					}
				}
			}
			return false
		}(clusterIds)

		if uidFilter && clusterFilter {
			apiKeys = append(apiKeys, a)
		}
	}
	return apiKeys
}

var (
	resourceNotFoundErrMsg = `{"error":{"code":404,"message":"resource not found","nested_errors":{},"details":[],"stack":null},"cluster":null}`
)

func writeResourceNotFoundError(w http.ResponseWriter) error {
	_, err := io.WriteString(w, resourceNotFoundErrMsg)
	return err
}

func getBaseDescribeCluster(id string, name string) *schedv1.KafkaCluster {
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
	}
}

func buildUser(id int32, email string, firstName string, lastName string, resourceId string) *orgv1.User {
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

func isValidEnvironmentId(environments []*orgv1.Account, reqEnvId string) (bool, *orgv1.Account) {
	for _, env := range environments {
		if reqEnvId == env.Id {
			return true, env
		}
	}
	return false, nil
}
