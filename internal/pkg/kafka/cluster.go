package kafka

import (
	"context"
	"net/http"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
)

func cmkApiContext(authToken string) context.Context {
	auth := context.WithValue(context.Background(), cmkv2.ContextAccessToken, authToken)
	return auth
}

func ListCmkKafkaClusters(client *ccloudv2.Client, environment string) (cmkv2.CmkV2ClusterList, *http.Response, error) {
	req := client.CmkClient.ClustersCmkV2Api.ListCmkV2Clusters(cmkApiContext(client.AuthToken)).Environment(environment)
	return client.CmkClient.ClustersCmkV2Api.ListCmkV2ClustersExecute(req)
}
