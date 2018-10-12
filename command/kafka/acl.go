package kafka

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"strings"
	"reflect"
	"fmt"
)

const ALL = "*"

type AclBindings []AclBinding

// Java reference:
// https://github.com/confluentinc/cc-kafka/blob/trunk/clients/src/main/java/org/apache/kafka/common/acl/AclBinding.java
// C3 reference:
// https://github.com/confluentinc/blueway/blob/master/control-center/src/main/java/io/confluent/controlcenter/rest/jackson/KafkaModule.java#L115-L126
type AclBinding struct {
	Pattern *ResourcePattern     	`json:"pattern"`
	Entry   *AccessControlEntry 	`json:"entry"`
	errors 	[]string				`json:"-"`
}

// Java reference:
// https://github.com/confluentinc/cc-kafka/blob/trunk/clients/src/main/java/org/apache/kafka/common/acl/AccessControlEntry.java
// C3 reference:
// https://github.com/confluentinc/blueway/blob/master/control-center/src/main/java/io/confluent/controlcenter/rest/jackson/KafkaModule.java#L148-L170
type AccessControlEntry struct {
	Principal       string              `json:"principal"`
	//Host            string              `json:"host"`
	Operation       AclOperation        `json:"operation"`
	PermissionType  AclPermissionType   `json:"permissionType"`
}

// Java reference:
// https://github.com/confluentinc/cc-kafka/blob/trunk/clients/src/main/java/org/apache/kafka/common/resource/PatternType.java
// CONFLUENT_* pattern types are not handled by clients, ANY and MATCH are used for filters not setting
type PatternType string

// Java reference
// https://github.com/confluentinc/cc-kafka/blob/trunk/clients/src/main/java/org/apache/kafka/common/resource/ResourceType.java
type ResourceType string

func AclBindingFlags() *pflag.FlagSet {
	//AclBindingFlags.String("host", "*", "Binds ACL to host; Not supported with CCloud")
	flgSet := ResourceFlags()
	flgSet.AddFlagSet(EntryFlags())
	return flgSet
}

// Add Flags for configuring ACL Entry
func EntryFlags() *pflag.FlagSet {
	flgSet := pflag.NewFlagSet("acl-entry", pflag.ExitOnError)
	flgSet.String("user", "", "Bind ACL to user principal")
	flgSet.String("operation", "", "Set ACL operation")
	flgSet.String("producer", "", "Set all ACLs necessary to operate a producer")
	flgSet.String("consumer", "", "Set all ACLs necessary to operate a consumer")
	flgSet.Bool("allow", false, "Set ACL grant access")
	flgSet.Bool("deny", false, "Set ACL to restricts access")
	return flgSet
}

func ResourceFlags() *pflag.FlagSet {
	flgSet := pflag.NewFlagSet("acl-resource", pflag.ExitOnError)

	resourceTypes := []string{
		"topic",
		"consumer-group",
		"cluster",
		"transaction_id",
		// "delegation_token", not supported in ccloud
	}

	for _, rt := range resourceTypes {
		flgSet.String(rt, "", "Bind ACL to resource " + rt)
	}
	return flgSet
}

func AclBindingsFromCMD(cmd *cobra.Command) {
	aclBinding = &AclBinding{}
	cmd.Flags().Visit(visitor(aclBinding))

	return
}

// Java reference:
// https://github.com/confluentinc/cc-kafka/blob/trunk/clients/src/main/java/org/apache/kafka/common/acl/AclOperation.java
type AclOperation string
var AclOperations = []string{
	"READ",
	"WRITE",
	"CREATE",
	"DELETE",
	"ALTER",
	"DESCRIBE",
	"CLUSTER_ACTION",
	"DESCRIBE_CONFIGS",
	"ALTER_CONFIGS",
	"IDEMPOTENT_WRITE",
	"PRODUCER",
	"CONSUMER",
}

// Java reference:
// https://github.com/confluentinc/cc-kafka/blob/trunk/clients/src/main/java/org/apache/kafka/common/acl/AclPermissionType.java
type AclPermissionType string
var AclPermissions = []string{
	"deny",
	"allow",
}

// Java reference:
// https://github.com/confluentinc/cc-kafka/blob/trunk/clients/src/main/java/org/apache/kafka/common/resource/ResourcePattern.java
// C3 reference:
// https://github.com/confluentinc/blueway/blob/master/control-center/src/main/java/io/confluent/controlcenter/rest/jackson/KafkaModule.java#L128-L146
type ResourcePattern struct {
	Resource    ResourceType    `json:"resourceType"`
	Name        string          `json:"name"`
	Type   		PatternType     `json:"patternType"`
}

func visitor(b *AclBinding) func(*pflag.Flag) {
	return func(flag *pflag.Flag) {
		v := flag.Value.String()
		n := strings.ToUpper(flag.Name)
		switch n {
		case "TOPIC":
			fallthrough
		case "CONSUMER-GROUP":
			fallthrough
		case "CLUSTER":
			fallthrough
		case "TRANSACTION_ID":
			if isSet(b.Pattern.Name) {
				b.errors = append(b.errors, "only one resource can be specified per execution")
				break
			}
			b.Pattern.Name = v
			b.Pattern.Resource = ResourceType(n)
			if len(v) > 1 && strings.HasSuffix(v, "*") {
				b.Pattern.Name = v[:len(v)-1]
				b.Pattern.Type = "PREFIXED"
				break
			}
			b.Pattern.Name = v
			b.Pattern.Type = PatternType("LITERAL")
		case "ALLOW":
			fallthrough
		case "DENY":
			b.Entry.PermissionType = AclPermissionType(n)
		case "USER":
			b.Entry.Principal = v
		case "OPERATION":
			if isValid(flag.Value.String(), AclOperations) {
				b.Entry.Operation = AclOperation(v)
				break
			}
			b.errors = append(b.errors, "Invalid operation: " + v)
		}
	}
}

func isSet(v interface{}) bool {
	fmt.Printf("test %s", v)
	return !reflect.DeepEqual(v, reflect.Zero(reflect.TypeOf(v)).Interface())
}

func isValid(actual string, allowable []string) bool {
	for _, val := range allowable {
		if actual == val {
			return true
		}
	}
	return false
}


