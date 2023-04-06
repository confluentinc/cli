# ListMeta

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Self** | Pointer to **string** | Self is a Uniform Resource Locator (URL) at which an object can be addressed. This URL encodes the service location, API version, and other particulars necessary to locate the resource at a point in time | [optional] 
**Next** | Pointer to **string** | An opaque token that can be sent to get the next batch of results | [optional] 
**CreatedAt** | Pointer to **time.Time** | The date and time at which this object was created. It is represented in RFC3339 format and is in UTC. | [optional] [readonly] 

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

### GetSelf

`func (o *ListMeta) GetSelf() string`

GetSelf returns the Self field if non-nil, zero value otherwise.

### GetSelfOk

`func (o *ListMeta) GetSelfOk() (*string, bool)`

GetSelfOk returns a tuple with the Self field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSelf

`func (o *ListMeta) SetSelf(v string)`

SetSelf sets Self field to given value.

### HasSelf

`func (o *ListMeta) HasSelf() bool`

HasSelf returns a boolean if a field has been set.

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

### GetCreatedAt

`func (o *ListMeta) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *ListMeta) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *ListMeta) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *ListMeta) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


