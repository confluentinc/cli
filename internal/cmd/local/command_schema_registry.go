package local

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/confluentinc/cli/internal/pkg/local"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
)

var (
	defaultValues = map[string]interface{}{
		"add":       defaultBool,
		"list":      defaultBool,
		"remove":    defaultBool,
		"operation": defaultString,
		"principal": defaultString,
		"subject":   defaultString,
		"topic":     defaultString,
	}
)

func NewSchemaRegistryACLCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	schemaRegistryACLCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "acl",
			Short: "Specify ACL for schema-registry.",
			Args:  cobra.NoArgs,
			RunE:  runSchemaRegistryACLCommand,
		},
		cfg, prerunner)

	schemaRegistryACLCommand.Flags().Bool("add", false, "Indicates you are trying to add ACLs.")
	schemaRegistryACLCommand.Flags().Bool("list", false, "List all the current ACLs")
	schemaRegistryACLCommand.Flags().Bool("remove", false, "Indicates you are trying to remove ACLs.")

	schemaRegistryACLCommand.Flags().StringP("operation", "o", "", "Operation that is being authorized. Valid operation names are: [SUBJECT_READ, SUBJECT_WRITE, SUBJECT_DELETE, SUBJECT_COMPATIBILITY_READ, SUBJECT_COMPATIBILITY_WRITE, GLOBAL_COMPATIBILITY_READ, GLOBAL_COMPATIBILITY_WRITE, GLOBAL_SUBJECTS_READ]")
	schemaRegistryACLCommand.Flags().StringP("principal", "p", "", "Principal to which the ACL is being applied to. Use * to apply to all principals.")
	schemaRegistryACLCommand.Flags().StringP("subject", "s", "", "Subject to which the ACL is being applied to. Only applicable for SUBJECT operations. Use * to apply to all subjects.")
	schemaRegistryACLCommand.Flags().StringP("topic", "t", "", "Topic to which the ACL is being applied to. The corresponding subjects would be topic-key and topic-value. Only applicable for SUBJECT operations. Use * to apply to all subjects.")

	schemaRegistryACLCommand.Flags().SortFlags = false

	return schemaRegistryACLCommand.Command
}

func runSchemaRegistryACLCommand(command *cobra.Command, _ []string) error {
	actions := 0
	for _, flag := range []string{"add", "list", "remove"} {
		isUsed, err := command.Flags().GetBool(flag)
		if err != nil {
			return err
		}
		if isUsed {
			actions++
		}
	}
	if actions != 1 {
		return fmt.Errorf("command must include exactly one action: --add, --list, or --remove")
	}

	ch := local.NewConfluentHomeManager()

	file, err := ch.GetACLCLIFile()
	if err != nil {
		return err
	}

	cc := local.NewConfluentCurrentManager()

	configFile, err := cc.GetConfigFile("schema-registry")
	if err != nil {
		return err
	}

	args, err := collectFlags(command.Flags(), defaultValues)
	if err != nil {
		return err
	}
	args = append(args, "--config", configFile)

	acl := exec.Command(file, args...)
	acl.Stdout = os.Stdout
	acl.Stderr = os.Stderr

	return acl.Run()
}
