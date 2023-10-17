package store

import (
	"fmt"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"pgregory.net/rapid"

	"github.com/confluentinc/cli/v3/pkg/flink/config"
)

type UserPropertiesTestSuite struct {
	suite.Suite
	defaultKey     string
	defaultValue   string
	userProperties UserProperties
}

func TestUserPropertiesTestSuite(t *testing.T) {
	suite.Run(t, new(UserPropertiesTestSuite))
}

func (s *UserPropertiesTestSuite) SetupTest() {
	s.defaultKey = "default-key"
	s.defaultValue = "default-value"
	s.userProperties = NewUserProperties(map[string]string{s.defaultKey: s.defaultValue}, map[string]string{})
	require.Equal(s.T(), s.defaultValue, s.userProperties.Get(s.defaultKey))
}

func (s *UserPropertiesTestSuite) addSomeKeys() map[string]string {
	newMap := map[string]string{s.defaultKey: s.defaultValue}
	numKeysToAdd := rapid.IntRange(5, 15).Example()
	for i := 0; i < numKeysToAdd; i++ {
		keyToAdd := fmt.Sprintf("new-key-%v", i)
		valToAdd := fmt.Sprintf("new-val-%v", i)
		newMap[keyToAdd] = valToAdd
		s.userProperties.Set(keyToAdd, valToAdd)
	}
	return newMap
}

func (s *UserPropertiesTestSuite) getRandomNonDefaultKey(fromMap map[string]string) string {
	keysWithoutDefault := make([]string, 0)
	for key := range fromMap {
		if key == s.defaultKey {
			continue
		}
		keysWithoutDefault = append(keysWithoutDefault, key)
	}
	return rapid.SampledFrom(keysWithoutDefault).Example()
}

func (s *UserPropertiesTestSuite) TestMapShouldAddKeys() {
	standardMap := s.addSomeKeys()

	require.Equal(s.T(), standardMap, s.userProperties.GetProperties())
}

func (s *UserPropertiesTestSuite) TestMapShouldOverwriteKey() {
	valToAdd := "new-val-for-default-key"

	s.userProperties.Set(s.defaultKey, valToAdd)

	require.Equal(s.T(), valToAdd, s.userProperties.Get(s.defaultKey))
}

func (s *UserPropertiesTestSuite) TestMapShouldGetKey() {
	standardMap := s.addSomeKeys()
	keyToGet := s.getRandomNonDefaultKey(standardMap)

	require.Equal(s.T(), standardMap[keyToGet], s.userProperties.Get(keyToGet))
}

func (s *UserPropertiesTestSuite) TestMapShouldGetEmptyStringIfKeyDoesNotExist() {
	nonExistingKey := "non-existing-key"

	require.Equal(s.T(), "", s.userProperties.Get(nonExistingKey))
}

func (s *UserPropertiesTestSuite) TestMapShouldGetDefaultValueIfKeyDoesNotExist() {
	nonExistingKey := "non-existing-key"
	defaultVal := "default"

	require.Equal(s.T(), defaultVal, s.userProperties.GetOrDefault(nonExistingKey, defaultVal))
}

func (s *UserPropertiesTestSuite) TestMapHasKeyReturnsFalseIfKeyDoesNotExist() {
	nonExistingKey := "non-existing-key"

	require.False(s.T(), s.userProperties.HasKey(nonExistingKey))
}

func (s *UserPropertiesTestSuite) TestMapHasKeyReturnsTrueIfKeyDoesExist() {
	standardMap := s.addSomeKeys()
	keyToGet := s.getRandomNonDefaultKey(standardMap)

	require.True(s.T(), s.userProperties.HasKey(keyToGet))
}

func (s *UserPropertiesTestSuite) TestDeleteRemovesNonDefaultKey() {
	standardMap := s.addSomeKeys()
	keyToDelete := s.getRandomNonDefaultKey(standardMap)
	delete(standardMap, keyToDelete)

	s.userProperties.Delete(keyToDelete)

	require.False(s.T(), s.userProperties.HasKey(keyToDelete))
	require.Equal(s.T(), standardMap, s.userProperties.GetProperties())
}

func (s *UserPropertiesTestSuite) TestDeleteDoesNotRemoveDefaultKey() {
	s.userProperties.Delete(s.defaultKey)

	require.True(s.T(), s.userProperties.HasKey(s.defaultKey))
}

func (s *UserPropertiesTestSuite) TestDeleteResetsDefaultKeyToDefaultValue() {
	valToAdd := "new-val-for-default-key"
	s.userProperties.Set(s.defaultKey, valToAdd)

	s.userProperties.Delete(s.defaultKey)

	require.Equal(s.T(), s.defaultValue, s.userProperties.Get(s.defaultKey))
}

func (s *UserPropertiesTestSuite) TestClearShouldRemoveAllNonDefaultKeysAndResetDefaultKeysToDefaultValues() {
	s.addSomeKeys()
	valToAdd := "new-val-for-default-key"
	s.userProperties.Set(s.defaultKey, valToAdd)

	s.userProperties.Clear()

	require.Len(s.T(), s.userProperties.GetProperties(), 1)
	require.Equal(s.T(), s.defaultValue, s.userProperties.Get(s.defaultKey))
}

func (s *UserPropertiesTestSuite) TestToSortedSlice() {
	for i := 0; i < 5; i++ {
		keyToAdd := fmt.Sprintf("new-key-%v", i)
		valToAdd := fmt.Sprintf("new-val-%v", i)
		s.userProperties.Set(keyToAdd, valToAdd)
	}

	s.T().Run("with annotated default values", func(t *testing.T) {
		cupaloy.SnapshotT(t, s.userProperties.ToSortedSlice(true))
	})
	s.T().Run("without annotated default values", func(t *testing.T) {
		cupaloy.SnapshotT(t, s.userProperties.ToSortedSlice(false))
	})
	s.T().Run("with annotated empty default value", func(t *testing.T) {
		userPropertiesWithEmptyDefault := NewUserProperties(map[string]string{
			s.defaultKey:                 s.defaultValue,
			"default-key-with-empty-val": "",
		}, map[string]string{})
		cupaloy.SnapshotT(t, userPropertiesWithEmptyDefault.ToSortedSlice(true))
	})
	s.T().Run("with initial values", func(t *testing.T) {
		userPropertiesWithInitialValues := NewUserProperties(map[string]string{
			s.defaultKey: s.defaultValue,
		}, map[string]string{
			"initial.key.1": "value1",
			"initial.key.2": "value2",
		})
		cupaloy.SnapshotT(t, userPropertiesWithInitialValues.ToSortedSlice(true))
	})
}

func (s *UserPropertiesTestSuite) TestShouldOnlyReturnNonLocalNamespaceProperties() {
	s.userProperties.Set(config.KeyResultsTimeout, "1000")
	s.userProperties.Set(config.KeyCatalog, "test-catalog")

	require.Equal(s.T(), map[string]string{
		config.KeyCatalog: "test-catalog",
		"default-key":     "default-value",
	}, s.userProperties.GetNonLocalProperties())
}
