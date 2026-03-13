# CLI Live Integration Tests

Live integration tests run the real CLI binary against Confluent Cloud. They create, read, update, and delete real resources to verify end-to-end CLI behavior.

## Prerequisites

1. **CLI binary** — Build with `make build-for-live-test`
2. **Authentication** — Use one of the following methods:

### Option A: Email/Password (default, used in CI)

| Variable | Required | Description |
|---|---|---|
| `CONFLUENT_CLOUD_EMAIL` | Yes | Confluent Cloud login email |
| `CONFLUENT_CLOUD_PASSWORD` | Yes | Confluent Cloud login password |

### Option B: Pre-Authenticated Session

| Variable | Required | Description |
|---|---|---|
| `CLI_LIVE_TEST_CONFIG_DIR` | Yes | Path to your HOME directory (or `.confluent/` config directory). The framework copies this config into an isolated temp dir so tests don't mutate your real CLI state. |

Example: `CLI_LIVE_TEST_CONFIG_DIR=~ make live-test-core`

### Cloud Configuration

| Variable | Required | Description |
|---|---|---|
| `CLI_LIVE_TEST_CLOUD` | No | Cloud provider: `aws` (default), `gcp`, `azure` |
| `CLI_LIVE_TEST_REGION` | No | Cloud region (default: `us-east-1`) |
| `CLI_LIVE_TEST_CLUSTER_TYPE` | No | Cluster type: `basic` (default), `standard`, `dedicated` |
| `CLI_LIVE_TEST_VARIANTS` | No | Comma-separated `cloud:region:type` triples for multi-cloud testing (e.g., `aws:us-east-1:basic,gcp:us-east1:basic`) |
| `LIVE_TEST_ENVIRONMENT_ID` | Kafka topics only | Pre-existing environment ID for topic tests |
| `KAFKA_STANDARD_AWS_CLUSTER_ID` | Kafka topics only | Pre-existing cluster ID for topic tests |

## Running Tests

### All tests
```bash
make live-test
```

### By group
```bash
make live-test-core        # environment, service account, API key
make live-test-kafka       # kafka cluster, topic, ACL, consumer group
make live-test-essential   # core + kafka + schema_registry + auth
make live-test CLI_LIVE_TEST_GROUPS="kafka"  # kafka only
```

### Single resource
```bash
make live-test-resource RESOURCE=kafka_cluster
make live-test-resource RESOURCE=environment
make live-test-resource   # prints all available resources
```

### Multi-cloud
```bash
# Convenience target: runs kafka tests across aws, gcp, and azure
make live-test-multicloud

# Custom variants
CLI_LIVE_TEST_VARIANTS="aws:us-east-1:basic,gcp:us-east1:basic,azure:eastus:basic" make live-test-kafka
```

### Single test (go test directly)
```bash
CLI_LIVE_TEST=1 go test ./test/live/ -v -run TestLive/TestKafkaClusterCRUDLive \
    -tags="live_test,kafka" -timeout 30m
```

## Test Groups

Tests are organized into groups via Go build tags:

| Group | Tag | Makefile Target |
|---|---|---|
| Core | `core` | `make live-test-core` |
| Kafka | `kafka` | `make live-test-kafka` |
| Schema Registry | `schema_registry` | `make live-test-schema-registry` |
| IAM | `iam` | `make live-test-iam` |
| Auth | `auth` | `make live-test-auth` |
| Connect | `connect` | `make live-test-connect` |
| Essential | `core,kafka,schema_registry,auth` | `make live-test-essential` |
| All | `all` | `make live-test` |

## Resource Lookup Table

| Resource | Test Function | Group | File |
|---|---|---|---|
| Environment | `TestEnvironmentCRUDLive` | `core` | `environment_live_test.go` |
| Service Account | `TestServiceAccountCRUDLive` | `core` | `service_account_live_test.go` |
| API Key | `TestApiKeyCRUDLive` | `core` | `api_key_live_test.go` |
| Kafka Cluster | `TestKafkaClusterCRUDLive` | `kafka` | `kafka_cluster_live_test.go` |
| Kafka Topic | `TestKafkaTopicCRUDLive` | `kafka` | `kafka_topic_live_test.go` |
| Kafka ACL | `TestKafkaACLCRUDLive` | `kafka` | `kafka_acl_live_test.go` |
| Kafka Consumer Group | `TestKafkaConsumerGroupListLive` | `kafka` | `kafka_consumer_group_live_test.go` |
| Schema Registry | `TestSchemaRegistrySchemaCRUDLive` | `schema_registry` | `schema_registry_live_test.go` |
| IAM RBAC | `TestRBACRoleBindingCRUDLive` | `iam` | `iam_rbac_live_test.go` |
| Login/Logout | `TestLoginLogoutLive` | `auth` | `login_live_test.go` |
| Connect | `TestConnectClusterCRUDLive` | `connect` | `connect_live_test.go` |

