package tableflow

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	tableflowv1 "github.com/confluentinc/ccloud-sdk-go-v2/tableflow/v1"
)

const variantFeature = "VARIANT"

// featuresOut is the "-o json" and "-o yaml" form of "config.features", keyed by feature name.
type featuresOut map[string]featureOut

type featureOut struct {
	Enabled          bool                `serialized:"enabled"`
	VariantShredding variantShreddingOut `serialized:"shredding"`
}

type variantShreddingOut struct {
	Enabled bool `serialized:"enabled"`
}

func addFeaturesFlags(cmd *cobra.Command) {
	cmd.Flags().StringSlice("features", nil, "A comma-separated list of table features, in the form NAME=BOOL, for example VARIANT=true.")
	cmd.Flags().Bool("variant-shredding", false, `Write VARIANT columns shredded. Requires "--features VARIANT=true". It cannot be disabled after it is enabled.`)
}

func featureEntries(cmd *cobra.Command) ([]tableflowv1.TableflowV1TableFlowTopicConfigsSpecFeaturesOneOf, error) {
	var features map[string]bool
	if cmd.Flags().Changed("features") {
		values, err := cmd.Flags().GetStringSlice("features")
		if err != nil {
			return nil, err
		}
		if features, err = parseFeatures(values); err != nil {
			return nil, err
		}
	}

	var variantShredding *bool
	if cmd.Flags().Changed("variant-shredding") {
		enabled, err := cmd.Flags().GetBool("variant-shredding")
		if err != nil {
			return nil, err
		}
		variantShredding = &enabled
	}

	if _, ok := features[variantFeature]; variantShredding != nil && !ok {
		return nil, fmt.Errorf("`--variant-shredding` requires \"%s\" in `--features`", variantFeature)
	}

	names := make([]string, 0, len(features))
	for name := range features {
		names = append(names, name)
	}
	sort.Strings(names)

	var entries []tableflowv1.TableflowV1TableFlowTopicConfigsSpecFeaturesOneOf
	for _, name := range names {
		if name == variantFeature {
			entries = append(entries, variantEntry(features[name], variantShredding))
		}
	}

	return entries, nil
}

// parseFeatures turns the values of --features, in the form NAME=BOOL, into a map keyed by the uppercase name.
func parseFeatures(values []string) (map[string]bool, error) {
	features := make(map[string]bool, len(values))

	for _, value := range values {
		value = strings.TrimSpace(value)

		name, enabledStr, found := strings.Cut(value, "=")
		if !found {
			return nil, fmt.Errorf("invalid value \"%s\" for `--features`: use NAME=BOOL, for example \"VARIANT=true\"", value)
		}
		name = strings.TrimSpace(name)

		enabled, err := strconv.ParseBool(strings.TrimSpace(enabledStr))
		if err != nil {
			return nil, fmt.Errorf("invalid value \"%s\" for \"%s\" in `--features`: use \"true\" or \"false\"", strings.TrimSpace(enabledStr), name)
		}

		if !strings.EqualFold(name, variantFeature) {
			return nil, fmt.Errorf("unknown feature \"%s\" in `--features`: allowed values are \"%s\"", name, variantFeature)
		}
		name = strings.ToUpper(name)

		if _, ok := features[name]; ok {
			return nil, fmt.Errorf("`--features` has more than one value for \"%s\"", name)
		}
		features[name] = enabled
	}

	return features, nil
}

// variantEntry returns the VARIANT entry of "config.features". A nil variantShredding leaves out "shredding".
func variantEntry(enabled bool, variantShredding *bool) tableflowv1.TableflowV1TableFlowTopicConfigsSpecFeaturesOneOf {
	feature := tableflowv1.NewTableflowV1VariantFeature(variantFeature, enabled)
	if variantShredding != nil {
		entry := tableflowv1.NewTableflowV1VariantShredding()
		entry.SetEnabled(*variantShredding)
		feature.SetShredding(*entry)
	}

	return tableflowv1.TableflowV1VariantFeatureAsTableflowV1TableFlowTopicConfigsSpecFeaturesOneOf(feature)
}

// getFeatures returns the table features in "config.features", keyed by feature name, in two forms:
// for human output, a value such as "{enabled=true, shredding=false}";
// for serialized output, a featureOut. Both maps are nil if the topic has no features.
// The API always returns "shredding.enabled" for VARIANT; a missing value reads as false.
func getFeatures(topic tableflowv1.TableflowV1TableflowTopic) (map[string]string, featuresOut) {
	var human map[string]string
	var serialized featuresOut

	config := topic.Spec.GetConfig()
	for _, entry := range config.GetFeatures() {
		variant := entry.TableflowV1VariantFeature
		if variant == nil {
			continue
		}

		variantShredding := variant.GetShredding()
		feature := featureOut{
			Enabled:          variant.GetEnabled(),
			VariantShredding: variantShreddingOut{Enabled: variantShredding.GetEnabled()},
		}

		if human == nil {
			human = make(map[string]string)
			serialized = make(featuresOut)
		}
		human[variant.GetFeature()] = fmt.Sprintf("{enabled=%t, shredding=%t}", feature.Enabled, feature.VariantShredding.Enabled)
		serialized[variant.GetFeature()] = feature
	}

	return human, serialized
}
