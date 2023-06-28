package flink

import (
	"github.com/spf13/cobra"
)

type iamBindingOut struct {
	Id           string `human:"ID" serialized:"id"`
	Cloud        string `human:"Cloud" serialized:"cloud"`
	Region       string `human:"Region" serialized:"region"`
	Environment  string `human:"Environment" serialized:"environment"`
	IdentityPool string `human:"Identity Pool" serialized:"identity_pool"`
}

func (c *command) newIamBindingCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "iam-binding",
		Short: "Manage Flink IAM bindings.",
	}

	cmd.AddCommand(c.newIamBindingCreateCommand())
	cmd.AddCommand(c.newIamBindingDeleteCommand())
	cmd.AddCommand(c.newIamBindingListCommand())

	return cmd
}
