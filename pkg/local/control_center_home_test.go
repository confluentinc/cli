package local

import (
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ControlCenterHomeTestSuite struct {
	suite.Suite
	c3h *ControlCenterHomeManager
}

func TestControlCenterHomeTestSuite(t *testing.T) {
	suite.Run(t, new(ControlCenterHomeTestSuite))
}

func (s *ControlCenterHomeTestSuite) SetupTest() {
	s.c3h = NewControlCenterHomeManager()
	dir, _ := createTestDir()
	os.Setenv("CONTROL_CENTER_HOME", dir)
}

func (s *ControlCenterHomeTestSuite) TearDownTest() {
	dir, _ := s.c3h.getRootDir()
	os.RemoveAll(dir)
	os.Clearenv()
}

func (s *ControlCenterHomeTestSuite) TestGetC3File() {
	req := require.New(s.T())

	dir, err := s.c3h.getRootDir()
	req.NoError(err)

	file, err := s.c3h.GetC3File(exampleDir, exampleFile)
	req.NoError(err)
	req.Equal(filepath.Join(dir, exampleDir, exampleFile), file)
}

func (s *ControlCenterHomeTestSuite) TestGetC3Script() {
	req := require.New(s.T())

	dir, err := s.c3h.getRootDir()
	req.NoError(err)

	file, err := s.c3h.GetServiceScriptC3("start", "prometheus")
	req.NoError(err)
	req.Equal(filepath.Join(dir, "bin/prometheus-start"), file)
}
