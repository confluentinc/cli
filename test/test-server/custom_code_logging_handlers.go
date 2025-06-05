package testserver

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	cclv1 "github.com/confluentinc/ccloud-sdk-go-v2/ccl/v1"
)

const (
	KAFKA = "Kafka"
)

// Handler for: "/ccl/v1/custom-code-loggings"
func handleCustomCodeLoggings(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var customCodeLogging cclv1.CclV1CustomCodeLogging
			customCodeLogging = cclv1.CclV1CustomCodeLogging{
				Id:          cclv1.PtrString("ccl-123456"),
				Cloud:       cclv1.PtrString("AWS"),
				Region:      cclv1.PtrString("us-west-2"),
				Environment: &cclv1.EnvScopedObjectReference{Id: "env-000000"},
				DestinationSettings: &cclv1.CclV1CustomCodeLoggingDestinationSettingsOneOf{
					CclV1KafkaDestinationSettings: &cclv1.CclV1KafkaDestinationSettings{
						Kind:      KAFKA,
						Topic:     "topic-123",
						ClusterId: "cluster-123",
						LogLevel:  cclv1.PtrString("INFO"),
					},
				},
			}
			err := json.NewEncoder(w).Encode(customCodeLogging)
			require.NoError(t, err)
		case http.MethodGet:
			customCodeLoggingList := &cclv1.CclV1CustomCodeLoggingList{
				Data: []cclv1.CclV1CustomCodeLogging{
					{
						Id:          cclv1.PtrString("ccl-123456"),
						Cloud:       cclv1.PtrString("AWS"),
						Region:      cclv1.PtrString("us-west-2"),
						Environment: &cclv1.EnvScopedObjectReference{Id: "env-000000"},
						DestinationSettings: &cclv1.CclV1CustomCodeLoggingDestinationSettingsOneOf{
							CclV1KafkaDestinationSettings: &cclv1.CclV1KafkaDestinationSettings{
								Kind:      KAFKA,
								Topic:     "topic-123",
								ClusterId: "cluster-123",
								LogLevel:  cclv1.PtrString("INFO"),
							},
						},
					},
					{
						Id:          cclv1.PtrString("ccl-456789"),
						Cloud:       cclv1.PtrString("AWS"),
						Region:      cclv1.PtrString("us-west-2"),
						Environment: &cclv1.EnvScopedObjectReference{Id: "env-111111"},
						DestinationSettings: &cclv1.CclV1CustomCodeLoggingDestinationSettingsOneOf{
							CclV1KafkaDestinationSettings: &cclv1.CclV1KafkaDestinationSettings{
								Kind:      KAFKA,
								Topic:     "topic-456",
								ClusterId: "cluster-456",
								LogLevel:  cclv1.PtrString("ERROR"),
							},
						},
					},
					{
						Id:          cclv1.PtrString("ccl-789012"),
						Cloud:       cclv1.PtrString("AWS"),
						Region:      cclv1.PtrString("us-west-2"),
						Environment: &cclv1.EnvScopedObjectReference{Id: "env-222222"},
						DestinationSettings: &cclv1.CclV1CustomCodeLoggingDestinationSettingsOneOf{
							CclV1KafkaDestinationSettings: &cclv1.CclV1KafkaDestinationSettings{
								Kind:      KAFKA,
								Topic:     "topic-789",
								ClusterId: "cluster-789",
								LogLevel:  cclv1.PtrString("DEBUG"),
							},
						},
					},
				},
			}
			setPageToken(customCodeLoggingList, &customCodeLoggingList.Metadata, r.URL)
			err := json.NewEncoder(w).Encode(customCodeLoggingList)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/ccl/v1/custom-code-loggings/{id}"
func handleCustomCodeLoggingsId(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			vars := mux.Vars(r)
			id := vars["id"]
			var customCodeLogging cclv1.CclV1CustomCodeLogging
			if id == "ccl-123456" {
				customCodeLogging = cclv1.CclV1CustomCodeLogging{
					Id:          cclv1.PtrString("ccl-123456"),
					Cloud:       cclv1.PtrString("AWS"),
					Region:      cclv1.PtrString("us-west-2"),
					Environment: &cclv1.EnvScopedObjectReference{Id: "env-000000"},
					DestinationSettings: &cclv1.CclV1CustomCodeLoggingDestinationSettingsOneOf{
						CclV1KafkaDestinationSettings: &cclv1.CclV1KafkaDestinationSettings{
							Kind:      KAFKA,
							Topic:     "topic-123",
							ClusterId: "cluster-123",
							LogLevel:  cclv1.PtrString("INFO"),
						},
					},
				}
			} else if id == "ccl-789012" {
				customCodeLogging = cclv1.CclV1CustomCodeLogging{
					Id:          cclv1.PtrString("ccl-789012"),
					Cloud:       cclv1.PtrString("AWS"),
					Region:      cclv1.PtrString("us-west-2"),
					Environment: &cclv1.EnvScopedObjectReference{Id: "env-222222"},
					DestinationSettings: &cclv1.CclV1CustomCodeLoggingDestinationSettingsOneOf{
						CclV1KafkaDestinationSettings: &cclv1.CclV1KafkaDestinationSettings{
							Kind:      KAFKA,
							Topic:     "topic-789",
							ClusterId: "cluster-789",
							LogLevel:  cclv1.PtrString("DEBUG"),
						},
					},
				}
			} else {
				customCodeLogging = cclv1.CclV1CustomCodeLogging{
					Id:          cclv1.PtrString("ccl-456789"),
					Cloud:       cclv1.PtrString("AWS"),
					Region:      cclv1.PtrString("us-west-2"),
					Environment: &cclv1.EnvScopedObjectReference{Id: "env-111111"},
					DestinationSettings: &cclv1.CclV1CustomCodeLoggingDestinationSettingsOneOf{
						CclV1KafkaDestinationSettings: &cclv1.CclV1KafkaDestinationSettings{
							Kind:      KAFKA,
							Topic:     "topic-456",
							ClusterId: "cluster-456",
							LogLevel:  cclv1.PtrString("ERROR"),
						},
					},
				}
			}
			err := json.NewEncoder(w).Encode(customCodeLogging)
			require.NoError(t, err)
		case http.MethodPatch:
			customCodeLogging := cclv1.CclV1CustomCodeLogging{
				Id: cclv1.PtrString("ccp-123456"),
				DestinationSettings: &cclv1.CclV1CustomCodeLoggingDestinationSettingsOneOf{
					CclV1KafkaDestinationSettings: &cclv1.CclV1KafkaDestinationSettings{
						Topic:     "topic-456",
						ClusterId: "cluster-456",
						LogLevel:  cclv1.PtrString("ERROR"),
					},
				},
			}
			err := json.NewEncoder(w).Encode(customCodeLogging)
			require.NoError(t, err)
		case http.MethodDelete:
			err := json.NewEncoder(w).Encode(cclv1.CclV1CustomCodeLogging{})
			require.NoError(t, err)
		}
	}
}
