# Architecture

**Analysis Date:** 2026-03-08

## Pattern Overview

**Overall:** Dual-Mode Command-Line Application with Plugin Architecture

**Key Characteristics:**
- Dual-mode operation supporting both Confluent Cloud (SaaS) and Confluent Platform (on-prem)
- Cobra-based hierarchical command structure with consistent PreRun lifecycle
- Lazy client initialization via factory pattern
- Multi-tenant authentication with automatic credential management and token refresh
- Plugin-based extensibility for third-party commands

## Layers

**Command Layer:**
- Purpose: Define CLI command structure and handle user interaction
- Location: `internal/`
- Contains: Top-level command definitions, subcommand handlers, flag parsing, validation
- Depends on: pkg/cmd (command infrastructure), pkg/auth (authentication), pkg/config (state)
- Used by: Entry point (cmd/confluent/main.go)

**Command Infrastructure Layer:**
- Purpose: Provide reusable command lifecycle and client management
- Location: `pkg/cmd/`
- Contains: PreRunner system, CLICommand/AuthenticatedCLICommand base types, error catching
- Depends on: pkg/config, pkg/auth, pkg/ccloudv2 (API clients)
- Used by: All commands in internal/

**Authentication Layer:**
- Purpose: Manage login credentials, token lifecycle, and dual-mode authentication
- Location: `pkg/auth/`
- Contains: Credential managers, token handlers, MDS client factories, SSO support
- Depends on: pkg/config, pkg/jwt, external SDKs (ccloud-sdk, mds-sdk)
- Used by: pkg/cmd (PreRunner), internal/login, internal/logout

**Configuration Layer:**
- Purpose: Persist CLI state across invocations
- Location: `pkg/config/`
- Contains: Config file management, context switching, credential encryption, machine state
- Depends on: pkg/secret (encryption), pkg/errors
- Used by: All layers (injected from main)

**API Client Layer:**
- Purpose: Abstract HTTP communication with Confluent APIs
- Location: `pkg/ccloudv2/`, external SDKs
- Contains: Retryable HTTP clients, resource-specific API wrappers, response handling
- Depends on: pkg/config (auth tokens), external API SDKs
- Used by: Command implementations in internal/

**Output Layer:**
- Purpose: Format command output for human and machine consumption
- Location: `pkg/output/`
- Contains: Table formatting, JSON/YAML serialization, color output
- Depends on: None (leaf layer)
- Used by: All command implementations

**Error Handling Layer:**
- Purpose: Standardize error messages and suggestions
- Location: `pkg/errors/`
- Contains: Typed errors, error catching, message formatting, suggestion system
- Depends on: None (leaf layer)
- Used by: All layers

## Data Flow

**Command Execution Flow:**

1. User invokes CLI with arguments (e.g., `confluent kafka topic list`)
2. `cmd/confluent/main.go` loads config from disk, constructs root command tree
3. Cobra routes to appropriate command in `internal/kafka/command_topic.go`
4. **PersistentPreRunE** executes (via PreRunner):
   - Anonymous PreRun: Parse flags, check verbosity, validate run requirements
   - Authenticated PreRun: Decrypt credentials, auto-login if needed, validate JWT token
   - ParseFlagsIntoContext: Store flag values in context for subcommands
5. **RunE** handler executes command logic:
   - Lazily initialize API client (c.V2Client, c.GetKafkaREST, etc.)
   - Make API calls via client layer
   - Format results via output layer
   - Return error (caught and formatted by CatchErrors wrapper)
6. **PersistentPostRun** reports usage telemetry (if cloud login)
7. Exit with appropriate code

**Authentication Flow:**

1. User runs `confluent login` with credentials
2. AuthTokenHandler.GetCCloudTokens() exchanges credentials for JWT
3. PersistCCloudCredentialsToConfig() stores encrypted tokens in ~/.confluent/config.json
4. Context is set to current org/env
5. Subsequent commands auto-refresh expired tokens via PreRunner.updateToken()

**Dual-Mode Routing:**

