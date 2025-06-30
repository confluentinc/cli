package flink

// Local struct for Statement
// ... localStatement and related types ...
type localStatement struct {
	ApiVersion string                 `yaml:"apiVersion" json:"apiVersion"`
	Kind       string                 `yaml:"kind" json:"kind"`
	Metadata   localStatementMetadata `yaml:"metadata" json:"metadata"`
	Spec       localStatementSpec     `yaml:"spec" json:"spec"`
	Status     *localStatementStatus  `yaml:"status,omitempty" json:"status,omitempty"`
	Result     *localStatementResult  `yaml:"result,omitempty" json:"result,omitempty"`
}

type localStatementMetadata struct {
	Name              string             `yaml:"name" json:"name"`
	CreationTimestamp *string            `yaml:"creationTimestamp,omitempty" json:"creationTimestamp,omitempty"`
	UpdateTimestamp   *string            `yaml:"updateTimestamp,omitempty" json:"updateTimestamp,omitempty"`
	Uid               *string            `yaml:"uid,omitempty" json:"uid,omitempty"`
	Labels            *map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	Annotations       *map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
}

type localStatementSpec struct {
	Statement          string                  `yaml:"statement" json:"statement"`
	Properties         *map[string]interface{} `yaml:"properties,omitempty" json:"properties,omitempty"`
	FlinkConfiguration *map[string]interface{} `yaml:"flinkConfiguration,omitempty" json:"flinkConfiguration,omitempty"`
	ComputePoolName    string                  `yaml:"computePoolName" json:"computePoolName"`
	Parallelism        interface{}             `yaml:"parallelism,omitempty" json:"parallelism,omitempty"`
	Stopped            *bool                   `yaml:"stopped,omitempty" json:"stopped,omitempty"`
}

type localStatementStatus struct {
	Phase  string                `yaml:"phase" json:"phase"`
	Detail *string               `yaml:"detail,omitempty" json:"detail,omitempty"`
	Traits *localStatementTraits `yaml:"traits,omitempty" json:"traits,omitempty"`
}

type localStatementTraits struct {
	SqlKind       *string            `yaml:"sqlKind,omitempty" json:"sqlKind,omitempty"`
	IsBounded     *bool              `yaml:"isBounded,omitempty" json:"isBounded,omitempty"`
	IsAppendOnly  *bool              `yaml:"isAppendOnly,omitempty" json:"isAppendOnly,omitempty"`
	UpsertColumns *[]int32           `yaml:"upsertColumns,omitempty" json:"upsertColumns,omitempty"`
	Schema        *localResultSchema `yaml:"schema,omitempty" json:"schema,omitempty"`
}

type localResultSchema struct {
	Columns []localResultSchemaColumn `yaml:"columns" json:"columns"`
}

type localResultSchemaColumn struct {
	Name string        `yaml:"name" json:"name"`
	Type localDataType `yaml:"type" json:"type"`
}

type localDataType struct {
	Type                string                `yaml:"type" json:"type"`
	Nullable            bool                  `yaml:"nullable" json:"nullable"`
	Length              *int32                `yaml:"length,omitempty" json:"length,omitempty"`
	Precision           *int32                `yaml:"precision,omitempty" json:"precision,omitempty"`
	Scale               *int32                `yaml:"scale,omitempty" json:"scale,omitempty"`
	KeyType             *localDataType        `yaml:"keyType,omitempty" json:"keyType,omitempty"`
	ValueType           *localDataType        `yaml:"valueType,omitempty" json:"valueType,omitempty"`
	ElementType         *localDataType        `yaml:"elementType,omitempty" json:"elementType,omitempty"`
	Fields              *[]localDataTypeField `yaml:"fields,omitempty" json:"fields,omitempty"`
	Resolution          *string               `yaml:"resolution,omitempty" json:"resolution,omitempty"`
	FractionalPrecision *int32                `yaml:"fractionalPrecision,omitempty" json:"fractionalPrecision,omitempty"`
}

