package testserver

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	rtcev1 "github.com/confluentinc/ccloud-sdk-go-v2/rtce/v1"
)

// Handler for "/rtce/v1/rtce-topics"
func handleRtceV1RtceTopics(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			rtceTopic := readRtceV1RtceTopicFile(t, "read_created_rtce_topic.json")

			rtceTopicList := &rtcev1.RtceV1RtceTopicList{
				Data: []rtcev1.RtceV1RtceTopic{rtceTopic},
			}

			err := json.NewEncoder(w).Encode(rtceTopicList)
			require.NoError(t, err)
		case http.MethodPost:
			rtceTopic := readRtceV1RtceTopicFile(t, "create_rtce_topic.json")

			// Overwrite updated fields using the request body
			err := json.NewDecoder(r.Body).Decode(&rtceTopic)
			require.NoError(t, err)

			err = json.NewEncoder(w).Encode(rtceTopic)
			require.NoError(t, err)
		}
	}
}

// Handler for "/rtce/v1/rtce-topics/{topic_name}"
func handleRtceV1RtceTopicsTopicName(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		topic_name := mux.Vars(r)["topic_name"]
		switch r.Method {
		case http.MethodGet:
			switch topic_name {
			case "invalid":
				w.WriteHeader(http.StatusNotFound)
			default:
				rtceTopic := readRtceV1RtceTopicFile(t, "read_created_rtce_topic.json")

				err := json.NewEncoder(w).Encode(&rtceTopic)
				require.NoError(t, err)
			}
		case http.MethodPatch:
			switch topic_name {
			case "invalid":
				w.WriteHeader(http.StatusNotFound)
			default:
				rtceTopic := readRtceV1RtceTopicFile(t, "read_created_rtce_topic.json")

				// Overwrite updated fields using the request body
				err := json.NewDecoder(r.Body).Decode(&rtceTopic)
				require.NoError(t, err)

				err = json.NewEncoder(w).Encode(rtceTopic)
				require.NoError(t, err)
			}
		case http.MethodDelete:
			switch topic_name {
			case "invalid":
				w.WriteHeader(http.StatusNotFound)
			default:
				w.WriteHeader(http.StatusNoContent)
			}
		}
	}
}

func readRtceV1RtceTopicFile(t *testing.T, filename string) rtcev1.RtceV1RtceTopic {
	jsonPath := filepath.Join("..", "fixtures", "input", "rtce", "rtce_topic", filename)
	jsonData, err := os.ReadFile(jsonPath)
	require.NoError(t, err)

	rtceTopic := rtcev1.RtceV1RtceTopic{}
	err = json.Unmarshal(jsonData, &rtceTopic)
	require.NoError(t, err)

	return rtceTopic
}
