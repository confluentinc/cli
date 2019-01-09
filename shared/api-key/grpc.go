package apiKey

import (
	"context"
	authv1 "github.com/confluentinc/ccloudapis/auth/v1"
	"github.com/confluentinc/cli/shared"
	chttp "github.com/confluentinc/ccloud-sdk-go"
)

var _ chttp.APIKey = (*GRPCClient)(nil)

// GRPCClient is an implementation of ApiKeyClient that talks over RPC.
type GRPCClient struct {
	client ApiKeyClient
}

// Create API key
func (c *GRPCClient) Create(ctx context.Context, key *authv1.APIKey) (*authv1.APIKey, error) {
	resp, err := c.client.Create(ctx, &authv1.CreateAPIKeyRequest{ApiKey: key})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.ApiKey, nil
}

// Delete API key
func (c *GRPCClient) Delete(ctx context.Context, key *authv1.APIKey) error {
	_, err := c.client.Delete(ctx, &authv1.DeleteAPIKeyRequest{ApiKey: key})
	if err != nil {
		return shared.ConvertGRPCError(err)
	}
	return nil
}

// GRPCServer the GPRClient talks to. Plugin authors implement this if they're using Go.
type GRPCServer struct {
	Impl chttp.APIKey
}

// Create API Key
func (s *GRPCServer) Create(ctx context.Context, req *authv1.CreateAPIKeyRequest) (*authv1.CreateAPIKeyReply, error) {
	r, err := s.Impl.Create(ctx, req.ApiKey)
	return &authv1.CreateAPIKeyReply{ApiKey: r}, err
}

// Delete API Key
func (s *GRPCServer) Delete(ctx context.Context, req *authv1.DeleteAPIKeyRequest)(*authv1.DeleteAPIKeyReply, error) {
	err := s.Impl.Delete(ctx, req.ApiKey)
	return &authv1.DeleteAPIKeyReply{},err
}
