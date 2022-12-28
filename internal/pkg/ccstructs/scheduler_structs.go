package ccstructs

import (
	proto "github.com/gogo/protobuf/proto"
)

// ACLBindng: binds an ACL to to a resource pattern.
// Apache Kafka reference:
// https://github.com/apache/kafka/blob/trunk/clients/src/main/java/org/apache/kafka/common/acl/AclBinding.java
type ACLBinding struct {
	Pattern              *ResourcePatternConfig    `protobuf:"bytes,1,opt,name=pattern,proto3" json:"pattern,omitempty" db:"pattern,omitempty" url:"pattern,omitempty"`
	Entry                *AccessControlEntryConfig `protobuf:"bytes,2,opt,name=entry,proto3" json:"entry,omitempty" db:"entry,omitempty" url:"entry,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                  `json:"-"`
	XXX_unrecognized     []byte                    `json:"-"`
	XXX_sizecache        int32                     `json:"-"`
}

func (m *ACLBinding) GetEntry() *AccessControlEntryConfig {
	if m != nil {
		return m.Entry
	}
	return nil
}

func (m *ACLBinding) GetPattern() *ResourcePatternConfig {
	if m != nil {
		return m.Pattern
	}
	return nil
}

// ResourcePatternConfig: matches ACLs with resources.
// Apache Kafka reference:
// https://github.com/apache/kafka/blob/trunk/clients/src/main/java/org/apache/kafka/common/resource/ResourcePattern.java
type ResourcePatternConfig struct {
	ResourceType         ResourceTypes_ResourceType `protobuf:"varint,1,opt,name=resource_type,json=resourceType,proto3,enum=ResourceTypes_ResourceType" json:"resourceType" db:"resource_type,omitempty" url:"resource_type,omitempty"`
	Name                 string                     `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty" db:"name,omitempty" url:"name,omitempty"`
	PatternType          PatternTypes_PatternType   `protobuf:"varint,3,opt,name=pattern_type,json=patternType,proto3,enum=PatternTypes_PatternType" json:"patternType" db:"pattern_type,omitempty" url:"pattern_type,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                   `json:"-"`
	XXX_unrecognized     []byte                     `json:"-"`
	XXX_sizecache        int32                      `json:"-"`
}

func (m *ResourcePatternConfig) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *ResourcePatternConfig) GetResourceType() ResourceTypes_ResourceType {
	if m != nil {
		return m.ResourceType
	}
	return ResourceTypes_UNKNOWN
}

func (m *ResourcePatternConfig) GetPatternType() PatternTypes_PatternType {
	if m != nil {
		return m.PatternType
	}
	return PatternTypes_UNKNOWN
}

type ResourceTypes_ResourceType int32

const (
	ResourceTypes_UNKNOWN          ResourceTypes_ResourceType = 0
	ResourceTypes_ANY              ResourceTypes_ResourceType = 1
	ResourceTypes_TOPIC            ResourceTypes_ResourceType = 2
	ResourceTypes_GROUP            ResourceTypes_ResourceType = 3
	ResourceTypes_CLUSTER          ResourceTypes_ResourceType = 4
	ResourceTypes_TRANSACTIONAL_ID ResourceTypes_ResourceType = 5
)

func (x ResourceTypes_ResourceType) String() string {
	return proto.EnumName(ResourceTypes_ResourceType_name, int32(x))
}

var ResourceTypes_ResourceType_name = map[int32]string{
	0: "UNKNOWN",
	1: "ANY",
	2: "TOPIC",
	3: "GROUP",
	4: "CLUSTER",
	5: "TRANSACTIONAL_ID",
}

var ResourceTypes_ResourceType_value = map[string]int32{
	"UNKNOWN":          0,
	"ANY":              1,
	"TOPIC":            2,
	"GROUP":            3,
	"CLUSTER":          4,
	"TRANSACTIONAL_ID": 5,
}

type PatternTypes_PatternType int32

const (
	PatternTypes_UNKNOWN  PatternTypes_PatternType = 0
	PatternTypes_ANY      PatternTypes_PatternType = 1
	PatternTypes_LITERAL  PatternTypes_PatternType = 2
	PatternTypes_PREFIXED PatternTypes_PatternType = 3
)

func (x PatternTypes_PatternType) String() string {
	return proto.EnumName(PatternTypes_PatternType_name, int32(x))
}

var PatternTypes_PatternType_name = map[int32]string{
	0: "UNKNOWN",
	1: "ANY",
	2: "LITERAL",
	3: "PREFIXED",
}

// AccessControlEntryConfig(ACE): a tuple of principal, host, operation, and permissionType.
// Apache Kafka reference:
// https://github.com/apache/kafka/blob/trunk/clients/src/main/java/org/apache/kafka/common/acl/AccessControlEntry.java
type AccessControlEntryConfig struct {
	Principal            string                               `protobuf:"bytes,1,opt,name=principal,proto3" json:"principal,omitempty" db:"principal,omitempty" url:"principal,omitempty"`
	Operation            ACLOperations_ACLOperation           `protobuf:"varint,2,opt,name=operation,proto3,enum=ACLOperations_ACLOperation" json:"operation,omitempty" db:"operation,omitempty" url:"operation,omitempty"`
	Host                 string                               `protobuf:"bytes,3,opt,name=host,proto3" json:"host,omitempty" db:"host,omitempty" url:"host,omitempty"`
	PermissionType       ACLPermissionTypes_ACLPermissionType `protobuf:"varint,4,opt,name=permission_type,json=permissionType,proto3,enum=ACLPermissionTypes_ACLPermissionType" json:"permissionType" db:"permission_type,omitempty" url:"permission_type,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                             `json:"-"`
	XXX_unrecognized     []byte                               `json:"-"`
	XXX_sizecache        int32                                `json:"-"`
}

