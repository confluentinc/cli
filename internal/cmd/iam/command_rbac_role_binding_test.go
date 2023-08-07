package iam

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseAndValidateResourcePattern_Prefixed(t *testing.T) {
	pattern, err := parseAndValidateResourcePattern("Topic:test", true)
	require.NoError(t, err)
	require.Equal(t, "PREFIXED", pattern.PatternType)
}

func TestParseAndValidateResourcePattern_Literal(t *testing.T) {
	pattern, err := parseAndValidateResourcePattern("Topic:a", false)
	require.NoError(t, err)
	require.Equal(t, "LITERAL", pattern.PatternType)
}

func TestParseAndValidateResourcePattern_Topic(t *testing.T) {
	pattern, err := parseAndValidateResourcePattern("Topic:a", true)
	require.NoError(t, err)
	require.Equal(t, "Topic", pattern.ResourceType)
	require.Equal(t, "a", pattern.Name)
}

func TestParseAndValidateResourcePattern_TopicWithColon(t *testing.T) {
	pattern, err := parseAndValidateResourcePattern("Topic:a:b", true)
	require.NoError(t, err)
	require.Equal(t, "a:b", pattern.Name)
}

func TestParseAndValidateResourcePattern_ErrIncorrectResourceFormat(t *testing.T) {
	_, err := parseAndValidateResourcePattern("string with no colon", true)
	require.Error(t, err)
}
