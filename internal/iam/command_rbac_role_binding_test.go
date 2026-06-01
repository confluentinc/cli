package iam

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
)

func TestParseAndValidateResourcePattern_Prefixed(t *testing.T) {
	pattern, err := parseAndValidateResourcePattern("Topic:test", true)
	require.NoError(t, err)
	require.Equal(t, "PREFIXED", pattern.PatternType)
}

func TestParseAndValidateResourcePattern_Literal(t *testing.T) {
	pattern, err := parseAndValidateResourcePattern("Topic:a", false)
	require.NoError(t, err)
	require.Equal(t, "LITERAL", pattern.PatternType)
}

func TestParseAndValidateResourcePattern_Topic(t *testing.T) {
	pattern, err := parseAndValidateResourcePattern("Topic:a", true)
	require.NoError(t, err)
	require.Equal(t, "Topic", pattern.ResourceType)
	require.Equal(t, "a", pattern.Name)
}

func TestParseAndValidateResourcePattern_TopicWithColon(t *testing.T) {
	pattern, err := parseAndValidateResourcePattern("Topic:a:b", true)
	require.NoError(t, err)
	require.Equal(t, "a:b", pattern.Name)
}

func TestParseAndValidateResourcePattern_ErrIncorrectResourceFormat(t *testing.T) {
	_, err := parseAndValidateResourcePattern("string with no colon", true)
	require.Error(t, err)
}

// newRoleBindingTestCommand returns a roleBindingCommand with a minimal context that only
// supplies the current organization, which is all parseV2BaseCrnPattern reads from context.
func newRoleBindingTestCommand() *roleBindingCommand {
	return &roleBindingCommand{
		AuthenticatedCLICommand: &pcmd.AuthenticatedCLICommand{
			Context: &config.Context{LastOrgId: "abc-123"},
		},
	}
}

// newCloudRoleBindingFlagSet registers the cloud scope flags that parseV2BaseCrnPattern reads.
func newCloudRoleBindingFlagSet() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("role", "", "")
	cmd.Flags().String("environment", "", "")
	cmd.Flags().Bool("current-environment", false, "")
	cmd.Flags().String("cloud-cluster", "", "")
	cmd.Flags().String("schema-registry-cluster", "", "")
	cmd.Flags().String("ksql-cluster", "", "")
	cmd.Flags().String("kafka-cluster", "", "")
	cmd.Flags().String("flink-region", "", "")
	cmd.Flags().String("usm-kafka-cluster", "", "")
	cmd.Flags().String("usm-connect-cluster", "", "")
	return cmd
}

func TestParseV2BaseCrnPattern_UsmKafkaCluster(t *testing.T) {
	cmd := newCloudRoleBindingFlagSet()
	require.NoError(t, cmd.Flags().Set("role", "UsmKafkaClusterAdmin"))
	require.NoError(t, cmd.Flags().Set("environment", "env-596"))
	require.NoError(t, cmd.Flags().Set("usm-kafka-cluster", "usmkc-123456"))

	crnPattern, err := newRoleBindingTestCommand().parseV2BaseCrnPattern(cmd)
	require.NoError(t, err)
	require.Equal(t, "crn://confluent.cloud/organization=abc-123/environment=env-596/usm-kafka-cluster=usmkc-123456", crnPattern)
}

func TestParseV2BaseCrnPattern_UsmConnectCluster(t *testing.T) {
	cmd := newCloudRoleBindingFlagSet()
	require.NoError(t, cmd.Flags().Set("role", "UsmConnectClusterAdmin"))
	require.NoError(t, cmd.Flags().Set("environment", "env-596"))
	require.NoError(t, cmd.Flags().Set("usm-connect-cluster", "usmcc-123456"))

	crnPattern, err := newRoleBindingTestCommand().parseV2BaseCrnPattern(cmd)
	require.NoError(t, err)
	require.Equal(t, "crn://confluent.cloud/organization=abc-123/environment=env-596/usm-connect-cluster=usmcc-123456", crnPattern)
}

func TestParseV2BaseCrnPattern_UsmKafkaRolesRequireClusterFlag(t *testing.T) {
	for _, role := range []string{"UsmKafkaClusterAdmin", "UsmKafkaOperator", "UsmKafkaMetricsViewer"} {
		cmd := newCloudRoleBindingFlagSet()
		require.NoError(t, cmd.Flags().Set("role", role))
		require.NoError(t, cmd.Flags().Set("environment", "env-596"))

		_, err := newRoleBindingTestCommand().parseV2BaseCrnPattern(cmd)
		require.EqualError(t, err, specifyUsmKafkaClusterErrorMsg, "role %q must require --usm-kafka-cluster", role)
	}
}

func TestParseV2BaseCrnPattern_UsmConnectRolesRequireClusterFlag(t *testing.T) {
	for _, role := range []string{"UsmConnectClusterAdmin", "UsmConnectOperator", "UsmConnectMetricsViewer"} {
		cmd := newCloudRoleBindingFlagSet()
		require.NoError(t, cmd.Flags().Set("role", role))
		require.NoError(t, cmd.Flags().Set("environment", "env-596"))

		_, err := newRoleBindingTestCommand().parseV2BaseCrnPattern(cmd)
		require.EqualError(t, err, specifyUsmConnectClusterErrorMsg, "role %q must require --usm-connect-cluster", role)
	}
}

func TestParseV2BaseCrnPattern_UsmRoleRequiresEnvironment(t *testing.T) {
	cmd := newCloudRoleBindingFlagSet()
	require.NoError(t, cmd.Flags().Set("role", "UsmKafkaClusterAdmin"))
	require.NoError(t, cmd.Flags().Set("usm-kafka-cluster", "usmkc-123456"))

	_, err := newRoleBindingTestCommand().parseV2BaseCrnPattern(cmd)
	require.EqualError(t, err, specifyEnvironmentErrorMsg)
}

func TestParseV2BaseCrnPattern_UsmClusterFlagRequiresEnvironment(t *testing.T) {
	cmd := newCloudRoleBindingFlagSet()
	require.NoError(t, cmd.Flags().Set("usm-connect-cluster", "usmcc-123456"))

	_, err := newRoleBindingTestCommand().parseV2BaseCrnPattern(cmd)
	require.EqualError(t, err, specifyEnvironmentErrorMsg)
}
