package auditlog

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/confluentinc/go-editor"
	mds "github.com/confluentinc/mds-sdk-go-public/mdsv1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func (c *configCommand) newEditCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit the audit-log configuration spec interactively.",
		Long:  "Edit the audit-log configuration spec object interactively, using the $EDITOR specified in your environment (for example, vim).",
		Args:  cobra.NoArgs,
		RunE:  c.edit,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
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
	result, httpResp, err := c.MDSClient.AuditLogConfigurationApi.PutConfig(c.createContext(), putSpec)
	if err != nil {
		if httpResp.StatusCode == http.StatusConflict {
			_ = enc.Encode(result)
			// We can just ignore this extra error. Why?
			// We expected a payload we could display as JSON, but got something unexpected.
			// That's OK though, we'll still handle and show the API error message.
		}
		return HandleMdsAuditLogApiError(cmd, err, httpResp)
	}

	return enc.Encode(result)
}
