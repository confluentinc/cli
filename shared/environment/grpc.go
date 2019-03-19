package environment

import (
	"context"
	chttp "github.com/confluentinc/ccloud-sdk-go"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	"github.com/confluentinc/cli/shared"
)

var _ chttp.Account = (*GRPCClient)(nil)

// GRPCClient is an implementation of AccountClient that talks over RPC.
type GRPCClient struct {
	client AccountClient
}

// Create account
func (c *GRPCClient) Create(ctx context.Context, account *orgv1.Account) (*orgv1.Account, error) {
	resp, err := c.client.Create(ctx, &orgv1.CreateAccountRequest{Account: account})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Account, nil
}

// Update account
func (c *GRPCClient) Update(ctx context.Context, account *orgv1.Account) error {
	_, err := c.client.Update(ctx, &orgv1.UpdateAccountRequest{Account: account})
	if err != nil {
		return shared.ConvertGRPCError(err)
	}
	return nil
}

// Delete account
func (c *GRPCClient) Delete(ctx context.Context, account *orgv1.Account) error {
	_, err := c.client.Delete(ctx, &orgv1.DeleteAccountRequest{Account: account})
	if err != nil {
		return shared.ConvertGRPCError(err)
	}
	return nil
}

// Get account
func (c *GRPCClient) Get(ctx context.Context, account *orgv1.Account) (*orgv1.Account, error) {
	reply, err := c.client.Get(ctx, &orgv1.GetAccountRequest{Account: account})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return reply.Account, nil
}

// List accounts
func (c *GRPCClient) List(ctx context.Context, account *orgv1.Account) ([]*orgv1.Account, error) {
	// TODO: we should pass in the account here, but it currently expects an org.  Should investigate/change that behavior.
	reply, err := c.client.List(ctx, &orgv1.ListAccountsRequest{})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return reply.Accounts, nil
}

// GRPCServer the GPRClient talks to. Plugin authors implement this if they're using Go.
type GRPCServer struct {
	Impl chttp.Account
}

// Create account
func (s *GRPCServer) Create(ctx context.Context, req *orgv1.CreateAccountRequest) (*orgv1.CreateAccountReply, error) {
	r, err := s.Impl.Create(ctx, req.Account)
	return &orgv1.CreateAccountReply{Account: r}, shared.ConvertGRPCError(err)
}

// Update account
func (s *GRPCServer) Update(ctx context.Context, req *orgv1.UpdateAccountRequest) (*orgv1.UpdateAccountReply, error) {
	err := s.Impl.Update(ctx, req.Account)
	return &orgv1.UpdateAccountReply{}, shared.ConvertGRPCError(err)
}

// Delete account
func (s *GRPCServer) Delete(ctx context.Context, req *orgv1.DeleteAccountRequest) (*orgv1.DeleteAccountReply, error) {
	err := s.Impl.Delete(ctx, req.Account)
	return &orgv1.DeleteAccountReply{}, shared.ConvertGRPCError(err)
}

// Get account
func (s *GRPCServer) Get(ctx context.Context, req *orgv1.GetAccountRequest) (*orgv1.GetAccountReply, error) {
	r, err := s.Impl.Get(ctx, req.Account)
	return &orgv1.GetAccountReply{Account: r}, shared.ConvertGRPCError(err)
}

// List accounts
func (s *GRPCServer) List(ctx context.Context, req *orgv1.ListAccountsRequest) (*orgv1.ListAccountsReply, error) {
	r, err := s.Impl.List(ctx, nil)
	return &orgv1.ListAccountsReply{Accounts: r}, shared.ConvertGRPCError(err)
}
