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
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/oauth2"
)

var (
	jsonCheck = regexp.MustCompile(`(?i:(?:application|text)/(?:vnd\.[^;]+\+)?json)`)
	xmlCheck  = regexp.MustCompile(`(?i:(?:application|text)/xml)`)
)

// APIClient manages communication with the Confluent Cloud Metrics API API v
// In most cases there should be only one, shared, APIClient.
type APIClient struct {
	cfg    *Configuration
	common service // Reuse a single struct instead of allocating one for each service on the heap.

	// API Services

	Version1Api Version1Api

	Version2Api Version2Api
}

type service struct {
	client *APIClient
}

// NewAPIClient creates a new API client. Requires a userAgent string describing your application.
// optionally a custom http.Client to allow for advanced features such as caching.
func NewAPIClient(cfg *Configuration) *APIClient {
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = http.DefaultClient
	}

	c := &APIClient{}
	c.cfg = cfg
	c.common.client = c

	// API Services
	c.Version1Api = (*Version1ApiService)(&c.common)
	c.Version2Api = (*Version2ApiService)(&c.common)

	return c
}

func atoi(in string) (int, error) {
	return strconv.Atoi(in)
}

// selectHeaderContentType select a content type from the available list.
func selectHeaderContentType(contentTypes []string) string {
	if len(contentTypes) == 0 {
		return ""
	}
	if contains(contentTypes, "application/json") {
		return "application/json"
	}
	return contentTypes[0] // use the first content type specified in 'consumes'
}

// selectHeaderAccept join all accept types and return
func selectHeaderAccept(accepts []string) string {
	if len(accepts) == 0 {
		return ""
	}

	if contains(accepts, "application/json") {
		return "application/json"
	}

	return strings.Join(accepts, ",")
}

// contains is a case insensitive match, finding needle in a haystack
func contains(haystack []string, needle string) bool {
	for _, a := range haystack {
		if strings.ToLower(a) == strings.ToLower(needle) {
			return true
		}
	}
	return false
}

// Verify optional parameters are of the correct type.
func typeCheckParameter(obj interface{}, expected string, name string) error {
	// Make sure there is an object.
	if obj == nil {
		return nil
	}

	// Check the type is as expected.
	if reflect.TypeOf(obj).String() != expected {
		return fmt.Errorf("Expected %s to be of type %s but received %s.", name, expected, reflect.TypeOf(obj).String())
	}
	return nil
}

// parameterToString convert interface{} parameters to string, using a delimiter if format is provided.
func parameterToString(obj interface{}, collectionFormat string) string {
	var delimiter string

	switch collectionFormat {
	case "pipes":
		delimiter = "|"
	case "ssv":
		delimiter = " "
	case "tsv":
		delimiter = "\t"
	case "csv":
		delimiter = ","
	}

	if reflect.TypeOf(obj).Kind() == reflect.Slice {
		return strings.Trim(strings.Replace(fmt.Sprint(obj), " ", delimiter, -1), "[]")
	} else if t, ok := obj.(time.Time); ok {
		return t.Format(time.RFC3339)
	}

	return fmt.Sprintf("%v", obj)
}

// helper for converting interface{} parameters to json strings
func parameterToJson(obj interface{}) (string, error) {
	jsonBuf, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(jsonBuf), err
}

// callAPI do the request.
func (c *APIClient) callAPI(request *http.Request) (*http.Response, error) {
	if c.cfg.Debug {
		dump, err := httputil.DumpRequestOut(request, true)
		if err != nil {
			return nil, err
		}
		log.Printf("\n%s\n", string(dump))
	}

	resp, err := c.cfg.HTTPClient.Do(request)
	if err != nil {
		return resp, err
	}

	if c.cfg.Debug {
		dump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return resp, err
		}
		log.Printf("\n%s\n", string(dump))
	}
	return resp, err
}

// Allow modification of underlying config for alternate implementations and testing
// Caution: modifying the configuration while live can cause data races and potentially unwanted behavior
func (c *APIClient) GetConfig() *Configuration {
	return c.cfg
}

type formFile struct {
		fileBytes []byte
		fileName string
		formFileName string
}

