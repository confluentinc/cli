package auditlog

import (
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/mds-sdk-go"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"

	"context"
	"encoding/json"
	"fmt"
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
	describeCmd.Flags().String("use", "", "Select an option (\"current\" or \"configured\") to resolve a configuration discrepancy/conflict.")
	c.AddCommand(describeCmd)

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Accepts an audit log configuration spec JSON object and submits it to the API as an update.",
		RunE:  c.update,
		Args:  cobra.NoArgs,
	}
	updateCmd.Flags().String("file", "", "A local file path, read as input. Otherwise the command will read from STDIN.")
	c.AddCommand(updateCmd)
}

func (c *configCommand) describe(cmd *cobra.Command, args []string) error {
	spec, _, err := c.client.AuditLogConfigurationApi.GetConfig(c.ctx)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	result := &spec

	if cmd.Flags().Changed("use") {
		use, err := cmd.Flags().GetString("use")
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		var conflictResolution *mds.AuditLogConfigConflictResolution
		if spec.Metadata.Errors != nil && len(spec.Metadata.Errors) > 0 {
			for _, metadataErr := range spec.Metadata.Errors {
				if match, ok := metadataErr.Conflict[use]; ok {
					conflictResolution = &match
					break
				}
			}
		}
		if conflictResolution == nil {
			err := fmt.Errorf("you asked to --use=\"%s\", but there was no such resolution available", use)
			return errors.HandleCommon(err, cmd)
		}
		result, err = merge(&spec, conflictResolution)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
	}

	err = json.NewEncoder(c.OutOrStdout()).Encode(*result)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	return nil
}

func merge(spec *mds.AuditLogConfigSpec, resolution *mds.AuditLogConfigConflictResolution) (*mds.AuditLogConfigSpec, error) {
	//TODO: implement this method
	return spec, nil
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

	spec := mds.AuditLogConfigSpec{}
	err = json.Unmarshal(data, &spec)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	result, r, err := c.client.AuditLogConfigurationApi.PutConfig(c.ctx, spec)
	if err != nil {
		if r.StatusCode == http.StatusConflict {
			err2 := json.NewEncoder(c.OutOrStdout()).Encode(result)
			if err2 != nil {
				return errors.HandleCommon(err2, cmd)
				//Ignore it, I guess?
			}
		}
		return errors.HandleCommon(err, cmd)
	}

	err = json.NewEncoder(c.OutOrStdout()).Encode(result)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	return nil
}
