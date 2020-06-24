package auditlog

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/go-editor"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"

	"net/http"
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
			Short: "Manage audit log configuration specification (since 6.0).",
			Long:  "Manage audit log defaults and routing rules that determine which auditable events are logged, and where (since 6.0).",
		}, prerunner)
	cmd := &configCommand{
		AuthenticatedCLICommand: cliCmd,
		prerunner:               prerunner,
	}
	cmd.init()
	return cmd.Command
}

func (c *configCommand) init() {
	describeCmd := &cobra.Command{
		Use:   "describe",
		Short: "Prints the audit log configuration spec object (since 6.0).",
		RunE:  c.describe,
		Args:  cobra.NoArgs,
	}
	c.AddCommand(describeCmd)

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Submits audit-log config spec object to the API (since 6.0).",
		Long:  "Submits an audit-log configuration specification JSON object to the API (since 6.0).",
		RunE:  c.update,
		Args:  cobra.NoArgs,
	}
	updateCmd.Flags().String("file", "", "A local file path, read as input. Otherwise the command will read from standard in.")
	updateCmd.Flags().Bool("force", false, "Tries to update even with concurrent modifications.")
	updateCmd.Flags().SortFlags = false
	c.AddCommand(updateCmd)

	editCmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit the audit-log config spec interactively (since 6.0).",
		Long:  "Edit the audit-log config spec object interactively, using the EDITOR specified in your env (since 6.0).",
		RunE:  c.edit,
		Args:  cobra.NoArgs,
	}
	c.AddCommand(editCmd)
}

func (c *configCommand) createContext() context.Context {
	return context.WithValue(context.Background(), mds.ContextAccessToken, c.State.AuthToken)
}

func (c *configCommand) describe(cmd *cobra.Command, args []string) error {
	spec, _, err := c.MDSClient.AuditLogConfigurationApi.GetConfig(c.createContext())
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	enc := json.NewEncoder(c.OutOrStdout())
	enc.SetIndent("", "  ")
	if err = enc.Encode(spec); err != nil {
		return errors.HandleCommon(err, cmd)
	}
	return nil
}

func (c *configCommand) update(cmd *cobra.Command, args []string) error {
	var data []byte
	var err error
	if cmd.Flags().Changed("file") {
		fileName, err := cmd.Flags().GetString("file")
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		data, err = ioutil.ReadFile(fileName)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
	} else {
		data, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
	}

	fileSpec := mds.AuditLogConfigSpec{}
	err = json.Unmarshal(data, &fileSpec)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	putSpec := &fileSpec

	if cmd.Flags().Changed("force") {
		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		if force {
			gotSpec, _, err := c.MDSClient.AuditLogConfigurationApi.GetConfig(c.createContext())
			if err != nil {
				return errors.HandleCommon(err, cmd)
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
					// Ignore it, I guess
				}
			}
		}
		return errors.HandleCommon(err, cmd)
	}
	if err = enc.Encode(result); err != nil {
		return errors.HandleCommon(err, cmd)
	}
	return nil
}

func (c *configCommand) edit(cmd *cobra.Command, args []string) error {
	gotSpec, _, err := c.MDSClient.AuditLogConfigurationApi.GetConfig(c.createContext())
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	gotSpecBytes, err := json.MarshalIndent(gotSpec, "", "  ")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	edit := editor.NewEditor()
	edited, path, err := edit.LaunchTempFile("audit-log", bytes.NewBuffer(gotSpecBytes))
	defer os.Remove(path)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	putSpec := mds.AuditLogConfigSpec{}
	if err = json.Unmarshal(edited, &putSpec); err != nil {
		return errors.HandleCommon(err, cmd)
	}
	enc := json.NewEncoder(c.OutOrStdout())
	enc.SetIndent("", "  ")
	result, r, err := c.MDSClient.AuditLogConfigurationApi.PutConfig(c.createContext(), putSpec)
	if err != nil {
		if r.StatusCode == http.StatusConflict {
			if err2 := enc.Encode(result); err2 != nil {
				// Ignore it, I guess
			}
		}
		return errors.HandleCommon(err, cmd)
	}
	if err = enc.Encode(result); err != nil {
		return errors.HandleCommon(err, cmd)
	}
	return nil
}
