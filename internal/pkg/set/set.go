package set

type Set map[string]bool

func New(keys ...string) Set {
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
	l := make([]string, len(s))
	i := 0

	for key := range s {
		l[i] = key
		i++
	}

	return l
}
