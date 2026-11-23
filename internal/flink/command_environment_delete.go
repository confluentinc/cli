package flink

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/plural"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

const (
	environmentCascadeSuggestions = "If the environment still has applications, statements or compute pools, add `--cascade` to delete them with it. This requires CMF 2.5.0 or later.\n" +
		"The environment's artifacts and catalog objects must always be deleted first.\n" +
		"If an earlier deletion is stuck, use `--cascade --skip-cleanup`, or `--cascade --force` if the Kubernetes cluster is disconnected or decommissioned."
	environmentForceSuggestions             = "If the Kubernetes cluster is disconnected or decommissioned, add `--force` to delete without requiring it."
	environmentCascadePermissionSuggestions = "With `--cascade`, you also need permission to delete every application, statement and compute pool in the environment."
)

func (c *command) newEnvironmentDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name-1> [name-2] ... [name-n]",
		Short: "Delete one or more Flink environments.",
		Long:  "Delete one or more Flink environments. An environment's artifacts, and the databases and other objects created in its catalog, must be deleted first. So must its applications, statements and compute pools, unless you add `--cascade`, which deletes them with the environment in the background, along with its savepoints and secret mappings. If its Kubernetes cluster is disconnected or decommissioned, add `--cascade --force` to remove them from CMF at once.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.environmentDelete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete Flink environment "my-env" with its applications, statements and compute pools, in the background.`,
				Code: "confluent flink environment delete my-env --cascade",
			},
			examples.Example{
				Text: `Unblock a deletion of Flink environment "my-env" that is stuck.`,
				Code: "confluent flink environment delete my-env --cascade --skip-cleanup",
			},
			examples.Example{
				Text: `Delete Flink environment "my-env" with its applications, statements and compute pools at once, when its Kubernetes cluster is disconnected or decommissioned.`,
				Code: "confluent flink environment delete my-env --cascade --force",
			},
		),
	}

	cmd.Flags().Bool("cascade", false, "Delete the environment's applications, statements, compute pools, savepoints and secret mappings with it, in the background. This stops all its Flink jobs. Requires CMF 2.5.0 or later.")
	cmd.Flags().Bool("skip-cleanup", false, "With --cascade, unblock a deletion that is stuck by no longer waiting for Kubernetes cleanup: no final savepoints, a statement's job may keep running, and savepoint files stay in storage. Use only if the deletion is stuck.")
	cmd.Flags().Bool("force", false, "Skip the confirmation prompt, and delete from CMF without requiring the Kubernetes cluster. Without --cascade, the environment's Kubernetes objects, such as its service account and secret, are left in its namespace. With --cascade, the deletion happens at once instead of in the background, Kubernetes objects are deleted on a best-effort basis, and savepoint files are kept.")
	addCmfFlagSet(cmd)

	// CMF rejects skipCleanup together with force. It also rejects skipCleanup without cascade, which Cobra can't express.
	cmd.MarkFlagsMutuallyExclusive("skip-cleanup", "force")

	return cmd
}

func (c *command) environmentDelete(cmd *cobra.Command, args []string) error {
	cascade, err := cmd.Flags().GetBool("cascade")
	if err != nil {
		return err
	}

	skipCleanup, err := cmd.Flags().GetBool("skip-cleanup")
	if err != nil {
		return err
	}

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	existenceFunc := func(name string) bool {
		_, err := client.DescribeEnvironment(c.createContext(), name)
		return err == nil
	}

	if err := deletion.ValidateAndConfirmWithExtraWarning(cmd, args, existenceFunc, resource.FlinkEnvironment, cascadeWarning(cascade, skipCleanup, len(args))); err != nil {
		// We are validating only the existence of the resources (there is no prefix validation). Thus we can
		// add some extra context for the error.
		suggestions := "List available Flink environments with `confluent flink environment list`."
		suggestions += "\nCheck that CMF is running and accessible."
		return errors.NewErrorWithSuggestions(err.Error(), suggestions)
	}

	// The HTTP statuses of CMF's refusals, which tell what flags could resolve them.
	refusals := map[int]bool{}
	deleteFunc := func(name string) error {
		result, status, err := client.DeleteEnvironment(c.createContext(), name, force, cascade, skipCleanup)
		if err != nil {
			refusals[status] = true
			return err
		}
		if cascade {
			printCascadeDeletion(name, status, result)
		}
		return nil
	}

	if cascade {
		_, err = deletion.DeleteWithoutMessage(cmd, args, deleteFunc)
	} else {
		_, err = deletion.Delete(cmd, args, deleteFunc, resource.FlinkEnvironment)
	}
	if err != nil {
		if suggestions := environmentDeleteSuggestions(refusals, cascade, force); suggestions != "" {
			// Only the suggestions of the returned error itself are shown, not those inside the multierror.
			return errors.NewErrorWithSuggestions(err.Error(), suggestions)
		}
	}
	return err
}

