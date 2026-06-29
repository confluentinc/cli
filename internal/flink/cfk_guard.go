package flink

import (
	"fmt"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

// CFK stamps these ownership annotations on every CMF resource it creates (RFC 68).
const (
	cfkManagedByAnnotation          = "cmf.platform.confluent.io/managed-by"
	cfkManagedByValue               = "confluent-operator"
	cfkManagedByNamespaceAnnotation = "cmf.platform.confluent.io/managed-by-namespace"
	cfkManagedByNameAnnotation      = "cmf.platform.confluent.io/managed-by-name"
)

// errIfCfkManaged returns an error naming the owning custom resource if the annotations mark it CFK-owned, else nil.
func errIfCfkManaged(resourceType, name string, annotations map[string]string) error {
	if annotations[cfkManagedByAnnotation] != cfkManagedByValue {
		return nil
	}

	errorMsg := fmt.Sprintf(`%s "%s" is managed by Confluent for Kubernetes (CFK) and cannot be modified through the CLI`, resourceType, name)

	suggestions := "This resource was created by CFK. Edit or delete it through the owning Kubernetes custom resource"
	if owner := cfkOwnerReference(annotations); owner != "" {
		suggestions += fmt.Sprintf(" %s", owner)
	}
	suggestions += ", for example with `kubectl edit` or `kubectl delete`."

	return errors.NewErrorWithSuggestions(errorMsg, suggestions)
}

// cfkOwnerReference renders the owning custom resource as `"namespace/name"` (or `"name"`, or "").
func cfkOwnerReference(annotations map[string]string) string {
	namespace := annotations[cfkManagedByNamespaceAnnotation]
	name := annotations[cfkManagedByNameAnnotation]
	switch {
	case namespace != "" && name != "":
		return fmt.Sprintf(`"%s/%s"`, namespace, name)
	case name != "":
		return fmt.Sprintf(`"%s"`, name)
	default:
		return ""
	}
}

// flinkApplicationAnnotations reads metadata.annotations from a FlinkApplication, whose metadata is untyped.
func flinkApplicationAnnotations(application cmfsdk.FlinkApplication) map[string]string {
	rawAnnotations, ok := application.GetMetadata()["annotations"].(map[string]interface{})
	if !ok {
		return nil
	}

	annotations := make(map[string]string, len(rawAnnotations))
	for key, value := range rawAnnotations {
		if stringValue, ok := value.(string); ok {
			annotations[key] = stringValue
		}
	}
	return annotations
}
