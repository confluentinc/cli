package ccloudv2

import (
	flinkv2 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink/v2"
	apikeysv2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"
	byokv1 "github.com/confluentinc/ccloud-sdk-go-v2/byok/v1"
	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2/cdx/v1"
	cliv1 "github.com/confluentinc/ccloud-sdk-go-v2/cli/v1"
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	connectv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1"
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	identityproviderv2 "github.com/confluentinc/ccloud-sdk-go-v2/identity-provider/v2"
	kafkaquotas "github.com/confluentinc/ccloud-sdk-go-v2/kafka-quotas/v1"
	ksql "github.com/confluentinc/ccloud-sdk-go-v2/ksql/v2"
	mdsv2 "github.com/confluentinc/ccloud-sdk-go-v2/mds/v2"
	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
	servicequotav1 "github.com/confluentinc/ccloud-sdk-go-v2/service-quota/v1"
	srcm "github.com/confluentinc/ccloud-sdk-go-v2/srcm/v2"
	streamdesignerv1 "github.com/confluentinc/ccloud-sdk-go-v2/stream-designer/v1"

	testserver "github.com/confluentinc/cli/test/test-server"
)

// Client represents a Confluent Cloud Client as defined by ccloud-sdk-go-v2
type Client struct {
	AuthToken string

	ApiKeysClient          *apikeysv2.APIClient
	ByokClient             *byokv1.APIClient
	CdxClient              *cdxv1.APIClient
	CliClient              *cliv1.APIClient
	CmkClient              *cmkv2.APIClient
	ConnectClient          *connectv1.APIClient
	FlinkClient            *flinkv2.APIClient
	IamClient              *iamv2.APIClient
	IdentityProviderClient *identityproviderv2.APIClient
	KsqlClient             *ksql.APIClient
	KafkaQuotasClient      *kafkaquotas.APIClient
	MdsClient              *mdsv2.APIClient
	OrgClient              *orgv2.APIClient
	SchemaRegistryClient   *srcm.APIClient
	StreamDesignerClient   *streamdesignerv1.APIClient
	ServiceQuotaClient     *servicequotav1.APIClient
}

func NewClient(baseUrl string, isTest bool, authToken, userAgent string, unsafeTrace bool) *Client {
	url := getServerUrl(baseUrl)
	if isTest {
		url = testserver.TestV2CloudUrl.String()
	}

	return &Client{
		AuthToken: authToken,

		ApiKeysClient:          newApiKeysClient(url, userAgent, unsafeTrace),
		ByokClient:             newByokV1Client(url, userAgent, unsafeTrace),
		CdxClient:              newCdxClient(url, userAgent, unsafeTrace),
		CliClient:              newCliClient(url, userAgent, unsafeTrace),
		CmkClient:              newCmkClient(url, userAgent, unsafeTrace),
		ConnectClient:          newConnectClient(url, userAgent, unsafeTrace),
		FlinkClient:            newFlinkClient(url, userAgent, unsafeTrace),
		IamClient:              newIamClient(url, userAgent, unsafeTrace),
		IdentityProviderClient: newIdentityProviderClient(url, userAgent, unsafeTrace),
		KsqlClient:             newKsqlClient(url, userAgent, unsafeTrace),
		KafkaQuotasClient:      newKafkaQuotasClient(url, userAgent, unsafeTrace),
		MdsClient:              newMdsClient(url, userAgent, unsafeTrace),
		OrgClient:              newOrgClient(url, userAgent, unsafeTrace),
		SchemaRegistryClient:   newSchemaRegistryClient(url, userAgent, unsafeTrace),
		StreamDesignerClient:   newStreamDesignerClient(url, userAgent, unsafeTrace),
		ServiceQuotaClient:     newServiceQuotaClient(url, userAgent, unsafeTrace),
	}
}
