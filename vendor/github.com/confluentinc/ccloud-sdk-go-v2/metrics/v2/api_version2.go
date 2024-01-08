// Copyright 2021 Confluent Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*
Confluent Cloud Metrics API

# Introduction  The Confluent Cloud Metrics API provides actionable operational metrics about your Confluent Cloud deployment. This is a queryable HTTP API in which the user will `POST` a query written in JSON and get back a time series of metrics specified by the query.  Comprehensive documentation is available on [docs.confluent.io](https://docs.confluent.io/current/cloud/metrics-api.html).  # Available Metrics Reference  <h3 style=\"margin-top: 0;\">Please see the <a href=\"/docs/descriptors\">Metrics Reference</a> for a list of available metrics.</h3>  This information is also available programmatically via the [descriptors endpoint](#tag/Version-2/paths/~1v2~1metrics~1{dataset}~1descriptors~1metrics/get).  # Authentication Confluent uses API keys for integrating with Confluent Cloud. Applications must be authorized and authenticated before they can access or manage resources in Confluent Cloud. You can manage your API keys in the Confluent Cloud Dashboard or Confluent Cloud CLI.  An API key is owned by a User or Service Account and inherits the permissions granted to the owner.  Today, you can divide API keys into two classes:  * **Cloud API Keys** - These grant access to the Confluent Cloud Control Plane APIs,   such as for Provisioning and Metrics integrations. * **Cluster API Keys** - These grant access to a single Confluent cluster, such as a specific   Kafka or Schema Registry cluster.  **Cloud API Keys are required for the Metrics API**. Cloud API Keys can be created using the [Confluent Cloud CLI](https://docs.confluent.io/current/cloud/cli/).  ``` ccloud api-key create --resource cloud ```  All API requests must be made over HTTPS. Calls made over plain HTTP will fail. API requests without authentication will also fail.  # Versioning  Confluent APIs ensure stability for your integrations by avoiding the introduction of breaking changes to customers unexpectedly. Confluent will make non-breaking API changes without advance notice. Thus, API clients **must** follow the [Compatibility Policy](#section/Versioning/Compatibility-Policy) below to ensure your ingtegration remains stable. All APIs follow the API Lifecycle Policy described below, which describes the guarantees API clients can rely on.  Breaking changes will be [widely communicated](#communication) in advance in accordance with our [Deprecation Policy](#section/Versioning/Deprecation-Policy). Confluent will provide timelines and a migration path for all API changes, where available. Be sure to subscribe to one or more [communication channels](#communication) so you don't miss any updates!  One exception to these guidelines is for critical security issues. We will take any necessary actions to mitigate any critical security issue as soon as possible, which may include disabling the vulnerable functionality until a proper solution is available.  Do not consume any Confluent API unless it is documented in the API Reference. All undocumented endpoints should be considered private, subject to change without notice, and not covered by any agreements.  > Note: The \"v1\" in the URL is not a \"major version\" in the [Semantic Versioning](https://semver.org/) sense. It is a \"generational version\" or \"meta version\", as seen in other APIs like <a href=\"https://developer.github.com/v3/versions/\" target=\"_blank\">Github API</a> or the <a href=\"https://stripe.com/docs/api/versioning\" target=\"_blank\">Stripe API</a>.  ## Changelog  ### 2021-09-23  #### API Version 1 is now deprecated All API Version 1 endpoints are now deprecated and will be removed on 2022-04-04. API users should migrate to API [Version 2](#tag/Version-2).  ### 2021-08-24  #### Metric-specific aggregation functions New metrics are being introduced that require alternative aggregation functions (e.g. `MAX`). When querying those metrics, using `agg: \"SUM\"` will return an error. It is recommended that clients **omit the `agg` field in the request** such that the required aggregation function for the specific metric is automatically applied on the backend.  > Note: The initial version of Metrics API required clients to effectively hardcode `agg: \"SUM\"` > in all queries.  In early 2021, the `agg` field was made optional, but many clients have not > been updated to omit the `agg` field.  #### Cursor-based pagination for `/query` endpoint The `/query` endpoint now supports cursor-based pagination similar to the `/descriptors` and `/attributes` endpoints.  ### 2021-02-10  #### API Version 2 is now Generally Available (GA) See the [Version 2](#tag/Version-2) section below for a detailed description of changes and migration guide.  ### 2020-12-04  #### API Version 2 *(Preview)* Version 2 of the Metrics API is now available in Preview. See the [Version 2](#tag/Version-2) section below for a detailed description of changes.  ### 2020-07-08  #### Correction for `active_connection_count` metric A bug in the `active_connection_count` metric that affected a subset of customers was fixed. Customers exporting the metric to an external monitoring system may observe a discontinuity between historical results and current results due to this one-time correction.  ### 2020-04-01 This release includes the following changes from the preview release:  #### New `format` request attribute The `/query` request now includes a `format` attribute which controls the result structuring in the response body.  See the `/query` endpoint definition for more details.  #### New `/available` endpoint The new `/available` endpoint allows determining which metrics are available for a set of resources (defined by labels). This endpoint can be used to determine which subset of metrics are currently available for a specific resource (e.g. a Confluent Cloud Kafka cluster).  #### Metric type changes The `CUMULATIVE_(INT|DOUBLE)` metric type enumeration was changed to `COUNTER_(INT|DOUBLE)`. This was done to better align with OpenTelemetry conventions. In tandem with this change, several metrics that were improperly classified as `GAUGE`s were re-classified as `COUNTER`s.  ### Metric name changes The `/delta` suffix has been removed from the following metrics: * `io.confluent.kafka.server/received_bytes/delta` * `io.confluent.kafka.server/sent_bytes/delta` * `io.confluent.kafka.server/request_count/delta`  ### 2020-09-15  #### Retire `/available` endpoint The `/available` endpoint (which was in _Preview_ status) has been removed from the API. The `/descriptors` endpoint can still be used to determine the universe of available metrics for Metrics API.  **The legacy metric names are deprecated and will stop functioning on 2020-07-01.**  ## API Lifecycle Policy  The following status labels are applicable to APIs, features, and SDK versions, based on the current support status of each:  * **Early Access** – May change at any time. Not recommended for production usage. Not   officially supported by Confluent. Intended for user feedback only. Users must be granted   explicit access to the API by Confluent. * **Preview** – Unlikely to change between Preview and General Availability. Not recommended   for production usage. Officially supported by Confluent for non-production usage.   For Closed Previews, users must be granted explicit access to the API by Confluent. * **Generally Available (GA)** – Will not change at short notice. Recommended for production   usage. Officially supported by Confluent for non-production and production usage. * **Deprecated** – No longer supported. Will be removed in the future at the announced date.   Use is discouraged and migration following the upgrade guide is recommended. * **Sunset** – Removed, and no longer supported or available.  Resources, operations, and individual fields in the <a href=\"./api.yaml\" target=\"_blank\">OpenAPI spec</a> will be annotated with `x-lifecycle-stage`, `x-deprecated-at`, and `x-sunset-at`. These annotations will appear in the corresponding API Reference Documentation. An API is \"Generally Available\" unless explicitly marked otherwise.  ## Compatibility Policy  Confluent APIs are governed by <a href=\"https://docs.confluent.io/current/cloud/limits.html#upgrade-policy\" target=\"_blank\"> Confluent Cloud Upgrade Policy</a> in which we will make backward incompatible changes and deprecations approximately once per year, and will provide 180 days notice via email to all registered Confluent Cloud users.  ### Backward Compatibility  > *An API version is backwards-compatible if a program written against the previous version of > the API will continue to work the same way, without modification, against this version of the > API.*  Confluent considers the following changes to be backwards-compatible:  * Adding new API resources. * Adding new optional parameters to existing API requests (e.g., query string or body). * Adding new properties to existing API responses. * Changing the order of properties in existing API responses. * Changing the length or format of object IDs or other opaque strings.   * Unless otherwise documented, you can safely assume object IDs we generate will never exceed     255 characters, but you should be able to handle IDs of up to that length.     If you're using MySQL, for example, you should store IDs in a     `VARCHAR(255) COLLATE utf8_bin` column.   * This includes adding or removing fixed prefixes (such as `lkc-` on kafka cluster IDs).   * This includes API keys, API tokens, and similar authentication mechanisms.   * This includes all strings described as \"opaque\" in the docs, such as pagination cursors. * Omitting properties with null values from existing API responses.  ### Client Responsibilities  * Resource and rate limits, and the default and maximum sizes of paginated data **are not**   considered part of the API contract and may change (possibly dynamically). It is the client's   responsibility to read the road signs and obey the speed limit. * If a property has a primitive type and the API documentation does not explicitly limit its   possible values, clients **must not** assume the values are constrained to a particular set   of possible responses. * If a property of an object is not explicitly declared as mandatory in the API, clients   **must not** assume it will be present. * A resource **may** be modified to return a \"redirection\" response (e.g. `301`, `307`) instead   of directly returning the resource. Clients **must** handle HTTP-level redirects, and respect   HTTP headers (e.g. `Location`).  ## Deprecation Policy  Confluent will announce deprecations at least 180 days in advance of a breaking change and we will continue to maintain the deprecated APIs in their original form during this time.  Exceptions to this policy apply in case of critical security vulnerabilities or functional defects.  ### Communication  When a deprecation is announced, the details and any relevant migration information will be available on the following channels:  * Publication in the [API Changelog](#section/Versioning/Changelog) * Lifecycle, deprecation and \"x-deprecated-at\" annotations in the   <a href=\"/docs/api.yaml\" target=\"_blank\">OpenAPI spec</a> * Announcements on the   <a href=\"https://www.confluent.io/blog/\" target=\"_blank\">Developer Blog</a>,   <a href=\"https://confluentcommunity.slack.com\" target=\"_blank\">Community Slack</a>   (<a href=\"https://slackpass.io/confluentcommunity\" target=\"_blank\">join!</a>),   <a href=\"https://groups.google.com/forum/#!forum/confluent-platform\" target=\"_blank\">   Google Group</a>,   the <a href=\"https://twitter.com/ConfluentInc\" target=\"_blank\">@ConfluentInc twitter</a>   account, and similar channels * Enterprise customers may receive information by email to their specified Confluent contact,   if applicable.  # Object Model The object model for the Metrics API is designed similarly to the [OpenTelemetry](https://opentelemetry.io/) standard.  ## Metrics A _metric_ is a numeric attribute of a resource, measured at a specific point in time, labeled with contextual metadata gathered at the point of instrumentation.  There are two types of metrics: * `GAUGE`: An instantaneous measurement of a value.   Gauge metrics are implicitly averaged when aggregating over time.   > Example: `io.confluent.kafka.server/retained_bytes` * `COUNTER`: The count of occurrences in a _single (one minute) sampling   interval_ (unless otherwise stated in the metric description).   Counter metrics are implicitly summed when aggregating over time.   > Example: `io.confluent.kafka.server/received_bytes`  The list of metrics and their labels is available at [/docs/descriptors](/docs/descriptors).  ## Resources A _resource_ represents an entity against which metrics are collected.  For example, a Kafka cluster, a Kafka Connector, a ksqlDB application, etc.  Each metric _descriptor_ is associated with one or more resource _descriptors_, representing the resource types to which that metric can apply.  A metric _data point_ is associated with a single resource _instance_, identified by the resource labels on that metric data point.  For example, metrics emitted by Kafka Connect are associated to the `connector` resource type. Data points for those metrics include resource labels identifying the specific `connector` instance that emitted the metric.  The list of resource types and labels are discoverable via the `/descriptors/resources` endpoint.  ## Labels A _label_ is a key-value attribute associated with a metric data point.  Labels can be used in queries to filter or group the results.  Labels must be prefixed when used in queries: * `metric.<label>` (for metric labels), for example `metric.topic` * `resource.<resource-type>.<label>` (for resource labels), for example `resource.kafka.id`.  The set of valid label keys for a metric include: * The label keys defined on that metric's descriptor itself * The label keys defined on the resource descriptor for the metric's associated resource type  For example, the `io.confluent.kafka.server/received_bytes` metric has the following labels: * `resource.kafka.id` - The Kafka cluster to which the metric pertains * `metric.topic` - The Kafka topic to which the bytes were produced * `metric.partition` - The partition to which the bytes were produced  ## Datasets A _dataset_ is a logical collection of metrics that can be queried together.  The `dataset` is a required URL template parameter for every endpoint in this API.  The following datasets are currently available:  <table> <thead>   <tr>     <th style=\"width: 250px;\">Dataset</th>     <th>Description</th>   </tr> </thead> <tbody>   <tr>     <td>       <code>cloud</code>       <p><img           src=\"https://img.shields.io/badge/Lifecycle%20Stage-Generally%20Available-%230074A2\"           alt=\"generally-available\">     </td>     <td>       Metrics originating from Confluent Cloud resources.       <p>Requests to this dataset require a resource <code>filter</code>          (e.g. Kafka cluster ID, Connector ID, etc.) in the query for authorization purposes.          The client's API key must be authorized for the resource referenced in the filter.     </td>   </tr> </tbody> </table>  # Client Considerations and Best Practices  ## Rate Limiting To protect the stability of the API and keep it available to all users, Confluent employs multiple safeguards. Users who send many requests in quick succession or perform too many concurrent operations may be throttled or have their requested rejected with an error. When a rate limit is breached, an HTTP `429 Too Many Requests` error is returned.  Rate limits are enforced at multiple scopes.  ### Global Rate Limits A global rate limit of **60 requests per IP address, per minute** is enforced.  ### Per-endpoint Rate Limits Additionally, some endpoint-specific rate limits are enforced.  | Endpoint  | Rate limit | | --------- | ---------- | | `/v2/metrics/{dataset}/export` | 80 requests per resource, per hour, per principal.<br/>See the [export endpoint documentation](#tag/Version-2/paths/~1v2~1metrics~1{dataset}~1export/get) for details. |  ## Retries Implement retry logic in your client to gracefully handle transient API failures. This should be done by watching for error responses and building in a retry mechanism. This mechanism should follow a capped exponential backoff policy to prevent retry amplification (\"retry storms\") and also introduce some randomness (\"jitter\") to avoid the [thundering herd effect](https://en.wikipedia.org/wiki/Thundering_herd_problem).  ## Metric Data Latency Metric data points are typically available for query in the API within **5 minutes** of their origination at the source.  This latency can vary based on network conditions and processing overhead.  Clients that are polling (or \"scraping\") metrics into an external monitoring system should account for this latency in their polling requests.  API requests that fail to incorporate the latency into the query `interval` may have incomplete data in the response.  ## Pagination Cursors, tokens, and corresponding pagination links may expire after a short amount of time. In this case, the API will return a `400 Bad Request` error and the client will need to restart from the beginning.  The client should have no trouble pausing between rate limiting windows, but persisting cursors for hours or days is not recommended. 