// prepareRequest build the request
func (c *APIClient) prepareRequest(
	ctx context.Context,
	path string, method string,
	postBody interface{},
	headerParams map[string]string,
	queryParams url.Values,
	formParams url.Values,
	formFiles []formFile) (localVarRequest *http.Request, err error) {

	var body *bytes.Buffer

	// Detect postBody type and post.
	if postBody != nil {
		contentType := headerParams["Content-Type"]
		if contentType == "" {
			contentType = detectContentType(postBody)
			headerParams["Content-Type"] = contentType
		}

		body, err = setBody(postBody, contentType)
		if err != nil {
			return nil, err
		}
	}

	// add form parameters and file if available.
	if strings.HasPrefix(headerParams["Content-Type"], "multipart/form-data") && len(formParams) > 0 || (len(formFiles) > 0) {
		if body != nil {
			return nil, errors.New("Cannot specify postBody and multipart form at the same time.")
		}
		body = &bytes.Buffer{}
		w := multipart.NewWriter(body)

		for k, v := range formParams {
			for _, iv := range v {
				if strings.HasPrefix(k, "@") { // file
					err = addFile(w, k[1:], iv)
					if err != nil {
						return nil, err
					}
				} else { // form value
					w.WriteField(k, iv)
				}
			}
		}
		for _, formFile := range formFiles {
			if len(formFile.fileBytes) > 0 && formFile.fileName != "" {
				w.Boundary()
				part, err := w.CreateFormFile(formFile.formFileName, filepath.Base(formFile.fileName))
				if err != nil {
						return nil, err
				}
				_, err = part.Write(formFile.fileBytes)
				if err != nil {
						return nil, err
				}
			}
		}

		// Set the Boundary in the Content-Type
		headerParams["Content-Type"] = w.FormDataContentType()

		// Set Content-Length
		headerParams["Content-Length"] = fmt.Sprintf("%d", body.Len())
		w.Close()
	}

	if strings.HasPrefix(headerParams["Content-Type"], "application/x-www-form-urlencoded") && len(formParams) > 0 {
		if body != nil {
			return nil, errors.New("Cannot specify postBody and x-www-form-urlencoded form at the same time.")
		}
		body = &bytes.Buffer{}
		body.WriteString(formParams.Encode())
		// Set Content-Length
		headerParams["Content-Length"] = fmt.Sprintf("%d", body.Len())
	}

	// Setup path and query parameters
	url, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	// Override request host, if applicable
	if c.cfg.Host != "" {
		url.Host = c.cfg.Host
	}

	// Override request scheme, if applicable
	if c.cfg.Scheme != "" {
		url.Scheme = c.cfg.Scheme
	}

	// Adding Query Param
	query := url.Query()
	for k, v := range queryParams {
		for _, iv := range v {
			query.Add(k, iv)
		}
	}

	// Encode the parameters.
	url.RawQuery = query.Encode()

	// Generate a new request
	if body != nil {
		localVarRequest, err = http.NewRequest(method, url.String(), body)
	} else {
		localVarRequest, err = http.NewRequest(method, url.String(), nil)
	}
	if err != nil {
		return nil, err
	}

	// add header parameters, if any
	if len(headerParams) > 0 {
		headers := http.Header{}
		for h, v := range headerParams {
			headers[h] = []string{v}
		}
		localVarRequest.Header = headers
	}

	// Add the user agent to the request.
	localVarRequest.Header.Add("User-Agent", c.cfg.UserAgent)

	if ctx != nil {
		// add context to the request
		localVarRequest = localVarRequest.WithContext(ctx)

		// Walk through any authentication.

		// OAuth2 authentication
		if tok, ok := ctx.Value(ContextOAuth2).(oauth2.TokenSource); ok {
			// We were able to grab an oauth2 token from the context
			var latestToken *oauth2.Token
			if latestToken, err = tok.Token(); err != nil {
				return nil, err
			}

			latestToken.SetAuthHeader(localVarRequest)
		}

		// Basic HTTP Authentication
		if auth, ok := ctx.Value(ContextBasicAuth).(BasicAuth); ok {
			localVarRequest.SetBasicAuth(auth.UserName, auth.Password)
		}

		// AccessToken Authentication
		if auth, ok := ctx.Value(ContextAccessToken).(string); ok {
			localVarRequest.Header.Add("Authorization", "Bearer "+auth)
		}

	}

	for header, value := range c.cfg.DefaultHeader {
		localVarRequest.Header.Add(header, value)
	}
	return localVarRequest, nil
}

func (c *APIClient) decode(v interface{}, b []byte, contentType string) (err error) {
	if len(b) == 0 {
		return nil
	}
	if s, ok := v.(*string); ok {
		*s = string(b)
		return nil
	}
	if f, ok := v.(**os.File); ok {
		*f, err = ioutil.TempFile("", "HttpClientFile")
		if err != nil {
			return
		}
		_, err = (*f).Write(b)
		if err != nil {
			return
		}
		_, err = (*f).Seek(0, io.SeekStart)
		return
	}
	if xmlCheck.MatchString(contentType) {
		if err = xml.Unmarshal(b, v); err != nil {
			return err
		}
		return nil
	}
	if jsonCheck.MatchString(contentType) {
		if actualObj, ok := v.(interface{ GetActualInstance() interface{} }); ok { // oneOf, anyOf schemas
			if unmarshalObj, ok := actualObj.(interface{ UnmarshalJSON([]byte) error }); ok { // make sure it has UnmarshalJSON defined
				if err = unmarshalObj.UnmarshalJSON(b); err != nil {
					return err
				}
			} else {
				return errors.New("Unknown type with GetActualInstance but no unmarshalObj.UnmarshalJSON defined")
			}
		} else if err = json.Unmarshal(b, v); err != nil { // simple model
			return err
		}
		return nil
	}
	return errors.New("undefined response type")
}

