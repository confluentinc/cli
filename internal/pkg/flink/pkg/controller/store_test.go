package controller

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type StoreTestSuite struct {
	suite.Suite
	store Store
}

func TestStoreTestSuite(t *testing.T) {
	suite.Run(t, new(StoreTestSuite))
}

func (s *StoreTestSuite) TestIsSETQuery() {
	assert.True(s.T(), true, queryStartsWithOp("SET", configOpSet))
	assert.True(s.T(), true, queryStartsWithOp("SET key", configOpSet))
	assert.True(s.T(), true, queryStartsWithOp("SET key=value", configOpSet))
	assert.True(s.T(), true, queryStartsWithOp("    SET key=value", configOpSet))
	assert.True(s.T(), true, queryStartsWithOp("    SET   ", configOpSet))
	assert.True(s.T(), true, queryStartsWithOp("    set   ", configOpSet))
	assert.True(s.T(), true, queryStartsWithOp("    SET key=value", configOpSet))

	assert.False(s.T(), false, queryStartsWithOp("SETting", configOpSet))
	assert.False(s.T(), false, queryStartsWithOp("", configOpSet))
	assert.False(s.T(), false, queryStartsWithOp("should be false", configOpSet))
	assert.False(s.T(), false, queryStartsWithOp("USE", configOpSet))
	assert.False(s.T(), false, queryStartsWithOp("SETTING", configOpSet))
}

func (s *StoreTestSuite) TestIsUSEQuery() {
	assert.True(s.T(), queryStartsWithOp("USE", configOpUse))
	assert.True(s.T(), queryStartsWithOp("USE catalog", configOpUse))
	assert.True(s.T(), queryStartsWithOp("USE CATALOG cat", configOpUse))
	assert.True(s.T(), queryStartsWithOp("    use CATALOG cat", configOpUse))
	assert.True(s.T(), queryStartsWithOp("    USE   ", configOpUse))
	assert.True(s.T(), queryStartsWithOp("    use   ", configOpUse))
	assert.True(s.T(), queryStartsWithOp("    USE CATALOG cat", configOpUse))

	assert.False(s.T(), queryStartsWithOp("SET", configOpUse))
	assert.False(s.T(), queryStartsWithOp("USES", configOpUse))
	assert.False(s.T(), queryStartsWithOp("", configOpUse))
	assert.False(s.T(), queryStartsWithOp("should be false", configOpUse))
}

func (s *StoreTestSuite) TestParseSETQuery() {
	key, value := parseSETQuery("SET key=value")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)

	key, value = parseSETQuery("  SET key=value;")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)

	key, value = parseSETQuery("  set key=value    ;")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)

	key, value = parseSETQuery("  set key = value    ")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)

	key, value = parseSETQuery("  set key     =    value    ")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)

	key, value = parseSETQuery("  set key= value    ")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)

	key, value = parseSETQuery("  set key =value    ")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)

	key, value = parseSETQuery("  set key		 =value    ")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)

	key, value = parseSETQuery("set")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)

	key, value = parseSETQuery("SET")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)

	key, value = parseSETQuery(" 		sET 	")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)

	key, value = parseSETQuery(" 		sET key	")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)

	key, value = parseSETQuery(" 		sET key=")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "", value)

	key, value = parseSETQuery(" 		sET key= 	")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "", value)

	key, value = parseSETQuery(" 		sET = value	")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)

	key, value = parseSETQuery(" 		sET key= \nvalue	")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)
}

func (s *StoreTestSuite) TestParseUSEQuery() {
	key, value := parseUSEQuery("USE CATALOG c;")
	assert.Equal(s.T(), configKeyCatalog, key)
	assert.Equal(s.T(), "c", value)

	key, value = parseUSEQuery("  use   catalog   \nc   ")
	assert.Equal(s.T(), configKeyCatalog, key)
	assert.Equal(s.T(), "c", value)

	key, value = parseUSEQuery("  use   catalog     ")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)

	key, value = parseUSEQuery("catalog   c")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)

	key, value = parseUSEQuery("  use     db   ")
	assert.Equal(s.T(), configKeyDatabase, key)
	assert.Equal(s.T(), "db", value)

	key, value = parseUSEQuery("dAtaBaSe  db   ")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)

	key, value = parseUSEQuery("  use     \ndatabase_name   ")
	assert.Equal(s.T(), configKeyDatabase, key)
	assert.Equal(s.T(), "database_name", value)
}
