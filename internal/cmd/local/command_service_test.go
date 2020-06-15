package local

import (
	"fmt"
	"net"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/mock"
)

const exampleService = "kafka"

func TestConfigService(t *testing.T) {
	req := require.New(t)

	ch := &mock.MockConfluentHome{
		GetConfigFunc: func(service string) ([]byte, error) {
			return []byte("replace=old\n# comment=old\n"), nil
		},
	}

	cc := &mock.MockConfluentCurrent{
		SetConfigFunc: func(service string, config []byte) error {
			req.NotContains(string(config), "replace=old")
			req.Contains(string(config), "replace=new")
			req.NotContains(string(config), "# comment=old")
			req.Contains(string(config), "comment=new")
			req.Contains(string(config), "append=new")
			return nil
		},
	}

	config := map[string]string{"replace": "new", "comment": "new", "append": "new"}
	req.NoError(configService(ch, cc, exampleService, config))
}

func TestIsRunning(t *testing.T) {
	req := require.New(t)

	cat := exec.Command("cat")
	req.NoError(cat.Start())
	defer cat.Process.Kill()

	cc := &mock.MockConfluentCurrent{
		HasPidFileFunc: func(service string) (bool, error) {
			return true, nil
		},
		GetPidFunc: func(service string) (int, error) {
			return cat.Process.Pid, nil
		},
	}

	isUp, err := isRunning(cc, exampleService)
	req.NoError(err)
	req.True(isUp)
}

func TestIsNotRunning(t *testing.T) {
	req := require.New(t)

	cc := &mock.MockConfluentCurrent{
		HasPidFileFunc: func(service string) (bool, error) {
			return false, nil
		},
	}

	isUp, err := isRunning(cc, exampleService)
	req.NoError(err)
	req.False(isUp)
}

func TestIsPortOpen(t *testing.T) {
	req := require.New(t)

	addr := fmt.Sprintf(":%d", services[exampleService].port)
	lis, err := net.Listen("tcp", addr)
	req.NoError(err)
	defer lis.Close()

	isOpen, err := isPortOpen(exampleService)
	req.NoError(err)
	req.True(isOpen)
}

func TestIsPortClosed(t *testing.T) {
	req := require.New(t)

	isOpen, err := isPortOpen(exampleService)
	req.NoError(err)
	req.False(isOpen)
}
