package kafka

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommentAndWarnAboutSr(t *testing.T) {
	t.Parallel()

	// comments should be at the beginning of the line
	original := "# Required connection configs for Confluent Cloud Schema Registry\n" +
		"schema.registry.url=https://{{ SR_ENDPOINT }}\n" +
		"basic.auth.credentials.source=USER_INFO\n" +
		"basic.auth.user.info={{ SR_API_KEY }}:{{ SR_API_SECRET }}\n"
	commented, err := commentAndWarnAboutSchemaRegistry("my-reason", "my-suggestions", original)
	require.NoError(t, err)
	require.Equal(t, "# Required connection configs for Confluent Cloud Schema Registry\n"+
		"#schema.registry.url=https://{{ SR_ENDPOINT }}\n"+
		"#basic.auth.credentials.source=USER_INFO\n"+
		"#basic.auth.user.info={{ SR_API_KEY }}:{{ SR_API_SECRET }}\n", string(commented))

	// comments should be right before each property, not the beginning of the line
	original = "  properties {\n" +
		"    # Required connection configs for Confluent Cloud Schema Registry\n" +
		"    schema.registry.url = \"https://{{ SR_ENDPOINT }}\"\n" +
		"    basic.auth.credentials.source = USER_INFO\n" +
		"    basic.auth.user.info = \"{{ SR_API_KEY }}:{{ SR_API_SECRET }}\"\n" +
		"  }"
	commented, err = commentAndWarnAboutSchemaRegistry("my-reason", "my-suggestions", original)
	require.NoError(t, err)
	require.Equal(t, "  properties {\n"+
		"    # Required connection configs for Confluent Cloud Schema Registry\n"+
		"    #schema.registry.url = \"https://{{ SR_ENDPOINT }}\"\n"+
		"    #basic.auth.credentials.source = USER_INFO\n"+
		"    #basic.auth.user.info = \"{{ SR_API_KEY }}:{{ SR_API_SECRET }}\"\n"+
		"  }", string(commented))
}
