package store

import (
	"fmt"
	"testing"

	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	"github.com/confluentinc/cli/v4/pkg/flink/config"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
)

func TestRemoveStatementTerminator(t *testing.T) {
	type args struct {
		statement string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "removeStatementTerminator() removes one terminator", args: args{statement: "SELECT * FROM table;"}, want: "SELECT * FROM table"},
		{name: "removeStatementTerminator() removes no terminator", args: args{statement: "SELECT * FROM table"}, want: "SELECT * FROM table"},
		{name: "removeStatementTerminator() removes multiple terminators", args: args{statement: "SELECT * FROM table;;;"}, want: "SELECT * FROM table"},
		{name: "removeStatementTerminator() doesn't remove terminators in between", args: args{statement: "SELECT * FROM table;;;wasas"}, want: "SELECT * FROM table;;;wasas"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := removeStatementTerminator(test.args.statement); got != test.want {
				require.Equal(t, test.want, got)
			}
		})
	}
}

func TestProcessSetStatement(t *testing.T) {
	// Create a new store
	client := ccloudv2.NewFlinkGatewayClient("url", "userAgent", false, "authToken")
	appOptions := &types.ApplicationOptions{
		Cloud:           true,
		EnvironmentName: "env-123",
	}
	userProperties := NewUserProperties(appOptions)
	s := NewStore(client, nil, userProperties, &types.ApplicationOptions{EnvironmentName: "env-123"}, tokenRefreshFunc).(*Store)
	// This is just a string, so really doesn't matter
	s.Properties.Set(config.KeyLocalTimeZone, "London/GMT")

	t.Run("should return an error message if statement is invalid", func(t *testing.T) {
		_, err := processSetStatement(s.Properties, "se key=value")
		assert.NotNil(t, err)
	})

	t.Run("should return all keys and values from config if configKey is empty", func(t *testing.T) {
		result, err := processSetStatement(s.Properties, "set")
		assert.Nil(t, err)
		assert.EqualValues(t, types.COMPLETED, result.Status)

		assert.Equal(t, 2, len(result.StatementResults.Headers))
		assert.Equal(t, "set", result.Statement)
		assert.Equal(t, len(s.Properties.GetProperties()), len(result.StatementResults.Rows))
		cupaloy.SnapshotT(t, result.StatementResults)
	})

	t.Run("should update config for valid configKey", func(t *testing.T) {
		result, err := processSetStatement(s.Properties, "set 'location'='USA'")
		assert.Nil(t, err)
		assert.EqualValues(t, types.COMPLETED, result.Status)
		assert.Equal(t, "set 'location'='USA'", result.Statement)
		assert.Equal(t, "configuration updated successfully", result.StatusDetail)
		cupaloy.SnapshotT(t, result.StatementResults)
	})

	t.Run("should return all keys and values from config if configKey is empty after updates", func(t *testing.T) {
		result, err := processSetStatement(s.Properties, "set")
		assert.Nil(t, err)
		assert.EqualValues(t, types.COMPLETED, result.Status)

		assert.Equal(t, 2, len(result.StatementResults.Headers))
		assert.Equal(t, "set", result.Statement)
		assert.Equal(t, len(s.Properties.GetProperties()), len(result.StatementResults.Rows))
		cupaloy.SnapshotT(t, result.StatementResults)
	})

	t.Run("should fail if user wants to set the catalog", func(t *testing.T) {
		_, err := processSetStatement(s.Properties, fmt.Sprintf("set '%s'='%s'", config.KeyCatalog, "catalog-name"))
		assert.Equal(t, &types.StatementError{
			Message:    "cannot set a catalog or a database with SET command",
			Suggestion: `please set a catalog with "USE CATALOG catalog-name" and a database with "USE db-name"`,
		}, err)
	})

	t.Run("should fail if user wants to set the database", func(t *testing.T) {
		_, err := processSetStatement(s.Properties, fmt.Sprintf("set '%s'='%s'", config.KeyDatabase, "db-name"))
		assert.Equal(t, &types.StatementError{
			Message:    "cannot set a catalog or a database with SET command",
			Suggestion: `please set a catalog with "USE CATALOG catalog-name" and a database with "USE db-name"`,
		}, err)
	})

	t.Run("should fail if user wants to set an empty statement name", func(t *testing.T) {
		_, err := processSetStatement(s.Properties, fmt.Sprintf("set '%s'='%s'", config.KeyStatementName, ""))
		assert.Equal(t, &types.StatementError{
			Message:    "cannot set an empty statement name",
			Suggestion: `please provide a non-empty statement name with "SET 'client.statement-name'='non-empty-name'"`,
		}, err)
	})

	t.Run("should parse and identify sensitive set statement", func(t *testing.T) {
		result, err := processSetStatement(s.Properties, "set 'sql.secrets.openai' = 'mysecret'")
		assert.Nil(t, err)
		assert.EqualValues(t, true, result.IsSensitiveStatement)

		result, err = processSetStatement(s.Properties, "set 'sql.secrets.opeenaai' = 'mysecret'")
		assert.Nil(t, err)
		assert.EqualValues(t, true, result.IsSensitiveStatement)
	})

	t.Run("should parse set statements with equal signs in the value", func(t *testing.T) {
		result, err := processSetStatement(s.Properties, "set 'sql.secrets.openai' = 'b64encodedABCD=='")
		assert.Nil(t, err)
		assert.EqualValues(t, true, result.IsSensitiveStatement)
	})

	t.Run("should parse set statements with equal signs in the key", func(t *testing.T) {
		result, err := processSetStatement(s.Properties, "set 'sql.secrets.openai==' = 'b64encodedABCD=='")
		assert.Nil(t, err)
		assert.EqualValues(t, true, result.IsSensitiveStatement)
	})
}

