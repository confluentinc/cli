package autocomplete

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"

	"github.com/confluentinc/flink-sql-language-service/pkg/api"
	prompt "github.com/confluentinc/go-prompt"

	"github.com/confluentinc/cli/v3/pkg/flink/types"
	"github.com/confluentinc/cli/v3/pkg/log"
)

type LSPClientWS struct {
	conn         *websocket.Conn
	documentURI  *lsp.DocumentURI
	store        types.StoreInterface
	responseChan chan *jsonrpc2.Response
}

func (c *LSPClientWS) LSPCompleter(in prompt.Document) []prompt.Suggest {
	err := c.didChange(in.Text)

	if err != nil {
		log.CliLogger.Debugf("Error sending didChange lsp request: %v\n", err)
		return []prompt.Suggest{}
	}

	textBeforeCursor := in.TextBeforeCursor()

	position := lsp.Position{
		Line:      0,
		Character: len(textBeforeCursor),
	}

	completions := []lsp.CompletionItem{}
	if textBeforeCursor != "" {
		completionList, err := c.completion(position)

		if err != nil {
			log.CliLogger.Debugf("Error sending completion lsp request: %v\n", err)
			return []prompt.Suggest{}
		}

		completions = completionList.Items
	}

	return lspCompletionsToSuggests(completions)
}

func (c *LSPClientWS) initialize() (*lsp.InitializeResult, error) {
	var resp lsp.InitializeResult

	if c.conn == nil {
		return nil, errors.New("connection to LSP server not established/nil")
	}

	err := c.sendMessage("initialize", lsp.InitializeParams{}, &resp, nil)
	if err != nil {
		log.CliLogger.Debugf("Error initializing LSP: %v\n", err)
	}

	return &resp, err
}

func (c *LSPClientWS) sendMessage(method string, params interface{}, resp interface{}, meta interface{}) error {
	requestBytes, err := buildRequest(method, params, meta).MarshalJSON()
	if err != nil {
		fmt.Printf("Error building request %v\n", err)
		return err
	}

	err = c.conn.WriteMessage(websocket.TextMessage, requestBytes)
	if err != nil {
		fmt.Printf("Error initializing LSP: %v\n", err)
		return err
	}

	response := <-c.responseChan
	if response == nil {
		return errors.New("error waiting for response")
	}

	err = json.Unmarshal(*response.Result, &resp)
	if err != nil {
		fmt.Printf("Error unmarshalling initialize response: %v\n", err)
		return err
	}

	return nil
}

func buildRequest(method string, params interface{}, meta interface{}) *jsonrpc2.Request {
	req := &jsonrpc2.Request{Method: method}
	if params != nil {
		if err := req.SetParams(params); err != nil {
			return nil
		}
	}
	if meta != nil {
		if err := req.SetMeta(meta); err != nil {
			return nil
		}
	}
	return req
}

func (c *LSPClientWS) didOpen() error {
	var resp interface{}

	documentURI := lsp.DocumentURI("temp_session_" + uuid.New().String() + ".sql")

	didOpenParams := lsp.DidOpenTextDocumentParams{
		TextDocument: lsp.TextDocumentItem{
			URI:  documentURI,
			Text: "",
		},
	}

	if c.conn == nil {
		return errors.New("connection to LSP server not established/nil")
	}

	err := c.sendMessage("textDocument/didOpen", didOpenParams, &resp, nil)

	if err != nil {
		log.CliLogger.Debugf("Error sending request: %v\n", err)
		return err
	}
	c.documentURI = &documentURI
	return nil
}

func (c *LSPClientWS) didChange(newText string) error {
	var resp interface{}

	if c.conn == nil || c.documentURI == nil {
		return errors.New("connection to LSP server not established/nil")
	}

	didchangeParams := lsp.DidChangeTextDocumentParams{
		TextDocument: lsp.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: lsp.TextDocumentIdentifier{
				URI: *c.documentURI,
			},
		},
		ContentChanges: []lsp.TextDocumentContentChangeEvent{
			{Text: newText},
		},
	}

	err := c.sendMessage("textDocument/didChange", didchangeParams, &resp, nil)

	if err != nil {
		log.CliLogger.Debugf("Error sending request: %v\n", err)
		return err
	}
	return nil
}

func (c *LSPClientWS) completion(position lsp.Position) (lsp.CompletionList, error) {
	var resp lsp.CompletionList

	if c.conn == nil || c.documentURI == nil {
		return lsp.CompletionList{}, errors.New("connection to LSP server not established/nil")
	}

	completionParams := lsp.CompletionParams{TextDocumentPositionParams: lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{
			URI: *c.documentURI,
		},
		Position: position,
	}}
	cliCtx := api.CliContext{
		AuthToken:     c.store.GetAuthToken(),
		Catalog:       c.store.GetCurrentCatalog(),
		Database:      c.store.GetCurrentDatabase(),
		ComputePoolId: c.store.GetComputePool(),
	}

	err := c.sendMessage("textDocument/completion", completionParams, &resp, cliCtx)

	if err != nil {
		log.CliLogger.Debugf("Error sending request: %v\n", err)
		return lsp.CompletionList{}, err
	}

	return resp, nil
}

func (c *LSPClientWS) ShutdownAndExit() {
	if c.conn == nil {
		return
	}

	err := c.sendMessage("shutdown", nil, nil, nil)

	if err != nil {
		log.CliLogger.Debugf("Error shutting down lsp server: %v\n", err)
		return
	}

	err = c.sendMessage("exit", nil, nil, nil)

	if err != nil {
		log.CliLogger.Debugf("Error existing lsp server: %v\n", err)
	}
}

func receiveHandler(connection *websocket.Conn, responseChan chan *jsonrpc2.Response) {
	for {
		response := &jsonrpc2.Response{}

		_, msg, err := connection.ReadMessage()
		if err != nil {
			fmt.Printf("Error in receive: %v\n", err)
			responseChan <- nil
			continue
		}

		err = response.UnmarshalJSON(msg)
		if err != nil {
			fmt.Printf("Error during unmarshalling: %v\n", err)
			responseChan <- nil
			continue
		}
		responseChan <- response
	}
}

func NewLSPClientWS(store types.StoreInterface) LSPClientInterface {
	socketUrl := "ws://localhost:8000/lsp"
	conn, _, err := websocket.DefaultDialer.Dial(socketUrl, nil)
	if err != nil {
		fmt.Printf("Error connecting to Websocket Server: %v \n", err)
	}
	responseChan := make(chan *jsonrpc2.Response)
	go receiveHandler(conn, responseChan)

	lspClient := &LSPClientWS{
		store:        store,
		conn:         conn,
		responseChan: responseChan,
	}

	lspInitParams, err := lspClient.initialize()
	if err != nil {
		fmt.Printf("Error initializing LSP: %v\n", err)
	} else {
		fmt.Printf("LSP init params: %v\n", lspInitParams)
		err = lspClient.didOpen()
		if err != nil {
			fmt.Printf("LSP didn't open: %v\n", err)
		}
	}

	return lspClient
}
