# SqlV1alpha1Statement

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ApiVersion** | Pointer to **string** | APIVersion defines the schema version of this representation of a resource. | [optional] [readonly] 
**Kind** | Pointer to **string** | Kind defines the object this REST resource represents. | [optional] [readonly] 
**Id** | Pointer to **string** | ID is the \&quot;natural identifier\&quot; for an object within its scope/namespace; it is normally unique across time but not space. That is, you can assume that the ID will not be reclaimed and reused after an object is deleted (\&quot;time\&quot;); however, it may collide with IDs for other object &#x60;kinds&#x60; or objects of the same &#x60;kind&#x60; within a different scope/namespace (\&quot;space\&quot;). | [optional] [readonly] 
**Metadata** | Pointer to [**ObjectMeta**](ObjectMeta.md) |  | [optional] 
**Spec** | Pointer to [**SqlV1alpha1StatementSpec**](SqlV1alpha1StatementSpec.md) |  | [optional] 
**Status** | Pointer to [**SqlV1alpha1StatementStatus**](SqlV1alpha1StatementStatus.md) |  | [optional] 

## Methods

### NewSqlV1alpha1Statement

`func NewSqlV1alpha1Statement() *SqlV1alpha1Statement`

NewSqlV1alpha1Statement instantiates a new SqlV1alpha1Statement object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSqlV1alpha1StatementWithDefaults

`func NewSqlV1alpha1StatementWithDefaults() *SqlV1alpha1Statement`

NewSqlV1alpha1StatementWithDefaults instantiates a new SqlV1alpha1Statement object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetApiVersion

`func (o *SqlV1alpha1Statement) GetApiVersion() string`

GetApiVersion returns the ApiVersion field if non-nil, zero value otherwise.

### GetApiVersionOk

`func (o *SqlV1alpha1Statement) GetApiVersionOk() (*string, bool)`

GetApiVersionOk returns a tuple with the ApiVersion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetApiVersion

`func (o *SqlV1alpha1Statement) SetApiVersion(v string)`

SetApiVersion sets ApiVersion field to given value.

### HasApiVersion

`func (o *SqlV1alpha1Statement) HasApiVersion() bool`

HasApiVersion returns a boolean if a field has been set.

### GetKind

`func (o *SqlV1alpha1Statement) GetKind() string`

GetKind returns the Kind field if non-nil, zero value otherwise.

### GetKindOk

`func (o *SqlV1alpha1Statement) GetKindOk() (*string, bool)`

GetKindOk returns a tuple with the Kind field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKind

`func (o *SqlV1alpha1Statement) SetKind(v string)`

SetKind sets Kind field to given value.

### HasKind

`func (o *SqlV1alpha1Statement) HasKind() bool`

HasKind returns a boolean if a field has been set.

### GetId

`func (o *SqlV1alpha1Statement) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *SqlV1alpha1Statement) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *SqlV1alpha1Statement) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *SqlV1alpha1Statement) HasId() bool`

HasId returns a boolean if a field has been set.

### GetMetadata

`func (o *SqlV1alpha1Statement) GetMetadata() ObjectMeta`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *SqlV1alpha1Statement) GetMetadataOk() (*ObjectMeta, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *SqlV1alpha1Statement) SetMetadata(v ObjectMeta)`

SetMetadata sets Metadata field to given value.

### HasMetadata

`func (o *SqlV1alpha1Statement) HasMetadata() bool`

HasMetadata returns a boolean if a field has been set.

### GetSpec

`func (o *SqlV1alpha1Statement) GetSpec() SqlV1alpha1StatementSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *SqlV1alpha1Statement) GetSpecOk() (*SqlV1alpha1StatementSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *SqlV1alpha1Statement) SetSpec(v SqlV1alpha1StatementSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *SqlV1alpha1Statement) HasSpec() bool`

HasSpec returns a boolean if a field has been set.

### GetStatus

`func (o *SqlV1alpha1Statement) GetStatus() SqlV1alpha1StatementStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *SqlV1alpha1Statement) GetStatusOk() (*SqlV1alpha1StatementStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *SqlV1alpha1Statement) SetStatus(v SqlV1alpha1StatementStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *SqlV1alpha1Statement) HasStatus() bool`

HasStatus returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


