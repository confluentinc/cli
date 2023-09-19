package store

import (
	"fmt"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/v3/pkg/ccloudv2"
	"github.com/confluentinc/cli/v3/pkg/flink/config"
	"github.com/confluentinc/cli/v3/pkg/flink/types"
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

func TestRemoveWhiteSpaces(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "removeTabNewLineAndWhitesSpaces() removes all whitespaces", args: args{str: " key=value "}, want: "key=value"},
		{name: "removeTabNewLineAndWhitesSpaces() removes all whitespaces", args: args{str: " key  =    value "}, want: "key=value"},

		{name: "removeTabNewLineAndWhitesSpaces() removes all newlines", args: args{str: "key=\nvalue"}, want: "key=value"},
		{name: "removeTabNewLineAndWhitesSpaces() removes all newlines", args: args{str: " key\n=value "}, want: "key=value"},
		{name: "removeTabNewLineAndWhitesSpaces() removes all newlines", args: args{str: "\nkey=\nvalue"}, want: "key=value"},
		{name: "removeTabNewLineAndWhitesSpaces() removes all newlines", args: args{str: "key=\nvalue\n"}, want: "key=value"},
		{name: "removeTabNewLineAndWhitesSpaces() removes all newlines", args: args{str: "\nkey=\nvalue\n"}, want: "key=value"},

		{name: "removeTabNewLineAndWhitesSpaces() removes all newlines", args: args{str: "key=\r\nvalue"}, want: "key=value"},
		{name: "removeTabNewLineAndWhitesSpaces() removes all newlines", args: args{str: " key\r\n=value "}, want: "key=value"},
		{name: "removeTabNewLineAndWhitesSpaces() removes all newlines", args: args{str: "\r\nkey=\r\nvalue"}, want: "key=value"},
		{name: "removeTabNewLineAndWhitesSpaces() removes all newlines", args: args{str: "key=\r\nvalue\r\n"}, want: "key=value"},
		{name: "removeTabNewLineAndWhitesSpaces() removes all newlines", args: args{str: "\r\nkey=\r\nvalue\r\n"}, want: "key=value"},

		{name: "removeTabNewLineAndWhitesSpaces() removes all tabs", args: args{str: "key=\tvalue"}, want: "key=value"},
		{name: "removeTabNewLineAndWhitesSpaces() removes all tabs", args: args{str: " key\t=value "}, want: "key=value"},
		{name: "removeTabNewLineAndWhitesSpaces() removes all tabs", args: args{str: "\tkey=\tvalue"}, want: "key=value"},
		{name: "removeTabNewLineAndWhitesSpaces() removes all tabs", args: args{str: "key=\tvalue\t"}, want: "key=value"},
		{name: "removeTabNewLineAndWhitesSpaces() removes all tabs", args: args{str: "\tkey=\tvalue\t"}, want: "key=value"},

		{name: "removeTabNewLineAndWhitesSpaces() removes all new lines, tabs and whitespaces", args: args{str: "\n \tkey\n=\n\tvalue\n"}, want: "key=value"},
		{name: "removeTabNewLineAndWhitesSpaces() removes all new lines, tabs and whitespaces", args: args{str: "\r\n \tkey\t=\t\tvalue\n"}, want: "key=value"},
		{name: "removeTabNewLineAndWhitesSpaces() removes all new lines, tabs and whitespaces", args: args{str: "\n \tkey\n = \n\tvalue\r\n"}, want: "key=value"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := removeTabNewLineAndWhitesSpaces(test.args.str); got != test.want {
				require.Equal(t, test.want, got)
			}
		})
	}
}

