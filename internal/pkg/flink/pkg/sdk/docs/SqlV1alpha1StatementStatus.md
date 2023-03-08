# SqlV1alpha1StatementStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Phase** | **string** | The lifecyle phase of the submitted SQL statement:  PENDING: SQL statement is pending execution;  RUNNING: SQL statement execution is in progress;  COMPLETED: SQL statement is completed;  DELETING: SQL statement deletion is in progress;  FAILED: SQL statement execution has failed;  | [readonly] 
**ErrorMessage** | Pointer to **string** | Displayable error message if SQL statement is in a failed state | [optional] [readonly] 

## Methods

### NewSqlV1alpha1StatementStatus

`func NewSqlV1alpha1StatementStatus(phase string, ) *SqlV1alpha1StatementStatus`

NewSqlV1alpha1StatementStatus instantiates a new SqlV1alpha1StatementStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSqlV1alpha1StatementStatusWithDefaults

`func NewSqlV1alpha1StatementStatusWithDefaults() *SqlV1alpha1StatementStatus`

NewSqlV1alpha1StatementStatusWithDefaults instantiates a new SqlV1alpha1StatementStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPhase

`func (o *SqlV1alpha1StatementStatus) GetPhase() string`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *SqlV1alpha1StatementStatus) GetPhaseOk() (*string, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *SqlV1alpha1StatementStatus) SetPhase(v string)`

SetPhase sets Phase field to given value.


### GetErrorMessage

`func (o *SqlV1alpha1StatementStatus) GetErrorMessage() string`

GetErrorMessage returns the ErrorMessage field if non-nil, zero value otherwise.

### GetErrorMessageOk

`func (o *SqlV1alpha1StatementStatus) GetErrorMessageOk() (*string, bool)`

GetErrorMessageOk returns a tuple with the ErrorMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetErrorMessage

`func (o *SqlV1alpha1StatementStatus) SetErrorMessage(v string)`

SetErrorMessage sets ErrorMessage field to given value.

### HasErrorMessage

`func (o *SqlV1alpha1StatementStatus) HasErrorMessage() bool`

HasErrorMessage returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


