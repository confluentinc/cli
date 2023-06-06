package auditlog

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	mds "github.com/confluentinc/mds-sdk-go-public/mdsv1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func (c *configCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Submits audit-log config spec object to the API.",
		Long:  "Submits an audit-log configuration specification JSON object to the API.",
		Args:  cobra.NoArgs,
		RunE:  c.update,
	}

	cmd.Flags().String("file", "", "A local file path to the JSON configuration file, read as input. Otherwise the command will read from standard input.")
	cmd.Flags().Bool("force", false, "Updates the configuration, overwriting any concurrent modifications.")
	pcmd.AddContextFlag(cmd, c.CLICommand)

	cobra.CheckErr(cmd.MarkFlagFilename("file", "json"))

	return cmd
}

func (c *configCommand) update(cmd *cobra.Command, _ []string) error {
	var data []byte
	var err error
	if cmd.Flags().Changed("file") {
		file, err := cmd.Flags().GetString("file")
		if err != nil {
			return err
		}
		data, err = os.ReadFile(file)
		if err != nil {
			return err
		}
	} else {
		data, err = io.ReadAll(os.Stdin)
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
	result, httpResp, err := c.MDSClient.AuditLogConfigurationApi.PutConfig(c.createContext(), *putSpec)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusConflict {
			if apiError, ok := err.(mds.GenericOpenAPIError); ok {
				_ = enc.Encode(apiError.Model())
				// We can just ignore this extra error. Why?
				// We expected a payload we could display as JSON, but got something unexpected.
				// That's OK though, we'll still handle and show the API error message.
			}
		}
		return HandleMdsAuditLogApiError(cmd, err, httpResp)
	}

	return enc.Encode(result)
}
