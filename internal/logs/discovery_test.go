package logs

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	loggingv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/logging/v1"
)

func TestCachedSource(t *testing.T) {
	now := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)
	key := discoveryCacheKey("https://confluent.cloud", "org-1")

	cache := withSource(nil, key, clusterHealth, now.Add(-time.Hour))

	got := cachedSource(cache, key, "cluster-health", now)
	require.NotNil(t, got)
	require.Equal(t, "cluster-health", got.Name)
	require.Equal(t, clusterHealth.GetIdPrefix(), got.GetIdPrefix())
	require.Equal(t, clusterHealth.SearchFields, got.SearchFields)

	require.Nil(t, cachedSource(nil, key, "cluster-health", now), "no cache")
	require.Nil(t, cachedSource(cache, key, "unknown-source", now), "unknown source")
	require.Nil(t, cachedSource(cache, discoveryCacheKey("https://confluent.cloud", "org-2"), "cluster-health", now), "other organization")
	require.Nil(t, cachedSource(cache, key, "cluster-health", now.Add(discoveryCacheTTL)), "expired")
	require.NotNil(t, cachedSource(cache, key, "cluster-health", now.Add(discoveryCacheTTL-2*time.Hour)), "still fresh")
	require.Nil(t, cachedSource(cache, key, "cluster-health", now.Add(-2*time.Hour)), "fetched in the future")
}

func TestWithSource(t *testing.T) {
	now := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)
	key := discoveryCacheKey("https://confluent.cloud", "org-1")

	cache := withSource(nil, key, clusterHealth, now)
	cache = withSource(cache, key, noPrefixSource, now)
	require.Len(t, cache.Sources, 2)

	// A newer fetch replaces the entry.
	later := now.Add(time.Hour)
	cache = withSource(cache, key, clusterHealth, later)
	require.Len(t, cache.Sources, 2)
	require.True(t, cache.Sources["cluster-health"].FetchedAt.Equal(later))

	// Entries for another organization are dropped, not mixed in.
	other := withSource(cache, discoveryCacheKey("https://confluent.cloud", "org-2"), clusterHealth, now)
	require.Len(t, other.Sources, 1)
	require.Equal(t, discoveryCacheKey("https://confluent.cloud", "org-2"), other.Key)
}

func TestCachedSources(t *testing.T) {
	now := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)
	key := discoveryCacheKey("https://confluent.cloud", "org-1")

	cache := withSources(key, []loggingv1.LoggingV1Source{*noPrefixSource, *clusterHealth}, now.Add(-time.Hour))

	got := cachedSources(cache, key, now)
	require.Len(t, got, 2)
	require.Equal(t, "cluster-health", got[0].Name, "sorted by name")
	require.Equal(t, "resource-source", got[1].Name)

	require.Nil(t, cachedSources(nil, key, now), "no cache")
	require.Nil(t, cachedSources(cache, discoveryCacheKey("https://confluent.cloud", "org-2"), now), "other organization")
	require.Nil(t, cachedSources(cache, key, now.Add(discoveryCacheTTL)), "expired")

	// Single sources alone do not tell which log sources exist.
	require.Nil(t, cachedSources(withSource(nil, key, clusterHealth, now), key, now), "never listed")

	// A full list replaces whatever was cached before.
	relisted := withSources(key, []loggingv1.LoggingV1Source{*clusterHealth}, now)
	require.Len(t, cachedSources(relisted, key, now), 1)
}

func TestResolveSource(t *testing.T) {
	stale := &loggingv1.LoggingV1Source{Name: "cluster-health"}
	fresh := &loggingv1.LoggingV1Source{Name: "cluster-health", SearchFields: []loggingv1.LoggingV1SearchFieldDescriptor{{Name: "level", Type: "string"}}}
	needsLevel := func(source *loggingv1.LoggingV1Source) error {
		if len(source.SearchFields) == 0 {
			return errors.New("no level field")
		}
		return nil
	}

	loads := func(sources ...*loggingv1.LoggingV1Source) (func(bool) (*loggingv1.LoggingV1Source, bool, error), *[]bool) {
		var refreshes []bool
		return func(refresh bool) (*loggingv1.LoggingV1Source, bool, error) {
			refreshes = append(refreshes, refresh)
			// The first load is served from the cache, and a refresh goes to the API.
			return sources[len(refreshes)-1], !refresh, nil
		}, &refreshes
	}

	// A cached source that passes validation is used as is.
	load, calls := loads(fresh)
	source, err := resolveSource(load, needsLevel)
	require.NoError(t, err)
	require.Equal(t, fresh, source)
	require.Equal(t, []bool{false}, *calls)

	// A cached source that fails validation is refreshed once, and the refreshed one is used.
	load, calls = loads(stale, fresh)
	source, err = resolveSource(load, needsLevel)
	require.NoError(t, err)
	require.Equal(t, fresh, source)
	require.Equal(t, []bool{false, true}, *calls)

	// If the refreshed source fails too, the error is reported after exactly one refresh.
	load, calls = loads(stale, stale)
	_, err = resolveSource(load, needsLevel)
	require.EqualError(t, err, "no level field")
	require.Equal(t, []bool{false, true}, *calls)

	// A source that was just fetched from the API is not fetched again.
	var apiCalls int
	_, err = resolveSource(func(bool) (*loggingv1.LoggingV1Source, bool, error) {
		apiCalls++
		return stale, false, nil
	}, needsLevel)
	require.EqualError(t, err, "no level field")
	require.Equal(t, 1, apiCalls)

	// Load errors are returned as they are.
	_, err = resolveSource(func(bool) (*loggingv1.LoggingV1Source, bool, error) {
		return nil, false, errors.New("unknown source")
	}, needsLevel)
	require.EqualError(t, err, "unknown source")
}
