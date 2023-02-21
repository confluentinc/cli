# \StatementsV1alpha1Api

All URIs are relative to *https://api.confluent.cloud*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateV1alpha1Statement**](StatementsV1alpha1Api.md#CreateV1alpha1Statement) | **Post** /v1alpha1/org/{org_id}/env/{env_id}/statements | Create a Statement
[**DeleteV1alpha1Statement**](StatementsV1alpha1Api.md#DeleteV1alpha1Statement) | **Delete** /v1alpha1/org/{org_id}/env/{env_id}/statements/{id} | Delete a Statement
[**GetV1alpha1Statement**](StatementsV1alpha1Api.md#GetV1alpha1Statement) | **Get** /v1alpha1/org/{org_id}/env/{env_id}/statements/{id} | Read a Statement
[**ListV1alpha1Statements**](StatementsV1alpha1Api.md#ListV1alpha1Statements) | **Get** /v1alpha1/org/{org_id}/env/{env_id}/statements | List of Statements



## CreateV1alpha1Statement

> V1alpha1Statement CreateV1alpha1Statement(ctx, orgId, envId).V1alpha1Statement(v1alpha1Statement).Execute()

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
    orgId := "orgId_example" // string | The Org
    envId := "envId_example" // string | The Env
    v1alpha1Statement := *openapiclient.NewV1alpha1Statement() // V1alpha1Statement |  (optional)

    configuration := openapiclient.NewConfiguration()
    api_client := openapiclient.NewAPIClient(configuration)
    resp, r, err := api_client.StatementsV1alpha1Api.CreateV1alpha1Statement(context.Background(), orgId, envId).V1alpha1Statement(v1alpha1Statement).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `StatementsV1alpha1Api.CreateV1alpha1Statement``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateV1alpha1Statement`: V1alpha1Statement
    fmt.Fprintf(os.Stdout, "Response from `StatementsV1alpha1Api.CreateV1alpha1Statement`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The Org | 
**envId** | **string** | The Env | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateV1alpha1StatementRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **v1alpha1Statement** | [**V1alpha1Statement**](V1alpha1Statement.md) |  | 

### Return type

[**V1alpha1Statement**](v1alpha1.Statement.md)

### Authorization

[api-key](../README.md#api-key)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteV1alpha1Statement

> DeleteV1alpha1Statement(ctx, orgId, envId, id).Execute()

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
    orgId := "orgId_example" // string | The Org
    envId := "envId_example" // string | The Env
    id := "id_example" // string | The unique identifier for the statement.

    configuration := openapiclient.NewConfiguration()
    api_client := openapiclient.NewAPIClient(configuration)
    resp, r, err := api_client.StatementsV1alpha1Api.DeleteV1alpha1Statement(context.Background(), orgId, envId, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `StatementsV1alpha1Api.DeleteV1alpha1Statement``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The Org | 
**envId** | **string** | The Env | 
**id** | **string** | The unique identifier for the statement. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteV1alpha1StatementRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




### Return type

 (empty response body)

### Authorization

[api-key](../README.md#api-key)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetV1alpha1Statement

> V1alpha1Statement GetV1alpha1Statement(ctx, orgId, envId, id).Execute()

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
    orgId := "orgId_example" // string | The Org
    envId := "envId_example" // string | The Env
    id := "id_example" // string | The unique identifier for the statement.

    configuration := openapiclient.NewConfiguration()
    api_client := openapiclient.NewAPIClient(configuration)
    resp, r, err := api_client.StatementsV1alpha1Api.GetV1alpha1Statement(context.Background(), orgId, envId, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `StatementsV1alpha1Api.GetV1alpha1Statement``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetV1alpha1Statement`: V1alpha1Statement
    fmt.Fprintf(os.Stdout, "Response from `StatementsV1alpha1Api.GetV1alpha1Statement`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The Org | 
**envId** | **string** | The Env | 
**id** | **string** | The unique identifier for the statement. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetV1alpha1StatementRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




### Return type

[**V1alpha1Statement**](v1alpha1.Statement.md)

### Authorization

[api-key](../README.md#api-key)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListV1alpha1Statements

> V1alpha1StatementList ListV1alpha1Statements(ctx, orgId, envId).PageSize(pageSize).PageToken(pageToken).Execute()

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
    orgId := "orgId_example" // string | The Org
    envId := "envId_example" // string | The Env
    pageSize := int32(56) // int32 | A pagination size for collection requests. (optional) (default to 10)
    pageToken := "pageToken_example" // string | An opaque pagination token for collection requests. (optional)

    configuration := openapiclient.NewConfiguration()
    api_client := openapiclient.NewAPIClient(configuration)
    resp, r, err := api_client.StatementsV1alpha1Api.ListV1alpha1Statements(context.Background(), orgId, envId).PageSize(pageSize).PageToken(pageToken).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `StatementsV1alpha1Api.ListV1alpha1Statements``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ListV1alpha1Statements`: V1alpha1StatementList
    fmt.Fprintf(os.Stdout, "Response from `StatementsV1alpha1Api.ListV1alpha1Statements`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The Org | 
**envId** | **string** | The Env | 

### Other Parameters

Other parameters are passed through a pointer to a apiListV1alpha1StatementsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **pageSize** | **int32** | A pagination size for collection requests. | [default to 10]
 **pageToken** | **string** | An opaque pagination token for collection requests. | 

### Return type

[**V1alpha1StatementList**](v1alpha1.StatementList.md)

### Authorization

[api-key](../README.md#api-key)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

