package logs

import (
	"encoding/json"
	"time"

	loggingv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/logging/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/log"
)

// discoveryCacheTTL is how long a discovered log source is reused. A stale entry is harmless, because the server
// validates every request again and a validation miss refreshes the entry.
const discoveryCacheTTL = 3 * time.Hour

// discoveryCacheKey ties cached entries to the platform and organization they were fetched for.
func discoveryCacheKey(platformServer, organizationId string) string {
	return platformServer + "|" + organizationId
}

// cachedSource returns the named log source from the cache, or nil when it is missing, expired or belongs to another
// platform or organization.
func cachedSource(cache *config.LogsDiscoveryCache, key, name string, now time.Time) *loggingv1.LoggingV1Source {
	if cache == nil || cache.Key != key {
		return nil
	}
	entry, ok := cache.Sources[name]
	if !ok || entry == nil || entry.FetchedAt.After(now) || now.Sub(entry.FetchedAt) >= discoveryCacheTTL {
		return nil
	}

	var source loggingv1.LoggingV1Source
	if err := json.Unmarshal(entry.Source, &source); err != nil {
		return nil
	}
	return &source
}

// withSource returns the cache with the log source added. A cache that belongs to another platform or organization
// is dropped first.
func withSource(cache *config.LogsDiscoveryCache, key string, source *loggingv1.LoggingV1Source, now time.Time) *config.LogsDiscoveryCache {
	if cache == nil || cache.Key != key {
		cache = &config.LogsDiscoveryCache{Key: key}
	}

	data, err := json.Marshal(source)
	if err != nil {
		return cache
	}
	if cache.Sources == nil {
		cache.Sources = map[string]*config.LogsCachedSource{}
	}
	cache.Sources[source.Name] = &config.LogsCachedSource{FetchedAt: now, Source: data}
	return cache
}

// loadLogSource returns the named log source from the cache when it is fresh, and from the discovery API otherwise.
// The second result reports whether it came from the cache. refresh skips the cache.
func loadLogSource(c *pcmd.AuthenticatedCLICommand, name string, refresh bool) (*loggingv1.LoggingV1Source, bool, error) {
	key := discoveryCacheKey(c.Context.GetPlatformServer(), c.Context.GetCurrentOrganization())
	now := time.Now()

	if !refresh {
		if source := cachedSource(c.Context.GetLogsDiscoveryCache(), key, name, now); source != nil {
			return source, true, nil
		}
	}

	source, err := c.V2Client.DescribeLogSource(name)
	if err != nil {
		return nil, false, err
	}
	storeLogSource(c, key, source, now)
	return source, false, nil
}

// storeLogSource saves a log source for later commands. Caching is best effort, so a failed save only gets logged.
func storeLogSource(c *pcmd.AuthenticatedCLICommand, key string, source *loggingv1.LoggingV1Source, now time.Time) {
	c.Context.SetLogsDiscoveryCache(withSource(c.Context.GetLogsDiscoveryCache(), key, source, now))
	if err := c.Config.Save(); err != nil {
		log.CliLogger.Debugf("Failed to cache log source %q: %v", source.Name, err)
	}
}

// resolveSource loads a log source and validates the query against it. A source that came from the cache may be
// stale, so when validation rejects it, the source is fetched again and the query is validated once more.
func resolveSource(load func(refresh bool) (*loggingv1.LoggingV1Source, bool, error), validate func(*loggingv1.LoggingV1Source) error) (*loggingv1.LoggingV1Source, error) {
	source, cached, err := load(false)
	if err != nil {
		return nil, err
	}

	err = validate(source)
	if err == nil || !cached {
		return source, err
	}

	if source, _, err = load(true); err != nil {
		return nil, err
	}
	return source, validate(source)
}
