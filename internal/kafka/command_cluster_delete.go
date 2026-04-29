package kafka

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

const (
	errorCodeDeletionProtectionEnabled = "deletion_protection_enabled"
	clusterDeletionProtectionDetail    = "Cluster deletion is blocked by deletion protection."
)

func (c *clusterCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete one or more Kafka clusters.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgsMultiple),
		RunE:              c.delete,
		Annotations:       map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *clusterCommand) delete(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	existenceFunc := func(id string) bool {
		_, _, err := c.V2Client.DescribeKafkaCluster(id, environmentId)
		return err == nil
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.KafkaCluster); err != nil {
		PluralClusterEnvironmentSuggestions := "Ensure the clusters you are specifying belong to the currently selected environment with `confluent kafka cluster list`, `confluent environment list`, and `confluent environment use`."
		return errors.NewErrorWithSuggestions(err.Error(), PluralClusterEnvironmentSuggestions)
	}

	deletionProtectionDetail := ""
	deleteFunc := func(id string) error {
		if httpResp, err := c.V2Client.DeleteKafkaCluster(id, environmentId); err != nil {
			if detail, ok := parseDeletionProtectionErrDetail(httpResp); ok {
				deletionProtectionDetail = detail
			}
			return errors.CatchKafkaNotFoundError(err, id, httpResp)
		}
		return nil
	}

	deletedIds, err := deletion.Delete(cmd, args, deleteFunc, resource.KafkaCluster)

	errs := multierror.Append(err, c.removeKafkaClusterConfigs(deletedIds))
	if errs.ErrorOrNil() != nil {
		if suggestion := deletionProtectionErrorToSuggestion(deletionProtectionDetail); suggestion != "" {
			return errors.NewErrorWithSuggestions(errs.Error(), suggestion)
		}
		if len(args)-len(deletedIds) > 1 {
			return errors.NewErrorWithSuggestions(errs.Error(),
				"Ensure the clusters are not associated with any active Connect clusters.")
		} else {
			return errors.NewErrorWithSuggestions(errs.Error(),
				"Ensure the cluster is not associated with any active Connect clusters.")
		}
	}

	return nil
}

type apiError struct {
	Code   string `json:"code"`
	Detail string `json:"detail"`
}

type apiErrorResponse struct {
	Errors []apiError `json:"errors"`
}

// parseDeletionProtectionErrDetail checks if the HTTP response indicates a deletion protection error
// Returns the error detail and true if the error is a deletion protection error, or empty string and false otherwise.
func parseDeletionProtectionErrDetail(r *http.Response) (string, bool) {
	if r == nil || r.StatusCode != http.StatusConflict || r.Body == nil {
		return "", false
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", false
	}
	// Restore the body so downstream handlers can read it
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	var res apiErrorResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return "", false
	}

	for _, apiErr := range res.Errors {
		if apiErr.Code == errorCodeDeletionProtectionEnabled {
			return apiErr.Detail, true
		}
	}

	return "", false
}

// deletionProtectionErrorToSuggestion maps a deletion protection error detail to a user-facing suggestion.
func deletionProtectionErrorToSuggestion(errorMsg string) string {
	switch {
	case strings.EqualFold(errorMsg, clusterDeletionProtectionDetail):
		return `Disable deletion_protection before deleting the cluster.`
	default:
		return ""
	}
}

func (c *clusterCommand) removeKafkaClusterConfigs(deletedIds []string) error {
	for _, id := range deletedIds {
		c.Context.KafkaClusterContext.RemoveKafkaCluster(id)
	}

	return c.Context.Save()
}