func (m *AccessControlEntryConfig) GetHost() string {
	if m != nil {
		return m.Host
	}
	return ""
}

func (m *AccessControlEntryConfig) GetPrincipal() string {
	if m != nil {
		return m.Principal
	}
	return ""
}

func (m *AccessControlEntryConfig) GetOperation() ACLOperations_ACLOperation {
	if m != nil {
		return m.Operation
	}
	return ACLOperations_UNKNOWN
}

func (m *AccessControlEntryConfig) GetPermissionType() ACLPermissionTypes_ACLPermissionType {
	if m != nil {
		return m.PermissionType
	}
	return ACLPermissionTypes_UNKNOWN
}

type ACLOperations_ACLOperation int32

const (
	ACLOperations_UNKNOWN          ACLOperations_ACLOperation = 0
	ACLOperations_ANY              ACLOperations_ACLOperation = 1
	ACLOperations_READ             ACLOperations_ACLOperation = 2
	ACLOperations_WRITE            ACLOperations_ACLOperation = 3
	ACLOperations_CREATE           ACLOperations_ACLOperation = 4
	ACLOperations_DELETE           ACLOperations_ACLOperation = 5
	ACLOperations_ALTER            ACLOperations_ACLOperation = 6
	ACLOperations_DESCRIBE         ACLOperations_ACLOperation = 7
	ACLOperations_CLUSTER_ACTION   ACLOperations_ACLOperation = 8
	ACLOperations_DESCRIBE_CONFIGS ACLOperations_ACLOperation = 9
	ACLOperations_ALTER_CONFIGS    ACLOperations_ACLOperation = 10
	ACLOperations_IDEMPOTENT_WRITE ACLOperations_ACLOperation = 11
)

func (x ACLOperations_ACLOperation) String() string {
	return proto.EnumName(ACLOperations_ACLOperation_name, int32(x))
}

var ACLOperations_ACLOperation_name = map[int32]string{
	0:  "UNKNOWN",
	1:  "ANY",
	2:  "READ",
	3:  "WRITE",
	4:  "CREATE",
	5:  "DELETE",
	6:  "ALTER",
	7:  "DESCRIBE",
	8:  "CLUSTER_ACTION",
	9:  "DESCRIBE_CONFIGS",
	10: "ALTER_CONFIGS",
	11: "IDEMPOTENT_WRITE",
}

var ACLOperations_ACLOperation_value = map[string]int32{
	"UNKNOWN":          0,
	"ANY":              1,
	"READ":             2,
	"WRITE":            3,
	"CREATE":           4,
	"DELETE":           5,
	"ALTER":            6,
	"DESCRIBE":         7,
	"CLUSTER_ACTION":   8,
	"DESCRIBE_CONFIGS": 9,
	"ALTER_CONFIGS":    10,
	"IDEMPOTENT_WRITE": 11,
}

type ACLPermissionTypes_ACLPermissionType int32

const (
	ACLPermissionTypes_UNKNOWN ACLPermissionTypes_ACLPermissionType = 0
	ACLPermissionTypes_ANY     ACLPermissionTypes_ACLPermissionType = 1
	ACLPermissionTypes_ALLOW   ACLPermissionTypes_ACLPermissionType = 2
	ACLPermissionTypes_DENY    ACLPermissionTypes_ACLPermissionType = 3
)

func (x ACLPermissionTypes_ACLPermissionType) String() string {
	return proto.EnumName(ACLPermissionTypes_ACLPermissionType_name, int32(x))
}

var ACLPermissionTypes_ACLPermissionType_name = map[int32]string{
	0: "UNKNOWN",
	1: "ANY",
	2: "ALLOW",
	3: "DENY",
}

// ACLFilter provides the criteria for matching  ACLBindings
// Apache Kafka reference:
// https://github.com/apache/kafka/blob/trunk/clients/src/main/java/org/apache/kafka/common/acl/AclBindingFilter.java
type ACLFilter struct {
	PatternFilter        *ResourcePatternConfig    `protobuf:"bytes,1,opt,name=pattern_filter,json=patternFilter,proto3" json:"patternFilter" db:"pattern_filter,omitempty" url:"pattern_filter,omitempty"`
	EntryFilter          *AccessControlEntryConfig `protobuf:"bytes,2,opt,name=entry_filter,json=entryFilter,proto3" json:"entryFilter" db:"entry_filter,omitempty" url:"entry_filter,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                  `json:"-"`
	XXX_unrecognized     []byte                    `json:"-"`
	XXX_sizecache        int32                     `json:"-"`
}
