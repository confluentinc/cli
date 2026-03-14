# pkg/mcp - MCP Runtime Execution Layer

The `mcp` package provides the runtime execution layer for Confluent CLI skills, enabling Claude Code to invoke CLI commands through the Model Context Protocol (MCP).

## Architecture

The package consists of four main components:

```
┌─────────────────┐
│   MCPServer     │  Loads skill manifest, routes tool executions
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ExecutionContext │  Manages CLI config and authentication
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│    Executor     │  Executes commands in-process with output capture
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│     Mapper      │  Maps skill parameters to CLI flags with validation
└─────────────────┘
```

## Components

### MCPServer (`server.go`)

- **Purpose**: Skill manifest loading and execution routing
- **Key methods**:
  - `NewMCPServer(manifestPath)`: Loads skills.json and builds tool lookup map
  - `ExecuteSkill(toolName, params)`: Routes skill execution to ExecutionContext
  - `Manifest()`: Returns loaded skill manifest for inspection
- **Browser login detection**: Implements EXEC-08 by detecting `confluent_login` tool without API key parameters

### ExecutionContext (`context.go`)

Manages CLI configuration and authentication context across multiple skill invocations.

#### Authentication Context (EXEC-04)

**How it works:**

1. `NewExecutionContext()` loads CLI config once from `~/.confluent/config`
2. Config contains authentication credentials:
   - **Cloud mode**: JWT tokens from `confluent login` or API key/secret
   - **Platform mode**: MDS credentials from `confluent login --url`
3. Config is reused across all `Execute()` calls
4. Authentication state persists for the lifetime of ExecutionContext

**Why this matters:**

- Avoids re-authentication overhead on every skill invocation
- Preserves session tokens that would be lost with subprocess execution
- Enables skills to work with existing authenticated CLI sessions

**Config location**: `~/.confluent/config` (standard CLI config path)

**Thread safety**: Config struct is immutable after loading. Executor rebuilds command tree per invocation to avoid flag pollution, but shares the same Config instance for auth.

#### Dual-Mode Support (EXEC-05)

**Cloud vs Platform detection:**

The ExecutionContext transparently supports both Confluent Cloud and Confluent Platform modes based on the loaded config:

- **Cloud mode**: Config has `platform: false` or no platform URL
  - Uses Cloud API endpoints
  - Authenticates with JWT or API key
  - Commands annotated with `RequireCloudLogin` work

- **Platform mode**: Config has `platform: true` and MDS URL
  - Uses Platform endpoints
  - Authenticates with MDS credentials
  - Commands annotated with `RequireNonCloudLogin` work

- **Dual-mode commands**: Some commands work in both modes (e.g., `version`)

**Accessing mode**: Call `ctx.Config()` to inspect current context and mode.

**No explicit mode switching**: Mode is determined by the loaded config. To switch modes, user must run `confluent login` (Cloud) or `confluent login --url <platform-url>` (Platform) before starting MCP server.

#### Timeout Handling (EXEC-06)

**Methods:**

- `Execute(commandPath, params)`: No timeout (runs until completion or error)
- `ExecuteWithTimeout(commandPath, params, timeout)`: Context-based cancellation

**How timeouts work:**

1. Creates `context.WithTimeout(timeout)`
2. Passes context to command execution (future: will use ctx for cancellation)
3. If timeout expires, returns `context.DeadlineExceeded` error
4. Gracefully cleans up without leaving zombie processes (in-process execution)

**Timeout recommendations:**

| Command Type | Recommended Timeout | Rationale |
|--------------|-------------------|-----------|
| Metadata queries (list, describe) | 30s | Cloud API latency + pagination |
| Create operations | 60s | Resource provisioning time |
| Delete operations | 45s | Cleanup + API confirmation |
| Version/help commands | 5s | Local operations only |
| Login operations | N/A | Use pre-auth (see LIMITATIONS.md) |

**Zero timeout**: `ExecuteWithTimeout(cmd, params, 0)` disables timeout enforcement (runs until completion).

**Future enhancement**: Pass context through to Cobra command execution for mid-execution cancellation (currently timeout is enforced at wrapper level).

### Executor (`executor.go`)

Executes CLI commands in-process without subprocess overhead.

**Key features:**

