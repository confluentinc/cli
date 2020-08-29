package local

import (
	"os"
	"os/exec"

	"github.com/confluentinc/cli/internal/pkg/local"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
)

var (
	usages = map[string]string{
		"add":    "Indicates you are trying to add ACLs.",
		"list":   "List all the current ACLs.",
		"remove": "Indicates you are trying to remove ACLs.",

		"operation": "Operation that is being authorized. Valid operation names are SUBJECT_READ, SUBJECT_WRITE, SUBJECT_DELETE, SUBJECT_COMPATIBILITY_READ, SUBJECT_COMPATIBILITY_WRITE, GLOBAL_COMPATIBILITY_READ, GLOBAL_COMPATIBILITY_WRITE, and GLOBAL_SUBJECTS_READ.",
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

func NewSchemaRegistryACLCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "acl",
			Short: "Specify an ACL for Schema Registry.",
			Args:  cobra.NoArgs,
		}, prerunner)

	c.Command.RunE = cmd.NewCLIRunE(c.runSchemaRegistryACLCommand)
	for flag, val := range defaultValues {
		switch val.(type) {
		case bool:
			c.Flags().BoolP(flag, shorthands[flag], val.(bool), usages[flag])
		case string:
			c.Flags().StringP(flag, shorthands[flag], val.(string), usages[flag])
		}
	}
	c.Flags().SortFlags = false

	return c.Command
}

func (c *Command) runSchemaRegistryACLCommand(command *cobra.Command, _ []string) error {
	isUp, err := c.isRunning("kafka")
	if err != nil {
		return err
	}
	if !isUp {
		return c.printStatus(command, "kafka")
	}

	file, err := c.ch.GetFile("bin", "sr-acl-cli")
	if err != nil {
		return err
	}

	configFile, err := c.cc.GetConfigFile("schema-registry")
	if err != nil {
		return err
	}

	args, err := local.CollectFlags(command.Flags(), defaultValues)
	if err != nil {
		return err
	}
	args = append(args, "--config", configFile)

	acl := exec.Command(file, args...)
	acl.Stdout = os.Stdout
	acl.Stderr = os.Stderr

	return acl.Run()
}