type localDataTypeField struct {
	Name        string        `yaml:"name" json:"name"`
	FieldType   localDataType `yaml:"fieldType" json:"fieldType"`
	Description *string       `yaml:"description,omitempty" json:"description,omitempty"`
}

type localStatementResult struct {
	ApiVersion string                       `yaml:"apiVersion" json:"apiVersion"`
	Kind       string                       `yaml:"kind" json:"kind"`
	Metadata   localStatementResultMetadata `yaml:"metadata" json:"metadata"`
	Results    localStatementResults        `yaml:"results" json:"results"`
}

type localStatementResultMetadata struct {
	CreationTimestamp *string            `yaml:"creationTimestamp,omitempty" json:"creationTimestamp,omitempty"`
	Annotations       *map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
}

type localStatementResults struct {
	Data *[]map[string]interface{} `yaml:"data,omitempty" json:"data,omitempty"`
}

// Local struct for Catalog
// ... localCatalog and related types ...
type localCatalog struct {
	ApiVersion string                `yaml:"apiVersion" json:"apiVersion"`
	Kind       string                `yaml:"kind" json:"kind"`
	Metadata   localCatalogMetadata  `yaml:"metadata" json:"metadata"`
	Spec       localKafkaCatalogSpec `yaml:"spec" json:"spec"`
}

type localCatalogMetadata struct {
	Name              string             `yaml:"name" json:"name"`
	CreationTimestamp *string            `yaml:"creationTimestamp,omitempty" json:"creationTimestamp,omitempty"`
	Uid               *string            `yaml:"uid,omitempty" json:"uid,omitempty"`
	Labels            *map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	Annotations       *map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
}

type localKafkaCatalogSpec struct {
	SrInstance    localKafkaCatalogSpecSrInstance     `yaml:"srInstance" json:"srInstance"`
	KafkaClusters []localKafkaCatalogSpecKafkaCluster `yaml:"kafkaClusters" json:"kafkaClusters"`
}

type localKafkaCatalogSpecSrInstance struct {
	ConnectionConfig   map[string]interface{} `yaml:"connectionConfig" json:"connectionConfig"`
	ConnectionSecretId *string                `yaml:"connectionSecretId,omitempty" json:"connectionSecretId,omitempty"`
}

type localKafkaCatalogSpecKafkaCluster struct {
	DatabaseName       string                 `yaml:"databaseName" json:"databaseName"`
	ConnectionConfig   map[string]interface{} `yaml:"connectionConfig" json:"connectionConfig"`
	ConnectionSecretId *string                `yaml:"connectionSecretId,omitempty" json:"connectionSecretId,omitempty"`
}

// Local struct for ComputePool
// ... localComputePool and related types ...
type localComputePool struct {
	ApiVersion string                   `yaml:"apiVersion" json:"apiVersion"`
	Kind       string                   `yaml:"kind" json:"kind"`
	Metadata   localComputePoolMetadata `yaml:"metadata" json:"metadata"`
	Spec       localComputePoolSpec     `yaml:"spec" json:"spec"`
	Status     *map[string]interface{}  `yaml:"status,omitempty" json:"status,omitempty"`
}

type localComputePoolMetadata struct {
	Name              string             `yaml:"name" json:"name"`
	CreationTimestamp *string            `yaml:"creationTimestamp,omitempty" json:"creationTimestamp,omitempty"`
	Uid               *string            `yaml:"uid,omitempty" json:"uid,omitempty"`
	Labels            *map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	Annotations       *map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
}

type localComputePoolSpec struct {
	Type        string                       `yaml:"type" json:"type"`
	ClusterSpec *localComputePoolClusterSpec `yaml:"clusterSpec,omitempty" json:"clusterSpec,omitempty"`
}

type localComputePoolClusterSpec struct {
	FlinkConfiguration *map[string]interface{}      `yaml:"flinkConfiguration,omitempty" json:"flinkConfiguration,omitempty"`
	Job                *localComputePoolJob         `yaml:"job,omitempty" json:"job,omitempty"`
	ServiceAccount     string                       `yaml:"serviceAccount" json:"serviceAccount"`
	TaskManager        *localComputePoolTaskManager `yaml:"taskManager,omitempty" json:"taskManager,omitempty"`
}