1. Command annotations define run requirements (cloud-login, on-prem-login, etc.)
2. Config.IsCloudLogin() / IsOnPremLogin() checks active context
3. ErrIfMissingRunRequirement() validates command can run in current mode
4. Commands with both modes use separate files: `command_X.go` (cloud), `command_X_onprem.go` (on-prem)
5. Client initialization branches on login mode (CCloud API vs MDS API)

**State Management:**
- Config persisted to `~/.confluent/config.json` after mutations (login, context switch)
- Context tracks: current org, environment, Kafka cluster, Schema Registry endpoint
- Credentials encrypted with OS keychain integration
- State updates synchronous (Save() after each mutation)

## Key Abstractions

**PreRunner:**
- Purpose: Encapsulate command lifecycle hooks for authentication and setup
- Examples: `pkg/cmd/prerunner.go`
- Pattern: Interface with multiple implementations (Anonymous, Authenticated, AuthenticatedWithMDS)

**CLICommand / AuthenticatedCLICommand:**
- Purpose: Base command types with embedded Cobra command and shared state
- Examples: `pkg/cmd/cli_command.go`, `pkg/cmd/authenticated_cli_command.go`
- Pattern: Composition (embed *cobra.Command) with lazy client getters

**Context:**
- Purpose: Represent active login session and resource selection state
- Examples: `pkg/config/context.go`
- Pattern: State object with nested contexts (environment, Kafka cluster, Flink)

**Client Factories:**
- Purpose: Decouple client construction from command logic
- Examples: `CCloudClientFactory`, `MDSClientManager`, `KafkaRESTProvider`
- Pattern: Function-type providers that lazily construct clients with auth

**Command Factory Pattern:**
- Purpose: Build command trees with consistent structure
- Examples: All `internal/*/command.go` files with New() functions
- Pattern: Factory functions that construct cobra.Command, register subcommands, attach PreRunner

## Entry Points

**CLI Binary:**
- Location: `cmd/confluent/main.go`
- Triggers: Direct user invocation
- Responsibilities: Load config, construct root command, execute Cobra, handle panics

**Root Command:**
- Location: `internal/command.go` (NewConfluentCommand)
- Triggers: Called by main.go
- Responsibilities: Register all top-level commands, setup PreRunner dependencies, configure help

**Plugin Entry:**
- Location: `pkg/plugin/plugin.go` (FindPlugin, ExecPlugin)
- Triggers: Command not found in built-in tree
- Responsibilities: Search for external executables named `confluent-<command>`, exec with remaining args

**Test Entry:**
- Location: `test/cli_test.go` (TestCLI suite)
- Triggers: `go test ./test`
- Responsibilities: Build instrumented binary, start mock server, run golden file tests

## Error Handling

**Strategy:** Centralized error catching with typed errors and user-friendly suggestions

**Patterns:**
- All RunE functions return error (never panic in user code)
- CatchErrors wrapper transforms generic errors into typed errors with suggestions
- Typed errors (NotLoggedInError, RunRequirementError) trigger specific flows (auto-login, mode check)
- Error messages lowercase without period, suggestions full sentences with period
- errors.NewErrorWithSuggestions() combines error + actionable fix
- DisplaySuggestionsMessage() formats errors for terminal output

## Cross-Cutting Concerns

**Logging:**
- Centralized via `pkg/log.CliLogger`
- Verbosity controlled by -v flags (warn, info, debug, trace, unsafe-trace)
- Unsafe-trace logs HTTP request/response bodies (may contain secrets)
- Debug logs auto-flushed for troubleshooting

**Validation:**
- Flag validation via Cobra (ExactArgs, MinimumNArgs, ValidArgsFunction)
- Run requirement validation via annotations (RequireCloudLogin, RequireOnPremLogin)
- Auto-completion via ValidArgsFunction with dynamic fetching from APIs
- Multi-delete validation via `pkg/deletion` helpers (existence check, confirmation prompt)

**Authentication:**
- JWT-based for Confluent Cloud (access token + refresh token)
- Bearer token for Confluent Platform MDS
- Auto-login on token expiry via non-interactive credentials (env vars, keychain)
- Dataplane tokens for Kafka REST and Flink Gateway (separate from auth token)
- Token refresh transparent to commands (handled in PreRunner)

---

*Architecture analysis: 2026-03-08*