// Add a file to the multipart request
func addFile(w *multipart.Writer, fieldName, path string) error {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}

	part, err := w.CreateFormFile(fieldName, filepath.Base(path))
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)

	return err
}

// Prevent trying to import "fmt"
func reportError(format string, a ...interface{}) error {
	return fmt.Errorf(format, a...)
}

// A wrapper for strict JSON decoding
func newStrictDecoder(data []byte) *json.Decoder {
	dec := json.NewDecoder(bytes.NewBuffer(data))
	dec.DisallowUnknownFields()
	return dec
}

// Set request body from an interface{}
func setBody(body interface{}, contentType string) (bodyBuf *bytes.Buffer, err error) {
	if bodyBuf == nil {
		bodyBuf = &bytes.Buffer{}
	}

	if reader, ok := body.(io.Reader); ok {
		_, err = bodyBuf.ReadFrom(reader)
	} else if fp, ok := body.(**os.File); ok {
		_, err = bodyBuf.ReadFrom(*fp)
	} else if b, ok := body.([]byte); ok {
		_, err = bodyBuf.Write(b)
	} else if s, ok := body.(string); ok {
		_, err = bodyBuf.WriteString(s)
	} else if s, ok := body.(*string); ok {
		_, err = bodyBuf.WriteString(*s)
	} else if jsonCheck.MatchString(contentType) {
		err = json.NewEncoder(bodyBuf).Encode(body)
	} else if xmlCheck.MatchString(contentType) {
		err = xml.NewEncoder(bodyBuf).Encode(body)
	}

	if err != nil {
		return nil, err
	}

	if bodyBuf.Len() == 0 {
		err = fmt.Errorf("Invalid body type %s\n", contentType)
		return nil, err
	}
	return bodyBuf, nil
}

// detectContentType method is used to figure out `Request.Body` content type for request header
func detectContentType(body interface{}) string {
	contentType := "text/plain; charset=utf-8"
	kind := reflect.TypeOf(body).Kind()

	switch kind {
	case reflect.Struct, reflect.Map, reflect.Ptr:
		contentType = "application/json; charset=utf-8"
	case reflect.String:
		contentType = "text/plain; charset=utf-8"
	default:
		if b, ok := body.([]byte); ok {
			contentType = http.DetectContentType(b)
		} else if kind == reflect.Slice {
			contentType = "application/json; charset=utf-8"
		}
	}

	return contentType
}

// Ripped from https://github.com/gregjones/httpcache/blob/master/httpcache.go
type cacheControl map[string]string

func parseCacheControl(headers http.Header) cacheControl {
	cc := cacheControl{}
	ccHeader := headers.Get("Cache-Control")
	for _, part := range strings.Split(ccHeader, ",") {
		part = strings.Trim(part, " ")
		if part == "" {
			continue
		}
		if strings.ContainsRune(part, '=') {
			keyval := strings.Split(part, "=")
			cc[strings.Trim(keyval[0], " ")] = strings.Trim(keyval[1], ",")
		} else {
			cc[part] = ""
		}
	}
	return cc
}

// CacheExpires helper function to determine remaining time before repeating a request.
func CacheExpires(r *http.Response) time.Time {
	// Figure out when the cache expires.
	var expires time.Time
	now, err := time.Parse(time.RFC1123, r.Header.Get("date"))
	if err != nil {
		return time.Now()
	}
	respCacheControl := parseCacheControl(r.Header)

	if maxAge, ok := respCacheControl["max-age"]; ok {
		lifetime, err := time.ParseDuration(maxAge + "s")
		if err != nil {
			expires = now
		} else {
			expires = now.Add(lifetime)
		}
	} else {
		expiresHeader := r.Header.Get("Expires")
		if expiresHeader != "" {
			expires, err = time.Parse(time.RFC1123, expiresHeader)
			if err != nil {
				expires = now
			}
		}
	}
	return expires
}

func strlen(s string) int {
	return utf8.RuneCountInString(s)
}

// GenericOpenAPIError Provides access to the body, error and model on returned errors.
type GenericOpenAPIError struct {
	body  []byte
	error string
	model interface{}
}

// Error returns non-empty string if there was an error.
func (e GenericOpenAPIError) Error() string {
	return e.error
}

// Body returns the raw bytes of the response
func (e GenericOpenAPIError) Body() []byte {
	return e.body
}

// Model returns the unpacked model of the error
func (e GenericOpenAPIError) Model() interface{} {
	return e.model
}
