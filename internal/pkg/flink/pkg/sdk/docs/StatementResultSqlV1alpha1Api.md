# \StatementResultSqlV1alpha1Api

All URIs are relative to *https://flink.region.provider.confluent.cloud*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetSqlV1alpha1StatementResult**](StatementResultSqlV1alpha1Api.md#GetSqlV1alpha1StatementResult) | **Get** /sql/v1alpha1/environments/{environment_id}/statements/{statement_name}/results | Read Statement Result



## GetSqlV1alpha1StatementResult

> SqlV1alpha1StatementResult GetSqlV1alpha1StatementResult(ctx, environmentId, statementName).PageToken(pageToken).Execute()

Read Statement Result



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
    environmentId := "environmentId_example" // string | The unique identifier for the environment.
    statementName := "statementName_example" // string | The unique identifier for the statement.
    pageToken := "pageToken_example" // string | It contains the field offset in the CollectSinkFunction protocol. On the first request, it should be unset. The offset is assumed to start at 0. (optional)

    configuration := openapiclient.NewConfiguration()
    api_client := openapiclient.NewAPIClient(configuration)
    resp, r, err := api_client.StatementResultSqlV1alpha1Api.GetSqlV1alpha1StatementResult(context.Background(), environmentId, statementName).PageToken(pageToken).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `StatementResultSqlV1alpha1Api.GetSqlV1alpha1StatementResult``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetSqlV1alpha1StatementResult`: SqlV1alpha1StatementResult
    fmt.Fprintf(os.Stdout, "Response from `StatementResultSqlV1alpha1Api.GetSqlV1alpha1StatementResult`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**environmentId** | **string** | The unique identifier for the environment. | 
**statementName** | **string** | The unique identifier for the statement. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetSqlV1alpha1StatementResultRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **pageToken** | **string** | It contains the field offset in the CollectSinkFunction protocol. On the first request, it should be unset. The offset is assumed to start at 0. | 

### Return type

[**SqlV1alpha1StatementResult**](sql.v1alpha1.StatementResult.md)

### Authorization

[api-key](../README.md#api-key)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