func TestProcessSetStatement(t *testing.T) {
	// Create a new store
	client := ccloudv2.NewFlinkGatewayClient("url", "userAgent", false, "authToken")
	s := NewStore(client, nil, &types.ApplicationOptions{EnvironmentName: "env-123"}, tokenRefreshFunc).(*Store)
	// This is just a string, so really doesn't matter
	s.Properties.Set(config.ConfigKeyLocalTimeZone, "London/GMT")

	t.Run("should return an error message if statement is invalid", func(t *testing.T) {
		_, err := s.processSetStatement("se key=value")
		assert.NotNil(t, err)
	})

	t.Run("should return all keys and values from config if configKey is empty", func(t *testing.T) {
		result, err := s.processSetStatement("set")
		assert.Nil(t, err)
		assert.EqualValues(t, types.COMPLETED, result.Status)

		assert.Equal(t, 2, len(result.StatementResults.Headers))
		assert.Equal(t, len(s.Properties.GetProperties()), len(result.StatementResults.Rows))
		cupaloy.SnapshotT(t, result.StatementResults)
	})

	t.Run("should update config for valid configKey", func(t *testing.T) {
		result, err := s.processSetStatement("set 'location'='USA'")
		assert.Nil(t, err)
		assert.EqualValues(t, types.COMPLETED, result.Status)
		assert.Equal(t, "configuration updated successfully", result.StatusDetail)
		cupaloy.SnapshotT(t, result.StatementResults)
	})

	t.Run("should return all keys and values from config if configKey is empty after updates", func(t *testing.T) {
		result, err := s.processSetStatement("set")
		assert.Nil(t, err)
		assert.EqualValues(t, types.COMPLETED, result.Status)

		assert.Equal(t, 2, len(result.StatementResults.Headers))
		assert.Equal(t, len(s.Properties.GetProperties()), len(result.StatementResults.Rows))
		cupaloy.SnapshotT(t, result.StatementResults)
	})

	t.Run("should fail if user wants to set the catalog", func(t *testing.T) {
		_, err := s.processSetStatement("set" + config.ConfigKeyCatalog)
		assert.NotNil(t, err)
	})

	t.Run("should fail if user wants to set the database", func(t *testing.T) {
		_, err := s.processSetStatement("set" + config.ConfigKeyDatabase)
		assert.NotNil(t, err)
	})
}

func TestProcessResetStatement(t *testing.T) {
	// Create a new store
	client := ccloudv2.NewFlinkGatewayClient("url", "userAgent", false, "authToken")
	appOptions := types.ApplicationOptions{
		OrgResourceId:    "orgId",
		EnvironmentName:  "envName",
		Database:         "database",
		ServiceAccountId: "sa-123",
	}
	s := NewStore(client, nil, &appOptions, tokenRefreshFunc).(*Store)
	s.Properties.Set(config.ConfigKeyLocalTimeZone, "London/GMT")

	defaultSetOutput := createStatementResults([]string{"Key", "Value"}, [][]string{
		{config.ConfigKeyLocalTimeZone, fmt.Sprintf("%s (default)", getLocalTimezone())},
		{config.ConfigKeyServiceAcount, fmt.Sprintf("%s (default)", appOptions.ServiceAccountId)},
	})

	t.Run("should return all keys and values including default and initial values before reseting", func(t *testing.T) {
		result, err := s.processSetStatement("set")
		assert.Nil(t, err)
		assert.EqualValues(t, types.COMPLETED, result.Status)

		assert.Equal(t, 2, len(result.StatementResults.Headers))
		assert.Equal(t, len(s.Properties.GetProperties()), len(result.StatementResults.Rows))
		cupaloy.SnapshotT(t, result.StatementResults)
	})

	t.Run("should return an error message if statement is invalid", func(t *testing.T) {
		_, err := s.processResetStatement("res key")
		assert.NotNil(t, err)
	})

	t.Run("should reset all keys and values from config to their default or delete them if no default", func(t *testing.T) {
		s.Properties.Set(config.ConfigKeyCatalog, "job1")
		s.Properties.Set("timeout", "30")
		result, _ := s.processResetStatement("reset")
		assert.EqualValues(t, types.COMPLETED, result.Status)
		assert.Equal(t, "configuration has been reset successfully", result.StatusDetail)
		assert.ElementsMatch(t, defaultSetOutput.GetHeaders(), result.StatementResults.GetHeaders())
		assert.ElementsMatch(t, defaultSetOutput.GetRows(), result.StatementResults.GetRows())
	})

	t.Run("should return an error message if configKey does not exist", func(t *testing.T) {
		result, err := s.processResetStatement("reset 'location'")
		assert.NotNil(t, err)
		assert.Equal(t, `Error: configuration key "location" is not set`, err.Error())
		assert.Nil(t, result)
	})

	t.Run("should reset config for valid configKey", func(t *testing.T) {
		s.Properties.Set("catalog", "job1")
		result, _ := s.processResetStatement("reset 'catalog'")
		assert.EqualValues(t, types.COMPLETED, result.Status)
		assert.Equal(t, `configuration key "catalog" has been reset successfully`, result.StatusDetail)
		assert.ElementsMatch(t, defaultSetOutput.GetHeaders(), result.StatementResults.GetHeaders())
		assert.ElementsMatch(t, defaultSetOutput.GetRows(), result.StatementResults.GetRows())
	})

	t.Run("should reset database if catalog is reset", func(t *testing.T) {
		s.Properties.Set(config.ConfigKeyCatalog, "catalog")
		s.Properties.Set(config.ConfigKeyDatabase, "db")
		statement := fmt.Sprintf("reset '%s'", config.ConfigKeyCatalog)
		_, err := s.processResetStatement(statement)
		assert.Nil(t, err)
		assert.False(t, s.Properties.HasKey(config.ConfigKeyCatalog))
		assert.False(t, s.Properties.HasKey(config.ConfigKeyDatabase))
	})
}

