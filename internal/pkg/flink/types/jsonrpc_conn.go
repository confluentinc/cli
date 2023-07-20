package types

import (
	"context"

	"github.com/sourcegraph/jsonrpc2"
)

// Interface for the jsonrpc2 conn struct here https://github.com/sourcegraph/jsonrpc2/blob/master/conn.go.
type Conn interface {
	Call(ctx context.Context, method string, params, result interface{}, opts ...jsonrpc2.CallOption) error
	DispatchCall(ctx context.Context, method string, params interface{}, opts ...jsonrpc2.CallOption) (jsonrpc2.Waiter, error)
	Notify(ctx context.Context, method string, params interface{}, opts ...jsonrpc2.CallOption) error
	ReplyWithError(ctx context.Context, id jsonrpc2.ID, respErr *jsonrpc2.Error) error
	SendResponse(ctx context.Context, resp *jsonrpc2.Response) error
	DisconnectNotify() <-chan struct{}
	Close() error
}
