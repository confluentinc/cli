package lsp

import (
	"context"
	"encoding/json"
	"math/rand"
	"time"

	"github.com/confluentinc/cli/v3/pkg/flink/types"
	"github.com/confluentinc/cli/v3/pkg/log"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

type ClientHandler struct {
	inputController func() types.InputControllerInterface
}

func (h *ClientHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	switch req.Method {
	case "textDocument/publishDiagnostics":
		var params lsp.PublishDiagnosticsParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return
		}

		if h.inputController() == nil {
			log.CliLogger.Debug("input controller is nil")
			return
		}

		h.inputController().SetDiagnostics(params.Diagnostics)
	}

}

func NewLspHandler(inputController func() types.InputControllerInterface) *ClientHandler {

	go func() {
		for {
			time.Sleep(5 * time.Second)
			if inputController() == nil {
				continue
			}

			diagnostics := []lsp.Diagnostic{}

			diagnosticsCount := rand.Intn(3) + 1
			for i := 0; i < diagnosticsCount; i++ {
				diagnosticPos := rand.Intn(10)

				diagnostics = append(diagnostics,
					lsp.Diagnostic{
						Range: lsp.Range{
							Start: lsp.Position{Line: 0, Character: diagnosticPos},
							End:   lsp.Position{Line: 0, Character: diagnosticPos + 3},
						},
						Severity: 1,
						Code:     "1234",
						Source:   "mock source",
						Message:  "Error: this is a lsp diagnostic",
					})
			}

			inputController().SetDiagnostics(diagnostics)

		}
	}()

	return &ClientHandler{
		inputController: inputController,
	}
}
