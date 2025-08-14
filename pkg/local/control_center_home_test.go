package local

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
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
	fmt.Println(filepath.Join(dir, exampleDir, exampleFile))

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

func (s *ControlCenterHomeTestSuite) TestReadServiceConfigC3() {
	req := require.New(s.T())
	dir, err := s.c3h.getRootDir()
	req.NoError(err)

	path := filepath.Join(dir, "etc/confluent-control-center/prometheus-generated-local.yml")
	err = os.MkdirAll(filepath.Dir(path), 0777)
	if err != nil {
		return
	}
	b := []byte{'h', 'e', 'l', 'l', 'o'}
	err = os.WriteFile(path, b, 0644)
	if err != nil {
		return
	}
	req.NoError(err)
	config, err := s.c3h.ReadServiceConfigC3("prometheus")
	req.NoError(err)
	req.Equal([]byte{'h', 'e', 'l', 'l', 'o'}, config)
}

func (s *ControlCenterHomeTestSuite) TestReadPortC3() {
	req := require.New(s.T())
	dir, err := s.c3h.getRootDir()
	req.NoError(err)

	path := filepath.Join(dir, "etc/confluent-control-center/prometheus-generated-local.yml")
	err = os.MkdirAll(filepath.Dir(path), 0777)
	if err != nil {
		return
	}
	b := []byte{'h', 'e', 'l', 'l', 'o'}
	err = os.WriteFile(path, b, 0644)
	if err != nil {
		return
	}
	req.NoError(err)
	_, err = s.c3h.ReadServicePortC3("prometheus", false)
	req.Error(err, "no port specified")
}
