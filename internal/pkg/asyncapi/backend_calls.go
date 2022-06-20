package asyncapi

import (
	"io"
	"io/ioutil"
	"net/http"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/log"
)

func GetSchemaLevelTags(srEndpoint, schemaClusterId, schemaId, apiKey, apiSecret string) ([]byte, error) {
	dataCatalogUrl := srEndpoint + "/catalog/v1/entity/type/sr_schema/name/" + schemaClusterId + ":.:" + schemaId + "/tags"
	req, err := http.NewRequest("GET", dataCatalogUrl, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(apiKey, apiSecret)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.CliLogger.Warnf("error getting tags: %v", err)
		}
	}(resp.Body)
	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}

func GetTagDefinitions(srEndpoint, tagName, apiKey, apiSecret string) ([]byte, error) {
	tagDefsUrl := srEndpoint + "/catalog/v1/types/tagdefs/" + tagName
	req, err := http.NewRequest("GET", tagDefsUrl, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(apiKey, apiSecret)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.CliLogger.Warnf("error getting tags: %v", err)
		}
	}(resp.Body)
	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}

func GetClusterCleanupPolicy(clusterEndpoint, clusterId, topicName string, clusterCreds *v1.APIKeyPair) ([]byte, error) {
	cleanupPolicyUrl := clusterEndpoint + "/kafka/v3/clusters/" + clusterId + "/topics/" + topicName + "/configs/cleanup.policy"
	req, err := http.NewRequest("GET", cleanupPolicyUrl, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(clusterCreds.Key, clusterCreds.Secret)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	} else {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.CliLogger.Warnf("error in getting bindings: %v", err)
			}
		}(resp.Body)
		body, err := ioutil.ReadAll(resp.Body)
		return body, err
	}
}

func GetClusterDeleteRetentionMs(clusterEndpoint, clusterId, topicName string, clusterCreds *v1.APIKeyPair) ([]byte, error) {
	deleteRetentionMsUrl := clusterEndpoint + "/kafka/v3/clusters/" + clusterId + "/topics/" + topicName + "/configs/delete.retention.ms"
	req, err := http.NewRequest("GET", deleteRetentionMsUrl, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(clusterCreds.Key, clusterCreds.Secret)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	} else {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.CliLogger.Warn("Error in getting Bindings")
			}
		}(resp.Body)
		body, err := ioutil.ReadAll(resp.Body)
		return body, err
	}
}
