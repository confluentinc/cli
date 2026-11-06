package logs

import (
	"encoding/json"

	"github.com/tidwall/sjson"

	loggingv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/logging/v1"
)

// sourceJson renders a log source for JSON and YAML output. The SDK model still carries the `allowed_scopes` field
// that the API no longer documents, so it is removed here, keeping the other fields in the API's order.
func sourceJson(source loggingv1.LoggingV1Source) (json.RawMessage, error) {
	data, err := json.Marshal(source)
	if err != nil {
		return nil, err
	}
	return sjson.DeleteBytes(data, "allowed_scopes")
}
