//go:build cuj_test

package cuj

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// uniqueName generates a unique resource name with the given prefix.
func uniqueName(prefix string) string {
	return fmt.Sprintf("cli-cuj-%s-%06d", prefix, rand.Intn(1000000))
}

// confluent runs the built CLI with the isolated HOME and returns its combined output.
func (s *CUJTestSuite) confluent(args ...string) (string, error) {
	out, err := s.cliCommand(context.Background(), args...).CombinedOutput()
	return string(out), err
}

// confluentJSON runs `confluent <args> -o json` and decodes the output into v.
func (s *CUJTestSuite) confluentJSON(v any, args ...string) error {
	cmd := s.cliCommand(context.Background(), append(args, "-o", "json")...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("`confluent %s` failed: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	if err := json.Unmarshal(bytes.TrimSpace(out), v); err != nil {
		return fmt.Errorf("`confluent %s` returned unexpected JSON: %w", strings.Join(args, " "), err)
	}
	return nil
}

// topicExists reports whether the topic exists on the suite's cluster, from live state.
func (s *CUJTestSuite) topicExists(name string) (bool, error) {
	var topics []map[string]any
	if err := s.confluentJSON(&topics, "kafka", "topic", "list", "--cluster", s.clusterID, "--environment", s.envID); err != nil {
		return false, err
	}
	for _, topic := range topics {
		if n, _ := topic["name"].(string); n == name {
			return true, nil
		}
	}
	return false, nil
}

// deleteTopicOnCleanup deletes the topic after the test and verifies it is gone.
// Set CLI_CUJ_KEEP_RESOURCES=1 to leave it for manual inspection.
func (s *CUJTestSuite) deleteTopicOnCleanup(t *testing.T, name string) {
	t.Cleanup(func() {
		if keepResources(t, fmt.Sprintf("topic %q on cluster %s", name, s.clusterID)) {
			return
		}
		exists, err := s.topicExists(name)
		if err != nil {
			t.Errorf("cleanup: %v", err)
			return
		}
		if !exists {
			return
		}
		if out, err := s.confluent("kafka", "topic", "delete", name, "--cluster", s.clusterID, "--environment", s.envID, "--force"); err != nil {
			t.Errorf("cleanup: failed to delete topic %q: %v\n%s", name, err, out)
			return
		}
		if exists, err := s.topicExists(name); err != nil || exists {
			t.Errorf("cleanup: topic %q still exists after delete (err=%v)", name, err)
			return
		}
		t.Logf("cleanup: deleted topic %q", name)
	})
}

// keepResources reports whether cleanup should be skipped (CLI_CUJ_KEEP_RESOURCES=1), logging what is kept.
func keepResources(t *testing.T, what string) bool {
	if os.Getenv("CLI_CUJ_KEEP_RESOURCES") != "1" {
		return false
	}
	t.Logf("CLI_CUJ_KEEP_RESOURCES=1: leaving %s", what)
	return true
}

// createTopic creates an empty topic owned by the harness, deleted after the test.
func (s *CUJTestSuite) createTopic(t *testing.T, name string) {
	t.Helper()
	s.deleteTopicOnCleanup(t, name)
	out, err := s.confluent("kafka", "topic", "create", name, "--cluster", s.clusterID, "--environment", s.envID)
	require.NoError(t, err, out)
}

type apiKey struct {
	Key          string `json:"key"`
	Description  string `json:"description"`
	ResourceType string `json:"resource_type"`
	Resource     string `json:"resource"`
}

func keyIDs(keys []apiKey) []string {
	ids := make([]string, len(keys))
	for i, k := range keys {
		ids[i] = k.Key
	}
	return ids
}

// listAPIKeys lists the current user's API keys only. In a shared org this
// keeps other people's keys from being counted as the agent's, or deleted.
func (s *CUJTestSuite) listAPIKeys() ([]apiKey, error) {
	var keys []apiKey
	err := s.confluentJSON(&keys, "api-key", "list", "--current-user")
	return keys, err
}

// apiKeySet records the current user's API key IDs, for a before/after comparison.
func (s *CUJTestSuite) apiKeySet(t *testing.T) map[string]bool {
	t.Helper()
	keys, err := s.listAPIKeys()
	require.NoError(t, err)
	set := make(map[string]bool, len(keys))
	for _, k := range keys {
		set[k.Key] = true
	}
	return set
}

// newAPIKeys returns the current user's API keys that are not in before.
func (s *CUJTestSuite) newAPIKeys(before map[string]bool) ([]apiKey, error) {
	keys, err := s.listAPIKeys()
	if err != nil {
		return nil, err
	}
	var added []apiKey
	for _, k := range keys {
		if !before[k.Key] {
			added = append(added, k)
		}
	}
	return added, nil
}

// createAPIKey creates an API key scoped to the suite's cluster, deleted after the
// test, and returns the key and its secret.
func (s *CUJTestSuite) createAPIKey(t *testing.T, description string) (string, string) {
	t.Helper()
	var created struct {
		Key    string `json:"api_key"`
		Secret string `json:"api_secret"`
	}
	require.NoError(t, s.confluentJSON(&created, "api-key", "create", "--resource", s.clusterID, "--environment", s.envID, "--description", description))
	s.deleteAPIKeysOnCleanup(t, func() ([]string, error) { return []string{created.Key}, nil })
	return created.Key, created.Secret
}

// deleteAPIKeysOnCleanup deletes the API keys returned by keys, which is called at
// cleanup time, and verifies they are gone. Keys that no longer exist (for
// example because the agent deleted them) are skipped.
func (s *CUJTestSuite) deleteAPIKeysOnCleanup(t *testing.T, keys func() ([]string, error)) {
	t.Cleanup(func() {
		ids, err := keys()
		if err != nil {
			t.Errorf("cleanup: %v", err)
			return
		}
		existing, err := s.listAPIKeys()
		if err != nil {
			t.Errorf("cleanup: %v", err)
			return
		}
		present := make(map[string]bool, len(existing))
		for _, k := range existing {
			present[k.Key] = true
		}
		var toDelete []string
		for _, id := range ids {
			if present[id] {
				toDelete = append(toDelete, id)
			}
		}
		if len(toDelete) == 0 || keepResources(t, fmt.Sprintf("API keys %v", toDelete)) {
			return
		}
		for _, id := range toDelete {
			if out, err := s.confluent("api-key", "delete", id, "--force"); err != nil {
				t.Errorf("cleanup: failed to delete API key %q: %v\n%s", id, err, out)
			}
		}
		remaining, err := s.listAPIKeys()
		if err != nil {
			t.Errorf("cleanup: %v", err)
			return
		}
		for _, k := range remaining {
			for _, id := range toDelete {
				if k.Key == id {
					t.Errorf("cleanup: API key %q still exists after delete", id)
				}
			}
		}
		t.Logf("cleanup: deleted API keys %v", toDelete)
	})
}

// consumeAll reads the messages on a topic from the beginning, authenticating
// with the given API key. The consumer never exits on its own, so it is stopped
// after wait. If it fails (for example while a new key is still propagating),
// it is retried.
func (s *CUJTestSuite) consumeAll(t *testing.T, topic, key, secret string, wait time.Duration) ([]string, error) {
	t.Helper()
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), wait)
		cmd := s.cliCommand(ctx, "kafka", "topic", "consume", topic, "--from-beginning",
			"--api-key", key, "--api-secret", secret, "--cluster", s.clusterID, "--environment", s.envID,
			"--value-format", "string", "--group", uniqueName("verify-group"))
		var stdout, stderr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		err := cmd.Run()
		stopped := ctx.Err() == context.DeadlineExceeded
		cancel()

		if err == nil || stopped {
			var messages []string
			for _, line := range strings.Split(stdout.String(), "\n") {
				if line = strings.TrimSpace(line); line != "" {
					messages = append(messages, line)
				}
			}
			if len(messages) == 0 {
				t.Logf("consumer read no messages from %q; consumer stderr: %s", topic, strings.TrimSpace(stderr.String()))
			}
			return messages, nil
		}
		lastErr = fmt.Errorf("consuming %q failed (attempt %d): %w: %s", topic, attempt, err, strings.TrimSpace(stderr.String()))
		t.Log(lastErr)
		time.Sleep(10 * time.Second)
	}
	return nil, lastErr
}
