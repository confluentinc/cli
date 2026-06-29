package flink

import (
	"fmt"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

// CFK (Confluent for Kubernetes) stamps every CMF resource it creates with these
// ownership annotations (RFC 68). Their presence is the read-side signal that a
// resource must be managed through its Kubernetes custom resource rather than the
// CLI. The CLI blocks mutations on these resources until out-of-band edit
// detection ships; reads remain unrestricted.
const (
	cfkManagedByAnnotation          = "cmf.platform.confluent.io/managed-by"
	cfkManagedByValue               = "confluent-operator"
	cfkManagedByNamespaceAnnotation = "cmf.platform.confluent.io/managed-by-namespace"
	cfkManagedByNameAnnotation      = "cmf.platform.confluent.io/managed-by-name"
)

// errIfCfkManaged returns a user-facing error when the given annotations mark the
// resource as owned by CFK, and nil otherwise. resourceType is a display label
// (e.g. resource.FlinkStatement) and name is the resource name; both appear in the
// message so the user can act without guessing.
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

// cfkOwnerReference renders the owning custom resource as `"namespace/name"` from
// the CFK ownership annotations, falling back to `"name"` or "" as the annotations
// allow.
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

// flinkApplicationAnnotations extracts metadata.annotations from a FlinkApplication.
// Unlike the other CMF resources, FlinkApplication exposes its metadata as an
// untyped map, so annotations are read with type assertions. Returns nil when no
// string-valued annotations are present.
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
