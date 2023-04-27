package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/confluentinc/flink-sql-client/pkg/types"

	v1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink-gateway/v1alpha1"
	"github.com/samber/lo"

	"github.com/confluentinc/flink-sql-client/config"
)

type StatementType string

const (
	SET_STATEMENT   StatementType = configOpSet
	USE_STATEMENT   StatementType = configOpUse
	RESET_STATEMENT StatementType = configOpReset
	EXIT_STATEMENT  StatementType = configOpExit
	OTHER_STATEMENT StatementType = "OTHER"
)

func createStatementResults(columnNames []string, rows [][]string) types.StatementResults {
	var statementResultRows []types.StatementResultRow
	for _, row := range rows {
		var statementResultRow types.StatementResultRow
		for _, field := range row {
			statementResultRow.Fields = append(statementResultRow.Fields, types.AtomicStatementResultField{
				Type:  types.VARCHAR,
				Value: field,
			})
		}
		statementResultRows = append(statementResultRows, statementResultRow)
	}
	return types.StatementResults{
		Headers: columnNames,
		Rows:    statementResultRows,
	}
}

func (s *Store) processSetStatement(statement string) (*types.ProcessedStatement, *types.StatementError) {
	configKey, configVal, err := parseSetStatement(statement)
	if err != nil {
		return nil, err.(*types.StatementError)
	}
	if configKey == "" {
		statementResults := createStatementResults([]string{"Key", "Value"}, lo.MapToSlice(s.Properties, func(key, val string) []string { return []string{key, val} }))
		return &types.ProcessedStatement{
			Kind:             configOpSet,
			Status:           types.COMPLETED,
			StatementResults: &statementResults,
			IsLocalStatement: true,
		}, nil
	}
	s.Properties[configKey] = configVal

	statementResults := createStatementResults([]string{"Key", "Value"}, [][]string{{configKey, configVal}})
	return &types.ProcessedStatement{
		Kind:             configOpSet,
		StatusDetail:     "Config updated successfuly.",
		Status:           types.COMPLETED,
		StatementResults: &statementResults,
		IsLocalStatement: true,
	}, nil
}

func (s *Store) processResetStatement(statement string) (*types.ProcessedStatement, *types.StatementError) {
	configKey, err := parseResetStatement(statement)
	if err != nil {
		return nil, &types.StatementError{Msg: err.Error()}
	}
	if configKey == "" {
		s.Properties = make(map[string]string)
		return &types.ProcessedStatement{
			Kind:             configOpReset,
			StatusDetail:     "Configuration has been reset successfuly.",
			Status:           types.COMPLETED,
			IsLocalStatement: true,
		}, nil
	} else {
		_, keyExists := s.Properties[configKey]
		if !keyExists {
			return nil, &types.StatementError{Msg: fmt.Sprintf("Error: Config key \"%s\" is currently not set.", configKey)}
		}

		delete(s.Properties, configKey)
		statementResults := createStatementResults([]string{"Key", "Value"}, lo.MapToSlice(s.Properties, func(key, val string) []string { return []string{key, val} }))
		return &types.ProcessedStatement{
			Kind:             configOpReset,
			StatusDetail:     fmt.Sprintf("Config key \"%s\" has been reset successfuly.", configKey),
			Status:           types.COMPLETED,
			StatementResults: &statementResults,
			IsLocalStatement: true,
		}, nil
	}
}

func (s *Store) processUseStatement(statement string) (*types.ProcessedStatement, *types.StatementError) {
	configKey, configVal, err := parseUseStatement(statement)
	if err != nil {
		return nil, &types.StatementError{Msg: err.Error()}
	}

	s.Properties[configKey] = configVal
	statementResults := createStatementResults([]string{"Key", "Value"}, [][]string{{configKey, configVal}})
	return &types.ProcessedStatement{
		Kind:             configOpUse,
		StatusDetail:     "Config updated successfuly.",
		Status:           types.COMPLETED,
		StatementResults: &statementResults,
		IsLocalStatement: true,
	}, nil
}