- **In-process execution**: Builds Cobra command tree via `internal.NewConfluentCommand(cfg)` (EXEC-01)
- **Thread-safe output capture**: Uses `outputMutex` to protect stdout/stderr redirection during concurrent executions
- **Fresh command tree**: Rebuilds command tree per invocation to avoid flag pollution between executions
- **Output capture**: Redirects `os.Stdout` and `os.Stderr` to buffers, restores after execution
- **Error propagation**: Returns command errors with output for debugging

**Why in-process:**

- Preserves authentication context from config (no env var passing needed)
- Faster execution (no fork/exec overhead)
- Simpler error handling (direct Go error returns)
- Enables concurrent execution with mutex protection

### Mapper (`mapper.go`)

Maps skill parameters to CLI flags with validation.

**Key features:**

- **Pre-execution validation** (EXEC-07): `ValidateParams()` catches unknown/missing flags before `cmd.Execute()`
- **Type conversion**: Handles string, bool, int via `fmt.Sprint()` conversion
- **Required flag detection**: Uses `BashCompOneRequiredFlag` annotation from pflag
- **Nil value skipping**: Nil params use flag defaults instead of failing
- **Automatic JSON output** (EXEC-03): `ForceJSONOutput()` sets `--output=json` when flag exists

**Validation errors:**

```go
// Unknown parameter
"unknown parameter: invalid-flag"

// Missing required parameter
"required parameter missing: environment"

// Type mismatch (rare - fmt.Sprint handles most types)
"error setting flag cluster: invalid value"
```

## Usage Examples

### Basic Execution

```go
package main

import (
    "fmt"
    "log"
    "github.com/confluentinc/cli/v4/pkg/mcp"
)

func main() {
    // Load config and create execution context
    ctx, err := mcp.NewExecutionContext()
    if err != nil {
        log.Fatal(err)
    }

    // Execute version command (no parameters)
    output, err := ctx.Execute("version", nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(output)
}
```

### Execution with Parameters

```go
// List Kafka clusters in specific environment
params := map[string]interface{}{
    "environment": "env-12345",
    "output":      "json",
}

output, err := ctx.Execute("kafka cluster list", params)
if err != nil {
    log.Fatalf("Failed to list clusters: %v", err)
}
fmt.Println(output)
```

### Execution with Timeout

```go
import "time"

// Execute with 30-second timeout
output, err := ctx.ExecuteWithTimeout(
    "kafka topic list",
    map[string]interface{}{"cluster": "lkc-abc123"},
    30 * time.Second,
)
if err != nil {
    // Check if timeout occurred
    if err == context.DeadlineExceeded {
        log.Fatal("Command timed out after 30s")
    }
    log.Fatal(err)
}
```

### MCP Server Usage

```go
// Load skill manifest and create server
server, err := mcp.NewMCPServer("pkg/mcp/skills.json")
if err != nil {
    log.Fatal(err)
}

// Execute skill by tool name
output, err := server.ExecuteSkill("confluent_kafka_cluster_list", map[string]interface{}{
    "environment": "env-12345",
})
if err != nil {
    log.Fatalf("Skill execution failed: %v", err)
}

fmt.Println(output)
```

### Checking Available Skills

```go
server, _ := mcp.NewMCPServer("skills.json")

manifest := server.Manifest()
fmt.Printf("Loaded %d skills\n", manifest.Metadata.SkillCount)

for _, tool := range manifest.Tools {
    fmt.Printf("- %s: %s\n", tool.Name, tool.Description)
}
```

## Skills Development Workflow

### When Skills Need Regeneration

Regenerate skills when you modify CLI **command structure**:

- Adding/removing commands or subcommands
- Adding/removing flags
- Changing flag names, types, or descriptions
- Modifying command descriptions or help text
- Changing command annotations (cloud/on-prem requirements)

**Examples:**
```go
// Added new flag → regenerate
cmd.Flags().String("new-flag", "", "Description")

// Changed command description → regenerate
Use: "cluster list",
Short: "New description here",  // Changed

// Added new subcommand → regenerate
cmd.AddCommand(newSubCommand())
```

### When Skills Do NOT Need Regeneration

Skip regeneration for **implementation changes**:

- Changing command logic/business logic
- Modifying API client calls
- Updating error messages
- Refactoring internal functions
- Adding tests

