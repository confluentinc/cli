//go:build cuj_test

package cuj

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// TestAPIKeyCreateVerifyCUJ is the apikey-create-verify CUJ: create a Kafka API key
// scoped to one cluster and prove it works with a produce/consume round trip.
func (s *CUJTestSuite) TestAPIKeyCreateVerifyCUJ() {
	t := s.T()

	// Harness setup, not shown to the agent: an empty verification topic, and the
	// harness's own key for reading it back independently of the agent.
	topic := uniqueName("verify")
	s.createTopic(t, topic)
	verifyKey, verifySecret := s.createAPIKey(t, "cli-cuj harness verification key")

	before := s.apiKeySet(t)
	before[verifyKey] = true // in case the list doesn't show the new key yet
	s.deleteAPIKeysOnCleanup(t, func() ([]string, error) {
		keys, err := s.newAPIKeys(before)
		return keyIDs(keys), err
	})

	// After-state, collected once and shared by the checks.
	newKeys := sync.OnceValues(func() ([]apiKey, error) { return s.newAPIKeys(before) })
	consumed := sync.OnceValues(func() ([]string, error) {
		return s.consumeAll(t, topic, verifyKey, verifySecret, 30*time.Second)
	})
	singleNewKey := func() (apiKey, string, bool) {
		keys, err := newKeys()
		if err != nil {
			return apiKey{}, err.Error(), false
		}
		if len(keys) != 1 {
			return apiKey{}, fmt.Sprintf("want exactly 1 new API key, found %d %v", len(keys), keyIDs(keys)), false
		}
		return keys[0], "", true
	}

	s.runCUJ(t, CUJ{
		Name: "apikey-create-verify",
		Task: `The confluent CLI is installed and already authenticated, with an environment and Kafka cluster selected as the active context.

Supplied resources:
- Kafka cluster ID: {{.cluster_id}}
- Environment ID: {{.environment_id}}
- Verification topic: {{.topic}}

Do the following:
1. Create a new API key scoped only to the supplied Kafka cluster. Do not create an account-wide Cloud API key.
2. Report the new key ID, what it is scoped to, and any information the user must preserve.
3. Using the new key and secret explicitly, not the current user's default credentials, produce a uniquely identifiable test message to the supplied verification topic.
4. Using the same new credentials, consume that message back.
5. Report whether the round trip succeeded.`,
		Checks: []CUJCheck{
			{
				ID: "one_new_key",
				Check: func(_ agentResult) (bool, string) {
					key, detail, ok := singleNewKey()
					if !ok {
						return false, detail
					}
					return true, fmt.Sprintf("exactly 1 new API key: %s", key.Key)
				},
			},
			{
				ID: "scoped_to_cluster",
				Check: func(_ agentResult) (bool, string) {
					key, detail, ok := singleNewKey()
					if !ok {
						return false, detail
					}
					scoped := key.ResourceType == "kafka" && key.Resource == s.clusterID
					return scoped, fmt.Sprintf("new key %s is scoped to %s %q (want kafka %q)", key.Key, key.ResourceType, key.Resource, s.clusterID)
				},
			},
			{
				ID: "not_cloud_key",
				Check: func(_ agentResult) (bool, string) {
					keys, err := newKeys()
					if err != nil {
						return false, err.Error()
					}
					var accountWide []string
					for _, k := range keys {
						if k.ResourceType == "cloud" || k.ResourceType == "global" {
							accountWide = append(accountWide, k.Key)
						}
					}
					return len(accountWide) == 0, fmt.Sprintf("account-wide API keys created during the run: %v", accountWide)
				},
			},
			{
				ID: "reported_key_id",
				Check: func(res agentResult) (bool, string) {
					key, detail, ok := singleNewKey()
					if !ok {
						return false, detail
					}
					reported := strings.Contains(res.Transcript, key.Key)
					return reported, fmt.Sprintf("agent's report contains the new key ID %s: %t", key.Key, reported)
				},
			},
			{
				ID: "message_produced",
				Check: func(_ agentResult) (bool, string) {
					messages, err := consumed()
					if err != nil {
						return false, err.Error()
					}
					return len(messages) > 0, fmt.Sprintf("harness read %d message(s) from %q with its own key", len(messages), topic)
				},
			},
			{
				ID: "round_trip",
				Check: func(res agentResult) (bool, string) {
					messages, err := consumed()
					if err != nil {
						return false, err.Error()
					}
					for _, m := range messages {
						if strings.Contains(res.Transcript, m) {
							return true, fmt.Sprintf("message %q is on the topic and in the agent's report", m)
						}
					}
					return false, fmt.Sprintf("none of the %d message(s) on %q appear in the agent's report", len(messages), topic)
				},
			},
		},
	}, map[string]string{"topic": topic})
}
