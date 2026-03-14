# Limitations and Constraints

This document describes the known limitations of Confluent CLI skills for Claude Code, the technical reasons for these constraints, and available workarounds.

## Authentication Limitations

### Browser-Based Login Not Supported

**What doesn't work:** The `--browser` flag for authentication flows.

**Why:** Skills execute CLI commands programmatically without access to a browser session. Browser-based OAuth flows require interactive user input that cannot be automated through the MCP protocol.

**Workaround:** Use CLI-based authentication before starting Claude Code:

```bash
# Authenticate via CLI first
confluent login

# Then use skills in Claude Code
```

Your authentication session persists across skill invocations. Skills inherit the active login from your CLI configuration.

### Login State Not Preserved Across Sessions

**What doesn't work:** Logging in through skills without prior CLI authentication.

**Why:** Skills invoke CLI commands using the existing authentication state stored in `~/.confluent/config.json`. They cannot create new authentication sessions.

**Workaround:** Maintain your login session using the CLI:

```bash
# Before using skills, ensure you're logged in
confluent login

# Check login state
confluent environment list
```

If your session expires, re-authenticate via CLI. Skills will automatically use the renewed session.

### Dual-Mode Detection

**What doesn't work:** Automatic switching between Confluent Cloud and Confluent Platform modes within a single session.

**Why:** The CLI operates in one mode at a time based on authentication state. Skills cannot change the active login context.

**Workaround:** Use separate terminal sessions or switch contexts manually:

```bash
# Switch to Confluent Cloud
confluent login

# Switch to Confluent Platform
confluent login --url https://platform-url

# Ask Claude to perform operations in current mode
```

The skill invocations will target whichever environment is currently active in your CLI configuration.

## Command Type Limitations

### Interactive TUI Commands

**What doesn't work:** Commands that launch interactive terminal user interfaces.

**Why:** Skills execute commands in non-interactive mode. Terminal UIs require user input and cannot be automated.

**Workaround:** Use command-line flags for non-interactive execution:

```bash
# TUI version (doesn't work in skills)
confluent kafka topic create

# Non-interactive version (works in skills)
confluent kafka topic create my-topic --partitions 6
```

Ask Claude to create resources with specific parameters. Skills will construct the appropriate non-interactive command.

### Commands Requiring stdin

**What doesn't work:** Commands that read input from standard input.

**Why:** Skills cannot provide interactive input to running processes.

**Workaround:** Use file-based input or flags:

```bash
# stdin version (doesn't work in skills)
confluent iam rbac role-binding create < binding.json

# File-based version (works in skills)
confluent iam rbac role-binding create --file binding.json
```

Provide configuration as files or inline parameters when requesting operations through skills.

### Streaming Commands

**What doesn't work:** Long-running commands that stream output continuously (e.g., `confluent kafka topic consume` without `--max-messages`).

**Why:** Skills have execution timeouts and expect commands to complete. Infinite streams cannot be captured.

**Workaround:** Use bounded operations:

```bash
# Streaming version (doesn't work in skills)
confluent kafka topic consume my-topic

# Bounded version (works in skills)
confluent kafka topic consume my-topic --max-messages 100
```

Request a specific number of messages when asking Claude to consume from topics.

## Output Limitations

### Formatted Output Only

**What doesn't work:** Raw streaming output or real-time updates.

**Why:** Skills capture complete command output after execution completes. Real-time streaming is not supported.

**Workaround:** Skills automatically format output in human-readable tables or JSON. For raw data, use the CLI directly:

```bash
# Direct CLI usage for raw output
confluent kafka topic consume my-topic --print-key
```

Skills excel at summarizing and formatting results for readability. For precise binary data or raw formats, invoke the CLI manually.

### Large Result Sets

**What doesn't work:** Operations returning thousands of results may be truncated or slow.

**Why:** Skills optimize for readability. Extremely large outputs may be summarized.

**Workaround:** Request filtered results:

```
Show me topics in my production cluster that have more than 100 partitions
```

Claude can filter and summarize large result sets before presenting them.

## Resource Limitations

### Timeout Constraints

**What doesn't work:** Operations that take longer than the skill timeout (typically 120 seconds).

**Why:** Skills have maximum execution time limits to prevent hanging.

**Workaround:** For long-running operations, use the CLI directly:

```bash
# Long-running cluster creation
confluent kafka cluster create production-cluster \
  --cloud aws \
  --region us-west-2 \
  --type dedicated
```

Skills work best for operations that complete quickly. For cluster provisioning or large data migrations, use direct CLI invocation.

### Concurrent Operation Limits

**What doesn't work:** Requesting multiple independent operations simultaneously may execute sequentially.

**Why:** Skills invoke CLI commands serially to maintain state consistency.

**Workaround:** Request batch operations where the CLI supports them:

```bash
# Batch deletion (if supported by command)
confluent kafka topic delete topic-1 topic-2 topic-3
```

Claude will use batch commands when available, but some operations will execute one at a time.

## Platform-Specific Limitations

### Windows Path Handling

**What doesn't work:** Skills may misinterpret Windows paths in some contexts.

**Why:** Path separators and escaping differ between platforms.

**Workaround:** Use forward slashes in paths or provide full absolute paths:

```
Create a connector using config from C:/configs/connector.json
```

Claude will normalize paths when invoking commands.

### Environment Variable Expansion

**What doesn't work:** Shell environment variables in skill requests are not expanded.

**Why:** Skills do not have access to your shell environment.

**Workaround:** Provide explicit values instead of environment variable references:

```
# Instead of: "Use cluster $PROD_CLUSTER"
# Use: "Use cluster lkc-123456"
```

## Feature-Specific Limitations

### Schema Registry Complex Operations

**What doesn't work:** Some advanced schema evolution operations may require manual CLI usage.

**Why:** Schema compatibility rules can have complex edge cases that require precise control.

**Workaround:** Use skills for common operations (register schema, check compatibility). For complex migrations, use the CLI with full configuration files.

### Connector Plugin Uploads

**What doesn't work:** Uploading custom connector plugins through skills.

**Why:** Large file uploads are not optimized for the skill execution model.

**Workaround:** Upload plugins via CLI:

```bash
confluent connect plugin upload --plugin-file my-connector.zip
```

Skills can configure and manage connectors after plugins are uploaded.

### Multi-Region Coordination

**What doesn't work:** Coordinating operations across multiple regions simultaneously.

**Why:** Skills execute in the context of a single CLI configuration.

**Workaround:** Perform operations sequentially by switching contexts:

```
List clusters in us-west-2, then list clusters in eu-west-1
```

Claude will execute region-specific operations in sequence.

## Working Within Constraints

Understanding these limitations helps you use skills effectively. General strategies:

- **Authenticate via CLI first** - Keep an active session before using skills
- **Use explicit parameters** - Provide specific values rather than interactive inputs
- **Request bounded operations** - Ask for limited results (top N items, specific date ranges)
- **Combine skills with direct CLI** - Use skills for exploration and discovery, CLI for complex operations

For common errors and diagnostic steps, see [TROUBLESHOOTING.md](TROUBLESHOOTING.md).