**Examples:**
```go
// Changed implementation → no regeneration needed
func (c *command) list(cmd *cobra.Command) error {
    // New API call logic
    clusters, err := c.V2Client.ListClusters(ctx)
    // ...
}
```

### Manual Regeneration

Regenerate skills manually when needed:

```bash
# Regenerate skill manifest
make generate-skills

# Verify the generated manifest
make validate-skills

# Check tool count
jq '.metadata.tool_count' pkg/mcp/skills.json

# Check CLI version
jq '.metadata.cli_version' pkg/mcp/skills.json
```

## Embedded vs File Loading

### How Embedding Works

**Production builds** embed skills.json at compile time:

```go
//go:embed skills.json
var skillsJSON []byte

func LoadSkills() (*Manifest, error) {
    // Load from embedded data
    return json.Unmarshal(skillsJSON)
}
```

**Development builds** fall back to file loading if embed fails:

```go
// If embedded data is empty, try loading from file
if len(skillsJSON) == 0 {
    data, err := os.ReadFile("pkg/mcp/skills.json")
    return json.Unmarshal(data)
}
```

### Why Rebuild Is Needed

**go:embed freezes content at build time:**

1. You modify command files
2. Run `make generate-skills` → updates skills.json on disk
3. **But** previously built binary still has old embedded data
4. You must `make build` to re-embed the new skills.json

**When to rebuild:**

- After regenerating skills (`make generate-skills`)
- Before testing skill changes in MCP server
- Before committing (pre-commit hook does this automatically)

**Quick test cycle:**

```bash
# 1. Make command changes
vim internal/kafka/command_cluster.go

# 2. Regenerate and rebuild
make generate-skills && make build

# 3. Test with MCP server
./dist/confluent_*/mcp-server --manifest pkg/mcp/skills.json
```

## VS Code Integration

### Available Tasks

Press `Cmd+Shift+P` → "Tasks: Run Task" to access:

| Task | Purpose | When to Use |
|------|---------|-------------|
| **Generate Skills** | Regenerates skills.json | After modifying commands |
| **Validate Skills** | Validates manifest structure | Before committing |
| **Build CLI with Skills** | Full build pipeline | Before testing changes |
| **Clean Skills** | Removes generated files | When resetting state |

### Keyboard Shortcuts

Add to your `.vscode/keybindings.json` for faster access:

```json
[
  {
    "key": "cmd+shift+g",
    "command": "workbench.action.tasks.runTask",
    "args": "Generate Skills"
  },
  {
    "key": "cmd+shift+v",
    "command": "workbench.action.tasks.runTask",
    "args": "Validate Skills"
  }
]
```

### Task Outputs

Tasks show output in the integrated terminal:

```
> Executing task: make generate-skills <

🔨 Generating skills manifest...
✓ Generated 420 tools
✓ Validated manifest structure
✓ Wrote pkg/mcp/skills.json (1.2 MB)

Terminal will be reused by tasks, press any key to close it.
```

## Verification

### Check Tool Count

```bash
jq '.metadata.tool_count' pkg/mcp/skills.json
# Expected: 400-500 tools
```

### Check CLI Version

```bash
jq '.metadata.cli_version' pkg/mcp/skills.json
# Expected: matches version in version/version.go
```

### Validate Manifest

```bash
make validate-skills
# Expected: ✓ Validation passed (tool count: 420)
```

### Test Skill Execution

```bash
# Build with latest skills
make build

# Run MCP server
./dist/confluent_darwin_arm64/mcp-server --manifest pkg/mcp/skills.json &

# Test a skill (requires MCP client)
# Example: list Kafka clusters
```

## Troubleshooting

### "Validation failed: duplicate tool names"

**Cause:** IR generation created duplicate tools from similar commands.

**Solution:**

1. Check the intermediate representation:
   ```bash
   jq '.[] | select(.tool_name == "duplicate_name")' cmd/generate-skills/ir.json
   ```

2. Review command registration in `internal/command.go`
3. Ensure unique command paths (e.g., `kafka cluster list` vs `kafka topic list`)
4. Check for duplicate aliases or hidden commands

### "Tool count exceeds 200"

**Cause:** CLI has grown beyond the recommended MCP tool limit (currently ~420 tools).

