package kafka

import (
	"github.com/confluentinc/cli/command/common"
	proto "github.com/confluentinc/cli/shared/kafka"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"strings"
)

// ACLConfiguration wrapper used for flag parsing and validation
type ACLConfiguration struct {
	*proto.KafkaAPIACLRequest
	errors []string
}

// ACLConfigFlags returns a flag set which can be parsed to create an ACLConfiguration object.
func ACLConfigFlags() *pflag.FlagSet {
	flgSet := ACLEntryFlags()
	flgSet.SortFlags = false
	flgSet.AddFlagSet(ResourceFlags())
	return flgSet
}

// ACLEntryFlags returns a flag set which can be parsed to create an AccessControlEntry object.
func ACLEntryFlags() *pflag.FlagSet {
	flgSet := pflag.NewFlagSet("acl-entry", pflag.ExitOnError)
	flgSet.Bool("allow", false, "Set ACL to grant access")
	flgSet.Bool("deny", false, "Set ACL to restrict access")
	flgSet.String("principal", "", "Set Kafka principal")
	//flgSet.String( "host", "*", "Set Kafka principal host. Note: Not supported on CCLOUD.")
	flgSet.String("operation", "", "Set ACL operation")

	return flgSet
}

// ResourceFlags returns a flag set which can be parsed to create a ResourcePattern object.
func ResourceFlags() *pflag.FlagSet {
	flgSet := pflag.NewFlagSet("acl-resource", pflag.ExitOnError)
	flgSet.String("topic", "", "Set TOPIC resource")
	flgSet.String("consumer_group", "", "Set CONSUMER_GROUP resource")
	flgSet.String("transactional_id", "", "Set TRANSACTIONAL_ID resource")
	//flgSet.String("delegation_token", "", "Set DELEGATION_TOKEN resource. Note: Not supported on CCLOUD.")

	return flgSet
}

// parse returns ACLConfiguration from the contents of cmd
func parse(cmd *cobra.Command) *ACLConfiguration {
	aclBinding := &ACLConfiguration{
		KafkaAPIACLRequest: &proto.KafkaAPIACLRequest{
			Entry: &proto.AccessControlEntryConfig{
				Host: "*",
			},
		},
	}
	cmd.Flags().Visit(fromArgs(aclBinding))
	return aclBinding
}

// fromArgs maps command flag values to the appropriate ACLConfiguration field
func fromArgs(conf *ACLConfiguration) func(*pflag.Flag) {
	return func(flag *pflag.Flag) {
		v := flag.Value.String()
		n := strings.ToUpper(flag.Name)
		switch n {
		case "CONSUMER_GROUP":
			// CONSUMER_GROUP is an alias for GROUP
			n = "GROUP"
			fallthrough
		case "TOPIC":
			fallthrough
		case "CLUSTER":
			fallthrough
		case "DELEGATION_TOKEN":
			fallthrough
		case "TRANSACTIONAL_ID":
			if common.IsSet(conf.Pattern) {
				conf.errors = append(conf.errors, "only one resource can be specified per command execution")
				break
			}
			conf.Pattern = &proto.ResourcePatternConfig{}
			conf.Pattern.Name = v
			conf.Pattern.ResourceType = n

			if len(v) > 1 && strings.HasSuffix(v, "*") {
				conf.Pattern.Name = v[:len(v)-1]
				conf.Pattern.PatternType = proto.ResourcePatternConfig_PREFIXED.String()
				break
			}
			conf.Pattern.Name = v
			conf.Pattern.PatternType = proto.ResourcePatternConfig_LITERAL.String()
		case "ALLOW":
			fallthrough
		case "DENY":
			if common.IsSet(conf.Entry.PermissionType) {
				conf.errors = append(conf.errors, "only one resource can be specified per command execution")
				break
			}
			conf.Entry.PermissionType = n
		case "PRINCIPAL":
			conf.Entry.Principal = "user:" + v
		case "OPERATION":
			v = strings.ToUpper(v)
			if _, ok := proto.AccessControlEntryConfig_ACLOperation_value[v]; ok {
				conf.Entry.Operation = v
				break
			}
			conf.errors = append(conf.errors, "Invalid operation value: "+v)
		}
	}
}
