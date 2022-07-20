package asyncapi

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/confluentinc/cli/internal/pkg/log"
)

type Catalog struct {
	GetSchemaLevelTagsRequest func(srEndpoint, schemaClusterId, schemaId, apiKey, apiSecret string) ([]byte, error)
	GetTagDefinitionsRequest  func(srEndpoint, tagName, apiKey, apiSecret string) ([]byte, error)
}

func (c Catalog) GetSchemaLevelTags(srEndpoint, schemaClusterId, schemaId, apiKey, apiSecret string) ([]byte, error) {
	if c.GetSchemaLevelTagsRequest == nil {
		c.GetSchemaLevelTagsRequest = func(srEndpoint, schemaClusterId, schemaId, apiKey, apiSecret string) ([]byte, error) {
			dataCatalogUrl := fmt.Sprintf("%s/catalog/v1/entity/type/sr_schema/name/%s:.:%s/tags", srEndpoint, schemaClusterId, schemaId)
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
			return ioutil.ReadAll(resp.Body)
		}
	}
	return c.GetSchemaLevelTagsRequest(srEndpoint, schemaClusterId, schemaId, apiKey, apiSecret)
}

func (c Catalog) GetTagDefinitions(srEndpoint, tagName, apiKey, apiSecret string) ([]byte, error) {
	if c.GetTagDefinitionsRequest == nil {
		c.GetTagDefinitionsRequest = func(srEndpoint, tagName, apiKey, apiSecret string) ([]byte, error) {
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
			return ioutil.ReadAll(resp.Body)
		}
	}
	return c.GetTagDefinitionsRequest(srEndpoint, tagName, apiKey, apiSecret)
}
