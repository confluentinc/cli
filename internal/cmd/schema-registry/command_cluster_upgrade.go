package schemaregistry

import (
	"fmt"
	"strings"

	srcm "github.com/confluentinc/ccloud-sdk-go-v2/srcm/v2"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/version"
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
				Code: fmt.Sprintf("%s schema-registry cluster upgrade --package advanced --environment env-12345", version.CLIName),
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
	environmentId, err := c.EnvironmentId()
	if err != nil {
		return err
	}

	clusterList, _, err := c.V2Client.GetSchemaRegistryClusterByEnvironment(environmentId)
	if err != nil {
		return err
	}
	if len(clusterList.Data) == 0 {
		return errors.NewSRNotEnabledError()
	}
	cluster := clusterList.GetData()[0]
	clusterSpec := cluster.GetSpec()

	packageToUpgradeTo, err := cmd.Flags().GetString("package")
	if err != nil {
		return err
	}

	if strings.ToLower(clusterSpec.GetPackage()) == strings.ToLower(packageToUpgradeTo) {
		output.ErrPrintf(errors.SRInvalidPackageUpgrade, environmentId, packageToUpgradeTo)
		return nil
	}

	clusterUpdateRequest := createClusterUpdateRequest(packageToUpgradeTo, environmentId)
	_, _, err = c.V2Client.UpgradeSchemaRegistryCluster(*clusterUpdateRequest, cluster.GetId())
	if err != nil {
		return err
	}
	output.Printf(errors.SchemaRegistryClusterUpgradedMsg, environmentId, packageToUpgradeTo)
	return nil
}

func createClusterUpdateRequest(packageType, environmentId string) *srcm.SrcmV2ClusterUpdate {
	newClusterUpdateRequest := srcm.NewSrcmV2ClusterUpdateWithDefaults()
	spec := srcm.NewSrcmV2ClusterSpecUpdateWithDefaults()
	envObjectReference := srcm.NewGlobalObjectReferenceWithDefaults()
	envObjectReference.SetId(environmentId)

	spec.SetPackage(packageType)
	spec.SetEnvironment(*envObjectReference)
	newClusterUpdateRequest.SetSpec(*spec)

	return newClusterUpdateRequest
}
