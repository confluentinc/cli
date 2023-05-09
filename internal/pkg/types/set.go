package types

type Set map[string]bool

func NewSet(keys ...string) Set {
	s := make(Set)
	for _, key := range keys {
		s.Add(key)
	}
	return s
}

func (s Set) Add(key string) {
	s[key] = true
}

func (s Set) Contains(key string) bool {
	_, ok := s[key]
	return ok
}

func (s Set) Slice() []string {
	return GetKeys(s)
}
