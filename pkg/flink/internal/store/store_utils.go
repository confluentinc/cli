package store

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	flinkgatewayv1beta1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1beta1"

	"github.com/confluentinc/cli/v3/pkg/flink/config"
	"github.com/confluentinc/cli/v3/pkg/flink/types"
)

type StatementType string

const (
	SetStatement   StatementType = config.OpSet
	UseStatement   StatementType = config.OpUse
	ResetStatement StatementType = config.OpReset
	ExitStatement  StatementType = config.OpExit
	OtherStatement StatementType = "OTHER"
)

func createStatementResults(columnNames []string, rows [][]string) *types.StatementResults {
	statementResultRows := make([]types.StatementResultRow, len(rows))
	for idx, row := range rows {
		var statementResultRow types.StatementResultRow
		for _, field := range row {
			statementResultRow.Fields = append(statementResultRow.Fields, types.AtomicStatementResultField{
				Type:  types.Varchar,
				Value: field,
			})
		}
		statementResultRows[idx] = statementResultRow
	}

	return &types.StatementResults{
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
		return &types.ProcessedStatement{
			Kind:             config.OpSet,
			Status:           types.COMPLETED,
			StatementResults: createStatementResults([]string{"Key", "Value"}, s.Properties.ToSortedSlice(true)),
			IsLocalStatement: true,
		}, nil
	}
	if configKey == config.KeyDatabase || configKey == config.KeyCatalog {
		return nil, &types.StatementError{
			Message:    "cannot set a catalog or a database with SET command",
			Suggestion: `please set a catalog with "USE CATALOG catalog-name" and a database with "USE db-name"`,
		}
	}
	if configKey == config.KeyStatementName && strings.TrimSpace(configVal) == "" {
		return nil, &types.StatementError{
			Message:    "cannot set an empty statement name",
			Suggestion: `please provide a non-empty statement name with "SET 'client.statement-name'='non-empty-name'"`,
		}
	}
	hasUserSecret := isUserSecretKey(configKey)

	s.Properties.Set(configKey, configVal)

	return &types.ProcessedStatement{
		Kind:                 config.OpSet,
		StatusDetail:         "configuration updated successfully",
		Status:               types.COMPLETED,
		StatementResults:     createStatementResults([]string{"Key", "Value"}, [][]string{{configKey, configVal}}),
		IsLocalStatement:     true,
		IsSensitiveStatement: hasUserSecret,
	}, nil
}

func (s *Store) processResetStatement(statement string) (*types.ProcessedStatement, *types.StatementError) {
	configKey, err := parseResetStatement(statement)
	if err != nil {
		return nil, &types.StatementError{Message: err.Error()}
	}
	if configKey == "" {
		s.Properties.Clear()
		return &types.ProcessedStatement{
			Kind:             config.OpReset,
			StatusDetail:     "configuration has been reset successfully",
			Status:           types.COMPLETED,
			StatementResults: createStatementResults([]string{"Key", "Value"}, s.Properties.ToSortedSlice(true)),
			IsLocalStatement: true,
		}, nil
	}
	if !s.Properties.HasKey(configKey) {
		return nil, &types.StatementError{Message: fmt.Sprintf(`configuration key "%s" is not set`, configKey)}
	}
	// if catalog is reset, also reset the database
	if configKey == config.KeyCatalog {
		s.Properties.Delete(config.KeyDatabase)
	}

	s.Properties.Delete(configKey)
	return &types.ProcessedStatement{
		Kind:             config.OpReset,
		StatusDetail:     fmt.Sprintf(`configuration key "%s" has been reset successfully`, configKey),
		Status:           types.COMPLETED,
		StatementResults: createStatementResults([]string{"Key", "Value"}, s.Properties.ToSortedSlice(true)),
		IsLocalStatement: true,
	}, nil
}

func (s *Store) processUseStatement(statement string) (*types.ProcessedStatement, *types.StatementError) {
	configKey, configVal, err := parseUseStatement(statement)
	if err != nil {
		return nil, &types.StatementError{Message: err.Error()}
	}

	// require catalog to be set before running USE <database>
	if configKey == config.KeyDatabase && !s.Properties.HasKey(config.KeyCatalog) {
		return nil, &types.StatementError{
			Message:    "no catalog was set",
			Suggestion: `please set a catalog first with "USE CATALOG catalog-name" before setting a database`,
		}
	}

	// USE CATALOG <catalog> will remove the current database
	if configKey == config.KeyCatalog {
		s.Properties.Delete(config.KeyDatabase)
	}

	s.Properties.Set(configKey, configVal)

	return &types.ProcessedStatement{
		Kind:             config.OpUse,
		StatusDetail:     "configuration updated successfully",
		Status:           types.COMPLETED,
		StatementResults: createStatementResults([]string{"Key", "Value"}, [][]string{{configKey, configVal}}),
		IsLocalStatement: true,
	}, nil
}

