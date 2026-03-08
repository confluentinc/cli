# Coding Conventions

**Analysis Date:** 2026-03-08

## Naming Patterns

**Files:**
- Top-level commands: `command.go`
- Subcommands: `command_<subcommand>.go` (e.g., `command_create.go`, `command_delete.go`)
- On-prem variants: `command_<subcommand>_onprem.go` (e.g., `command_cluster_list_onprem.go`)
- Tests: `command_<subcommand>_test.go` or `command_test.go`
- Package names match directory name with hyphens converted to no separator (e.g., `internal/api-key/` → `package apikey`)

**Functions:**
- camelCase for unexported: `newCreateCommand()`, `validArgs()`, `validArgsMultiple()`
- PascalCase for exported: `NewAuthenticatedCLICommand()`, `ValidateAndConfirm()`
- Factory methods: `New()` for top-level commands, `new<Subcommand>Command()` for subcommands
- Handler methods: Named after the command (e.g., `create()`, `delete()`, `describe()`)

**Variables:**
- camelCase for local and unexported: `environmentId`, `kafkaREST`, `userKey`
- PascalCase for exported: `Client`, `V2Client`, `MDSClient`
- Constants: SCREAMING_SNAKE_CASE in local scope, or PascalCase for exported

**Types:**
- Unexported command structs: `type command struct` or `type clusterCommand struct`
- Unexported output structs: `type createOut struct`, `type listOut struct`
- Exported types: PascalCase (`AuthenticatedCLICommand`, `PreRunner`)

**Error and Message Variables:**
- Error messages: `<Description>ErrorMsg` (e.g., `KafkaClusterNotFoundErrorMsg`)
- Suggestions: `<Description>Suggestions` (e.g., `KafkaClusterNotFoundSuggestions`)
- General messages: `<Description>Msg` (e.g., `useAPIKeyMsg`)
- Warning messages: Defined in `warning_message.go`

## Code Style

**Formatting:**
- Tool: `gofmt` (enforced via golangci-lint)
- Tab indentation (Go standard)
- No trailing whitespace

**Linting:**
- Tool: golangci-lint v1.64.8
- Config: `.golangci.yml`
- Enabled linters: dupword, gci, gofmt, goimports, gomoddirectives, govet, ineffassign, misspell, nakedret, nolintlint, nonamedreturns, prealloc, predeclared, unconvert, unparam, unused, usestdlibvars, whitespace
- Naked returns: Disallowed (max-func-lines: 0)
- Exclude directories: `mock/`, `pkg/flink/test/mock/`

## Import Organization

**Order (enforced by gci linter):**
1. Standard library packages
2. Default (third-party) packages
3. Confluent packages (`github.com/confluentinc/`)
4. CLI-specific packages (`github.com/confluentinc/cli/`)

**Pattern:**
```go
import (
    "fmt"
    "net/http"

    "github.com/spf13/cobra"

    apikeysv2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"

    pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
    "github.com/confluentinc/cli/v4/pkg/config"
    "github.com/confluentinc/cli/v4/pkg/errors"
)
```

**Import Aliases:**
- SDK packages often aliased: `apikeysv2`, `srcmv3`, `ccloudv1`
- Internal packages use short prefixes: `pcmd` (pkg/cmd), `pauth` (pkg/auth)

## Error Handling

**Patterns:**
- Use error wrapping with context
- Return early on errors
- Prefer typed errors for catchable conditions

**Error Creation:**

1. **Basic errors** - Use `fmt.Errorf()` with constants from `pkg/errors/error_message.go`:
```go
return fmt.Errorf(errors.KafkaClusterNotFoundErrorMsg, clusterId)
```

2. **Errors with suggestions** - Use `errors.NewErrorWithSuggestions()`:
```go
return errors.NewErrorWithSuggestions(
    fmt.Sprintf(errors.ResourceNotFoundErrorMsg, resourceId),
    fmt.Sprintf(errors.ResourceNotFoundSuggestions, resourceId),
)
```

3. **Wrapped errors** - Use `errors.NewWrapErrorWithSuggestions()`:
```go
return errors.NewWrapErrorWithSuggestions(err,
    fmt.Sprintf(errors.LookUpRoleErrorMsg, roleName),
    errors.LookUpRoleSuggestions,
)
```

4. **Typed errors** - Implement `CLITypedError` interface for reusable/catchable errors

**Error Message Format:**
- Lowercase first letter (unless proper noun)
- No period at end
- Short, one-line description
- Variable name ends with `ErrorMsg`

**Suggestions Format:**
- Full sentences
- Capitalized first letter
- End with period
- Separate multiple suggestions with newlines
- Variable name ends with `Suggestions`