func TestProcessResetStatement(t *testing.T) {
	// Create a new store
	client := ccloudv2.NewFlinkGatewayClient("url", "userAgent", false, "authToken")
	appOptions := types.ApplicationOptions{
		Cloud:            true,
		OrganizationId:   "orgId",
		EnvironmentName:  "envName",
		Database:         "database",
		ServiceAccountId: "sa-123",
	}

	userProperties := NewUserProperties(&appOptions)
	s := NewStore(client, nil, userProperties, &appOptions, tokenRefreshFunc).(*Store)
	s.Properties.Set(config.KeyLocalTimeZone, "London/GMT")

	defaultSetOutput := createStatementResults([]string{"Key", "Value"}, [][]string{
		{config.KeyLocalTimeZone, fmt.Sprintf("%s (default)", getLocalTimezone())},
		{config.KeyServiceAccount, fmt.Sprintf("%s (default)", appOptions.ServiceAccountId)},
		{config.KeyOutputFormat, fmt.Sprintf("%s (default)", config.OutputFormatStandard)},
	})

	t.Run("should return all keys and values including default and initial values before reseting", func(t *testing.T) {
		result, err := processSetStatement(s.Properties, "set")
		assert.Nil(t, err)
		assert.EqualValues(t, types.COMPLETED, result.Status)

		assert.Equal(t, 2, len(result.StatementResults.Headers))
		assert.Equal(t, "set", result.Statement)
		assert.Equal(t, len(s.Properties.GetProperties()), len(result.StatementResults.Rows))
		cupaloy.SnapshotT(t, result.StatementResults)
	})

	t.Run("should return an error message if statement is invalid", func(t *testing.T) {
		_, err := processResetStatement(s.Properties, "res key")
		assert.NotNil(t, err)
	})

	t.Run("should reset all keys and values from config to their default or delete them if no default", func(t *testing.T) {
		s.Properties.Set(config.KeyCatalog, "job1")
		s.Properties.Set("timeout", "30")
		result, _ := processResetStatement(s.Properties, "reset")
		assert.EqualValues(t, types.COMPLETED, result.Status)
		assert.Equal(t, "configuration has been reset successfully", result.StatusDetail)
		assert.Equal(t, "reset", result.Statement)
		assert.ElementsMatch(t, defaultSetOutput.GetHeaders(), result.StatementResults.GetHeaders())
		assert.ElementsMatch(t, defaultSetOutput.GetRows(), result.StatementResults.GetRows())
	})

	t.Run("should return an error message if configKey does not exist", func(t *testing.T) {
		result, err := processResetStatement(s.Properties, "reset 'location'")
		assert.NotNil(t, err)
		assert.Equal(t, `Error: configuration key "location" is not set`, err.Error())
		assert.Nil(t, result)
	})

	t.Run("should reset config for valid configKey", func(t *testing.T) {
		s.Properties.Set("catalog", "job1")
		result, _ := processResetStatement(s.Properties, "reset 'catalog'")
		assert.EqualValues(t, types.COMPLETED, result.Status)
		assert.Equal(t, `configuration key "catalog" has been reset successfully`, result.StatusDetail)
		assert.Equal(t, "reset 'catalog'", result.Statement)
		assert.ElementsMatch(t, defaultSetOutput.GetHeaders(), result.StatementResults.GetHeaders())
		assert.ElementsMatch(t, defaultSetOutput.GetRows(), result.StatementResults.GetRows())
	})

	t.Run("should reset database if catalog is reset", func(t *testing.T) {
		s.Properties.Set(config.KeyCatalog, "catalog")
		s.Properties.Set(config.KeyDatabase, "db")
		statement := fmt.Sprintf("reset '%s'", config.KeyCatalog)
		result, err := processResetStatement(s.Properties, statement)
		assert.Nil(t, err)
		assert.Equal(t, statement, result.Statement)
		assert.False(t, s.Properties.HasKey(config.KeyCatalog))
		assert.False(t, s.Properties.HasKey(config.KeyDatabase))
	})
}

