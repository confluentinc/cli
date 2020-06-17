package local

import (
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
