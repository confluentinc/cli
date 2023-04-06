# SqlV1alpha1StatementList

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ApiVersion** | **string** | APIVersion defines the schema version of this representation of a resource. | 
**Kind** | **string** | Kind defines the object this REST resource represents. | 
**Metadata** | [**ListMeta**](ListMeta.md) |  | 
**Data** | [**[]SqlV1alpha1Statement**](SqlV1alpha1Statement.md) | A data property that contains an array of resource items. Each entry in the array is a separate resource. | 

## Methods

### NewSqlV1alpha1StatementList

`func NewSqlV1alpha1StatementList(apiVersion string, kind string, metadata ListMeta, data []SqlV1alpha1Statement, ) *SqlV1alpha1StatementList`

NewSqlV1alpha1StatementList instantiates a new SqlV1alpha1StatementList object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSqlV1alpha1StatementListWithDefaults

`func NewSqlV1alpha1StatementListWithDefaults() *SqlV1alpha1StatementList`

NewSqlV1alpha1StatementListWithDefaults instantiates a new SqlV1alpha1StatementList object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetApiVersion

`func (o *SqlV1alpha1StatementList) GetApiVersion() string`

GetApiVersion returns the ApiVersion field if non-nil, zero value otherwise.

### GetApiVersionOk

`func (o *SqlV1alpha1StatementList) GetApiVersionOk() (*string, bool)`

GetApiVersionOk returns a tuple with the ApiVersion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetApiVersion

`func (o *SqlV1alpha1StatementList) SetApiVersion(v string)`

SetApiVersion sets ApiVersion field to given value.


### GetKind

`func (o *SqlV1alpha1StatementList) GetKind() string`

GetKind returns the Kind field if non-nil, zero value otherwise.

### GetKindOk

`func (o *SqlV1alpha1StatementList) GetKindOk() (*string, bool)`

GetKindOk returns a tuple with the Kind field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKind

`func (o *SqlV1alpha1StatementList) SetKind(v string)`

SetKind sets Kind field to given value.


### GetMetadata

`func (o *SqlV1alpha1StatementList) GetMetadata() ListMeta`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *SqlV1alpha1StatementList) GetMetadataOk() (*ListMeta, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *SqlV1alpha1StatementList) SetMetadata(v ListMeta)`

SetMetadata sets Metadata field to given value.


### GetData

`func (o *SqlV1alpha1StatementList) GetData() []SqlV1alpha1Statement`

GetData returns the Data field if non-nil, zero value otherwise.

### GetDataOk

`func (o *SqlV1alpha1StatementList) GetDataOk() (*[]SqlV1alpha1Statement, bool)`

GetDataOk returns a tuple with the Data field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetData

`func (o *SqlV1alpha1StatementList) SetData(v []SqlV1alpha1Statement)`

SetData sets Data field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


