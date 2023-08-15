package types

import "golang.org/x/exp/constraints"

type SortableSlice[T constraints.Ordered] []T

func NewSortableSlice[T constraints.Ordered](size int) SortableSlice[T] {
	return make([]T, size)
}

func (s SortableSlice[T]) Len() int {
	return len(s)
}

func (s SortableSlice[T]) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s SortableSlice[T]) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
