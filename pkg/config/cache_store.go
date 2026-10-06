package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type cacheStore struct{ dir string }

func newCacheStore() *cacheStore { return &cacheStore{dir: CacheDir()} }

// readJSON loads name into v. A missing or unreadable/corrupt cache file returns
// false with v untouched: the cache is disposable, so a bad read is a miss.
func (s *cacheStore) readJSON(name string, v any) bool {
	data, err := os.ReadFile(filepath.Join(s.dir, name))
	if err != nil {
		return false
	}
	return json.Unmarshal(data, v) == nil
}

func (s *cacheStore) writeJSON(name string, v any) error {
	if err := os.MkdirAll(s.dir, 0700); err != nil {
		return fmt.Errorf("unable to create cache directory %s: %w", s.dir, err)
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal cache %s: %w", name, err)
	}
	return writeFileAtomic(filepath.Join(s.dir, name), data)
}