API version: 
Contact: support@confluent.io
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package v2

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"reflect"
)


type Version2Api interface {

	/*
	V2MetricsDatasetAttributesPost Query label values

	Enumerates label values for a single metric.


	@param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
	@param dataset The dataset to query.
	@return ApiV2MetricsDatasetAttributesPostRequest
	*/
	V2MetricsDatasetAttributesPost(ctx context.Context, dataset string) ApiV2MetricsDatasetAttributesPostRequest

	// V2MetricsDatasetAttributesPostExecute executes the request
	//  @return AttributesResponse
	V2MetricsDatasetAttributesPostExecute(r ApiV2MetricsDatasetAttributesPostRequest) (*AttributesResponse, *http.Response, error)

	/*
	V2MetricsDatasetDescriptorsMetricsGet List metric descriptors

	Lists all the metric descriptors for a dataset.

A metric descriptor represents metadata for a metric, including its data type and labels.
This metadata is provided programmatically to enable clients to dynamically adjust as new
metrics are added to the dataset, rather than hardcoding metric names in client code.


	@param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
	@param dataset The dataset to list metric descriptors for. Currently the only supported dataset name is `cloud`. See [here](#section/Object-Model/Datasets).
	@return ApiV2MetricsDatasetDescriptorsMetricsGetRequest
	*/
	V2MetricsDatasetDescriptorsMetricsGet(ctx context.Context, dataset string) ApiV2MetricsDatasetDescriptorsMetricsGetRequest

	// V2MetricsDatasetDescriptorsMetricsGetExecute executes the request
	//  @return ListMetricDescriptorsResponse
	V2MetricsDatasetDescriptorsMetricsGetExecute(r ApiV2MetricsDatasetDescriptorsMetricsGetRequest) (*ListMetricDescriptorsResponse, *http.Response, error)

	/*
	V2MetricsDatasetDescriptorsResourcesGet List resource descriptors

	Lists all the resource descriptors for a dataset.


	@param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
	@param dataset The dataset to list resource descriptors for. Currently the only supported dataset name is `cloud`. See [here](#section/Object-Model/Datasets).
	@return ApiV2MetricsDatasetDescriptorsResourcesGetRequest
	*/
	V2MetricsDatasetDescriptorsResourcesGet(ctx context.Context, dataset string) ApiV2MetricsDatasetDescriptorsResourcesGetRequest

	// V2MetricsDatasetDescriptorsResourcesGetExecute executes the request
	//  @return ListResourceDescriptorsResponse
	V2MetricsDatasetDescriptorsResourcesGetExecute(r ApiV2MetricsDatasetDescriptorsResourcesGetRequest) (*ListResourceDescriptorsResponse, *http.Response, error)

	/*
	V2MetricsDatasetExportGet Export metric values

	Export current metric values in [OpenMetrics format](https://github.com/OpenObservability/OpenMetrics/blob/main/specification/OpenMetrics.md)
or [Prometheus format](https://prometheus.io/docs/instrumenting/exposition_formats/#text-based-format),
suitable for import into an external monitoring system. Returns the single most recent
data point for each metric, for each distinct combination of labels.

#### Supported datasets and metrics
Only the `cloud` dataset is supported for this endpoint.

Only a subset of metrics and labels from the dataset are included in the export response. To request
a particular metric or label be added, please contact [Confluent Support](https://support.confluent.io).

#### Metric translation
Metric and label names are translated to adhere to [Prometheus restrictions](https://prometheus.io/docs/concepts/data_model/#metric-names-and-labels).
The `resource.` and `metric.` prefixes from label names are also dropped to simplify consumption in downstream systems.

Counter metrics are classified as the Prometheus `gauge` type to conform to required semantics.
> The `counter` type in Prometheus must be monotonically increasing, whereas Confluent
Metrics API counters are represented as deltas.

#### Timestamp offset
To account for [metric data latency](#section/Client-Considerations-and-Best-Practices/Metric-Data-Latency),
this endpoint returns metrics from the current timestamp minus a fixed offset. The current
offset is 5 minutes rounded down to the start of the minute. For example, if a request is
received at `12:06:41`, the returned metrics will have the timestamp `12:01:00` and represent the
data for the interval `12:01:00` through `12:02:00` (exclusive).

> **NOTE:** Confluent may choose to lengthen or shorten this offset based on operational
considerations. _Doing so is considered a backwards-compatible change_.

To accommodate this offset, the timestamps in the response should be honored when importing
the metrics. For example, in prometheus this can be controlled using the `honor_timestamps`
flag.

#### Rate limits
Since metrics are available at minute granularity, it is expected that clients scrape this
endpoint at most once per minute. To allow for ad-hoc testing, the rate limit is enforced
at hourly granularity. To accommodate retries, the rate limit is 80 requests per hour
rather than 60 per hour.

The rate limit is evaluated on a per-resource basis. For example, the following requests would
each be allowed an 80-requests-per-hour rate:
* `GET /v2/metrics/cloud/export?resource.kafka.id=lkc-1&resource.kafka.id=lkc-2`
* `GET /v2/metrics/cloud/export?resource.kafka.id=lkc-3`

Rate limits for this endpoint are also scoped to the authentication principal. This allows multiple systems
to export metrics for the same resources by configuring each with a separate service account.

If the rate limit is exceeded, the response body will include a message indicating which
resource exceeded the limit.
```json
{
  "errors": [
    {
      "status": "429",
      "detail": "Too many requests have been made for the following resources:
kafka.id:lkc-12345. Please see the documentation for current rate limits."
    }
  ]
}
```

#### Example Prometheus scrape configuration
Here is an example [prometheus configuration](https://prometheus.io/docs/prometheus/latest/configuration/configuration/)
for scraping this endpoint:

```yaml
scrape_configs:
  - job_name: Confluent Cloud
    scrape_interval: 1m
    scrape_timeout: 1m
    honor_timestamps: true
    static_configs:
      - targets:
        - api.telemetry.confluent.cloud
    scheme: https
    basic_auth:
      username: <Cloud API Key>
      password: <Cloud API Secret>
    metrics_path: /v2/metrics/cloud/export
    params:
      "resource.kafka.id":
        - lkc-1
        - lkc-2
```


	@param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
	@param dataset The dataset to export metrics for. Currently the only supported dataset name is `cloud`. See [here](#section/Object-Model/Datasets).
	@return ApiV2MetricsDatasetExportGetRequest
	*/
	V2MetricsDatasetExportGet(ctx context.Context, dataset string) ApiV2MetricsDatasetExportGetRequest

	// V2MetricsDatasetExportGetExecute executes the request
	//  @return string
	V2MetricsDatasetExportGetExecute(r ApiV2MetricsDatasetExportGetRequest) (string, *http.Response, error)

	/*
	V2MetricsDatasetQueryPost Query metric values

	Query for metric values in a dataset.


	@param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
	@param dataset The dataset to query. Currently the only supported dataset name is `cloud`. See [here](#section/Object-Model/Datasets).
	@return ApiV2MetricsDatasetQueryPostRequest
	*/
	V2MetricsDatasetQueryPost(ctx context.Context, dataset string) ApiV2MetricsDatasetQueryPostRequest

	// V2MetricsDatasetQueryPostExecute executes the request
	//  @return QueryResponse
	V2MetricsDatasetQueryPostExecute(r ApiV2MetricsDatasetQueryPostRequest) (*QueryResponse, *http.Response, error)
}

