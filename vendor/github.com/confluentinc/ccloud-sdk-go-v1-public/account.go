package ccloud

import (
	"fmt"
	"net/http"

	"github.com/dghubble/sling"
)

const (
	// ACCOUNTS base endpoint for accounts operations
	ACCOUNTS = "/api/accounts"
	// ACCOUNT specific user operation endpoint
	ACCOUNT = ACCOUNTS + "/%s"
)

// AccountService provides methods for managing accounts on Confluent Control Plane.
type AccountService struct {
	client *http.Client
	sling  *sling.Sling
}

var _ AccountInterface = (*AccountService)(nil)

// NewAccountService returns a new AccountService.
func NewAccountService(client *Client) *AccountService {
	return &AccountService{
		client: client.HttpClient,
		sling:  client.sling,
	}
}

// Create creates a new account
// When Creating a new Account following fields are required
// name, organization_id
func (s *AccountService) Create(account *Account) (*Account, error) {
	body := &CreateAccountRequest{Account: account}
	reply := new(CreateAccountReply)

	resp, err := s.sling.New().Post(ACCOUNTS).BodyProvider(Request(body)).Receive(reply, reply)
	if err := ReplyErr(reply, err); err != nil {
		errMsg := "error creating account"
		if resp != nil {
			errMsg = fmt.Sprintf("%s with response %s", errMsg, resp.Status)
		}
		return nil, WrapErr(err, errMsg)
	}
	return reply.Account, nil
}

// Get returns details for a given account.
func (s *AccountService) Get(account *Account) (*Account, error) {
	if !isValidResource(account.Id) {
		return nil, ErrNotFound
	}
	reply := new(GetAccountReply)
	resp, err := s.sling.New().Get(fmt.Sprintf(ACCOUNT, account.Id)).Receive(reply, reply)
	if err := ReplyErr(reply, err); err != nil {
		errMsg := "error retrieving account"
		if resp != nil {
			errMsg = fmt.Sprintf("%s with response %s", errMsg, resp.Status)
		}
		return nil, WrapErr(err, errMsg)
	}
	if reply.Account == nil {
		return nil, ErrNotFound
	}
	return reply.Account, nil
}

// List returns the accounts that match
func (s *AccountService) List(account *Account) ([]*Account, error) {
	reply := new(ListAccountsReply)
	resp, err := s.sling.New().Get(ACCOUNTS).QueryStruct(account).Receive(reply, reply)
	if err := ReplyErr(reply, err); err != nil {
		errMsg := "error retrieving accounts"
		if resp != nil {
			errMsg = fmt.Sprintf("%s with response %s", errMsg, resp.Status)
		}
		return nil, WrapErr(err, errMsg)
	}
	return reply.Accounts, nil
}
