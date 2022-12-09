package testserver

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	streamdesignerv1 "github.com/confluentinc/ccloud-sdk-go-v2/stream-designer/v1"
	"github.com/stretchr/testify/require"
)

// Handler for: "/sd/v1/pipelines/{id}"
func handlePipeline(t *testing.T) http.HandlerFunc {
	CreatedAt := time.Date(2022, 10, 4, 06, 00, 00, 000000000, time.UTC)
	UpdatedAt := time.Date(2022, 10, 6, 06, 00, 00, 000000000, time.UTC)
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			w.WriteHeader(202)

		case http.MethodGet:
			pipeline := &streamdesignerv1.SdV1Pipeline{
				Id: streamdesignerv1.PtrString("pipe-12345"),
				Spec: &streamdesignerv1.SdV1PipelineSpec{
					DisplayName: streamdesignerv1.PtrString("testPipeline"),
					Description: streamdesignerv1.PtrString("description"),
					KsqlCluster: &streamdesignerv1.ObjectReference{Id: "lksqlc-12345"},
				},

				Status: &streamdesignerv1.SdV1PipelineStatus{
					State: streamdesignerv1.PtrString("draft"),
				},

				Metadata: &streamdesignerv1.ObjectMeta{
					CreatedAt: &CreatedAt,
					UpdatedAt: &UpdatedAt,
				},
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(202)
			err := json.NewEncoder(w).Encode(pipeline)
			require.NoError(t, err)

		case http.MethodPatch:
			var body streamdesignerv1.SdV1Pipeline
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
				Id: &id,
				Spec: &streamdesignerv1.SdV1PipelineSpec{
					DisplayName: &name,
					Description: streamdesignerv1.PtrString("description"),
					KsqlCluster: &streamdesignerv1.ObjectReference{Id: "lksqlc-12345"},
				},

				Status: &streamdesignerv1.SdV1PipelineStatus{
					State: &state,
				},

				Metadata: &streamdesignerv1.ObjectMeta{
					CreatedAt: &CreatedAt,
					UpdatedAt: &UpdatedAt,
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
	CreatedAt := time.Date(2022, 10, 4, 06, 00, 00, 000000000, time.UTC)
	UpdatedAt := time.Date(2022, 10, 6, 06, 00, 00, 000000000, time.UTC)

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			pipeline := &streamdesignerv1.SdV1Pipeline{
				Id: streamdesignerv1.PtrString("pipe-12345"),
				Spec: &streamdesignerv1.SdV1PipelineSpec{
					DisplayName: streamdesignerv1.PtrString("testPipeline"),
					Description: streamdesignerv1.PtrString("description"),
					KsqlCluster: &streamdesignerv1.ObjectReference{Id: "lksqlc-12345"},
				},

				Status: &streamdesignerv1.SdV1PipelineStatus{
					State: streamdesignerv1.PtrString("draft"),
				},

				Metadata: &streamdesignerv1.ObjectMeta{
					CreatedAt: &CreatedAt,
					UpdatedAt: &UpdatedAt,
				},
			}
			pipeline2 := &streamdesignerv1.SdV1Pipeline{
				Id: streamdesignerv1.PtrString("pipe-12346"),
				Spec: &streamdesignerv1.SdV1PipelineSpec{
					DisplayName: streamdesignerv1.PtrString("testPipeline2"),
					Description: streamdesignerv1.PtrString("description2"),
					KsqlCluster: &streamdesignerv1.ObjectReference{Id: "lksqlc-54321"},
				},

				Status: &streamdesignerv1.SdV1PipelineStatus{
					State: streamdesignerv1.PtrString("draft"),
				},

				Metadata: &streamdesignerv1.ObjectMeta{
					CreatedAt: &CreatedAt,
					UpdatedAt: &UpdatedAt,
				},
			}
			pipelineList := streamdesignerv1.SdV1PipelineList{Data: []streamdesignerv1.SdV1Pipeline{*pipeline, *pipeline2}}
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(pipelineList)
			require.NoError(t, err)

		case http.MethodPost:
			var body streamdesignerv1.SdV1Pipeline
			err := json.NewDecoder(r.Body).Decode(&body)
			require.NoError(t, err)

			pipeline := &streamdesignerv1.SdV1Pipeline{
				Id: streamdesignerv1.PtrString("pipe-12345"),
				Spec: &streamdesignerv1.SdV1PipelineSpec{
					DisplayName: body.Spec.DisplayName,
					Description: streamdesignerv1.PtrString("description"),
					SourceCode:  body.Spec.SourceCode,
					Secrets:     body.Spec.Secrets,
					KsqlCluster: &streamdesignerv1.ObjectReference{Id: "lksqlc-12345"},
				},

				Status: &streamdesignerv1.SdV1PipelineStatus{
					State: streamdesignerv1.PtrString("draft"),
				},

				Metadata: &streamdesignerv1.ObjectMeta{
					CreatedAt: &CreatedAt,
					UpdatedAt: &UpdatedAt,
				},
			}

			w.Header().Set("Content-Type", "application/json")
			err = json.NewEncoder(w).Encode(pipeline)
			require.NoError(t, err)
		}
	}
}