/*
Expected statement: "SET 'key'='value'"
Steps to parse:
1. Remove the semicolon if present
2. Extract the substring after SET: "SET 'key'='value'" -> "'key'='value'"
3. Replace all whitespaces from this substring
4. Then split the substring by the equals sign: "'key'='value'" -> ["'key'", "'value'"]
5. If the resulting array length is not equal to two or the extracted key is empty, return directly
6. If key and value are not enclosed by single quotes, return error
7. Otherwise, return the extracted key and value (value is allowed to be empty)
*/
func parseSetStatement(statement string) (string, string, error) {
	statement = removeStatementTerminator(statement)

	indexOfSet := strings.Index(strings.ToUpper(statement), config.OpSet)
	if indexOfSet == -1 {
		return "", "", &types.StatementError{
			Message: "invalid syntax for SET",
			Usage:   []string{"SET 'key'='value'"},
		}
	}
	startOfStrAfterSet := indexOfSet + len(config.OpSet)
	// This is the case when the statement is simply "SET", which is used to display current config.
	if startOfStrAfterSet >= len(statement) {
		return "", "", nil
	}
	strAfterSet := strings.TrimSpace(statement[startOfStrAfterSet:])

	// This is the case when the statement is simply "SET  " (with empty spaces), which is used to display current config.
	if strAfterSet == "" {
		return "", "", nil
	}

	if !strings.Contains(strAfterSet, "=") {
		return "", "", &types.StatementError{
			Message: `missing "="`,
			Usage:   []string{"SET 'key'='value'"},
		}
	}

	keyValuePair := strings.Split(strAfterSet, "=")

	if len(keyValuePair) != 2 {
		return "", "", &types.StatementError{
			Message: `"=" should only appear once`,
			Usage:   []string{"SET 'key'='value'"},
		}
	}

	keyWithQuotes := strings.TrimSpace(keyValuePair[0])
	valueWithQuotes := strings.TrimSpace(keyValuePair[1])

	if keyWithQuotes != "" && valueWithQuotes == "" {
		return "", "", &types.StatementError{
			Message:    "value for key not present",
			Suggestion: `if you want to reset a key, use "RESET 'key'"`,
		}
	}

	if keyWithQuotes == "" && valueWithQuotes != "" {
		return "", "", &types.StatementError{
			Message: "key not present",
			Usage:   []string{"SET 'key'='value'"},
		}
	}

	if keyWithQuotes == "" && valueWithQuotes == "" {
		return "", "", &types.StatementError{
			Message: "key and value not present",
			Usage:   []string{"SET 'key'='value'"},
		}
	}

	if !strings.HasPrefix(keyWithQuotes, "'") || !strings.HasSuffix(keyWithQuotes, "'") ||
		!strings.HasPrefix(valueWithQuotes, "'") || !strings.HasSuffix(valueWithQuotes, "'") {
		return "", "", &types.StatementError{
			Message: "key and value must be enclosed by single quotes (')",
			Usage:   []string{"SET 'key'='value'"},
		}
	}

	// remove enclosing quotes
	keyWithQuotes = keyWithQuotes[1 : len(keyWithQuotes)-1]
	valueWithQuotes = valueWithQuotes[1 : len(valueWithQuotes)-1]

	if containsUnescapedSingleQuote(keyWithQuotes) {
		return "", "", &types.StatementError{
			Message:    "key contains unescaped single quotes (')",
			Usage:      []string{"SET 'key'='value'"},
			Suggestion: `please escape all single quotes with another single quote "''key''"`,
		}
	}

	if containsUnescapedSingleQuote(valueWithQuotes) {
		return "", "", &types.StatementError{
			Message:    "value contains unescaped single quotes (')",
			Usage:      []string{"SET 'key'='value'"},
			Suggestion: `please escape all single quotes with another single quote "''key''"`,
		}
	}

	// replace escaped quotes
	key := strings.ReplaceAll(keyWithQuotes, "''", "'")
	value := strings.ReplaceAll(valueWithQuotes, "''", "'")
	return key, value, nil
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
		return "", "", &types.StatementError{
			Message: "missing database/catalog name",
			Usage:   []string{"USE CATALOG my_catalog", "USE my_database"},
		}
	}

	isFirstWordUse := strings.ToUpper(words[0]) == config.OpUse
	isSecondWordCatalog := strings.ToUpper(words[1]) == config.OpUseCatalog
	// handle "USE database_name" statement
	if len(words) == 2 && isFirstWordUse {
		if isSecondWordCatalog {
			// handle empty catalog name -> "USE CATALOG "
			return "", "", &types.StatementError{
				Message: "missing catalog name",
				Usage:   []string{"USE CATALOG my_catalog"},
			}
		} else {
			return config.KeyDatabase, words[1], nil
		}
	}

	// handle "USE CATALOG catalog_name" statement
	if len(words) == 3 && isFirstWordUse && isSecondWordCatalog {
		return config.KeyCatalog, words[2], nil
	}

	return "", "", &types.StatementError{
		Message: "invalid syntax for USE",
		Usage:   []string{"USE CATALOG my_catalog", "USE my_database"},
	}
}

