package update

import (
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/require"
)

func TestFilterUpdates(t *testing.T) {
	content := `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 3.2 Final//EN">
<html>
  <head>
    <title>Index of confluent-cli/binaries</title>
  </head>
  <body>
    <h1>Index of confluent-cli/binaries</h1>
    <hr/>
    <pre>
      <a href="/confluent-cli/">../</a>
      <a href="/confluent-cli/binaries/4.6.0/">4.6.0/</a>
      <a href="/confluent-cli/binaries/4.7.0/">4.7.0/</a>
    </pre>
    <hr/>
  </body>
</html>`

	minorVersions, majorVersions := FilterUpdates(content, version.Must(version.NewVersion("4.6.0")))
	require.Equal(t, 1, len(minorVersions))
	require.Equal(t, 0, len(majorVersions))
}
