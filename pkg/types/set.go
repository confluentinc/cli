package types

import (
	"fmt"
)

type Set[V comparable] map[V]bool

func NewSet[V comparable](keys ...V) Set[V] {
	s := make(Set[V])
	for _, key := range keys {
		s.Add(key)
	}
	return s
}

func (s Set[V]) Contains(key V) bool {
	_, ok := s[key]
	return ok
}

func (s Set[V]) Add(key V) {
	s[key] = true
}

func (s Set[V]) Remove(key V) {
	delete(s, key)
}

func (s Set[V]) Slice() []V {
	return GetKeys(s)
}

func RemoveDuplicates(slice []string) []string {
	return NewSet(slice...).Slice()
}

func AddAndRemove(existing, add, remove []string) ([]string, []string) {
	var warnings []string

	s := NewSet(add...)

	for _, x := range remove {
		if s.Contains(x) {
			warnings = append(warnings, fmt.Sprintf(`"%s" is marked for addition and deletion`, x))
		}
	}

	for _, x := range existing {
		if s.Contains(x) {
			warnings = append(warnings, fmt.Sprintf(`"%s" is marked for addition but already exists`, x))
		} else {
			s.Add(x)
		}
	}

	for _, x := range remove {
		if s.Contains(x) {
			s.Remove(x)
		} else {
			warnings = append(warnings, fmt.Sprintf(`"%s" is marked for deletion but does not exist`, x))
		}
	}

	return s.Slice(), warnings
}
