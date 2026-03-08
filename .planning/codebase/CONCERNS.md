# Codebase Concerns

**Analysis Date:** 2026-03-08

## Tech Debt

**Deprecated Field Mappings:**
- Issue: Availability zone mappings (`single-zone`/`multi-zone`) are deprecated but still in use
- Files: `internal/kafka/command_cluster.go`
- Impact: Breaking change scheduled for v5 - users relying on old values will break
- Fix approach: Add deprecation warnings to CLI output, update documentation to guide users to new `low`/`high` values, plan migration path for v5

**Incomplete Mock Implementations:**
- Issue: Mock implementations use `panic("implement me")` instead of proper stubs
- Files: `mock/confluent_current.go` (lines 106, 111)
- Impact: Tests cannot exercise code paths that use these methods, reducing test coverage
- Fix approach: Implement proper mock behavior or mark methods as explicitly unsupported with better error messages

**TODO Comments Indicating Missing Features:**
- Issue: Config flag missing for commands, forcing use of standard config file location
- Files: `test/connect_test.go` (lines 15, 50, 112), `test/kafka_test.go` (line 289), `test/ksql_test.go` (line 4), `test/api_key_test.go` (line 14)
- Impact: Reduced flexibility in testing and CI/CD environments
- Fix approach: Add `--config` flag or `CONFLUENT_CONFIG` environment variable support

**Relative Path Limitation in Connect Plugin Detection:**
- Issue: Connect worker config detection fails when workers started with relative paths unless CLI run in same directory
- Files: `internal/connect/install_utils.go` (line 338)
- Impact: Plugin installation can fail to detect running workers, leading to configuration errors
- Fix approach: Resolve relative paths against worker process working directory, or require absolute paths

**Incomplete Context Handling:**
- Issue: Extensive use of `context.Background()` instead of propagating request context
- Files: Found in 20+ files including `internal/cluster/command_*.go`, `internal/local/command_kafka*.go`, `internal/audit-log/*.go`
- Impact: Cannot properly cancel long-running operations, timeout handling is inconsistent
- Fix approach: Add context parameter to command interfaces, propagate from cobra.Command context

**HACK Comments in Test Infrastructure:**
- Issue: Non-POSIX shell parsing workaround in test framework
- Files: `test/cli_test.go` (line 212)
- Impact: Tests may not accurately reflect real-world shell behavior
- Fix approach: Use proper shell parsing library or restrict test syntax to POSIX-compatible subset

**Config Deletion Between Non-Workflow Tests:**
- Issue: Tests delete current config to isolate test cases
- Files: `test/cli_test.go` (line 238)
- Impact: Potential for test pollution and inconsistent test state
- Fix approach: Use isolated config directories per test via temp directories

**Deprecated Error Message Formatting:**
- Issue: Auto-login flag message marked for removal in V4
- Files: `pkg/errors/error_message.go` (line 155)
- Impact: User-facing breaking change when removed
- Fix approach: Remove in v4.0 release, update documentation

**Domain ID Migration Incomplete:**
- Issue: Deprecated `domain_id` field still present alongside `cert_id`
- Files: `pkg/ccstructs/scheduler_structs.go` (line 398)
- Impact: Confusion about which field to use, potential for inconsistent behavior
- Fix approach: Complete migration to cert_id, add validation warnings for domain_id usage

**Cobra Helper Function Unused:**
- Issue: `appendIfNotPresent` function marked for removal in v2
- Files: `pkg/help/cobra.go` (line 55)
- Impact: Dead code maintenance burden
- Fix approach: Remove in next major version

## Known Bugs

**Windows File Lock Issue:**
- Symptoms: Proto files remain locked after tests complete, cleanup fails
- Files: `pkg/serdes/serdes_test.go` (lines 28-38)
- Trigger: Running test suite on Windows
- Workaround: 2-second sleep added before cleanup, skip cleanup on Windows entirely

**HACK: Newline Needed for Test Output Parsing:**
- Symptoms: Test output parsing fails without explicit newline
- Files: `pkg/auth/login_credentials_manager_test.go` (line 208)
- Trigger: Running specific login credential tests
- Workaround: Manual `fmt.Println("")` added

**macOS File Descriptor Limit:**
- Symptoms: Integration tests fail with "too many open files" error
- Files: Documented in `CONTRIBUTING.md` (lines 36-42)
- Trigger: Running integration tests on macOS with default limits
- Workaround: Requires manual system configuration changes and restart

