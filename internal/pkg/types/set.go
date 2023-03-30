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

// Returns the elements of list which are contained in set s (intersection) and which aren't (difference)
func (s Set) IntersectionAndDifference(list []string) ([]string, []string) {
	var commonElements, remainingElements []string
	for _, element := range list {
		if s.Contains(element) {
			commonElements = append(commonElements, element)
		} else {
			remainingElements = append(remainingElements, element)
		}
	}

	return commonElements, remainingElements
}
