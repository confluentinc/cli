# ListMeta

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**First** | Pointer to **NullableString** | A link to the first page of results. If a response does not contain a first link, then direct navigation to the first page is not supported. | [optional] 
**Last** | Pointer to **NullableString** | A link to the last page of results. If a response does not contain a last link, then direct navigation to the last page is not supported. | [optional] 
**Prev** | Pointer to **NullableString** | A link to the previous page of results. If a response does not contain a prev link, then either there is no previous data or backwards traversal through the result set is not supported. | [optional] 
**Next** | Pointer to **NullableString** | A link to the next page of results. If a response does not contain a next link, then there is no more data available. | [optional] 
**TotalSize** | Pointer to **int32** | Number of records in the full result set. This response may be paginated and have a smaller number of records. | [optional] 

## Methods

### NewListMeta

`func NewListMeta() *ListMeta`

NewListMeta instantiates a new ListMeta object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewListMetaWithDefaults

`func NewListMetaWithDefaults() *ListMeta`

NewListMetaWithDefaults instantiates a new ListMeta object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFirst

`func (o *ListMeta) GetFirst() string`

GetFirst returns the First field if non-nil, zero value otherwise.

### GetFirstOk

`func (o *ListMeta) GetFirstOk() (*string, bool)`

GetFirstOk returns a tuple with the First field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFirst

`func (o *ListMeta) SetFirst(v string)`

SetFirst sets First field to given value.

### HasFirst

`func (o *ListMeta) HasFirst() bool`

HasFirst returns a boolean if a field has been set.

### SetFirstNil

`func (o *ListMeta) SetFirstNil(b bool)`

 SetFirstNil sets the value for First to be an explicit nil

### UnsetFirst
`func (o *ListMeta) UnsetFirst()`

UnsetFirst ensures that no value is present for First, not even an explicit nil
### GetLast

`func (o *ListMeta) GetLast() string`

GetLast returns the Last field if non-nil, zero value otherwise.

### GetLastOk

`func (o *ListMeta) GetLastOk() (*string, bool)`

GetLastOk returns a tuple with the Last field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLast

`func (o *ListMeta) SetLast(v string)`

SetLast sets Last field to given value.

### HasLast

`func (o *ListMeta) HasLast() bool`

HasLast returns a boolean if a field has been set.

### SetLastNil

`func (o *ListMeta) SetLastNil(b bool)`

 SetLastNil sets the value for Last to be an explicit nil

### UnsetLast
`func (o *ListMeta) UnsetLast()`

UnsetLast ensures that no value is present for Last, not even an explicit nil
### GetPrev

`func (o *ListMeta) GetPrev() string`

GetPrev returns the Prev field if non-nil, zero value otherwise.

### GetPrevOk

`func (o *ListMeta) GetPrevOk() (*string, bool)`

GetPrevOk returns a tuple with the Prev field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrev

`func (o *ListMeta) SetPrev(v string)`

SetPrev sets Prev field to given value.

### HasPrev

`func (o *ListMeta) HasPrev() bool`

HasPrev returns a boolean if a field has been set.

### SetPrevNil

`func (o *ListMeta) SetPrevNil(b bool)`

 SetPrevNil sets the value for Prev to be an explicit nil

### UnsetPrev
`func (o *ListMeta) UnsetPrev()`

UnsetPrev ensures that no value is present for Prev, not even an explicit nil
### GetNext

`func (o *ListMeta) GetNext() string`

GetNext returns the Next field if non-nil, zero value otherwise.

### GetNextOk

`func (o *ListMeta) GetNextOk() (*string, bool)`

GetNextOk returns a tuple with the Next field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNext

`func (o *ListMeta) SetNext(v string)`

SetNext sets Next field to given value.

### HasNext

`func (o *ListMeta) HasNext() bool`

HasNext returns a boolean if a field has been set.

### SetNextNil

`func (o *ListMeta) SetNextNil(b bool)`

 SetNextNil sets the value for Next to be an explicit nil

### UnsetNext
`func (o *ListMeta) UnsetNext()`

UnsetNext ensures that no value is present for Next, not even an explicit nil
### GetTotalSize

`func (o *ListMeta) GetTotalSize() int32`

GetTotalSize returns the TotalSize field if non-nil, zero value otherwise.

### GetTotalSizeOk

`func (o *ListMeta) GetTotalSizeOk() (*int32, bool)`

GetTotalSizeOk returns a tuple with the TotalSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalSize

`func (o *ListMeta) SetTotalSize(v int32)`

SetTotalSize sets TotalSize field to given value.

### HasTotalSize

`func (o *ListMeta) HasTotalSize() bool`

HasTotalSize returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