/* Expected statement: "RESET 'pipeline.name'" */
func parseResetStatement(statement string) (string, error) {
	statement = removeStatementTerminator(statement)

	indexOfReset := strings.Index(strings.ToUpper(statement), config.OpReset)
	if indexOfReset == -1 {
		return "", &types.StatementError{
			Message: "invalid syntax for RESET",
			Usage:   []string{"RESET 'key'"},
		}
	}
	startOfStrAfterReset := indexOfReset + len(config.OpReset)
	// This is the case where we reset the entire config (e.g. "RESET")
	if startOfStrAfterReset >= len(statement) {
		return "", nil
	}
	strAfterReset := strings.TrimSpace(statement[startOfStrAfterReset:])

	// This is the case when the statement is simply "RESET  " (with empty spaces), where we reset the entire config
	if strAfterReset == "" {
		return "", nil
	}

	if !strings.HasPrefix(strAfterReset, "'") || !strings.HasSuffix(strAfterReset, "'") {
		return "", &types.StatementError{
			Message: "invalid syntax for RESET, key must be enclosed by single quotes ''",
			Usage:   []string{"RESET 'key'"},
		}
	}

	// remove enclosing quotes
	strAfterReset = strAfterReset[1 : len(strAfterReset)-1]

	if containsUnescapedSingleQuote(strAfterReset) {
		return "", &types.StatementError{
			Message:    "key contains unescaped single quotes (')",
			Usage:      []string{"RESET 'key'"},
			Suggestion: `please escape all single quotes with another single quote "''key''"`,
		}
	}

	// replace escaped quotes
	key := strings.ReplaceAll(strAfterReset, "''", "'")
	return key, nil
}

func processHttpErrors(resp *http.Response, err error) error {
	if err != nil {
		return &types.StatementError{Message: err.Error()}
	}

	if resp != nil && resp.StatusCode >= 400 {
		if resp.StatusCode == http.StatusUnauthorized {
			return &types.StatementError{
				Message:    "unauthorized",
				Suggestion: `Please run "confluent login"`,
				StatusCode: resp.StatusCode,
			}
		}

		statementErr := flinkgatewayv1beta1.NewError()
		body, err := io.ReadAll(resp.Body)

		if err != nil {
			return &types.StatementError{Message: fmt.Sprintf(`received error with code "%d" from server but could not parse it. This is not expected. Please contact support`, resp.StatusCode)}
		}

		err = json.Unmarshal(body, &statementErr)

		if err != nil || statementErr == nil || statementErr.Title == nil || statementErr.Detail == nil {
			return &types.StatementError{Message: fmt.Sprintf(`received error with code "%d" from server but could not parse it. This is not expected. Please contact support`, resp.StatusCode)}
		}

		return &types.StatementError{Message: fmt.Sprintf("%s: %s", statementErr.GetTitle(), statementErr.GetDetail())}
	}

	return nil
}

