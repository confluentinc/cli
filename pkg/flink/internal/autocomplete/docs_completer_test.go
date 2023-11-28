package autocomplete

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/go-prompt"
)

type DocsCompleterTestSuite struct {
	suite.Suite
	completer prompt.Completer
}

func TestDocsCompleterTestSuite(t *testing.T) {
	suite.Run(t, new(DocsCompleterTestSuite))
}

func (s *DocsCompleterTestSuite) SetupSuite() {
	s.completer = GenerateDocsCompleter()
}

func (s *DocsCompleterTestSuite) TestSelectDocsAutoCompletionSnapshot() {
	input := "select spec FROM"
	buffer := prompt.NewBuffer()
	buffer.InsertText(input, false, true)

	actual := s.completer(*buffer.Document())

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *DocsCompleterTestSuite) TestCreateDocsAutoCompletionSnapshot() {
	input := "create"
	buffer := prompt.NewBuffer()
	buffer.InsertText(input, false, true)

	actual := s.completer(*buffer.Document())

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *DocsCompleterTestSuite) TestUseDocsAutoCompletionSnapshot() {
	input := "use"
	buffer := prompt.NewBuffer()
	buffer.InsertText(input, false, true)

	actual := s.completer(*buffer.Document())

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *DocsCompleterTestSuite) TestSetDocsAutoCompletionSnapshot() {
	input := "set"
	buffer := prompt.NewBuffer()
	buffer.InsertText(input, false, true)

	actual := s.completer(*buffer.Document())

	cupaloy.SnapshotT(s.T(), actual)
}
