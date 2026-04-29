# Codebase Structure

**Analysis Date:** 2026-03-08

## Directory Layout

```
cli/
├── cmd/                    # Binary entry points
│   ├── confluent/          # Main CLI binary (production)
│   ├── docs/               # Documentation generator
│   ├── lint/               # Custom linter for CLI conventions
│   └── whitelist/          # Command whitelist checker
├── internal/               # Command implementations (38 packages)
│   ├── kafka/              # Kafka commands (topic, cluster, acl, etc.)
│   ├── flink/              # Flink SQL commands
│   ├── iam/                # IAM commands (users, service accounts, RBAC)
│   ├── schema-registry/    # Schema Registry commands
│   ├── login/              # Login command
│   ├── logout/             # Logout command
│   ├── api-key/            # API key management
│   ├── environment/        # Environment management
│   ├── context/            # Context switching
│   ├── connect/            # Kafka Connect commands
│   ├── ksql/               # ksqlDB commands
│   ├── network/            # Networking (peering, private link)
│   ├── plugin/             # Plugin management
│   └── ...                 # 25+ other domain commands
├── pkg/                    # Reusable packages (50+ packages)
│   ├── cmd/                # Command infrastructure (PreRunner, base commands)
│   ├── auth/               # Authentication logic
│   ├── config/             # Config management
│   ├── errors/             # Error handling and messages
│   ├── output/             # Output formatting
│   ├── ccloudv2/           # Confluent Cloud v2 API client
│   ├── kafka/              # Kafka utilities
│   ├── flink/              # Flink shell and utilities
│   ├── deletion/           # Multi-resource deletion helpers
│   ├── schemaregistry/     # Schema Registry client wrapper
│   ├── featureflags/       # LaunchDarkly feature flags
│   ├── log/                # Logging infrastructure
│   ├── utils/              # HTTP, file, crypto utilities
│   └── ...                 # 35+ other packages
├── test/                   # Integration tests
│   ├── fixtures/           # Golden files for test output
│   │   ├── input/          # Test input files
│   │   └── output/         # Expected output by command
│   ├── test-server/        # Mock HTTP server for tests
│   └── *_test.go           # Integration test suites (one per command)
├── mock/                   # Generated mocks for unit tests
├── debian/                 # Debian package configuration
├── docker/                 # Docker build files
├── .github/                # GitHub Actions workflows
├── go.mod                  # Go module definition
├── Makefile               # Build automation
└── CLAUDE.md              # Project documentation for Claude
```

## Directory Purposes

**cmd/:**
- Purpose: Entry points for executable binaries
- Contains: main.go files that construct and execute commands
- Key files: `cmd/confluent/main.go` (CLI binary entry point)

**internal/:**
- Purpose: Command implementations organized by domain (not importable by other projects)
- Contains: Cobra command definitions, RunE handlers, domain-specific logic
- Key files: `internal/command.go` (root command registration)

**internal/\<domain\>/:**
- Purpose: Implement all commands for a specific domain (kafka, flink, iam, etc.)
- Contains:
  - `command.go`: Top-level command (e.g., "kafka")
  - `command_<subcommand>.go`: Cloud subcommands (e.g., "kafka cluster")
  - `command_<subcommand>_onprem.go`: On-prem variants
  - `command_<subcommand>_test.go`: Unit tests
- Key pattern: One file per subcommand level

**pkg/cmd/:**
- Purpose: Command infrastructure shared by all commands
- Contains: PreRunner, CLICommand, AuthenticatedCLICommand, error catchers, flag helpers
- Key files:
  - `prerunner.go`: Authentication and setup lifecycle
  - `authenticated_cli_command.go`: Base for authenticated commands
  - `run_requirements.go`: Mode validation

**pkg/config/:**
- Purpose: CLI configuration and state management
- Contains: Config struct, context switching, credential encryption, persistence
- Key files:
  - `config.go`: Main config type and load/save logic
  - `context.go`: Active session state
  - `machine.go`: Machine-specific credential storage

**pkg/auth/:**
- Purpose: Authentication and credential management
- Contains: Token handlers, credential managers, client factories, SSO support
- Key files:
  - `auth_token_handler.go`: Token exchange and refresh
  - `login_credentials_manager.go`: Credential sourcing (env, keychain, config)
  - `mds_client.go`: On-prem MDS client factory

**pkg/ccloudv2/:**
- Purpose: Confluent Cloud v2 API client wrappers
- Contains: Resource-specific client methods, retry logic, response parsing
- Key files: API client for each Confluent Cloud resource type

**pkg/errors/:**
- Purpose: Error handling and message standardization
- Contains: Typed errors, error catchers, message constants, suggestion builders
- Key files:
  - `error_message.go`: All error message constants
  - `catcher.go`: Error transformation logic

**pkg/output/:**
- Purpose: Format command output for humans and machines
- Contains: Table rendering, JSON/YAML serialization, color support
- Key files:
  - `table.go`: Tabular output formatting
  - `printer.go`: Output to stdout/stderr