func TestProcessUseStatement(t *testing.T) {
	// Create a new store
	client := ccloudv2.NewFlinkGatewayClient("url", "userAgent", false, "authToken")
	appOptions := types.ApplicationOptions{
		OrganizationId:  "orgId",
		EnvironmentName: "envName",
		Database:        "database",
	}
	userProperties := NewUserProperties(&appOptions)
	s := NewStore(client, nil, userProperties, &appOptions, tokenRefreshFunc).(*Store)

	t.Run("should return an error message if statement is invalid", func(t *testing.T) {
		_, err := processUseStatement(s.Properties, "us")
		require.Error(t, err)
	})

	t.Run("should update the database name in config", func(t *testing.T) {
		result, err := processUseStatement(s.Properties, "use db1")
		require.Nil(t, err)
		require.Equal(t, config.OpUse, result.Kind)
		require.Equal(t, "use db1", result.Statement)
		require.EqualValues(t, types.COMPLETED, result.Status)
		require.Equal(t, "configuration updated successfully", result.StatusDetail)
		expectedResult := createStatementResults([]string{"Key", "Value"}, [][]string{{config.KeyDatabase, "db1"}})
		assert.Equal(t, expectedResult, result.StatementResults)
		assert.Equal(t, "db1", s.Properties.Get(config.KeyDatabase))
	})

	t.Run("should return an error message if catalog name is missing", func(t *testing.T) {
		_, err := processUseStatement(s.Properties, "use catalog")
		require.Error(t, err)
	})

	t.Run("should update the catalog name in config", func(t *testing.T) {
		result, err := processUseStatement(s.Properties, "use catalog metadata")
		require.Nil(t, err)
		require.Equal(t, config.OpUse, result.Kind)
		require.Equal(t, "use catalog metadata", result.Statement)
		require.EqualValues(t, types.COMPLETED, result.Status)
		require.Equal(t, "configuration updated successfully", result.StatusDetail)
		expectedResult := createStatementResults([]string{"Key", "Value"}, [][]string{{config.KeyCatalog, "metadata"}})
		assert.Equal(t, expectedResult, result.StatementResults)
		assert.Equal(t, "metadata", s.Properties.Get(config.KeyCatalog))
	})

	t.Run("should update the catalog and database name in config", func(t *testing.T) {
		result, err := processUseStatement(s.Properties, "USE `my catalog`.`my database`")
		require.Nil(t, err)
		require.Equal(t, config.OpUse, result.Kind)
		require.Equal(t, "USE `my catalog`.`my database`", result.Statement)
		require.EqualValues(t, types.COMPLETED, result.Status)
		require.Equal(t, "configuration updated successfully", result.StatusDetail)
		expectedResult := createStatementResults([]string{"Key", "Value"}, [][]string{
			{config.KeyCatalog, "my catalog"},
			{config.KeyDatabase, "my database"},
		})
		assert.Equal(t, expectedResult, result.StatementResults)
		assert.Equal(t, "my catalog", s.Properties.Get(config.KeyCatalog))
		assert.Equal(t, "my database", s.Properties.Get(config.KeyDatabase))
	})

	t.Run("should return an error message for invalid syntax", func(t *testing.T) {
		_, err := processUseStatement(s.Properties, "use db1 catalog metadata")
		require.Error(t, err)
	})

	t.Run("use database should fail if no catalog was set", func(t *testing.T) {
		// delete the catalog
		catalogName := s.Properties.Get(config.KeyCatalog)
		s.Properties.Delete(config.KeyCatalog)

		_, err := processUseStatement(s.Properties, "use db1")
		require.Error(t, err)

		// restore previous props state
		s.Properties.Set(config.KeyCatalog, catalogName)
	})

	t.Run("use catalog should reset the current database", func(t *testing.T) {
		// set a test DB
		dbName := s.Properties.Get(config.KeyDatabase)
		s.Properties.Set(config.KeyDatabase, "test-db")

		// use catalog should remove the DB property
		_, err := processUseStatement(s.Properties, "use catalog test")
		require.Nil(t, err)
		require.False(t, s.Properties.HasKey(config.KeyDatabase))

		// restore previous props state
		s.Properties.Set(config.KeyDatabase, dbName)
	})
}

