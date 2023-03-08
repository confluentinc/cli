# \StatementsSqlV1alpha1Api

All URIs are relative to *https://flink.region.provider.confluent.cloud*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateSqlV1alpha1Statement**](StatementsSqlV1alpha1Api.md#CreateSqlV1alpha1Statement) | **Post** /sql/v1alpha1/statements | Create a Statement
[**DeleteSqlV1alpha1Statement**](StatementsSqlV1alpha1Api.md#DeleteSqlV1alpha1Statement) | **Delete** /sql/v1alpha1/statements/{id} | Delete a Statement
[**GetSqlV1alpha1Statement**](StatementsSqlV1alpha1Api.md#GetSqlV1alpha1Statement) | **Get** /sql/v1alpha1/statements/{id} | Read a Statement
[**ListSqlV1alpha1Statements**](StatementsSqlV1alpha1Api.md#ListSqlV1alpha1Statements) | **Get** /sql/v1alpha1/statements | List of Statements



## CreateSqlV1alpha1Statement

> SqlV1alpha1Statement CreateSqlV1alpha1Statement(ctx).SqlV1alpha1Statement(sqlV1alpha1Statement).Execute()

Create a Statement



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    sqlV1alpha1Statement := *openapiclient.NewSqlV1alpha1Statement() // SqlV1alpha1Statement |  (optional)

    configuration := openapiclient.NewConfiguration()
    api_client := openapiclient.NewAPIClient(configuration)
    resp, r, err := api_client.StatementsSqlV1alpha1Api.CreateSqlV1alpha1Statement(context.Background()).SqlV1alpha1Statement(sqlV1alpha1Statement).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `StatementsSqlV1alpha1Api.CreateSqlV1alpha1Statement``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateSqlV1alpha1Statement`: SqlV1alpha1Statement
    fmt.Fprintf(os.Stdout, "Response from `StatementsSqlV1alpha1Api.CreateSqlV1alpha1Statement`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateSqlV1alpha1StatementRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **sqlV1alpha1Statement** | [**SqlV1alpha1Statement**](SqlV1alpha1Statement.md) |  | 

### Return type

[**SqlV1alpha1Statement**](sql.v1alpha1.Statement.md)

### Authorization

[api-key](../README.md#api-key), [confluent-sts-access-token](../README.md#confluent-sts-access-token)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteSqlV1alpha1Statement

> DeleteSqlV1alpha1Statement(ctx, id).Environment(environment).SpecComputePool(specComputePool).Execute()

Delete a Statement



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    environment := "env-00000" // string | Scope the operation to the given environment.
    specComputePool := "fcp-00000" // string | Scope the operation to the given spec.compute_pool.
    id := "id_example" // string | The unique identifier for the statement.

    configuration := openapiclient.NewConfiguration()
    api_client := openapiclient.NewAPIClient(configuration)
    resp, r, err := api_client.StatementsSqlV1alpha1Api.DeleteSqlV1alpha1Statement(context.Background(), id).Environment(environment).SpecComputePool(specComputePool).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `StatementsSqlV1alpha1Api.DeleteSqlV1alpha1Statement``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | The unique identifier for the statement. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteSqlV1alpha1StatementRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **environment** | **string** | Scope the operation to the given environment. | 
 **specComputePool** | **string** | Scope the operation to the given spec.compute_pool. | 


### Return type

 (empty response body)

### Authorization

[api-key](../README.md#api-key), [confluent-sts-access-token](../README.md#confluent-sts-access-token)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetSqlV1alpha1Statement

> SqlV1alpha1Statement GetSqlV1alpha1Statement(ctx, id).Environment(environment).SpecComputePool(specComputePool).Execute()

Read a Statement



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    environment := "env-00000" // string | Scope the operation to the given environment.
    specComputePool := "fcp-00000" // string | Scope the operation to the given spec.compute_pool.
    id := "id_example" // string | The unique identifier for the statement.

    configuration := openapiclient.NewConfiguration()
    api_client := openapiclient.NewAPIClient(configuration)
    resp, r, err := api_client.StatementsSqlV1alpha1Api.GetSqlV1alpha1Statement(context.Background(), id).Environment(environment).SpecComputePool(specComputePool).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `StatementsSqlV1alpha1Api.GetSqlV1alpha1Statement``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetSqlV1alpha1Statement`: SqlV1alpha1Statement
    fmt.Fprintf(os.Stdout, "Response from `StatementsSqlV1alpha1Api.GetSqlV1alpha1Statement`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | The unique identifier for the statement. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetSqlV1alpha1StatementRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **environment** | **string** | Scope the operation to the given environment. | 
 **specComputePool** | **string** | Scope the operation to the given spec.compute_pool. | 


### Return type

[**SqlV1alpha1Statement**](sql.v1alpha1.Statement.md)

### Authorization

[api-key](../README.md#api-key), [confluent-sts-access-token](../README.md#confluent-sts-access-token)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListSqlV1alpha1Statements

> SqlV1alpha1StatementList ListSqlV1alpha1Statements(ctx).Environment(environment).SpecComputePool(specComputePool).SpecStatementName(specStatementName).StatusPhase(statusPhase).PageSize(pageSize).PageToken(pageToken).Execute()

List of Statements



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    environment := "env-00000" // string | Filter the results by exact match for environment.
    specComputePool := "fcp-00000" // string | Filter the results by exact match for spec.compute_pool.
    specStatementName := *openapiclient.NewMultipleSearchFilter() // MultipleSearchFilter | Filter the results by exact match for spec.statement_name. Pass multiple times to see results matching any of the values. (optional)
    statusPhase := *openapiclient.NewMultipleSearchFilter() // MultipleSearchFilter | Filter the results by exact match for status.phase. Pass multiple times to see results matching any of the values. (optional)
    pageSize := int32(56) // int32 | A pagination size for collection requests. (optional) (default to 10)
    pageToken := "pageToken_example" // string | An opaque pagination token for collection requests. (optional)

    configuration := openapiclient.NewConfiguration()
    api_client := openapiclient.NewAPIClient(configuration)
    resp, r, err := api_client.StatementsSqlV1alpha1Api.ListSqlV1alpha1Statements(context.Background()).Environment(environment).SpecComputePool(specComputePool).SpecStatementName(specStatementName).StatusPhase(statusPhase).PageSize(pageSize).PageToken(pageToken).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `StatementsSqlV1alpha1Api.ListSqlV1alpha1Statements``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ListSqlV1alpha1Statements`: SqlV1alpha1StatementList
    fmt.Fprintf(os.Stdout, "Response from `StatementsSqlV1alpha1Api.ListSqlV1alpha1Statements`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListSqlV1alpha1StatementsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **environment** | **string** | Filter the results by exact match for environment. | 
 **specComputePool** | **string** | Filter the results by exact match for spec.compute_pool. | 
 **specStatementName** | [**MultipleSearchFilter**](MultipleSearchFilter.md) | Filter the results by exact match for spec.statement_name. Pass multiple times to see results matching any of the values. | 
 **statusPhase** | [**MultipleSearchFilter**](MultipleSearchFilter.md) | Filter the results by exact match for status.phase. Pass multiple times to see results matching any of the values. | 
 **pageSize** | **int32** | A pagination size for collection requests. | [default to 10]
 **pageToken** | **string** | An opaque pagination token for collection requests. | 

### Return type

[**SqlV1alpha1StatementList**](sql.v1alpha1.StatementList.md)

### Authorization

[api-key](../README.md#api-key), [confluent-sts-access-token](../README.md#confluent-sts-access-token)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

