package testserver

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	streamdesignerv1 "github.com/confluentinc/ccloud-sdk-go-v2/stream-designer/v1"
)

// Handler for: "/sd/v1/pipelines/{id}"
func handlePipeline(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(202)
		}
		if r.Method == http.MethodPatch {
			b, _ := io.ReadAll(r.Body)
			body := make(map[string]interface{})
			json.Unmarshal([]byte(b), &body)
			spec := body["spec"].(map[string]interface{})

			pipeline := streamdesignerv1.NewSdV1Pipeline()
			pipeline.SetId("pipe-12345")

			pipelineSpec := streamdesignerv1.NewSdV1PipelineSpec()
			name, found := spec["display_name"]
			if found {
				pipelineSpec.SetDisplayName(name.(string))
			} else {
				pipelineSpec.SetDisplayName("testPipeline")
			}

			pipelineStatus := streamdesignerv1.NewSdV1PipelineStatus()
			activated, found := spec["activated"]
			if found {
				if activated.(bool) {
					pipelineStatus.SetState("activating")
				} else {
					pipelineStatus.SetState("deactivating")
				}
			} else {
				pipelineStatus.SetState("draft")
			}
			pipeline.SetSpec(*pipelineSpec)
			pipeline.SetStatus(*pipelineStatus)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(202)
			err := json.NewEncoder(w).Encode(pipeline)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/sd/v1/pipelines"
func handlePipelines(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			pipeline := streamdesignerv1.NewSdV1Pipeline()
			pipeline.SetId("pipe-12345")
			pipelineSpec := streamdesignerv1.NewSdV1PipelineSpec()
			pipelineSpec.SetDisplayName("testPipeline")
			pipelineStatus := streamdesignerv1.NewSdV1PipelineStatus()
			pipelineStatus.SetState("draft")
			pipeline.SetSpec(*pipelineSpec)
			pipeline.SetStatus(*pipelineStatus)

			pipeline2 := streamdesignerv1.NewSdV1Pipeline()
			pipeline2.SetId("pipe-12346")
			pipelineSpec2 := streamdesignerv1.NewSdV1PipelineSpec()
			pipelineSpec2.SetDisplayName("testPipeline2")
			pipelineStatus2 := streamdesignerv1.NewSdV1PipelineStatus()
			pipelineStatus2.SetState("draft")
			pipeline2.SetSpec(*pipelineSpec2)
			pipeline2.SetStatus(*pipelineStatus2)
			pipelineList := streamdesignerv1.NewSdV1PipelineListWithDefaults()

			pipelineList.SetData([]streamdesignerv1.SdV1Pipeline{*pipeline, *pipeline2})

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(pipelineList)
			require.NoError(t, err)
		}
		if r.Method == http.MethodPost {
			b, _ := io.ReadAll(r.Body)
			body := make(map[string]interface{})
			json.Unmarshal([]byte(b), &body)
			spec := body["spec"].(map[string]interface{})

			pipeline := streamdesignerv1.NewSdV1Pipeline()
			pipeline.SetId("pipe-12345")
			pipelineSpec := streamdesignerv1.NewSdV1PipelineSpec()
			pipelineSpec.SetDisplayName(spec["display_name"].(string))
			pipelineStatus := streamdesignerv1.NewSdV1PipelineStatus()
			pipelineStatus.SetState("draft")
			pipeline.SetSpec(*pipelineSpec)
			pipeline.SetStatus(*pipelineStatus)

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(pipeline)
			require.NoError(t, err)
		}
	}
}
