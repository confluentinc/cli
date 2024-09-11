package update

import (
	"regexp"
	"sort"

	"github.com/hashicorp/go-version"
)

// FilterUpdates takes the HTML content of the binaries page and returns the versions that are valid minor and major updates.
func FilterUpdates(binaries string, current *version.Version, major bool) (version.Collection, version.Collection) {
	matches := regexp.MustCompile(`<a href="/confluent-cli/binaries/([0-9]+\.[0-9]+\.[0-9]+)/">`).FindAllStringSubmatch(binaries, -1)

	versions := make(version.Collection, len(matches))
	for i, match := range matches {
		versions[i] = version.Must(version.NewVersion(match[1]))
	}

	// Versions are sorted in lexigraphical order instead of semver order
	sort.Sort(versions)

	// Remove versions less than or equal to the current version
	if idx := sort.Search(len(versions), func(i int) bool { return versions[i].GreaterThanOrEqual(current) }); idx < len(versions) {
		versions = versions[idx+1:]
	}

	if !major {
		// Remove major versions
		if idx := sort.Search(len(versions), func(i int) bool { return current.Segments()[0] < versions[i].Segments()[0] }); idx < len(versions) {
			return versions[:idx], versions
		}
	}

	return versions, versions
}
