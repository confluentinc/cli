package releasenotes

import (
	"github.com/google/go-github/v50/github"
	"github.com/hashicorp/go-version"
)

type byVersion []*github.RepositoryTag

func (v byVersion) Len() int      { return len(v) }
func (v byVersion) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v byVersion) Less(i, j int) bool {
	vi, err := version.NewSemver(v[i].GetName())
	if err != nil {
		return true
	}
	vj, err := version.NewSemver(v[j].GetName())
	if err != nil {
		return false
	}

	return vi.LessThan(vj)
}
