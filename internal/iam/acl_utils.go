package iam

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/confluentinc/mds-sdk-go-public/mdsv1"

	"github.com/confluentinc/cli/v3/pkg/acl"
	"github.com/confluentinc/cli/v3/pkg/utils"
)

// ACLConfiguration wrapper used for flag parsing and validation
type ACLConfiguration struct {
	*mdsv1.CreateAclRequest
	errors error
}

func aclFlags() *pflag.FlagSet {
	flgSet := pflag.NewFlagSet("acl-config", pflag.ExitOnError)
	flgSet.String("kafka-cluster", "", "Kafka cluster ID for scope of ACL commands.")
	flgSet.String("principal", "", "Principal for this operation with User: or Group: prefix.")
	flgSet.String("operation", "", fmt.Sprintf("Set ACL Operation to: (%s).", acl.ConvertToLower(acl.AclOperations)))
	flgSet.String("host", "*", "Set host for access. Only IP addresses are supported.")
	flgSet.Bool("allow", false, "ACL permission to allow access.")
	flgSet.Bool("deny", false, "ACL permission to restrict access to resource.")
	flgSet.Bool("cluster-scope", false, `Set the cluster resource. With this option the ACL grants
access to the provided operations on the Kafka cluster itself.`)
	flgSet.String("consumer-group", "", "Set the Consumer Group resource.")
	flgSet.String("transactional-id", "", "Set the TransactionalID resource.")
	flgSet.String("topic", "", `Set the topic resource. With this option the ACL grants the provided
operations on the topics that start with that prefix, depending on whether
the --prefix option was also passed.`)
	flgSet.Bool("prefix", false, "Set to match all resource names prefixed with this value.")
	flgSet.SortFlags = false
	return flgSet
}

// parse returns ACLConfiguration from the contents of cmd
func parse(cmd *cobra.Command) *ACLConfiguration {
	aclConfiguration := &ACLConfiguration{
		CreateAclRequest: &mdsv1.CreateAclRequest{
			Scope: mdsv1.KafkaScope{
				Clusters: mdsv1.KafkaScopeClusters{},
			},
			AclBinding: mdsv1.AclBinding{
				Entry: mdsv1.AccessControlEntry{
					Host: "*",
				},
				Pattern: mdsv1.KafkaResourcePattern{},
			},
		},
	}
	cmd.Flags().Visit(fromArgs(aclConfiguration))
	return aclConfiguration
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
		case "kafka-cluster":
			conf.Scope.Clusters.KafkaCluster = v
		case "topic":
			fallthrough
		case "delegation-token":
			fallthrough
		case "transactional-id":
			setResourcePattern(conf, n, v)
		case "allow":
			conf.AclBinding.Entry.PermissionType = mdsv1.ACLPERMISSIONTYPE_ALLOW
		case "deny":
			conf.AclBinding.Entry.PermissionType = mdsv1.ACLPERMISSIONTYPE_DENY
		case "prefix":
			conf.AclBinding.Pattern.PatternType = mdsv1.PATTERNTYPE_PREFIXED
		case "principal":
			conf.AclBinding.Entry.Principal = v
		case "host":
			conf.AclBinding.Entry.Host = v
		case "operation":
			v = strings.ToUpper(v)
			v = strings.ReplaceAll(v, "-", "_")
			enumUtils := utils.EnumUtils{}
			enumUtils.Init(
				mdsv1.ACLOPERATION_UNKNOWN,
				mdsv1.ACLOPERATION_ANY,
				mdsv1.ACLOPERATION_ALL,
				mdsv1.ACLOPERATION_READ,
				mdsv1.ACLOPERATION_WRITE,
				mdsv1.ACLOPERATION_CREATE,
				mdsv1.ACLOPERATION_DELETE,
				mdsv1.ACLOPERATION_ALTER,
				mdsv1.ACLOPERATION_DESCRIBE,
				mdsv1.ACLOPERATION_CLUSTER_ACTION,
				mdsv1.ACLOPERATION_DESCRIBE_CONFIGS,
				mdsv1.ACLOPERATION_ALTER_CONFIGS,
				mdsv1.ACLOPERATION_IDEMPOTENT_WRITE,
			)
			if op, ok := enumUtils[v]; ok {
				conf.AclBinding.Entry.Operation = op.(mdsv1.AclOperation)
				break
			}
			conf.errors = multierror.Append(conf.errors, fmt.Errorf("invalid operation value: %s", v))
		}
	}
}

func setResourcePattern(conf *ACLConfiguration, n, v string) {
	if conf.AclBinding.Pattern.ResourceType != "" {
		// A resourceType has already been set with a previous flag
		conf.errors = multierror.Append(conf.errors, fmt.Errorf("exactly one of %v must be set",
			convertToFlags(mdsv1.ACLRESOURCETYPE_TOPIC, mdsv1.ACLRESOURCETYPE_GROUP,
				mdsv1.ACLRESOURCETYPE_CLUSTER, mdsv1.ACLRESOURCETYPE_TRANSACTIONAL_ID)))
		return
	}

	// Normalize the resource pattern name
	n = strings.ToUpper(n)
	n = strings.ReplaceAll(n, "-", "_")

	enumUtils := utils.EnumUtils{}
	enumUtils.Init(mdsv1.ACLRESOURCETYPE_TOPIC, mdsv1.ACLRESOURCETYPE_GROUP,
		mdsv1.ACLRESOURCETYPE_CLUSTER, mdsv1.ACLRESOURCETYPE_TRANSACTIONAL_ID)
	conf.AclBinding.Pattern.ResourceType = enumUtils[n].(mdsv1.AclResourceType)

	if conf.AclBinding.Pattern.ResourceType == mdsv1.ACLRESOURCETYPE_CLUSTER {
		conf.AclBinding.Pattern.PatternType = mdsv1.PATTERNTYPE_LITERAL
	}
	conf.AclBinding.Pattern.Name = v
}

func convertToFlags(operations ...any) string {
	ops := make([]string, len(operations))

	for i, v := range operations {
		if v == mdsv1.ACLRESOURCETYPE_GROUP {
			v = "consumer-group"
		}
		if v == mdsv1.ACLRESOURCETYPE_CLUSTER {
			v = "cluster-scope"
		}
		s := strings.ToLower(strings.ReplaceAll(fmt.Sprint(v), "_", "-"))
		ops[i] = fmt.Sprintf("`--%s`", s)
	}

	sort.Strings(ops)
	return strings.Join(ops, ", ")
}
