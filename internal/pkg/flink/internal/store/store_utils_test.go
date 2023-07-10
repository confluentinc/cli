package store

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	"github.com/confluentinc/cli/internal/pkg/flink/config"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
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
	s := NewStore(client, nil, &types.ApplicationOptions{Context: newContext()}, tokenRefreshFunc).(*Store)
	s.Properties[config.ConfigKeyLocalTimeZone] = "UTC+01:00"

	t.Run("should return an error message if statement is invalid", func(t *testing.T) {
		_, err := s.processSetStatement("se key=value")
		assert.NotNil(t, err)
	})

	t.Run("should return all keys and values from config if configKey is empty", func(t *testing.T) {
		result, err := s.processSetStatement("set")
		assert.Nil(t, err)
		assert.EqualValues(t, types.COMPLETED, result.Status)
		cupaloy.SnapshotT(t, result.StatementResults)

		// Add some key-value pairs to the config
		s.Properties["catalog"] = "job1"
		s.Properties["timeout"] = "30"
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
		expectedKeyValuePairs := map[string]string{"catalog": "job1", "timeout": "30", "location": "USA"}

		// check row and column lengths match
		assert.Equal(t, 2, len(result.StatementResults.Headers))
		// + 2 for the other default properties apart from catalog
		assert.Equal(t, len(expectedKeyValuePairs)+2, len(result.StatementResults.Rows))
		// check if all expected key value pairs are in the results
		cupaloy.SnapshotT(t, result.StatementResults)
	})
}

func TestProcessResetStatement(t *testing.T) {
	// Create a new store
	client := ccloudv2.NewFlinkGatewayClient("url", "userAgent", false, "authToken")
	s := NewStore(client, nil, nil, tokenRefreshFunc).(*Store)

	t.Run("should return an error message if statement is invalid", func(t *testing.T) {
		_, err := s.processResetStatement("res key")
		assert.NotNil(t, err)
	})

	t.Run("should reset all keys and values from config", func(t *testing.T) {
		s.Properties["catalog"] = "job1"
		s.Properties["timeout"] = "30"
		result, _ := s.processResetStatement("reset")
		assert.EqualValues(t, types.COMPLETED, result.Status)
		assert.Equal(t, "configuration has been reset successfully", result.StatusDetail)
		assert.Nil(t, result.StatementResults)
	})

	t.Run("should return an error message if configKey does not exist", func(t *testing.T) {
		result, err := s.processResetStatement("reset 'location'")
		assert.NotNil(t, err)
		assert.Equal(t, `Error: configuration key "location" is not set`, err.Error())
		assert.Nil(t, result)
	})

	t.Run("should reset config for valid configKey", func(t *testing.T) {
		s.Properties["catalog"] = "job1"
		result, _ := s.processResetStatement("reset 'catalog'")
		assert.EqualValues(t, types.COMPLETED, result.Status)
		assert.Equal(t, `configuration key "catalog" has been reset successfully`, result.StatusDetail)
		expectedResult := createStatementResults([]string{"Key", "Value"}, [][]string{})
		assert.Equal(t, &expectedResult, result.StatementResults)
	})
	t.Run("should return all keys and values from config after updates", func(t *testing.T) {
		result, _ := s.processResetStatement("reset")

		assert.EqualValues(t, types.COMPLETED, result.Status)
		assert.Equal(t, "configuration has been reset successfully", result.StatusDetail)
		assert.Nil(t, result.StatementResults)
	})
}

func TestProcessUseStatement(t *testing.T) {
	// Create a new store
	client := ccloudv2.NewFlinkGatewayClient("url", "userAgent", false, "authToken")
	s := NewStore(client, nil, nil, tokenRefreshFunc).(*Store)

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
		assert.Equal(t, &expectedResult, result.StatementResults)
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
		assert.Equal(t, &expectedResult, result.StatementResults)
	})

	t.Run("should return an error message for invalid syntax", func(t *testing.T) {
		_, err := s.processUseStatement("use db1 catalog metadata")
		require.Error(t, err)
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
			expected:      "UTC+05:30",
		},
		{
			offsetSeconds: hoursToSeconds(-6),
			expected:      "UTC-06:00",
		},
		{
			offsetSeconds: hoursToSeconds(0),
			expected:      "UTC+00:00",
		},
		{
			offsetSeconds: hoursToSeconds(-2.25),
			expected:      "UTC-02:15",
		},
		{
			offsetSeconds: hoursToSeconds(3.75),
			expected:      "UTC+03:45",
		},
	}

	for _, tc := range testCases {
		actual := formatUTCOffsetToTimezone(tc.offsetSeconds)
		require.Equal(t, tc.expected, actual)
	}
}
