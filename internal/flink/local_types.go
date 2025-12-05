package flink

import "time"

type LocalAllStatementDefaults1 struct {
	Detached    *LocalStatementDefaults `json:"detached,omitempty" yaml:"detached,omitempty"`
	Interactive *LocalStatementDefaults `json:"interactive,omitempty" yaml:"interactive,omitempty"`
}

type LocalCatalogMetadata struct {
	Name              string             `json:"name" yaml:"name"`
	CreationTimestamp *string            `json:"creationTimestamp,omitempty" yaml:"creationTimestamp,omitempty"`
	Uid               *string            `json:"uid,omitempty" yaml:"uid,omitempty"`
	Labels            *map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations       *map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

type LocalSavepoint struct {
	ApiVersion string                 `json:"apiVersion" yaml:"apiVersion"`
	Kind       string                 `json:"kind" yaml:"kind"`
	Metadata   LocalSavepointMetadata `json:"metadata" yaml:"metadata"`
	Spec       LocalSavepointSpec     `json:"spec" yaml:"spec"`
	Status     *LocalSavepointStatus  `json:"status,omitempty" yaml:"status,omitempty"`
}

type LocalSavepointMetadata struct {
	Name              string             `json:"name" yaml:"name"`
	CreationTimestamp *string            `json:"creationTimestamp,omitempty" yaml:"creationTimestamp,omitempty"`
	Uid               *string            `json:"uid,omitempty" yaml:"uid,omitempty"`
	Labels            *map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations       *map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

type LocalSavepointSpec struct {
	Path *string `json:"path,omitempty" yaml:"path,omitempty"`

	BackoffLimit *int32 `json:"backoffLimit,omitempty" yaml:"backoffLimit,omitempty"`

	FormatType *string `json:"formatType,omitempty" yaml:"formatType,omitempty"`
}

type LocalSavepointStatus struct {
	State *string `json:"state,omitempty" yaml:"state,omitempty"`

	Path *string `json:"path,omitempty" yaml:"path,omitempty"`

	TriggerTimestamp *string `json:"triggerTimestamp,omitempty" yaml:"triggerTimestamp,omitempty"`

	ResultTimestamp *string `json:"resultTimestamp,omitempty" yaml:"resultTimestamp,omitempty"`

	Failures *int32 `json:"failures,omitempty" yaml:"failures,omitempty"`

	Error *string `json:"error,omitempty" yaml:"error,omitempty"`

	PendingDeletion *bool `json:"pendingDeletion,omitempty" yaml:"pendingDeletion,omitempty"`
}

type LocalComputePool struct {
	ApiVersion string                   `json:"apiVersion" yaml:"apiVersion"`
	Kind       string                   `json:"kind" yaml:"kind"`
	Metadata   LocalComputePoolMetadata `json:"metadata" yaml:"metadata"`
	Spec       LocalComputePoolSpec     `json:"spec" yaml:"spec"`
	Status     *LocalComputePoolStatus  `json:"status,omitempty" yaml:"status,omitempty"`
}

type LocalComputePoolMetadata struct {
	Name              string             `json:"name" yaml:"name"`
	CreationTimestamp *string            `json:"creationTimestamp,omitempty" yaml:"creationTimestamp,omitempty"`
	Uid               *string            `json:"uid,omitempty" yaml:"uid,omitempty"`
	Labels            *map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations       *map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

type LocalComputePoolSpec struct {
	Type        string                 `json:"type" yaml:"type"`
	ClusterSpec map[string]interface{} `json:"clusterSpec" yaml:"clusterSpec"`
}

type LocalComputePoolStatus struct {
	Phase string `json:"phase" yaml:"phase"`
}

type LocalDataType struct {
	Type                string                `json:"type" yaml:"type"`
	Nullable            bool                  `json:"nullable" yaml:"nullable"`
	Length              *int32                `json:"length,omitempty" yaml:"length,omitempty"`
	Precision           *int32                `json:"precision,omitempty" yaml:"precision,omitempty"`
	Scale               *int32                `json:"scale,omitempty" yaml:"scale,omitempty"`
	KeyType             *LocalDataType        `json:"keyType,omitempty" yaml:"keyType,omitempty"`
	ValueType           *LocalDataType        `json:"valueType,omitempty" yaml:"valueType,omitempty"`
	ElementType         *LocalDataType        `json:"elementType,omitempty" yaml:"elementType,omitempty"`
	Fields              *[]LocalDataTypeField `json:"fields,omitempty" yaml:"fields,omitempty"`
	Resolution          *string               `json:"resolution,omitempty" yaml:"resolution,omitempty"`
	FractionalPrecision *int32                `json:"fractionalPrecision,omitempty" yaml:"fractionalPrecision,omitempty"`
}

type LocalDataTypeField struct {
	Name        string        `json:"name" yaml:"name"`
	FieldType   LocalDataType `json:"fieldType" yaml:"fieldType"`
	Description *string       `json:"description,omitempty" yaml:"description,omitempty"`
}

type LocalEnvironment struct {
	Secrets                  *map[string]string          `json:"secrets,omitempty" yaml:"secrets,omitempty"`
	Name                     string                      `json:"name" yaml:"name"`
	CreatedTime              *time.Time                  `json:"created_time,omitempty" yaml:"created_time,omitempty"`
	UpdatedTime              *time.Time                  `json:"updated_time,omitempty" yaml:"updated_time,omitempty"`
	FlinkApplicationDefaults *map[string]interface{}     `json:"flinkApplicationDefaults,omitempty" yaml:"flinkApplicationDefaults,omitempty"`
	KubernetesNamespace      string                      `json:"kubernetesNamespace" yaml:"kubernetesNamespace"`
	ComputePoolDefaults      *map[string]interface{}     `json:"computePoolDefaults,omitempty" yaml:"computePoolDefaults,omitempty"`
	StatementDefaults        *LocalAllStatementDefaults1 `json:"statementDefaults,omitempty" yaml:"statementDefaults,omitempty"`
}

type LocalFlinkApplication struct {
	ApiVersion string                  `json:"apiVersion" yaml:"apiVersion"`
	Kind       string                  `json:"kind" yaml:"kind"`
	Metadata   map[string]interface{}  `json:"metadata" yaml:"metadata"`
	Spec       map[string]interface{}  `json:"spec" yaml:"spec"`
	Status     *map[string]interface{} `json:"status,omitempty" yaml:"status,omitempty"`
}

type LocalKafkaCatalog struct {
	ApiVersion string                `json:"apiVersion" yaml:"apiVersion"`
	Kind       string                `json:"kind" yaml:"kind"`
	Metadata   LocalCatalogMetadata  `json:"metadata" yaml:"metadata"`
	Spec       LocalKafkaCatalogSpec `json:"spec" yaml:"spec"`
}

type LocalKafkaCatalogSpec struct {
	SrInstance    LocalKafkaCatalogSpecSrInstance      `json:"srInstance" yaml:"srInstance"`
	KafkaClusters []LocalKafkaCatalogSpecKafkaClusters `json:"kafkaClusters" yaml:"kafkaClusters"`
}

type LocalKafkaCatalogSpecKafkaClusters struct {
	DatabaseName       string            `json:"databaseName" yaml:"databaseName"`
	ConnectionConfig   map[string]string `json:"connectionConfig" yaml:"connectionConfig"`
	ConnectionSecretId *string           `json:"connectionSecretId,omitempty" yaml:"connectionSecretId,omitempty"`
}

type LocalKafkaCatalogSpecSrInstance struct {
	ConnectionConfig   map[string]string `json:"connectionConfig" yaml:"connectionConfig"`
	ConnectionSecretId *string           `json:"connectionSecretId,omitempty" yaml:"connectionSecretId,omitempty"`
}

type LocalResultSchema struct {
	Columns []LocalResultSchemaColumn `json:"columns" yaml:"columns"`
}

type LocalResultSchemaColumn struct {
	Name string        `json:"name" yaml:"name"`
	Type LocalDataType `json:"type" yaml:"type"`
}

type LocalStatement struct {
	ApiVersion string                 `json:"apiVersion" yaml:"apiVersion"`
	Kind       string                 `json:"kind" yaml:"kind"`
	Metadata   LocalStatementMetadata `json:"metadata" yaml:"metadata"`
	Spec       LocalStatementSpec     `json:"spec" yaml:"spec"`
	Status     *LocalStatementStatus  `json:"status,omitempty" yaml:"status,omitempty"`
	Result     *LocalStatementResult  `json:"result,omitempty" yaml:"result,omitempty"`
}

type LocalStatementDefaults struct {
	FlinkConfiguration *map[string]string `json:"flinkConfiguration,omitempty" yaml:"flinkConfiguration,omitempty"`
}

type LocalStatementMetadata struct {
	Name              string             `json:"name" yaml:"name"`
	CreationTimestamp *string            `json:"creationTimestamp,omitempty" yaml:"creationTimestamp,omitempty"`
	UpdateTimestamp   *string            `json:"updateTimestamp,omitempty" yaml:"updateTimestamp,omitempty"`
	Uid               *string            `json:"uid,omitempty" yaml:"uid,omitempty"`
	Labels            *map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations       *map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

type LocalStatementResult struct {
	ApiVersion string                       `json:"apiVersion" yaml:"apiVersion"`
	Kind       string                       `json:"kind" yaml:"kind"`
	Metadata   LocalStatementResultMetadata `json:"metadata" yaml:"metadata"`
	Results    LocalStatementResults        `json:"results" yaml:"results"`
}

type LocalStatementResultMetadata struct {
	CreationTimestamp *string            `json:"creationTimestamp,omitempty" yaml:"creationTimestamp,omitempty"`
	Annotations       *map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

type LocalStatementResults struct {
	Data *[]map[string]interface{} `json:"data,omitempty" yaml:"data,omitempty"`
}

type LocalStatementSpec struct {
	Statement          string             `json:"statement" yaml:"statement"`
	Properties         *map[string]string `json:"properties,omitempty" yaml:"properties,omitempty"`
	FlinkConfiguration *map[string]string `json:"flinkConfiguration,omitempty" yaml:"flinkConfiguration,omitempty"`
	ComputePoolName    string             `json:"computePoolName" yaml:"computePoolName"`
	Parallelism        *int32             `json:"parallelism,omitempty" yaml:"parallelism,omitempty"`
	Stopped            *bool              `json:"stopped,omitempty" yaml:"stopped,omitempty"`
}

type LocalStatementStatus struct {
	Phase  string                `json:"phase" yaml:"phase"`
	Detail *string               `json:"detail,omitempty" yaml:"detail,omitempty"`
	Traits *LocalStatementTraits `json:"traits,omitempty" yaml:"traits,omitempty"`
}

type LocalStatementTraits struct {
	SqlKind       *string            `json:"sqlKind,omitempty" yaml:"sqlKind,omitempty"`
	IsBounded     *bool              `json:"isBounded,omitempty" yaml:"isBounded,omitempty"`
	IsAppendOnly  *bool              `json:"isAppendOnly,omitempty" yaml:"isAppendOnly,omitempty"`
	UpsertColumns *[]int32           `json:"upsertColumns,omitempty" yaml:"upsertColumns,omitempty"`
	Schema        *LocalResultSchema `json:"schema,omitempty" yaml:"schema,omitempty"`
}
