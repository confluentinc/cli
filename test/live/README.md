# CLI Live Integration Tests

Live integration tests run the real CLI binary against Confluent Cloud. They create, read, update, and delete real resources to verify end-to-end CLI behavior.

## Prerequisites

1. **CLI binary** — Build with `make build-for-live-test`
2. **Confluent Cloud credentials** — Set the following environment variables:

| Variable | Required | Description |
|---|---|---|
| `CONFLUENT_CLOUD_EMAIL` | Yes | Confluent Cloud login email |
| `CONFLUENT_CLOUD_PASSWORD` | Yes | Confluent Cloud login password |
| `CLI_LIVE_TEST_CLOUD` | No | Cloud provider: `aws` (default), `gcp`, `azure` |
| `CLI_LIVE_TEST_REGION` | No | Cloud region (default: `us-east-1`) |
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
make live-test-essential   # core + kafka
make live-test CLI_LIVE_TEST_GROUPS="kafka"  # kafka only
```

### Multi-cloud
```bash
CLI_LIVE_TEST_CLOUD=gcp CLI_LIVE_TEST_REGION=us-east1 make live-test-essential
```

### Single test
```bash
CLI_LIVE_TEST=1 go test ./test/live/ -v -run TestLive/TestKafkaClusterCRUDLive \
    -tags="live_test,kafka" -timeout 30m
```

## Test Groups

Tests are organized into groups via Go build tags:

| Group | Tag | Tests |
|---|---|---|
| Core | `core` | Environment, Service Account, API Key CRUD |
| Kafka | `kafka` | Kafka Cluster CRUD, Kafka Topic CRUD |
| All | `all` | Everything |

The `essential` group in Semaphore/Makefile maps to `core,kafka`.

## Concurrency Model

- Each test method calls `s.setupTestContext(t)` which creates an **isolated HOME directory** and logs in. This means each test has its own CLI config — no shared state.
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
3. Update the Semaphore promotion parameters if the group should be selectable in CI.

## CI (Semaphore)

Live tests are triggered via the "Run live integration tests" promotion in `.semaphore/semaphore.yml`. Parameters:

- **CLI_LIVE_TEST_GROUPS** — Test group to run (default: `essential`)
- **CLI_LIVE_TEST_CLOUD** — Cloud provider (default: `aws`)
- **CLI_LIVE_TEST_REGION** — Cloud region (default: `us-east-1`)

Credentials are loaded from Vault secrets in the Semaphore pipeline.
