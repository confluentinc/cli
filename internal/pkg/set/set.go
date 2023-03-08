package set

import "github.com/confluentinc/cli/internal/pkg/utils"

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
	return utils.GetKeys(s)
}
