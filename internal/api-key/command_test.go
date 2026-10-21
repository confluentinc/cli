package apikey

import (
	"fmt"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	apikeysv2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func TestGetResourceType(t *testing.T) {
	require.Equal(t, "cloud", getResourceType(apikeysv2.TypedEnvScopedObjectReference{Kind: apikeysv2.PtrString("Cloud")}))
	require.Equal(t, "global", getResourceType(apikeysv2.TypedEnvScopedObjectReference{Kind: apikeysv2.PtrString("Global")}))
	require.Equal(t, "flink-region", getResourceType(apikeysv2.TypedEnvScopedObjectReference{ApiVersion: apikeysv2.PtrString("fcpm/v2"), Kind: apikeysv2.PtrString("Region")}))
	require.Equal(t, "kafka", getResourceType(apikeysv2.TypedEnvScopedObjectReference{ApiVersion: apikeysv2.PtrString("cmk/v2"), Kind: apikeysv2.PtrString("Cluster")}))
	require.Equal(t, "ksql", getResourceType(apikeysv2.TypedEnvScopedObjectReference{Kind: apikeysv2.PtrString("ksqlDB")}))
	require.Equal(t, "tableflow", getResourceType(apikeysv2.TypedEnvScopedObjectReference{Kind: apikeysv2.PtrString("Tableflow")}))
	require.Equal(t, "schema-registry", getResourceType(apikeysv2.TypedEnvScopedObjectReference{Kind: apikeysv2.PtrString("SchemaRegistry")}))
}

func newCreateFlagsCommand() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("expiration", "", "")
	cmd.Flags().String("service-account", "", "")
	cmd.Flags().String("cloud", "", "")
	cmd.Flags().String("region", "", "")
	return cmd
}

func TestGetExpiration(t *testing.T) {
	cmd := newCreateFlagsCommand()

	expiration, err := getExpiration(cmd)
	require.NoError(t, err)
	require.Empty(t, expiration)

	require.NoError(t, cmd.Flags().Set("expiration", "2026-12-31"))
	expiration, err = getExpiration(cmd)
	require.NoError(t, err)
	require.Equal(t, "2026-12-31", expiration)

	require.NoError(t, cmd.Flags().Set("expiration", "12/31/2026"))
	_, err = getExpiration(cmd)
	require.EqualError(t, err, `invalid expiration date "12/31/2026"`)

	// An explicitly empty value (e.g. an unset shell variable) is rejected, not treated as "never expires".
	require.NoError(t, cmd.Flags().Set("expiration", ""))
	_, err = getExpiration(cmd)
	require.EqualError(t, err, `invalid expiration date ""`)

	_, err = getExpiration(&cobra.Command{})
	require.Error(t, err)
}

func TestGetOwner(t *testing.T) {
	c := &command{AuthenticatedCLICommand: &pcmd.AuthenticatedCLICommand{Context: &config.Context{}}}

	cmd := newCreateFlagsCommand()
	require.NoError(t, cmd.Flags().Set("service-account", "sa-123456"))
	serviceAccount, ownerId, err := c.getOwner(cmd)
	require.NoError(t, err)
	require.Equal(t, "sa-123456", serviceAccount)
	require.Equal(t, "sa-123456", ownerId)

	_, _, err = c.getOwner(&cobra.Command{})
	require.Error(t, err)
}

func TestBuildResourceRef(t *testing.T) {
	c := &command{AuthenticatedCLICommand: &pcmd.AuthenticatedCLICommand{Context: &config.Context{CurrentEnvironment: "env-123456"}}}
	cmd := newCreateFlagsCommand()

	for resourceType, expected := range map[string]string{
		resource.Cloud:        resource.Cloud,
		resource.Global:       resource.Global,
		resource.Tableflow:    resource.Tableflow,
		resource.KafkaCluster: "lkc-123456",
	} {
		ref, err := c.buildResourceRef(cmd, resourceType, "lkc-123456")
		require.NoError(t, err)
		require.Equal(t, expected, ref.Id)
	}

	_, err := c.buildResourceRef(cmd, resource.Flink, "")
	require.EqualError(t, err, "must provide both `--cloud` and `--region`")

	require.NoError(t, cmd.Flags().Set("cloud", "aws"))
	require.NoError(t, cmd.Flags().Set("region", "us-east-1"))
	ref, err := c.buildResourceRef(cmd, resource.Flink, "")
	require.NoError(t, err)
	require.Equal(t, "aws.us-east-1", ref.Id)
	require.Equal(t, "env-123456", ref.GetEnvironment())

	_, err = c.buildResourceRef(&cobra.Command{}, resource.Flink, "")
	require.Error(t, err)

	noEnv := &command{AuthenticatedCLICommand: &pcmd.AuthenticatedCLICommand{Context: &config.Context{}}}
	_, err = noEnv.buildResourceRef(cmd, resource.Flink, "")
	require.Error(t, err)
}

func TestUseCreatedKey_UnsupportedResource(t *testing.T) {
	c := &command{AuthenticatedCLICommand: &pcmd.AuthenticatedCLICommand{}}
	err := c.useCreatedKey("KEY", resource.Cloud, resource.Cloud)
	require.EqualError(t, err, "`--use` set but ineffective: "+nonKafkaNotImplementedErrorMsg)
}

func TestValidateGlobalKeyResource(t *testing.T) {
	require.NoError(t, validateGlobalKeyResource("KEY", resource.Global, true))
	require.NoError(t, validateGlobalKeyResource("KEY", "", true))
	require.Error(t, validateGlobalKeyResource("KEY", resource.Global, false))
	require.Error(t, validateGlobalKeyResource("KEY", resource.KafkaCluster, true))
}

func TestGetStoreTargetCluster(t *testing.T) {
	cluster := &config.KafkaClusterConfig{ID: "lkc-123456", Bootstrap: "pkc-123456:9092", LastUpdate: time.Now()}
	c := &command{AuthenticatedCLICommand: &pcmd.AuthenticatedCLICommand{Context: &config.Context{
		KafkaClusterContext: &config.KafkaClusterContext{
			ActiveKafkaCluster:  "lkc-123456",
			KafkaClusterConfigs: map[string]*config.KafkaClusterConfig{"lkc-123456": cluster},
		},
	}}}

	// --resource resolves to a Kafka cluster
	got, err := c.getStoreTargetCluster(resource.KafkaCluster, "lkc-123456", nil)
	require.NoError(t, err)
	require.Equal(t, cluster, got)

	// --resource resolves to a non-Kafka resource
	_, err = c.getStoreTargetCluster(resource.SchemaRegistryCluster, "lsrc-123456", nil)
	require.EqualError(t, err, nonKafkaNotImplementedErrorMsg)

	// --resource failed to resolve: fall back to the active Kafka cluster
	got, err = c.getStoreTargetCluster("", "", fmt.Errorf("unable to resolve"))
	require.NoError(t, err)
	require.Equal(t, cluster, got)

	// no active Kafka cluster
	noCluster := &command{AuthenticatedCLICommand: &pcmd.AuthenticatedCLICommand{Context: &config.Context{}}}
	_, err = noCluster.getStoreTargetCluster("", "", nil)
	require.EqualError(t, err, errors.NoKafkaSelectedErrorMsg)
}