**Missing Text Input Support in Integration Tests:**
- Symptoms: Cannot test interactive text input features
- Files: `test/login_test.go` (line 319)
- Trigger: Attempting to test stdin input flows
- Workaround: Tests marked TODO, feature untested

**Deprecated Environment Variable Credentials:**
- Symptoms: Deprecated env var names still accepted but may cause confusion
- Files: `pkg/auth/login_credentials_manager_test.go` (lines 20-21)
- Trigger: Using old `CONFLUENT_PLATFORM_USERNAME` environment variable
- Workaround: Migration to new variable names incomplete

## Security Considerations

**Hardcoded Auth0 Port Requirement:**
- Risk: Auth0 integration requires hardcoded port instead of random port
- Files: `pkg/auth/mfa/auth_server.go`, `pkg/auth/sso/auth_server.go`
- Current mitigation: Port documented and required by Auth0
- Recommendations: Request Auth0 support dynamic redirect URIs or use port range with fallback

**Global Logger State:**
- Risk: Global `CliLogger` instance can leak sensitive information across commands
- Files: `pkg/log/logger.go` (lines 11-19)
- Current mitigation: UNSAFE_TRACE level for sensitive data
- Recommendations: Refactor to use context-scoped loggers, audit UNSAFE_TRACE usage

**Panic in Mock Code:**
- Risk: Panics in mock implementations could expose error handling gaps
- Files: `pkg/mock/*.go` - multiple files with panic-based mocks
- Current mitigation: Only used in test code
- Recommendations: Replace panics with proper error returns to catch bugs earlier

**Deprecated Credential Handling:**
- Risk: Multiple credential sources with unclear precedence
- Files: `pkg/config/credential.go`, `pkg/config/context.go`
- Current mitigation: Migration in progress
- Recommendations: Complete migration, remove deprecated credential storage, audit all credential flows

## Performance Bottlenecks

**Large Handler Files:**
- Problem: Test server handlers are very large (3000+ lines)
- Files: `test/test-server/networking_handlers.go` (3242 lines), `test/test-server/kafka_rest_router.go` (2447 lines)
- Cause: All handler logic in single file
- Improvement path: Split by resource type into separate files

**Large Store Test File:**
- Problem: Flink store tests are extremely large
- Files: `pkg/flink/internal/store/store_test.go` (2766 lines)
- Cause: Comprehensive testing without organization
- Improvement path: Split into focused test files by feature area

**Synchronous Sleep Calls:**
- Problem: Blocking sleeps in production code for polling
- Files: `pkg/flink/internal/store/store.go` (lines 263, 362), `pkg/flink/internal/store/store_onprem.go` (lines 254, 336)
- Cause: Simple polling implementation
- Improvement path: Use exponential backoff with context cancellation, consider event-driven approach

**Password Protection Complexity:**
- Problem: Very large password protection implementation
- Files: `pkg/secret/password_protection_plugin.go` (659 lines), `pkg/secret/password_protection_test.go` (1486 lines)
- Cause: Complex encryption logic with multiple providers
- Improvement path: Refactor into smaller composable components

## Fragile Areas

**Dual-Mode Command System:**
- Files: Commands throughout `internal/` with both cloud and on-prem variants
- Why fragile: Easy to forget to implement both modes, inconsistent feature parity
- Safe modification: Always check for `*_onprem.go` variant, test both login modes
- Test coverage: Integration tests cover both modes but easy to miss edge cases

**Test Golden File System:**
- Files: `test/fixtures/output/**/*.golden`
- Why fragile: Any output change breaks tests, requires manual update review
- Safe modification: Run tests with `-update` flag, carefully review diffs before committing
- Test coverage: Good coverage but brittle to formatting changes

**Local Services Management:**
- Files: `internal/local/command_service.go` (937 lines), `internal/local/command_services.go` (644 lines)
- Why fragile: Process management, file system state, platform-specific behavior
- Safe modification: Test on all platforms (macOS, Linux, Windows), verify cleanup on failure paths
- Test coverage: Limited platform-specific testing

**Flink Shell Application:**
- Files: `pkg/flink/app/application.go`, `pkg/flink/lsp/*.go`, `pkg/flink/internal/controller/*.go`
- Why fragile: Complex state management with goroutines, channels, and mutex locks
- Safe modification: Run with race detector, test cancellation paths, verify resource cleanup
- Test coverage: Good unit tests but complex concurrent behavior

