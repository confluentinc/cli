# SqlV1alpha1StatementSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**StatementName** | Pointer to **string** | The name of the resource, unique within this environment. | [optional] 
**Statement** | Pointer to **string** | The raw SQL text statement. | [optional] 
**Properties** | Pointer to **map[string]string** | A map (key-value pairs) of statement properties. | [optional] 
**ComputePoolId** | Pointer to **string** | The id associated with the compute pool in context. | [optional] 

## Methods

### NewSqlV1alpha1StatementSpec

`func NewSqlV1alpha1StatementSpec() *SqlV1alpha1StatementSpec`

NewSqlV1alpha1StatementSpec instantiates a new SqlV1alpha1StatementSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSqlV1alpha1StatementSpecWithDefaults

`func NewSqlV1alpha1StatementSpecWithDefaults() *SqlV1alpha1StatementSpec`

NewSqlV1alpha1StatementSpecWithDefaults instantiates a new SqlV1alpha1StatementSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStatementName

`func (o *SqlV1alpha1StatementSpec) GetStatementName() string`

GetStatementName returns the StatementName field if non-nil, zero value otherwise.

### GetStatementNameOk

`func (o *SqlV1alpha1StatementSpec) GetStatementNameOk() (*string, bool)`

GetStatementNameOk returns a tuple with the StatementName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatementName

`func (o *SqlV1alpha1StatementSpec) SetStatementName(v string)`

SetStatementName sets StatementName field to given value.

### HasStatementName

`func (o *SqlV1alpha1StatementSpec) HasStatementName() bool`

HasStatementName returns a boolean if a field has been set.

### GetStatement

`func (o *SqlV1alpha1StatementSpec) GetStatement() string`

GetStatement returns the Statement field if non-nil, zero value otherwise.

### GetStatementOk

`func (o *SqlV1alpha1StatementSpec) GetStatementOk() (*string, bool)`

GetStatementOk returns a tuple with the Statement field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatement

`func (o *SqlV1alpha1StatementSpec) SetStatement(v string)`

SetStatement sets Statement field to given value.

### HasStatement

`func (o *SqlV1alpha1StatementSpec) HasStatement() bool`

HasStatement returns a boolean if a field has been set.

### GetProperties

`func (o *SqlV1alpha1StatementSpec) GetProperties() map[string]string`

GetProperties returns the Properties field if non-nil, zero value otherwise.

### GetPropertiesOk

`func (o *SqlV1alpha1StatementSpec) GetPropertiesOk() (*map[string]string, bool)`

GetPropertiesOk returns a tuple with the Properties field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProperties

`func (o *SqlV1alpha1StatementSpec) SetProperties(v map[string]string)`

SetProperties sets Properties field to given value.

### HasProperties

`func (o *SqlV1alpha1StatementSpec) HasProperties() bool`

HasProperties returns a boolean if a field has been set.

### GetComputePoolId

`func (o *SqlV1alpha1StatementSpec) GetComputePoolId() string`

GetComputePoolId returns the ComputePoolId field if non-nil, zero value otherwise.

### GetComputePoolIdOk

`func (o *SqlV1alpha1StatementSpec) GetComputePoolIdOk() (*string, bool)`

GetComputePoolIdOk returns a tuple with the ComputePoolId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetComputePoolId

`func (o *SqlV1alpha1StatementSpec) SetComputePoolId(v string)`

SetComputePoolId sets ComputePoolId field to given value.

### HasComputePoolId

`func (o *SqlV1alpha1StatementSpec) HasComputePoolId() bool`

HasComputePoolId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


