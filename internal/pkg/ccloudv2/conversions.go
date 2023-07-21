package ccloudv2

import (
	"github.com/confluentinc/cli/internal/pkg/errors"
)

const (
	ResourceNameNotFoundErrorMsg          = `resource "%s" was not found`
	DuplicateResourceNameErrorMsg         = `the resource name "%s" is shared across multiple resources`
	DuplicateResourceNameErrorSuggestions = "retry the previous command using a resource id"
)

/*
This code handles converting from a valid resource name to its ID
when a user chooses to use a resource name as an argument.

All resource object structs have a pointer receiver function GetId(),
so resourcePtr is the top level interface that all object structs implement.

Some resource structs have a *Spec field, which has a pointer receiver
function GetDisplayName(), whereas the remaining resource structs have the
pointer receiver GetDisplayName() function at the resource struct level.
Therefore, specPtr is an interface implemented only by the *Specs of the first
kind of resource and v2ResourcePtr is an interface implemented by the second
kind of resource.
*/

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

func ConvertToPtrSlice[R resource](resources []R) []*R {
	ptrs := make([]*R, len(resources))
	for i := range resources {
		ptrs[i] = &resources[i]
	}
	return ptrs
}

// ConvertSpecNameToId ConvertNamesToID returns a resource spec's name's corresponding ID or returns an error if not found
func ConvertSpecNameToId[R resourcePtr, S specPtr](input string, resources []R, specs []S) (string, error) {
	namesToIds, err := GetSpecNamesToIds(resources, specs)
	if err != nil {
		return input, err
	}
	if resourceIds, ok := namesToIds[input]; ok {
		if len(resourceIds) == 1 {
			return resourceIds[0], nil
		} else {
			return input, errors.NewErrorWithSuggestions(DuplicateResourceNameErrorMsg, DuplicateResourceNameErrorSuggestions)
		}
	} else {
		return input, errors.Errorf(ResourceNameNotFoundErrorMsg, input)
	}
}

// GetSpecNamesToIds returns a mapping from spec resource names to their respective IDs
func GetSpecNamesToIds[R resourcePtr, S specPtr](resources []R, specs []S) (map[string][]string, error) {
	namesToIds := make(map[string][]string, len(resources))
	for i := range resources {
		name := specs[i].GetDisplayName()
		if _, ok := namesToIds[name]; !ok {
			namesToIds[name] = []string{}
		}
		namesToIds[name] = append(namesToIds[name], resources[i].GetId())
	}
	return namesToIds, nil
}

// ConvertV2NameToId ConvertNamesToID returns a v2 resource name's corresponding ID or returns an error if not found
func ConvertV2NameToId[V v2ResourcePtr](input string, resources []V) (string, error) {
	namesToIds, err := GetV2NamesToIds(resources)
	if err != nil {
		return input, err
	}
	if resourceIds, ok := namesToIds[input]; ok {
		if len(resourceIds) == 1 {
			return resourceIds[0], nil
		} else {
			return input, errors.NewErrorWithSuggestions(DuplicateResourceNameErrorMsg, DuplicateResourceNameErrorSuggestions)
		}
	} else {
		return input, errors.Errorf(ResourceNameNotFoundErrorMsg, input)
	}
}

// GetV2NamesToIds returns a mapping from resource names to their respective IDs
func GetV2NamesToIds[V v2ResourcePtr](resources []V) (map[string][]string, error) {
	namesToIds := make(map[string][]string, len(resources))
	for _, res := range resources {
		name := res.GetDisplayName()
		if _, ok := namesToIds[name]; !ok {
			namesToIds[name] = []string{}
		}
		namesToIds[name] = append(namesToIds[name], res.GetId())
	}
	return namesToIds, nil
}
