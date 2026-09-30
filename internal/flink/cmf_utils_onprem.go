package flink

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

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