func TestParseStatementType(t *testing.T) {
	require.Equal(t, SetStatement, parseStatementType("set ..."))
	require.Equal(t, UseStatement, parseStatementType("use ..."))
	require.Equal(t, ResetStatement, parseStatementType("reset ..."))
	require.Equal(t, ExitStatement, parseStatementType("exit;"))
	require.Equal(t, QuitStatement, parseStatementType("quit;"))
	require.Equal(t, QuitStatement, parseStatementType("quit"))
	require.Equal(t, OtherStatement, parseStatementType("Some other statement"))
}

func hoursToSeconds(hours float32) int {
	return int(hours * 60 * 60)
}

func TestIsUserSecretKey(t *testing.T) {
	require.True(t, hasSensitiveKey("sql.secrets.openai"))
	require.True(t, hasSensitiveKey("sql.secrets.openai"))
	require.True(t, hasSensitiveKey("sql.secrets.penaik"))
	require.True(t, hasSensitiveKey("sql.secrets.oopenai"))
	require.True(t, hasSensitiveKey("sql.secrets.oenaik"))
	require.True(t, hasSensitiveKey("sql.secrets.oppenai"))
	require.True(t, hasSensitiveKey("sql.secrets.opnai"))
	require.True(t, hasSensitiveKey("sql.secrets.opeenai"))
	require.True(t, hasSensitiveKey("sql.secrets.opeai"))
	require.True(t, hasSensitiveKey("sql.secrets.opennai"))
	require.True(t, hasSensitiveKey("sql.secrets.openi"))
	require.True(t, hasSensitiveKey("sql.secrets.openaai"))
	require.True(t, hasSensitiveKey("sql.secrets.opena"))
	require.True(t, hasSensitiveKey("sql.secrets.openaii"))
	require.True(t, hasSensitiveKey("sql.secrets.openaiii"))

	require.True(t, hasSensitiveKey("SQL.SECRETS.openai"))
	require.True(t, hasSensitiveKey("SQL.SECRETS.openai"))
	require.True(t, hasSensitiveKey("SQL.SECRETS.penaik"))
	require.True(t, hasSensitiveKey("SQL.SECRETS.oopenai"))
	require.True(t, hasSensitiveKey("SQL.SECRETS.oenaik"))
	require.True(t, hasSensitiveKey("SQL.SECRETS.oppenai"))
	require.True(t, hasSensitiveKey("SQL.SECRETS.opnai"))
	require.True(t, hasSensitiveKey("SQL.SECRETS.opeenai"))
	require.True(t, hasSensitiveKey("SQL.SECRETS.opeai"))
	require.True(t, hasSensitiveKey("SQL.SECRETS.opennai"))
	require.True(t, hasSensitiveKey("SQL.SECRETS.openi"))
	require.True(t, hasSensitiveKey("SQL.SECRETS.openaai"))
	require.True(t, hasSensitiveKey("SQL.SECRETS.opena"))
	require.True(t, hasSensitiveKey("SQL.SECRETS.openaii"))
	require.True(t, hasSensitiveKey("SQL.SECRETS.openaiii"))

	require.True(t, hasSensitiveKey("SQL.SECRETS.OPENAI"))
	require.True(t, hasSensitiveKey("SQL.SECRETS.OPENAI"))
	require.True(t, hasSensitiveKey("SQL.SECRETS.PENAIK"))
	require.True(t, hasSensitiveKey("SQL.SECRETS.OOPENAI"))
	require.True(t, hasSensitiveKey("SQL.SECRETS.OENAIK"))
	require.True(t, hasSensitiveKey("SQL.SECRETS.OPPENAI"))
	require.True(t, hasSensitiveKey("SQL.SECRETS.OPNAI"))
	require.True(t, hasSensitiveKey("SQL.SECRETS.OPEENAI"))
	require.True(t, hasSensitiveKey("SQL.SECRETS.OPEAI"))
	require.True(t, hasSensitiveKey("SQL.SECRETS.OPENNAI"))
	require.True(t, hasSensitiveKey("SQL.SECRETS.OPENI"))
	require.True(t, hasSensitiveKey("SQL.SECRETS.OPENAAI"))
	require.True(t, hasSensitiveKey("SQL.SECRETS.OPENA"))
	require.True(t, hasSensitiveKey("SQL.SECRETS.OPENAII"))
	require.True(t, hasSensitiveKey("SQL.SECRETS.OPENAIII"))
	require.True(t, hasSensitiveKey("SQL.SECRETS.NAME"))
	require.True(t, hasSensitiveKey("SQL.SECRETS.SCERECETASDT"))
	require.True(t, hasSensitiveKey("SQL.SECRETS.SECCCCCCCCRET"))

	require.False(t, hasSensitiveKey(""))
	require.False(t, hasSensitiveKey("gustavo"))
	require.False(t, hasSensitiveKey("sql.current-catalog"))
	require.False(t, hasSensitiveKey("client.results-timeout"))
	require.False(t, hasSensitiveKey("OPENAPI.KEY"))
	require.False(t, hasSensitiveKey("SEEEEECRET.OPENAPI.KEY"))
	require.False(t, hasSensitiveKey("SECRET.OPENAPI.KEY"))
}