// environmentDeleteSuggestions points to the flags that could resolve CMF's refusals, judging by their HTTP statuses,
// because CMF's messages name its query parameters or don't mention the flags at all.
func environmentDeleteSuggestions(refusals map[int]bool, cascade, force bool) string {
	var suggestions []string
	if !cascade && (refusals[http.StatusBadRequest] || refusals[http.StatusUnprocessableEntity]) {
		suggestions = append(suggestions, environmentCascadeSuggestions)
	}
	// A disconnected or decommissioned cluster, or CMF restoring after a restart.
	if !force && refusals[http.StatusServiceUnavailable] {
		suggestions = append(suggestions, environmentForceSuggestions)
	}
	// CMF answers a missing permission with 401.
	if cascade && (refusals[http.StatusUnauthorized] || refusals[http.StatusForbidden]) {
		suggestions = append(suggestions, environmentCascadePermissionSuggestions)
	}
	return strings.Join(suggestions, "\n")
}

// cascadeWarning is added to the confirmation prompt of a delete with --cascade.
func cascadeWarning(cascade, skipCleanup bool, environments int) string {
	if !cascade {
		return ""
	}
	its := "its"
	if environments > 1 {
		its = "their"
	}
	if skipCleanup {
		return fmt.Sprintf(" This also deletes %[1]s applications, statements, compute pools, savepoints and secret mappings without waiting for Kubernetes to clean them up: no final savepoints, a statement's job may keep running, and savepoint files stay in storage.", its)
	}
	return fmt.Sprintf(" This also stops %[1]s Flink jobs and deletes %[1]s applications, statements, compute pools, savepoints and secret mappings.", its)
}

// printCascadeDeletion reports the outcome of a delete with --cascade, which depends on the HTTP status of CMF's response.
func printCascadeDeletion(name string, status int, result cmfsdk.EnvironmentDeletionResult) {
	switch status {
	case http.StatusAccepted:
		output.Println(false, fmt.Sprintf("Requested deletion of Flink environment \"%[1]s\" and everything in it%[2]s. The deletion runs in the background; until it finishes, nothing in \"%[1]s\" can be created or changed. To check whether it is still being deleted, run `confluent flink environment describe %[1]s`.", name, formatEnvironmentResources(result.GetToBeDeleted())))
	case http.StatusOK:
		deleted := result.GetDeleted()
		output.Println(false, fmt.Sprintf(`Deleted Flink environment "%s" and everything in it%s.`, name, formatEnvironmentResources(deleted)))
		warnAboutForceDeletionLeftovers(result.GetNotCleanedUp(), deleted.GetSavepoints())
	default:
		// CMF versions before 2.5.0 ignore cascade, and only delete an empty environment.
		output.Println(false, fmt.Sprintf(`Deleted Flink environment "%s".`, name))
	}
}

// formatEnvironmentResources counts each kind of resource in names, as in ": 2 applications, 1 statement and
// 1 secret mapping", or returns "" when names is empty.
func formatEnvironmentResources(names cmfsdk.EnvironmentResourceNames) string {
	kinds := []struct {
		noun  string
		count int
	}{
		{"application", len(names.GetApplications())},
		{"statement", len(names.GetStatements())},
		{"compute pool", len(names.GetComputePools())},
		{"savepoint", len(names.GetSavepoints())},
		{"secret mapping", len(names.GetSecretMappings())},
	}

	var counts []string
	for _, kind := range kinds {
		if kind.count == 1 {
			counts = append(counts, "1 "+kind.noun)
		} else if kind.count > 1 {
			counts = append(counts, fmt.Sprintf("%d %s", kind.count, plural.Plural(kind.noun)))
		}
	}

	switch len(counts) {
	case 0:
		return ""
	case 1:
		return ": " + counts[0]
	default:
		return ": " + strings.Join(counts[:len(counts)-1], ", ") + " and " + counts[len(counts)-1]
	}
}

// warnAboutForceDeletionLeftovers lists what a cascade deletion with force left behind: Kubernetes objects that might
// still exist, and the files of the deleted savepoints, which CMF doesn't delete.
func warnAboutForceDeletionLeftovers(notCleanedUp []cmfsdk.KubernetesCleanupFailure, savepoints []cmfsdk.SavepointReference) {
	if len(notCleanedUp) > 0 {
		output.ErrPrintln(false, "Warning: Kubernetes objects of these deleted resources might still exist. Delete them yourself if you still use the cluster:")
		for _, failure := range notCleanedUp {
			line := fmt.Sprintf(`    %s "%s"`, failure.GetKind(), failure.GetName())
			if namespace := failure.GetNamespace(); namespace != "" {
				line += fmt.Sprintf(` in namespace "%s"`, namespace)
			}
			output.ErrPrintln(false, line+": "+failure.GetReason())
		}
	}

	keptFiles := make([]string, 0, len(savepoints))
	for _, savepoint := range savepoints {
		if savepoint.GetPath() == "" {
			continue
		}
		owner := fmt.Sprintf(`application "%s"`, savepoint.GetApplication())
		if savepoint.GetStatement() != "" {
			owner = fmt.Sprintf(`statement "%s"`, savepoint.GetStatement())
		}
		keptFiles = append(keptFiles, fmt.Sprintf(`    "%s" of %s: %s`, savepoint.GetName(), owner, savepoint.GetPath()))
	}
	if len(keptFiles) > 0 {
		output.ErrPrintln(false, "Warning: savepoint files were kept:")
		for _, line := range keptFiles {
			output.ErrPrintln(false, line)
		}
	}
}
