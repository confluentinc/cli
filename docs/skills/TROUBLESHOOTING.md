# Troubleshooting Guide

This guide provides diagnostic steps and solutions for common issues with Confluent CLI skills for Claude Code.

## Skills Not Loading

### Symptom

Claude Code does not respond to Confluent-related requests. Natural language questions about Kafka clusters, topics, or connectors receive generic responses instead of skill invocations.

### Diagnosis

The MCP server is not configured, or the CLI version is too old.

### Fix

1. **Check CLI version:**

   ```bash
   confluent version
   ```

   Verify the version is v3.45.0 or later. If older, update the CLI:

   ```bash
   # Homebrew
   brew upgrade confluentinc/tap/cli

   # APT
   sudo apt update && sudo apt upgrade confluent-cli

   # YUM
   sudo yum update confluent-cli
   ```

2. **Verify MCP configuration:**

   Open `~/.claude/config.json` and confirm the `confluent` server is configured:

   ```json
   {
     "mcpServers": {
       "confluent": {
         "command": "confluent",
         "args": ["mcp"]
       }
     }
   }
   ```

   If missing, add the entry and save the file.

3. **Restart Claude Code:**

   Completely quit and restart Claude Code to reload the MCP configuration.

4. **Test skills:**

   Ask a simple question: "List my Kafka clusters"

   If skills are working, Claude will invoke the CLI and return cluster information.

## Authentication Errors

### Symptom

Skill invocations fail with authentication errors such as "not logged in" or "invalid credentials."

### Diagnosis

The CLI does not have an active authentication session.

### Fix

1. **Check current authentication state:**

   ```bash
   confluent environment list
   ```

   If this command fails with an authentication error, you need to log in.

2. **Log in to Confluent Cloud:**

   ```bash
   confluent login
   ```

   Follow the prompts to authenticate.

3. **Log in to Confluent Platform:**

   ```bash
   confluent login --url https://your-platform-url
   ```

   Provide your credentials when prompted.

4. **Verify login succeeded:**

   ```bash
   confluent environment list
   ```

   If you see a list of environments, authentication is working.

5. **Retry skills:**

   Ask Claude to perform the operation again. Skills will use the active session.

### Session Expiration

If authentication worked previously but stopped, your session may have expired.

**Fix:** Re-authenticate using the steps above. Sessions expire after a period of inactivity.

## Command Not Found

### Symptom

Skills report that a specific command does not exist, even though you know the command is valid.

### Diagnosis

Version mismatch between the CLI and the skills manifest, or the command was removed in a recent CLI release.

### Fix

1. **Verify CLI version:**

   ```bash
   confluent version
   ```

   Ensure you have the latest stable release.

2. **Check command availability:**

   Test the command directly:

   ```bash
   confluent kafka cluster list
   ```

   If the command works via CLI but not through skills, the skills manifest may be outdated.

3. **Update CLI:**

   ```bash
   # Homebrew
   brew upgrade confluentinc/tap/cli

   # APT
   sudo apt update && sudo apt upgrade confluent-cli

   # YUM
   sudo yum update confluent-cli
   ```

   Newer CLI versions include updated skill manifests.

4. **Restart Claude Code:**

   After updating the CLI, restart Claude Code to reload the skills.

5. **Verify command format:**

   Some commands require specific flags. Check the CLI help:

   ```bash
   confluent kafka cluster --help
   ```

   Ensure your request includes required parameters.

## Timeout Errors

### Symptom

Long-running commands timeout before completing. Skills report an error after waiting for an extended period.

### Diagnosis

The operation exceeds the default skill timeout limit (typically 120 seconds).

### Fix

**For cluster provisioning or large migrations:**

Use the CLI directly. Skills are optimized for quick operations:

```bash
# Create dedicated cluster (may take several minutes)
confluent kafka cluster create production-cluster \
  --cloud aws \
  --region us-west-2 \
  --type dedicated
```

**No skill-based workaround exists for timeout limits.** For operations that take longer than 2 minutes, use direct CLI invocation.

**For data consumption:**

Request bounded results:

```
Consume 100 messages from my-topic
```

Instead of open-ended consumption which may timeout.

## Output Formatting Issues

### Symptom

Skill output appears garbled, contains ANSI codes, or is incomplete.

### Diagnosis

Edge case in output formatting or CLI output contains unexpected characters.

### Fix

1. **Test command directly:**

   Run the same command via CLI to see if the issue reproduces:

   ```bash
   confluent kafka topic list
   ```

   If the CLI output looks correct but skills format it incorrectly, this indicates a formatter issue.

2. **Check CLI version:**

   Ensure you're running the latest CLI version, as older versions may have output quirks:

   ```bash
   confluent version
   ```

3. **Report the issue:**

   If the problem persists, report it with the following information:

   - CLI version (`confluent version`)
   - Exact command that produces incorrect output
   - Sample of the incorrect output
   - Expected output format

   File an issue at: https://github.com/confluentinc/cli/issues

**Workaround:** Use the CLI directly for commands with formatting issues until the bug is fixed.