// Used to help mocking answers for now - will be removed in the future
//  Or replaced with a call to a /validate endpoint
func startsWithValidSQL(statement string) bool {
	if statement == "" {
		return false
	}

	words := strings.Fields(statement)
	firstWord := strings.ToUpper(words[0])
	return config.SQLKeywords.Contains(firstWord)
}

func isUserSecretKey(key string) bool {
	parts := strings.Split(key, ".")
	count := len(parts)

	if count < 1 {
		return false
	}

	last := strings.ToLower(parts[count-1])

	// add *ecrets, s*crets and so on. Maybe not * but only two letter that are wrong after the s? hmm
	matches, err := regexp.MatchString(".*?ecret|s.{0,2}?cret|se.{0,2}?ret|sec.{0,2}?et|secr.{0,2}?t|secre.{0,2}?", last)
	if matches && err == nil {
		return true
	}

	return false
}

// Removes leading, trailling spaces, and semicolon from end, if present
func removeStatementTerminator(s string) string {
	for strings.HasSuffix(s, config.StatementTerminator) {
		s = strings.TrimSuffix(s, config.StatementTerminator)
	}
	return s
}

// Removes spaces, tabs and newlines
func removeTabNewLineAndWhitesSpaces(str string) string {
	replacer := strings.NewReplacer(" ", "", "\t", "", "\n", "", "\r\n", "")
	return replacer.Replace(str)
}

func statementStartsWithOp(statement, op string) bool {
	cleanedStatement := strings.ToUpper(statement)
	pattern := fmt.Sprintf("^%s\\b", op)
	startsWithOp, _ := regexp.MatchString(pattern, cleanedStatement)
	return startsWithOp
}

func parseStatementType(statement string) StatementType {
	if statementStartsWithOp(statement, string(SetStatement)) {
		return SetStatement
	} else if statementStartsWithOp(statement, string(UseStatement)) {
		return UseStatement
	} else if statementStartsWithOp(statement, string(ResetStatement)) {
		return ResetStatement
	} else if statementStartsWithOp(statement, string(ExitStatement)) {
		return ExitStatement
	} else {
		return OtherStatement
	}
}

// This returns the local timezone as a custom timezone along with the offset to UTC/GMT
// Example: GMT+02:00 or GMT-08:00
func getLocalTimezone() string {
	_, offsetSeconds := time.Now().Zone()
	return formatUTCOffsetToTimezone(offsetSeconds)
}

func formatUTCOffsetToTimezone(offsetSeconds int) string {
	timeOffset := time.Duration(offsetSeconds) * time.Second
	sign := "+"
	if offsetSeconds < 0 {
		sign = "-"
		timeOffset *= -1
	}
	offsetStr := fmt.Sprintf("%02d:%02d", int(timeOffset.Hours()), int(timeOffset.Minutes())%60)
	return fmt.Sprintf("GMT%s%s", sign, offsetStr)
}

// This increases function calculates a wait time that starts at 300 ms and increases 300 ms every 10 retries.
// This should provide a better UX than exponential backoff. He're are two simulations in an excel sheet
// Exponential: https://docs.google.com/spreadsheets/d/14lHRcC_NGoF4KBtA_lrEivv05XYc3nNo5jaIvsHpgi0/edit?usp=sharing
// Discrete: https://docs.google.com/spreadsheets/d/1fMIOBIDbhZ6zH6bLq9iJXRs8jBLdA7beHef4vOW__tw/edit?usp=sharing
func calcWaitTime(retries int) time.Duration {
	return config.InitialWaitTime + time.Duration(config.WaitTimeIncrease*(retries/10))*time.Millisecond
}

// Function to extract timeout for waiting for results.
// We either use the value set by user using set or use a default value of 10 minutes (as of today)
func (s *Store) getTimeout() time.Duration {
	if s.Properties.HasKey(config.KeyResultsTimeout) {
		timeoutInMilliseconds, err := strconv.Atoi(s.Properties.Get(config.KeyResultsTimeout))
		if err == nil {
			// TODO - check for error when setting the property so user knows he hasn't set the results-timeout property properly
			return time.Duration(timeoutInMilliseconds) * time.Millisecond
		} else {
			return config.DefaultTimeoutDuration
		}
	} else {
		return config.DefaultTimeoutDuration
	}
}

func containsUnescapedSingleQuote(str string) bool {
	// remove escaped quotes and check if there are still single quotes
	str = strings.ReplaceAll(str, "''", "")
	return strings.Contains(str, "'")
}
