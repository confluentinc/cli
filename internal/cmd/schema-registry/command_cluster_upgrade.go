package schemaregistry

import (
	"strings"

	"github.com/spf13/cobra"

	srcmv2 "github.com/confluentinc/ccloud-sdk-go-v2/srcm/v2"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newClusterUpgradeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "upgrade",
		Short:       "Upgrade the Schema Registry package for this environment.",
		Args:        cobra.NoArgs,
		RunE:        c.clusterUpgrade,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Upgrade Schema Registry to the "advanced" package for environment "env-12345".`,
				Code: "confluent schema-registry cluster upgrade --package advanced --environment env-12345",
			},
		),
	}

	addPackageFlag(cmd, "")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("package"))

	return cmd
}

func (c *command) clusterUpgrade(cmd *cobra.Command, _ []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	clusters, err := c.V2Client.GetSchemaRegistryClustersByEnvironment(environmentId)
	if err != nil {
		return err
	}

	if len(clusters) == 0 {
		return errors.NewSRNotEnabledError()
	}

	cluster := clusters[0]
	clusterSpec := cluster.GetSpec()

	packageDisplayName, err := cmd.Flags().GetString("package")
	if err != nil {
		return err
	}

	_, err = getPackageInternalName(packageDisplayName)
	if err != nil {
		return err
	}

	if strings.EqualFold(clusterSpec.GetPackage(), packageDisplayName) {
		output.ErrPrintf(errors.SRInvalidPackageUpgrade, environmentId, packageDisplayName)
		return nil
	}

	clusterUpdateRequest := createClusterUpdateRequest(packageDisplayName, environmentId)
	_, err = c.V2Client.UpgradeSchemaRegistryCluster(*clusterUpdateRequest, cluster.GetId())
	if err != nil {
		return err
	}

	output.Printf(errors.SchemaRegistryClusterUpgradedMsg, environmentId, packageDisplayName)
	return nil
}

func createClusterUpdateRequest(packageType, environmentId string) *srcmv2.SrcmV2ClusterUpdate {
	newClusterUpdateRequest := srcmv2.NewSrcmV2ClusterUpdateWithDefaults()
	spec := srcmv2.NewSrcmV2ClusterSpecUpdateWithDefaults()
	envObjectReference := srcmv2.NewGlobalObjectReferenceWithDefaults()
	envObjectReference.SetId(environmentId)

	spec.SetPackage(packageType)
	spec.SetEnvironment(*envObjectReference)
	newClusterUpdateRequest.SetSpec(*spec)

	return newClusterUpdateRequest
}
