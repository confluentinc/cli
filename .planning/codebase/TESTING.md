# Testing Patterns

**Analysis Date:** 2026-03-08

## Test Framework

**Runner:**
- Go testing (standard library)
- go test v1.25.7
- Config: `Makefile` targets, no separate test config files

**Assertion Library:**
- `github.com/stretchr/testify` v1.10.0
- Primary packages: `require`, `suite`, `assert`

**Mocking:**
- `go.uber.org/mock` v0.4.0 (uber's fork of gomock)
- Generated mocks in `mock/` directories

**Run Commands:**
```bash
make test                  # Run all tests (unit + integration)
make unit-test             # Run unit tests only
make integration-test      # Run integration tests only
make lint                  # Run linters (golangci-lint + custom CLI linter)
```

## Test File Organization

**Location:**
- Unit tests: Co-located with source files in same package (e.g., `internal/api-key/command_test.go`)
- Integration tests: Separate `test/` directory at repo root

**Naming:**
- Unit tests: `*_test.go` (e.g., `command_create_test.go`, `command_test.go`)
- Integration tests: `test/<command>_test.go` (e.g., `test/kafka_test.go`, `test/api_key_test.go`)
- Live tests: `test/live/*_live_test.go`

**Structure:**
```
cli/
├── internal/
│   ├── api-key/
│   │   ├── command.go
│   │   ├── command_create.go
│   │   └── command_test.go           # Unit tests
├── test/
│   ├── cli_test.go                    # Integration test suite setup
│   ├── kafka_test.go                  # Integration tests for kafka commands
│   ├── api_key_test.go                # Integration tests for api-key commands
│   ├── fixtures/
│   │   └── output/
│   │       ├── kafka/
│   │       │   ├── 1.golden           # Golden files for expected output
│   │       │   └── cluster/
│   │       │       └── list.golden
│   └── live/                          # Live tests (run against real services)
│       └── *_live_test.go
```

## Test Structure

**Unit Tests:**
```go
package apikey

import (
    "testing"

    "github.com/stretchr/testify/require"
)

func TestGetResourceType(t *testing.T) {
    require.Equal(t, "cloud", getResourceType(apikeysv2.ObjectReference{Kind: apikeysv2.PtrString("Cloud")}))
    require.Equal(t, "kafka", getResourceType(apikeysv2.ObjectReference{ApiVersion: apikeysv2.PtrString("cmk/v2"), Kind: apikeysv2.PtrString("Cluster")}))
}
```

**Integration Tests (Suite Pattern):**
```go
package test

import (
    "testing"

    "github.com/stretchr/testify/suite"
)

type CLITestSuite struct {
    suite.Suite
    TestBackend *testserver.TestBackend
}

func TestCLI(t *testing.T) {
    suite.Run(t, new(CLITestSuite))
}

func (s *CLITestSuite) SetupSuite() {
    // Build CLI binary with coverage instrumentation
    // Start mock backend server
}

func (s *CLITestSuite) TearDownSuite() {
    s.TestBackend.Close()
}
```

**Patterns:**
- Unit tests: Table-driven tests using `require` assertions
- Integration tests: Golden file pattern with test suite
- Use `require.New(t)` for cleaner syntax in complex tests
- Named test cases in subtests: `t.Run(test.name, func(t *testing.T) {...})`

## Mocking

**Framework:**
- `go.uber.org/mock` for generating mocks
- Mocks stored in `mock/` subdirectories

**Patterns:**
```go
// Mock definition (generated)
type LoginCredentialsManager struct {
    GetCloudCredentialsFromEnvVarFunc func(string) func() (*pauth.Credentials, error)
    GetCloudCredentialsFromPromptFunc func(string) func() (*pauth.Credentials, error)
}

// Mock usage in test
mockLoginCredentialsManager := &climock.LoginCredentialsManager{
    GetCloudCredentialsFromEnvVarFunc: func(_ string) func() (*pauth.Credentials, error) {
        return func() (*pauth.Credentials, error) {
            return nil, nil
        }
    },
}
```

**What to Mock:**
- External SDK clients (ccloud-sdk-go, mds-sdk-go)
- HTTP clients and backend services
- Auth/credential providers
- File system operations (where appropriate)

**What NOT to Mock:**
- Standard library types (use real implementations)
- Simple data structures
- Pure functions without side effects

## Fixtures and Factories

**Golden Files:**
- Location: `test/fixtures/output/<command>/<subcommand>/*.golden`
- Format: Plain text output that CLI should produce
- Organization: Mirrors command structure

**Test Data Pattern:**
```go
type CLITest struct {
    name            string   // Test name for output
    args            string   // CLI arguments to run
    env             []string // Environment variables
    login           string   // "cloud" or "onprem"
    loginURL        string   // Optional custom URL
    useKafka        string   // Kafka cluster to select
    authKafka       bool     // Create API key for Kafka
    fixture         string   // Path to golden file
    disableAuditLog bool     // Audit log state
    regex           bool     // Treat fixture as regex
    contains        string   // String to check in output
    notContains     string   // String to ensure absent
    exitCode        int      // Expected exit code
    workflow        bool     // Maintain state between tests
    wantFunc        func(t *testing.T) // Custom assertions
    input           string   // Stdin input
}
```

**Integration Test Execution:**
```go
tests := []CLITest{
    {args: "kafka cluster list", fixture: "kafka/cluster/list.golden"},
    {args: "kafka cluster create my-cluster --cloud aws --region us-east-1", fixture: "kafka/cluster/create.golden"},
    {args: "kafka cluster delete lkc-123456 --force", fixture: "kafka/cluster/delete.golden"},
}
for _, test := range tests {
    s.runIntegrationTest(test)
}
```

**Factory Pattern:**
- Minimal in unit tests (tests are simple, focused)
- Test backends provide data factories in integration tests
- `testserver.TestBackend` provides mock API responses

## Coverage

**Requirements:**
- No enforced minimum coverage
- Coverage collected separately for unit and integration tests

**View Coverage:**
```bash
make unit-test             # Generates coverage.unit.out
make integration-test      # Generates coverage.integration.out
make coverage              # Merges into coverage.txt
```

**Coverage Collection:**
- Unit tests: Standard `-coverprofile=coverage.unit.out -covermode=atomic`
- Integration tests: Build CLI with `-cover` flag, use `GOCOVERDIR` for runtime coverage collection
- Combined coverage: Merge both outputs with `coverage` target

**CI vs Local:**
- CI uses `gotestsum` for JUnit XML output
- Local runs use standard `go test`
- Both collect coverage data

## Test Types

**Unit Tests:**
- Scope: Individual functions, pure logic, data transformations
- Location: Co-located with source (`internal/<package>/*_test.go`)
- Examples:
  - `TestGetResourceType` - Type mapping logic
  - `TestFormatBalance` - Output formatting
  - `TestFormatExpiration` - Date formatting
  - `TestValidateUrl` - Input validation
- Pattern: Fast, focused, no external dependencies

**Integration Tests:**
- Scope: Full CLI command execution against mock backend
- Location: `test/<command>_test.go`
- Examples:
  - `TestKafka` - All kafka command workflows
  - `TestLogin` - Login flows with different credential sources
  - `TestAPIKey` - API key lifecycle
- Pattern: Build CLI binary, execute as subprocess, compare output to golden files
- Mock backend: `testserver.TestBackend` provides API responses

**Live Tests:**
- Scope: Real API calls against staging/prod environments
- Location: `test/live/*_live_test.go`
- Build tag: `live_test`
- Groups: `core`, `kafka`, `schema_registry`, `iam`, `auth`, `connect`
- Run command:
```bash
make live-test                              # All live tests
make live-test-kafka                        # Kafka tests only
CLI_LIVE_TEST_GROUPS="core,kafka" make live-test  # Specific groups
```
- Pattern: Long-running, parallel execution, requires real credentials

## Common Patterns

**Async Testing:**
- Not heavily used (CLI is mostly synchronous)
- Where needed, use channels and timeouts:
```go
done := make(chan bool)
go func() {
    // async operation
    done <- true
}()
select {
case <-done:
    // success
case <-time.After(5 * time.Second):
    t.Fatal("timeout")
}
```

**Error Testing:**
```go
// Unit tests
require.Error(t, err)
require.Contains(t, err.Error(), "expected substring")

// Integration tests
{args: "kafka cluster create", fixture: "error.golden", exitCode: 1}
```

**Table-Driven Tests:**
```go
func TestFormatBalance(t *testing.T) {
    tests := []struct {
        spent    int32
        total    int32
        expected string
    }{
        {0, 10000, "$0.00/1.00 USD"},
        {5000, 10000, "$0.50/1.00 USD"},
    }
    for _, tt := range tests {
        require.Equal(t, tt.expected, formatBalance(tt.spent, tt.total))
    }
}
```

**Workflow Tests:**
- Set `workflow: true` in `CLITest` to maintain state between test steps
- Useful for testing multi-command workflows (create → use → update → delete)
- Config/state persists across tests in the workflow

**Golden File Updates:**
```bash
make integration-test INTEGRATION_TEST_ARGS="-update"
```

**Selective Test Execution:**
```bash
# Unit tests
make unit-test UNIT_TEST_ARGS="-run TestApiTestSuite"
make unit-test UNIT_TEST_ARGS="-run TestApiTestSuite/TestCreateCloudAPIKey"

# Integration tests
make integration-test INTEGRATION_TEST_ARGS="-run TestCLI/TestKafka"
make integration-test INTEGRATION_TEST_ARGS="-run TestCLI/TestKafka/kafka_cluster_list"
```

**Test Setup/Teardown:**
- Suite-level: `SetupSuite()`, `TearDownSuite()`
- Test-level: `SetupTest()`, `TearDownTest()` (less common)
- Integration tests: `resetConfiguration()` between non-workflow tests

**Environment Variable Handling:**
```go
// Set env vars for test
env := []string{pauth.ConfluentCloudEmail + "=fake@user.com"}
for _, e := range env {
    keyVal := strings.Split(e, "=")
    os.Setenv(keyVal[0], keyVal[1])
}
defer func() {
    for _, e := range env {
        keyVal := strings.Split(e, "=")
        os.Unsetenv(keyVal[0])
    }
}()
```

**Build Requirements:**
- Integration tests require CLI rebuild: `make build-for-integration-test`
- Skip rebuild with: `make integration-test INTEGRATION_TEST_ARGS="-no-rebuild"`
- Binary location: `test/bin/confluent` (or `.exe` on Windows)
- Build includes coverage instrumentation and test flags

**Platform-Specific Testing:**
- Windows: Special handling in `runCommand()` for non-POSIX shell parsing
- macOS: File descriptor limit considerations (see CONTRIBUTING.md)
- Cross-platform: `runtime.GOOS` checks where needed

---

*Testing analysis: 2026-03-08*
