package asyncapi

import (
	"io"
	"io/ioutil"
	"net/http"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/log"
)

func GetSchemaLevelTags(srEndpoint, schemaClusterId, schemaId, apiKey, apiSecret string) []byte {
	dataCatalogUrl := srEndpoint + "/catalog/v1/entity/type/sr_schema/name/" + schemaClusterId + ":.:" + schemaId + "/tags"
	req, _ := http.NewRequest("GET", dataCatalogUrl, nil)
	req.SetBasicAuth(apiKey, apiSecret)
	resp, _ := http.DefaultClient.Do(req)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.CliLogger.Warnf("error getting tags: %v", err)
		}
	}(resp.Body)
	body, _ := ioutil.ReadAll(resp.Body)
	return body
}

func GetTagDefinitions(srEndpoint, tagName, apiKey, apiSecret string) []byte {
	tagDefsUrl := srEndpoint + "/catalog/v1/types/tagdefs/" + tagName
	req, _ := http.NewRequest("GET", tagDefsUrl, nil)
	req.SetBasicAuth(apiKey, apiSecret)
	resp, _ := http.DefaultClient.Do(req)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.CliLogger.Warnf("error getting tags: %v", err)
		}
	}(resp.Body)
	body, _ := ioutil.ReadAll(resp.Body)
	return body
}

func GetClusterCleanupPolicy(clusterEndpoint, clusterId, topicName string, clusterCreds *v1.APIKeyPair) []byte {
	cleanupPolicyUrl := clusterEndpoint + "/kafka/v3/clusters/" + clusterId + "/topics/" + topicName + "/configs/cleanup.policy"
	req, _ := http.NewRequest("GET", cleanupPolicyUrl, nil)
	req.SetBasicAuth(clusterCreds.Key, clusterCreds.Secret)
	resp, _ := http.DefaultClient.Do(req)
	if resp == nil {
		return nil
	} else {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.CliLogger.Warnf("error in getting bindings: %v", err)
			}
		}(resp.Body)
		body, _ := ioutil.ReadAll(resp.Body)
		return body
	}
}

func GetClusterDeleteRetentionMs(clusterEndpoint, clusterId, topicName string, clusterCreds *v1.APIKeyPair) []byte {
	deleteRetentionMsUrl := clusterEndpoint + "/kafka/v3/clusters/" + clusterId + "/topics/" + topicName + "/configs/delete.retention.ms"
	req, _ := http.NewRequest("GET", deleteRetentionMsUrl, nil)
	req.SetBasicAuth(clusterCreds.Key, clusterCreds.Secret)
	resp, _ := http.DefaultClient.Do(req)
	if resp == nil {
		return nil
	} else {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.CliLogger.Warn("Error in getting Bindings")
			}
		}(resp.Body)
		body, _ := ioutil.ReadAll(resp.Body)
		return body
	}
}
