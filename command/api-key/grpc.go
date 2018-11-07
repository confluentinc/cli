package apiKey

import (
	"context"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/shared"
	proto "github.com/confluentinc/cli/shared/api-key"
)

// GRPCClient is an implementation of Counter that talks over RPC.
type GRPCClient struct {
	client proto.ApiKeyClient
}

// Create API key
func (c *GRPCClient) Create(ctx context.Context, key *schedv1.ApiKey) (*schedv1.ApiKey, error) {
	resp, err := c.client.Create(ctx, &schedv1.CreateApiKeyRequest{ApiKey: key})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.ApiKey, nil
}

// Delete API key
func (c *GRPCClient) Delete(ctx context.Context, key *schedv1.ApiKey) error {
	_, err := c.client.Delete(ctx, &schedv1.DeleteApiKeyRequest{ApiKey: key})
	if err != nil {
		return shared.ConvertGRPCError(err)
	}
	return nil
}

// GRPCServer the GPRClient talks to. Plugin authors implement this if they're using Go.
type GRPCServer struct {
	Impl ApiKey
}

// Create API Key
func (s *GRPCServer) Create(ctx context.Context, req *schedv1.CreateApiKeyRequest) (*schedv1.CreateApiKeyReply, error) {
	r, err := s.Impl.Create(ctx, req.ApiKey)
	return &schedv1.CreateApiKeyReply{ApiKey: r}, err
}

// Delete API Key
func (s *GRPCServer) Delete(ctx context.Context, req *schedv1.DeleteApiKeyRequest)(*schedv1.DeleteApiKeyReply, error) {
	err := s.Impl.Delete(ctx, req.ApiKey)
	return &schedv1.DeleteApiKeyReply{},err
}
