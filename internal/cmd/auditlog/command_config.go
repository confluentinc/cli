package auditlog

import (
	"bytes"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/go-editor"
	"github.com/confluentinc/mds-sdk-go"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"

	"context"
	"encoding/json"
	"net/http"
)

type configCommand struct {
	*cobra.Command
	config *config.Config
	client *mds.APIClient
	ctx    context.Context
}

// NewRouteCommand returns the sub-command object for interacting with audit log route rules.
func NewConfigCommand(config *config.Config, client *mds.APIClient) *cobra.Command {
	cmd := &configCommand{
		Command: &cobra.Command{
			Use:   "config",
			Short: "Manage audit log configuration specification.",
			Long:  "Manage audit log defaults and routing rules that determine which auditable events are logged, and where.",
		},
		config: config,
		client: client,
		ctx:    context.WithValue(context.Background(), mds.ContextAccessToken, config.AuthToken),
	}

	cmd.init()
	return cmd.Command
}

func (c *configCommand) init() {
	describeCmd := &cobra.Command{
		Use:   "describe",
		Short: "Returns the entire audit log configuration spec object.",
		RunE:  c.describe,
		Args:  cobra.NoArgs,
	}
	c.AddCommand(describeCmd)

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Submits audit-log config spec object to the API.",
		Long:  "Submits an audit-log configuration specification JSON object to the API.",
		RunE:  c.update,
		Args:  cobra.NoArgs,
	}
	updateCmd.Flags().String("file", "", "A local file path, read as input. Otherwise the command will read from standard in.")
	updateCmd.Flags().Bool("force", false, "Tries to update even with concurrent modifications.")
	updateCmd.Flags().SortFlags = false
	c.AddCommand(updateCmd)

	editCmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit the audit-log config spec object interactively.",
		RunE:  c.edit,
		Args:  cobra.NoArgs,
	}
	c.AddCommand(editCmd)
}

func (c *configCommand) describe(cmd *cobra.Command, args []string) error {
	spec, _, err := c.client.AuditLogConfigurationApi.GetConfig(c.ctx)
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
			gotSpec, _, err := c.client.AuditLogConfigurationApi.GetConfig(c.ctx)
			if err != nil {
				return errors.HandleCommon(err, cmd)
			}
			putSpec = &mds.AuditLogConfigSpec{
				Destinations:       fileSpec.Destinations,
				ExcludedPrincipals: fileSpec.ExcludedPrincipals,
				DefaultTopics:      fileSpec.DefaultTopics,
				Routes:             fileSpec.Routes,
				Metadata: mds.AuditLogConfigMetadata{
					ResourceVersion: gotSpec.Metadata.ResourceVersion,
				},
			}
		}
	}

	enc := json.NewEncoder(c.OutOrStdout())
	enc.SetIndent("", "  ")
	result, r, err := c.client.AuditLogConfigurationApi.PutConfig(c.ctx, *putSpec)
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
	gotSpec, _, err := c.client.AuditLogConfigurationApi.GetConfig(c.ctx)
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
	result, r, err := c.client.AuditLogConfigurationApi.PutConfig(c.ctx, putSpec)
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
