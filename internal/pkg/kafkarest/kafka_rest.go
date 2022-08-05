package kafkarest

import (
	"encoding/json"
	"fmt"
	"net/http"
	neturl "net/url"
	"strings"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

const (
	BadRequestErrorCode              = 40002
	UnknownTopicOrPartitionErrorCode = 40403
)

const (
	selfSignedCertError   = "x509: certificate is not authorized to sign other certificates"
	unauthorizedCertError = "x509: certificate signed by unknown authority"
)

type Error struct {
	Code    int    `json:"error_code"`
	Message string `json:"message"`
}

func NewError(url string, err error, httpResp *http.Response) error {
	switch e := err.(type) {
	case *neturl.Error:
		if strings.Contains(e.Error(), selfSignedCertError) || strings.Contains(e.Error(), unauthorizedCertError) {
			return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.KafkaRestConnectionMsg, url, e.Err), errors.KafkaRestCertErrorSuggestions)
		}
		return errors.Errorf(errors.KafkaRestConnectionMsg, url, e.Err)
	case kafkarestv3.GenericOpenAPIError:
		openAPIError, parseErr := ParseOpenAPIError(err)
		if parseErr == nil {
			if strings.Contains(openAPIError.Message, "invalid_token") {
				return errors.NewErrorWithSuggestions(errors.InvalidMDSToken, errors.InvalidMDSTokenSuggestions)
			}
			return fmt.Errorf("REST request failed: %v (%v)", openAPIError.Message, openAPIError.Code)
		}
		if httpResp != nil && httpResp.StatusCode >= 400 {
			return kafkaRestHttpError(httpResp)
		}
		return errors.NewErrorWithSuggestions(errors.UnknownErrorMsg, errors.InternalServerErrorSuggestions)
	}
	return err
}

func ParseOpenAPIError(err error) (*Error, error) {
	if openAPIError, ok := err.(kafkarestv3.GenericOpenAPIError); ok {
		var decodedError Error
		err = json.Unmarshal(openAPIError.Body(), &decodedError)
		if err != nil {
			return nil, err
		}
		return &decodedError, nil
	}
	return nil, fmt.Errorf("unexpected type")
}

func kafkaRestHttpError(httpResp *http.Response) error {
	return errors.NewErrorWithSuggestions(
		fmt.Sprintf(errors.KafkaRestErrorMsg, httpResp.Request.Method, httpResp.Request.URL, httpResp.Status),
		errors.InternalServerErrorSuggestions)
}
