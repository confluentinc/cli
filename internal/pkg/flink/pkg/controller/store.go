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
	ProcessStatement(statement string) (*StatementResult, error)
}

type Store struct {
	MockData         []StatementResult `json:"data"`
	index            int
	Config           map[string]string
	StatementResults []StatementResult
	client           *GatewayClient
	appOptions       *ApplicationOptions
}

func (s *Store) ProcessLocalStatement(statement string) (*StatementResult, error) {
	switch statementType := parseStatementType(statement); statementType {
	case SET_STATEMENT:
		return processSetStatement(statement, s)
	case RESET_STATEMENT:
		return processResetStatement(statement, s)
	case USE_STATEMENT:
		return processUseStatement(statement, s)
	default:
		return nil, nil
	}
}

func (s *Store) ProcessStatement(statement string) (*StatementResult, error) {
	// We trim the statement here once so we don't have to do it in every function
	statement = strings.TrimSpace(statement)

	// Process local statements: set, use, reset
	result, err := s.ProcessLocalStatement(statement)
	if result != nil || err != nil {
		return result, err
	}

	// TODO: Remove this once we have a real backend
	if s.appOptions.MOCK_STATEMENTS_OUTPUT_DEMO {
		if !startsWithValidSQL(statement) {
			return nil, &StatementError{"Error: Invalid syntax. Please check your statement."}
		} else {
			s.index++
			return &s.MockData[s.index%len(s.MockData)], nil
		}
	}

	// Process remote statements
	statementObj, resp, err := s.client.CreateStatement(context.Background(), statement, s.Config)
	err = processHttpErrors(resp, err)

	if err != nil {
		return nil, &StatementError{err.Error()}
	}

	return &StatementResult{
		Message: *statementObj.Status.Detail,
		Status:  PHASE(statementObj.Status.Phase),
	}, nil

	/* TODO Result handling
	here's where we will probably fetch results - at least the first page
	*/
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
