package set

type Set map[string]bool

func New() Set {
	return make(map[string]bool)
}

func (s Set) Add(key string) {
	s[key] = true
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