**test/:**
- Purpose: Integration tests using golden file pattern
- Contains: Test suites (one per command), mock server, golden files
- Key files:
  - `cli_test.go`: Test suite setup and runner
  - `test-server/`: Mock HTTP backend

**test/fixtures/output/:**
- Purpose: Expected output for integration tests
- Contains: Subdirectories mirroring command structure with .golden files
- Key pattern: `<command>/<subcommand>/<test-name>.golden`

## Key File Locations

**Entry Points:**
- `cmd/confluent/main.go`: CLI binary entry point
- `internal/command.go`: Root command construction

**Configuration:**
- `go.mod`: Go module dependencies
- `Makefile`: Build targets and tooling
- `.golangci.yml`: Linter configuration
- `.go-version`: Required Go version (1.25.7)

**Core Logic:**
- `pkg/cmd/prerunner.go`: Command lifecycle
- `pkg/auth/auth.go`: Authentication logic
- `pkg/config/config.go`: State management

**Testing:**
- `test/cli_test.go`: Integration test framework
- `Makefile`: test, unit-test, integration-test targets

## Naming Conventions

**Files:**
- `command.go`: Top-level command for a domain (e.g., `internal/kafka/command.go`)
- `command_<subcommand>.go`: Cloud subcommand (e.g., `command_cluster.go`)
- `command_<subcommand>_onprem.go`: On-prem variant (e.g., `command_cluster_onprem.go`)
- `command_<subcommand>_test.go`: Unit tests (e.g., `command_cluster_test.go`)
- `<name>_test.go`: Test files (package_test or integration tests)

**Directories:**
- `internal/<domain>`: Domain-aligned with CLI command hierarchy
- `pkg/<utility>`: Utility/infrastructure packages (lowercase, no hyphens in import)
- `internal/<hyphenated>`: Hyphens allowed for multi-word domains matching CLI commands

**Packages:**
- Import path: `github.com/confluentinc/cli/v4/<path>`
- Internal packages: Not importable outside this module
- Pkg packages: Reusable but CLI-specific

**Functions:**
- Public: PascalCase (exported)
- Private: camelCase (unexported)
- Factory: `New()`, `New<Type>()` (e.g., `NewAuthenticatedCLICommand`)
- Subcommand factories: `new<Subcommand>Command()` (e.g., `newClusterCommand()`)

## Where to Add New Code

**New Top-Level Command:**
- Create: `internal/<command>/command.go`
- Register: Add to `internal/command.go` NewConfluentCommand()
- Tests: `test/<command>_test.go` and `test/fixtures/output/<command>/`

**New Subcommand (Cloud):**
- Implementation: `internal/<domain>/command_<subcommand>.go`
- Register: Add to parent command in `internal/<domain>/command.go`
- Tests: Add test cases to `test/<domain>_test.go`
- Golden files: `test/fixtures/output/<domain>/<subcommand>/<test>.golden`

**New Subcommand (On-Prem):**
- Implementation: `internal/<domain>/command_<subcommand>_onprem.go`
- Annotation: Add `annotations.RunRequirement: annotations.RequireOnPremLogin` to command
- Both modes: Create both `command_<subcommand>.go` and `command_<subcommand>_onprem.go`

**New API Client Method:**
- Location: `pkg/ccloudv2/<resource>.go`
- Pattern: Add method to existing client or create new resource file
- Dependencies: Use generated SDK clients from go.mod

**New Error Message:**
- Location: `pkg/errors/error_message.go`
- Pattern: Add constant with `ErrorMsg` suffix (lowercase, no period)
- Suggestion: Add corresponding `Suggestions` constant (full sentences with periods)

**New Output Format:**
- Tables: Use `pkg/output.Table` in command RunE handler
- JSON/YAML: Use `output.SerializedOutput` interface
- Custom: Extend `pkg/output/` with new formatter

**Utilities:**
- Shared helpers: `pkg/utils/` (HTTP, file I/O, crypto)
- Command helpers: `pkg/cmd/` (flags, validation, autocompletion)
- Domain-specific: `pkg/<domain>/` (kafka, flink, schemaregistry)

## Special Directories

**.planning/:**
- Purpose: GSD planning documents (created by /gsd commands)
- Generated: Yes (by Claude)
- Committed: Yes (planning docs are part of workflow)

**dist/:**
- Purpose: Compiled binaries from build
- Generated: Yes (by make build)
- Committed: No (.gitignore)

**test/bin/:**
- Purpose: Test binaries with coverage instrumentation
- Generated: Yes (by make build-for-integration-test)
- Committed: No

**mock/:**
- Purpose: Generated mocks for unit testing
- Generated: Yes (by mockery tool)
- Committed: Yes (checked in for CI)

**debian/:**
- Purpose: Debian package configuration
- Generated: Partially (package metadata)
- Committed: Yes

**test/fixtures/:**
- Purpose: Test inputs and expected outputs
- Generated: Golden files updated with -update flag
- Committed: Yes (golden files are test assertions)

---

*Structure analysis: 2026-03-08*
