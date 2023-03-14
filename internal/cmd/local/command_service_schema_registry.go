package local

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/local"
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

	defaultValues = map[string]any{
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

func NewSchemaRegistryACLCommand(prerunner pcmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "acl",
			Short: "Specify an ACL for Schema Registry.",
			Args:  cobra.NoArgs,
		}, prerunner)

	c.Command.RunE = c.runSchemaRegistryACLCommand

	for _, flag := range []string{"add", "list", "remove", "operation", "principal", "subject", "topic"} {
		switch val := defaultValues[flag].(type) {
		case bool:
			c.Flags().BoolP(flag, shorthands[flag], val, usages[flag])
		case string:
			c.Flags().StringP(flag, shorthands[flag], val, usages[flag])
		}
	}

	return c.Command
}

func (c *Command) runSchemaRegistryACLCommand(cmd *cobra.Command, _ []string) error {
	isUp, err := c.isRunning("kafka")
	if err != nil {
		return err
	}
	if !isUp {
		return c.printStatus("kafka")
	}

	file, err := c.ch.GetFile("bin", "sr-acl-cli")
	if err != nil {
		return err
	}

	configFile, err := c.cc.GetConfigFile("schema-registry")
	if err != nil {
		return err
	}

	args, err := local.CollectFlags(cmd.Flags(), defaultValues)
	if err != nil {
		return err
	}
	args = append(args, "--config", configFile)

	acl := exec.Command(file, args...)
	acl.Stdout = os.Stdout
	acl.Stderr = os.Stderr

	return acl.Run()
}
