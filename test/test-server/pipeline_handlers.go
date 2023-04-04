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

var (
	createdAt = time.Date(2022, 10, 4, 6, 0, 0, 0, time.UTC)
	updatedAt = time.Date(2022, 10, 6, 6, 0, 0, 0, time.UTC)
	pipelines = []*streamdesignerv1.SdV1Pipeline{
		{
			Id: streamdesignerv1.PtrString("pipe-12345"),
			Spec: &streamdesignerv1.SdV1PipelineSpec{
				DisplayName:         streamdesignerv1.PtrString("testPipeline"),
				Description:         streamdesignerv1.PtrString("description"),
				SourceCode:          &streamdesignerv1.SdV1SourceCodeObject{Sql: "CREATE STREAM `upstream` (id INTEGER, name STRING) WITH (kafka_topic = 'topic', partitions=1, value_format='JSON');\n\nCREATE STREAM `downstream` AS SELECT * FROM upstream;"},
				Secrets:             &map[string]string{"name1": "*****************", "name2": "*****************"},
				KsqlCluster:         &streamdesignerv1.ObjectReference{Id: "lksqlc-12345"},
				ActivationPrivilege: streamdesignerv1.PtrBool(false),
			},
			Status: &streamdesignerv1.SdV1PipelineStatus{
				State: streamdesignerv1.PtrString("draft"),
			},
			Metadata: &streamdesignerv1.ObjectMeta{
				CreatedAt: &createdAt,
				UpdatedAt: &updatedAt,
			},
		},
		{
			Id: streamdesignerv1.PtrString("pipe-12346"),
			Spec: &streamdesignerv1.SdV1PipelineSpec{
				DisplayName:         streamdesignerv1.PtrString("testPipeline2"),
				Description:         streamdesignerv1.PtrString("description2"),
				SourceCode:          &streamdesignerv1.SdV1SourceCodeObject{Sql: "CREATE STREAM `upstream` (id INTEGER, name STRING) WITH (kafka_topic = 'topic', partitions=1, value_format='JSON');\n\nCREATE STREAM `downstream` AS SELECT * FROM upstream;"},
				Secrets:             &map[string]string{"name1": "*****************", "name2": "*****************"},
				KsqlCluster:         &streamdesignerv1.ObjectReference{Id: "lksqlc-54321"},
				ActivationPrivilege: streamdesignerv1.PtrBool(false),
			},

			Status: &streamdesignerv1.SdV1PipelineStatus{
				State: streamdesignerv1.PtrString("draft"),
			},

			Metadata: &streamdesignerv1.ObjectMeta{
				CreatedAt: &createdAt,
				UpdatedAt: &updatedAt,
			},
		},
	}
)

// Handler for: "/sd/v1/pipelines/{id}"
func handlePipeline(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pipelineId := mux.Vars(r)["id"]
		if i := getV2Index(pipelines, pipelineId); i != -1 {
			pipeline := pipelines[i]
			switch r.Method {
			case http.MethodDelete:
				w.WriteHeader(http.StatusAccepted)
			case http.MethodGet:
				w.WriteHeader(http.StatusAccepted)
				err := json.NewEncoder(w).Encode(pipeline)
				require.NoError(t, err)
			case http.MethodPatch:
				pipelinePatch := &streamdesignerv1.SdV1Pipeline{ // make a deep copy so changes don't reflect in subsequent tests
					Id:       streamdesignerv1.PtrString(pipeline.GetId()),
					Spec:     getV2Ptr(pipeline.GetSpec()),
					Status:   getV2Ptr(pipeline.GetStatus()),
					Metadata: getV2Ptr(pipeline.GetMetadata()),
				}
				var body streamdesignerv1.SdV1Pipeline
				err := json.NewDecoder(r.Body).Decode(&body)
				require.NoError(t, err)
				if body.Spec.DisplayName != nil {
					pipelinePatch.Spec.DisplayName = body.Spec.DisplayName
				}

				if body.Spec.Description != nil {
					pipelinePatch.Spec.Description = body.Spec.Description
				}

				if body.Spec.ActivationPrivilege != nil {
					pipelinePatch.Spec.ActivationPrivilege = body.Spec.ActivationPrivilege
				}

				state := "draft"
				if body.Spec.Activated != nil {
					if *body.Spec.Activated {
						state = "activating"
					} else {
						state = "deactivating"
					}
				}
				pipelinePatch.Status.State = &state

				if body.Spec.Secrets != nil {
					for name := range *body.Spec.Secrets {
						value := (*body.Spec.Secrets)[name]
						if len(value) > 0 {
							(*pipelinePatch.Spec.Secrets)[name] = "*****************"
						} else {
							delete(*pipelinePatch.Spec.Secrets, name)
						}
					}
				}

				w.WriteHeader(http.StatusAccepted)
				err = json.NewEncoder(w).Encode(pipelinePatch)
				require.NoError(t, err)
			}
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
	}
}

// Handler for: "/sd/v1/pipelines"
func handlePipelines(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			pipelineList := &streamdesignerv1.SdV1PipelineList{Data: getV2List(pipelines)}
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
					CreatedAt: &createdAt,
					UpdatedAt: &updatedAt,
				},
			}

			err = json.NewEncoder(w).Encode(pipeline)
			require.NoError(t, err)
		}
	}
}
