package ccloudv2

import (
	"context"
	"net/http"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
)

func newKafkaRestClient(url, userAgent string, unsafeTrace bool) *kafkarestv3.APIClient {
	cfg := kafkarestv3.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = newRetryableHttpClient(unsafeTrace)
	cfg.Servers = kafkarestv3.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return kafkarestv3.NewAPIClient(cfg)
}

func (c *Client) GetKafkaRestUrl() string {
	return c.KafkaRestClient.GetConfig().Servers[0].URL
}

func (c *Client) kafkaRestApiContext() context.Context {
	return context.WithValue(context.Background(), kafkarestv3.ContextAccessToken, c.AuthToken)
}

func (c *Client) CreateKafkaAcls(clusterId string, data kafkarestv3.CreateAclRequestData) (*http.Response, error) {
	req := c.KafkaRestClient.ACLV3Api.CreateKafkaAcls(c.kafkaRestApiContext(), clusterId).CreateAclRequestData(data)
	return c.KafkaRestClient.ACLV3Api.CreateKafkaAclsExecute(req)
}

func (c *Client) GetKafkaAcls(clusterId string, acl *schedv1.ACLBinding) (kafkarestv3.AclDataList, *http.Response, error) {
	req := c.KafkaRestClient.ACLV3Api.GetKafkaAcls(c.kafkaRestApiContext(), clusterId).Host(acl.Entry.Host).Principal(acl.Entry.Principal).ResourceName(acl.Pattern.Name)

	if acl.Entry.Operation != schedv1.ACLOperations_UNKNOWN {
		req.Operation(acl.Entry.Operation.String())
	}

	if acl.Pattern.PatternType != schedv1.PatternTypes_UNKNOWN {
		req.PatternType(acl.Pattern.PatternType.String())
	}

	if acl.Entry.PermissionType != schedv1.ACLPermissionTypes_UNKNOWN {
		req.Permission(acl.Entry.PermissionType.String())
	}

	if acl.Pattern.ResourceType != schedv1.ResourceTypes_UNKNOWN {
		req.ResourceType(kafkarestv3.AclResourceType(acl.Pattern.ResourceType.String()))
	}

	return c.KafkaRestClient.ACLV3Api.GetKafkaAclsExecute(req)
}

func (c *Client) DeleteKafkaAcls(clusterId string, acl *schedv1.ACLFilter) (kafkarestv3.InlineResponse200, *http.Response, error) {
	req := c.KafkaRestClient.ACLV3Api.DeleteKafkaAcls(c.kafkaRestApiContext(), clusterId).Host(acl.EntryFilter.Host).Principal(acl.EntryFilter.Principal).ResourceName(acl.PatternFilter.Name)

	if acl.EntryFilter.Operation != schedv1.ACLOperations_UNKNOWN {
		req.Operation(acl.EntryFilter.Operation.String())
	}

	if acl.PatternFilter.PatternType != schedv1.PatternTypes_UNKNOWN {
		req.PatternType(acl.PatternFilter.PatternType.String())
	}

	if acl.EntryFilter.PermissionType != schedv1.ACLPermissionTypes_UNKNOWN {
		req.Permission(acl.EntryFilter.PermissionType.String())
	}

	if acl.PatternFilter.ResourceType != schedv1.ResourceTypes_UNKNOWN {
		req.ResourceType(kafkarestv3.AclResourceType(acl.PatternFilter.ResourceType.String()))
	}

	return c.KafkaRestClient.ACLV3Api.DeleteKafkaAclsExecute(req)
}
