# Usage Examples

This guide demonstrates common workflows using Confluent CLI skills in Claude Code. Each example shows the before/after comparison between raw CLI commands and natural language requests.

The examples follow a typical user journey from initial setup through daily operations, troubleshooting, and advanced scenarios.

## Setup and First Use

### Workflow 1: Setting Up Your First Environment and Cluster

**Goal:** New user creates an environment and provisions a basic Kafka cluster.

#### Before (Raw CLI)

```bash
# Create environment
confluent environment create dev-environment --cloud aws

# Set environment context
confluent environment use env-123456

# Create basic cluster
confluent kafka cluster create dev-cluster \
  --cloud aws \
  --region us-west-2 \
  --type basic

# Wait for cluster provisioning
confluent kafka cluster describe lkc-abc123
```

#### After (Natural Language)

**User:** "Create a development environment in AWS and provision a basic Kafka cluster in us-west-2"

**Claude:** Invokes `confluent_environment_create` with cloud="aws" and name="dev-environment". Then invokes `confluent_kafka_kafka_cluster_create` with type="basic", cloud="aws", region="us-west-2".

Returns formatted summary:
- Environment created: env-123456
- Cluster created: lkc-abc123 (status: provisioning)
- Endpoint will be available when provisioning completes

**Outcome:** Environment and cluster created. User can immediately proceed with topic creation or configuration.

---

### Workflow 2: Managing Authentication Across Cloud and On-Prem

**Goal:** User switches between Confluent Cloud and Confluent Platform deployments.

#### Before (Raw CLI)

```bash
# Log in to Confluent Cloud
confluent login
# Follow browser authentication flow

# Log in to Confluent Platform
confluent login --url https://kafka.internal.company.com
# Provide username and password

# Check current context
confluent context list
```

#### After (Natural Language)

**User:** "I need to log in to my on-prem cluster at kafka.internal.company.com"

**Claude:** Explains that authentication happens via CLI first (browser-based login not supported in skills). Provides the exact command to run:

```bash
confluent login --url https://kafka.internal.company.com
```

After authentication completes, skills will use the active session.

**Outcome:** User authenticates via CLI, then returns to Claude for subsequent operations in the correct context.

