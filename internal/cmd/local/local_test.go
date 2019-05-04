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
	req := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shellRunner := mock_local.NewMockShellRunner(ctrl)
	shellRunner.EXPECT().Init(os.Stdout, os.Stderr)
	shellRunner.EXPECT().Export("CONFLUENT_HOME", "blah")
	shellRunner.EXPECT().Source("cp_cli/confluent.sh", gomock.Any())
	shellRunner.EXPECT().Run("main", gomock.Eq([]string{"local", "help"})).Return(0, nil)
	localCmd := New(&cliMock.Commander{}, shellRunner)
	_, err := cmd.ExecuteCommand(localCmd, "local", "--path", "blah", "help")
	req.NoError(err)
}

func TestLocalErrorDuringSource(t *testing.T) {
	req := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shellRunner := mock_local.NewMockShellRunner(ctrl)
	shellRunner.EXPECT().Init(os.Stdout, os.Stderr)
	shellRunner.EXPECT().Export("CONFLUENT_HOME", "blah")
	shellRunner.EXPECT().Source("cp_cli/confluent.sh", gomock.Any()).Return(errors.New("oh no"))
	localCmd := New(&cliMock.Commander{}, shellRunner)
	_, err := cmd.ExecuteCommand(localCmd, "local", "--path", "blah", "help")
	req.Error(err)
}
