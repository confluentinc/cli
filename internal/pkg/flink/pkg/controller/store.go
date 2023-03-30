package controller

import (
	"context"
	_ "embed"
	"encoding/json"
	"strings"
)

//go:embed mock-data.json
var mockData []byte

const (
	//ops
	configOpSet               = "SET"
	configOpUse               = "USE"
	configOpReset             = "RESET"
	configOpUseCatalog        = "CATALOG"
	configStatementTerminator = ";"
	//keys
	configKeyCatalog  = "default_catalog"
	configKeyDatabase = "default_database"
)

const MOCK_STATEMENTS_OUTPUT_DEMO = true

type StoreInterface interface {
	ProcessStatement(statement string) (*StatementResult, *StatementError)
}

type Store struct {
	MockData         []StatementResult `json:"data"`
	index            int
	Config           map[string]string
	StatementResults []StatementResult
	client           *GatewayClient
	appOptions       *ApplicationOptions
}

func (s *Store) ProcessLocalStatement(statement string) (*StatementResult, *StatementError) {
	switch statementType := parseStatementType(statement); statementType {
	case SET_STATEMENT:
		return s.processSetStatement(statement)
	case RESET_STATEMENT:
		return s.processResetStatement(statement)
	case USE_STATEMENT:
		return s.processUseStatement(statement)
	default:
		return nil, nil
	}
}

func (s *Store) ProcessStatement(statement string) (*StatementResult, *StatementError) {
	// We trim the statement here once so we don't have to do it in every function
	statement = strings.TrimSpace(statement)

	// Process local statements: set, use, reset
	result, sErr := s.ProcessLocalStatement(statement)
	if result != nil || sErr != nil {
		return result, sErr
	}

	// TODO: Remove this once we have a real backend
	if s.appOptions.MOCK_STATEMENTS_OUTPUT_DEMO {

		if !startsWithValidSQL(statement) {
			return nil, &StatementError{Msg: "Error: Invalid syntax. Please check your statement."}
		} else {
			s.index++
			return &s.MockData[s.index%len(s.MockData)], nil
		}
	}

	// Process remote statements
	statementObj, resp, err := s.client.CreateStatement(context.Background(), statement, s.Config)
	err = processHttpErrors(resp, err)

	if err != nil {
		return nil, &StatementError{Msg: err.Error()}
	}

	return &StatementResult{
		StatusDetail: *statementObj.Status.Detail,
		Status:       PHASE(statementObj.Status.GetPhase()),
	}, nil
}

func NewStore(client *GatewayClient, appOptions *ApplicationOptions) StoreInterface {

	store := Store{
		Config:     map[string]string{},
		index:      0,
		client:     client,
		appOptions: appOptions,
	}
	// Opening mock data
	json.Unmarshal(mockData, &store)

	return &store
}
