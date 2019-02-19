package user

import (
	"context"
	chttp "github.com/confluentinc/ccloud-sdk-go"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	"github.com/confluentinc/cli/shared"
)

var _ chttp.User = (*GRPCClient)(nil)

// GRPCClient is an implementation of UserClient that talks over RPC.
type GRPCClient struct {
	client UserClient
}

// Create Service Account
func (c *GRPCClient) CreateServiceAccount(ctx context.Context, user *orgv1.User) (*orgv1.User, error) {

	resp, err := c.client.CreateServiceAccount(ctx, &orgv1.CreateServiceAccountRequest{User: user})
	if err != nil {
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
func (c *GRPCClient) GetServiceAccounts(ctx context.Context) ([]*orgv1.User, error) {
	resp, err := c.client.GetServiceAccounts(ctx, &orgv1.GetServiceAccountsRequest{})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Users, nil
}

// Delete Service Account
func (c *GRPCClient) DeleteServiceAccount(ctx context.Context, user *orgv1.User) error {
	_, err := c.client.DeleteServiceAccount(ctx, &orgv1.DeleteServiceAccountRequest{User: user})
	if err != nil {
		return shared.ConvertGRPCError(err)
	}
	return nil
}

// Describe User
func (c *GRPCClient) Describe(ctx context.Context, user *orgv1.User) (*orgv1.User, error) {
	resp, err := c.client.Describe(ctx, &orgv1.GetUserRequest{User: user})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.User, nil
}

// List Users
func (c *GRPCClient) List(ctx context.Context) ([]*orgv1.User, error) {
	resp, err := c.client.List(ctx, &orgv1.GetUsersRequest{})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Users, nil
}

// GRPCServer the GPRClient talks to. Plugin authors implement this if they're using Go.
type GRPCServer struct {
	Impl chttp.User
}

// Create Service Account
func (s *GRPCServer) CreateServiceAccount(ctx context.Context, req *orgv1.CreateServiceAccountRequest) (*orgv1.CreateServiceAccountReply, error) {
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
	r, err := s.Impl.GetServiceAccounts(ctx)
	return &orgv1.GetServiceAccountsReply{Users: r}, err
}

// Deactivate Users
func (s *GRPCServer) DeleteServiceAccount(ctx context.Context, req *orgv1.DeleteServiceAccountRequest) (*orgv1.DeleteServiceAccountReply, error) {
	err := s.Impl.DeleteServiceAccount(ctx, req.User)
	return &orgv1.DeleteServiceAccountReply{}, err
}

// Describe User
func (s *GRPCServer) Describe(ctx context.Context, req *orgv1.GetUserRequest) (*orgv1.GetUserReply, error) {
	r, err := s.Impl.Describe(ctx, req.User)
	return &orgv1.GetUserReply{User: r}, err
}

// List User
func (s *GRPCServer) List(ctx context.Context, req *orgv1.GetUsersRequest) (*orgv1.GetUsersReply, error) {
	r, err := s.Impl.List(ctx)
	return &orgv1.GetUsersReply{Users: r}, err
}