func TestIsUserSecretKey2(t *testing.T) {
	require.True(t, hasSensitiveKey("sql.secrets.mysecret"))
}

func TestFormatUTCOffsetToTimezone(t *testing.T) {
	testCases := []struct {
		offsetSeconds int
		expected      string
	}{
		{
			offsetSeconds: hoursToSeconds(5.5),
			expected:      "GMT+05:30",
		},
		{
			offsetSeconds: hoursToSeconds(-6),
			expected:      "GMT-06:00",
		},
		{
			offsetSeconds: hoursToSeconds(0),
			expected:      "GMT+00:00",
		},
		{
			offsetSeconds: hoursToSeconds(-2.25),
			expected:      "GMT-02:15",
		},
		{
			offsetSeconds: hoursToSeconds(3.75),
			expected:      "GMT+03:45",
		},
	}

	for _, tc := range testCases {
		actual := formatUTCOffsetToTimezone(tc.offsetSeconds)
		require.Equal(t, tc.expected, actual)
	}
}

func TestTokenizeSQL(t *testing.T) {
	require := require.New(t)
	// Test escaped backticks
	input := "`a``b`"
	expected := []string{"a`b"}
	require.Equal(expected, TokenizeSQL(input))

	// Test trailing whitespaces
	input = "   The dog  "
	expected = []string{"The", "dog"}
	require.Equal(expected, TokenizeSQL(input))

	// Test whitespaces inside backticks
	input = "   `The dog`  "
	expected = []string{"The dog"}
	require.Equal(expected, TokenizeSQL(input))

	// Test basic string with backticks
	input = "   The  `quick`  `brown``fox`  jumps over   the  lazy dog  "
	expected = []string{"The", "quick", "brown`fox", "jumps", "over", "the", "lazy", "dog"}
	require.Equal(expected, TokenizeSQL(input))

	// Test string with escaped backticks
	input = "   The  `quick`  `brown``f``o``x`  jumps over   the  lazy dog  "
	expected = []string{"The", "quick", "brown`f`o`x", "jumps", "over", "the", "lazy", "dog"}
	require.Equal(expected, TokenizeSQL(input))

	// Test string with unclosed backtick
	input = "   The  `quick`  `brown``fox  jumps over   the  lazy dog  "
	expected = []string{"The", "quick", "brown`fox  jumps over   the  lazy dog  "}
	require.Equal(expected, TokenizeSQL(input))

	// Test closed closing backtick
	input = "   The  `quick`  `brown``fox  jumps over   the  lazy dog `"
	expected = []string{"The", "quick", "brown`fox  jumps over   the  lazy dog "}
	require.Equal(expected, TokenizeSQL(input))

	input = "   USE  CATALOG   `catalog`"
	expected = []string{"USE", "CATALOG", "catalog"}
	require.Equal(expected, TokenizeSQL(input))

	input = "   USE  CATALOG catalog"
	expected = []string{"USE", "CATALOG", "catalog"}
	require.Equal(expected, TokenizeSQL(input))

	input = "   UsE  CATalOG   `catalog`"
	expected = []string{"UsE", "CATalOG", "catalog"}
	require.Equal(expected, TokenizeSQL(input))

	input = "   USE  CATALOG catAlog"
	expected = []string{"USE", "CATALOG", "catAlog"}
	require.Equal(expected, TokenizeSQL(input))

	input = "   USE  db"
	expected = []string{"USE", "db"}
	require.Equal(expected, TokenizeSQL(input))

	input = "   USE  `db`"
	expected = []string{"USE", "db"}
	require.Equal(expected, TokenizeSQL(input))

	input = "   USE  `my catalog`.`my db`"
	expected = []string{"USE", "my catalog", ".", "my db"}
	require.Equal(expected, TokenizeSQL(input))

	input = "   USE  `catalog`.`db`"
	expected = []string{"USE", "catalog", ".", "db"}
	require.Equal(expected, TokenizeSQL(input))

	input = "   USE  `catalog`.`db.1`"
	expected = []string{"USE", "catalog", ".", "db.1"}
	require.Equal(expected, TokenizeSQL(input))

	input = "   USE  `catalog`.db"
	expected = []string{"USE", "catalog", ".", "db"}
	require.Equal(expected, TokenizeSQL(input))

	input = "   USE  catalog.`db`"
	expected = []string{"USE", "catalog", ".", "db"}
	require.Equal(expected, TokenizeSQL(input))

	input = "   USE  catalog.db"
	expected = []string{"USE", "catalog", ".", "db"}
	require.Equal(expected, TokenizeSQL(input))

	input = "   USE  catalog   .  db"
	expected = []string{"USE", "catalog", ".", "db"}
	require.Equal(expected, TokenizeSQL(input))

	input = "   USE  catalog.  db"
	expected = []string{"USE", "catalog", ".", "db"}
	require.Equal(expected, TokenizeSQL(input))

	input = "USE catalog.`my database`"
	expected = []string{"USE", "catalog", ".", "my database"}
	require.Equal(expected, TokenizeSQL(input))

	// Test empty string
	input = ""
	expected = []string{}
	require.Equal(expected, TokenizeSQL(input))

	// Test string with only whitespace
	input = "   \t\n\r "
	expected = []string{}
	require.Equal(expected, TokenizeSQL(input))

	// Test string with only backticks
	input = "````"
	expected = []string{"`"}
	require.Equal(expected, TokenizeSQL(input))
}