**Kafka Topic Produce Command:**
- Files: `internal/kafka/command_topic_produce.go` (809 lines)
- Why fragile: Handles multiple serialization formats, schema registry integration, complex error handling
- Safe modification: Test with all serialization types (Avro, JSON, Protobuf), verify schema registry fallback
- Test coverage: Integration tests exist but format combinations are extensive

**RBAC Role Binding Commands:**
- Files: `internal/iam/command_rbac_role_binding.go` (646 lines), `internal/iam/command_rbac_role_binding_list.go` (537 lines)
- Why fragile: Complex permission logic, dual-mode support, principal type handling
- Safe modification: Test with all principal types, verify both cloud and on-prem modes
- Test coverage: Partial - some edge cases may be untested

## Scaling Limits

**Integration Test File Descriptors:**
- Current capacity: Default macOS limit is 256 open files
- Limit: Integration tests exceed this during full suite run
- Scaling path: Documented workaround requires system-level changes, consider reducing concurrent file access in tests

**In-Memory Test Server:**
- Current capacity: All test fixtures stored in memory
- Limit: Large test suites with extensive fixtures consume significant memory
- Scaling path: Consider lazy loading of fixtures or external test data storage

**Flink Materialized Results:**
- Current capacity: Results buffered in memory
- Limit: Large result sets can exhaust memory
- Scaling path: Implement streaming or pagination for large result sets

## Dependencies at Risk

**Generated Protobuf Structs:**
- Risk: `pkg/ccstructs/scheduler_structs.go` contains generated code with XXX_ fields
- Impact: Protobuf format changes require regeneration, potential serialization breaks
- Migration plan: Document generation process, version proto definitions, test backward compatibility

**Cobra Shell Dependency:**
- Risk: `github.com/brianstrauch/cobra-shell` is relatively low-maintenance
- Impact: Shell mode could break with updates
- Migration plan: Consider vendoring or forking if maintenance becomes issue

**Multiple Cloud SDK Versions:**
- Risk: 50+ different `ccloud-sdk-go-v2` packages in dependencies
- Impact: Version skew between packages could cause API inconsistencies
- Migration plan: Coordinate updates across all SDK packages, test thoroughly after updates

## Missing Critical Features

**Config Flag for Commands:**
- Problem: Cannot specify config file location via flag
- Blocks: Testing in non-standard environments, CI/CD flexibility
- Priority: Medium

**Text Input in Integration Tests:**
- Problem: Cannot test interactive prompts with stdin
- Blocks: Testing save credential flows, prompt-based features
- Priority: Medium

**Comprehensive Error Context Propagation:**
- Problem: Many commands don't properly propagate context for cancellation
- Blocks: Graceful shutdown, request timeout enforcement
- Priority: High

**Windows Platform-Specific File Handling:**
- Problem: File locking and cleanup issues on Windows
- Blocks: Reliable testing and local development on Windows
- Priority: Medium

## Test Coverage Gaps

**Platform-Specific Code:**
- What's not tested: Windows-specific file handling edge cases
- Files: `pkg/serdes/serdes_test.go` - cleanup explicitly skipped on Windows
- Risk: Windows users may encounter file lock issues in production
- Priority: Medium

**On-Prem Command Variants:**
- What's not tested: Full parity testing between cloud and on-prem command variants
- Files: All `*_onprem.go` files throughout `internal/`
- Risk: Feature drift between cloud and on-prem implementations
- Priority: High

**Error Path Coverage:**
- What's not tested: Many error returns are not exercised in tests
- Files: Commands with complex error handling, especially in `internal/kafka/`, `internal/flink/`
- Risk: Silent failures or poor error messages in production
- Priority: High

**Concurrent Operation Safety:**
- What's not tested: Race conditions in Flink shell and local service management
- Files: `pkg/flink/app/application.go`, `internal/local/command_service.go`
- Risk: Goroutine leaks, deadlocks, or data races under concurrent use
- Priority: High

**Schema Registry Serialization Edge Cases:**
- What's not tested: All combinations of serialization formats with schema registry
- Files: `pkg/serdes/*.go`, `internal/kafka/command_topic_produce.go`
- Risk: Serialization failures with specific schema types
- Priority: Medium

**Mock Method Coverage:**
- What's not tested: Unimplemented mock methods prevent testing certain code paths
- Files: `mock/confluent_current.go` (GetConfigFileC3, WriteConfigC3)
- Risk: Features using these mocks are untestable
- Priority: Low

---

*Concerns audit: 2026-03-08*
