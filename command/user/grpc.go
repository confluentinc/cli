package user

import (
	"context"
	"fmt"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/cli/shared"
	proto "github.com/confluentinc/cli/shared/user"
)

// GRPCClient is an implementation of Counter that talks over RPC.
type GRPCClient struct {
	client proto.UserClient
}

// Create Service Account
func (c *GRPCClient) CreateServiceAccount(ctx context.Context, user *orgv1.User) (*orgv1.User, error) {

	fmt.Printf("GRPC CALLing ")
	resp, err := c.client.CreateServiceAccount(ctx, &orgv1.CreateServiceAccountRequest{User: user})
	if err != nil {
		fmt.Printf("GRPC Error:  ")
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.User, nil
}

// Update Service Account
func (c *GRPCClient) UpdateServiceAccount(ctx context.Context, user *orgv1.User) error {
	_, err := c.client.UpdateServiceAccount(ctx, &orgv1.UpdateServiceAccountRequest{User: user})
	if err != nil {
		return shared.ConvertGRPCError(err)
	}
	return nil
}

// Get Service Accounts
func (c *GRPCClient) GetServiceAccounts(ctx context.Context, user *orgv1.User) ([]*orgv1.User, error) {
	resp, err := c.client.GetServiceAccounts(ctx, &orgv1.GetServiceAccountsRequest{User: user})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Users, nil
}

// Deactivate Service Account
func (c *GRPCClient) DeactivateServiceAccount(ctx context.Context, user *orgv1.User) error {
	_, err := c.client.DeactivateServiceAccount(ctx, &orgv1.DeactivateServiceAccountRequest{User: user})
	if err != nil {
		return shared.ConvertGRPCError(err)
	}
	return nil
}

// GRPCServer the GPRClient talks to. Plugin authors implement this if they're using Go.
type GRPCServer struct {
	Impl User
}

// Create Service Account
func (s *GRPCServer) CreateServiceAccount(ctx context.Context, req *orgv1.CreateServiceAccountRequest) (*orgv1.CreateServiceAccountReply, error) {
	fmt.Printf("GRPC Server")
	r, err := s.Impl.CreateServiceAccount(ctx, req.User)
	return &orgv1.CreateServiceAccountReply{User: r}, err
}

// Update Service Account
func (s *GRPCServer) UpdateServiceAccount(ctx context.Context, req *orgv1.UpdateServiceAccountRequest) (*orgv1.UpdateServiceAccountReply, error) {
	err := s.Impl.UpdateServiceAccount(ctx, req.User)
	return &orgv1.UpdateServiceAccountReply{}, err
}

// Get Service Accounts
func (s *GRPCServer) GetServiceAccounts(ctx context.Context, req *orgv1.GetServiceAccountsRequest) (*orgv1.GetServiceAccountsReply, error) {
	r, err := s.Impl.GetServiceAccounts(ctx, req.User)
	return &orgv1.GetServiceAccountsReply{Users: r}, err
}

// Deactivate Service Account
func (s *GRPCServer) DeactivateServiceAccount(ctx context.Context, req *orgv1.DeactivateServiceAccountRequest) (*orgv1.DeactivateServiceAccountReply, error) {
	err := s.Impl.DeactivateServiceAccount(ctx, req.User)
	return &orgv1.DeactivateServiceAccountReply{}, err
}
