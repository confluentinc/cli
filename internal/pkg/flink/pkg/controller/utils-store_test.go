package controller

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeStatementTerminator(tt.args.statement); got != tt.want {
				require.Equal(t, tt.want, got)
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeTabNewLineAndWhitesSpaces(tt.args.str); got != tt.want {
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestProcessSetStatement(t *testing.T) {
	// Create a new store
	client := NewGatewayClient("envId", "computePoolId", "authToken")
	s := NewStore(client, nil).(*Store)

	t.Run("should return an error message if statement is invalid", func(t *testing.T) {
		_, err := s.processSetStatement("se key=value")
		assert.NotNil(t, err)
	})

	t.Run("should return all keys and values from config if configKey is empty", func(t *testing.T) {
		result, err := s.processSetStatement("set")
		assert.Nil(t, err)
		assert.EqualValues(t, "Completed", result.Status)
		assert.Equal(t, []string{"Key", "Value"}, result.Columns)
		assert.Equal(t, [][]string{}, result.Rows)

		// Add some key-value pairs to the config
		s.Config["pipeline.name"] = "job1"
		s.Config["timeout"] = "30"
	})

	t.Run("should update config for valid configKey", func(t *testing.T) {
		result, err := s.processSetStatement("set location=USA")
		assert.Nil(t, err)
		assert.EqualValues(t, "Completed", result.Status)
		assert.Equal(t, "Config updated successfuly.", result.StatusDetail)
		assert.Equal(t, []string{"Key", "Value"}, result.Columns)
		assert.Equal(t, [][]string{{"location", "USA"}}, result.Rows)
	})

	t.Run("should return all keys and values from config if configKey is empty after updates", func(t *testing.T) {
		result, err := s.processSetStatement("set")
		assert.Nil(t, err)
		assert.EqualValues(t, "Completed", result.Status)
		assert.Equal(t, []string{"Key", "Value"}, result.Columns)

		// sort both the expected and actual rows to compare equality
		sort.Slice(result.Rows, func(i, j int) bool {
			return result.Rows[i][0] < result.Rows[j][0]
		})

		expectedRows := [][]string{{"pipeline.name", "job1"}, {"timeout", "30"}, {"location", "USA"}}
		sort.Slice(expectedRows, func(i, j int) bool {
			return expectedRows[i][0] < expectedRows[j][0]
		})
		assert.Equal(t, expectedRows, result.Rows)
	})
}

func TestProcessResetStatement(t *testing.T) {
	// Create a new store
	client := NewGatewayClient("envId", "computePoolId", "authToken")
	s := NewStore(client, nil).(*Store)

	t.Run("should return an error message if statement is invalid", func(t *testing.T) {
		_, err := s.processResetStatement("res key")
		assert.NotNil(t, err)
	})

	t.Run("should reset all keys and values from config", func(t *testing.T) {
		s.Config["pipeline.name"] = "job1"
		s.Config["timeout"] = "30"
		result, _ := s.processResetStatement("reset")
		assert.EqualValues(t, "Completed", result.Status)
		assert.Equal(t, "Configuration has been reset successfuly.", result.StatusDetail)
		assert.Equal(t, []string(nil), result.Columns)
		assert.Equal(t, [][]string(nil), result.Rows)
	})

	t.Run("should return an error message if configKey does not exist", func(t *testing.T) {
		result, err := s.processResetStatement("reset location")
		assert.NotNil(t, err)
		assert.Equal(t, "Error: Config key \"location\" is currently not set.", err.Error())
		assert.Nil(t, result)
	})

	t.Run("should reset config for valid configKey", func(t *testing.T) {
		s.Config["pipeline.name"] = "job1"
		result, _ := s.processResetStatement("reset pipeline.name")
		assert.EqualValues(t, "Completed", result.Status)
		assert.Equal(t, "Config key \"pipeline.name\" has been reset successfuly.", result.StatusDetail)
		assert.Equal(t, []string{"Key", "Value"}, result.Columns)
		assert.Equal(t, [][]string{}, result.Rows)
	})
	t.Run("should return all keys and values from config after updates", func(t *testing.T) {
		result, _ := s.processResetStatement("reset")

		assert.EqualValues(t, "Completed", result.Status)
		assert.Equal(t, "Configuration has been reset successfuly.", result.StatusDetail)
		assert.Equal(t, []string(nil), result.Columns)
		assert.Equal(t, [][]string(nil), result.Rows)
	})
}

func TestProcessUseStatement(t *testing.T) {
	// Create a new store
	client := NewGatewayClient("envId", "computePoolId", "authToken")
	s := NewStore(client, nil).(*Store)

	t.Run("should return an error message if statement is invalid", func(t *testing.T) {
		_, err := s.processUseStatement("us")
		require.Error(t, err)
	})

	t.Run("should update the database name in config", func(t *testing.T) {
		result, err := s.processUseStatement("use db1")
		require.NoError(t, err)
		require.Equal(t, configOpUse, result.Statement)
		require.EqualValues(t, "Completed", result.Status)
		require.Equal(t, "Config updated successfuly.", result.StatusDetail)
		require.Equal(t, []string{"Key", "Value"}, result.Columns)
		require.Equal(t, [][]string{{configKeyDatabase, "db1"}}, result.Rows)
	})

	t.Run("should return an error message if catalog name is missing", func(t *testing.T) {
		_, err := s.processUseStatement("use catalog")
		require.Error(t, err)
	})

	t.Run("should update the catalog name in config", func(t *testing.T) {
		result, err := s.processUseStatement("use catalog metadata")
		require.NoError(t, err)
		require.Equal(t, configOpUse, result.Statement)
		require.EqualValues(t, "Completed", result.Status)
		require.Equal(t, "Config updated successfuly.", result.StatusDetail)
		require.Equal(t, []string{"Key", "Value"}, result.Columns)
		require.Equal(t, [][]string{{configKeyCatalog, "metadata"}}, result.Rows)
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
	require.Equal(t, SET_STATEMENT, parseStatementType("set ..."))
	require.Equal(t, USE_STATEMENT, parseStatementType("use ..."))
	require.Equal(t, RESET_STATEMENT, parseStatementType("reset ..."))
	require.Equal(t, OTHER_STATEMENT, parseStatementType("Some other statement"))
}