## Invalid Parameter Errors

### Symptom

Skills report invalid parameters or missing required flags when you believe the request is correct.

### Diagnosis

The natural language request did not map to valid CLI parameters, or required flags are missing.

### Fix

1. **Check command requirements:**

   View the help text for the command:

   ```bash
   confluent kafka topic create --help
   ```

   Note which flags are required.

2. **Provide explicit parameters:**

   Include all required information in your request:

   ```
   Create a topic called my-topic with 6 partitions in environment env-123456
   ```

   Instead of:

   ```
   Create a topic called my-topic
   ```

3. **Verify resource IDs:**

   Ensure you're using valid IDs for environments, clusters, etc. List resources first:

   ```
   List my environments
   ```

   Then use the correct ID in subsequent requests.

## Dual-Mode Confusion

### Symptom

Skills operate on the wrong environment (Confluent Cloud instead of Platform, or vice versa).

### Diagnosis

The CLI is logged into a different environment than expected.

### Fix

1. **Check current login context:**

   ```bash
   confluent context list
   ```

   The active context is marked with an asterisk.

2. **Switch contexts:**

   ```bash
   # For Confluent Cloud
   confluent login

   # For Confluent Platform
   confluent login --url https://platform-url
   ```

3. **Verify active context:**

   ```bash
   confluent environment list  # Cloud
   # or
   confluent cluster list       # Platform
   ```

4. **Retry skills:**

   Skills will now operate in the correct context.

**Limitation:** Skills cannot switch contexts automatically. You must change the active login via CLI before using skills. See [LIMITATIONS.md](LIMITATIONS.md#dual-mode-detection) for details.

## Connector Issues

### Symptom

Connector operations fail or produce unexpected results.

### Diagnosis

Connector plugins may not be installed, or connector configuration is invalid.

### Fix

1. **Verify plugin availability:**

   ```bash
   confluent connect plugin list
   ```

   Ensure the required connector plugin is installed.

2. **Validate connector configuration:**

   Test configuration via CLI:

   ```bash
   confluent connect cluster create --config-file connector.json --dry-run
   ```

   The dry-run flag validates configuration without creating the connector.

3. **Check connector status:**

   ```bash
   confluent connect cluster describe <connector-id>
   ```

   Look for error messages in the status output.

4. **Use explicit configuration:**

   When requesting connector creation through skills, provide complete configuration:

   ```
   Create a connector using config from /path/to/connector.json
   ```

## Schema Registry Compatibility Errors

### Symptom

Schema registration fails with compatibility errors.

### Diagnosis

The new schema violates the subject's compatibility rules.

### Fix

1. **Check current compatibility level:**

   ```bash
   confluent schema-registry compatibility level <subject>
   ```

2. **Test schema compatibility:**

   ```bash
   confluent schema-registry schema validate --subject <subject> --schema-file schema.avsc
   ```

3. **Adjust compatibility if needed:**

   ```bash
   confluent schema-registry compatibility update <subject> --level BACKWARD
   ```

   Valid levels: BACKWARD, FORWARD, FULL, NONE

4. **Retry registration:**

   Use skills or CLI to register the schema after resolving compatibility.

**For complex schema evolution:** Use the CLI directly with full control over compatibility settings. See [LIMITATIONS.md](LIMITATIONS.md#schema-registry-complex-operations).

## Debugging Steps

If none of the above solutions resolve your issue, follow this general diagnostic process:

### 1. Isolate the Problem

**Test via CLI first:**

```bash
# Try the command directly
confluent kafka cluster list
```

If the command fails via CLI, the issue is not skill-specific.

**Test via skills:**

Ask Claude to perform the same operation. If CLI succeeds but skills fail, the issue is skill-related.

### 2. Gather Information

Collect the following:

- CLI version: `confluent version`
- Operating system and version
- MCP configuration from `~/.claude/config.json`
- Exact natural language request that failed
- Error message or unexpected output

### 3. Check Recent Changes

- Did you recently update the CLI?
- Did you change authentication contexts?
- Did you modify MCP configuration?

### 4. Report Issue

If the problem persists after troubleshooting, file an issue at:

https://github.com/confluentinc/cli/issues

Include all information from step 2. Provide a minimal reproduction case if possible.

## Common Error Patterns

### "command exited with non-zero status"

**Cause:** CLI command failed. Check authentication and parameters.

**Fix:** Run command directly to see detailed error, then fix underlying issue.

### "skill execution timeout"

**Cause:** Operation took too long.

**Fix:** Use CLI for long-running operations. See [Timeout Errors](#timeout-errors).

### "environment not found"

**Cause:** Invalid environment ID or not logged in.

**Fix:** List environments (`confluent environment list`) to find valid IDs.

### "cluster not selected"

**Cause:** Command requires active cluster context.

**Fix:** Specify cluster ID explicitly in your request or use `confluent kafka cluster use <cluster-id>`.

### "permission denied"

**Cause:** Insufficient permissions for the operation.

**Fix:** Verify your account has required roles. Contact your administrator if needed.
