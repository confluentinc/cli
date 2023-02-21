# V1alpha1Statement

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ApiVersion** | Pointer to **string** | APIVersion defines the schema version of this representation of a resource. | [optional] [readonly] 
**Kind** | Pointer to **string** | Kind defines the object this REST resource represents. | [optional] [readonly] 
**Id** | Pointer to **string** | ID of the SQL statement | [optional] 
**Metadata** | Pointer to [**ObjectMeta**](ObjectMeta.md) |  | [optional] 
**Properties** | Pointer to [**V1alphaProperties**](v1alpha.Properties.md) | Request/client properties. | [optional] 
**ComputePoolId** | Pointer to **string** | The SQL statement will be executed under the scope of this compute pool id. | [optional] 
**Statement** | Pointer to **string** | SQL statement executed. | [optional] 
**StatementKind** | Pointer to **string** | Type of the SQL statement as identified. | [optional] [readonly] 
**Status** | Pointer to [**V1alpha1StatementStatus**](V1alpha1StatementStatus.md) |  | [optional] 

## Methods

### NewV1alpha1Statement

`func NewV1alpha1Statement() *V1alpha1Statement`

NewV1alpha1Statement instantiates a new V1alpha1Statement object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewV1alpha1StatementWithDefaults

`func NewV1alpha1StatementWithDefaults() *V1alpha1Statement`

NewV1alpha1StatementWithDefaults instantiates a new V1alpha1Statement object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetApiVersion

`func (o *V1alpha1Statement) GetApiVersion() string`

GetApiVersion returns the ApiVersion field if non-nil, zero value otherwise.

### GetApiVersionOk

`func (o *V1alpha1Statement) GetApiVersionOk() (*string, bool)`

GetApiVersionOk returns a tuple with the ApiVersion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetApiVersion

`func (o *V1alpha1Statement) SetApiVersion(v string)`

SetApiVersion sets ApiVersion field to given value.

### HasApiVersion

`func (o *V1alpha1Statement) HasApiVersion() bool`

HasApiVersion returns a boolean if a field has been set.

### GetKind

`func (o *V1alpha1Statement) GetKind() string`

GetKind returns the Kind field if non-nil, zero value otherwise.

### GetKindOk

`func (o *V1alpha1Statement) GetKindOk() (*string, bool)`

GetKindOk returns a tuple with the Kind field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKind

`func (o *V1alpha1Statement) SetKind(v string)`

SetKind sets Kind field to given value.

### HasKind

`func (o *V1alpha1Statement) HasKind() bool`

HasKind returns a boolean if a field has been set.

### GetId

`func (o *V1alpha1Statement) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *V1alpha1Statement) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *V1alpha1Statement) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *V1alpha1Statement) HasId() bool`

HasId returns a boolean if a field has been set.

### GetMetadata

`func (o *V1alpha1Statement) GetMetadata() ObjectMeta`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *V1alpha1Statement) GetMetadataOk() (*ObjectMeta, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *V1alpha1Statement) SetMetadata(v ObjectMeta)`

SetMetadata sets Metadata field to given value.

### HasMetadata

`func (o *V1alpha1Statement) HasMetadata() bool`

HasMetadata returns a boolean if a field has been set.

### GetProperties

`func (o *V1alpha1Statement) GetProperties() V1alphaProperties`

GetProperties returns the Properties field if non-nil, zero value otherwise.

### GetPropertiesOk

`func (o *V1alpha1Statement) GetPropertiesOk() (*V1alphaProperties, bool)`

GetPropertiesOk returns a tuple with the Properties field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProperties

`func (o *V1alpha1Statement) SetProperties(v V1alphaProperties)`

SetProperties sets Properties field to given value.

### HasProperties

`func (o *V1alpha1Statement) HasProperties() bool`

HasProperties returns a boolean if a field has been set.

### GetComputePoolId

`func (o *V1alpha1Statement) GetComputePoolId() string`

GetComputePoolId returns the ComputePoolId field if non-nil, zero value otherwise.

### GetComputePoolIdOk

`func (o *V1alpha1Statement) GetComputePoolIdOk() (*string, bool)`

GetComputePoolIdOk returns a tuple with the ComputePoolId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetComputePoolId

`func (o *V1alpha1Statement) SetComputePoolId(v string)`

SetComputePoolId sets ComputePoolId field to given value.

### HasComputePoolId

`func (o *V1alpha1Statement) HasComputePoolId() bool`

HasComputePoolId returns a boolean if a field has been set.

### GetStatement

`func (o *V1alpha1Statement) GetStatement() string`

GetStatement returns the Statement field if non-nil, zero value otherwise.

### GetStatementOk

`func (o *V1alpha1Statement) GetStatementOk() (*string, bool)`

GetStatementOk returns a tuple with the Statement field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatement

`func (o *V1alpha1Statement) SetStatement(v string)`

SetStatement sets Statement field to given value.

### HasStatement

`func (o *V1alpha1Statement) HasStatement() bool`

HasStatement returns a boolean if a field has been set.

### GetStatementKind

`func (o *V1alpha1Statement) GetStatementKind() string`

GetStatementKind returns the StatementKind field if non-nil, zero value otherwise.

### GetStatementKindOk

`func (o *V1alpha1Statement) GetStatementKindOk() (*string, bool)`

GetStatementKindOk returns a tuple with the StatementKind field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatementKind

`func (o *V1alpha1Statement) SetStatementKind(v string)`

SetStatementKind sets StatementKind field to given value.

### HasStatementKind

`func (o *V1alpha1Statement) HasStatementKind() bool`

HasStatementKind returns a boolean if a field has been set.

### GetStatus

`func (o *V1alpha1Statement) GetStatus() V1alpha1StatementStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *V1alpha1Statement) GetStatusOk() (*V1alpha1StatementStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *V1alpha1Statement) SetStatus(v V1alpha1StatementStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *V1alpha1Statement) HasStatus() bool`

HasStatus returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


