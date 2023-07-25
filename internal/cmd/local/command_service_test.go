package local

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInjectConfigs(t *testing.T) {
	req := require.New(t)

	data := []byte("replace=old\n# replace=commented-duplicate\n# comment=old\n")

	config := map[string]string{
		"replace": "new",
		"comment": "new",
		"append":  "new",
	}

	data = injectConfig(data, config)

	req.Contains(string(data), "replace=new")
	req.Contains(string(data), "# replace=commented-duplicate")
	req.Contains(string(data), "comment=new")
	req.Contains(string(data), "append=new")
}

func TestInjectConfigsNoNewline(t *testing.T) {
	req := require.New(t)

	data := []byte("replace=old\n# replace=commented-duplicate\n# comment=old")

	config := map[string]string{
		"replace": "new",
		"append":  "new",
	}

	data = injectConfig(data, config)

	req.Contains(string(data), "replace=new")
	req.Contains(string(data), "# replace=commented-duplicate")
	req.Contains(string(data), "comment=old\n")
	req.Contains(string(data), "append=new")
}

func TestSetServiceEnvs(t *testing.T) {
	req := require.New(t)

	req.NoError(os.Setenv("KAFKA_LOG4J_OPTS", "saveme"))
	req.NoError(os.Setenv("CONNECT_LOG4J_OPTS", "useme"))

	req.NoError(setServiceEnvs("connect"))

	req.Equal("saveme", os.Getenv("SAVED_KAFKA_LOG4J_OPTS"))
	req.Equal("useme", os.Getenv("KAFKA_LOG4J_OPTS"))
}

func TestIsValidJavaVersion(t *testing.T) {
	req := require.New(t)

	var isValid bool
	var err error

	isValid, err = isValidJavaVersion("", "1.8.0_152")
	req.NoError(err)
	req.True(isValid)

	isValid, err = isValidJavaVersion("", "9.0.4")
	req.NoError(err)
	req.False(isValid)

	isValid, err = isValidJavaVersion("zookeeper", "13")
	req.NoError(err)
	req.True(isValid)

	isValid, err = isValidJavaVersion("", "13")
	req.NoError(err)
	req.False(isValid)
}
