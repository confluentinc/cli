package testserver

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	streamdesignerv1 "github.com/confluentinc/ccloud-sdk-go-v2/stream-designer/v1"
)

// Handler for: "/sd/v1/pipelines/{id}"
func handlePipeline(t *testing.T) http.HandlerFunc {
	CreatedAt := time.Date(2022, 10, 4, 6, 0, 0, 0, time.UTC)
	UpdatedAt := time.Date(2022, 10, 6, 6, 0, 0, 0, time.UTC)
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		if id != "pipe-12345" && id != "pipe-54321" {
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		switch r.Method {
		case http.MethodGet:
			pipeline := &streamdesignerv1.SdV1Pipeline{
				Id: streamdesignerv1.PtrString(id),
				Spec: &streamdesignerv1.SdV1PipelineSpec{
					DisplayName:         streamdesignerv1.PtrString("testPipeline"),
					Description:         streamdesignerv1.PtrString("description"),
					SourceCode:          &streamdesignerv1.SdV1SourceCodeObject{Sql: "CREATE STREAM `upstream` (id INTEGER, name STRING) WITH (kafka_topic = 'topic', partitions=1, value_format='JSON');\n\nCREATE STREAM `downstream` AS SELECT * FROM upstream;"},
					Secrets:             &map[string]string{"name1": "secret1", "name2": "secret2"},
					KsqlCluster:         &streamdesignerv1.ObjectReference{Id: "lksqlc-12345"},
					ActivationPrivilege: streamdesignerv1.PtrBool(false),
				},

				Status: &streamdesignerv1.SdV1PipelineStatus{
					State: streamdesignerv1.PtrString("draft"),
				},

				Metadata: &streamdesignerv1.ObjectMeta{
					CreatedAt: &CreatedAt,
					UpdatedAt: &UpdatedAt,
				},
			}

			// if request to use SR cluster
			if id == "pipe-11001" {
				pipeline.Spec.StreamGovernanceCluster = &streamdesignerv1.ObjectReference{Id: "lsrc-12345"}
			}

			err := json.NewEncoder(w).Encode(pipeline)
			require.NoError(t, err)
		case http.MethodPatch:
			var body streamdesignerv1.SdV1Pipeline
			err := json.NewDecoder(r.Body).Decode(&body)
			require.NoError(t, err)

			pipeline := &streamdesignerv1.SdV1Pipeline{
				Id: streamdesignerv1.PtrString(id),
				Spec: &streamdesignerv1.SdV1PipelineSpec{
					DisplayName:         streamdesignerv1.PtrString("testPipeline"),
					Description:         streamdesignerv1.PtrString("description"),
					SourceCode:          &streamdesignerv1.SdV1SourceCodeObject{Sql: "CREATE STREAM `upstream` (id INTEGER, name STRING) WITH (kafka_topic = 'topic', partitions=1, value_format='JSON');\n\nCREATE STREAM `downstream` AS SELECT * FROM upstream;"},
					Secrets:             &map[string]string{"name1": "*****************", "name2": "*****************", "name3": "*****************"},
					KsqlCluster:         &streamdesignerv1.ObjectReference{Id: "lksqlc-12345"},
					ActivationPrivilege: streamdesignerv1.PtrBool(false),
				},

				Status: &streamdesignerv1.SdV1PipelineStatus{
					State: streamdesignerv1.PtrString("draft"),
				},

				Metadata: &streamdesignerv1.ObjectMeta{
					CreatedAt: &CreatedAt,
					UpdatedAt: &UpdatedAt,
				},
			}

			if body.Spec.DisplayName != nil {
				pipeline.Spec.DisplayName = body.Spec.DisplayName
			}

			if body.Spec.Description != nil {
				pipeline.Spec.Description = body.Spec.Description
			}

			if body.Spec.ActivationPrivilege != nil {
				pipeline.Spec.ActivationPrivilege = body.Spec.ActivationPrivilege
			}

			state := "draft"
			if body.Spec.Activated != nil {
				if *body.Spec.Activated {
					state = "activating"
				} else {
					state = "deactivating"
				}
			}
			pipeline.Status.State = &state

			if body.Spec.Secrets != nil {
				for name := range *body.Spec.Secrets {
					value := (*body.Spec.Secrets)[name]
					if len(value) > 0 {
						(*pipeline.Spec.Secrets)[name] = "*****************"
					} else {
						// for PATCH operation, empty secret value will be removed
						delete(*pipeline.Spec.Secrets, name)
					}
				}
			}
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
			err := json.NewEncoder(w).Encode(pipelineList)
			require.NoError(t, err)

		case http.MethodPost:
			var body streamdesignerv1.SdV1Pipeline
			err := json.NewDecoder(r.Body).Decode(&body)
			require.NoError(t, err)

			pipeline := &streamdesignerv1.SdV1Pipeline{
				Id: streamdesignerv1.PtrString("pipe-12345"),
				Spec: &streamdesignerv1.SdV1PipelineSpec{
					DisplayName:         body.Spec.DisplayName,
					Description:         streamdesignerv1.PtrString("description"),
					SourceCode:          body.Spec.SourceCode,
					Secrets:             body.Spec.Secrets,
					KsqlCluster:         &streamdesignerv1.ObjectReference{Id: "lksqlc-12345"},
					ActivationPrivilege: streamdesignerv1.PtrBool(false),
				},

				Status: &streamdesignerv1.SdV1PipelineStatus{
					State: streamdesignerv1.PtrString("draft"),
				},

				Metadata: &streamdesignerv1.ObjectMeta{
					CreatedAt: &CreatedAt,
					UpdatedAt: &UpdatedAt,
				},
			}

			// if request to use SR cluster
			if body.Spec.StreamGovernanceCluster != nil {
				pipeline.Spec.StreamGovernanceCluster = &streamdesignerv1.ObjectReference{Id: "lsrc-12345"}
				pipeline.Id = streamdesignerv1.PtrString("pipe-11001")
			}

			err = json.NewEncoder(w).Encode(pipeline)
			require.NoError(t, err)
		}
	}
}