func TestProcessUseStatement(t *testing.T) {
	// Create a new store
	client := ccloudv2.NewFlinkGatewayClient("url", "userAgent", false, "authToken")
	appOptions := types.ApplicationOptions{
		OrgResourceId:   "orgId",
		EnvironmentName: "envName",
		Database:        "database",
	}
	s := NewStore(client, nil, &appOptions, tokenRefreshFunc).(*Store)

	t.Run("should return an error message if statement is invalid", func(t *testing.T) {
		_, err := s.processUseStatement("us")
		require.Error(t, err)
	})

	t.Run("should update the database name in config", func(t *testing.T) {
		result, err := s.processUseStatement("use db1")
		require.Nil(t, err)
		require.Equal(t, config.ConfigOpUse, result.Kind)
		require.EqualValues(t, types.COMPLETED, result.Status)
		require.Equal(t, "configuration updated successfully", result.StatusDetail)
		expectedResult := createStatementResults([]string{"Key", "Value"}, [][]string{{config.ConfigKeyDatabase, "db1"}})
		assert.Equal(t, expectedResult, result.StatementResults)
	})

	t.Run("should return an error message if catalog name is missing", func(t *testing.T) {
		_, err := s.processUseStatement("use catalog")
		require.Error(t, err)
	})

	t.Run("should update the catalog name in config", func(t *testing.T) {
		result, err := s.processUseStatement("use catalog metadata")
		require.Nil(t, err)
		require.Equal(t, config.ConfigOpUse, result.Kind)
		require.EqualValues(t, types.COMPLETED, result.Status)
		require.Equal(t, "configuration updated successfully", result.StatusDetail)
		expectedResult := createStatementResults([]string{"Key", "Value"}, [][]string{{config.ConfigKeyCatalog, "metadata"}})
		assert.Equal(t, expectedResult, result.StatementResults)
	})

	t.Run("should return an error message for invalid syntax", func(t *testing.T) {
		_, err := s.processUseStatement("use db1 catalog metadata")
		require.Error(t, err)
	})

	t.Run("use database should fail if no catalog was set", func(t *testing.T) {
		// delete the catalog
		catalogName := s.Properties.Get(config.ConfigKeyCatalog)
		s.Properties.Delete(config.ConfigKeyCatalog)

		_, err := s.processUseStatement("use db1")
		require.Error(t, err)

		// restore previous props state
		s.Properties.Set(config.ConfigKeyCatalog, catalogName)
	})

	t.Run("use catalog should reset the current database", func(t *testing.T) {
		// set a test DB
		dbName := s.Properties.Get(config.ConfigKeyDatabase)
		s.Properties.Set(config.ConfigKeyDatabase, "test-db")

		// use catalog should remove the DB property
		_, err := s.processUseStatement("use catalog test")
		require.Nil(t, err)
		require.False(t, s.Properties.HasKey(config.ConfigKeyDatabase))

		// restore previous props state
		s.Properties.Set(config.ConfigKeyDatabase, dbName)
	})
}

func TestStartsWithValidSQL(t *testing.T) {
	require.True(t, startsWithValidSQL("SELECT * FROM users"))
	require.True(t, startsWithValidSQL("INSERT INTO orders (customer_id, product_id) VALUES (1, 2)"))
	require.False(t, startsWithValidSQL("asdasd"))
	require.False(t, startsWithValidSQL(";"))
	require.False(t, startsWithValidSQL(""))
	require.False(t, startsWithValidSQL("This is not a valid SQL statement"))
}

func TestParseStatementType(t *testing.T) {
	require.Equal(t, SetStatement, parseStatementType("set ..."))
	require.Equal(t, UseStatement, parseStatementType("use ..."))
	require.Equal(t, ResetStatement, parseStatementType("reset ..."))
	require.Equal(t, ExitStatement, parseStatementType("exit;"))
	require.Equal(t, OtherStatement, parseStatementType("Some other statement"))
}

func hoursToSeconds(hours float32) int {
	return int(hours * 60 * 60)
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
