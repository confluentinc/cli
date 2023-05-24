package kafkarest

import (
	"encoding/json"
	"fmt"
	"net/http"
	neturl "net/url"
	"strings"

	cckafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

const (
	SelfSignedCertError   = "x509: certificate is not authorized to sign other certificates"
	UnauthorizedCertError = "x509: certificate signed by unknown authority"
)

type V3Error struct {
	Code    int    `json:"error_code"`
	Message string `json:"message"`
}

func NewError(url string, err error, httpResp *http.Response) error {
	switch e := err.(type) {
	case *neturl.Error:
		if strings.Contains(e.Error(), SelfSignedCertError) || strings.Contains(e.Error(), UnauthorizedCertError) {
			return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.KafkaRestConnectionErrorMsg, url, e.Err), errors.KafkaRestCertErrorSuggestions)
		}
		return errors.Errorf(errors.KafkaRestConnectionErrorMsg, url, e.Err)
	case cckafkarestv3.GenericOpenAPIError:
		openAPIError, parseErr := ParseOpenAPIErrorCloud(err)
		if parseErr == nil {
			if strings.Contains(openAPIError.Message, "invalid_token") {
				return errors.NewErrorWithSuggestions(errors.InvalidMDSTokenErrorMsg, errors.InvalidMDSTokenSuggestions)
			}
			return fmt.Errorf("REST request failed: %v (%v)", openAPIError.Message, openAPIError.Code)
		}
		if httpResp != nil && httpResp.StatusCode >= 400 {
			return kafkaRestHttpError(httpResp)
		}
		return errors.NewErrorWithSuggestions(errors.UnknownErrorMsg, errors.InternalServerErrorSuggestions)
	case kafkarestv3.GenericOpenAPIError:
		openAPIError, parseErr := parseOpenAPIError(err)
		if parseErr == nil {
			if strings.Contains(openAPIError.Message, "invalid_token") {
				return errors.NewErrorWithSuggestions(errors.InvalidMDSTokenErrorMsg, errors.InvalidMDSTokenSuggestions)
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

func ParseOpenAPIErrorCloud(err error) (*V3Error, error) {
	if openAPIError, ok := err.(cckafkarestv3.GenericOpenAPIError); ok {
		var decodedError V3Error
		if err := json.Unmarshal(openAPIError.Body(), &decodedError); err != nil {
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

func parseOpenAPIError(err error) (*V3Error, error) {
	if openAPIError, ok := err.(kafkarestv3.GenericOpenAPIError); ok {
		var decodedError V3Error
		if err := json.Unmarshal(openAPIError.Body(), &decodedError); err != nil {
			return nil, err
		}
		return &decodedError, nil
	}
	return nil, fmt.Errorf("unexpected type")
}
