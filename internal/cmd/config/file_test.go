package config

// Example unit tests for the example command config test
// Unit tests are for testing internal methods
// See apikey_test.go for a good example

import (
	"github.com/stretchr/testify/suite"
)

type ConfigFileTestSuite struct {
	suite.Suite // from "github.com/stretchr/testify/suite" - can use a suite if I want (there are mocks to set up)
	// --include mocks
}
