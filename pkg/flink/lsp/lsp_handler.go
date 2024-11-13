package lsp

import (
	"context"

	"github.com/sourcegraph/jsonrpc2"

	"github.com/confluentinc/cli/v4/pkg/log"
)

type LSPHandler struct {
	handlerCh chan *jsonrpc2.Request
}

//  All we do here is to pass the request to the channel since the controller that has to process and do something with the request is the InputController
func (h *LSPHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	select {
	case h.handlerCh <- req:
		// Successfully sent the request
	default:
		// Channel is closed or full
		log.CliLogger.Warn("handlerCh is closed or full, unable to send request to handler")
	}
}

func NewLspHandler(handlerCh chan *jsonrpc2.Request) *LSPHandler {
	return &LSPHandler{
		handlerCh: handlerCh,
	}
}
