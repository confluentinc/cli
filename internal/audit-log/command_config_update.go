package auditlog

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/confluentinc/mds-sdk-go-public/mdsv1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

func (c *configCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Submits audit-log configuration specification object to the API.",
		Long:  "Submits an audit-log configuration specification JSON object to the API.",
		Args:  cobra.NoArgs,
		RunE:  c.update,
	}

	cmd.Flags().String("file", "", "A local file path to the JSON configuration file, read as input. Otherwise the command will read from standard input.")
	cmd.Flags().Bool("force", false, "Updates the configuration, overwriting any concurrent modifications.")
	pcmd.AddMDSOnPremMTLSFlags(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	cobra.CheckErr(cmd.MarkFlagFilename("file", "json"))

	return cmd
}

func (c *configCommand) update(cmd *cobra.Command, _ []string) error {
	client, err := c.GetMDSClient(cmd)
	if err != nil {
		return err
	}

	var data []byte
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

	fileSpec := mdsv1.AuditLogConfigSpec{}
	if err := json.Unmarshal(data, &fileSpec); err != nil {
		return err
	}
	putSpec := &fileSpec

	if cmd.Flags().Changed("force") {
		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}
		if force {
			gotSpec, response, err := client.AuditLogConfigurationApi.GetConfig(c.createContext())
			if err != nil {
				return HandleMdsAuditLogApiError(err, response)
			}
			putSpec = &mdsv1.AuditLogConfigSpec{
				Destinations:       fileSpec.Destinations,
				ExcludedPrincipals: fileSpec.ExcludedPrincipals,
				DefaultTopics:      fileSpec.DefaultTopics,
				Routes:             fileSpec.Routes,
				Metadata: &mdsv1.AuditLogConfigMetadata{
					ResourceVersion: gotSpec.Metadata.ResourceVersion,
				},
			}
		}
	}

	enc := json.NewEncoder(c.OutOrStdout())
	enc.SetIndent("", "  ")
	result, httpResp, err := client.AuditLogConfigurationApi.PutConfig(c.createContext(), *putSpec)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusConflict {
			if apiError, ok := err.(mdsv1.GenericOpenAPIError); ok {
				_ = enc.Encode(apiError.Model())
				// We can just ignore this extra error. Why?
				// We expected a payload we could display as JSON, but got something unexpected.
				// That's OK though, we'll still handle and show the API error message.
			}
		}
		return HandleMdsAuditLogApiError(err, httpResp)
	}

	return enc.Encode(result)
}
