# V1alpha1StatementStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Phase** | **string** | Denotes the current state of the submitted SQL job. | [readonly] 
**Detail** | Pointer to **string** | Error details if phase is \&quot;failed\&quot;. For other phase states, it will contain human readable description of what each state represents. | [optional] 

## Methods

### NewV1alpha1StatementStatus

`func NewV1alpha1StatementStatus(phase string, ) *V1alpha1StatementStatus`

NewV1alpha1StatementStatus instantiates a new V1alpha1StatementStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewV1alpha1StatementStatusWithDefaults

`func NewV1alpha1StatementStatusWithDefaults() *V1alpha1StatementStatus`

NewV1alpha1StatementStatusWithDefaults instantiates a new V1alpha1StatementStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPhase

`func (o *V1alpha1StatementStatus) GetPhase() string`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *V1alpha1StatementStatus) GetPhaseOk() (*string, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *V1alpha1StatementStatus) SetPhase(v string)`

SetPhase sets Phase field to given value.


### GetDetail

`func (o *V1alpha1StatementStatus) GetDetail() string`

GetDetail returns the Detail field if non-nil, zero value otherwise.

### GetDetailOk

`func (o *V1alpha1StatementStatus) GetDetailOk() (*string, bool)`

GetDetailOk returns a tuple with the Detail field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDetail

`func (o *V1alpha1StatementStatus) SetDetail(v string)`

SetDetail sets Detail field to given value.

### HasDetail

`func (o *V1alpha1StatementStatus) HasDetail() bool`

HasDetail returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


