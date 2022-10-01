package testserver

import (
	"encoding/json"
	"net/http"
	"testing"

	streamdesignerv1 "github.com/confluentinc/ccloud-sdk-go-v2/stream-designer/v1"
	"github.com/stretchr/testify/require"
)

type PipelineRequestBody struct {
	Spec struct {
		Activated   *bool   `json:"activated,omitempty"`
		Description *string `json:"description,omitempty"`
		DisplayName *string `json:"display_name,omitempty"`
		Environment struct {
			ID           *string `json:"id,omitempty"`
			Related      *string `json:"related,omitempty"`
			ResourceName *string `json:"resource_name,omitempty"`
		} `json:"environment,omitempty"`
		KafkaCluster struct {
			ID           *string `json:"id,omitempty"`
			Related      *string `json:"related,omitempty"`
			ResourceName *string `json:"resource_name,omitempty"`
		} `json:"kafka_cluster,omitempty"`
	} `json:"spec,omitempty"`
}

// Handler for: "/sd/v1/pipelines/{id}"
func handlePipeline(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			w.WriteHeader(202)

		case http.MethodPatch:
			var body PipelineRequestBody
			err := json.NewDecoder(r.Body).Decode(&body)
			require.NoError(t, err)

			id := "pipe-12345"
			name := "testPipeline"
			if body.Spec.DisplayName != nil {
				name = *body.Spec.DisplayName
			}

			state := "draft"
			if body.Spec.Activated != nil {
				if *body.Spec.Activated {
					state = "activating"
				} else {
					state = "deactivating"
				}
			}

			pipeline := &streamdesignerv1.SdV1Pipeline{
				Id:   &id,
				Spec: &streamdesignerv1.SdV1PipelineSpec{DisplayName: &name},

				Status: &streamdesignerv1.SdV1PipelineStatus{
					State: &state,
				},
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(202)
			err = json.NewEncoder(w).Encode(pipeline)
			require.NoError(t, err)
		}
	}

}

// Handler for: "/sd/v1/pipelines"
func handlePipelines(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			id := "pipe-12345"
			name := "testPipeline"
			state := "draft"
			pipeline := &streamdesignerv1.SdV1Pipeline{
				Id:   &id,
				Spec: &streamdesignerv1.SdV1PipelineSpec{DisplayName: &name},

				Status: &streamdesignerv1.SdV1PipelineStatus{
					State: &state,
				},
			}
			id2 := "pipe-12346"
			name2 := "testPipeline2"
			pipeline2 := &streamdesignerv1.SdV1Pipeline{
				Id:   &id2,
				Spec: &streamdesignerv1.SdV1PipelineSpec{DisplayName: &name2},

				Status: &streamdesignerv1.SdV1PipelineStatus{
					State: &state,
				},
			}
			pipelineList := streamdesignerv1.SdV1PipelineList{Data: []streamdesignerv1.SdV1Pipeline{*pipeline, *pipeline2}}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(pipelineList)
			require.NoError(t, err)

		case http.MethodPost:
			var body PipelineRequestBody
			err := json.NewDecoder(r.Body).Decode(&body)
			require.NoError(t, err)

			id := "pipe-12345"
			state := "draft"
			pipeline := &streamdesignerv1.SdV1Pipeline{
				Id:   &id,
				Spec: &streamdesignerv1.SdV1PipelineSpec{DisplayName: body.Spec.DisplayName},

				Status: &streamdesignerv1.SdV1PipelineStatus{
					State: &state,
				},
			}

			w.Header().Set("Content-Type", "application/json")
			err = json.NewEncoder(w).Encode(pipeline)
			require.NoError(t, err)
		}
	}
}
