# MCP Server Limitations

This document describes known limitations of the Confluent CLI MCP (Model Context Protocol) server and provides workarounds where available.

## Browser-Based Login (EXEC-08)

**Issue:** Browser-based OAuth login flow (`confluent login`) cannot be invoked from MCP server.

**Why:**
- `confluent login` opens system browser for OAuth redirect
- MCP server typically runs headless or in background process
- Browser callback cannot be captured by MCP server
- Interactive prompts incompatible with MCP protocol

**Workaround:**

### Option A: API Key Authentication (Recommended)
1. Create API key in Confluent Cloud UI:
   - Navigate to: Cloud Console → Settings → API Keys
   - Click "Create Key" and select scope (global or resource-specific)
   - Save the API Key and Secret securely
2. Configure API key in `~/.confluent/config` before starting MCP server:
   ```
   [default]
   api_key = <your-api-key>
   api_secret = <your-api-secret>
   ```
3. Start MCP server - it inherits API key from config
4. API key auth is non-interactive and works from MCP context

### Option B: Pre-Login Session
1. Run `confluent login` manually in terminal before starting MCP server
2. Complete OAuth flow in browser
3. Start MCP server after successful login
4. MCP server inherits authentication from existing session
5. Note: Session may expire, requiring re-login

**Detection:**
- MCP server detects login commands without API key parameters
- Returns helpful error: "Browser-based login cannot be invoked from MCP server. Use API key authentication instead."
- Prevents hanging on browser prompt

**Future Enhancement:**
- Phase 7 documentation will include detailed API key setup guide
- Consider adding `--api-key` and `--api-secret` flags to login command for MCP compatibility

## Interactive Commands

**Issue:** Commands requiring stdin input (prompts, confirmations) will hang when invoked through MCP server.

**Why:**
- MCP protocol doesn't support interactive TTY input
- Commands waiting for user input appear to hang indefinitely
- Timeout may occur if ExecuteWithTimeout is used

**Workaround:**
- Use `--force` or `--yes` flags to skip confirmations where available
- Example: `confluent kafka topic delete --force` instead of interactive confirmation
- Check command help (`--help`) for available non-interactive flags

**Affected Commands:**
- Delete operations without `--force` flag
- Commands with confirmation prompts
- Commands requiring manual input (e.g., password prompts)

## Long-Running Commands

**Issue:** Commands without progress indicators appear to hang.

**Why:**
- MCP server captures stdout/stderr only after command completes
- No intermediate output during execution
- User has no indication command is progressing

**Workaround:**
- Use `ExecuteWithTimeout()` to enforce maximum duration
- Phase 5 will add `--timeout` flag to MCP server for global timeout
- Consider background execution for known long-running operations

**Affected Commands:**
- Large cluster operations (create, update, delete)
- Schema Registry operations on large schemas
- Connector deployments

## Non-JSON Output

**Issue:** Some commands don't support `--output json` flag.

**Why:**
- Certain commands (login, version, update) are informational only
- Output is meant for human consumption, not programmatic parsing
- No structured data to serialize

**Behavior:**
- Raw output returned as-is (not parsed or formatted)
- May contain ANSI color codes and formatting
- Phase 4 formatter will handle ANSI stripping and formatting

**Affected Commands:**
- `confluent version` - returns plain text version info
- `confluent login` - returns login status message
- `confluent update` - returns update status
- `confluent help` - returns help text

**Workaround:**
- Parse raw text output in skill layer
- Most data-oriented commands support `--output json`
- Check command help to verify JSON support

## Authentication Context

**Issue:** MCP server requires pre-existing authentication.

**Why:**
- ExecutionContext loads config at initialization
- No mechanism to authenticate mid-execution
- Session state shared across all skill invocations

**Implication:**
- User must be logged in before starting MCP server
- Use API key auth (Option A above) for reliable headless operation
- MCP server inherits authentication from CLI config

## Command Discovery

**Issue:** Tool names don't exactly match CLI command paths.

**Why:**
- Tool names use underscore separators: `confluent_kafka_cluster_list`
- CLI commands use spaces: `confluent kafka cluster list`
- Heuristic conversion in Phase 3 may have edge cases

**Workaround:**
- Phase 4 will add explicit `command_path` metadata to tools
- For now, `extractCommandPath()` handles common patterns
- Report issues if tool name → command path conversion fails

**Known Patterns:**
- Hyphenated resources: `service_account` → `service-account`
- Nested namespaces: `kafka_topic_list` → `kafka topic list`
- See `pkg/mcp/server.go` for full conversion logic

## Configuration Management

**Issue:** MCP server uses single fixed config context.

**Why:**
- ExecutionContext loads config once at startup
- No support for switching contexts mid-execution
- All skills execute in same config context

**Implication:**
- Cannot switch between Cloud/Platform mid-execution
- Cannot switch between different environments
- Restart MCP server to change config context

**Future Enhancement:**
- Phase 6 may add context switching support
- Consider per-skill context parameter

## Performance

**Issue:** Each skill execution creates new CLI command execution.

**Why:**
- No command result caching
- No connection pooling to Confluent APIs
- Each skill invocation is independent

**Implication:**
- Repeated calls to same command re-fetch data
- API rate limits may be hit with high-frequency calls
- Consider caching at skill layer if needed

**Future Enhancement:**
- Phase 6 may add result caching
- Phase 7 may add connection pooling