**Error Handling in Commands:**
- All RunE and PreRunE functions return errors
- Use `errors.HandleCommon()` wrapper in command execution

## Logging

**Framework:** Custom logger in `pkg/log`

**Patterns:**
- Debug logging: `log.CliLogger.Debugf("message", args...)`
- Used sparingly, primarily for troubleshooting
- Not for user-facing output

**User-Facing Output:**
- Use `output.Printf()`, `output.ErrPrintln()` from `pkg/output`
- Format with color support: `output.Printf(c.Config.EnableColor, format, args...)`

## Comments

**When to Comment:**
- Complex logic requiring explanation
- TODO/FIXME/HACK for technical debt
- CLI-specific issue tracker references: `// CLI-1544: Warn users if...`
- Deprecation notices: `// TODO: This mapping is deprecated and will be removed in v5`

**Documentation Comments:**
- Package-level comments rare (most packages are internal commands)
- Exported types and functions: Standard Go doc comments
- Struct field tags for output formatting: `human:"API Key" serialized:"api_key"`

## Function Design

**Size:**
- No hard limit, but functions > 100 lines are rare
- Extract helper functions for readability

**Parameters:**
- Commands: `func (c *command) handler(cmd *cobra.Command, args []string) error`
- Cobra Args validators: `cobra.NoArgs`, `cobra.ExactArgs(1)`, `cobra.MinimumNArgs(1)`
- Helpers: Accept only required parameters, not entire command context

**Return Values:**
- Commands always return `error` (or nil)
- Helpers return `(result, error)` pattern
- Multi-valued deletions return `([]string, error)` for deleted IDs

**Receiver Patterns:**
- Command methods: Pointer receivers `(c *command)`
- Embed `*pcmd.AuthenticatedCLICommand` or `*pcmd.CLICommand` for access to SDK clients

## Module Design

**Command Structure:**
- Each command is a package in `internal/<command>/`
- Root command registration in `internal/command.go`
- Top-level command factory: `func New(prerunner pcmd.PreRunner) *cobra.Command`
- Subcommand factories: `func (c *command) newCreateCommand() *cobra.Command`

**Exports:**
- Only `New()` function exported per command package
- Everything else unexported (internal to command)

**Dual-Mode Support:**
- Cloud vs On-Prem commands use Cobra annotations: `annotations.RequireCloudLogin`, `annotations.RequireOnPremLogin`
- Conditional subcommand registration based on `cfg.IsCloudLogin()`
- Separate files for on-prem variants: `command_<subcommand>_onprem.go`

**Struct Embedding:**
```go
type command struct {
    *pcmd.AuthenticatedCLICommand  // Provides Client, V2Client, Context, etc.
}
```

**Client Initialization:**
- Lazy initialization via embedded struct methods
- `c.V2Client` - Confluent Cloud v2 API
- `c.MDSClient` - MDS (RBAC) for on-prem
- `c.GetKafkaREST()` - Kafka REST API
- `c.GetFlinkGatewayClient()` - Flink Gateway

## User-Facing String Formatting

**CLI Commands/Flags:**
- Use backticks: `` `confluent kafka cluster list` ``

**Resources (IDs, names):**
- Use double quotes: `"lkc-123456"`, `"my-cluster"`

**Combined:**
- Escape quotes in strings: `"Update cluster \"lkc-123456\" with \`confluent kafka cluster update lkc-123456\`."`

**Flag Descriptions:**
- Cannot contain backticks (pflag limitation)
- Use quotes for commands, flags, and resources

## Command Patterns

**Validation:**
- Use `cobra.CheckErr(cmd.MarkFlagRequired("flag-name"))` for required flags
- Use `cmd.MarkFlagsRequiredTogether("flag1", "flag2")` for dependent flags
- Custom validation in handler function before SDK calls

**Autocompletion:**
- Implement `validArgs` or `validArgsMultiple` functions
- Register with: `ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs)`
- Handle error cases gracefully (return nil/empty on error)

**Multiple Resource Deletion:**
- Use `cobra.MinimumNArgs(1)` instead of `cobra.ExactArgs(1)`
- Use `deletion.ValidateAndConfirm()` to check existence and prompt
- Use `deletion.Delete()` to delete all resources, collecting errors
- Use `deletion.DeleteWithoutMessage()` for custom messages

**Output Formatting:**
- Define output struct with struct tags: `` `human:"Display Name" serialized:"api_name"` ``
- Use `output.NewTable(cmd)` and `table.Add()` pattern
- Support `-o json` and `-o yaml` flags via `pcmd.AddOutputFlag(cmd)`

---

*Convention analysis: 2026-03-08*
