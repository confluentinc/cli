package acl

import (
	"io"

	"github.com/spf13/cobra"

	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func PrintAcls(cmd *cobra.Command, bindingsObj []*kafkav1.ACLBinding, writer io.Writer) error {

	// non list commands which do not have -o flags also uses this function, need to set default
	_, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	}

	aclListFields := []string{"ServiceAccountId", "Permission", "Operation", "Resource", "Name", "Type"}
	aclListStructuredRenames := []string{"service_account_id", "permission", "operation", "resource", "name", "type"}
	outputWriter, err := output.NewListOutputCustomizableWriter(cmd, aclListFields, aclListFields, aclListStructuredRenames, writer)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	for _, binding := range bindingsObj {
		record := &struct {
			ServiceAccountId string
			Permission       string
			Operation        string
			Resource         string
			Name             string
			Type             string
		}{
			binding.Entry.Principal,
			binding.Entry.PermissionType.String(),
			binding.Entry.Operation.String(),
			binding.Pattern.ResourceType.String(),
			binding.Pattern.Name,
			binding.Pattern.PatternType.String(),
		}
		outputWriter.AddElement(record)
	}
	return outputWriter.Out()
}
