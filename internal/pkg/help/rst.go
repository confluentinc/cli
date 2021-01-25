package help

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/utils"
)

var (
	// Annoying duplicate of bits of https://github.com/confluentinc/docs/blob/master/conf.py#L254
	// and basic ReST linking capabilities, like pointing to URLs
	//
	// Thankfully there's only a couple:
	//   grep -E "\|[^ ]+\|" internal/* -rn
	//   grep :ref: internal/* -rn
	replacements = map[string]string{
		"|ccloud-ent|":                         "Confluent Cloud Enterprise",
		":ref:`only available <cloud-limits>`": "only available",
		":ref:`acl-manage`":                    "https://docs.confluent.io/current/cloud/access-management/acl.html",
		":ref:`kafka_authorization`":           "https://docs.confluent.io/current/kafka/authorization.html",
		".. include:: ../includes/example-ref.rst": `  For a complete example of |ccloud| user account administration, service
  account management, and topic management, see https://docs.confluent.io/current/cloud/access-management/user-service-example.html`,
		".. important::": "",
	}
)

func ResolveReST(template string, cmd *cobra.Command) error {
	//cmd.mergePersistentFlags()
	err := resolveReSTHelper(cmd)
	if err != nil {
		return err
	}
	err = tmpl(cmd.OutOrStderr(), template, cmd)
	if err != nil {
		utils.Println(cmd, err)
	}
	return err
}

func resolveReSTHelper(cmd *cobra.Command) error {
	for rest, text := range replacements {
		cmd.Short = strings.ReplaceAll(cmd.Short, rest, text)
		cmd.Long = strings.ReplaceAll(cmd.Long, rest, text)
		cmd.Example = strings.ReplaceAll(cmd.Example, rest, text)
	}
	if cmd.HasAvailableSubCommands() {
		for _, c := range cmd.Commands() {
			err := resolveReSTHelper(c)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