/*
Expected statement: "SET key=value"
Steps to parse:
1. Remove the semicolon if present
2. Extract the substring after SET: "SET key=value" -> "key=value"
3. Replace all whitespaces from this substring
4. Then split the substring by the equals sign: "key=value" -> ["key", "value"]
5. If the resulting array length is not equal to two or the extracted key is empty, return directly
6. Otherwise, return the extracted key and value (value is allowed to be empty)
*/
func parseSetStatement(statement string) (string, string, error) {
	statement = removeStatementTerminator(statement)

	indexOfSet := strings.Index(strings.ToUpper(statement), configOpSet)
	if indexOfSet == -1 {
		return "", "", &types.StatementError{Msg: "Error: Invalid syntax for SET. Usage example: SET key=value."}
	}
	startOfStrAfterSet := indexOfSet + len(configOpSet)
	// This is the case when the statement is simply "SET", which is used to display current config.
	if startOfStrAfterSet >= len(statement) {
		return "", "", nil
	}
	strAfterSet := statement[startOfStrAfterSet:]

	strAfterSet = removeTabNewLineAndWhitesSpaces(strAfterSet)

	// This is the case when the statement is simply "SET  " (with empty spaces), which is used to display current config.
	if strAfterSet == "" {
		return "", "", nil
	}

	if !strings.Contains(strAfterSet, "=") {
		return "", "", &types.StatementError{Msg: "Error: missing \"=\". Usage example: SET key=value."}
	}

	keyValuePair := strings.Split(strAfterSet, "=")

	if len(keyValuePair) != 2 {
		return "", "", &types.StatementError{Msg: "Error: \"=\" should only appear once. Usage example: SET key=value."}
	}

	if keyValuePair[0] != "" && keyValuePair[1] == "" {
		return "", "", &types.StatementError{Msg: "Error: Value for key not present. If you want to reset a key, use \"RESET key\"."}
	}

	if keyValuePair[0] == "" && keyValuePair[1] != "" {
		return "", "", &types.StatementError{Msg: "Error: Key not present. Usage example: SET key=value."}
	}

	if keyValuePair[0] == "" && keyValuePair[1] == "" {
		return "", "", &types.StatementError{Msg: "Error: Key and value not present. Usage example: SET key=value."}
	}

	return keyValuePair[0], keyValuePair[1], nil
}

/*
Expected statement: "USE CATALOG catalog_name" or "USE database_name"
Steps to parse:
1. Remove semicolon if present
2. Split into words
3. If resulting array length is smaller than 2 directly return
4. If word length is 2, first word is "use" and second word IS NOT "catalog", second word is the database name
5. If word length is 3, first word is "use" and second word IS "catalog", third word is the catalog name
6. Otherwise, return empty
*/
func parseUseStatement(statement string) (string, string, error) {
	statement = removeStatementTerminator(statement)
	words := strings.Fields(statement)
	if len(words) < 2 {
		return "", "", &types.StatementError{Msg: "Error: Missing database/catalog name: Usage examples: USE DB1 OR USE CATALOG METADATA."}
	}

	isFirstWordUse := strings.ToUpper(words[0]) == configOpUse
	isSecondWordCatalog := strings.ToUpper(words[1]) == configOpUseCatalog
	//handle "USE database_name" statement
	if len(words) == 2 && isFirstWordUse {
		if isSecondWordCatalog {
			// handle empty catalog name -> "USE CATALOG "
			return "", "", &types.StatementError{Msg: "Error: Missing catalog name: Usage example: USE CATALOG METADATA."}
		} else {
			return configKeyDatabase, words[1], nil
		}
	}

	//handle "USE CATALOG catalog_name" statement
	if len(words) == 3 && isFirstWordUse && isSecondWordCatalog {
		return configKeyCatalog, words[2], nil
	}

	return "", "", &types.StatementError{Msg: "Invalid syntax for USE. Usage examples: USE CATALOG my_catalog or USE my_database"}
}

