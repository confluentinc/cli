package types

import (
	"sort"

	"golang.org/x/exp/constraints"
)

func GetKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, len(m))
	i := 0
	for key := range m {
		keys[i] = key
		i++
	}
	return keys
}

func GetSortedKeys[K constraints.Ordered, V any](m map[K]V) []K {
	keys := NewSortableSlice[K](len(m))
	i := 0
	for key := range m {
		keys[i] = key
		i++
	}
	sort.Sort(keys)
	return keys
}

func GetSortedValues[K comparable, V constraints.Ordered](m map[K]V) []V {
	values := NewSortableSlice[V](len(m))
	i := 0
	for _, value := range m {
		values[i] = value
		i++
	}
	sort.Sort(values)
	return values
}
