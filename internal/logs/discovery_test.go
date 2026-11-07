package logs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	loggingv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/logging/v1"
)

var clusterHealth = &loggingv1.LoggingV1Source{
	Name:          "cluster-health",
	ResourceType:  "kafka",
	IdPrefix:      loggingv1.PtrString("lkc"),
	AllowedScopes: []string{"ORGANIZATION", "ENVIRONMENT", "RESOURCE"},
	SearchFields:  []loggingv1.LoggingV1SearchFieldDescriptor{{Name: "id", Type: "string"}},
}

var connectorLogs = &loggingv1.LoggingV1Source{Name: "connector", ResourceType: "connector", AllowedScopes: []string{"RESOURCE"}}

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
	require.Nil(t, cachedSource(cache, key, "connector", now), "unknown source")
	require.Nil(t, cachedSource(cache, discoveryCacheKey("https://confluent.cloud", "org-2"), "cluster-health", now), "other organization")
	require.Nil(t, cachedSource(cache, key, "cluster-health", now.Add(discoveryCacheTTL)), "expired")
	require.NotNil(t, cachedSource(cache, key, "cluster-health", now.Add(discoveryCacheTTL-2*time.Hour)), "still fresh")
	require.Nil(t, cachedSource(cache, key, "cluster-health", now.Add(-2*time.Hour)), "fetched in the future")
}

func TestWithSource(t *testing.T) {
	now := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)
	key := discoveryCacheKey("https://confluent.cloud", "org-1")

	cache := withSource(nil, key, clusterHealth, now)
	cache = withSource(cache, key, connectorLogs, now)
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

	cache := withSources(key, []loggingv1.LoggingV1Source{*connectorLogs, *clusterHealth}, now.Add(-time.Hour))

	got := cachedSources(cache, key, now)
	require.Len(t, got, 2)
	require.Equal(t, "cluster-health", got[0].Name, "sorted by name")
	require.Equal(t, "connector", got[1].Name)

	require.Nil(t, cachedSources(nil, key, now), "no cache")
	require.Nil(t, cachedSources(cache, discoveryCacheKey("https://confluent.cloud", "org-2"), now), "other organization")
	require.Nil(t, cachedSources(cache, key, now.Add(discoveryCacheTTL)), "expired")

	// Single sources alone do not tell which log sources exist.
	require.Nil(t, cachedSources(withSource(nil, key, clusterHealth, now), key, now), "never listed")

	// A full list replaces whatever was cached before.
	relisted := withSources(key, []loggingv1.LoggingV1Source{*clusterHealth}, now)
	require.Len(t, cachedSources(relisted, key, now), 1)
}
