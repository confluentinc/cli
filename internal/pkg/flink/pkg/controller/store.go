package controller

import (
	"context"
	_ "embed"
	"github.com/confluentinc/flink-sql-client/pkg/converter"
	"github.com/confluentinc/flink-sql-client/pkg/types"
	"github.com/confluentinc/flink-sql-client/test/generators"
	"strings"
)

const (
	//ops
	configOpSet               = "SET"
	configOpUse               = "USE"
	configOpReset             = "RESET"
	configOpUseCatalog        = "CATALOG"
	configStatementTerminator = ";"

	//keys
	configKeyCatalog          = "default_catalog"
	configKeyDatabase         = "default_database"
	configKeyOrgResourceId    = "org-resource-id"
	configKeyExecutionRuntime = "execution.runtime-mode"
)

const MOCK_STATEMENTS_OUTPUT_DEMO = true

type StoreInterface interface {
	ProcessStatement(statement string) (*types.ProcessedStatement, *types.StatementError)
	FetchStatementResults(types.ProcessedStatement) (*types.ProcessedStatement, *types.StatementError)
}

type Store struct {
	Properties          map[string]string
	ProcessedStatements []types.ProcessedStatement
	client              *GatewayClient
	appOptions          *ApplicationOptions
}

func (s *Store) ProcessLocalStatement(statement string) (*types.ProcessedStatement, *types.StatementError) {
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

func (s *Store) ProcessStatement(statement string) (*types.ProcessedStatement, *types.StatementError) {
	// We trim the statement here once so we don't have to do it in every function
	statement = strings.TrimSpace(statement)

	// Process local statements: set, use, reset
	result, sErr := s.ProcessLocalStatement(statement)
	if result != nil || sErr != nil {
		return result, sErr
	}

	// TODO: Remove this once we have a real backend
	if s.appOptions != nil && s.appOptions.MOCK_STATEMENTS_OUTPUT_DEMO {

		if !startsWithValidSQL(statement) {
			return nil, &types.StatementError{Msg: "Error: Invalid syntax '" + statement + "'. Please check your statement."}
		} else {
			return &types.ProcessedStatement{}, nil
		}
	}

	// Process remote statements
	statementObj, resp, err := s.client.CreateStatement(context.Background(), statement, s.Properties)
	err = processHttpErrors(resp, err)

	if err != nil {
		return nil, &types.StatementError{Msg: err.Error()}
	}
	return &types.ProcessedStatement{
		StatementName: statementObj.Spec.GetStatementName(),
		StatusDetail:  statementObj.Status.GetDetail(),
		Status:        types.PHASE(statementObj.Status.GetPhase()),
		ResultSchema:  statementObj.Status.GetResultSchema(),
	}, nil
}

func (s *Store) FetchStatementResults(statement types.ProcessedStatement) (*types.ProcessedStatement, *types.StatementError) {
	if statement.IsLocalStatement {
		return &statement, nil
	}
	// Process remote statements
	statementResultObj, resp, err := s.client.FetchStatementResults(context.Background(), statement.StatementName, statement.PageToken)
	err = processHttpErrors(resp, err)

	// TODO: Remove this once we have a real backend
	if s.appOptions != nil && s.appOptions.MOCK_STATEMENTS_OUTPUT_DEMO {
		mockResults := generators.MockResults(5, 2).Example()
		statementResults := mockResults.StatementResults
		resultSchema := mockResults.ResultSchema
		convertedResults, err := converter.ConvertToInternalResults(statementResults.Results.GetData(), resultSchema)
		if err != nil {
			return nil, &types.StatementError{Msg: err.Error()}
		}
		statement.StatementResults = convertedResults
		return &statement, nil
	}
	if err != nil {
		return nil, &types.StatementError{Msg: err.Error()}
	}

	results := statementResultObj.GetResults()
	convertedResults, err := converter.ConvertToInternalResults(results.GetData(), statement.ResultSchema)
	if err != nil {
		return nil, &types.StatementError{Msg: err.Error()}
	}

	statement.StatementResults = convertedResults
	metadata := statementResultObj.GetMetadata()
	statement.PageToken = metadata.GetNext()
	return &statement, nil
}

func NewStore(client *GatewayClient, appOptions *ApplicationOptions) StoreInterface {
	defaultProperties := make(map[string]string)

	if appOptions != nil && appOptions.DEFAULT_PROPERTIES != nil {
		defaultProperties = appOptions.DEFAULT_PROPERTIES
	}

	store := Store{
		Properties: defaultProperties,
		client:     client,
		appOptions: appOptions,
	}

	return &store
}
