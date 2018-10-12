package kafka

import (
	"fmt"

	"github.com/spf13/cobra"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/shared"
)

type aclCommand struct {
	*cobra.Command
	config *shared.Config
	kafka  Kafka
}

var cluster *schedv1.KafkaCluster
var aclBinding *AclBinding

type stringValue string

func (v *stringValue) Set(s string) error {
	if *v != "" {
		return fmt.Errorf("duplicate values are not permitted")
	}
	*v = stringValue(s)
	return nil
}

// NewTopicCommand returns the Cobra clusterCommand for Kafka Cluster.
func NewAclCommand(config *shared.Config, kafka Kafka) *cobra.Command {
	cmd := &aclCommand{
		Command: &cobra.Command{
			Use:   "acl",
			Short: "Manage Kafka ACLs.",
		},
		config: config,
		kafka:  kafka,
	}

	cmd.init()
	return cmd.Command
}

func (c *aclCommand) init() {
	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Create a Kafka ACL.",
		RunE:  c.add,
		Args:  cobra.NoArgs,
	}
	addCmd.Flags().AddFlagSet(AclBindingFlags())
	c.AddCommand(addCmd)

	delCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a Kafka ACL.",
		RunE:  c.delete,
		Args:  cobra.NoArgs,
	}
	delCmd.Flags().AddFlagSet(AclBindingFlags())
	c.AddCommand(delCmd)

	lstCmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka ACLs for resource.",
		RunE:  c.list,
		Args:  cobra.NoArgs,
	}
	lstCmd.Flags().AddFlagSet(ResourceFlags())
	c.AddCommand(lstCmd)
}

func (c *aclCommand) list(cmd *cobra.Command, args []string) error {
	return nil
}

func (c *aclCommand) add(cmd *cobra.Command, args []string) error {
	AclBindingsFromCMD(cmd)
	return nil
}

func (c *aclCommand) delete(cmd *cobra.Command, args []string) error {
	AclBindingsFromCMD(cmd)
	return nil
}
