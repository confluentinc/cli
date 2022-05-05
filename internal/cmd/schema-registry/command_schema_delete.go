package schemaregistry

import (
	"context"
	"fmt"

	"github.com/antihax/optional"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

func (c *schemaCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "delete",
		Short:       "Delete one or more schemas.",
		Long:        "Delete one or more schemas. This command should only be used if absolutely necessary.",
		Args:        cobra.NoArgs,
		RunE:        c.delete,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Soft delete the latest version of subject "payments".`,
				Code: fmt.Sprintf("%s schema-registry schema delete --subject payments --version latest", pversion.CLIName),
			},
		),
	}

	cmd.Flags().StringP("subject", "S", "", SubjectUsage)
	cmd.Flags().StringP("version", "V", "", `Version of the schema. Can be a specific version, "all", or "latest".`)
	cmd.Flags().BoolP("permanent", "P", false, "Permanently delete the schema.")
	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	_ = cmd.MarkFlagRequired("subject")
	_ = cmd.MarkFlagRequired("version")

	return cmd
}

func (c *schemaCommand) delete(cmd *cobra.Command, _ []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	return deleteSchema(cmd, srClient, ctx)
}

func deleteSchema(cmd *cobra.Command, srClient *srsdk.APIClient, ctx context.Context) error {
	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}

	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return err
	}

	permanent, err := cmd.Flags().GetBool("permanent")
	if err != nil {
		return err
	}

	deleteType := "soft"
	if permanent {
		deleteType = "hard"
	}
	if version == "all" {
		deleteSubjectOpts := &srsdk.DeleteSubjectOpts{Permanent: optional.NewBool(permanent)}
		versions, httpResp, err := srClient.DefaultApi.DeleteSubject(ctx, subject, deleteSubjectOpts)
		if err != nil {
			return errors.CatchSchemaNotFoundError(err, httpResp)
		}
		utils.Printf(cmd, errors.DeletedAllSubjectVersionMsg, deleteType, subject)
		printVersions(versions)
		return nil
	} else {
		deleteVersionOpts := &srsdk.DeleteSchemaVersionOpts{Permanent: optional.NewBool(permanent)}
		versionResult, httpResp, err := srClient.DefaultApi.DeleteSchemaVersion(ctx, subject, version, deleteVersionOpts)
		if err != nil {
			return errors.CatchSchemaNotFoundError(err, httpResp)
		}
		utils.Printf(cmd, errors.DeletedSubjectVersionMsg, deleteType, version, subject)
		printVersions([]int32{versionResult})
		return nil
	}
}
