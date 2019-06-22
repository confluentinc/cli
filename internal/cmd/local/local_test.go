package local

import (
	"errors"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	cliMock "github.com/confluentinc/cli/mock"
	mock_local "github.com/confluentinc/cli/mock/local"
)

func TestLocal(t *testing.T) {
	oldCurrent := os.Getenv("CONFLUENT_CURRENT")
	_ = os.Setenv("CONFLUENT_CURRENT", "/path/to/confluent/workdir")
	defer func() { _ = os.Setenv("CONFLUENT_CURRENT", oldCurrent) }()

	oldTmp := os.Getenv("TMPDIR")
	_ = os.Setenv("TMPDIR", "/var/folders/some/junk")
	defer func() { _ = os.Setenv("TMPDIR", oldTmp) }()

	req := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shellRunner := mock_local.NewMockShellRunner(ctrl)
	shellRunner.EXPECT().Init(os.Stdout, os.Stderr)
	shellRunner.EXPECT().Export("CONFLUENT_CURRENT", "/path/to/confluent/workdir")
	shellRunner.EXPECT().Export("CONFLUENT_HOME", "blah")
	shellRunner.EXPECT().Export("TMPDIR", "/var/folders/some/junk")
	shellRunner.EXPECT().Source("cp_cli/confluent.sh", gomock.Any())
	shellRunner.EXPECT().Run("main", gomock.Eq([]string{"local", "help"})).Return(0, nil)
	localCmd := New(&cliMock.Commander{}, shellRunner)
	_, err := cmd.ExecuteCommand(localCmd, "local", "--path", "blah", "help")
	req.NoError(err)
}

func TestLocalErrorDuringSource(t *testing.T) {
	oldCurrent := os.Getenv("CONFLUENT_CURRENT")
	_ = os.Setenv("CONFLUENT_CURRENT", "/path/to/confluent/workdir")
	defer func() { _ = os.Setenv("CONFLUENT_CURRENT", oldCurrent) }()

	oldTmp := os.Getenv("TMPDIR")
	_ = os.Setenv("TMPDIR", "/var/folders/some/junk")
	defer func() { _ = os.Setenv("TMPDIR", oldTmp) }()

	req := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shellRunner := mock_local.NewMockShellRunner(ctrl)
	shellRunner.EXPECT().Init(os.Stdout, os.Stderr)
	shellRunner.EXPECT().Export("CONFLUENT_CURRENT", "/path/to/confluent/workdir")
	shellRunner.EXPECT().Export("CONFLUENT_HOME", "blah")
	shellRunner.EXPECT().Export("TMPDIR", "/var/folders/some/junk")
	shellRunner.EXPECT().Source("cp_cli/confluent.sh", gomock.Any()).Return(errors.New("oh no"))
	localCmd := New(&cliMock.Commander{}, shellRunner)
	_, err := cmd.ExecuteCommand(localCmd, "local", "--path", "blah", "help")
	req.Error(err)
}
