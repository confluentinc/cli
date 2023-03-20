package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	v1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink-gateway/v1alpha1"

	"github.com/confluentinc/flink-sql-client/config"
)

type StatementError struct {
	Msg string
}

func (e *StatementError) Error() string {
	return e.Msg
}

// Custom Internal type that shall be used internally by the client
type StatementResult struct {
	Message string     `json:"message"` // Shown at the top before the table
	Status  PHASE      `json:"status"`
	Columns []string   `json:"columns"`
	Rows    [][]string `json:"rows"`
	//ResultsSuffix string     `json:"ResultsSuffix"` // Property that probably will be necessary in some cases Message shown after the table. i.e: "Press enter for next page..."
}

type PHASE string

const (
	PENDING   PHASE = "PENDING"   // Results are not available yet
	RUNNING   PHASE = "RUNNING"   // More results are available (pagination)
	COMPLETED PHASE = "COMPLETED" // All results were fetched
	DELETING  PHASE = "DELETING"
	FAILED    PHASE = "FAILED"
)

type StatementType string

const (
	SET_STATEMENT   StatementType = configOpSet
	USE_STATEMENT   StatementType = configOpUse
	RESET_STATEMENT StatementType = configOpReset
	OTHER_STATEMENT StatementType = "OTHER"
)

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
func parseSETStatement(statement string) (string, string, error) {
	statement = removeStatementTerminator(statement)

	indexOfSet := strings.Index(strings.ToUpper(statement), configOpSet)
	if indexOfSet == -1 {
		return "", "", &StatementError{"Error: Invalid syntax for SET. Usage example: SET key=value."}
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
		return "", "", &StatementError{"Error: missing \"=\". Usage example: SET key=value."}
	}

	keyValuePair := strings.Split(strAfterSet, "=")

	if len(keyValuePair) != 2 {
		return "", "", &StatementError{"Error: \"=\" should only appear once. Usage example: SET key=value."}
	}

	if keyValuePair[0] != "" && keyValuePair[1] == "" {
		return "", "", &StatementError{"Error: Value for key not present. If you want to reset a key, use \"RESET key\"."}
	}

	if keyValuePair[0] == "" && keyValuePair[1] != "" {
		return "", "", &StatementError{"Error: Key not present. Usage example: SET key=value."}
	}

	if keyValuePair[0] == "" && keyValuePair[1] == "" {
		return "", "", &StatementError{"Error: Key and value not present. Usage example: SET key=value."}
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
func parseUSEStatement(statement string) (string, string, error) {
	statement = removeStatementTerminator(statement)
	words := strings.Fields(statement)
	if len(words) < 2 {
		return "", "", &StatementError{"Error: Missing database/catalog name: Usage examples: USE DB1 OR USE CATALOG METADATA."}
	}

	isFirstWordUse := strings.ToUpper(words[0]) == configOpUse
	isSecondWordCatalog := strings.ToUpper(words[1]) == configOpUseCatalog
	//handle "USE database_name" statement
	if len(words) == 2 && isFirstWordUse {
		if isSecondWordCatalog {
			// handle empty catalog name -> "USE CATALOG "
			return "", "", &StatementError{"Error: Missing catalog name: Usage example: USE CATALOG METADATA."}
		} else {
			return configKeyDatabase, words[1], nil
		}
	}

	//handle "USE CATALOG catalog_name" statement
	if len(words) == 3 && isFirstWordUse && isSecondWordCatalog {
		return configKeyCatalog, words[2], nil
	}

	return "", "", &StatementError{"Invalid syntax for USE. Usage examples: USE CATALOG my_catalog or USE my_database"}
}

func processHttpErrors(resp *http.Response, err error) error {
	if resp.StatusCode != http.StatusAccepted {

		if resp.StatusCode == http.StatusUnauthorized {
			return &StatementError{"Error: Unauthorized. Please consider running confluent login again."}
		}

		statementErr := v1.NewError()
		body, err := io.ReadAll(resp.Body)

		if err != nil {
			return &StatementError{fmt.Sprintf("Error: received error with code \"%d\" from server but could not parse it. This is not expected. Please contact support.", resp.StatusCode)}
		}

		err = json.Unmarshal(body, &statementErr)

		if err != nil || statementErr == nil || statementErr.Error.Title == nil || statementErr.Error.Detail == nil {
			return &StatementError{fmt.Sprintf("Error: received error with code \"%d\" from server but could not parse it. This is not expected. Please contact support.", resp.StatusCode)}
		}

		return &StatementError{*statementErr.Error.Title + ": " + *statementErr.Error.Detail}

	}

	if err != nil {
		return &StatementError{err.Error()}
	}

	return nil
}

// Used to to help mocking answers for now - will be removed in the future
func startsWithValidSQL(statement string) bool {
	if statement == "" {
		return false
	}

	words := strings.Fields(statement)
	firstWord := strings.ToUpper(words[0])
	_, exists := config.SQLKeywords[firstWord]

	return exists
}

// Removes semicolon from end if it is present while ignoring whitespaces
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
	cleanedStatement := strings.TrimSpace(strings.ToUpper(statement))
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
	} else {
		return OTHER_STATEMENT
	}
}
