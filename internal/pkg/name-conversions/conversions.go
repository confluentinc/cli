package name_conversions

import (
	"fmt"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

const (
	DuplicateResourceNameErrorMsg         = `the resource name "%s" is shared across multiple resources`
	DuplicateResourceNameErrorSuggestions = "retry the previous command using a resource id"
)

type resource any

type resourcePtr interface {
	GetId() string
}

type specPtr interface {
	GetDisplayName() string
}

type v2ResourcePtr interface {
	GetDisplayName() string
	resourcePtr
}

func ConvertToPtrSlice[V resource](resources []V) []*V {
	ptrs := make([]*V, len(resources))
	for i := range resources {
		ptrs[i] = &resources[i]
	}
	return ptrs
}

// ConvertSpecNameToId ConvertNamesToID returns a resource spec's name's corresponding ID or returns the input string if not found
func ConvertSpecNameToId[V resourcePtr, T specPtr](input string, resources []V, specs []T) (string, error) {
	namesToIds, err := GetSpecNamesToIds(resources, specs)
	if err != nil {
		return input, err
	}
	if resourceId, ok := namesToIds[input]; ok {
		return resourceId, nil
	} else {
		return input, nil
	}
}

// GetSpecNamesToIds returns a mapping from spec resource names to their respective IDs
func GetSpecNamesToIds[V resourcePtr, T specPtr](resources []V, specs []T) (map[string]string, error) {
	namesToIds := make(map[string]string, len(resources))
	for i := range resources {
		name := specs[i].GetDisplayName()
		if _, ok := namesToIds[name]; !ok {
			namesToIds[name] = resources[i].GetId()
		} else {
			return nil, errors.NewErrorWithSuggestions(fmt.Sprintf(DuplicateResourceNameErrorMsg, name), DuplicateResourceNameErrorSuggestions)
		}
	}
	return namesToIds, nil
}

// ConvertV2NameToId ConvertNamesToID returns a v2 resource name's corresponding ID or returns the input string if not found
func ConvertV2NameToId[V v2ResourcePtr](input string, resources []V) (string, error) {
	namesToIds, err := GetV2NamesToIds(resources)
	if err != nil {
		return input, err
	}
	if resourceId, ok := namesToIds[input]; ok {
		return resourceId, nil
	} else {
		return input, nil
	}
}

// GetV2NamesToIds returns a mapping from resource names to their respective IDs
func GetV2NamesToIds[V v2ResourcePtr](resources []V) (map[string]string, error) {
	namesToIDs := make(map[string]string, len(resources))
	for _, res := range resources {
		name := res.GetDisplayName()
		if _, ok := namesToIDs[name]; !ok {
			namesToIDs[name] = res.GetId()
		} else {
			return nil, errors.NewErrorWithSuggestions(DuplicateResourceNameErrorMsg, DuplicateResourceNameErrorSuggestions)
		}
	}
	return namesToIDs, nil
}
