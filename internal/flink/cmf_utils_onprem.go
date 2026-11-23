package flink

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	"github.com/confluentinc/cli/v4/pkg/output"
)

func addCmfFlagSet(cmd *cobra.Command) {
	cmd.Flags().String("url", "", `Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.`)
	cmd.Flags().String("client-key-path", "", `Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.`)
	cmd.Flags().String("client-cert-path", "", `Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.`)
	cmd.Flags().String("certificate-authority-path", "", `Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CMF_CERTIFICATE_AUTHORITY_PATH" may be set in place of this flag.`)
}

func addPageSizeFlag(cmd *cobra.Command) {
	cmd.Flags().Int32("page-size", 100, "Number of results to fetch per API request while paginating; does not cap the total results returned.")
}

func getPageSize(cmd *cobra.Command) (int32, error) {
	return cmd.Flags().GetInt32("page-size")
}

// resourceEventOut is a row of the events table, with the columns of the CMF UI's "Events log".
type resourceEventOut struct {
	Severity        string `human:"Severity" serialized:"severity"`
	Reason          string `human:"Reason" serialized:"reason"`
	Message         string `human:"Message" serialized:"message"`
	Count           int32  `human:"Count" serialized:"count"`
	FirstOccurrence string `human:"First Occurrence" serialized:"first_occurrence"`
	LastOccurrence  string `human:"Last Occurrence" serialized:"last_occurrence"`
}

// printResourceEvents prints the events of a statement or compute pool in the order given: as a table, or as the
// ResourceEvents that CMF returned.
func printResourceEvents(cmd *cobra.Command, events []cmfsdk.ResourceEvent) error {
	if output.GetFormat(cmd) != output.Human {
		return output.SerializedOutputFromJsonTags(cmd, events)
	}

	list := output.NewList(cmd)
	list.Sort(false)
	for _, event := range events {
		// The CMF UI's rule: the reason of the Kubernetes event behind it, if any, otherwise the event type.
		reason := event.Status.GetType()
		if data := event.Status.Data; data != nil && data.EventDataReasonMessage.GetReason() != "" {
			reason = data.EventDataReasonMessage.GetReason()
		}

		var lastOccurrence string
		if updateTimestamp, ok := event.Metadata.GetUpdateTimestampOk(); ok {
			lastOccurrence = updateTimestamp.UTC().Format(time.RFC3339)
		}

		list.Add(&resourceEventOut{
			Severity:        event.Status.GetSeverity(),
			Reason:          reason,
			Message:         event.Status.GetMessage(),
			Count:           event.Status.GetCount(),
			FirstOccurrence: formatEventTimestamp(event.Metadata.GetCreationTimestamp()),
			LastOccurrence:  lastOccurrence,
		})
	}
	return list.Print()
}

// formatEventTimestamp prints CMF's creationTimestamp to the second, in UTC, as updateTimestamp is printed. A value
// that doesn't parse is printed as it is.
func formatEventTimestamp(timestamp string) string {
	parsed, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return timestamp
	}
	return parsed.UTC().Format(time.RFC3339)
}

func (c *command) createContext() context.Context {
	if !c.Config.IsOnPremLogin() {
		return context.Background()
	}
	return context.WithValue(context.Background(), cmfsdk.ContextAccessToken, c.Context.GetAuthToken())
}

func (c *command) handleFlinkWebUiForwardRequest(userResponseWriter http.ResponseWriter, userRequest *http.Request, url, environmentName, resourceType, resourceName, userAgent string, client *http.Client) {
	body, err := io.ReadAll(userRequest.Body)
	if err != nil {
		http.Error(userResponseWriter, fmt.Sprintf("Failed to read request body: %s", err), http.StatusInternalServerError)
		return
	}

	newUrl := fmt.Sprintf("%s/cmf/api/v1/environments/%s/%s/%s/flink-web-ui%s", url, environmentName, resourceType, resourceName, userRequest.RequestURI)
	reqToCmf, err := http.NewRequest(userRequest.Method, newUrl, bytes.NewReader(body))
	if err != nil {
		http.Error(userResponseWriter, fmt.Sprintf("failed to forward the web UI: %s", err), http.StatusInternalServerError)
		return
	}
	reqToCmf.Header = userRequest.Header
	reqToCmf.Header.Set("x-confluent-cli-version", userAgent)

	if c.Config.IsOnPremLogin() {
		accessToken := c.Context.GetAuthToken()
		reqToCmf.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	}

	resFromCmf, err := client.Do(reqToCmf)
	if err != nil {
		http.Error(userResponseWriter, fmt.Sprintf("failed to forward the request: %s", err), http.StatusInternalServerError)
		return
	}
	defer resFromCmf.Body.Close()

	// Copy response headers - this includes content type.
	for key, values := range resFromCmf.Header {
		for _, value := range values {
			userResponseWriter.Header().Set(key, value)
		}
	}
	userResponseWriter.WriteHeader(resFromCmf.StatusCode)

	// Copy response body.
	resBody, err := io.ReadAll(resFromCmf.Body)
	if err != nil {
		http.Error(userResponseWriter, fmt.Sprintf("failed to return response from the web UI: %s", err), http.StatusInternalServerError)
		return
	}
	_, err = userResponseWriter.Write(resBody)
	if err != nil {
		output.ErrPrintf(false, "Failed to write response body: %s", err)
	}
}

// copyDataType recursively converts the SDK DataType to a LocalDataType.
func copyDataType(sdkType cmfsdk.DataType) LocalDataType {
	localType := LocalDataType{
		Type:                sdkType.Type,
		Nullable:            sdkType.Nullable,
		Length:              sdkType.Length,
		Precision:           sdkType.Precision,
		Scale:               sdkType.Scale,
		Resolution:          sdkType.Resolution,
		FractionalPrecision: sdkType.FractionalPrecision,
	}
	if sdkType.KeyType != nil {
		copiedKeyType := copyDataType(*sdkType.KeyType)
		localType.KeyType = &copiedKeyType
	}
	if sdkType.ValueType != nil {
		copiedValueType := copyDataType(*sdkType.ValueType)
		localType.ValueType = &copiedValueType
	}
	if sdkType.ElementType != nil {
		copiedElementType := copyDataType(*sdkType.ElementType)
		localType.ElementType = &copiedElementType
	}
	if sdkType.Fields != nil {
		localFields := make([]LocalDataTypeField, 0, len(*sdkType.Fields))
		for _, sdkField := range *sdkType.Fields {
			localFields = append(localFields, LocalDataTypeField{
				Name:        sdkField.Name,
				FieldType:   copyDataType(sdkField.FieldType),
				Description: sdkField.Description,
			})
		}
		localType.Fields = &localFields
	}
	return localType
}