**Solution:**

This is informational, not an error. The validation limit is set to 500. If tool count approaches 500:

1. Review grouping strategy in skill generator
2. Consider consolidating similar commands
3. Use command categories to reduce granularity
4. Discuss with team if restructuring is needed

**Current status:** 420 tools (within limits, but watch for growth)

### "Skills not embedded"

**Cause:** Built with `go build` instead of `make build`.

**Solution:**

Always use `make build`:

```bash
# Wrong - skills not embedded
go build -o confluent cmd/confluent/main.go

# Correct - skills embedded
make build
```

`pkg/mcp/skills.json` is committed in the repo and embedded via `go:embed` at build time. Run `make generate-skills` to update it when commands change.

### "Changes not reflected"

**Cause:** go:embed freezes content at build time.

**Solution:**

Rebuild after regenerating:

```bash
# Regenerate and rebuild
make generate-skills && make build

# Or use the combined task
make build  # Runs generate-skills first
```

**Why this happens:**

1. `make generate-skills` updates `pkg/mcp/skills.json` on disk
2. Previously built binary still has old embedded `//go:embed skills.json` data
3. Must rebuild to re-embed the new file

**Verification:**

```bash
# Check file on disk
jq '.metadata.generated_at' pkg/mcp/skills.json

# Check embedded version (requires running binary)
./dist/confluent_*/mcp-server --version
# Should show matching timestamp
```

## Known Limitations

See [LIMITATIONS.md](./LIMITATIONS.md) for detailed documentation of known constraints:

1. **Browser-based login** (EXEC-08): Cannot invoke `confluent login` from MCP server. Use API key authentication or pre-login.
2. **Interactive commands**: Commands requiring stdin (prompts, confirmations) will hang.
3. **TUI commands**: Terminal UI commands (e.g., `flink shell`) incompatible with output capture.
4. **Concurrent execution limits**: While thread-safe, extremely high concurrency may hit OS file descriptor limits.
5. **Output buffering**: Very large outputs (>100MB) may cause memory pressure.
6. **ANSI escape codes**: Raw output includes color codes. Phase 4 will add formatter to strip these.

## Testing

**Unit tests:**
```bash
go test ./pkg/mcp -run TestExecutor      # Executor tests
go test ./pkg/mcp -run TestMapper        # Mapper tests
go test ./pkg/mcp -run TestExecution     # Context tests
go test ./pkg/mcp -run TestMCPServer     # Server tests
```

**Integration tests:**
```bash
go test ./pkg/mcp -run TestIntegration   # End-to-end with real CLI
```

**Coverage:**
```bash
go test -cover ./pkg/mcp
# Current coverage: 93.6%
```

## Requirements Satisfied

This package implements the following Phase 3 requirements:

- **EXEC-01**: In-process CLI execution without subprocess overhead
- **EXEC-02**: Skill parameter to CLI flag mapping with type conversion
- **EXEC-03**: Automatic `--output json` forcing on all invocations
- **EXEC-04**: Authentication context loading and persistence across invocations
- **EXEC-05**: Dual-mode support (Confluent Cloud and Platform) via config detection
- **EXEC-06**: Command timeout enforcement with graceful cancellation
- **EXEC-07**: Parameter validation before CLI invocation (unknown/missing/type errors)
- **EXEC-08**: Browser-based login detection and rejection with helpful error message

## Future Enhancements

**Phase 4 (Output Formatting):**
- ANSI escape code stripping
- JSON to human-readable summary conversion
- Markdown table rendering
- Resource ID shortening

**Phase 5 (Build Integration):**
- Embed skill manifest in CLI binary
- Version-locked skill distribution

**Phase 6 (Testing):**
- Cross-platform validation (macOS, Linux, Windows)
- Multi-version compatibility testing
- Error scenario coverage expansion

## See Also

- [../skillgen/](../skillgen/) - Skill manifest generation from CLI command tree
- [../../cmd/mcp-server/](../../cmd/mcp-server/) - Runnable MCP server binary
- [LIMITATIONS.md](./LIMITATIONS.md) - Known constraints and workarounds
- [../../.planning/phases/03-execution-runtime/](../../.planning/phases/03-execution-runtime/) - Phase 3 planning documents
