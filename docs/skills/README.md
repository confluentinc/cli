# Claude Code Skills for Confluent CLI

The Confluent CLI includes Claude Code skills that provide a natural language interface to Confluent Cloud and Confluent Platform. Skills enable you to manage deployments through conversational requests, without memorizing command syntax or flags.

## Overview

Claude Code skills translate your natural language requests into properly formatted CLI commands. Instead of looking up syntax in documentation, you can describe what you want to accomplish, and Claude will invoke the appropriate skills to execute your request.

**Example:**
- You: "List my Kafka clusters"
- Claude: Invokes `confluent_kafka_cluster_list` skill
- Result: Formatted table of your Kafka clusters

Skills are organized by CLI command structure, corresponding to the `confluent` command namespaces (kafka, connect, flink, schema-registry, etc.).

## How to Discover Skills

There are three ways to discover available skills:

### 1. Ask Claude Directly

The easiest method is to ask Claude what skills are available:

- "What Confluent CLI skills are available?"
- "What skills can help me manage Kafka topics?"
- "Show me all flink-related skills"

Claude will list relevant skills and explain their usage.

### 2. Browse the Skill Reference

The [REFERENCE.md](REFERENCE.md) file contains a complete catalog of all available skills, organized by namespace. Each skill listing includes:

- Description with usage examples
- Required parameters
- Optional parameters with defaults
- Priority level (high, medium, low)

This is useful for comprehensive exploration or when you want to understand all capabilities within a specific namespace.

### 3. Claude Code Skill Browser (if available)

If your Claude Code UI includes a skill browser panel:

1. Open the skills panel
2. Search for "confluent"
3. Browse skills organized by category
4. View detailed parameter documentation

## Skill Organization

Skills mirror the Confluent CLI command structure:

- **Kafka skills** → `confluent kafka` commands (clusters, topics, ACLs, consumer groups)
- **Connect skills** → `confluent connect` commands (clusters, connectors, plugins)
- **Flink skills** → `confluent flink` commands (compute pools, statements, artifacts)
- **Schema Registry skills** → `confluent schema` commands (schemas, subjects, exporters)
- **IAM skills** → `confluent iam` commands (roles, bindings, service accounts)
- **Network skills** → `confluent network` commands (peering, private links, DNS)
- **Environment skills** → `confluent environment` commands (create, list, update)
- **API Key skills** → `confluent api-key` commands (create, list, delete)

And many more. See [REFERENCE.md](REFERENCE.md) for the complete catalog.

## How Skills Work

**You don't invoke skills explicitly.** Claude automatically selects and invokes the appropriate skills based on your natural language request.

**Example workflow:**

1. You: "Create a Kafka topic called 'orders' with 6 partitions in cluster lkc-123456"
2. Claude recognizes this requires the `confluent_kafka_topic_create` skill
3. Claude invokes the skill with parameters:
   - `topic`: "orders"
   - `partitions`: 6
   - `cluster`: "lkc-123456"
4. The CLI executes and returns results
5. Claude formats the output for you

This natural language interface eliminates the need to remember flag names, parameter formats, or command syntax.

## Quick Start

### Prerequisites

- Confluent CLI v3.45.0 or later installed
- Claude Code (https://claude.ai/code)
- Authenticated via `confluent login` (Cloud) or `confluent login --url` (Platform)

### MCP Configuration

Add the Confluent CLI MCP server to your Claude Code configuration (`~/.claude/config.json` or equivalent):

```json
{
  "mcpServers": {
    "confluent": {
      "command": "confluent",
      "args": ["mcp", "start"],
      "env": {}
    }
  }
}
```

Restart Claude Code after updating the configuration.

### First Example

Once configured, try a simple request:

**You:** "List my Kafka clusters"

**Claude will:**
1. Invoke the `confluent_kafka_cluster_list` skill
2. Execute `confluent kafka cluster list --output json`
3. Format the results as a readable table showing:
   - Cluster ID
   - Name
   - Type (BASIC, STANDARD, DEDICATED)
   - Cloud provider and region
   - Status

**Sample output:**
```
Found 3 Kafka clusters:

| ID          | Name         | Type      | Cloud | Region    | Status |
|-------------|--------------|-----------|-------|------------ |--------|
| lkc-123456  | prod-cluster | DEDICATED | aws   | us-west-2 | UP     |
| lkc-234567  | dev-cluster  | BASIC     | gcp   | us-east1  | UP     |
| lkc-345678  | staging      | STANDARD  | azure | eastus    | UP     |
```

## Additional Resources

- **[INSTALLATION.md](INSTALLATION.md)** - Detailed setup instructions, including troubleshooting for various environments
- **[USAGE.md](USAGE.md)** - 8-10 complete workflow examples covering common tasks (topic management, connector deployment, schema evolution, etc.)
- **[LIMITATIONS.md](LIMITATIONS.md)** - Known constraints and unsupported operations
- **[TROUBLESHOOTING.md](TROUBLESHOOTING.md)** - Solutions for common issues (authentication errors, skill discovery failures, output formatting problems)
- **[REFERENCE.md](REFERENCE.md)** - Complete skill catalog with all 420+ skills organized by namespace

## Skill Priority

Skills are tagged with priority levels to help Claude select the most appropriate tool:

- **High priority**: Core operations (list, create, describe, delete)
- **Medium priority**: Secondary operations (update, configure, pause/resume)
- **Low priority**: Utility operations (shell, completion, feedback)

Claude uses priority along with your request context to select the best skill for your task.

## Authentication

Skills inherit authentication from your CLI login state:

- **Confluent Cloud**: Requires `confluent login` (browser-based OAuth)
- **Confluent Platform**: Requires `confluent login --url <mds-url>` (username/password or API key)

The MCP server uses your active context (see `confluent context list`). Switch contexts with `confluent context use <name>` if needed.

## Output Formats

Skills default to structured output formats (JSON/YAML) which Claude then formats for readability. You can request specific formats:

- "List Kafka topics as a table"
- "Show me the cluster details as JSON"
- "Export the connector config as YAML"

## Feedback and Support

- Report skill issues: https://github.com/confluentinc/cli/issues
- Confluent CLI documentation: https://docs.confluent.io/confluent-cli/current/
- Claude Code documentation: https://claude.ai/code/docs

---

**Ready to get started?** Try asking Claude: "What Confluent environments do I have access to?"
