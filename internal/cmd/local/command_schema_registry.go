package local

import (
	"os"
	"os/exec"

	"github.com/confluentinc/cli/internal/pkg/local"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
)

var (
	usages = map[string]string{
		"add":    "Indicates you are trying to add ACLs.",
		"list":   "List all the current ACLs",
		"remove": "Indicates you are trying to remove ACLs.",

		"operation": "Operation that is being authorized. Valid operation names are: [SUBJECT_READ, SUBJECT_WRITE, SUBJECT_DELETE, SUBJECT_COMPATIBILITY_READ, SUBJECT_COMPATIBILITY_WRITE, GLOBAL_COMPATIBILITY_READ, GLOBAL_COMPATIBILITY_WRITE, GLOBAL_SUBJECTS_READ]",
		"principal": "Principal to which the ACL is being applied to. Use * to apply to all principals.",
		"subject":   "Subject to which the ACL is being applied to. Only applicable for SUBJECT operations. Use * to apply to all subjects.",
		"topic":     "Topic to which the ACL is being applied to. The corresponding subjects would be topic-key and topic-value. Only applicable for SUBJECT operations. Use * to apply to all subjects.",
	}

	defaultValues = map[string]interface{}{
		"add":    defaultBool,
		"list":   defaultBool,
		"remove": defaultBool,

		"operation": defaultString,
		"principal": defaultString,
		"subject":   defaultString,
		"topic":     defaultString,
	}

	shorthands = map[string]string{
		"add":    "",
		"list":   "",
		"remove": "",

		"operation": "o",
		"principal": "p",
		"subject":   "s",
		"topic":     "t",
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

	for flag, val := range defaultValues {
		switch val.(type) {
		case bool:
			schemaRegistryACLCommand.Flags().BoolP(flag, shorthands[flag], val.(bool), usages[flag])
		case string:
			schemaRegistryACLCommand.Flags().StringP(flag, shorthands[flag], val.(string), usages[flag])
		}
	}
	schemaRegistryACLCommand.Flags().SortFlags = false

	return schemaRegistryACLCommand.Command
}

func runSchemaRegistryACLCommand(command *cobra.Command, _ []string) error {
	ch := local.NewConfluentHomeManager()

	file, err := ch.GetFile("bin", "sr-acl-cli")
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