func TestTokenizeSQLSpecialCharacters(t *testing.T) {
	require := require.New(t)

	input := "my clusté€r"
	expected := []string{"my", "clusté€r"}
	require.Equal(expected, TokenizeSQL(input))

	input = "my cluster αβγбвг汉字あア한😀"
	expected = []string{"my", "cluster", "αβγбвг汉字あア한😀"}
	require.Equal(expected, TokenizeSQL(input))
}

func TestGetSubstringUpToSecondDot(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal case with two dots",
			input:    "sql.secrets.mysecret",
			expected: "sql.secrets.",
		},
		{
			name:     "normal case with two dots",
			input:    "sql.secretxx.mysecret",
			expected: "sql.secretxx.",
		},
		{
			name:     "normal case with two dots",
			input:    "x.y.z",
			expected: "x.y.",
		},
		{
			name:     "only one dot",
			input:    "x.y",
			expected: "",
		},
		{
			name:     "no dots",
			input:    "xyz",
			expected: "",
		},
		{
			name:     "more than two dots",
			input:    "x.y.z.a",
			expected: "x.y.",
		},
		{
			name:     "dots at start",
			input:    ".x.y.z",
			expected: ".x.",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "dots only",
			input:    "...",
			expected: "..",
		},
		{
			name:     "space between dots",
			input:    "x. .z",
			expected: "x. .",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getSubstringUpToSecondDot(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}