// Version2ApiService Version2Api service
type Version2ApiService service

type ApiV2MetricsDatasetAttributesPostRequest struct {
	ctx context.Context
	ApiService Version2Api
	dataset string
	pageToken *string
	attributesRequest *AttributesRequest
}

// The next page token. The token is returned by the previous request as part of &#x60;meta.pagination&#x60;.
func (r ApiV2MetricsDatasetAttributesPostRequest) PageToken(pageToken string) ApiV2MetricsDatasetAttributesPostRequest {
	r.pageToken = &pageToken
	return r
}

func (r ApiV2MetricsDatasetAttributesPostRequest) AttributesRequest(attributesRequest AttributesRequest) ApiV2MetricsDatasetAttributesPostRequest {
	r.attributesRequest = &attributesRequest
	return r
}

func (r ApiV2MetricsDatasetAttributesPostRequest) Execute() (*AttributesResponse, *http.Response, error) {
	return r.ApiService.V2MetricsDatasetAttributesPostExecute(r)
}

/*
V2MetricsDatasetAttributesPost Query label values

Enumerates label values for a single metric.


 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param dataset The dataset to query.
 @return ApiV2MetricsDatasetAttributesPostRequest
*/
func (a *Version2ApiService) V2MetricsDatasetAttributesPost(ctx context.Context, dataset string) ApiV2MetricsDatasetAttributesPostRequest {
	return ApiV2MetricsDatasetAttributesPostRequest{
		ApiService: a,
		ctx: ctx,
		dataset: dataset,
	}
}

