package types

type Set[V comparable] map[V]bool

func NewSet[V comparable](keys ...V) Set[V] {
	s := make(Set[V])
	for _, key := range keys {
		s.Add(key)
	}
	return s
}

func (s Set[V]) Add(key V) {
	s[key] = true
}

func (s Set[V]) Contains(key V) bool {
	_, ok := s[key]
	return ok
}

func (s Set[V]) Slice() []V {
	return GetKeys(s)
}
