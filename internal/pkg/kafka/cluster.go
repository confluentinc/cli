package kafka

import (
	"context"
	"net/http"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
)

func CmkApiContext(authToken string) context.Context {
	auth := context.WithValue(context.Background(), cmkv2.ContextAccessToken, authToken)
	return auth
}

func ListKafkaClusters(client *cmkv2.APIClient, environment string, authToken string) (cmkv2.CmkV2ClusterList, *http.Response, error) {
	req := client.ClustersCmkV2Api.ListCmkV2Clusters(CmkApiContext(authToken)).Environment(environment)
	return client.ClustersCmkV2Api.ListCmkV2ClustersExecute(req)
}