type localComputePoolJob struct {
	Parallelism interface{} `yaml:"parallelism" json:"parallelism"`
}

type localComputePoolTaskManager struct {
	Resource localComputePoolResource `yaml:"resource" json:"resource"`
}

type localComputePoolResource struct {
	CPU    interface{} `yaml:"cpu" json:"cpu"`
	Memory string      `yaml:"memory" json:"memory"`
}

// Local struct for FlinkApplication
// ... localFlinkApplication and related types ...
type localFlinkApplication struct {
	ApiVersion string                        `yaml:"apiVersion" json:"apiVersion"`
	Kind       string                        `yaml:"kind" json:"kind"`
	Metadata   localFlinkApplicationMetadata `yaml:"metadata" json:"metadata"`
	Spec       localFlinkApplicationSpec     `yaml:"spec" json:"spec"`
	Status     *map[string]interface{}       `yaml:"status,omitempty" json:"status,omitempty"`
}

type localFlinkApplicationMetadata struct {
	Name              string             `yaml:"name" json:"name"`
	CreationTimestamp *string            `yaml:"creationTimestamp,omitempty" json:"creationTimestamp,omitempty"`
	Uid               *string            `yaml:"uid,omitempty" json:"uid,omitempty"`
	Labels            *map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	Annotations       *map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
}

type localFlinkApplicationSpec struct {
	FlinkConfiguration *map[string]interface{}           `yaml:"flinkConfiguration,omitempty" json:"flinkConfiguration,omitempty"`
	FlinkEnvironment   string                            `yaml:"flinkEnvironment" json:"flinkEnvironment"`
	FlinkVersion       string                            `yaml:"flinkVersion" json:"flinkVersion"`
	Image              string                            `yaml:"image" json:"image"`
	Job                *localFlinkApplicationJob         `yaml:"job,omitempty" json:"job,omitempty"`
	JobManager         *localFlinkApplicationJobManager  `yaml:"jobManager,omitempty" json:"jobManager,omitempty"`
	ServiceAccount     string                            `yaml:"serviceAccount" json:"serviceAccount"`
	TaskManager        *localFlinkApplicationTaskManager `yaml:"taskManager,omitempty" json:"taskManager,omitempty"`
}

type localFlinkApplicationJob struct {
	JarURI      string      `yaml:"jarURI" json:"jarURI"`
	Parallelism interface{} `yaml:"parallelism" json:"parallelism"`
	State       string      `yaml:"state" json:"state"`
	UpgradeMode string      `yaml:"upgradeMode" json:"upgradeMode"`
}

type localFlinkApplicationJobManager struct {
	Resource localFlinkApplicationResource `yaml:"resource" json:"resource"`
}

type localFlinkApplicationTaskManager struct {
	Resource localFlinkApplicationResource `yaml:"resource" json:"resource"`
}

type localFlinkApplicationResource struct {
	CPU    interface{} `yaml:"cpu" json:"cpu"`
	Memory string      `yaml:"memory" json:"memory"`
}

// Summary output struct for Flink Application list
// Used in command_application_list.go
type flinkApplicationSummaryOut struct {
	Name        string `human:"Name" serialized:"name"`
	Environment string `human:"Environment" serialized:"environment"`
	JobName     string `human:"Job Name" serialized:"job_name"`
	JobStatus   string `human:"Job Status" serialized:"job_status"`
}

// Local struct for ComputePool YAML output (on-prem)
// Used in command_compute_pool_create_onprem.go and command_compute_pool_describe_onprem.go
type localComputePoolOnPrem struct {
	ApiVersion string                  `yaml:"apiVersion" json:"apiVersion"`
	Kind       string                  `yaml:"kind" json:"kind"`
	Metadata   map[string]interface{}  `yaml:"metadata" json:"metadata"`
	Spec       map[string]interface{}  `yaml:"spec" json:"spec"`
	Status     *map[string]interface{} `yaml:"status,omitempty" json:"status,omitempty"`
}
