package ccloudv2

import (
	"context"

	aiv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/ai/v1"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

func newAiClient(url, userAgent string, unsafeTrace bool) *aiv1.APIClient {
	cfg := aiv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = NewRetryableHttpClient(unsafeTrace)
	cfg.Servers = aiv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return aiv1.NewAPIClient(cfg)
}

func (c *Client) aiApiContext() context.Context {
	return context.WithValue(context.Background(), aiv1.ContextAccessToken, c.AuthToken)
}

func (c *Client) QueryChatCompletion(req aiv1.AiV1ChatCompletionsRequest) (aiv1.AiV1ChatCompletionsReply, error) {
	res, httpResp, err := c.AiClient.ChatCompletionsAiV1Api.QueryAiV1ChatCompletion(c.aiApiContext()).AiV1ChatCompletionsRequest(req).Execute()
	return res, errors.CatchCCloudV2Error(err, httpResp)
}
