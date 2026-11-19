package tableflow

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	tableflowv1 "github.com/confluentinc/ccloud-sdk-go-v2/tableflow/v1"
)

func TestParseFeatures(t *testing.T) {
	tests := []struct {
		name    string
		values  []string
		want    map[string]bool
		wantErr string
	}{
		{name: "missing value", values: []string{"VARIANT"}, wantErr: "invalid value \"VARIANT\" for `--features`: use NAME=BOOL, for example \"VARIANT=true\""},
		{name: "not a bool", values: []string{"VARIANT=maybe"}, wantErr: "invalid value \"maybe\" for \"VARIANT\" in `--features`: use \"true\" or \"false\""},
		{name: "unknown feature", values: []string{"OTHER=true"}, wantErr: "unknown feature \"OTHER\" in `--features`: allowed values are \"VARIANT\""},
		{name: "repeated feature", values: []string{"VARIANT=true", "VARIANT=false"}, wantErr: "`--features` has more than one value for \"VARIANT\""},
		{name: "repeated feature in another case", values: []string{"VARIANT=true", "variant=false"}, wantErr: "`--features` has more than one value for \"VARIANT\""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseFeatures(test.values)
			if test.wantErr != "" {
				require.EqualError(t, err, test.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.want, got)
		})
	}
}

func newFeaturesFlagsTestCommand(t *testing.T, args ...string) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{Use: "test"}
	addFeaturesFlags(cmd)
	require.NoError(t, cmd.ParseFlags(args))
	return cmd
}

func TestFeatureEntries(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr string
	}{
		{name: "no flags", args: nil, want: "null"},
		{name: "variant with shredding", args: []string{"--features", "VARIANT=true", "--variant-shredding=true"}, want: `[{"feature":"VARIANT","enabled":true,"shredding":{"enabled":true}}]`},
		{name: "variant without shredding", args: []string{"--features", "VARIANT=true"}, want: `[{"feature":"VARIANT","enabled":true}]`},
		{name: "variant with shredding off", args: []string{"--features", "VARIANT=true", "--variant-shredding=false"}, want: `[{"feature":"VARIANT","enabled":true,"shredding":{"enabled":false}}]`},
		{name: "shredding without features", args: []string{"--variant-shredding=true"}, wantErr: "`--variant-shredding` requires \"VARIANT\" in `--features`"},
		{name: "shredding off without features", args: []string{"--variant-shredding=false"}, wantErr: "`--variant-shredding` requires \"VARIANT\" in `--features`"},
		{name: "shredding on a disabled variant is left to the API", args: []string{"--features", "VARIANT=false", "--variant-shredding"}, want: `[{"feature":"VARIANT","enabled":false,"shredding":{"enabled":true}}]`},
		{name: "lowercase variant with shredding", args: []string{"--features", "variant=true", "--variant-shredding"}, want: `[{"feature":"VARIANT","enabled":true,"shredding":{"enabled":true}}]`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entries, err := featureEntries(newFeaturesFlagsTestCommand(t, test.args...))
			if test.wantErr != "" {
				require.EqualError(t, err, test.wantErr)
				return
			}
			require.NoError(t, err)
			body, err := json.Marshal(entries)
			require.NoError(t, err)
			require.JSONEq(t, test.want, string(body))
		})
	}
}

func topicWithFeatures(features ...tableflowv1.TableflowV1VariantFeature) tableflowv1.TableflowV1TableflowTopic {
	config := tableflowv1.TableflowV1TableFlowTopicConfigsSpec{}
	entries := make([]tableflowv1.TableflowV1TableFlowTopicConfigsSpecFeaturesOneOf, len(features))
	for i := range features {
		entries[i] = tableflowv1.TableflowV1VariantFeatureAsTableflowV1TableFlowTopicConfigsSpecFeaturesOneOf(&features[i])
	}
	config.Features = &entries
	return tableflowv1.TableflowV1TableflowTopic{Spec: &tableflowv1.TableflowV1TableflowTopicSpec{Config: &config}}
}

func TestGetFeatures(t *testing.T) {
	tests := []struct {
		name           string
		topic          tableflowv1.TableflowV1TableflowTopic
		wantHuman      map[string]string
		wantSerialized featuresOut
	}{
		{
			name:           "no shredding object",
			topic:          topicWithFeatures(tableflowv1.TableflowV1VariantFeature{Feature: "VARIANT", Enabled: true}),
			wantHuman:      map[string]string{"VARIANT": "{enabled=true, shredding=false}"},
			wantSerialized: featuresOut{"VARIANT": {Enabled: true, VariantShredding: variantShreddingOut{Enabled: false}}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			human, serialized := getFeatures(test.topic)
			require.Equal(t, test.wantHuman, human)
			require.Equal(t, test.wantSerialized, serialized)
		})
	}
}
