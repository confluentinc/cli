package flink

import (
	"fmt"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

// cfkManagedByAnnotation marks a CMF resource as CFK-managed. CFK stamps it on every
// resource it creates (RFC 68); its value is the owning CR identity "<namespace>/<name>".
// Read-side clients block mutations based on the annotation's presence.
const cfkManagedByAnnotation = "cmf.platform.confluent.io/managed-by"

// errIfCfkManaged returns an error naming the owning custom resource if the resource is CFK-managed, else nil.
func errIfCfkManaged(resourceType, name string, annotations map[string]string) error {
	owner := annotations[cfkManagedByAnnotation]
	if owner == "" {
		return nil
	}

	errorMsg := fmt.Sprintf(`%s "%s" is managed by Confluent for Kubernetes (CFK) and cannot be modified through the CLI`, resourceType, name)
	suggestions := fmt.Sprintf("This resource is owned by the Kubernetes custom resource %q. Edit or delete it there, for example with `kubectl edit` or `kubectl delete`.", owner)

	return errors.NewErrorWithSuggestions(errorMsg, suggestions)
}

// flinkApplicationAnnotations reads string-valued metadata.annotations from a FlinkApplication
// (whose metadata is untyped), returning nil when none are present.
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
	if len(annotations) == 0 {
		return nil
	}
	return annotations
}