/* Expected statement: "RESET pipeline.name" */
func parseResetStatement(statement string) (string, error) {
	statement = removeStatementTerminator(statement)
	words := strings.Fields(statement)
	if len(words) == 0 {
		return "", &types.StatementError{Msg: "Error: Invalid syntax for RESET. Usage example: RESET key."}
	}

	// This is the case where we reset the entire config (e.g. "RESET")
	if len(words) == 1 {
		return "", nil
	}

	if len(words) == 2 {
		isFirstWordReset := strings.ToUpper(words[0]) == configOpReset
		key := strings.ToLower(words[1])
		if isFirstWordReset {
			return key, nil
		}
	}

	if len(words) > 2 {
		return "", &types.StatementError{Msg: "Error: too many keys for RESET provided. Usage example: RESET key."}
	}

	return "", &types.StatementError{Msg: "Error: Invalid syntax for RESET. Usage example: RESET key."}
}

func processHttpErrors(resp *http.Response, err error) error {
	if err != nil {
		return &types.StatementError{Msg: "Error: " + err.Error()}
	}

	if resp != nil && resp.StatusCode >= 400 {
		if resp.StatusCode == http.StatusUnauthorized {
			return &types.StatementError{Msg: "Error: Unauthorized. Please consider running confluent login again.", HttpResponseCode: resp.StatusCode}
		}

		statementErr := v1.NewError()
		body, err := io.ReadAll(resp.Body)

		if err != nil {
			return &types.StatementError{Msg: fmt.Sprintf("Error: received error with code \"%d\" from server but could not parse it. This is not expected. Please contact support.", resp.StatusCode)}
		}

		err = json.Unmarshal(body, &statementErr)

		if err != nil || statementErr == nil || statementErr.Title == nil || statementErr.Detail == nil {
			return &types.StatementError{Msg: fmt.Sprintf("Error: received error with code \"%d\" from server but could not parse it. This is not expected. Please contact support.", resp.StatusCode)}
		}

		return &types.StatementError{Msg: statementErr.GetTitle() + ": " + statementErr.GetDetail()}

	}

	return nil
}

// Used to to help mocking answers for now - will be removed in the future
//  Or replaced with a call to a /validate endpoint
func startsWithValidSQL(statement string) bool {
	if statement == "" {
		return false
	}

	words := strings.Fields(statement)
	firstWord := strings.ToUpper(words[0])
	_, exists := config.SQLKeywords[firstWord]

	return exists
}

// Removes leading, trailling spaces, and semicolon from end, if present
func removeStatementTerminator(s string) string {
	for strings.HasSuffix(s, configStatementTerminator) {
		s = strings.TrimSuffix(s, configStatementTerminator)
	}
	return s
}

// Removes spaces, tabs and newlines
func removeTabNewLineAndWhitesSpaces(str string) string {
	replacer := strings.NewReplacer(" ", "", "\t", "", "\n", "", "\r\n", "")
	return replacer.Replace(str)
}

func statementStartsWithOp(statement string, op string) bool {
	cleanedStatement := strings.ToUpper(statement)
	pattern := fmt.Sprintf("^%s\\b", op)
	startsWithOp, _ := regexp.MatchString(pattern, cleanedStatement)
	return startsWithOp
}

func parseStatementType(statement string) StatementType {
	if statementStartsWithOp(statement, string(SET_STATEMENT)) {
		return SET_STATEMENT
	} else if statementStartsWithOp(statement, string(USE_STATEMENT)) {
		return USE_STATEMENT
	} else if statementStartsWithOp(statement, string(RESET_STATEMENT)) {
		return RESET_STATEMENT
	} else if statementStartsWithOp(statement, string(EXIT_STATEMENT)) {
		return EXIT_STATEMENT
	} else {
		return OTHER_STATEMENT
	}
}
