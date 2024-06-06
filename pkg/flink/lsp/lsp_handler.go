package lsp

import (
	"context"

	"github.com/sourcegraph/jsonrpc2"
)

type LSPHandler struct {
	handlerCh chan *jsonrpc2.Request
}

//  All we do here is to pass the request to the channel since the controller that has to process and do something with the request is the InputController
func (h *LSPHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	h.handlerCh <- req
}

func NewLspHandler(handlerCh chan *jsonrpc2.Request) *LSPHandler {
	return &LSPHandler{
		handlerCh: handlerCh,
	}
}
