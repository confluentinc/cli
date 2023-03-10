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
	return s[key]
}

func (s Set) Slice() []string {
	return GetKeys(s)
}

// Returns the elements of list which are not contained in set s
func (s Set) Difference(list []string) []string {
	var remainingElements []string
	for _, element := range list {
		if !s.Contains(element) {
			remainingElements = append(remainingElements, element)
		}
	}

	return remainingElements
}
