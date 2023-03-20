package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	v1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink-gateway/v1alpha1"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

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

type StoreInterface interface {
	ProcessStatement(statement string) (*StatementResult, error)
}

type Store struct {
	MockData         []StatementResult `json:"data"`
	index            int
	Config           map[string]string
	StatementResults []StatementResult
	client           *v1.APIClient
}

func (s *Store) submitStatement(ctx context.Context, authToken, envId, orgId, computePoolId, statement string) (v1.SqlV1alpha1Statement, *http.Response, error) {
	statementName, ok := s.Config["pipeline.name"]
	if !ok || strings.TrimSpace(statementName) == "" {
		statementName = uuid.New().String()
	}
	statementObj := v1.SqlV1alpha1Statement{
		Spec: &v1.SqlV1alpha1StatementSpec{
			StatementName: &statementName,
			Statement:     &statement,
			ComputePoolId: &computePoolId,
			// Properties: todo - add local config to properties
		},
	}

	ctx = context.WithValue(ctx, v1.ContextAccessToken, authToken)
	createdStatement, resp, err := s.client.StatementsSqlV1alpha1Api.CreateSqlV1alpha1Statement(ctx, envId).SqlV1alpha1Statement(statementObj).Execute()
	return createdStatement, resp, err
}

func (s *Store) waitForStatementExecution(envId, statementId string) (*StatementResult, error) {
	//TODO result handling: https://confluentinc.atlassian.net/wiki/spaces/FLINK/pages/3004703887/WIP+Flink+Gateway+-+Results+handling
	return &StatementResult{
		Status:  "Completed",
		Columns: []string{},
		Rows:    [][]string{{}},
	}, nil
}

func (s *Store) ProcessLocalStatement(statement string) (*StatementResult, error) {
	statementType := parseStatementType(statement)
	switch statementType {
	case SET_STATEMENT:
		configKey, configVal, err := parseSETStatement(statement)
		if err != nil {
			return nil, err
		}
		if configKey == "" {
			//return current config
			return &StatementResult{
				Status:  "Completed",
				Columns: []string{"Key", "Value"},
				Rows:    lo.MapToSlice(s.Config, func(key, val string) []string { return []string{key, val} }),
			}, nil
		}
		s.Config[configKey] = configVal
		//return only new config row
		return &StatementResult{
			Message: "Config updated successfuly.",
			Status:  "Completed",
			Columns: []string{"Key", "Value"},
			Rows:    [][]string{{configKey, configVal}},
		}, nil
	case USE_STATEMENT:
		configKey, configVal, err := parseUSEStatement(statement)
		if err != nil {
			return nil, err
		}

		s.Config[configKey] = configVal
		return &StatementResult{
			Message: "Config updated successfuly.",
			Status:  "Completed",
			Columns: []string{"Key", "Value"},
			Rows:    [][]string{{configKey, configVal}},
		}, nil
	default:
		return nil, nil
	}
}

func (s *Store) ProcessStatement(statement string) (*StatementResult, error) {
	// Process local statements: set, use, reset
	result, err := s.ProcessLocalStatement(statement)
	if result != nil || err != nil {
		return result, err
	}

	// This is where we currently mock results, since we don't have a real backend yet
	// TODO -> we'll receive these from the cli
	authToken := ""
	orgId := ""
	envId := ""
	computePoolId := ""
	//return mock data
	if authToken == "" {
		if !startsWithValidSQL(statement) {
			return nil, &StatementError{"Error: Invalid syntax. Please check your statement."}
		} else {
			s.index++
			return &s.MockData[s.index%len(s.MockData)], nil
		}
	}

	// Process remote statements
	statementObj, resp, err := s.submitStatement(context.Background(), authToken, envId, orgId, computePoolId, statement)
	err = processHttpErrors(resp, err)
	if err != nil {
		return nil, &StatementError{err.Error()}
	}

	return &StatementResult{
		Message: *statementObj.Status.Detail,
		Status:  PHASE(statementObj.Status.Phase),
	}, nil

	/* // TODO Result handling
	executionResult, err := s.waitForStatementExecution(envId, statement.GetId())
	return executionResult, err */
}

func NewStore(client *v1.APIClient) StoreInterface {
	store := Store{
		Config: map[string]string{},
		index:  0,
		client: client,
	}
	// Opening mock data
	jsonFile, err := os.ReadFile("mock-data.json")
	if err != nil {
		fmt.Println(err)
	}
	json.Unmarshal(jsonFile, &store)

	return &store
}
