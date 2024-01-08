# Go API client for v2

# Introduction

The Confluent Cloud Metrics API provides actionable operational metrics about your Confluent
Cloud deployment. This is a queryable HTTP API in which the user will `POST` a query written in
JSON and get back a time series of metrics specified by the query.

Comprehensive documentation is available on
[docs.confluent.io](https://docs.confluent.io/current/cloud/metrics-api.html).

# Available Metrics Reference

<h3 style=\"margin-top: 0;\">Please see the <a href=\"/docs/descriptors\">Metrics Reference</a> for
a list of available metrics.</h3>

This information is also available programmatically via the
[descriptors endpoint](#tag/Version-2/paths/~1v2~1metrics~1{dataset}~1descriptors~1metrics/get).

# Authentication
Confluent uses API keys for integrating with Confluent Cloud. Applications must be
authorized and authenticated before they can access or manage resources in Confluent Cloud.
You can manage your API keys in the Confluent Cloud Dashboard or Confluent Cloud CLI.

An API key is owned by a User or Service Account and inherits the permissions granted
to the owner.

Today, you can divide API keys into two classes:

* **Cloud API Keys** - These grant access to the Confluent Cloud Control Plane APIs,
  such as for Provisioning and Metrics integrations.
* **Cluster API Keys** - These grant access to a single Confluent cluster, such as a specific
  Kafka or Schema Registry cluster.

**Cloud API Keys are required for the Metrics API**. Cloud API Keys can be created using the
[Confluent Cloud CLI](https://docs.confluent.io/current/cloud/cli/).

```
ccloud api-key create --resource cloud
```

All API requests must be made over HTTPS. Calls made over plain HTTP will fail. API requests
without authentication will also fail.

# Versioning

Confluent APIs ensure stability for your integrations by avoiding the introduction
of breaking changes to customers unexpectedly. Confluent will make non-breaking
API changes without advance notice. Thus, API clients **must** follow the
[Compatibility Policy](#section/Versioning/Compatibility-Policy) below to ensure your
ingtegration remains stable. All APIs follow the API Lifecycle Policy described below,
which describes the guarantees API clients can rely on.

Breaking changes will be [widely communicated](#communication) in advance in accordance
with our [Deprecation Policy](#section/Versioning/Deprecation-Policy). Confluent will provide
timelines and a migration path for all API changes, where available. Be sure to subscribe
to one or more [communication channels](#communication) so you don't miss any updates!

One exception to these guidelines is for critical security issues. We will take any necessary
actions to mitigate any critical security issue as soon as possible, which may include disabling
the vulnerable functionality until a proper solution is available.

Do not consume any Confluent API unless it is documented in the API Reference. All undocumented
endpoints should be considered private, subject to change without notice, and not covered by any
agreements.

> Note: The \"v1\" in the URL is not a \"major version\" in the
[Semantic Versioning](https://semver.org/) sense. It is a \"generational version\" or
\"meta version\", as seen in other APIs like
<a href=\"https://developer.github.com/v3/versions/\" target=\"_blank\">Github API</a> or the
<a href=\"https://stripe.com/docs/api/versioning\" target=\"_blank\">Stripe API</a>.

## Changelog

### 2021-09-23

#### API Version 1 is now deprecated
All API Version 1 endpoints are now deprecated and will be removed on 2022-04-04. API users
should migrate to API [Version 2](#tag/Version-2).

### 2021-08-24

#### Metric-specific aggregation functions
New metrics are being introduced that require alternative aggregation functions (e.g. `MAX`).
When querying those metrics, using `agg: \"SUM\"` will return an error.
It is recommended that clients **omit the `agg` field in the request** such that the required
aggregation function for the specific metric is automatically applied on the backend.

> Note: The initial version of Metrics API required clients to effectively hardcode `agg: \"SUM\"`
> in all queries.  In early 2021, the `agg` field was made optional, but many clients have not
> been updated to omit the `agg` field.

#### Cursor-based pagination for `/query` endpoint
The `/query` endpoint now supports cursor-based pagination similar to the `/descriptors` and
`/attributes` endpoints.

### 2021-02-10

#### API Version 2 is now Generally Available (GA)
See the [Version 2](#tag/Version-2) section below for a detailed description of changes and
migration guide.

### 2020-12-04

#### API Version 2 *(Preview)*
Version 2 of the Metrics API is now available in Preview. See the [Version 2](#tag/Version-2)
section below for a detailed description of changes.

### 2020-07-08

#### Correction for `active_connection_count` metric
A bug in the `active_connection_count` metric that affected a subset of customers was fixed.
Customers exporting the metric to an external monitoring system may observe a discontinuity
between historical results and current results due to this one-time correction.

### 2020-04-01
This release includes the following changes from the preview release:

#### New `format` request attribute
The `/query` request now includes a `format` attribute which controls the result structuring in
the response body.  See the `/query` endpoint definition for more details.

#### New `/available` endpoint
The new `/available` endpoint allows determining which metrics are available for a set of
resources (defined by labels). This endpoint can be used to determine which subset of metrics
are currently available for a specific resource (e.g. a Confluent Cloud Kafka cluster).

#### Metric type changes
The `CUMULATIVE_(INT|DOUBLE)` metric type enumeration was changed to `COUNTER_(INT|DOUBLE)`.
This was done to better align with OpenTelemetry conventions. In tandem with this change,
several metrics that were improperly classified as `GAUGE`s were re-classified as `COUNTER`s.

### Metric name changes
The `/delta` suffix has been removed from the following metrics:
* `io.confluent.kafka.server/received_bytes/delta`
* `io.confluent.kafka.server/sent_bytes/delta`
* `io.confluent.kafka.server/request_count/delta`

### 2020-09-15

#### Retire `/available` endpoint
The `/available` endpoint (which was in _Preview_ status) has been removed from the API.
The `/descriptors` endpoint can still be used to determine the universe of available
metrics for Metrics API.

**The legacy metric names are deprecated and will stop functioning on 2020-07-01.**

## API Lifecycle Policy

The following status labels are applicable to APIs, features, and SDK versions, based on
the current support status of each:

* **Early Access** – May change at any time. Not recommended for production usage. Not
  officially supported by Confluent. Intended for user feedback only. Users must
be granted
  explicit access to the API by Confluent.
* **Preview** – Unlikely to change between Preview and General Availability. Not recommended
  for production usage. Officially supported by Confluent for non-production usage.
  For Closed Previews, users must be granted explicit access to the API by Confluent.
* **Generally Available (GA)** – Will not change at short notice. Recommended for production
  usage. Officially supported by Confluent for non-production and production usage.
* **Deprecated** – No longer supported. Will be removed in the future at the announced date.
  Use is discouraged and migration following the upgrade guide is recommended.
* **Sunset** – Removed, and no longer supported or available.

Resources, operations, and individual fields in the
<a href=\"./api.yaml\" target=\"_blank\">OpenAPI spec</a> will be annotated with
`x-lifecycle-stage`, `x-deprecated-at`, and `x-sunset-at`. These annotations will appear in the
corresponding API Reference Documentation. An API is \"Generally Available\" unless explicitly
marked otherwise.

## Compatibility Policy

Confluent APIs are governed by
<a href=\"https://docs.confluent.io/current/cloud/limits.html#upgrade-policy\" target=\"_blank\">
Confluent Cloud Upgrade Policy</a> in which we will make backward incompatible changes and
deprecations approximately once per year, and will provide 180 days notice via email to all
registered Confluent Cloud users.

### Backward Compatibility

> *An API version is backwards-compatible if a program written against the previous version of
> the API will continue to work the same way, without modification, against this version of the
> API.*

Confluent considers the following changes to be backwards-compatible:

* Adding new API resources.
* Adding new optional parameters to existing API requests (e.g., query string or body).
* Adding new properties to existing API responses.
* Changing the order of properties in existing API responses.
* Changing the length or format of object IDs or other opaque strings.
  * Unless otherwise documented, you can safely assume object IDs we generate
will never exceed
    255 characters, but you should be able to handle IDs of up to that length.
    If you're using MySQL, for example, you should store IDs in a
    `VARCHAR(255) COLLATE utf8_bin` column.
  * This includes adding or removing fixed prefixes (such as `lkc-` on kafka cluster
IDs).
  * This includes API keys, API tokens, and similar authentication mechanisms.
  * This includes all strings described as \"opaque\" in the docs, such as pagination
cursors. * Omitting properties with null values from existing API responses.

### Client Responsibilities

* Resource and rate limits, and the default and maximum sizes of paginated data **are not**
  considered part of the API contract and may change (possibly dynamically). It
is the client's
  responsibility to read the road signs and obey the speed limit.
* If a property has a primitive type and the API documentation does not explicitly limit its
  possible values, clients **must not** assume the values are constrained to a
particular set
  of possible responses.
* If a property of an object is not explicitly declared as mandatory in the API, clients
  **must not** assume it will be present.
* A resource **may** be modified to return a \"redirection\" response (e.g. `301`, `307`) instead
  of directly returning the resource. Clients **must** handle HTTP-level redirects,
and respect
  HTTP headers (e.g. `Location`).

## Deprecation Policy

Confluent will announce deprecations at least 180 days in advance of a breaking change
and we will continue to maintain the deprecated APIs in their original form during this time.

Exceptions to this policy apply in case of critical security vulnerabilities or functional
defects.

### Communication

When a deprecation is announced, the details and any relevant migration
information will be available on the following channels:

* Publication in the [API Changelog](#section/Versioning/Changelog)
* Lifecycle, deprecation and \"x-deprecated-at\" annotations in the
  <a href=\"/docs/api.yaml\" target=\"_blank\">OpenAPI spec</a>
* Announcements on the
  <a href=\"https://www.confluent.io/blog/\" target=\"_blank\">Developer Blog</a>,
  <a href=\"https://confluentcommunity.slack.com\" target=\"_blank\">Community Slack</a>
  (<a href=\"https://slackpass.io/confluentcommunity\" target=\"_blank\">join!</a>),
  <a href=\"https://groups.google.com/forum/#!forum/confluent-platform\" target=\"_blank\">
  Google Group</a>,
  the <a href=\"https://twitter.com/ConfluentInc\" target=\"_blank\">@ConfluentInc
twitter</a>
  account, and similar channels
* Enterprise customers may receive information by email to their specified Confluent contact,
  if applicable.

# Object Model
The object model for the Metrics API is designed similarly to the
[OpenTelemetry](https://opentelemetry.io/) standard.

## Metrics
A _metric_ is a numeric attribute of a resource, measured at a specific point in time, labeled
with contextual metadata gathered at the point of instrumentation.

There are two types of metrics:
* `GAUGE`: An instantaneous measurement of a value.
  Gauge metrics are implicitly averaged when aggregating over time.
  > Example: `io.confluent.kafka.server/retained_bytes`
* `COUNTER`: The count of occurrences in a _single (one minute) sampling
  interval_ (unless otherwise stated in the metric description).
  Counter metrics are implicitly summed when aggregating over time.
  > Example: `io.confluent.kafka.server/received_bytes`

The list of metrics and their labels is available at [/docs/descriptors](/docs/descriptors).

## Resources
A _resource_ represents an entity against which metrics are collected.  For example, a Kafka
cluster, a Kafka Connector, a ksqlDB application, etc.

Each metric _descriptor_ is associated with one or more resource _descriptors_, representing
the resource types to which that metric can apply.  A metric _data point_ is associated with a
single resource _instance_, identified by the resource labels on that metric data point.

For example, metrics emitted by Kafka Connect are associated to the `connector` resource type.
Data points for those metrics include resource labels identifying the specific `connector`
instance that emitted the metric.

The list of resource types and labels are discoverable via the `/descriptors/resources`
endpoint.

## Labels
A _label_ is a key-value attribute associated with a metric data point.

Labels can be used in queries to filter or group the results.  Labels must be prefixed when
used in queries:
* `metric.<label>` (for metric labels), for example `metric.topic`
* `resource.<resource-type>.<label>` (for resource labels), for example `resource.kafka.id`.

The set of valid label keys for a metric include:
* The label keys defined on that metric's descriptor itself
* The label keys defined on the resource descriptor for the metric's associated resource type

For example, the `io.confluent.kafka.server/received_bytes` metric has the following labels:
* `resource.kafka.id` - The Kafka cluster to which the metric pertains
* `metric.topic` - The Kafka topic to which the bytes were produced
* `metric.partition` - The partition to which the bytes were produced

## Datasets
A _dataset_ is a logical collection of metrics that can be queried together.  The `dataset` is
a required URL template parameter for every endpoint in this API.  The following datasets are
currently available:

<table>
<thead>
  <tr>
    <th style=\"width: 250px;\">Dataset</th>
    <th>Description</th>
  </tr>
</thead>
<tbody>
  <tr>
    <td>
      <code>cloud</code>
      <p><img
          src=\"https://img.shields.io/badge/Lifecycle%20Stage-Generally%20Available-%230074A2\"
          alt=\"generally-available\">
    </td>
    <td>
      Metrics originating from Confluent Cloud resources.
      <p>Requests to this dataset require a resource <code>filter</code>
         (e.g. Kafka cluster ID, Connector ID, etc.) in the query for authorization
purposes.
         The client's API key must be authorized for the resource referenced in
the filter.
    </td>
  </tr>
</tbody>
</table>

# Client Considerations and Best Practices

## Rate Limiting
To protect the stability of the API and keep it available to all users, Confluent employs
multiple safeguards. Users who send many requests in quick succession or perform too many
concurrent operations may be throttled or have their requested rejected with an error.
When a rate limit is breached, an HTTP `429 Too Many Requests` error is returned.

Rate limits are enforced at multiple scopes.

### Global Rate Limits
A global rate limit of **60 requests per IP address, per minute** is enforced.

### Per-endpoint Rate Limits
Additionally, some endpoint-specific rate limits are enforced.

| Endpoint  | Rate limit |
| --------- | ---------- |
| `/v2/metrics/{dataset}/export` | 80 requests per resource, per hour, per principal.<br/>See the [export endpoint documentation](#tag/Version-2/paths/~1v2~1metrics~1{dataset}~1export/get) for details. |

## Retries
Implement retry logic in your client to gracefully handle transient API failures.
This should be done by watching for error responses and building in a retry mechanism.
This mechanism should follow a capped exponential backoff policy to prevent retry
amplification (\"retry storms\") and also introduce some randomness (\"jitter\") to avoid the
[thundering herd effect](https://en.wikipedia.org/wiki/Thundering_herd_problem).

## Metric Data Latency
Metric data points are typically available for query in the API within **5 minutes** of their
origination at the source.  This latency can vary based on network conditions and processing
overhead.  Clients that are polling (or \"scraping\") metrics into an external monitoring system
should account for this latency in their polling requests.  API requests that fail to
incorporate the latency into the query `interval` may have incomplete data in the response.

## Pagination
Cursors, tokens, and corresponding pagination links may expire after a short amount of time.
In this case, the API will return a `400 Bad Request` error and the client will need to restart
from the beginning.

The client should have no trouble pausing between rate limiting windows, but persisting cursors
for hours or days is not recommended.


## Overview
This API client was generated by the [OpenAPI Generator](https://openapi-generator.tech) project.  By using the [OpenAPI-spec](https://www.openapis.org/) from a remote server, you can easily generate an API client.

- API version: 
- Package version: 1.0.0
- Build package: org.openapitools.codegen.languages.GoClientCodegen
For more information, please visit [https://confluent.io](https://confluent.io)

## Installation

Install the following dependencies:

```shell
go get github.com/stretchr/testify/assert
go get golang.org/x/oauth2
go get golang.org/x/net/context
```

Put the package under your project folder and add the following in import:

```golang
import v2 "github.com/confluentinc/ccloud-sdk-go-v2"
```

To use a proxy, set the environment variable `HTTP_PROXY`:

```golang
os.Setenv("HTTP_PROXY", "http://proxy_name:proxy_port")
```

## Configuration of Server URL

Default configuration comes with `Servers` field that contains server objects as defined in the OpenAPI specification.

### Select Server Configuration

For using other server than the one defined on index 0 set context value `sw.ContextServerIndex` of type `int`.

```golang
ctx := context.WithValue(context.Background(), v2.ContextServerIndex, 1)
```

### Templated Server URL

Templated server URL is formatted using default variables from configuration or from context value `sw.ContextServerVariables` of type `map[string]string`.

```golang
ctx := context.WithValue(context.Background(), v2.ContextServerVariables, map[string]string{
	"basePath": "v2",
})
```

Note, enum values are always validated and all unused variables are silently ignored.

### URLs Configuration per Operation

Each operation can use different server URL defined using `OperationServers` map in the `Configuration`.
An operation is uniquely identified by `"{classname}Service.{nickname}"` string.
Similar rules for overriding default operation server index and variables applies by using `sw.ContextOperationServerIndices` and `sw.ContextOperationServerVariables` context maps.

```
ctx := context.WithValue(context.Background(), v2.ContextOperationServerIndices, map[string]int{
	"{classname}Service.{nickname}": 2,
})
ctx = context.WithValue(context.Background(), v2.ContextOperationServerVariables, map[string]map[string]string{
	"{classname}Service.{nickname}": {
		"port": "8443",
	},
})
```

## Documentation for API Endpoints

All URIs are relative to *https://api.telemetry.confluent.cloud*

Class | Method | HTTP request | Description
------------ | ------------- | ------------- | -------------
*Version1Api* | [**V1MetricsDatasetAttributesPost**](docs/Version1Api.md#v1metricsdatasetattributespost) | **Post** /v1/metrics/{dataset}/attributes | Query label values
*Version1Api* | [**V1MetricsDatasetDescriptorsGet**](docs/Version1Api.md#v1metricsdatasetdescriptorsget) | **Get** /v1/metrics/{dataset}/descriptors | List all metric descriptors
*Version1Api* | [**V1MetricsDatasetQueryPost**](docs/Version1Api.md#v1metricsdatasetquerypost) | **Post** /v1/metrics/{dataset}/query | Query metric values
*Version2Api* | [**V2MetricsDatasetAttributesPost**](docs/Version2Api.md#v2metricsdatasetattributespost) | **Post** /v2/metrics/{dataset}/attributes | Query label values
*Version2Api* | [**V2MetricsDatasetDescriptorsMetricsGet**](docs/Version2Api.md#v2metricsdatasetdescriptorsmetricsget) | **Get** /v2/metrics/{dataset}/descriptors/metrics | List metric descriptors
*Version2Api* | [**V2MetricsDatasetDescriptorsResourcesGet**](docs/Version2Api.md#v2metricsdatasetdescriptorsresourcesget) | **Get** /v2/metrics/{dataset}/descriptors/resources | List resource descriptors
*Version2Api* | [**V2MetricsDatasetExportGet**](docs/Version2Api.md#v2metricsdatasetexportget) | **Get** /v2/metrics/{dataset}/export | Export metric values
*Version2Api* | [**V2MetricsDatasetQueryPost**](docs/Version2Api.md#v2metricsdatasetquerypost) | **Post** /v2/metrics/{dataset}/query | Query metric values


## Documentation For Models

 - [Aggregation](docs/Aggregation.md)
 - [AggregationFunction](docs/AggregationFunction.md)
 - [AttributesRequest](docs/AttributesRequest.md)
 - [AttributesResponse](docs/AttributesResponse.md)
 - [CompoundFilter](docs/CompoundFilter.md)
 - [ErrorResponse](docs/ErrorResponse.md)
 - [ErrorResponseErrorsInner](docs/ErrorResponseErrorsInner.md)
 - [FieldFilter](docs/FieldFilter.md)
 - [FieldFilterValue](docs/FieldFilterValue.md)
 - [Filter](docs/Filter.md)
 - [FlatQueryResponse](docs/FlatQueryResponse.md)
 - [Granularity](docs/Granularity.md)
 - [GroupedQueryResponse](docs/GroupedQueryResponse.md)
 - [GroupedQueryResponseDataInner](docs/GroupedQueryResponseDataInner.md)
 - [LabelDescriptor](docs/LabelDescriptor.md)
 - [Links](docs/Links.md)
 - [ListMetricDescriptorsResponse](docs/ListMetricDescriptorsResponse.md)
 - [ListResourceDescriptorsResponse](docs/ListResourceDescriptorsResponse.md)
 - [Meta](docs/Meta.md)
 - [MetricDescriptor](docs/MetricDescriptor.md)
 - [OrderBy](docs/OrderBy.md)
 - [Pagination](docs/Pagination.md)
 - [Point](docs/Point.md)
 - [QueryRequest](docs/QueryRequest.md)
 - [QueryResponse](docs/QueryResponse.md)
 - [ResourceDescriptor](docs/ResourceDescriptor.md)
 - [ResponseFormat](docs/ResponseFormat.md)
 - [UnaryFilter](docs/UnaryFilter.md)
 - [V1MetricsDatasetAttributesPostRequest](docs/V1MetricsDatasetAttributesPostRequest.md)


## Documentation For Authorization



### api-key

- **Type**: HTTP basic authentication

Example

```golang
auth := context.WithValue(context.Background(), sw.ContextBasicAuth, sw.BasicAuth{
    UserName: "username",
    Password: "password",
})
r, err := client.Service.Operation(auth, args)
```


## Documentation for Utility Methods

Due to the fact that model structure members are all pointers, this package contains
a number of utility functions to easily obtain pointers to values of basic types.
Each of these functions takes a value of the given basic type and returns a pointer to it:

* `PtrBool`
* `PtrInt`
* `PtrInt32`
* `PtrInt64`
* `PtrFloat`
* `PtrFloat32`
* `PtrFloat64`
* `PtrString`
* `PtrTime`

## Author

support@confluent.io

