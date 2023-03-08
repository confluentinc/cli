package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	v2 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const (
	//ops
	configOpSet        = "SET"
	configOpUse        = "USE"
	configOpUseCatalog = "CATALOG"
	//keys
	configKeyCatalog  = "default_catalog"
	configKeyDatabase = "default_database"
)

// custom type for now until we have the SDK for result API
type StatementResult struct {
	Status  string     `json:"status"`
	Columns []string   `json:"columns"`
	Rows    [][]string `json:"rows"`
}

type Store struct {
	MockData         []StatementResult `json:"data"`
	index            int
	Config           map[string]string
	StatementResults []StatementResult
	client           *v2.APIClient
}

func generateUUID() string {
	id := uuid.New()
	return id.String()
}

// removes semicolon from end if it is present while ignoring whitespaces
func removeQueryTerminator(query string) string {
	idxOfSemicolon := strings.LastIndex(query, ";")
	if idxOfSemicolon != -1 {
		query = query[:idxOfSemicolon] + query[idxOfSemicolon+1:]
	}
	return query
}

// removes spaces, tabs and newlines
func removeWhiteSpaces(str string) string {
	replacer := strings.NewReplacer(" ", "", "\t", "", "\n", "")
	return replacer.Replace(str)
}

func queryStartsWithOp(query string, op string) bool {
	pattern := fmt.Sprintf("^%s\\b", op)
	cleanedQuery := strings.TrimSpace(strings.ToUpper(query))
	startsWithOp, _ := regexp.MatchString(pattern, cleanedQuery)
	return startsWithOp
}

/*
Expected query: "SET key=value"
Steps to parse:
1. Remove the semicolon if present
2. Extract the substring after SET: "SET key=value" -> "key=value"
3. Replace all whitespaces from this substring
4. Then split the substring by the equals sign: "key=value" -> ["key", "value"]
5. If the resulting array length is not equal to two or the extracted key is empty, return directly
6. Otherwise, return the extracted key and value (value is allowed to be empty)
*/
func parseSETQuery(query string) (string, string) {
	query = removeQueryTerminator(query)

	indexOfSet := strings.Index(strings.ToUpper(query), configOpSet)
	if indexOfSet == -1 {
		return "", ""
	}
	startOfStrAfterSet := indexOfSet + len(configOpSet)
	if startOfStrAfterSet >= len(query) {
		return "", ""
	}
	strAfterSet := query[startOfStrAfterSet:]

	strAfterSet = removeWhiteSpaces(strAfterSet)
	keyValuePair := strings.Split(strAfterSet, "=")
	if len(keyValuePair) != 2 || keyValuePair[0] == "" {
		return "", ""
	}

	return keyValuePair[0], keyValuePair[1]
}

/*
Expected query: "USE CATALOG catalog_name" or "USE database_name"
Steps to parse:
1. Remove semicolon if present
2. Split into words
3. If resulting array length is smaller than 2 directly return
4. If word length is 2, first word is "use" and second word IS NOT "catalog", second word is the database name
5. If word length is 3, first word is "use" and second word IS "catalog", third word is the catalog name
6. Otherwise, return empty
*/
func parseUSEQuery(query string) (string, string) {
	query = removeQueryTerminator(query)
	words := strings.Fields(query)
	if len(words) < 2 {
		return "", ""
	}

	isFirstWordUse := strings.ToUpper(words[0]) == configOpUse
	isSecondWordCatalog := strings.ToUpper(words[1]) == configOpUseCatalog
	//handle "USE database_name" query
	if len(words) == 2 && isFirstWordUse && !isSecondWordCatalog {
		return configKeyDatabase, words[1]
	}

	//handle "USE CATALOG catalog_name" query
	if len(words) == 3 && isFirstWordUse && isSecondWordCatalog {
		return configKeyCatalog, words[2]
	}

	return "", ""
}

func (s *Store) waitForStatementExecution(envId, statementId string) (*StatementResult, error) {
	//TODO result handling: https://confluentinc.atlassian.net/wiki/spaces/FLINK/pages/3004703887/WIP+Flink+Gateway+-+Results+handling
	return &StatementResult{
		Status:  "Completed",
		Columns: []string{},
		Rows:    [][]string{{}},
	}, nil
}

func (s *Store) submitStatement(ctx context.Context, authToken, envId, computePoolId, query string) (v2.SqlV1alpha1Statement, *http.Response, error) {
	statementName, ok := s.Config["pipeline.name"]
	if !ok || strings.TrimSpace(statementName) == "" {
		statementName = generateUUID()
	}
	statement := v2.SqlV1alpha1Statement{
		Spec: &v2.SqlV1alpha1StatementSpec{
			StatementName: &statementName,
			Statement:     &query,
			Environment:   &v2.GlobalObjectReference{Id: envId},
			ComputePool:   &v2.EnvScopedObjectReference{Id: computePoolId},
		},
	}
	ctx = context.WithValue(ctx, v2.ContextAccessToken, authToken)
	createdStatement, resp, err := s.client.StatementsSqlV1alpha1Api.CreateSqlV1alpha1Statement(ctx).SqlV1alpha1Statement(statement).Execute()
	return createdStatement, resp, err
}

func (s *Store) ProcessQuery(query string) (*StatementResult, error) {
	isSETQuery := queryStartsWithOp(query, configOpSet)
	if isSETQuery {
		configKey, configVal := parseSETQuery(query)
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
			Status:  "Completed",
			Columns: []string{"Key", "Value"},
			Rows:    [][]string{{configKey, configVal}},
		}, nil
	}

	isUSEQuery := queryStartsWithOp(query, configOpUse)
	if isUSEQuery {
		configKey, configVal := parseUSEQuery(query)
		if configKey == "" {
			return nil, errors.New("Parsing USE query failed")
		}
		s.Config[configKey] = configVal
		return &StatementResult{
			Status:  "Completed",
			Columns: []string{"Key", "Value"},
			Rows:    [][]string{{configKey, configVal}},
		}, nil
	}

	//TODO
	authToken := ""
	envId := ""
	computePoolId := ""
	//return mock data
	if authToken == "" {
		s.index++
		return &s.MockData[s.index%len(s.MockData)], nil
	}

	statement, _, err := s.submitStatement(context.Background(), authToken, envId, computePoolId, query)
	if err != nil {
		return nil, errors.New("Could not create statement")
	}
	spec := statement.GetSpec()
	environment := spec.GetEnvironment()
	executionResult, err := s.waitForStatementExecution(environment.GetId(), statement.GetId())
	return executionResult, err
}

func NewStore(client *v2.APIClient) Store {
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

	return store
}
