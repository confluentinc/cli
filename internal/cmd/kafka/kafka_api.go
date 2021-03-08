package kafka

import (
	"fmt"
	"sort"
	"strings"

	"github.com/confluentinc/cli/internal/pkg/errors"

	"github.com/hashicorp/go-multierror"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
)

// ACLConfiguration wrapper used for flag parsing and validation
type ACLConfiguration struct {
	*schedv1.ACLBinding
	errors error
}

func NewACLConfig() *ACLConfiguration {
	return &ACLConfiguration{
		ACLBinding: &schedv1.ACLBinding{
			Entry: &schedv1.AccessControlEntryConfig{
				Host: "*",
			},
			Pattern: &schedv1.ResourcePatternConfig{},
		},
	}
}

// aclConfigFlags returns a flag set which can be parsed to create an ACLConfiguration object.
func aclConfigFlags() *pflag.FlagSet {
	flgSet := aclEntryFlags()
	flgSet.SortFlags = false
	flgSet.AddFlagSet(resourceFlags())
	return flgSet
}

// aclEntryFlags returns a flag set which can be parsed to create an AccessControlEntry object.
func aclEntryFlags() *pflag.FlagSet {
	flgSet := pflag.NewFlagSet("acl-entry", pflag.ExitOnError)
	//flgSet.String("cluster", "", "Confluent Cloud cluster ID.")
	flgSet.Bool("allow", false, "Access to the resource is allowed.") // permission
	flgSet.Bool("deny", false, "Access to the resource is denied.")
	//flgSet.String( "host", "*", "Set Kafka principal host. Note: Not supported on CCLOUD.")
	flgSet.Int("service-account", 0, "The service account ID.")
	operationHelpOneLine := fmt.Sprintf("The ACL Operation: (%s).\nNote: This flag may be specified more than once.",
		listEnum(schedv1.ACLOperations_ACLOperation_name, []string{"ANY", "UNKNOWN"}))
	operationHelpParts := strings.SplitAfter(operationHelpOneLine, "delete, ")
	operationHelp := operationHelpParts[0] + "\n" + operationHelpParts[1]
	flgSet.StringArray("operation", []string{""}, operationHelp)
	// An error is only returned if the flag name is not present.
	// We know the flag name is present so its safe to ignore this.
	_ = cobra.MarkFlagRequired(flgSet, "service-account")
	_ = cobra.MarkFlagRequired(flgSet, "operation")
	return flgSet
}

// resourceFlags returns a flag set which can be parsed to create a ResourcePattern object.
func resourceFlags() *pflag.FlagSet {
	flgSet := pflag.NewFlagSet("acl-resource", pflag.ExitOnError)
	//flgSet.String("cluster", "", "The Confluent Cloud cluster ID.")
	flgSet.Bool("cluster-scope", false, `Modify ACLs for the cluster.`)
	flgSet.String("topic", "", `Modify ACLs for the specified topic resource.`)
	flgSet.String("consumer-group", "", "Modify ACLs for the specified consumer group resource.")
	flgSet.String("transactional-id", "", "Modify ACLs for the specified TransactionalID resource.")
	flgSet.Bool("prefix", false, `When this flag is set, the specified resource name is interpreted as
a prefix.`)

	return flgSet
}

// parse returns ACLConfiguration from the contents of cmd
func parse(cmd *cobra.Command) ([]*ACLConfiguration, error) {
	var aclConfigs []*ACLConfiguration

	if cmd.Name() == listCmd.Name() {
		aclConfig := NewACLConfig()
		cmd.Flags().Visit(fromArgs(aclConfig))
		aclConfigs = append(aclConfigs, aclConfig)
		return aclConfigs, nil
	}

	operations, err := cmd.Flags().GetStringArray("operation")
	if err != nil {
		return nil, err
	}
	for _, operation := range operations {
		aclConfig := NewACLConfig()
		op, err := getACLOperation(operation)
		if err != nil {
			return nil, err
		}
		aclConfig.Entry.Operation = op
		cmd.Flags().Visit(fromArgs(aclConfig))
		aclConfigs = append(aclConfigs, aclConfig)
	}
	return aclConfigs, nil
}

// fromArgs maps command flag values to the appropriate ACLConfiguration field
func fromArgs(conf *ACLConfiguration) func(*pflag.Flag) {
	return func(flag *pflag.Flag) {
		v := flag.Value.String()
		switch n := flag.Name; n {
		case "consumer-group":
			setResourcePattern(conf, "GROUP", v)
		case "cluster-scope":
			// The only valid name for a cluster is kafka-cluster
			// https://github.com/confluentinc/cc-kafka/blob/88823c6016ea2e306340938994d9e122abf3c6c0/core/src/main/scala/kafka/security/auth/Resource.scala#L24
			setResourcePattern(conf, "cluster", "kafka-cluster")
		case "topic":
			fallthrough
		case "delegation-token":
			fallthrough
		case "transactional-id":
			setResourcePattern(conf, n, v)
		case "allow":
			conf.Entry.PermissionType = schedv1.ACLPermissionTypes_ALLOW
		case "deny":
			conf.Entry.PermissionType = schedv1.ACLPermissionTypes_DENY
		case "prefix":
			conf.Pattern.PatternType = schedv1.PatternTypes_PREFIXED
		case "service-account":
			if v == "0" {
				conf.Entry.Principal = "User:*"
				break
			}
			conf.Entry.Principal = "User:" + v
		}
	}
}

func setResourcePattern(conf *ACLConfiguration, n, v string) {
	/* Normalize the resource pattern name */
	if conf.Pattern.ResourceType != schedv1.ResourceTypes_UNKNOWN {
		conf.errors = multierror.Append(conf.errors, fmt.Errorf(errors.ExactlyOneSetErrorMsg,
			listEnum(schedv1.ResourceTypes_ResourceType_name, []string{"ANY", "UNKNOWN"})))
		return
	}

	n = strings.ToUpper(n)
	n = strings.ReplaceAll(n, "-", "_")

	conf.Pattern.ResourceType = schedv1.ResourceTypes_ResourceType(schedv1.ResourceTypes_ResourceType_value[n])

	if conf.Pattern.ResourceType == schedv1.ResourceTypes_CLUSTER {
		conf.Pattern.PatternType = schedv1.PatternTypes_LITERAL
	}
	conf.Pattern.Name = v
}

func listEnum(enum map[int32]string, exclude []string) string {
	var ops []string

OUTER:
	for _, v := range enum {
		for _, exclusion := range exclude {
			if v == exclusion {
				continue OUTER
			}
		}
		if v == "GROUP" {
			v = "consumer-group"
		}
		if v == "CLUSTER" {
			v = "cluster-scope"
		}
		v = strings.ReplaceAll(v, "_", "-")
		ops = append(ops, strings.ToLower(v))
	}

	sort.Strings(ops)
	return strings.Join(ops, ", ")
}

func getACLOperation(operation string) (schedv1.ACLOperations_ACLOperation, error) {
	op := strings.ToUpper(operation)
	op = strings.ReplaceAll(op, "-", "_")
	if operation, ok := schedv1.ACLOperations_ACLOperation_value[op]; ok {
		return schedv1.ACLOperations_ACLOperation(operation), nil
	}
	return schedv1.ACLOperations_UNKNOWN, fmt.Errorf(errors.InvalidOperationValueErrorMsg, op)
}
