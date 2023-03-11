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
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

func (c *schemaCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "delete",
		Short:       "Delete one or more schema versions.",
		Long:        "Delete one or more schema versions. This command should only be used if absolutely necessary.",
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

	cmd.Flags().String("subject", "", SubjectUsage)
	cmd.Flags().String("version", "", `Version of the schema. Can be a specific version, "all", or "latest".`)
	cmd.Flags().Bool("permanent", false, "Permanently delete the schema.")
	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddForceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	_ = cmd.MarkFlagRequired("subject")
	_ = cmd.MarkFlagRequired("version")

	return cmd
}

func (c *schemaCommand) delete(cmd *cobra.Command, _ []string) error {
	srClient, ctx, err := getApiClient(cmd, c.srClient, c.Config, c.Version)
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

	checkVersion := version
	if version == "all" {
		// check that at least one version for the input subject exists
		checkVersion = "latest"
	}
	_, httpResp, err := srClient.DefaultApi.GetSchemaByVersion(ctx, subject, checkVersion, nil)
	if err != nil {
		return errors.CatchSchemaNotFoundError(err, httpResp)
	}

	subjectWithVersion := fmt.Sprintf("%s (version %s)", subject, version)
	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmMsg, "schema", subjectWithVersion, subject)
	if _, err := form.ConfirmDeletionTypeCustomPrompt(cmd, promptMsg, subject); err != nil {
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

	var versions []int32
	if version == "all" {
		opts := &srsdk.DeleteSubjectOpts{Permanent: optional.NewBool(permanent)}
		v, httpResp, err := srClient.DefaultApi.DeleteSubject(ctx, subject, opts)
		if err != nil {
			return errors.CatchSchemaNotFoundError(err, httpResp)
		}
		output.Printf(errors.DeletedAllSubjectVersionMsg, deleteType, subject)
		versions = v
	} else {
		opts := &srsdk.DeleteSchemaVersionOpts{Permanent: optional.NewBool(permanent)}
		v, httpResp, err := srClient.DefaultApi.DeleteSchemaVersion(ctx, subject, version, opts)
		if err != nil {
			return errors.CatchSchemaNotFoundError(err, httpResp)
		}
		output.Printf(errors.DeletedSubjectVersionMsg, deleteType, version, subject)
		versions = []int32{v}
	}

	list := output.NewList(cmd)
	for _, version := range versions {
		list.Add(&versionOut{Version: version})
	}
	return list.Print()
}
