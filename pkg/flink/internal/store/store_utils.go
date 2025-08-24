package store

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/texttheater/golang-levenshtein/levenshtein"

	"github.com/confluentinc/cli/v4/pkg/flink/config"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
)

type StatementType string

const (
	SetStatement   StatementType = config.OpSet
	UseStatement   StatementType = config.OpUse
	ResetStatement StatementType = config.OpReset
	ExitStatement  StatementType = config.OpExit
	QuitStatement  StatementType = config.OpQuit
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

func processSetStatement(properties types.UserPropertiesInterface, statement string) (*types.ProcessedStatement, *types.StatementError) {
	configKey, configVal, err := parseSetStatement(statement)
	if err != nil {
		return nil, err.(*types.StatementError)
	}
	if configKey == "" {
		return &types.ProcessedStatement{
			Kind:             config.OpSet,
			Status:           types.COMPLETED,
			Statement:        statement,
			StatementResults: createStatementResults([]string{"Key", "Value"}, properties.ToSortedSlice(true)),
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

	if configKey == config.KeyOutputFormat {
		outputFormat := config.OutputFormat(configVal)
		if outputFormat != config.OutputFormatStandard && outputFormat != config.OutputFormatPlainText {
			return nil, &types.StatementError{
				Message:    fmt.Sprintf(`invalid output format for "%s"`, config.KeyOutputFormat),
				Suggestion: fmt.Sprintf(`please provide a valid output format: "%s" or "%s"`, config.OutputFormatStandard, config.OutputFormatPlainText),
			}
		}
	}

	properties.Set(configKey, configVal)
	return &types.ProcessedStatement{
		Kind:                 config.OpSet,
		StatusDetail:         "configuration updated successfully",
		Status:               types.COMPLETED,
		Statement:            statement,
		StatementResults:     createStatementResults([]string{"Key", "Value"}, [][]string{{configKey, configVal}}),
		IsLocalStatement:     true,
		IsSensitiveStatement: hasSensitiveKey(configKey),
	}, nil
}

func processResetStatement(properties types.UserPropertiesInterface, statement string) (*types.ProcessedStatement, *types.StatementError) {
	configKey, err := parseResetStatement(statement)
	if err != nil {
		return nil, &types.StatementError{Message: err.Error()}
	}
	if configKey == "" {
		properties.Clear()
		return &types.ProcessedStatement{
			Kind:             config.OpReset,
			StatusDetail:     "configuration has been reset successfully",
			Status:           types.COMPLETED,
			Statement:        statement,
			StatementResults: createStatementResults([]string{"Key", "Value"}, properties.ToSortedSlice(true)),
			IsLocalStatement: true,
		}, nil
	}
	if !properties.HasKey(configKey) {
		return nil, &types.StatementError{Message: fmt.Sprintf(`configuration key "%s" is not set`, configKey)}
	}
	// if catalog is reset, also reset the database
	if configKey == config.KeyCatalog {
		properties.Delete(config.KeyDatabase)
	}

	properties.Delete(configKey)
	return &types.ProcessedStatement{
		Kind:             config.OpReset,
		StatusDetail:     fmt.Sprintf(`configuration key "%s" has been reset successfully`, configKey),
		Status:           types.COMPLETED,
		Statement:        statement,
		StatementResults: createStatementResults([]string{"Key", "Value"}, properties.ToSortedSlice(true)),
		IsLocalStatement: true,
	}, nil
}

func processUseStatement(properties types.UserPropertiesInterface, statement string) (*types.ProcessedStatement, *types.StatementError) {
	catalog, database, err := parseUseStatement(statement)
	if err != nil {
		return nil, &types.StatementError{Message: err.Error()}
	}
	addedConfig := [][]string{}

	// "USE CATALOG catalog_name" statement
	if catalog != "" && database == "" {
		// USE CATALOG <catalog> will remove the current database
		properties.Delete(config.KeyDatabase)

		properties.Set(config.KeyCatalog, catalog)
		addedConfig = append(addedConfig, []string{config.KeyCatalog, catalog})

		// "USE database" statement
	} else if catalog == "" && database != "" {
		// require catalog to be set before running USE <database>
		if !properties.HasKey(config.KeyCatalog) {
			return nil, &types.StatementError{
				Message:    "no catalog was set",
				Suggestion: `please set a catalog first with "USE CATALOG catalog-name" or  before setting a database`,
			}
		}

		properties.Set(config.KeyDatabase, database)
		addedConfig = append(addedConfig, []string{config.KeyDatabase, database})

		// "USE `catalog_name`.`database_name`" statement
	} else if catalog != "" && database != "" {
		properties.Set(config.KeyCatalog, catalog)
		properties.Set(config.KeyDatabase, database)
		addedConfig = append(addedConfig, []string{config.KeyCatalog, catalog})
		addedConfig = append(addedConfig, []string{config.KeyDatabase, database})
	} else {
		return nil, useError()
	}

	return &types.ProcessedStatement{
		Kind:             config.OpUse,
		StatusDetail:     "configuration updated successfully",
		Status:           types.COMPLETED,
		Statement:        statement,
		StatementResults: createStatementResults([]string{"Key", "Value"}, addedConfig),
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

	// Trim whitespace
	strAfterSet = strings.TrimSpace(strAfterSet)

	// Checks to give helpful suggestions to the users.
	// Contains only an equal sign
	if strAfterSet == "=" {
		return "", "", &types.StatementError{
			Message: "key and value not present",
			Usage:   []string{"SET 'key'='value'"},
		}
	}

	// Check that the string doesn't end in an equal sign (+ optional whitespace),
	// and suggest using RESET to remove a key
	if strings.HasSuffix(strAfterSet, "=") {
		return "", "", &types.StatementError{
			Message:    "value for key not present",
			Suggestion: `if you want to reset a key, use "RESET 'key'"`,
		}
	}

	// Check that it doesn't begin with an equal sign (+ optional whitespace),
	if strings.HasPrefix(strAfterSet, "=") {
		return "", "", &types.StatementError{
			Message: "key not present",
			Usage:   []string{"SET 'key'='value'"},
		}
	}

	// The key and value must be enclosed by single quotes
	checkquotes := regexp.MustCompile(`^'.+'\s*=\s*'.*'$`)
	if !checkquotes.MatchString(strAfterSet) {
		return "", "", &types.StatementError{
			Message: "key and value must be enclosed by single quotes (')",
			Usage:   []string{"SET 'key'='value'"},
		}
	}

	// The actual final regex for parsing the key and value
	// The only possible error left is an unescaped single quote
	re := regexp.MustCompile(`^'(([^']|'')*?)'\s*=\s*'(([^']|'')*?)'$`)
	matches := re.FindStringSubmatch(strAfterSet)
	if len(matches) != 5 {
		return "", "", &types.StatementError{
			Message:    "key or value contains unescaped single quotes (')",
			Usage:      []string{"SET 'key'='value'"},
			Suggestion: `please escape all single quotes with another single quote "''key''"`,
		}
	}

	key := matches[1]
	value := matches[3]

	// replace escaped quotes
	key = strings.ReplaceAll(key, "''", "'")
	value = strings.ReplaceAll(value, "''", "'")
	return key, value, nil
}

func TokenizeSQL(statement string) []string {
	var tokens []string
	var buffer bytes.Buffer
	var inBacktick bool
	input := []rune(statement)

	// Iterate over each character in the input string
	for i := 0; i < len(input); i++ {
		c := input[i]

		// Ignore whitespace
		if unicode.IsSpace(c) && !inBacktick {
			tokens = appendToTokens(tokens, &buffer)
			continue
		}

		// Dot is a separator
		if input[i] == '.' && !inBacktick {
			tokens = appendToTokens(tokens, &buffer)
			tokens = append(tokens, ".")
			continue
		}

		// Handle backticks
		if c == '`' {
			if inBacktick {
				// escaped backtick
				if i+1 < len(input) && input[i+1] == '`' {
					i++
					buffer.WriteRune(c)
					continue
				}

				// End of backtick
				tokens = append(tokens, buffer.String())
				buffer.Reset()
				inBacktick = false
			} else {
				// Start of backtick
				tokens = appendToTokens(tokens, &buffer)
				inBacktick = true
			}
			continue
		}

		buffer.WriteRune(c)
	}

	// Add last token if in backtick
	if inBacktick {
		tokens = append(tokens, buffer.String())
	} else if buffer.Len() > 0 {
		tokens = append(tokens, buffer.String())
	}

	if len(tokens) == 0 {
		return []string{}
	}

	return tokens
}

func appendToTokens(tokens []string, buffer *bytes.Buffer) []string {
	if str := buffer.String(); len(str) > 0 {
		tokens = append(tokens, str)
		buffer.Reset()
	}
	return tokens
}

/*
Expected statement: "USE CATALOG `catalog_name`" or "USE `database_name` or "USE `catalog_name`.`database_name`"
Returns the catalog and database extracted if the present, otherwise returns an error
*/
func parseUseStatement(statement string) (string, string, error) {
	statement = removeStatementTerminator(statement)
	tokens := TokenizeSQL(statement)
	if len(tokens) < 2 {
		return "", "", useError()
	}

	isFirstWordUse := strings.ToUpper(tokens[0]) == config.OpUse
	isSecondWordCatalog := strings.ToUpper(tokens[1]) == config.OpUseCatalog

	// handle "USE CATALOG catalog_name" statement
	if isFirstWordUse && isSecondWordCatalog {
		if len(tokens) == 3 {
			return tokens[2], "", nil
		} else {
			return "", "", &types.StatementError{
				Message: "invalid syntax for USE CATALOG",
				Usage:   []string{"USE CATALOG `my_catalog`"},
			}
		}
	}

	if isFirstWordUse {
		switch len(tokens) {
		// handle "USE database_name" statement
		case 2:
			return "", tokens[1], nil
		// handle "USE `catalog_name`.`database_name`" statement
		case 4:
			if tokens[2] == "." {
				return tokens[1], tokens[3], nil
			}
		default:
			return "", "", useError()
		}
	}

	return "", "", useError()
}

func useError() *types.StatementError {
	return &types.StatementError{
		Message: "invalid syntax for USE",
		Usage:   []string{"USE CATALOG `my_catalog`", "USE `my_database`", "USE `my_catalog`.`my_database`"},
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

func hasSensitiveKey(key string) bool {
	secretsPrefix := getSubstringUpToSecondDot(key)
	if secretsPrefix == "" {
		return false
	}

	secretsPrefix = strings.ToLower(secretsPrefix)
	distance := levenshtein.DistanceForStrings([]rune(config.KeySqlSecrets), []rune(secretsPrefix), levenshtein.DefaultOptions)

	return distance <= 2
}

func getSubstringUpToSecondDot(s string) string {
	firstDot := strings.Index(s, ".")
	if firstDot == -1 {
		return ""
	}
	if len(s) <= firstDot+2 {
		return ""
	}

	secondDot := strings.Index(s[firstDot+1:], ".")
	if secondDot == -1 {
		return ""
	}
	secondDot += firstDot + 1

	return s[:secondDot+1]
}

// Removes leading, trailling spaces, and semicolon from end, if present
func removeStatementTerminator(s string) string {
	for strings.HasSuffix(s, config.StatementTerminator) {
		s = strings.TrimSuffix(s, config.StatementTerminator)
	}
	return s
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
	} else if statementStartsWithOp(statement, string(QuitStatement)) {
		return QuitStatement
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
func getTimeout(properties types.UserPropertiesInterface) time.Duration {
	if properties.HasKey(config.KeyResultsTimeout) {
		timeoutInMilliseconds, err := strconv.Atoi(properties.Get(config.KeyResultsTimeout))
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
