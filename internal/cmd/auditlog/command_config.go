package auditlog

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/confluentinc/go-editor"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
)

type configCommand struct {
	*cmd.AuthenticatedCLICommand
	prerunner cmd.PreRunner
}

// NewRouteCommand returns the sub-command object for interacting with audit log route rules.
func NewConfigCommand(prerunner cmd.PreRunner) *cobra.Command {
	cliCmd := cmd.NewAuthenticatedWithMDSCLICommand(
		&cobra.Command{
			Use:   "config",
			Short: "Manage the audit log configuration specification.",
			Long:  "Manage the audit log defaults and routing rules that determine which auditable events are logged, and where.",
		}, prerunner)
	command := &configCommand{
		AuthenticatedCLICommand: cliCmd,
		prerunner:               prerunner,
	}
	command.init()
	return command.Command
}

func (c *configCommand) init() {
	describeCmd := &cobra.Command{
		Use:   "describe",
		Short: "Prints the audit log configuration spec object.",
		Long:  `Prints the audit log configuration spec object, where "spec" refers to the JSON blob that describes audit log routing rules.`,
		RunE:  cmd.NewCLIRunE(c.describe),
		Args:  cobra.NoArgs,
	}
	c.AddCommand(describeCmd)

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Submits audit-log config spec object to the API.",
		Long:  "Submits an audit-log configuration specification JSON object to the API.",
		RunE:  cmd.NewCLIRunE(c.update),
		Args:  cobra.NoArgs,
	}
	updateCmd.Flags().String("file", "", "A local file path to the JSON configuration file, read as input. Otherwise the command will read from standard input.")
	updateCmd.Flags().Bool("force", false, "Updates the configuration, overwriting any concurrent modifications.")
	updateCmd.Flags().SortFlags = false
	c.AddCommand(updateCmd)

	editCmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit the audit-log config spec interactively.",
		Long:  "Edit the audit-log config spec object interactively, using the $EDITOR specified in your environment (for example, vim).",
		RunE:  cmd.NewCLIRunE(c.edit),
		Args:  cobra.NoArgs,
	}
	c.AddCommand(editCmd)
}

func (c *configCommand) createContext() context.Context {
	return context.WithValue(context.Background(), mds.ContextAccessToken, c.State.AuthToken)
}

func (c *configCommand) describe(cmd *cobra.Command, _ []string) error {
	spec, _, err := c.MDSClient.AuditLogConfigurationApi.GetConfig(c.createContext())
	if err != nil {
		return err
	}
	enc := json.NewEncoder(c.OutOrStdout())
	enc.SetIndent("", "  ")
	if err = enc.Encode(spec); err != nil {
		return err
	}
	return nil
}

func (c *configCommand) update(cmd *cobra.Command, _ []string) error {
	var data []byte
	var err error
	if cmd.Flags().Changed("file") {
		fileName, err := cmd.Flags().GetString("file")
		if err != nil {
			return err
		}
		data, err = ioutil.ReadFile(fileName)
		if err != nil {
			return err
		}
	} else {
		data, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
	}

	fileSpec := mds.AuditLogConfigSpec{}
	err = json.Unmarshal(data, &fileSpec)
	if err != nil {
		return err
	}
	putSpec := &fileSpec

	if cmd.Flags().Changed("force") {
		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}
		if force {
			gotSpec, response, err := c.MDSClient.AuditLogConfigurationApi.GetConfig(c.createContext())
			if err != nil {
				return HandleMdsAuditLogApiError(cmd, err, response)
			}
			putSpec = &mds.AuditLogConfigSpec{
				Destinations:       fileSpec.Destinations,
				ExcludedPrincipals: fileSpec.ExcludedPrincipals,
				DefaultTopics:      fileSpec.DefaultTopics,
				Routes:             fileSpec.Routes,
				Metadata: &mds.AuditLogConfigMetadata{
					ResourceVersion: gotSpec.Metadata.ResourceVersion,
				},
			}
		}
	}

	enc := json.NewEncoder(c.OutOrStdout())
	enc.SetIndent("", "  ")
	result, r, err := c.MDSClient.AuditLogConfigurationApi.PutConfig(c.createContext(), *putSpec)
	if err != nil {
		if r != nil && r.StatusCode == http.StatusConflict {
			if apiError, ok := err.(mds.GenericOpenAPIError); ok {
				if err2 := enc.Encode(apiError.Model()); err2 != nil {
					// We can just ignore this extra error. Why?
					// We expected a payload we could display as JSON, but got something unexpected.
					// That's OK though, we'll still handle and show the API error message.
				}
			}
		}
		return HandleMdsAuditLogApiError(cmd, err, r)
	}
	if err = enc.Encode(result); err != nil {
		return err
	}
	return nil
}

func (c *configCommand) edit(cmd *cobra.Command, _ []string) error {
	gotSpec, response, err := c.MDSClient.AuditLogConfigurationApi.GetConfig(c.createContext())
	if err != nil {
		return HandleMdsAuditLogApiError(cmd, err, response)
	}
	gotSpecBytes, err := json.MarshalIndent(gotSpec, "", "  ")
	if err != nil {
		return err
	}
	edit := editor.NewEditor()
	edited, path, err := edit.LaunchTempFile("audit-log", bytes.NewBuffer(gotSpecBytes))
	defer os.Remove(path)
	if err != nil {
		return err
	}
	putSpec := mds.AuditLogConfigSpec{}
	if err = json.Unmarshal(edited, &putSpec); err != nil {
		return err
	}
	enc := json.NewEncoder(c.OutOrStdout())
	enc.SetIndent("", "  ")
	result, r, err := c.MDSClient.AuditLogConfigurationApi.PutConfig(c.createContext(), putSpec)
	if err != nil {
		if r.StatusCode == http.StatusConflict {
			if err2 := enc.Encode(result); err2 != nil {
				// We can just ignore this extra error. Why?
				// We expected a payload we could display as JSON, but got something unexpected.
				// That's OK though, we'll still handle and show the API error message.
			}
		}
		return HandleMdsAuditLogApiError(cmd, err, r)
	}
	if err = enc.Encode(result); err != nil {
		return err
	}
	return nil
}
