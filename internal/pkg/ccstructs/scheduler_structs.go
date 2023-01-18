package ccstructs

import (
	"time"

	proto "github.com/gogo/protobuf/proto"
	types "github.com/gogo/protobuf/types"
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

type KafkaCluster struct {
	Id   string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty" db:"id,omitempty" url:"id,omitempty"`
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty" url:"name" db:"name,omitempty"`
	// id of account, set by the auth/gateway service
	AccountId string `protobuf:"bytes,3,opt,name=account_id,json=accountId,proto3" json:"account_id,omitempty" url:"account_id" db:"account_id,omitempty"`
	// write throughput in MB/s (read only)
	NetworkIngress int32 `protobuf:"varint,4,opt,name=network_ingress,json=networkIngress,proto3" json:"network_ingress,omitempty" db:"network_ingress,omitempty" url:"network_ingress,omitempty"`
	// read throughput in MB/s (read only)
	NetworkEgress int32 `protobuf:"varint,5,opt,name=network_egress,json=networkEgress,proto3" json:"network_egress,omitempty" db:"network_egress,omitempty" url:"network_egress,omitempty"`
	// storage in GB
	Storage    int32         `protobuf:"varint,6,opt,name=storage,proto3" json:"storage,omitempty" db:"storage,omitempty" url:"storage,omitempty"`
	Durability Durability    `protobuf:"varint,7,opt,name=durability,proto3,enum=Durability" json:"durability,omitempty" db:"durability,omitempty" url:"durability,omitempty"`
	Status     ClusterStatus `protobuf:"varint,8,opt,name=status,proto3,enum=ClusterStatus" json:"status,omitempty" db:"status,omitempty" url:"status,omitempty"`
	// e.g. SASL_SSL://r0.kafka.confluent.aws.confluent.cloud:9092,r0.kafka.confluent.aws.confluent.cloud:9093,r0.kafka.confluent.aws.confluent.cloud:9094
	Endpoint string `protobuf:"bytes,9,opt,name=endpoint,proto3" json:"endpoint,omitempty" db:"endpoint,omitempty" url:"endpoint,omitempty"`
	// e.g. us-west-2
	Region   string           `protobuf:"bytes,10,opt,name=region,proto3" json:"region,omitempty" db:"region,omitempty" url:"region,omitempty"`
	Created  *types.Timestamp `protobuf:"bytes,11,opt,name=created,proto3" json:"created,omitempty" db:"created,omitempty" url:"created,omitempty"`
	Modified *types.Timestamp `protobuf:"bytes,12,opt,name=modified,proto3" json:"modified,omitempty" db:"modified,omitempty" url:"modified,omitempty"`
	// e.g. aws
	ServiceProvider string `protobuf:"bytes,13,opt,name=service_provider,json=serviceProvider,proto3" json:"service_provider,omitempty" db:"service_provider,omitempty" url:"service_provider,omitempty"`
	OrganizationId  int32  `protobuf:"varint,14,opt,name=organization_id,json=organizationId,proto3" json:"organization_id,omitempty" db:"organization_id,omitempty" url:"organization_id,omitempty"`
	// deprecated; use separate "durability" and "dedicated" attributes now
	Enterprise bool `protobuf:"varint,16,opt,name=enterprise,proto3" json:"enterprise,omitempty" db:"enterprise,omitempty" url:"enterprise,omitempty"`
	// internal only
	K8SClusterId string `protobuf:"bytes,17,opt,name=k8s_cluster_id,json=k8sClusterId,proto3" json:"k8s_cluster_id,omitempty" db:"k8s_cluster_id,omitempty" url:"k8s_cluster_id,omitempty"`
	// internal only
	PhysicalClusterId string `protobuf:"bytes,18,opt,name=physical_cluster_id,json=physicalClusterId,proto3" json:"physical_cluster_id,omitempty" db:"physical_cluster_id,omitempty" url:"physical_cluster_id,omitempty"`
	// hundredths of a cent
	PricePerHour uint64 `protobuf:"varint,19,opt,name=price_per_hour,json=pricePerHour,proto3" json:"price_per_hour,omitempty" db:"price_per_hour,omitempty" url:"price_per_hour,omitempty"`
	// in hundredths of a cent, the cost incurred by running this cluster so far in the current billing cycle
	AccruedThisCycle uint64 `protobuf:"varint,20,opt,name=accrued_this_cycle,json=accruedThisCycle,proto3" json:"accrued_this_cycle,omitempty" db:"accrued_this_cycle,omitempty" url:"accrued_this_cycle,omitempty"`
	LegacyEndpoint   bool   `protobuf:"varint,21,opt,name=legacy_endpoint,json=legacyEndpoint,proto3" json:"legacy_endpoint,omitempty" db:"legacy_endpoint,omitempty" url:"legacy_endpoint,omitempty"`
	// kafka, metrics
	Type string `protobuf:"bytes,22,opt,name=type,proto3" json:"type,omitempty" db:"type,omitempty" url:"type,omitempty"`
	// kafka-api HTTP endpoint for this cluster
	ApiEndpoint string `protobuf:"bytes,23,opt,name=api_endpoint,json=apiEndpoint,proto3" json:"api_endpoint,omitempty" db:"api_endpoint,omitempty" url:"api_endpoint,omitempty"`
	// internal_proxy
	InternalProxy bool `protobuf:"varint,24,opt,name=internal_proxy,json=internalProxy,proto3" json:"internal_proxy,omitempty" db:"internal_proxy,omitempty" url:"internal_proxy,omitempty"`
	// is_sla_enabled. for enterprise, always true; for non enterprise, true if the physical cluster durability is high
	IsSlaEnabled bool `protobuf:"varint,25,opt,name=is_sla_enabled,json=isSlaEnabled,proto3" json:"is_sla_enabled,omitempty" db:"is_sla_enabled,omitempty" url:"is_sla_enabled,omitempty"`
	// is_schedulable, used for CLI to set a physical cluster to be schedulable or not
	IsSchedulable bool `protobuf:"varint,26,opt,name=is_schedulable,json=isSchedulable,proto3" json:"is_schedulable,omitempty" db:"is_schedulable,omitempty" url:"is_schedulable,omitempty"`
	// deprecated; (See Deployment.Sku) whether this cluster is dedicated to a single customer
	Dedicated bool `protobuf:"varint,27,opt,name=dedicated,proto3" json:"dedicated,omitempty" db:"dedicated,omitempty" url:"dedicated,omitempty"`
	// maximum network ingress
	MaxNetworkIngress int32 `protobuf:"varint,29,opt,name=max_network_ingress,json=maxNetworkIngress,proto3" json:"max_network_ingress,omitempty" db:"max_network_ingress,omitempty" url:"max_network_ingress,omitempty"`
	// minimum network egress
	MaxNetworkEgress int32       `protobuf:"varint,30,opt,name=max_network_egress,json=maxNetworkEgress,proto3" json:"max_network_egress,omitempty" db:"max_network_egress,omitempty" url:"max_network_egress,omitempty"`
	Deployment       *Deployment `protobuf:"bytes,31,opt,name=deployment,proto3" json:"deployment,omitempty" db:"deployment,omitempty" url:"deployment,omitempty"`
	Cku              int32       `protobuf:"varint,32,opt,name=cku,proto3" json:"cku,omitempty" db:"cku,omitempty" url:"cku,omitempty"`
	// This field is meant for internal clients (e.g. UI) where implicit creation is desired.
	NetworkRegion *NetworkRegion `protobuf:"bytes,33,opt,name=network_region,json=networkRegion,proto3" json:"network_region,omitempty" db:"network_region,omitempty" url:"network_region,omitempty"`
	// DEPRECATED. See selected_network_type
	InitialNetworkType NetworkType `protobuf:"varint,34,opt,name=initial_network_type,json=initialNetworkType,proto3,enum=NetworkType" json:"initial_network_type,omitempty" db:"initial_network_type,omitempty" url:"initial_network_type,omitempty"`
	// used to store initial customer decision for networking type
	SelectedNetworkType NetworkType `protobuf:"varint,35,opt,name=selected_network_type,json=selectedNetworkType,proto3,enum=NetworkType" json:"selected_network_type,omitempty" db:"selected_network_type,omitempty" url:"selected_network_type,omitempty"`
	EncryptionKeyId     string      `protobuf:"bytes,36,opt,name=encryption_key_id,json=encryptionKeyId,proto3" json:"encryption_key_id,omitempty" db:"encryption_key_id,omitempty" url:"encryption_key_id,omitempty"`
	PendingCku          int32       `protobuf:"varint,37,opt,name=pending_cku,json=pendingCku,proto3" json:"pending_cku,omitempty" db:"pending_cku,omitempty" url:"pending_cku,omitempty"`
	IsExpandable        bool        `protobuf:"varint,38,opt,name=is_expandable,json=isExpandable,proto3" json:"is_expandable,omitempty" db:"is_expandable,omitempty" url:"is_expandable,omitempty"`
	InfiniteStorage     bool        `protobuf:"varint,39,opt,name=infinite_storage,json=infiniteStorage,proto3" json:"infinite_storage,omitempty" db:"infinite_storage,omitempty" url:"infinite_storage,omitempty"`
	// The maximum number of partitions a CKU tenant is allowed to have, computed from the number of CKUs
	MaxPartitions int32 `protobuf:"varint,40,opt,name=max_partitions,json=maxPartitions,proto3" json:"max_partitions,omitempty" db:"max_partitions,omitempty" url:"max_partitions,omitempty"`
	// kafka rest HTTP endpoint for this cluster
	RestEndpoint string `protobuf:"bytes,41,opt,name=rest_endpoint,json=restEndpoint,proto3" json:"rest_endpoint,omitempty" db:"rest_endpoint,omitempty" url:"rest_endpoint,omitempty"`
	IsShrinkable bool   `protobuf:"varint,42,opt,name=is_shrinkable,json=isShrinkable,proto3" json:"is_shrinkable,omitempty" db:"is_shrinkable,omitempty" url:"is_shrinkable,omitempty"`
	// Internal status that contains detailed status information about the cluster that will be used to derived the customer visible ClusterStatus
	InternalStatus *Status `protobuf:"bytes,43,opt,name=internal_status,json=internalStatus,proto3" json:"internal_status,omitempty" db:"internal_status,omitempty" url:"internal_status,omitempty"`
	OrgResourceId  string  `protobuf:"bytes,44,opt,name=org_resource_id,json=orgResourceId,proto3" json:"org_resource_id,omitempty" db:"org_resource_id,omitempty" url:"org_resource_id,omitempty"`
	// BYOKv1 API Confluent Cloud key ID reference
	ConfluentCloudKeyId  string   `protobuf:"bytes,48,opt,name=confluent_cloud_key_id,json=confluentCloudKeyId,proto3" json:"confluent_cloud_key_id,omitempty" db:"confluent_cloud_key_id,omitempty" url:"confluent_cloud_key_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *KafkaCluster) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

type Durability int32

type ClusterStatus int32

type NetworkType int32

type Deployment struct {
	// The unique identifier of this resource instance.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty" db:"id,omitempty" url:"id,omitempty"`
	// The time in which this resource was created.
	Created *time.Time `protobuf:"bytes,2,opt,name=created,proto3,stdtime" json:"created,omitempty" db:"created,omitempty" url:"created,omitempty"`
	// The time in which this resource was last udpated.
	Modified *time.Time `protobuf:"bytes,3,opt,name=modified,proto3,stdtime" json:"modified,omitempty" db:"modified,omitempty" url:"modified,omitempty"`
	// The time in which this resource was deactivated.
	Deactivated *time.Time `protobuf:"bytes,4,opt,name=deactivated,proto3,stdtime" json:"deactivated,omitempty" db:"deactivated,omitempty" url:"deactivated,omitempty"`
	// deprecated; use environment_id.
	AccountId string `protobuf:"bytes,5,opt,name=account_id,json=accountId,proto3" json:"account_id,omitempty" db:"account_id,omitempty" url:"account_id,omitempty"`
	// deprecated; use the NetworkAccess resource directly.
	// The network access modes that are enabled or disabled for this Deployment.
	NetworkAccess *Deployment_NetworkAccess `protobuf:"bytes,6,opt,name=network_access,json=networkAccess,proto3" json:"network_access,omitempty" db:"network_access,omitempty" url:"network_access,omitempty"`
	// The product tier associated with this resource
	// TODO: find out how to rename the type in a backward compatible manner.
	Sku Sku `protobuf:"varint,7,opt,name=sku,proto3,enum=Sku" json:"sku,omitempty" db:"sku,omitempty" url:"sku,omitempty"`
	// The NetworkRegion that this resource is associated with.
	NetworkRegionId string `protobuf:"bytes,8,opt,name=network_region_id,json=networkRegionId,proto3" json:"network_region_id,omitempty" db:"network_region_id,omitempty" url:"network_region_id,omitempty"`
	// The required zones from the customer.
	// CP resources in  this deployment will be scheduled in these zones.
	//
	// NOTE that network_region_id takes precedence over Provider.Cloud and Provider.Region
	Provider *Provider `protobuf:"bytes,9,opt,name=provider,proto3" json:"provider,omitempty" db:"provider,omitempty" url:"provider,omitempty"`
	// Will this deployment be single zone or multi zone?
	// This represents the "upper bound" on durability. This means that if a CP
	// component in this Deployment doesn't support MZ, then it will still be SZ.
	Durability Durability `protobuf:"varint,10,opt,name=durability,proto3,enum=Durability" json:"durability,omitempty" db:"durability,omitempty" url:"durability,omitempty"`
	// Environment associated with this resource.
	EnvironmentId string `protobuf:"bytes,11,opt,name=environment_id,json=environmentId,proto3" json:"environment_id,omitempty" db:"environment_id,omitempty" url:"environment_id,omitempty"`
	// Whether or not this Deployment has dedicated resources.
	Dedicated bool `protobuf:"varint,12,opt,name=dedicated,proto3" json:"dedicated,omitempty" db:"dedicated,omitempty" url:"dedicated,omitempty"`
	// Preserved physical resources about how to place the logical resources (LKC).
	CreationConstraint   *Deployment_CreationConstraint `protobuf:"bytes,13,opt,name=creation_constraint,json=creationConstraint,proto3" json:"creation_constraint,omitempty" db:"creation_constraint,omitempty" url:"creation_constraint,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                       `json:"-"`
	XXX_unrecognized     []byte                         `json:"-"`
	XXX_sizecache        int32                          `json:"-"`
}

// NetworkRegion is an abstraction that represents the region specific building
// blocks that make up a Network. This resource contains all the configuration
// necessary to bootstrap a regional instance of a Network.
type NetworkRegion struct {
	// The unique identifier of this resource instance.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty" db:"id,omitempty" url:"id,omitempty"`
	// Customer provided name for this resource.
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty" db:"name,omitempty" url:"name,omitempty"`
	// Customer provided description for this resource.
	Description string `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty" db:"description,omitempty" url:"description,omitempty"`
	// Which Cidr should Confluent use when provisiong this resource?
	RequestedCidr string `protobuf:"bytes,4,opt,name=requested_cidr,json=requestedCidr,proto3" json:"requested_cidr,omitempty" db:"requested_cidr,omitempty" url:"requested_cidr,omitempty"`
	// The provider configuration
	Provider *Provider `protobuf:"bytes,5,opt,name=provider,proto3" json:"provider,omitempty" db:"provider,omitempty" url:"provider,omitempty"`
	// The environment this resource belongs to. This is necessary for shard Networks.
	EnvironmentId string `protobuf:"bytes,6,opt,name=environment_id,json=environmentId,proto3" json:"environment_id,omitempty" db:"environment_id,omitempty" url:"environment_id,omitempty"`
	// These are the Confluent infrastructure details
	// that we choose to show to the clients.
	ServiceNetwork *NetworkRegion_ServiceNetwork `protobuf:"bytes,7,opt,name=service_network,json=serviceNetwork,proto3" json:"service_network,omitempty" db:"service_network,omitempty" url:"service_network,omitempty"`
	// This is the internal only site name.
	SiteName string `protobuf:"bytes,8,opt,name=site_name,json=siteName,proto3" json:"site_name,omitempty" db:"site_name,omitempty" url:"site_name,omitempty"`
	// The status for this resource
	Status *Status `protobuf:"bytes,9,opt,name=status,proto3" json:"status,omitempty" db:"status,omitempty" url:"status,omitempty"`
	// Is this NetworkRegion fronted by the v4 network data plane?
	SniEnabled bool `protobuf:"varint,10,opt,name=sni_enabled,json=sniEnabled,proto3" json:"sni_enabled,omitempty" db:"sni_enabled,omitempty" url:"sni_enabled,omitempty"`
	// The time in which this resource was created.
	Created *time.Time `protobuf:"bytes,11,opt,name=created,proto3,stdtime" json:"created,omitempty" db:"created,omitempty" url:"created,omitempty"`
	// The time in which this resource was last udpated.
	Modified *time.Time `protobuf:"bytes,12,opt,name=modified,proto3,stdtime" json:"modified,omitempty" db:"modified,omitempty" url:"modified,omitempty"`
	// The time in which this resource was deactivated.
	Deactivated *time.Time `protobuf:"bytes,13,opt,name=deactivated,proto3,stdtime" json:"deactivated,omitempty" db:"deactivated,omitempty" url:"deactivated,omitempty"`
	// If dedicated, only one account can schedule resources on this NetworkRegion.
	Dedicated bool `protobuf:"varint,14,opt,name=dedicated,proto3" json:"dedicated,omitempty" db:"dedicated,omitempty" url:"dedicated,omitempty"`
	// This field is present so we can transform v2 -> v1 for ORM manipulations.
	NetworkConnectionTypes []NetworkType `protobuf:"varint,15,rep,packed,name=network_connection_types,json=networkConnectionTypes,proto3,enum=NetworkType" json:"network_connection_types,omitempty" db:"network_connection_types,omitempty" url:"network_connection_types,omitempty"`
	// TODO - deprecate domain id after succesful migration to cert id
	// a unique ID identifying a dynamically-allocated DNS domain that is used for TLS endpoint addresses
	DomainId string `protobuf:"bytes,16,opt,name=domain_id,json=domainId,proto3" json:"domain_id,omitempty" db:"domain_id,omitempty" url:"domain_id,omitempty"`
	// a unique ID identifying a dynamically-allocated DNS domain that is used for TLS endpoint addresses
	CertificateId string `protobuf:"bytes,17,opt,name=certificate_id,json=certificateId,proto3" json:"certificate_id,omitempty" db:"certificate_id,omitempty" url:"certificate_id,omitempty"`
	// Return the supported network domain types
	DomainTypes []NetworkRegion_NetworkDomainType `protobuf:"varint,18,rep,packed,name=domain_types,json=domainTypes,proto3,enum=NetworkRegion_NetworkDomainType" json:"domain_types,omitempty" db:"domain_types,omitempty" url:"domain_types,omitempty"`
	// Return the dns domains for each supported network domain type
	// The key is the NetworkDomainType name
	DnsDomains           map[string]*NetworkRegion_DnsDomain `protobuf:"bytes,19,rep,name=dns_domains,json=dnsDomains,proto3" json:"dns_domains,omitempty" db:"dns_domains,omitempty" url:"dns_domains,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}                            `json:"-"`
	XXX_unrecognized     []byte                              `json:"-"`
	XXX_sizecache        int32                               `json:"-"`
}

// Status is a generic message that encapsulates
// status details and is used in various resources.
type Status struct {
	// The type of status.
	Type Status_Type `protobuf:"varint,1,opt,name=type,proto3,enum=Status_Type" json:"type,omitempty" db:"type,omitempty" url:"type,omitempty"`
	// Status details.
	Details string `protobuf:"bytes,2,opt,name=details,proto3" json:"details,omitempty" db:"details,omitempty" url:"details,omitempty"`
	// The time in which the status was last updated.
	LastModified *time.Time `protobuf:"bytes,3,opt,name=last_modified,json=lastModified,proto3,stdtime" json:"last_modified,omitempty" db:"last_modified,omitempty" url:"last_modified,omitempty"`
	// One word free formed reason word briefly describing the corresponding status and details,
	// it's used mainly for UI to easily match certain conditions associated with the
	// status type and details to make certain decisions.
	Reason               string   `protobuf:"bytes,4,opt,name=reason,proto3" json:"reason,omitempty" db:"reason,omitempty" url:"reason,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

type Deployment_NetworkAccess struct {
	// Public internet access settings.
	PublicInternet []*Deployment_NetworkAccess_PublicInternet `protobuf:"bytes,1,rep,name=public_internet,json=publicInternet,proto3" json:"public_internet,omitempty" db:"public_internet,omitempty" url:"public_internet,omitempty"`
	// VPC Peering access settings.
	VpcPeering []*Deployment_NetworkAccess_VPCPeering `protobuf:"bytes,2,rep,name=vpc_peering,json=vpcPeering,proto3" json:"vpc_peering,omitempty" db:"vpc_peering,omitempty" url:"vpc_peering,omitempty"`
	// Private Link access settings.
	PrivateLink []*Deployment_NetworkAccess_PrivateLink `protobuf:"bytes,3,rep,name=private_link,json=privateLink,proto3" json:"private_link,omitempty" db:"private_link,omitempty" url:"private_link,omitempty"`
	// Transit Gateway access settings.
	TransitGateway []*Deployment_NetworkAccess_TransitGateway `protobuf:"bytes,4,rep,name=transit_gateway,json=transitGateway,proto3" json:"transit_gateway,omitempty" db:"transit_gateway,omitempty" url:"transit_gateway,omitempty"`
	// Internal access settings.
	Internal             []*Deployment_NetworkAccess_Internal `protobuf:"bytes,5,rep,name=internal,proto3" json:"internal,omitempty" db:"internal,omitempty" url:"internal,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                             `json:"-"`
	XXX_unrecognized     []byte                               `json:"-"`
	XXX_sizecache        int32                                `json:"-"`
}

type Provider struct {
	// The selected cloud provider.
	Cloud Provider_Cloud `protobuf:"varint,1,opt,name=cloud,proto3,enum=Provider_Cloud" json:"cloud,omitempty" db:"cloud,omitempty" url:"cloud,omitempty"`
	// The selected region.
	Region string `protobuf:"bytes,2,opt,name=region,proto3" json:"region,omitempty" db:"region,omitempty" url:"region,omitempty"`
	// The zones that are used for the NetworkRegion.
	Zones                []*AvailabilityZone `protobuf:"bytes,3,rep,name=zones,proto3" json:"zones,omitempty" db:"zones,omitempty" url:"zones,omitempty"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-"`
	XXX_unrecognized     []byte              `json:"-"`
	XXX_sizecache        int32               `json:"-"`
}

type Deployment_CreationConstraint struct {
	K8SClusterId         string   `protobuf:"bytes,1,opt,name=k8s_cluster_id,json=k8sClusterId,proto3" json:"k8s_cluster_id,omitempty" db:"k8s_cluster_id,omitempty" url:"k8s_cluster_id,omitempty"`
	PhysicalClusterId    string   `protobuf:"bytes,2,opt,name=physical_cluster_id,json=physicalClusterId,proto3" json:"physical_cluster_id,omitempty" db:"physical_cluster_id,omitempty" url:"physical_cluster_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

type NetworkRegion_ServiceNetwork struct {
	// Only one of the following may be set.
	Aws   *NetworkRegion_ServiceNetwork_Aws   `protobuf:"bytes,1,opt,name=aws,proto3" json:"aws,omitempty" db:"aws,omitempty" url:"aws,omitempty"`
	Gcp   *NetworkRegion_ServiceNetwork_Gcp   `protobuf:"bytes,2,opt,name=gcp,proto3" json:"gcp,omitempty" db:"gcp,omitempty" url:"gcp,omitempty"`
	Azure *NetworkRegion_ServiceNetwork_Azure `protobuf:"bytes,3,opt,name=azure,proto3" json:"azure,omitempty" db:"azure,omitempty" url:"azure,omitempty"`
	// dns_domain will be the following for v3 and v4 (post CoreDNS) network architectures:
	// v3: $region.$domain
	// v4: $networkRegionId.$region.glb.$domain
	DnsDomain string `protobuf:"bytes,4,opt,name=dns_domain,json=dnsDomain,proto3" json:"dns_domain,omitempty" db:"dns_domain,omitempty" url:"dns_domain,omitempty"`
	// deprecated
	ZonalSubdomains map[string]string `protobuf:"bytes,5,rep,name=zonal_subdomains,json=zonalSubdomains,proto3" json:"zonal_subdomains,omitempty" db:"zonal_subdomains,omitempty" url:"zonal_subdomains,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"` // Deprecated: Do not use.
	// The DNS domain for a given zone.
	// key: zone id, value: domain
	ZonalDomains map[string]string `protobuf:"bytes,6,rep,name=zonal_domains,json=zonalDomains,proto3" json:"zonal_domains,omitempty" db:"zonal_domains,omitempty" url:"zonal_domains,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// The suffix that can be concatenated to the name for brokers in the zone.
	// key: zone id, value: Suffix
	GlbZoneSuffixes map[string]string `protobuf:"bytes,7,rep,name=glb_zone_suffixes,json=glbZoneSuffixes,proto3" json:"glb_zone_suffixes,omitempty" db:"glb_zone_suffixes,omitempty" url:"glb_zone_suffixes,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// List of IPs that may generate outbound traffic from this region
	EgressIps            []string `protobuf:"bytes,8,rep,name=egress_ips,json=egressIps,proto3" json:"egress_ips,omitempty" db:"egress_ips,omitempty" url:"egress_ips,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

type NetworkRegion_NetworkDomainType int32

type NetworkRegion_DnsDomain struct {
	DnsDomain            string            `protobuf:"bytes,1,opt,name=dns_domain,json=dnsDomain,proto3" json:"dns_domain,omitempty" db:"dns_domain,omitempty" url:"dns_domain,omitempty"`
	ZonalDomains         map[string]string `protobuf:"bytes,2,rep,name=zonal_domains,json=zonalDomains,proto3" json:"zonal_domains,omitempty" db:"zonal_domains,omitempty" url:"zonal_domains,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

type Status_Type int32

type Deployment_NetworkAccess_PublicInternet struct {
	// Should this Deployment be accessible over the internet?
	Enabled bool `protobuf:"varint,1,opt,name=enabled,proto3" json:"enabled,omitempty" db:"enabled,omitempty" url:"enabled,omitempty"`
	// The network configs that should be used for this Deployment.
	NetworkConfigId string `protobuf:"bytes,3,opt,name=network_config_id,json=networkConfigId,proto3" json:"network_config_id,omitempty" db:"network_config_id,omitempty" url:"network_config_id,omitempty"`
	// Whitelisted Cidr ranges.
	AllowedCidrBlocks    []string `protobuf:"bytes,4,rep,name=allowed_cidr_blocks,json=allowedCidrBlocks,proto3" json:"allowed_cidr_blocks,omitempty" db:"allowed_cidr_blocks,omitempty" url:"allowed_cidr_blocks,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

type Deployment_NetworkAccess_VPCPeering struct {
	// deprecated; present for backward compatability reasons.
	Enabled bool `protobuf:"varint,1,opt,name=enabled,proto3" json:"enabled,omitempty" db:"enabled,omitempty" url:"enabled,omitempty"`
	// The network configs that should be used for this Deployment.
	// NOTE:  Mutating this field for VPC peering configs are not supported at this time.
	//        All VPC NetworkConfigs configured for the associated NetworkRegion will be enabled.
	NetworkConfigId      string   `protobuf:"bytes,3,opt,name=network_config_id,json=networkConfigId,proto3" json:"network_config_id,omitempty" db:"network_config_id,omitempty" url:"network_config_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

type Deployment_NetworkAccess_PrivateLink struct {
	// deprecated; present for backward compatability reasons.
	Enabled bool `protobuf:"varint,1,opt,name=enabled,proto3" json:"enabled,omitempty" db:"enabled,omitempty" url:"enabled,omitempty"`
	// Whitelisted source vpc endpoints.
	AllowedVpcEndpoints []string `protobuf:"bytes,2,rep,name=allowed_vpc_endpoints,json=allowedVpcEndpoints,proto3" json:"allowed_vpc_endpoints,omitempty" db:"allowed_vpc_endpoints,omitempty" url:"allowed_vpc_endpoints,omitempty"`
	// The network configs that should be used for this Deployment.
	NetworkConfigId      string   `protobuf:"bytes,3,opt,name=network_config_id,json=networkConfigId,proto3" json:"network_config_id,omitempty" db:"network_config_id,omitempty" url:"network_config_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

type Deployment_NetworkAccess_TransitGateway struct {
	// deprecated; present for backward compatability reasons.
	Enabled bool `protobuf:"varint,1,opt,name=enabled,proto3" json:"enabled,omitempty" db:"enabled,omitempty" url:"enabled,omitempty"`
	// The network configs that should be used for this Deployment.
	NetworkConfigId      string   `protobuf:"bytes,2,opt,name=network_config_id,json=networkConfigId,proto3" json:"network_config_id,omitempty" db:"network_config_id,omitempty" url:"network_config_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

type Deployment_NetworkAccess_Internal struct {
	Enabled              bool     `protobuf:"varint,1,opt,name=enabled,proto3" json:"enabled,omitempty" db:"enabled,omitempty" url:"enabled,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

type Provider_Cloud int32

type AvailabilityZone struct {
	// us-west-2
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty" db:"name,omitempty" url:"name,omitempty"`
	// AWS:   zone id
	// GCP:   zone name
	// AZURE: zone name
	ZoneId string `protobuf:"bytes,3,opt,name=zone_id,json=zoneId,proto3" json:"zone_id,omitempty" db:"zone_id,omitempty" url:"zone_id,omitempty"`
	// The internal mothership zone id
	Id                   string           `protobuf:"bytes,4,opt,name=id,proto3" json:"id,omitempty" db:"id,omitempty" url:"id,omitempty"`
	RegionId             string           `protobuf:"bytes,5,opt,name=region_id,json=regionId,proto3" json:"region_id,omitempty" db:"region_id,omitempty" url:"region_id,omitempty"`
	SniEnabled           *types.BoolValue `protobuf:"bytes,6,opt,name=sni_enabled,json=sniEnabled,proto3" json:"sni_enabled,omitempty" db:"sni_enabled,omitempty" url:"sni_enabled,omitempty"`
	Schedulable          *types.BoolValue `protobuf:"bytes,7,opt,name=schedulable,proto3" json:"schedulable,omitempty" db:"schedulable,omitempty" url:"schedulable,omitempty"`
	Created              *time.Time       `protobuf:"bytes,8,opt,name=created,proto3,stdtime" json:"created,omitempty" db:"created,omitempty" url:"created,omitempty"`
	Modified             *time.Time       `protobuf:"bytes,9,opt,name=modified,proto3,stdtime" json:"modified,omitempty" db:"modified,omitempty" url:"modified,omitempty"`
	Deactivated          *time.Time       `protobuf:"bytes,10,opt,name=deactivated,proto3,stdtime" json:"deactivated,omitempty" db:"deactivated,omitempty" url:"deactivated,omitempty"`
	Realm                string           `protobuf:"bytes,12,opt,name=realm,proto3" json:"realm,omitempty" db:"realm,omitempty" url:"realm,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

type NetworkRegion_ServiceNetwork_Aws struct {
	AccountId                  string   `protobuf:"bytes,1,opt,name=account_id,json=accountId,proto3" json:"account_id,omitempty" db:"account_id,omitempty" url:"account_id,omitempty"`
	VpcId                      string   `protobuf:"bytes,2,opt,name=vpc_id,json=vpcId,proto3" json:"vpc_id,omitempty" db:"vpc_id,omitempty" url:"vpc_id,omitempty"`
	PrivateLinkEndpointService string   `protobuf:"bytes,3,opt,name=private_link_endpoint_service,json=privateLinkEndpointService,proto3" json:"private_link_endpoint_service,omitempty" db:"private_link_endpoint_service,omitempty" url:"private_link_endpoint_service,omitempty"`
	XXX_NoUnkeyedLiteral       struct{} `json:"-"`
	XXX_unrecognized           []byte   `json:"-"`
	XXX_sizecache              int32    `json:"-"`
}

type NetworkRegion_ServiceNetwork_Gcp struct {
	ProjectId                               string            `protobuf:"bytes,1,opt,name=project_id,json=projectId,proto3" json:"project_id,omitempty" db:"project_id,omitempty" url:"project_id,omitempty"`
	VpcNetworkName                          string            `protobuf:"bytes,2,opt,name=vpc_network_name,json=vpcNetworkName,proto3" json:"vpc_network_name,omitempty" db:"vpc_network_name,omitempty" url:"vpc_network_name,omitempty"`
	PrivateServiceConnectServiceAttachments map[string]string `protobuf:"bytes,3,rep,name=private_service_connect_service_attachments,json=privateServiceConnectServiceAttachments,proto3" json:"private_service_connect_service_attachments,omitempty" db:"private_service_connect_service_attachments,omitempty" url:"private_service_connect_service_attachments,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral                    struct{}          `json:"-"`
	XXX_unrecognized                        []byte            `json:"-"`
	XXX_sizecache                           int32             `json:"-"`
}

type NetworkRegion_ServiceNetwork_Azure struct {
	SubscriptionId          string            `protobuf:"bytes,1,opt,name=subscription_id,json=subscriptionId,proto3" json:"subscription_id,omitempty" db:"subscription_id,omitempty" url:"subscription_id,omitempty"`
	VnetName                string            `protobuf:"bytes,3,opt,name=vnet_name,json=vnetName,proto3" json:"vnet_name,omitempty" db:"vnet_name,omitempty" url:"vnet_name,omitempty"`
	VnetResourceGroupName   string            `protobuf:"bytes,4,opt,name=vnet_resource_group_name,json=vnetResourceGroupName,proto3" json:"vnet_resource_group_name,omitempty" db:"vnet_resource_group_name,omitempty" url:"vnet_resource_group_name,omitempty"`
	PrivateLinkServiceAlias map[string]string `protobuf:"bytes,5,rep,name=private_link_service_alias,json=privateLinkServiceAlias,proto3" json:"private_link_service_alias,omitempty" db:"private_link_service_alias,omitempty" url:"private_link_service_alias,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral    struct{}          `json:"-"`
	XXX_unrecognized        []byte            `json:"-"`
	XXX_sizecache           int32             `json:"-"`
}