**Note:** See [LIMITATIONS.md](LIMITATIONS.md#browser-based-login-not-supported) for authentication constraints.

---

## Daily Operations

### Workflow 3: Listing and Inspecting Kafka Resources

**Goal:** Find clusters, topics, and consumer groups in an environment.

#### Before (Raw CLI)

```bash
# List all clusters
confluent kafka cluster list

# Select cluster
confluent kafka cluster use lkc-abc123

# List topics
confluent kafka topic list

# Describe specific topic
confluent kafka topic describe user-events

# List consumer groups
confluent kafka consumer list
```

#### After (Natural Language)

**User:** "Show me all topics in my production cluster with their partition counts"

**Claude:** Invokes `confluent_kafka_kafka_cluster_list` to find clusters, identifies the production cluster, then invokes `confluent_kafka_kafka_topic_list` with cluster context. Formats results as a table:

| Topic Name    | Partitions | Replication Factor |
|---------------|------------|--------------------|
| user-events   | 6          | 3                  |
| orders        | 12         | 3                  |
| inventory     | 3          | 3                  |

**Outcome:** Formatted table showing topic details without manual command construction.

---

### Workflow 4: Creating and Configuring Topics

**Goal:** Create a topic with specific partition and retention settings.

#### Before (Raw CLI)

```bash
# Create topic with configuration
confluent kafka topic create user-events \
  --cluster lkc-abc123 \
  --partitions 6 \
  --config retention.ms=604800000

# Verify creation
confluent kafka topic describe user-events --cluster lkc-abc123
```

#### After (Natural Language)

**User:** "Create a topic called user-events with 6 partitions and 7 day retention in my development cluster"

**Claude:** Invokes `confluent_kafka_kafka_topic_create` with:
- topic name: user-events
- partitions: 6
- config: retention.ms=604800000 (7 days in milliseconds)
- cluster: lkc-abc123 (identified from context)

Confirms creation:
- Topic "user-events" created
- Partitions: 6
- Retention: 7 days

**Outcome:** Topic created with correct settings. User can immediately begin producing/consuming.

---

### Workflow 5: Managing API Keys and ACLs

**Goal:** Create a service account API key with specific permissions.

#### Before (Raw CLI)

```bash
# Create service account
confluent iam service-account create kafka-producer \
  --description "Producer service"

# Create API key for service account
confluent api-key create \
  --resource lkc-abc123 \
  --service-account sa-123456

# Create ACL for write access
confluent kafka acl create \
  --allow \
  --service-account sa-123456 \
  --operation WRITE \
  --topic user-events \
  --cluster lkc-abc123
```

#### After (Natural Language)

**User:** "Create an API key for a new service account called kafka-producer with write access to the user-events topic"

**Claude:** Orchestrates multiple skill invocations:
1. `confluent_iam_iam_service-account_create` (name="kafka-producer")
2. `confluent_api-key_api-key_create` (resource=cluster, service-account=sa-123456)
3. `confluent_kafka_kafka_acl_create` (operation=WRITE, topic=user-events, principal=sa-123456)

Returns:
- Service account: sa-123456
- API Key: APIKEY123 (displays key and secret securely)
- ACL configured: WRITE access to user-events topic

**Outcome:** Complete service account setup with API key and permissions ready for use.

---

## Troubleshooting

### Workflow 6: Debugging Consumer Lag

**Goal:** Identify which consumer groups are experiencing lag.

#### Before (Raw CLI)

```bash
# List all consumer groups
confluent kafka consumer list --cluster lkc-abc123

# Describe each group to check lag
confluent kafka consumer describe group-1 --cluster lkc-abc123
confluent kafka consumer describe group-2 --cluster lkc-abc123
confluent kafka consumer describe group-3 --cluster lkc-abc123

# Manually parse output to find lagging groups
```

#### After (Natural Language)

**User:** "Which consumer groups have lag over 1000 messages in my production cluster?"

**Claude:**
1. Invokes `confluent_kafka_kafka_consumer_list` to get all groups
2. Invokes `confluent_kafka_kafka_consumer_describe` for each group
3. Filters groups where any partition has lag > 1000
4. Summarizes findings

Returns:
- analytics-group: 2,450 messages lag (partition 3)
- reporting-group: 1,200 messages lag (partition 0)

Groups with acceptable lag: data-sync-group (50 messages)

**Outcome:** Actionable list of lagging groups without manual parsing. User can investigate specific groups.

---

### Workflow 7: Investigating Cluster Health

**Goal:** Check cluster status and identify partition issues.

#### Before (Raw CLI)

```bash
# Describe cluster
confluent kafka cluster describe lkc-abc123

# List topics and check for under-replicated partitions
confluent kafka topic list --cluster lkc-abc123

# Manually correlate status information
```

#### After (Natural Language)

**User:** "Is my production cluster healthy? Check for under-replicated partitions"

**Claude:**
1. Invokes `confluent_kafka_kafka_cluster_describe` (lkc-abc123)
2. Invokes `confluent_kafka_kafka_topic_list` with detail
3. Analyzes replication status across topics

Reports:
- Cluster status: Running
- Availability: 99.9%
- Under-replicated partitions: None
- All topics fully replicated

**Outcome:** Clear health report with specific metrics. User confirms cluster is operating normally.

---

## Advanced Scenarios

### Workflow 8: Disaster Recovery - Recreating Topics from Backup

**Goal:** Restore topic configuration after accidental deletion.

#### Before (Raw CLI)

```bash
# Manually inspect source topic configuration
confluent kafka topic describe orders-topic --cluster lkc-abc123

# Note all configuration parameters
# Create new topic with matching settings
confluent kafka topic create user-events \
  --cluster lkc-abc123 \
  --partitions 12 \
  --replication-factor 3 \
  --config retention.ms=2592000000 \
  --config compression.type=lz4 \
  --config min.insync.replicas=2
```

#### After (Natural Language)

**User:** "Recreate the user-events topic with the same settings as orders-topic"

**Claude:**
1. Invokes `confluent_kafka_kafka_topic_describe` (orders-topic) to get full configuration
2. Extracts: partitions=12, retention=30 days, compression=lz4, min.insync.replicas=2
3. Invokes `confluent_kafka_kafka_topic_create` (user-events) with matching config

Confirms:
- Topic "user-events" created
- Configuration matches "orders-topic"
- Partitions: 12, Retention: 30 days, Compression: lz4

**Outcome:** Topic recreated with exact configuration from reference topic.

---

### Workflow 9: Multi-Region Cluster Management

**Goal:** Compare clusters across regions to balance workloads.

#### Before (Raw CLI)

```bash
# List all clusters
confluent kafka cluster list

# Manually group by region from output
# Record cluster counts per region
```

#### After (Natural Language)

**User:** "Show me all my Kafka clusters grouped by region"

**Claude:**
1. Invokes `confluent_kafka_kafka_cluster_list` across environments
2. Extracts region metadata for each cluster
3. Groups clusters by region
4. Formats summary table

Returns:

**us-west-2 (3 clusters)**
- prod-cluster-1 (Dedicated)
- dev-cluster-1 (Basic)
- test-cluster-1 (Basic)

**eu-west-1 (2 clusters)**
- prod-cluster-eu (Dedicated)
- dev-cluster-eu (Basic)

**ap-southeast-1 (1 cluster)**
- prod-cluster-apac (Dedicated)

**Outcome:** Formatted multi-region view showing cluster distribution. User can assess regional capacity.

---

### Workflow 10: Schema Registry Integration

**Goal:** Register an Avro schema with backward compatibility.

#### Before (Raw CLI)

```bash
# Register schema from file
confluent schema-registry schema create \
  --subject users-value \
  --schema @user.avsc \
  --compatibility BACKWARD \
  --cluster lsrc-123456

# Verify registration
confluent schema-registry schema describe users-value --version latest
```

#### After (Natural Language)

**User:** "Register this Avro schema for the users-value subject with BACKWARD compatibility"

```json
{
  "type": "record",
  "name": "User",
  "fields": [
    {"name": "id", "type": "int"},
    {"name": "username", "type": "string"},
    {"name": "email", "type": "string"}
  ]
}
```

**Claude:**
1. Saves schema to temporary file
2. Invokes `confluent_schema-registry_schema-registry_schema_create` with:
   - subject: users-value
   - schema-file: (temp file path)
   - compatibility: BACKWARD
3. Validates compatibility before registration

Confirms:
- Schema registered: users-value (version 1)
- Compatibility mode: BACKWARD
- Schema ID: 100042

**Outcome:** Schema registered and validated. Ready for producer/consumer use.

---

## Complex Workflow: End-to-End Pipeline Setup

**Goal:** Create complete data pipeline with environment, cluster, topics, connectors, and monitoring.

#### Before (Raw CLI)

Multiple commands across different namespaces:

```bash
# 1. Create environment
confluent environment create data-pipeline --cloud aws

# 2. Create Kafka cluster
confluent kafka cluster create pipeline-cluster \
  --cloud aws --region us-west-2 --type standard

# 3. Create topics
confluent kafka topic create raw-events --partitions 6
confluent kafka topic create processed-events --partitions 6

# 4. Create Schema Registry cluster
confluent schema-registry cluster enable --cloud aws --geo us

# 5. Register schemas
confluent schema-registry schema create --subject raw-events-value --schema @raw.avsc
confluent schema-registry schema create --subject processed-events-value --schema @processed.avsc

# 6. Create service account for connector
confluent iam service-account create s3-sink-connector

# 7. Create API key
confluent api-key create --resource lkc-abc123 --service-account sa-789

# 8. Set ACLs
confluent kafka acl create --allow --service-account sa-789 --operation READ --topic processed-events

# 9. Create connector
confluent connect cluster create s3-sink --config-file s3-sink.json
```

#### After (Natural Language)

**User:** "Set up a complete data pipeline in AWS us-west-2 with:
- Environment called data-pipeline
- Standard Kafka cluster
- Two topics: raw-events and processed-events (6 partitions each)
- Schema Registry enabled
- S3 sink connector with read access to processed-events topic"

**Claude:** Orchestrates the entire sequence:
1. Creates environment (data-pipeline)
2. Provisions cluster (standard, us-west-2)
3. Creates both topics in parallel (6 partitions)
4. Enables Schema Registry (geo: us)
5. Creates service account for connector
6. Generates API key for cluster access
7. Configures ACLs (READ on processed-events)
8. Creates S3 sink connector (pending config file)

Notes that connector requires configuration file for S3 credentials and asks:
"Please provide the S3 bucket name and AWS credentials configuration"

**Outcome:** Complete pipeline infrastructure created. Only connector-specific configuration remains, which requires user input for credentials.

---

## Tips for Effective Skill Usage

### Be Specific with Resource Names

Clear names help Claude identify the correct resources:

```
# Good: "Create topic user-events in my production cluster"
# Less specific: "Create a topic"
```

### Provide Context for Ambiguous Requests

When you have multiple environments or clusters, specify which one:

```
# Good: "List topics in my development environment's cluster"
# Ambiguous: "List topics"
```

### Use Natural Language for Complex Queries

Skills excel at filtering and aggregating data:

```
# Before (Raw CLI): Multiple commands + manual filtering
# After: "Show me all topics with more than 10 partitions and less than 1 hour retention"
```

### Combine Operations in Single Requests

Claude can orchestrate multi-step workflows:

```
"Create a service account called data-processor, generate an API key, and grant read access to all topics starting with 'sensor-'"
```

### Request Formatted Output

Ask for specific presentation formats:

```
"Show me my clusters as a table sorted by creation date"
"List consumer groups with lag formatted as JSON"
```

### Understand Limitations

Some operations require direct CLI usage. See [LIMITATIONS.md](LIMITATIONS.md) for complete constraints:

- Browser-based authentication
- Interactive TUI commands
- Long-running operations (>2 minutes)
- Large file uploads

For these scenarios, use the CLI directly and return to skills for management and monitoring.

---

## Next Steps

- Review [LIMITATIONS.md](LIMITATIONS.md) to understand constraints and workarounds
- Check [TROUBLESHOOTING.md](TROUBLESHOOTING.md) if skills don't respond as expected
- Explore the full skill catalog in REFERENCE.md (organized by namespace)

Skills work best when you combine natural language requests with knowledge of Confluent concepts. Experiment with different phrasings to find workflows that match your needs.
