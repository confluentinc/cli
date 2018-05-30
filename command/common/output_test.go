package common

import (
	"bytes"
	"testing"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	connectv1 "github.com/confluentinc/cli/shared/connect"
)

func TestRender(t *testing.T) {
	testStruct := &struct {
		Foo string
		Baz int
		Bat float32
	} {
		Foo: "bar",
		Baz: 23,
		Bat: 1.23,
	}
	testStructFields := []string{"Foo", "Baz"}
	testStructJSONTags := &struct {
		Foo string  `json:"foo"`
		Baz int     `json:"baz"`
		Bat float32 `json:"bat"`
	} {
		Foo: "bar",
		Baz: 23,
		Bat: 1.23,
	}
	testFieldsJSONTags := []string{"foo", "baz"}
	testFieldsJSONTagsRename := []string{"better_foo", "clearer_baz"}
	testMap := map[string]interface{}{
		"foo": "bar",
		"baz": 23,
		"bat": 1.23,
	}
	//testMapFields := []string{"foo", "baz"}
	testProto := &connectv1.ConnectS3SinkClusterConfig{
		Name: "connect",
		AccountId: "23",
		KafkaClusterName: "s3-logs",
		KafkaUserEmail: "cody@confluent.io",
		Servers: 2,
		Options: &schedv1.ConnectS3SinkOptions{
			SourceTopics: "foo,bar.*",
		},
	}
	testProtoFields := []string{"name", "account_id", "kafka_cluster_name"}
	testProtoFieldsRename:= []string{"Name", "Account", "Kafka"}

	type args struct {
		obj          interface{}
		fields       []string
		labels       []string
		outputFormat string
	}
	tests := []struct {
		name    string
		args    args
		wantOut string
		wantErr bool
	}{
		{
			name: "struct to json",
			args: args{
				obj:          testStruct,
				fields:       nil,
				labels:       nil,
				outputFormat: "json",
			},
			wantOut: `{
  "Foo": "bar",
  "Baz": 23,
  "Bat": 1.23
}
`,
			wantErr: false,
		},
		{
			name: "struct to json, with json tags",
			args: args{
				obj:          testStructJSONTags,
				fields:       nil,
				labels:       nil,
				outputFormat: "json",
			},
			wantOut: `{
  "foo": "bar",
  "baz": 23,
  "bat": 1.23
}
`,
			wantErr: false,
		},
		{
			name: "struct to json, with filtering",
			args: args{
				obj:          testStruct,
				fields:       testStructFields,
				labels:       testStructFields,
				outputFormat: "json",
			},
			wantOut: `{
  "Foo": "bar",
  "Baz": 23
}
`,
			wantErr: false,
		},
		{
			name: "struct to json, with json tags, with filtering",
			args: args{
				obj:          testStructJSONTags,
				fields:       testFieldsJSONTags,
				labels:       testFieldsJSONTags,
				outputFormat: "json",
			},
			wantOut: `{
  "foo": "bar",
  "baz": 23
}
`,
			wantErr: false,
		},
		{
			name: "struct to json, with json tags, with filtering, with renaming",
			args: args{
				obj:          testStructJSONTags,
				fields:       testFieldsJSONTags,
				labels:       testFieldsJSONTagsRename,
				outputFormat: "json",
			},
			wantOut: `{
  "better_foo": "bar",
  "clearer_baz": 23
}
`,
			wantErr: false,
		},
		{
			name: "map to json",
			args: args{
				obj:          testMap,
				fields:       nil,
				labels:       nil,
				outputFormat: "json",
			},
			wantOut: `{
  "bat": 1.23,
  "baz": 23,
  "foo": "bar"
}
`,
			wantErr: false,
		},
		//{
		//	name: "map to json, with filtering",
		//	args: args{
		//		obj:          testMap,
		//		fields:       testMapFields,
		//		labels:       testMapFields,
		//		outputFormat: "json",
		//	},
		//	wantErr: true, // TODO: return error instead of panic here
		//},
		{
			name: "protobuf to json",
			args: args{
				obj:          testProto,
				fields:       nil,
				labels:       nil,
				outputFormat: "json",
			},
			wantOut: `{
  "name": "connect",
  "account_id": "23",
  "kafka_cluster_name": "s3-logs",
  "kafka_user_email": "cody@confluent.io",
  "servers": 2,
  "options": {
    "source_topics": "foo,bar.*"
  }
}
`,
			wantErr: false,
		},
		{
			name: "protobuf to json, with filtering",
			args: args{
				obj:          testProto,
				fields:       testProtoFields,
				labels:       testProtoFields,
				outputFormat: "json",
			},
			wantOut: `{
  "name": "connect",
  "account_id": "23",
  "kafka_cluster_name": "s3-logs"
}
`,
			wantErr: false,
		},
		{
			name: "protobuf to json, with filtering, with renaming",
			args: args{
				obj:          testProto,
				fields:       testProtoFields,
				labels:       testProtoFieldsRename,
				outputFormat: "json",
			},
			wantOut: `{
  "Name": "connect",
  "Account": "23",
  "Kafka": "s3-logs"
}
`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			if err := renderOut(tt.args.obj, tt.args.fields, tt.args.labels, tt.args.outputFormat, out); (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOut := out.String(); gotOut != tt.wantOut {
				t.Errorf("Render() = %#v, want %#v", gotOut, tt.wantOut)
			}
		})
	}
}
