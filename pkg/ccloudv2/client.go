package ccloudv2

import (
	apikeysv2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"
	billingv1 "github.com/confluentinc/ccloud-sdk-go-v2/billing/v1"
	byokv1 "github.com/confluentinc/ccloud-sdk-go-v2/byok/v1"
	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2/cdx/v1"
	cliv1 "github.com/confluentinc/ccloud-sdk-go-v2/cli/v1"
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	connectcustompluginv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect-custom-plugin/v1"
	connectv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1"
	flinkv2 "github.com/confluentinc/ccloud-sdk-go-v2/flink/v2"
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	identityproviderv2 "github.com/confluentinc/ccloud-sdk-go-v2/identity-provider/v2"
	kafkaquotasv1 "github.com/confluentinc/ccloud-sdk-go-v2/kafka-quotas/v1"
	ksqlv2 "github.com/confluentinc/ccloud-sdk-go-v2/ksql/v2"
	mdsv2 "github.com/confluentinc/ccloud-sdk-go-v2/mds/v2"
	networkingdnsforwarderv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-dnsforwarder/v1"
	networkingipv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-ip/v1"
	networkingoutboundprivatelinkv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-outbound-privatelink/v1"
	networkingprivatelinkv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-privatelink/v1"
	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"
	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
	servicequotav1 "github.com/confluentinc/ccloud-sdk-go-v2/service-quota/v1"
	srcmv2 "github.com/confluentinc/ccloud-sdk-go-v2/srcm/v2"
	ssov2 "github.com/confluentinc/ccloud-sdk-go-v2/sso/v2"
	streamdesignerv1 "github.com/confluentinc/ccloud-sdk-go-v2/stream-designer/v1"

	"github.com/confluentinc/cli/v3/pkg/config"
	testserver "github.com/confluentinc/cli/v3/test/test-server"
)

// Client represents a Confluent Cloud Client as defined by ccloud-sdk-go-v2
type Client struct {
	cfg *config.Config

	ApiKeysClient                       *apikeysv2.APIClient
	BillingClient                       *billingv1.APIClient
	ByokClient                          *byokv1.APIClient
	CdxClient                           *cdxv1.APIClient
	CliClient                           *cliv1.APIClient
	CmkClient                           *cmkv2.APIClient
	ConnectClient                       *connectv1.APIClient
	ConnectCustomPluginClient           *connectcustompluginv1.APIClient
	FlinkClient                         *flinkv2.APIClient
	IamClient                           *iamv2.APIClient
	IdentityProviderClient              *identityproviderv2.APIClient
	KafkaQuotasClient                   *kafkaquotasv1.APIClient
	KsqlClient                          *ksqlv2.APIClient
	MdsClient                           *mdsv2.APIClient
	NetworkingClient                    *networkingv1.APIClient
	NetworkingDnsForwarderClient        *networkingdnsforwarderv1.APIClient
	NetworkingIpClient                  *networkingipv1.APIClient
	NetworkingOutboundPrivateLinkClient *networkingoutboundprivatelinkv1.APIClient
	NetworkingPrivateLinkClient         *networkingprivatelinkv1.APIClient
	OrgClient                           *orgv2.APIClient
	ServiceQuotaClient                  *servicequotav1.APIClient
	SrcmClient                          *srcmv2.APIClient
	SsoClient                           *ssov2.APIClient
	StreamDesignerClient                *streamdesignerv1.APIClient
}

func NewClient(cfg *config.Config, unsafeTrace bool) *Client {
	httpClient := NewRetryableHttpClient(cfg, unsafeTrace)

	url := getServerUrl(cfg.Context().GetPlatformServer())
	if cfg.IsTest {
		url = testserver.TestV2CloudUrl.String()
	}

	userAgent := cfg.Version.UserAgent

	return &Client{
		cfg: cfg,

		ApiKeysClient:                       newApiKeysClient(httpClient, url, userAgent, unsafeTrace),
		BillingClient:                       newBillingClient(httpClient, url, userAgent, unsafeTrace),
		ByokClient:                          newByokV1Client(httpClient, url, userAgent, unsafeTrace),
		CdxClient:                           newCdxClient(httpClient, url, userAgent, unsafeTrace),
		CliClient:                           newCliClient(url, userAgent, unsafeTrace),
		CmkClient:                           newCmkClient(httpClient, url, userAgent, unsafeTrace),
		ConnectClient:                       newConnectClient(httpClient, url, userAgent, unsafeTrace),
		ConnectCustomPluginClient:           newConnectCustomPluginClient(httpClient, url, userAgent, unsafeTrace),
		FlinkClient:                         newFlinkClient(httpClient, url, userAgent, unsafeTrace),
		IamClient:                           newIamClient(httpClient, url, userAgent, unsafeTrace),
		IdentityProviderClient:              newIdentityProviderClient(httpClient, url, userAgent, unsafeTrace),
		KafkaQuotasClient:                   newKafkaQuotasClient(httpClient, url, userAgent, unsafeTrace),
		KsqlClient:                          newKsqlClient(httpClient, url, userAgent, unsafeTrace),
		MdsClient:                           newMdsClient(httpClient, url, userAgent, unsafeTrace),
		NetworkingClient:                    newNetworkingClient(httpClient, url, userAgent, unsafeTrace),
		NetworkingDnsForwarderClient:        newNetworkingDnsForwarderClient(httpClient, url, userAgent, unsafeTrace),
		NetworkingIpClient:                  newNetworkingIpClient(httpClient, url, userAgent, unsafeTrace),
		NetworkingOutboundPrivateLinkClient: newNetworkingOutboundPrivateLinkClient(httpClient, url, userAgent, unsafeTrace),
		NetworkingPrivateLinkClient:         newNetworkingPrivateLinkClient(httpClient, url, userAgent, unsafeTrace),
		OrgClient:                           newOrgClient(httpClient, url, userAgent, unsafeTrace),
		ServiceQuotaClient:                  newServiceQuotaClient(httpClient, url, userAgent, unsafeTrace),
		SrcmClient:                          newSrcmClient(httpClient, url, userAgent, unsafeTrace),
		SsoClient:                           newSsoClient(httpClient, url, userAgent, unsafeTrace),
		StreamDesignerClient:                newStreamDesignerClient(httpClient, url, userAgent, unsafeTrace),
	}
}