// Execute executes the request
//  @return AttributesResponse
func (a *Version2ApiService) V2MetricsDatasetAttributesPostExecute(r ApiV2MetricsDatasetAttributesPostRequest) (*AttributesResponse, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodPost
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *AttributesResponse
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "Version2ApiService.V2MetricsDatasetAttributesPost")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/v2/metrics/{dataset}/attributes"
	localVarPath = strings.Replace(localVarPath, "{"+"dataset"+"}", url.PathEscape(parameterToString(r.dataset, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	if r.pageToken != nil {
		localVarQueryParams.Add("page_token", parameterToString(*r.pageToken, ""))
	}
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	// body params
	localVarPostBody = r.attributesRequest
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
			var v ErrorResponse
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiV2MetricsDatasetDescriptorsMetricsGetRequest struct {
	ctx context.Context
	ApiService Version2Api
	dataset string
	pageSize *int32
	pageToken *string
	resourceType *string
}

// The maximum number of results to return. The page size is an integer in the range from 1 through 1000.
func (r ApiV2MetricsDatasetDescriptorsMetricsGetRequest) PageSize(pageSize int32) ApiV2MetricsDatasetDescriptorsMetricsGetRequest {
	r.pageSize = &pageSize
	return r
}

// The next page token. The token is returned by the previous request as part of &#x60;meta.pagination&#x60;.
func (r ApiV2MetricsDatasetDescriptorsMetricsGetRequest) PageToken(pageToken string) ApiV2MetricsDatasetDescriptorsMetricsGetRequest {
	r.pageToken = &pageToken
	return r
}

// The type of the resource to list metric descriptors for.
func (r ApiV2MetricsDatasetDescriptorsMetricsGetRequest) ResourceType(resourceType string) ApiV2MetricsDatasetDescriptorsMetricsGetRequest {
	r.resourceType = &resourceType
	return r
}

func (r ApiV2MetricsDatasetDescriptorsMetricsGetRequest) Execute() (*ListMetricDescriptorsResponse, *http.Response, error) {
	return r.ApiService.V2MetricsDatasetDescriptorsMetricsGetExecute(r)
}

/*
V2MetricsDatasetDescriptorsMetricsGet List metric descriptors

Lists all the metric descriptors for a dataset.

A metric descriptor represents metadata for a metric, including its data type and labels.
This metadata is provided programmatically to enable clients to dynamically adjust as new
metrics are added to the dataset, rather than hardcoding metric names in client code.


 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param dataset The dataset to list metric descriptors for. Currently the only supported dataset name is `cloud`. See [here](#section/Object-Model/Datasets).
 @return ApiV2MetricsDatasetDescriptorsMetricsGetRequest
*/
func (a *Version2ApiService) V2MetricsDatasetDescriptorsMetricsGet(ctx context.Context, dataset string) ApiV2MetricsDatasetDescriptorsMetricsGetRequest {
	return ApiV2MetricsDatasetDescriptorsMetricsGetRequest{
		ApiService: a,
		ctx: ctx,
		dataset: dataset,
	}
}

// Execute executes the request
//  @return ListMetricDescriptorsResponse
func (a *Version2ApiService) V2MetricsDatasetDescriptorsMetricsGetExecute(r ApiV2MetricsDatasetDescriptorsMetricsGetRequest) (*ListMetricDescriptorsResponse, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodGet
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *ListMetricDescriptorsResponse
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "Version2ApiService.V2MetricsDatasetDescriptorsMetricsGet")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/v2/metrics/{dataset}/descriptors/metrics"
	localVarPath = strings.Replace(localVarPath, "{"+"dataset"+"}", url.PathEscape(parameterToString(r.dataset, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	if r.pageSize != nil {
		localVarQueryParams.Add("page_size", parameterToString(*r.pageSize, ""))
	}
	if r.pageToken != nil {
		localVarQueryParams.Add("page_token", parameterToString(*r.pageToken, ""))
	}
	if r.resourceType != nil {
		localVarQueryParams.Add("resource_type", parameterToString(*r.resourceType, ""))
	}
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
			var v ErrorResponse
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiV2MetricsDatasetDescriptorsResourcesGetRequest struct {
	ctx context.Context
	ApiService Version2Api
	dataset string
	pageSize *int32
	pageToken *string
}

// The maximum number of results to return. The page size is an integer in the range from 1 through 1000.
func (r ApiV2MetricsDatasetDescriptorsResourcesGetRequest) PageSize(pageSize int32) ApiV2MetricsDatasetDescriptorsResourcesGetRequest {
	r.pageSize = &pageSize
	return r
}

// The next page token. The token is returned by the previous request as part of &#x60;meta.pagination&#x60;.
func (r ApiV2MetricsDatasetDescriptorsResourcesGetRequest) PageToken(pageToken string) ApiV2MetricsDatasetDescriptorsResourcesGetRequest {
	r.pageToken = &pageToken
	return r
}

func (r ApiV2MetricsDatasetDescriptorsResourcesGetRequest) Execute() (*ListResourceDescriptorsResponse, *http.Response, error) {
	return r.ApiService.V2MetricsDatasetDescriptorsResourcesGetExecute(r)
}

/*
V2MetricsDatasetDescriptorsResourcesGet List resource descriptors

Lists all the resource descriptors for a dataset.


 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param dataset The dataset to list resource descriptors for. Currently the only supported dataset name is `cloud`. See [here](#section/Object-Model/Datasets).
 @return ApiV2MetricsDatasetDescriptorsResourcesGetRequest
*/
func (a *Version2ApiService) V2MetricsDatasetDescriptorsResourcesGet(ctx context.Context, dataset string) ApiV2MetricsDatasetDescriptorsResourcesGetRequest {
	return ApiV2MetricsDatasetDescriptorsResourcesGetRequest{
		ApiService: a,
		ctx: ctx,
		dataset: dataset,
	}
}

// Execute executes the request
//  @return ListResourceDescriptorsResponse
func (a *Version2ApiService) V2MetricsDatasetDescriptorsResourcesGetExecute(r ApiV2MetricsDatasetDescriptorsResourcesGetRequest) (*ListResourceDescriptorsResponse, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodGet
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *ListResourceDescriptorsResponse
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "Version2ApiService.V2MetricsDatasetDescriptorsResourcesGet")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/v2/metrics/{dataset}/descriptors/resources"
	localVarPath = strings.Replace(localVarPath, "{"+"dataset"+"}", url.PathEscape(parameterToString(r.dataset, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	if r.pageSize != nil {
		localVarQueryParams.Add("page_size", parameterToString(*r.pageSize, ""))
	}
	if r.pageToken != nil {
		localVarQueryParams.Add("page_token", parameterToString(*r.pageToken, ""))
	}
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
			var v ErrorResponse
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiV2MetricsDatasetExportGetRequest struct {
	ctx context.Context
	ApiService Version2Api
	dataset string
	resourceKafkaId *[]string
	resourceConnectorId *[]string
	resourceKsqlId *[]string
	resourceSchemaRegistryId *[]string
	metric *[]string
}

// The ID of the Kafka cluster to export metrics for. This parameter can be specified multiple times (e.g. &#x60;?resource.kafka.id&#x3D;lkc-1&amp;resource.kafka.id&#x3D;lkc-2&#x60;).
func (r ApiV2MetricsDatasetExportGetRequest) ResourceKafkaId(resourceKafkaId []string) ApiV2MetricsDatasetExportGetRequest {
	r.resourceKafkaId = &resourceKafkaId
	return r
}

// The ID of the Connector to export metrics for. This parameter can be specified multiple times.
func (r ApiV2MetricsDatasetExportGetRequest) ResourceConnectorId(resourceConnectorId []string) ApiV2MetricsDatasetExportGetRequest {
	r.resourceConnectorId = &resourceConnectorId
	return r
}

// The ID of the ksqlDB application to export metrics for. This parameter can be specified multiple times.
func (r ApiV2MetricsDatasetExportGetRequest) ResourceKsqlId(resourceKsqlId []string) ApiV2MetricsDatasetExportGetRequest {
	r.resourceKsqlId = &resourceKsqlId
	return r
}

// The ID of the Schema Registry to export metrics for. This parameter can be specified multiple times.
func (r ApiV2MetricsDatasetExportGetRequest) ResourceSchemaRegistryId(resourceSchemaRegistryId []string) ApiV2MetricsDatasetExportGetRequest {
	r.resourceSchemaRegistryId = &resourceSchemaRegistryId
	return r
}

// The metric to export. If this parameter is not specified, all metrics for the resource will be exported. This parameter can be specified multiple times.
func (r ApiV2MetricsDatasetExportGetRequest) Metric(metric []string) ApiV2MetricsDatasetExportGetRequest {
	r.metric = &metric
	return r
}

func (r ApiV2MetricsDatasetExportGetRequest) Execute() (string, *http.Response, error) {
	return r.ApiService.V2MetricsDatasetExportGetExecute(r)
}

/*
V2MetricsDatasetExportGet Export metric values

Export current metric values in [OpenMetrics format](https://github.com/OpenObservability/OpenMetrics/blob/main/specification/OpenMetrics.md)
or [Prometheus format](https://prometheus.io/docs/instrumenting/exposition_formats/#text-based-format),
suitable for import into an external monitoring system. Returns the single most recent
data point for each metric, for each distinct combination of labels.

#### Supported datasets and metrics
Only the `cloud` dataset is supported for this endpoint.

Only a subset of metrics and labels from the dataset are included in the export response. To request
a particular metric or label be added, please contact [Confluent Support](https://support.confluent.io).

#### Metric translation
Metric and label names are translated to adhere to [Prometheus restrictions](https://prometheus.io/docs/concepts/data_model/#metric-names-and-labels).
The `resource.` and `metric.` prefixes from label names are also dropped to simplify consumption in downstream systems.

Counter metrics are classified as the Prometheus `gauge` type to conform to required semantics.
> The `counter` type in Prometheus must be monotonically increasing, whereas Confluent
Metrics API counters are represented as deltas.

#### Timestamp offset
To account for [metric data latency](#section/Client-Considerations-and-Best-Practices/Metric-Data-Latency),
this endpoint returns metrics from the current timestamp minus a fixed offset. The current
offset is 5 minutes rounded down to the start of the minute. For example, if a request is
received at `12:06:41`, the returned metrics will have the timestamp `12:01:00` and represent the
data for the interval `12:01:00` through `12:02:00` (exclusive).

> **NOTE:** Confluent may choose to lengthen or shorten this offset based on operational
considerations. _Doing so is considered a backwards-compatible change_.

To accommodate this offset, the timestamps in the response should be honored when importing
the metrics. For example, in prometheus this can be controlled using the `honor_timestamps`
flag.

#### Rate limits
Since metrics are available at minute granularity, it is expected that clients scrape this
endpoint at most once per minute. To allow for ad-hoc testing, the rate limit is enforced
at hourly granularity. To accommodate retries, the rate limit is 80 requests per hour
rather than 60 per hour.

The rate limit is evaluated on a per-resource basis. For example, the following requests would
each be allowed an 80-requests-per-hour rate:
* `GET /v2/metrics/cloud/export?resource.kafka.id=lkc-1&resource.kafka.id=lkc-2`
* `GET /v2/metrics/cloud/export?resource.kafka.id=lkc-3`

Rate limits for this endpoint are also scoped to the authentication principal. This allows multiple systems
to export metrics for the same resources by configuring each with a separate service account.

If the rate limit is exceeded, the response body will include a message indicating which
resource exceeded the limit.
```json
{
  "errors": [
    {
      "status": "429",
      "detail": "Too many requests have been made for the following resources:
kafka.id:lkc-12345. Please see the documentation for current rate limits."
    }
  ]
}
```

#### Example Prometheus scrape configuration
Here is an example [prometheus configuration](https://prometheus.io/docs/prometheus/latest/configuration/configuration/)
for scraping this endpoint:

```yaml
scrape_configs:
  - job_name: Confluent Cloud
    scrape_interval: 1m
    scrape_timeout: 1m
    honor_timestamps: true
    static_configs:
      - targets:
        - api.telemetry.confluent.cloud
    scheme: https
    basic_auth:
      username: <Cloud API Key>
      password: <Cloud API Secret>
    metrics_path: /v2/metrics/cloud/export
    params:
      "resource.kafka.id":
        - lkc-1
        - lkc-2
```


 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param dataset The dataset to export metrics for. Currently the only supported dataset name is `cloud`. See [here](#section/Object-Model/Datasets).
 @return ApiV2MetricsDatasetExportGetRequest
*/
func (a *Version2ApiService) V2MetricsDatasetExportGet(ctx context.Context, dataset string) ApiV2MetricsDatasetExportGetRequest {
	return ApiV2MetricsDatasetExportGetRequest{
		ApiService: a,
		ctx: ctx,
		dataset: dataset,
	}
}

// Execute executes the request
//  @return string
func (a *Version2ApiService) V2MetricsDatasetExportGetExecute(r ApiV2MetricsDatasetExportGetRequest) (string, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodGet
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  string
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "Version2ApiService.V2MetricsDatasetExportGet")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/v2/metrics/{dataset}/export"
	localVarPath = strings.Replace(localVarPath, "{"+"dataset"+"}", url.PathEscape(parameterToString(r.dataset, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	if r.resourceKafkaId != nil {
		t := *r.resourceKafkaId
		if reflect.TypeOf(t).Kind() == reflect.Slice {
			s := reflect.ValueOf(t)
			for i := 0; i < s.Len(); i++ {
				localVarQueryParams.Add("resource.kafka.id", parameterToString(s.Index(i), "multi"))
			}
		} else {
			localVarQueryParams.Add("resource.kafka.id", parameterToString(t, "multi"))
		}
	}
	if r.resourceConnectorId != nil {
		t := *r.resourceConnectorId
		if reflect.TypeOf(t).Kind() == reflect.Slice {
			s := reflect.ValueOf(t)
			for i := 0; i < s.Len(); i++ {
				localVarQueryParams.Add("resource.connector.id", parameterToString(s.Index(i), "multi"))
			}
		} else {
			localVarQueryParams.Add("resource.connector.id", parameterToString(t, "multi"))
		}
	}
	if r.resourceKsqlId != nil {
		t := *r.resourceKsqlId
		if reflect.TypeOf(t).Kind() == reflect.Slice {
			s := reflect.ValueOf(t)
			for i := 0; i < s.Len(); i++ {
				localVarQueryParams.Add("resource.ksql.id", parameterToString(s.Index(i), "multi"))
			}
		} else {
			localVarQueryParams.Add("resource.ksql.id", parameterToString(t, "multi"))
		}
	}
	if r.resourceSchemaRegistryId != nil {
		t := *r.resourceSchemaRegistryId
		if reflect.TypeOf(t).Kind() == reflect.Slice {
			s := reflect.ValueOf(t)
			for i := 0; i < s.Len(); i++ {
				localVarQueryParams.Add("resource.schema_registry.id", parameterToString(s.Index(i), "multi"))
			}
		} else {
			localVarQueryParams.Add("resource.schema_registry.id", parameterToString(t, "multi"))
		}
	}
	if r.metric != nil {
		t := *r.metric
		if reflect.TypeOf(t).Kind() == reflect.Slice {
			s := reflect.ValueOf(t)
			for i := 0; i < s.Len(); i++ {
				localVarQueryParams.Add("metric", parameterToString(s.Index(i), "multi"))
			}
		} else {
			localVarQueryParams.Add("metric", parameterToString(t, "multi"))
		}
	}
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"text/plain;version=0.0.4", "application/openmetrics-text;version=1.0.0;charset=utf-8", "application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
			var v ErrorResponse
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiV2MetricsDatasetQueryPostRequest struct {
	ctx context.Context
	ApiService Version2Api
	dataset string
	pageToken *string
	queryRequest *QueryRequest
}

// The next page token. The token is returned by the previous request as part of &#x60;meta.pagination&#x60;. Pagination is only supported for requests containing a &#x60;group_by&#x60; element.
func (r ApiV2MetricsDatasetQueryPostRequest) PageToken(pageToken string) ApiV2MetricsDatasetQueryPostRequest {
	r.pageToken = &pageToken
	return r
}

func (r ApiV2MetricsDatasetQueryPostRequest) QueryRequest(queryRequest QueryRequest) ApiV2MetricsDatasetQueryPostRequest {
	r.queryRequest = &queryRequest
	return r
}

func (r ApiV2MetricsDatasetQueryPostRequest) Execute() (*QueryResponse, *http.Response, error) {
	return r.ApiService.V2MetricsDatasetQueryPostExecute(r)
}

/*
V2MetricsDatasetQueryPost Query metric values

Query for metric values in a dataset.


 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param dataset The dataset to query. Currently the only supported dataset name is `cloud`. See [here](#section/Object-Model/Datasets).
 @return ApiV2MetricsDatasetQueryPostRequest
*/
func (a *Version2ApiService) V2MetricsDatasetQueryPost(ctx context.Context, dataset string) ApiV2MetricsDatasetQueryPostRequest {
	return ApiV2MetricsDatasetQueryPostRequest{
		ApiService: a,
		ctx: ctx,
		dataset: dataset,
	}
}

// Execute executes the request
//  @return QueryResponse
func (a *Version2ApiService) V2MetricsDatasetQueryPostExecute(r ApiV2MetricsDatasetQueryPostRequest) (*QueryResponse, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodPost
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *QueryResponse
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "Version2ApiService.V2MetricsDatasetQueryPost")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/v2/metrics/{dataset}/query"
	localVarPath = strings.Replace(localVarPath, "{"+"dataset"+"}", url.PathEscape(parameterToString(r.dataset, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	if r.pageToken != nil {
		localVarQueryParams.Add("page_token", parameterToString(*r.pageToken, ""))
	}
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	// body params
	localVarPostBody = r.queryRequest
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
			var v ErrorResponse
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}
