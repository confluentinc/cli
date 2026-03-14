# Installation Guide

This guide covers the installation and configuration of Confluent CLI skills for Claude Code.

## Prerequisites

Before installing the skills, ensure you have:

- **Confluent CLI v3.45.0 or later** - Check your version with `confluent version`
- **Claude Code** - The official CLI for Claude from Anthropic
- **Authentication** - Active login to Confluent Cloud or Confluent Platform

### Verify CLI Version

Check that your CLI version meets the minimum requirement:

```bash
confluent version
```

If your version is older than v3.45.0, update the CLI using your package manager:

```bash
# Homebrew (macOS/Linux)
brew upgrade confluentinc/tap/cli

# APT (Ubuntu/Debian)
sudo apt update && sudo apt upgrade confluent-cli

# YUM (RHEL/CentOS)
sudo yum update confluent-cli
```

### Verify Authentication

Skills require an active CLI authentication session. Verify your login state:

```bash
# For Confluent Cloud
confluent environment list

# For Confluent Platform
confluent login --url https://your-platform-url
```

If you see authentication errors, log in before proceeding:

```bash
# Confluent Cloud
confluent login

# Confluent Platform
confluent login --url https://your-platform-url
```

## MCP Server Configuration

Claude Code uses the Model Context Protocol (MCP) to integrate with external tools. Configure the Confluent CLI as an MCP server by editing your Claude Code configuration file.

### Configuration File Location

The configuration file location varies by platform:

- **macOS**: `~/.claude/config.json`
- **Linux**: `~/.claude/config.json`
- **Windows**: `%USERPROFILE%\.claude\config.json`

### Add MCP Server Entry

Open your `config.json` file and add the Confluent CLI server to the `mcpServers` section:

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

If you already have other MCP servers configured, add the `confluent` entry alongside them:

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/allowed/files"]
    },
    "confluent": {
      "command": "confluent",
      "args": ["mcp"]
    }
  }
}
```

### Platform-Specific Notes

#### macOS

Ensure the `confluent` binary is in your PATH. If installed via Homebrew, this is typically `/usr/local/bin/confluent` (Intel) or `/opt/homebrew/bin/confluent` (Apple Silicon).

#### Linux

The package managers install `confluent` to `/usr/bin/confluent`, which is in the PATH by default.

#### Windows

If you installed the CLI from the ZIP file, provide the full path to the executable:

```json
{
  "mcpServers": {
    "confluent": {
      "command": "C:\\path\\to\\confluent.exe",
      "args": ["mcp"]
    }
  }
}
```

## Verification

After updating your configuration, restart Claude Code to load the MCP server.

### Test Skills Loading

Verify that Claude Code can access Confluent CLI skills by asking a simple question:

```
List my Kafka clusters
```

Claude should invoke the Confluent CLI skills and return your cluster list. If you see cluster information formatted as a response, the skills are working correctly.

### Verification Checklist

Confirm each step:

- [ ] CLI version is v3.45.0 or later (`confluent version`)
- [ ] Active authentication session (`confluent environment list` or similar command succeeds)
- [ ] MCP server added to `~/.claude/config.json`
- [ ] Claude Code restarted
- [ ] Skills respond to natural language requests

## Troubleshooting

If skills do not load or respond correctly, see [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for common issues and resolution steps.

Common installation issues include:

- **Skills Not Loading** - MCP server configuration missing or incorrect
- **Authentication Errors** - Not logged in via CLI
- **Command Not Found** - CLI not in PATH or wrong version

Detailed diagnostics and fixes are available in the troubleshooting guide.