Run any single resource with: `make live-test-resource RESOURCE=<name>` (e.g., `make live-test-resource RESOURCE=kafka_cluster`).

## Concurrency Model

- Each test method calls `s.setupTestContext(t)` which creates an **isolated HOME directory** and authenticates. This means each test has its own CLI config — no shared state.
- Tests opt in to concurrency by calling `t.Parallel()` at the start. All current tests do this.
- The `-parallel 10` flag in the Makefile controls max concurrent tests.
- Tests that need sequential execution (e.g., tests modifying shared external state) should simply omit the `t.Parallel()` call.

## Writing a New Test

### 1. Create a test file

```go
//go:build live_test && (all || mygroup)

package live

func (s *CLILiveTestSuite) TestMyResourceCRUDLive() {
    t := s.T()
    t.Parallel()
    state := s.setupTestContext(t)

    // ... test body ...
}
```

The test method name **must** end with `Live` to match the `-run=".*Live$"` filter.

### 2. Define test steps

Use `CLILiveTest` structs for each CLI command:

```go
steps := []CLILiveTest{
    {
        Name:     "Create resource",
        Args:     "resource create my-name -o json",
        JSONFieldsExist: []string{"id"},
        CaptureID: "resource_id",  // captures JSON "id" field into state
    },
    {
        Name:         "Describe resource",
        Args:         "resource describe {{.resource_id}} -o json",
        UseStateVars: true,  // enables {{.key}} template substitution
        JSONFields:   map[string]string{"name": "my-name"},
    },
}
```

### 3. Register cleanup

Always register cleanup **before** creating resources (LIFO execution order):

```go
s.registerCleanup(t, "resource delete {{.resource_id}} --force", state)
```

### 4. Run steps

```go
for _, step := range steps {
    t.Run(step.Name, func(t *testing.T) {
        s.runLiveCommand(t, step, state)
    })
}
```

### CLILiveTest Field Reference

| Field | Type | Description |
|---|---|---|
| `Name` | `string` | Step name shown in output |
| `Args` | `string` | CLI arguments (supports `{{.key}}` when `UseStateVars` is true) |
| `ExitCode` | `int` | Expected exit code (default 0) |
| `Input` | `string` | Stdin content |
| `Contains` | `[]string` | Strings that must appear in output |
| `NotContains` | `[]string` | Strings that must NOT appear in output |
| `Regex` | `[]string` | Regex patterns output must match |
| `JSONFields` | `map[string]string` | JSON fields to check (empty value = any non-empty value) |
| `JSONFieldsExist` | `[]string` | JSON fields that must exist (any value) |
| `WantFunc` | `func(t, output, state)` | Custom assertion function |
| `CaptureID` | `string` | State key to store extracted JSON "id" field |
| `UseStateVars` | `bool` | Enable `{{.key}}` template substitution in Args |

### Async Operations

For operations that take time (e.g., cluster provisioning), use `waitForCondition`:

```go
s.waitForCondition(t,
    "kafka cluster describe {{.cluster_id}} -o json",
    state,
    func(output string) bool {
        return strings.EqualFold(extractJSONField(t, output, "status"), "UP")
    },
    30*time.Second,   // poll interval
    10*time.Minute,   // timeout
)
```

## Adding a New Test Group

1. Create test file(s) with build tag: `//go:build live_test && (all || mygroup)`
2. Add a Makefile target:
   ```makefile
   .PHONY: live-test-mygroup
   live-test-mygroup:
       @$(MAKE) live-test CLI_LIVE_TEST_GROUPS="mygroup"
   ```
3. Add the resource to the lookup table in this README and in the `live-test-resource` Makefile target.
4. Update the Semaphore promotion parameters if the group should be selectable in CI.

## CI (Semaphore)

Live tests are triggered via the "Run live integration tests" promotion in `.semaphore/semaphore.yml`. Parameters:

- **CLI_LIVE_TEST_GROUPS** — Test group to run (default: `essential`)
- **CLI_LIVE_TEST_CLOUD** — Cloud provider (default: `aws`)
- **CLI_LIVE_TEST_REGION** — Cloud region (default: `us-east-1`)

Credentials are loaded from Vault secrets in the Semaphore pipeline.
