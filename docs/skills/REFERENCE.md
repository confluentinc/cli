# Skill Reference

**⚠️ This file is auto-generated from skills.json. Do not edit manually.**

**CLI Version:** v0.0.0-test
**Generated:** 2026-03-13T17:28:04-07:00
**Total Skills:** 420

This reference documents all available Claude Code skills for the Confluent CLI. Skills are organized by namespace, corresponding to the CLI command structure.

## Table of Contents

- [Ai Skills](#ai-skills) (1 skills)
- [Api Skills](#api-skills) (7 skills)
- [Asyncapi Skills](#asyncapi-skills) (1 skills)
- [Audit Skills](#audit-skills) (4 skills)
- [Billing Skills](#billing-skills) (4 skills)
- [Byok Skills](#byok-skills) (1 skills)
- [Ccpm Skills](#ccpm-skills) (5 skills)
- [Cloud Skills](#cloud-skills) (1 skills)
- [Cluster Skills](#cluster-skills) (1 skills)
- [Completion Skills](#completion-skills) (1 skills)
- [Configuration Skills](#configuration-skills) (1 skills)
- [Connect Skills](#connect-skills) (25 skills)
- [Context Skills](#context-skills) (1 skills)
- [Custom Skills](#custom-skills) (1 skills)
- [Environment Skills](#environment-skills) (6 skills)
- [Feedback Skills](#feedback-skills) (1 skills)
- [Flink Skills](#flink-skills) (56 skills)
- [IAM Skills](#iam-skills) (56 skills)
- [Kafka Skills](#kafka-skills) (60 skills)
- [Ksql Skills](#ksql-skills) (1 skills)
- [Local Skills](#local-skills) (47 skills)
- [Login Skills](#login-skills) (1 skills)
- [Logout Skills](#logout-skills) (1 skills)
- [Network Skills](#network-skills) (74 skills)
- [Organization Skills](#organization-skills) (1 skills)
- [Plugin Skills](#plugin-skills) (1 skills)
- [Prompt Skills](#prompt-skills) (1 skills)
- [Provider Skills](#provider-skills) (6 skills)
- [Schema Skills](#schema-skills) (34 skills)
- [Secret Skills](#secret-skills) (2 skills)
- [Service Skills](#service-skills) (1 skills)
- [Shell Skills](#shell-skills) (1 skills)
- [Stream Skills](#stream-skills) (5 skills)
- [Tableflow Skills](#tableflow-skills) (6 skills)
- [Unified Skills](#unified-skills) (3 skills)
- [Update Skills](#update-skills) (1 skills)
- [Version Skills](#version-skills) (1 skills)

## Ai Skills

### `confluent_ai`

**Description:** Start an interactive AI shell.

**Priority:** low

---

## Api Skills

### `confluent_api_key_create`

**Description:** Create API keys for a given resource.

Example:
Create a Cloud API key:

  $ confluent api-key create --resource cloud

Create a Flink API key for region "N. Virginia (us-east-1)":

  $ confluent api-key create --resource flink --cloud aws --region us-east-1

Create an API key with full access to Kafka cluster "lkc-123456":

  $ confluent api-key create --resource lkc-123456

Create an API key for Kafka cluster "lkc-123456" and service account "sa-123456":

  $ confluent api-key create --resource lkc-123456 --service-account sa-123456

Create an API key for Schema Registry cluster "lsrc-123456":

  $ confluent api-key create --resource lsrc-123456

Create an API key for KSQL cluster "lksqlc-123456":

  $ confluent api-key create --resource lksqlc-123456

Create a Tableflow API key:

  $ confluent api-key create --resource tableflow

**Required Parameters:**

- `resource`: The ID of the resource the API key is for. Use "cloud" for a Cloud API key, "flink" for a Flink API key, or "tableflow" for a Tableflow API key.

**Optional Parameters:**

- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `context`: CLI context name.
- `description`: Description of API key.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `region`: Cloud region for Flink (use "confluent flink region list" to see all).
- `service-account`: Service account ID.
- `use`: Use the created API key for the provided resource.

**Priority:** medium

---

### `confluent_api_key_delete`

**Description:** Delete one or more API keys.

**Optional Parameters:**

- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_api_key_describe`

**Description:** Describe an API key.

**Optional Parameters:**

- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_api_key_list`

**Description:** List the API keys.

Example:
List the API keys that belong to service account "sa-123456" on cluster "lkc-123456".

  $ confluent api-key list --resource lkc-123456 --service-account sa-123456

**Optional Parameters:**

- `current-user`: Show only API keys belonging to current user.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `resource`: The ID of the resource the API key is for. Use "cloud" for a Cloud API key, "flink" for a Flink API key, or "tableflow" for a Tableflow API key.
- `service-account`: Service account ID.

**Priority:** high

---

### `confluent_api_key_other`

**Description:** Store an API key/secret locally to use in the CLI.

Example:
Pass the API key and secret as arguments

  $ confluent api-key store my-key my-secret

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Force overwrite existing secret for this key.
- `resource`: The ID of the resource the API key is for.

**Priority:** low

---

### `confluent_api_key_update`

**Description:** Update an API key.

**Optional Parameters:**

- `description`: Description of the API key.

**Priority:** medium

---

### `confluent_api_key_use`

**Description:** Use an API key in subsequent commands.

**Optional Parameters:**

- `resource`: The ID of the resource the API key is for. Use "cloud" for a Cloud API key, "flink" for a Flink API key, or "tableflow" for a Tableflow API key.

**Priority:** low

---

## Asyncapi Skills

### `confluent_asyncapi`

**Description:** Export an AsyncAPI specification.

Example:
Export an AsyncAPI specification with topic "my-topic" and all topics starting with "prefix-".

  $ confluent asyncapi export --topics "my-topic,prefix-*"

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `consume-examples`: Consume messages from topics for populating examples.
- `environment`: Environment ID.
- `file`: Output file name. (default: `asyncapi-spec.yaml`)
- `group`: Consumer Group ID for getting messages. (default: `consumerApplication`)
- `kafka-api-key`: Kafka cluster API key.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `schema-context`: Use a specific schema context. (default: `default`)
- `schema-registry-api-key`: API key for Schema Registry.
- `schema-registry-api-secret`: API secret for Schema Registry.
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.
- `spec-version`: Version number of the output file. (default: `1.0.0`)
- `topics`: A comma-separated list of topics to export. Supports prefixes ending with a wildcard (*).
- `value-format`: Format message value as "string", "avro", "double", "integer", "jsonschema", or "protobuf". Note that schema references are not supported for Avro. (default: `string`)

**Priority:** low

---

## Audit Skills

### `confluent_audit_log_describe`

**Description:** Describe the audit log configuration for this organization.

**Optional Parameters:**

- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_audit_log_list`

**Description:** List routes matching a resource & sub-resources.

**Required Parameters:**

- `resource`: The Confluent resource name (CRN) that is the subject of the query.

**Optional Parameters:**

- `client-cert-path`: Path to client cert to be verified by MDS. Include for mTLS authentication.
- `client-key-path`: Path to client private key, include for mTLS authentication.
- `context`: CLI context name.

**Priority:** high

---

### `confluent_audit_log_other`

**Description:** Edit the audit-log configuration specification interactively.

**Optional Parameters:**

- `client-cert-path`: Path to client cert to be verified by MDS. Include for mTLS authentication.
- `client-key-path`: Path to client private key, include for mTLS authentication.
- `context`: CLI context name.

**Priority:** low

---

### `confluent_audit_log_update`

**Description:** Submits audit-log configuration specification object to the API.

**Optional Parameters:**

- `client-cert-path`: Path to client cert to be verified by MDS. Include for mTLS authentication.
- `client-key-path`: Path to client private key, include for mTLS authentication.
- `context`: CLI context name.
- `file`: A local file path to the JSON configuration file, read as input. Otherwise the command will read from standard input.
- `force`: Updates the configuration, overwriting any concurrent modifications.

**Priority:** medium

---

## Billing Skills

### `confluent_billing_describe`

**Description:** Describe the active payment method.

**Priority:** high

---

### `confluent_billing_list`

**Description:** List Confluent Cloud billing costs.

Example:
List billing costs from 2023-01-01 to 2023-01-10:

  $ confluent billing cost list --start-date 2023-01-01 --end-date 2023-01-10

**Required Parameters:**

- `start-date`: Start date.
- `end-date`: End date.

**Optional Parameters:**

- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_billing_other`

**Description:** Add a new promo code.

**Priority:** low

---

### `confluent_billing_update`

**Description:** Update the active payment method.

**Priority:** medium

---

## Byok Skills

### `confluent_byok`

**Description:** Register a self-managed encryption key.

Example:
Register a new self-managed encryption key for AWS:

  $ confluent byok create "arn:aws:kms:us-west-2:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab"

Register a new self-managed encryption key for Azure:

  $ confluent byok create "https://vault-name.vault.azure.net/keys/key-name" --tenant "00000000-0000-0000-0000-000000000000" --key-vault "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/resourcegroup-name/providers/Microsoft.KeyVault/vaults/vault-name"

Register a new self-managed encryption key for GCP:

  $ confluent byok create "projects/exampleproject/locations/us-central1/keyRings/testkeyring/cryptoKeys/testbyokkey/cryptoKeyVersions/3"

**Optional Parameters:**

- `display-name`: A human-readable name for the self-managed key.
- `key-vault`: The ID of the Azure Key Vault where the key is stored.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `tenant`: The ID of the Azure Active Directory tenant that the key vault belongs to.

**Priority:** medium

---

## Ccpm Skills

### `confluent_ccpm_create`

**Description:** Create a custom Connect plugin.

Example:
Create a custom Connect plugin for AWS.

  $ confluent ccpm plugin create --name "My Custom Plugin" --cloud AWS --description "A custom connector for data processing" --environment env-12345

Create a custom Connect plugin for GCP with minimal description.

  $ confluent ccpm plugin create --name "GCP Data Connector" --cloud GCP --environment env-abcdef

**Required Parameters:**

- `name`: Display name of the custom Connect plugin.
- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".

**Optional Parameters:**

- `description`: Description of the custom Connect plugin.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_ccpm_delete`

**Description:** Delete one or more custom Connect plugins.

Example:
Delete a custom Connect plugin by ID.

  $ confluent ccpm plugin delete plugin-123456 --environment env-12345

Force delete a custom Connect plugin without confirmation.

  $ confluent ccpm plugin delete plugin-123456 --environment env-12345 --force

**Optional Parameters:**

- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_ccpm_describe`

**Description:** Describe a custom Connect plugin.

Example:
Describe a custom Connect plugin by ID.

  $ confluent ccpm plugin describe plugin-123456 --environment env-12345

**Optional Parameters:**

- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_ccpm_list`

**Description:** List custom Connect plugins.

Example:
List all custom Connect plugins in an environment.

  $ confluent ccpm plugin list --environment env-12345

List custom Connect plugins filtered by cloud provider.

  $ confluent ccpm plugin list --environment env-12345 --cloud AWS

**Optional Parameters:**

- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_ccpm_update`

**Description:** Update a custom Connect plugin.

Example:
Update the name and description of a custom Connect plugin.

  $ confluent ccpm plugin update plugin-123456 --name "Updated Plugin Name" --description "Updated description" --environment env-12345

Update only the name of a custom Connect plugin.

  $ confluent ccpm plugin update plugin-123456 --name "New Plugin Name" --environment env-12345

**Optional Parameters:**

- `description`: Description of the custom Connect plugin.
- `environment`: Environment ID.
- `name`: Display name of the custom Connect plugin.

**Priority:** medium

---

## Cloud Skills

### `confluent_cloud_signup`

**Description:** Sign up for Confluent Cloud.

**Priority:** low

---

## Cluster Skills

### `confluent_cluster`

**Description:** Describe a Kafka cluster.

Example:
Discover the cluster ID and Kafka ID for Connect.

  $ confluent cluster describe --url http://localhost:8083

**Optional Parameters:**

- `certificate-authority-path`: Self-signed certificate chain in PEM format.
- `client-cert-path`: Path to client cert to be verified by MDS. Include for mTLS authentication.
- `client-key-path`: Path to client private key, include for mTLS authentication.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `url`: URL to a Confluent cluster.

**Priority:** high

---

## Completion Skills

### `confluent_completion`

**Description:** Print shell completion code.

**Priority:** low

---

## Configuration Skills

### `confluent_configuration`

**Description:** Describe a user-configurable field.

Example:
View the "disable_update_check" configuration.

  $ confluent configuration describe disable_update_check

**Optional Parameters:**

- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

## Connect Skills

### `confluent_connect_connect_artifact_create`

**Description:** Create a Connect artifact.

Example:
Create Connect artifact "my-connect-artifact".

  $ confluent connect artifact create my-connect-artifact --artifact-file artifact.jar --cloud aws --environment env-abc123 --description "This is my new Connect artifact"

**Required Parameters:**

- `artifact-file`: Connect artifact JAR file or ZIP file.
- `cloud`: Specify the cloud provider as "aws" or "azure".

**Optional Parameters:**

- `context`: CLI context name.
- `description`: Specify the Connect artifact description.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_connect_connect_artifact_delete`

**Description:** Delete one or more Connect artifacts.

Example:
Delete Connect artifact.

  $ confluent connect artifact delete cfa-abc123 --cloud aws --environment env-abc123

**Required Parameters:**

- `cloud`: Specify the cloud provider as "aws" or "azure".

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_connect_connect_artifact_describe`

**Description:** Describe a Connect artifact.

Example:
Describe a Connect artifact.

  $ confluent connect artifact describe cfa-abc123 --cloud aws --environment env-abc123

**Required Parameters:**

- `cloud`: Specify the cloud provider as "aws" or "azure".

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_connect_connect_artifact_list`

**Description:** List Connect artifacts.

Example:
List Connect artifacts.

  $ confluent connect artifact list --cloud aws --environment env-abc123

**Required Parameters:**

- `cloud`: Specify the cloud provider as "aws" or "azure".

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_connect_connect_cluster_create`

**Description:** Create a connector.

Example:
Create a configuration file with connector configs and offsets.

  {
    "name": "MyGcsLogsBucketConnector",
    "config": {
      "connector.class": "GcsSink",
      "data.format": "BYTES",
      "flush.size": "1000",
      "gcs.bucket.name": "APILogsBucket",
      "gcs.credentials.config": "****************",
      "kafka.api.key": "****************",
      "kafka.api.secret": "****************",
      "name": "MyGcsLogsBucketConnector",
      "tasks.max": "2",
      "time.interval": "DAILY",
      "topics": "APILogsTopic"
    },
    "offsets": [
  	{
  	  "partition": {
  		"kafka_partition": 0,
  		"kafka_topic": "topic_A"
  	  },
  	  "offset": {
  		"kafka_offset": 1000
  	  }
  	}
    ]
  }

Create a connector in the current or specified Kafka cluster context.

  $ confluent connect cluster create --config-file config.json

  $ confluent connect cluster create --config-file config.json --cluster lkc-123456

**Required Parameters:**

- `config-file`: JSON connector configuration file.

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_connect_connect_cluster_delete`

**Description:** Delete one or more connectors.

Example:
Delete a connector in the current or specified Kafka cluster context.

  $ confluent connect cluster delete

  $ confluent connect cluster delete --cluster lkc-123456

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_connect_connect_cluster_describe`

**Description:** Describe a connector.

Example:
Describe connector and task level details of a connector in the current or specified Kafka cluster context.

  $ confluent connect cluster describe lcc-123456

  $ confluent connect cluster describe lcc-123456 --cluster lkc-123456

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_connect_connect_cluster_list`

**Description:** List connectors.

Example:
List connectors in the current or specified Kafka cluster context.

  $ confluent connect cluster list

  $ confluent connect cluster list --cluster lkc-123456

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_connect_connect_cluster_other`

**Description:** Pause connectors.

Example:
Pause connectors "lcc-000001" and "lcc-000002":

  $ confluent connect cluster pause lcc-000001 lcc-000002

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.

**Priority:** low

---

### `confluent_connect_connect_cluster_update`

**Description:** Update a connector configuration.

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `config`: A comma-separated list of configuration overrides ("key=value") for the connector being updated.
- `config-file`: JSON connector configuration file.
- `context`: CLI context name.
- `environment`: Environment ID.

**Priority:** medium

---

### `confluent_connect_connect_custom_connector_runtime_list`

**Description:** List custom connector runtimes.

**Optional Parameters:**

- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_connect_connect_custom_plugin_create`

**Description:** Create a custom connector plugin.

Example:
Create custom connector plugin "my-plugin".

  $ confluent connect custom-plugin create my-plugin --plugin-file datagen.zip --connector-type source --connector-class io.confluent.kafka.connect.datagen.DatagenConnector --cloud aws

**Required Parameters:**

- `plugin-file`: ZIP/JAR custom plugin file.
- `connector-class`: Connector class of custom plugin.
- `connector-type`: Connector type of custom plugin.

**Optional Parameters:**

- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp". (default: `aws`)
- `context`: CLI context name.
- `description`: Description of custom plugin.
- `documentation-link`: Document link of custom plugin.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `sensitive-properties`: A comma-separated list of sensitive property names.

**Priority:** medium

---

### `confluent_connect_connect_custom_plugin_delete`

**Description:** Delete one or more custom connector plugins.

**Optional Parameters:**

- `context`: CLI context name.
- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_connect_connect_custom_plugin_describe`

**Description:** Describe a custom connector plugin.

Example:
Describe custom connector plugin

  $ confluent connect custom-plugin describe ccp-123456

**Optional Parameters:**

- `context`: CLI context name.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_connect_connect_custom_plugin_list`

**Description:** List custom connector plugins.

Example:
List custom connector plugins in the org

  $ confluent connect custom-plugin list --cloud aws

**Optional Parameters:**

- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `context`: CLI context name.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_connect_connect_custom_plugin_update`

**Description:** Update a custom connector plugin configuration.

**Optional Parameters:**

- `context`: CLI context name.
- `description`: Description of custom plugin.
- `documentation-link`: Document link of custom plugin.
- `name`: Name of custom plugin.
- `sensitive-properties`: A comma-separated list of sensitive property names.

**Priority:** medium

---

### `confluent_connect_connect_event_describe`

**Description:** Describe the Connect log events configuration.

**Optional Parameters:**

- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_connect_connect_offset_delete`

**Description:** Delete a connector's offsets.

Example:
Delete offsets for a connector in the current or specified Kafka cluster context.

  $ confluent connect offset delete lcc-123456

  $ confluent connect offset update lcc-123456 --cluster lkc-123456

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_connect_connect_offset_describe`

**Description:** Describe connector offsets.

Example:
Describe offsets for a connector in the current or specified Kafka cluster context.

  $ confluent connect offset describe lcc-123456

  $ confluent connect offset describe lcc-123456 --cluster lkc-123456

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `staleness-threshold`: Repeatedly fetch offsets, until receiving an offset with an observed time within the staleness threshold in seconds, for a minimum of 5 seconds. (default: `120`)
- `timeout`: Max time in seconds to wait until we get an offset within the staleness threshold. (default: `30`)

**Priority:** high

---

### `confluent_connect_connect_offset_status_describe`

**Description:** Describe connector offset update or delete status.

Example:
Describe the status of the latest offset update/delete operation for a connector in the current or specified Kafka cluster context.

  $ confluent connect offset status describe lcc-123456

  $ confluent connect offset status describe lcc-123456 --cluster lkc-123456

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_connect_connect_offset_update`

**Description:** Update a connector's offsets.

Example:
The configuration file contains offsets to be set for the connector.

  {
    "offsets": [
  	{
  	  "partition": {
  		"kafka_partition": 0,
  		"kafka_topic": "topic_A"
  	  },
  	  "offset": {
  		"kafka_offset": 1000
  	  }
  	}
    ]
  }

Update offsets for a connector in the current or specified Kafka cluster context.

  $ confluent connect offset update lcc-123456 --config-file config.json

  $ confluent connect offset update lcc-123456 --config-file config.json --cluster lkc-123456

**Required Parameters:**

- `config-file`: JSON file containing new connector offsets.

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_connect_connect_other`

**Description:** Manage logs for connectors.

Example:
Query connector logs with log level ERROR between the provided time window:

  $ confluent connect logs lcc-123456 --level ERROR --start-time "2025-02-01T00:00:00Z" --end-time "2025-02-01T23:59:59Z"

Query connector logs with log level ERROR and WARN between the provided time window:

  $ confluent connect logs lcc-123456 --level "ERROR|WARN" --start-time "2025-02-01T00:00:00Z" --end-time "2025-02-01T23:59:59Z"

Query subsequent pages of connector logs for the same query by executing the command with next flag until "No logs found for the current query" is printed to the console:

  $ confluent connect logs lcc-123456 --level ERROR --start-time "2025-02-01T00:00:00Z" --end-time "2025-02-01T23:59:59Z" --next

Query connector logs with log level ERROR and containing "example error" in logs between the provided time window, and store in file:

  $ confluent connect logs lcc-123456 --level "ERROR" --search-text "example error" --start-time "2025-02-01T00:00:00Z" --end-time "2025-02-01T23:59:59Z" --output-file errors.json

Query connector logs with log level ERROR and matching regex "exa*" in logs between the provided time window, and store in file:

  $ confluent connect logs lcc-123456 --level "ERROR" --search-text "exa*" --start-time "2025-02-01T00:00:00Z" --end-time "2025-02-01T23:59:59Z" --output-file errors.json

**Required Parameters:**

- `start-time`: Start time for log query in ISO 8601 (https://en.wikipedia.org/wiki/ISO_8601) UTC datetime format (e.g., 2025-02-01T00:00:00Z).
- `end-time`: End time for log query in ISO 8601 (https://en.wikipedia.org/wiki/ISO_8601) UTC datetime format (e.g., 2025-02-01T23:59:59Z).

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `environment`: Environment ID.
- `level`: Log level filter (INFO, WARN, ERROR). Defaults to ERROR. Use '|' to specify multiple levels (e.g., ERROR|WARN). (default: `ERROR`)
- `next`: Whether to fetch next page of logs after the next execution of the command.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `output-file`: Output file path to append connector logs.
- `search-text`: Search text within logs.

**Priority:** low

---

### `confluent_connect_connect_plugin_describe`

**Description:** Describe a connector plugin.

Example:
Describe the required connector configuration parameters for connector plugin "MySource".

  $ confluent connect plugin describe MySource

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_connect_connect_plugin_list`

**Description:** List connector plugin types.

Example:
List connectors in the current or specified Kafka cluster context.

  $ confluent connect plugin list

  $ confluent connect plugin list --cluster lkc-123456

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_connect_connect_plugin_other`

**Description:** Install a Connect plugin.

Example:
Install the latest version of the Datagen connector into your local Confluent Platform environment.

  $ confluent connect plugin install confluentinc/kafka-connect-datagen:latest

Install the latest version of the Datagen connector in a user-specified directory and update a worker configuration file.

  $ confluent connect plugin install confluentinc/kafka-connect-datagen:latest --plugin-directory $CONFLUENT_HOME/plugins --worker-configurations $CONFLUENT_HOME/etc/kafka/connect-distributed.properties

**Optional Parameters:**

- `confluent-platform`: The path to a Confluent Platform archive installation. By default, this command will search for Confluent Platform installations in common locations.
- `dry-run`: Run the command without committing changes.
- `force`: Proceed without user input.
- `plugin-directory`: The plugin installation directory. If not specified, a default will be selected based on your Confluent Platform installation.
- `worker-configurations`: A comma-separated list of paths to one or more Kafka Connect worker configuration files. Each worker file will be updated to load plugins from the plugin directory in addition to any prior directories.

**Priority:** low

---

## Context Skills

### `confluent_context`

**Description:** Delete one or more contexts.

**Optional Parameters:**

- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

## Custom Skills

### `confluent_custom_code_logging`

**Description:** Create a custom code logging.

Example:
Create custom code logging.

  $ confluent custom-code-logging create --cloud aws --region us-west-2 --topic topic-123 --cluster cluster-123 --environment env-000000

**Required Parameters:**

- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `region`: Cloud region for Kafka (use "confluent kafka region list" to see all).
- `cluster`: Kafka cluster ID.
- `topic`: Kafka topic of custom code logging destination.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `log-level`: Specify the Custom Code Logging Log Level as "INFO", "DEBUG", "ERROR", or "WARN". (default: `INFO`)
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

## Environment Skills

### `confluent_environment_create`

**Description:** Create a new Confluent Cloud environment.

**Optional Parameters:**

- `context`: CLI context name.
- `governance-package`: Specify the Stream Governance package as "essentials" or "advanced". Downgrading the package from "advanced" to "essentials" is not allowed once the Schema Registry cluster is provisioned. (default: `essentials`)
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_environment_delete`

**Description:** Delete one or more Confluent Cloud environments.

**Optional Parameters:**

- `context`: CLI context name.
- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_environment_describe`

**Description:** Describe a Confluent Cloud environment.

**Optional Parameters:**

- `context`: CLI context name.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_environment_list`

**Description:** List Confluent Cloud environments.

**Optional Parameters:**

- `context`: CLI context name.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_environment_update`

**Description:** Update an existing Confluent Cloud environment.

**Optional Parameters:**

- `context`: CLI context name.
- `governance-package`: Specify the Stream Governance package as "essentials" or "advanced". Downgrading the package from "advanced" to "essentials" is not allowed once the Schema Registry cluster is provisioned.
- `name`: New name for Confluent Cloud environment.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_environment_use`

**Description:** Use an environment in subsequent commands.

**Priority:** low

---

## Feedback Skills

### `confluent_feedback`

**Description:** Submit feedback for the Confluent CLI.

**Priority:** low

---

## Flink Skills

### `confluent_flink_flink_application_create`

**Description:** Create a Flink application.

**Required Parameters:**

- `environment`: Name of the Flink environment.

**Optional Parameters:**

- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CMF_CERTIFICATE_AUTHORITY_PATH" may be set in place of this flag.
- `client-cert-path`: Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.
- `client-key-path`: Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.
- `output`: Specify the output format as "json" or "yaml". (default: `json`)
- `url`: Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.

**Priority:** medium

---

### `confluent_flink_flink_application_delete`

**Description:** Delete one or more Flink applications.

**Required Parameters:**

- `environment`: Name of the Flink environment.

**Optional Parameters:**

- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CMF_CERTIFICATE_AUTHORITY_PATH" may be set in place of this flag.
- `client-cert-path`: Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.
- `client-key-path`: Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.
- `force`: Skip the deletion confirmation prompt.
- `url`: Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.

**Priority:** medium

---

### `confluent_flink_flink_application_describe`

**Description:** Describe a Flink application.

**Required Parameters:**

- `environment`: Name of the Flink environment.

**Optional Parameters:**

- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CMF_CERTIFICATE_AUTHORITY_PATH" may be set in place of this flag.
- `client-cert-path`: Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.
- `client-key-path`: Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.
- `output`: Specify the output format as "json" or "yaml". (default: `json`)
- `url`: Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.

**Priority:** high

---

### `confluent_flink_flink_application_list`

**Description:** List Flink applications.

**Required Parameters:**

- `environment`: Name of the Flink environment.

**Optional Parameters:**

- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CMF_CERTIFICATE_AUTHORITY_PATH" may be set in place of this flag.
- `client-cert-path`: Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.
- `client-key-path`: Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `url`: Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.

**Priority:** high

---

### `confluent_flink_flink_application_other`

**Description:** Forward the web UI of a Flink application.

**Required Parameters:**

- `environment`: Name of the Flink environment.

**Optional Parameters:**

- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CMF_CERTIFICATE_AUTHORITY_PATH" may be set in place of this flag.
- `client-cert-path`: Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.
- `client-key-path`: Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.
- `port`: Port to forward the web UI to. If not provided, a random, OS-assigned port will be used.
- `url`: Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.

**Priority:** low

---

### `confluent_flink_flink_application_update`

**Description:** Update a Flink application.

**Required Parameters:**

- `environment`: Name of the Flink environment.

**Optional Parameters:**

- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CMF_CERTIFICATE_AUTHORITY_PATH" may be set in place of this flag.
- `client-cert-path`: Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.
- `client-key-path`: Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.
- `output`: Specify the output format as "json" or "yaml". (default: `json`)
- `url`: Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.

**Priority:** medium

---

### `confluent_flink_flink_artifact_create`

**Description:** Create a Flink UDF artifact.

Example:
Create Flink artifact "my-flink-artifact".

  $ confluent flink artifact create my-flink-artifact --artifact-file artifact.jar --cloud aws --region us-west-2 --environment env-123456

Create Flink artifact "flink-java-artifact".

  $ confluent flink artifact create my-flink-artifact --artifact-file artifact.jar --cloud aws --region us-west-2 --environment env-123456 --description flinkJavaScalar

**Required Parameters:**

- `artifact-file`: Flink artifact JAR file or ZIP file.
- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `region`: Cloud region for Flink (use "confluent flink region list" to see all).

**Optional Parameters:**

- `context`: CLI context name.
- `description`: Specify the Flink artifact description.
- `documentation-link`: Specify the Flink artifact documentation link.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `runtime-language`: Specify the Flink artifact runtime language as "python" or "java". (default: `java`)

**Priority:** medium

---

### `confluent_flink_flink_artifact_delete`

**Description:** Delete one or more Flink UDF artifacts.

Example:
Delete Flink UDF artifact.

  $ confluent flink artifact delete --cloud aws --region us-west-2 --environment env-123456 cfa-123456

**Required Parameters:**

- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `region`: Cloud region for Flink (use "confluent flink region list" to see all).

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_flink_flink_artifact_describe`

**Description:** Describe a Flink UDF artifact.

Example:
Describe Flink UDF artifact.

  $ confluent flink artifact describe --cloud aws --region us-west-2 --environment env-123456 cfa-123456

**Required Parameters:**

- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `region`: Cloud region for Flink (use "confluent flink region list" to see all).

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_flink_flink_artifact_list`

**Description:** List Flink UDF artifacts.

Example:
List Flink UDF artifacts.

  $ confluent flink artifact list --cloud aws --region us-west-2 --environment env-123456

**Required Parameters:**

- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `region`: Cloud region for Flink (use "confluent flink region list" to see all).

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_flink_flink_catalog_create`

**Description:** Create a Flink catalog.

**Optional Parameters:**

- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CMF_CERTIFICATE_AUTHORITY_PATH" may be set in place of this flag.
- `client-cert-path`: Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.
- `client-key-path`: Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `url`: Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.

**Priority:** medium

---

### `confluent_flink_flink_catalog_delete`

**Description:** Delete one or more Flink catalogs in Confluent Platform.

**Optional Parameters:**

- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CMF_CERTIFICATE_AUTHORITY_PATH" may be set in place of this flag.
- `client-cert-path`: Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.
- `client-key-path`: Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.
- `force`: Skip the deletion confirmation prompt.
- `url`: Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.

**Priority:** medium

---

### `confluent_flink_flink_catalog_describe`

**Description:** Describe a Flink catalog in Confluent Platform.

**Optional Parameters:**

- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CMF_CERTIFICATE_AUTHORITY_PATH" may be set in place of this flag.
- `client-cert-path`: Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.
- `client-key-path`: Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `url`: Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.

**Priority:** high

---

### `confluent_flink_flink_catalog_list`

**Description:** List Flink catalogs in Confluent Platform.

**Optional Parameters:**

- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CMF_CERTIFICATE_AUTHORITY_PATH" may be set in place of this flag.
- `client-cert-path`: Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.
- `client-key-path`: Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `url`: Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.

**Priority:** high

---

### `confluent_flink_flink_compute_pool_create`

**Description:** Create a Flink compute pool.

Example:
Create Flink compute pool "my-compute-pool" in AWS with 5 CFUs.

  $ confluent flink compute-pool create my-compute-pool --cloud aws --region us-west-2 --max-cfu 5

**Required Parameters:**

- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `region`: Cloud region for Flink (use "confluent flink region list" to see all).

**Optional Parameters:**

- `environment`: Environment ID.
- `max-cfu`: Maximum number of Confluent Flink Units (CFU). (default: `5`)
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_flink_flink_compute_pool_delete`

**Description:** Delete one or more Flink compute pools.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_flink_flink_compute_pool_describe`

**Description:** Describe a Flink compute pool.

**Optional Parameters:**

- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_flink_flink_compute_pool_list`

**Description:** List Flink compute pools.

**Optional Parameters:**

- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `region`: Cloud region for Flink (use "confluent flink region list" to see all).

**Priority:** high

---

### `confluent_flink_flink_compute_pool_other`

**Description:** Unset the current Flink compute pool.

Example:
Unset default compute pool:

  $ confluent flink compute-pool unset

**Optional Parameters:**

- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** low

---

### `confluent_flink_flink_compute_pool_update`

**Description:** Update a Flink compute pool.

Example:
Update name and CFU count of a Flink compute pool.

  $ confluent flink compute-pool update lfcp-123456 --name "new name" --max-cfu 5

**Optional Parameters:**

- `environment`: Environment ID.
- `max-cfu`: Maximum number of Confluent Flink Units (CFU).
- `name`: Name of the compute pool.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_flink_flink_compute_pool_use`

**Description:** Use a Flink compute pool in subsequent commands.

**Priority:** low

---

### `confluent_flink_flink_connection_create`

**Description:** Create a Flink connection.

Example:
Create Flink connection "my-connection" in AWS us-west-2 for OpenAPI with endpoint and API key.

  $ confluent flink connection create my-connection --cloud aws --region us-west-2 --type openai --endpoint https://api.openai.com/v1/chat/completions --api-key 0000000000000000

**Required Parameters:**

- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `region`: Cloud region for Flink (use "confluent flink region list" to see all).
- `type`: Specify the connection type as "openai", "azureml", "azureopenai", "anthropic", "fireworksai", "a2a", "bedrock", "sagemaker", "googleai", "vertexai", "mongodb", "elastic", "pinecone", "couchbase", "confluent_jdbc", "rest", "mcp_server", "cosmosdb", or "s3vectors".
- `endpoint`: Specify endpoint for the connection.

**Optional Parameters:**

- `api-key`: Specify API key for the type: "openai", "azureml", "azureopenai", "anthropic", "fireworksai", "googleai", "elastic", "pinecone", "a2a", "mcp_server", or "cosmosdb".
- `aws-access-key`: Specify access key for the type: "bedrock", "sagemaker", or "s3vectors".
- `aws-secret-key`: Specify secret key for the type: "bedrock", "sagemaker", or "s3vectors".
- `aws-session-token`: Specify session token for the type: "bedrock", "sagemaker", or "s3vectors".
- `client-id`: Specify OAuth2 client ID for the type: "a2a", "rest", or "mcp_server".
- `client-secret`: Specify OAuth2 client secret for the type: "a2a", "rest", or "mcp_server".
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `password`: Specify password for the type: "mongodb", "couchbase", "confluent_jdbc", "a2a", "rest", or "mcp_server".
- `scope`: Specify OAuth2 scope for the type: "a2a", "rest", or "mcp_server".
- `service-key`: Specify service key for the type: "vertexai".
- `sse-endpoint`: Specify SSE endpoint for the type: "mcp_server".
- `token`: Specify bearer token for the type: "a2a", "rest", or "mcp_server".
- `token-endpoint`: Specify OAuth2 token endpoint for the type: "a2a", "rest", or "mcp_server".
- `transport-type`: Specify transport type for the type: "mcp_server". Default: SSE.
- `username`: Specify username for the type: "mongodb", "couchbase", "confluent_jdbc", "a2a", "rest", or "mcp_server".

**Priority:** medium

---

### `confluent_flink_flink_connection_delete`

**Description:** Delete one or more Flink connections.

**Required Parameters:**

- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `region`: Cloud region for Flink (use "confluent flink region list" to see all).

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_flink_flink_connection_describe`

**Description:** Describe a Flink connection.

**Required Parameters:**

- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `region`: Cloud region for Flink (use "confluent flink region list" to see all).

**Optional Parameters:**

- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_flink_flink_connection_list`

**Description:** List Flink connections.

**Required Parameters:**

- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `region`: Cloud region for Flink (use "confluent flink region list" to see all).

**Optional Parameters:**

- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `type`: Specify the connection type as "openai", "azureml", "azureopenai", "anthropic", "fireworksai", "a2a", "bedrock", "sagemaker", "googleai", "vertexai", "mongodb", "elastic", "pinecone", "couchbase", "confluent_jdbc", "rest", "mcp_server", "cosmosdb", or "s3vectors".

**Priority:** high

---

### `confluent_flink_flink_connection_update`

**Description:** Update a Flink connection. Only secret can be updated.

Example:
Update API key of Flink connection "my-connection".

  $ confluent flink connection update my-connection --cloud aws --region us-west-2 --api-key new-key

**Required Parameters:**

- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `region`: Cloud region for Flink (use "confluent flink region list" to see all).

**Optional Parameters:**

- `api-key`: Specify API key for the type: "openai", "azureml", "azureopenai", "anthropic", "fireworksai", "googleai", "elastic", "pinecone", "a2a", "mcp_server", or "cosmosdb".
- `aws-access-key`: Specify access key for the type: "bedrock", "sagemaker", or "s3vectors".
- `aws-secret-key`: Specify secret key for the type: "bedrock", "sagemaker", or "s3vectors".
- `aws-session-token`: Specify session token for the type: "bedrock", "sagemaker", or "s3vectors".
- `client-id`: Specify OAuth2 client ID for the type: "a2a", "rest", or "mcp_server".
- `client-secret`: Specify OAuth2 client secret for the type: "a2a", "rest", or "mcp_server".
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `password`: Specify password for the type: "mongodb", "couchbase", "confluent_jdbc", "a2a", "rest", or "mcp_server".
- `scope`: Specify OAuth2 scope for the type: "a2a", "rest", or "mcp_server".
- `service-key`: Specify service key for the type: "vertexai".
- `sse-endpoint`: Specify SSE endpoint for the type: "mcp_server".
- `token`: Specify bearer token for the type: "a2a", "rest", or "mcp_server".
- `token-endpoint`: Specify OAuth2 token endpoint for the type: "a2a", "rest", or "mcp_server".
- `transport-type`: Specify transport type for the type: "mcp_server". Default: SSE.
- `username`: Specify username for the type: "mongodb", "couchbase", "confluent_jdbc", "a2a", "rest", or "mcp_server".

**Priority:** medium

---

### `confluent_flink_flink_connectivity_type_use`

**Description:** Select a Flink connectivity type.

**Priority:** low

---

### `confluent_flink_flink_detached_savepoint_create`

**Description:** Create a Flink detached savepoint in Confluent Platform.

Example:
Create a Flink savepoint named "my-savepoint".

  $ confluent flink detached-savepoint create ds1 --path path1

**Required Parameters:**

- `path`: The path to the savepoint data.

**Optional Parameters:**

- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CMF_CERTIFICATE_AUTHORITY_PATH" may be set in place of this flag.
- `client-cert-path`: Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.
- `client-key-path`: Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `url`: Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.

**Priority:** medium

---

### `confluent_flink_flink_detached_savepoint_delete`

**Description:** Delete Flink detached savepoints in Confluent Platform.

**Optional Parameters:**

- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CMF_CERTIFICATE_AUTHORITY_PATH" may be set in place of this flag.
- `client-cert-path`: Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.
- `client-key-path`: Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.
- `force`: Skip the deletion confirmation prompt.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `url`: Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.

**Priority:** medium

---

### `confluent_flink_flink_detached_savepoint_describe`

**Description:** Describe a Flink detached savepoint in Confluent Platform.

Example:
Describe a Flink savepoint named "my-savepoint".

  $ confluent flink detached-savepoint describe my-savepoint

**Optional Parameters:**

- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CMF_CERTIFICATE_AUTHORITY_PATH" may be set in place of this flag.
- `client-cert-path`: Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.
- `client-key-path`: Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `url`: Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.

**Priority:** high

---

### `confluent_flink_flink_detached_savepoint_list`

**Description:** List Flink detached savepoints in Confluent Platform.

Example:
List Flink detached savepoints with filter filter1.

  $ confluent flink detached-savepoint list --filter name1

**Optional Parameters:**

- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CMF_CERTIFICATE_AUTHORITY_PATH" may be set in place of this flag.
- `client-cert-path`: Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.
- `client-key-path`: Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.
- `filter`: A filter expression to filter by detached savepoint name prefix.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `url`: Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.

**Priority:** high

---

### `confluent_flink_flink_endpoint_list`

**Description:** List Flink endpoint.

Example:
List the available Flink endpoints with current cloud provider and region.

  $ confluent flink endpoint list

**Optional Parameters:**

- `context`: CLI context name.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_flink_flink_endpoint_other`

**Description:** Unset the current Flink endpoint.

Example:
Unset the current Flink endpoint "https://flink.us-east-1.aws.confluent.cloud".

  $ confluent flink endpoint unset

**Priority:** low

---

### `confluent_flink_flink_endpoint_use`

**Description:** Use a Flink endpoint.

Example:
Use "https://flink.us-east-1.aws.confluent.cloud" for subsequent Flink dataplane commands.

  $ confluent flink endpoint use "https://flink.us-east-1.aws.confluent.cloud"

**Priority:** low

---

### `confluent_flink_flink_environment_create`

**Description:** Create a Flink environment.

**Required Parameters:**

- `kubernetes-namespace`: Kubernetes namespace to deploy Flink applications to.

**Optional Parameters:**

- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CMF_CERTIFICATE_AUTHORITY_PATH" may be set in place of this flag.
- `client-cert-path`: Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.
- `client-key-path`: Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.
- `compute-pool-defaults`: JSON string defining the environment's Flink compute pool defaults, or path to a file to read defaults from (with .yml, .yaml or .json extension).
- `defaults`: JSON string defining the environment's Flink application defaults, or path to a file to read defaults from (with .yml, .yaml or .json extension).
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `statement-defaults`: JSON string defining the environment's Flink statement defaults, or path to a file to read defaults from (with .yml, .yaml or .json extension).
- `url`: Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.

**Priority:** medium

---

### `confluent_flink_flink_environment_delete`

**Description:** Delete one or more Flink environments.

**Optional Parameters:**

- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CMF_CERTIFICATE_AUTHORITY_PATH" may be set in place of this flag.
- `client-cert-path`: Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.
- `client-key-path`: Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.
- `force`: Skip the deletion confirmation prompt.
- `url`: Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.

**Priority:** medium

---

### `confluent_flink_flink_environment_describe`

**Description:** Describe a Flink environment.

**Optional Parameters:**

- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CMF_CERTIFICATE_AUTHORITY_PATH" may be set in place of this flag.
- `client-cert-path`: Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.
- `client-key-path`: Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `url`: Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.

**Priority:** high

---

### `confluent_flink_flink_environment_list`

**Description:** List Flink environments.

**Optional Parameters:**

- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CMF_CERTIFICATE_AUTHORITY_PATH" may be set in place of this flag.
- `client-cert-path`: Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.
- `client-key-path`: Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `url`: Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.

**Priority:** high

---

### `confluent_flink_flink_environment_update`

**Description:** Update a Flink environment.

**Optional Parameters:**

- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CMF_CERTIFICATE_AUTHORITY_PATH" may be set in place of this flag.
- `client-cert-path`: Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.
- `client-key-path`: Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.
- `compute-pool-defaults`: JSON string defining the environment's Flink compute pool defaults, or path to a file to read defaults from (with .yml, .yaml or .json extension).
- `defaults`: JSON string defining the environment's Flink application defaults, or path to a file to read defaults from (with .yml, .yaml or .json extension).
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `statement-defaults`: JSON string defining the environment's Flink statement defaults, or path to a file to read defaults from (with .yml, .yaml or .json extension).
- `url`: Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.

**Priority:** medium

---

### `confluent_flink_flink_other`

**Description:** Start Flink interactive SQL client.

Example:
For a Quick Start with examples in context, see https://docs.confluent.io/cloud/current/flink/get-started/quick-start-shell.html.

**Optional Parameters:**

- `compute-pool`: Flink compute pool ID.
- `context`: CLI context name.
- `database`: The database which will be used as the default database. When using Kafka, this is the cluster ID.
- `environment`: Environment ID.
- `service-account`: Service account ID.

**Priority:** low

---

### `confluent_flink_flink_region_list`

**Description:** List Flink regions.

Example:
List the available Flink AWS regions.

  $ confluent flink region list --cloud aws

**Optional Parameters:**

- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `context`: CLI context name.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_flink_flink_region_other`

**Description:** Unset the current Flink cloud and region.

Example:
Unset the current Flink region us-west-1 with cloud provider = AWS.

  $ confluent flink region unset

**Priority:** low

---

### `confluent_flink_flink_region_use`

**Description:** Use a Flink region in subsequent commands.

Example:
Select region "N. Virginia (us-east-1)" for use in subsequent Flink commands.

  $ confluent flink region use --cloud aws --region us-east-1

**Required Parameters:**

- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `region`: Cloud region for Flink (use "confluent flink region list" to see all).

**Priority:** low

---

### `confluent_flink_flink_savepoint_create`

**Description:** Create a Flink savepoint.

Example:
Create a Flink savepoint named "my-savepoint".

  $ confluent flink savepoint create statement "SELECT 1;" --path path-to-savepoint --environment env1

**Required Parameters:**

- `environment`: Name of the Flink environment.

**Optional Parameters:**

- `application`: The name of the Flink application to create the savepoint for.
- `backoff-limit`: Maximum number of retries before the snapshot is considered failed. Set to -1 for unlimited or 0 for no retries.
- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CMF_CERTIFICATE_AUTHORITY_PATH" may be set in place of this flag.
- `client-cert-path`: Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.
- `client-key-path`: Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.
- `format`: The format of the savepoint. Defaults to CANONICAL. (default: `CANONICAL`)
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `path`: The directory where the savepoint should be stored.
- `statement`: The name of the Flink statement to create the savepoint for.
- `url`: Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.

**Priority:** medium

---

### `confluent_flink_flink_savepoint_delete`

**Description:** Delete Flink savepoint in Confluent Platform.

**Required Parameters:**

- `environment`: Name of the Flink environment.

**Optional Parameters:**

- `application`: The Name of the application from which to delete the savepoint.
- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CMF_CERTIFICATE_AUTHORITY_PATH" may be set in place of this flag.
- `client-cert-path`: Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.
- `client-key-path`: Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.
- `force`: Force delete the savepoint.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `statement`: The Name of the statement from which to delete the savepoint.
- `url`: Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.

**Priority:** medium

---

### `confluent_flink_flink_savepoint_describe`

**Description:** Describe a Flink savepoint in Confluent Platform.

**Required Parameters:**

- `environment`: Name of the Flink environment.

**Optional Parameters:**

- `application`: Name of the application to which the savepoint is attached to.
- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CMF_CERTIFICATE_AUTHORITY_PATH" may be set in place of this flag.
- `client-cert-path`: Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.
- `client-key-path`: Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `statement`: Name of the statement to which the savepoint is attached to.
- `url`: Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.

**Priority:** high

---

### `confluent_flink_flink_savepoint_list`

**Description:** List Flink savepoints in Confluent Platform.

**Required Parameters:**

- `environment`: Name of the Flink environment.

**Optional Parameters:**

- `application`: The name of the Flink application to list the savepoints.
- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CMF_CERTIFICATE_AUTHORITY_PATH" may be set in place of this flag.
- `client-cert-path`: Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.
- `client-key-path`: Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `statement`: The name of the Flink statement to list the savepoints.
- `url`: Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.

**Priority:** high

---

### `confluent_flink_flink_savepoint_other`

**Description:** Detach a Flink savepoint in Confluent Platform.

**Required Parameters:**

- `environment`: Name of the Flink environment.
- `application`: Name of the application from which to detach the savepoint.

**Optional Parameters:**

- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CMF_CERTIFICATE_AUTHORITY_PATH" may be set in place of this flag.
- `client-cert-path`: Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.
- `client-key-path`: Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `url`: Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.

**Priority:** low

---

### `confluent_flink_flink_statement_create`

**Description:** Create a Flink SQL statement.

Example:
Create a Flink SQL statement in the current compute pool.

  $ confluent flink statement create --sql "SELECT * FROM table;"

Create a Flink SQL statement named "my-statement" in compute pool "lfcp-123456" with service account "sa-123456", using Kafka cluster "my-cluster" as the default database, and with additional properties.

  $ confluent flink statement create my-statement --sql "SELECT * FROM my-topic;" --compute-pool lfcp-123456 --service-account sa-123456 --database my-cluster --property property1=value1,property2=value2

**Required Parameters:**

- `sql`: The Flink SQL statement.

**Optional Parameters:**

- `compute-pool`: Flink compute pool ID.
- `context`: CLI context name.
- `database`: The database which will be used as the default database. When using Kafka, this is the cluster ID.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `property`: A mechanism to pass properties in the form key=value when creating a Flink statement.
- `service-account`: Service account ID.
- `wait`: Block until the statement is running or has failed.

**Priority:** medium

---

### `confluent_flink_flink_statement_delete`

**Description:** Delete one or more Flink SQL statements.

**Optional Parameters:**

- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.
- `region`: Cloud region for Flink (use "confluent flink region list" to see all).

**Priority:** medium

---

### `confluent_flink_flink_statement_describe`

**Description:** Describe a Flink SQL statement.

**Optional Parameters:**

- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `region`: Cloud region for Flink (use "confluent flink region list" to see all).

**Priority:** high

---

### `confluent_flink_flink_statement_exception_list`

**Description:** List exceptions for a Flink SQL statement.

**Optional Parameters:**

- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `region`: Cloud region for Flink (use "confluent flink region list" to see all).

**Priority:** high

---

### `confluent_flink_flink_statement_list`

**Description:** List Flink SQL statements.

Example:
List running statements.

  $ confluent flink statement list --status running

**Optional Parameters:**

- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `compute-pool`: Flink compute pool ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `region`: Cloud region for Flink (use "confluent flink region list" to see all).
- `status`: Filter the results by statement status.

**Priority:** high

---

### `confluent_flink_flink_statement_other`

**Description:** Resume a Flink SQL statement.

Example:
Request to resume the currently stopped statement "my-statement" using original principal id and under the original compute pool.

  $ confluent flink statement resume my-statement

Request to resume the currently stopped statement "my-statement" using service account "sa-123456".

  $ confluent flink statement resume my-statement --principal sa-123456

Request to resume the currently stopped statement "my-statement" using user account "u-987654".

  $ confluent flink statement resume my-statement --principal u-987654

Request to resume the currently stopped statement "my-statement" and under a different compute pool "lfcp-123456".

  $ confluent flink statement resume my-statement --compute-pool lfcp-123456

Request to resume the currently stopped statement "my-statement" using service account "sa-123456" and under a different compute pool "lfcp-123456".

  $ confluent flink statement resume my-statement --principal sa-123456 --compute-pool lfcp-123456

**Optional Parameters:**

- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `compute-pool`: Flink compute pool ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `principal`: A user or service account the statement runs as.
- `region`: Cloud region for Flink (use "confluent flink region list" to see all).

**Priority:** low

---

### `confluent_flink_flink_statement_stop`

**Description:** Stop a Flink SQL statement.

Example:
Request to stop the currently running statement "my-statement".

  $ confluent flink statement stop my-statement

**Optional Parameters:**

- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `context`: CLI context name.
- `environment`: Environment ID.
- `region`: Cloud region for Flink (use "confluent flink region list" to see all).

**Priority:** low

---

### `confluent_flink_flink_statement_update`

**Description:** Update a Flink SQL statement.

Example:
Request to resume the currently stopped statement "my-statement" using original principal id and under the original compute pool.

  $ confluent flink statement update my-statement --stopped=false

Request to resume the currently stopped statement "my-statement" using service account "sa-123456".

  $ confluent flink statement update my-statement --stopped=false --principal sa-123456

Request to resume the currently stopped statement "my-statement" using user account "u-987654".

  $ confluent flink statement update my-statement --stopped=false --principal u-987654

Request to resume the currently stopped statement "my-statement" and under a different compute pool "lfcp-123456".

  $ confluent flink statement update my-statement --stopped=false --compute-pool lfcp-123456

Request to resume the currently stopped statement "my-statement" using service account "sa-123456" and under a different compute pool "lfcp-123456".

  $ confluent flink statement update my-statement --stopped=false --principal sa-123456 --compute-pool lfcp-123456

Request to stop the currently running statement "my-statement".

  $ confluent flink statement update my-statement --stopped=true

**Optional Parameters:**

- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `compute-pool`: Flink compute pool ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `principal`: A user or service account the statement runs as.
- `region`: Cloud region for Flink (use "confluent flink region list" to see all).
- `stopped`: Request to stop or resume the statement.

**Priority:** medium

---

## IAM Skills

### `confluent_iam_iam_acl_create`

**Description:** Create a centralized ACL.

Example:
Create an ACL that grants the specified user "read" permission to the specified consumer group in the specified Kafka cluster:

  $ confluent iam acl create --allow --principal User:User1 --operation read --consumer-group java_example_group_1 --kafka-cluster <kafka-cluster-id>

Create an ACL that grants the specified user "write" permission on all topics in the specified Kafka cluster:

  $ confluent iam acl create --allow --principal User:User1 --operation write --topic "*" --kafka-cluster <kafka-cluster-id>

Create an ACL that assigns a group "read" access to all topics that use the specified prefix in the specified Kafka cluster:

  $ confluent iam acl create --allow --principal Group:Finance --operation read --topic financial --prefix --kafka-cluster <kafka-cluster-id>

**Required Parameters:**

- `kafka-cluster`: Kafka cluster ID for scope of ACL commands.
- `principal`: Principal for this operation, prefixed with "User:" or "Group:".
- `operation`: Set ACL Operation to: (all, alter, alter-configs, cluster-action, create, delete, describe, describe-configs, idempotent-write, read, write).

**Optional Parameters:**

- `allow`: ACL permission to allow access.
- `client-cert-path`: Path to client cert to be verified by MDS. Include for mTLS authentication.
- `client-key-path`: Path to client private key, include for mTLS authentication.
- `cluster-scope`: Set the cluster resource. With this option the ACL grants access to the provided operations on the Kafka cluster itself.
- `consumer-group`: Set the Consumer Group resource.
- `context`: CLI context name.
- `deny`: ACL permission to restrict access to resource.
- `host`: Set host for access. Only IP addresses are supported. (default: `*`)
- `prefix`: Set to match all resource names prefixed with this value.
- `topic`: Set the topic resource. With this option the ACL grants the provided operations on the topics that start with that prefix, depending on whether the "--prefix" option was also passed.
- `transactional-id`: Set the TransactionalID resource.

**Priority:** medium

---

### `confluent_iam_iam_acl_delete`

**Description:** Delete a centralized ACL.

Example:
Delete an ACL that granted the specified user access to the "test" topic in the specified cluster.

  $ confluent iam acl delete --kafka-cluster <kafka-cluster-id> --allow --principal User:Jane --topic test --operation write --host "*"

**Required Parameters:**

- `kafka-cluster`: Kafka cluster ID for scope of ACL commands.
- `principal`: Principal for this operation, prefixed with "User:" or "Group:".
- `operation`: Set ACL Operation to: (all, alter, alter-configs, cluster-action, create, delete, describe, describe-configs, idempotent-write, read, write).
- `host`: Set host for access. Only IP addresses are supported.

**Optional Parameters:**

- `allow`: ACL permission to allow access.
- `client-cert-path`: Path to client cert to be verified by MDS. Include for mTLS authentication.
- `client-key-path`: Path to client private key, include for mTLS authentication.
- `cluster-scope`: Set the cluster resource. With this option the ACL grants access to the provided operations on the Kafka cluster itself.
- `consumer-group`: Set the Consumer Group resource.
- `context`: CLI context name.
- `deny`: ACL permission to restrict access to resource.
- `force`: Skip the deletion confirmation prompt.
- `prefix`: Set to match all resource names prefixed with this value.
- `topic`: Set the topic resource. With this option the ACL grants the provided operations on the topics that start with that prefix, depending on whether the "--prefix" option was also passed.
- `transactional-id`: Set the TransactionalID resource.

**Priority:** medium

---

### `confluent_iam_iam_acl_list`

**Description:** List centralized ACLs for a resource.

Example:
List all the ACLs for the specified Kafka cluster:

  $ confluent iam acl list --kafka-cluster <kafka-cluster-id>

List all the ACLs for the specified cluster that include "allow" permissions for user "Jane":

  $ confluent iam acl list --kafka-cluster <kafka-cluster-id> --allow --principal User:Jane

**Required Parameters:**

- `kafka-cluster`: Kafka cluster ID for scope of ACL commands.

**Optional Parameters:**

- `allow`: ACL permission to allow access.
- `client-cert-path`: Path to client cert to be verified by MDS. Include for mTLS authentication.
- `client-key-path`: Path to client private key, include for mTLS authentication.
- `cluster-scope`: Set the cluster resource. With this option the ACL grants access to the provided operations on the Kafka cluster itself.
- `consumer-group`: Set the Consumer Group resource.
- `context`: CLI context name.
- `deny`: ACL permission to restrict access to resource.
- `host`: Set host for access. Only IP addresses are supported. (default: `*`)
- `operation`: Set ACL Operation to: (all, alter, alter-configs, cluster-action, create, delete, describe, describe-configs, idempotent-write, read, write).
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `prefix`: Set to match all resource names prefixed with this value.
- `principal`: Principal for this operation, prefixed with "User:" or "Group:".
- `topic`: Set the topic resource. With this option the ACL grants the provided operations on the topics that start with that prefix, depending on whether the "--prefix" option was also passed.
- `transactional-id`: Set the TransactionalID resource.

**Priority:** high

---

### `confluent_iam_iam_certificate_authority_create`

**Description:** Create a certificate authority.

Example:
Create the certificate authority "my-ca" using the certificate chain stored in the "CERTIFICATE_CHAIN" environment variable:

  $ confluent iam certificate-authority create my-ca --description "my certificate authority" --certificate-chain $CERTIFICATE_CHAIN --certificate-chain-filename certificate.pem

An example of a certificate chain:

  -----BEGIN CERTIFICATE-----
  MIIDdTCCAl2gAwIBAgILBAAAAAABFUtaw5QwDQYJKoZIhvcNAQEFBQAwVzELMAkGA1UEBhMCQkUx
  GTAXBgNVBAoTEEdsb2JhbFNpZ24gbnYtc2ExEDAOBgNVBAsTB1Jvb3QgQ0ExGzAZBgNVBAMTEkds
  b2JhbFNpZ24gUm9vdCBDQTAeFw05ODA5MDExMjAwMDBaFw0yODAxMjgxMjAwMDBaMFcxCzAJBgNV
  BAYTAkJFMRkwFwYDVQQKExBHbG9iYWxTaWduIG52LXNhMRAwDgYDVQQLEwdSb290IENBMRswGQYD
  VQQDExJHbG9iYWxTaWduIFJvb3QgQ0EwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDa
  DuaZjc6j40+Kfvvxi4Mla+pIH/EqsLmVEQS98GPR4mdmzxzdzxtIK+6NiY6arymAZavpxy0Sy6sc
  THAHoT0KMM0VjU/43dSMUBUc71DuxC73/OlS8pF94G3VNTCOXkNz8kHp1Wrjsok6Vjk4bwY8iGlb
  Kk3Fp1S4bInMm/k8yuX9ifUSPJJ4ltbcdG6TRGHRjcdGsnUOhugZitVtbNV4FpWi6cgKOOvyJBNP
  c1STE4U6G7weNLWLBYy5d4ux2x8gkasJU26Qzns3dLlwR5EiUWMWea6xrkEmCMgZK9FGqkjWZCrX
  gzT/LCrBbBlDSgeF59N89iFo7+ryUp9/k5DPAgMBAAGjQjBAMA4GA1UdDwEB/wQEAwIBBjAPBgNV
  HRMBAf8EBTADAQH/MB0GA1UdDgQWBBRge2YaRQ2XyolQL30EzTSo//z9SzANBgkqhkiG9w0BAQUF
  AAOCAQEA1nPnfE920I2/7LqivjTFKDK1fPxsnCwrvQmeU79rXqoRSLblCKOzyj1hTdNGCbM+w6Dj
  Y1Ub8rrvrTnhQ7k4o+YviiY776BQVvnGCv04zcQLcFGUl5gE38NflNUVyRRBnMRddWQVDf9VMOyG
  j/8N7yy5Y0b2qvzfvGn9LhJIZJrglfCm7ymPAbEVtQwdpf5pLGkkeB6zpxxxYu7KyJesF12KwvhH
  hm4qxFYxldBniYUr+WymXUadDKqC5JlR3XC321Y9YeRq4VzW9v493kHMB65jUr9TU/Qr6cf9tveC
  X4XSQRjbgbMEHMUfpIBvFSDJ3gyICh3WZlXi/EjJKSZp4A==
  -----END CERTIFICATE-----

**Required Parameters:**

- `description`: Description of the certificate authority.
- `certificate-chain`: A base64 encoded string containing the signing certificate chain.
- `certificate-chain-filename`: The name of the certificate file.

**Optional Parameters:**

- `context`: CLI context name.
- `crl-chain`: A base64 encoded string containing the CRL for this certificate authority.
- `crl-url`: The URL from which to fetch the CRL (Certificate Revocation List) for the certificate authority.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_iam_iam_certificate_authority_delete`

**Description:** Delete one or more certificate authorities.

Example:
Delete certificate authority "op-123456":

  $ confluent iam certificate-authority delete op-123456

**Optional Parameters:**

- `context`: CLI context name.
- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_iam_iam_certificate_authority_describe`

**Description:** Describe a certificate authority.

**Optional Parameters:**

- `context`: CLI context name.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_iam_iam_certificate_authority_list`

**Description:** List certificate authorities.

**Optional Parameters:**

- `context`: CLI context name.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_iam_iam_certificate_authority_update`

**Description:** Update a certificate authority.

Example:
Update the certificate chain for certificate authority "op-123456" using the certificate chain stored in the "CERTIFICATE_CHAIN" environment variable:

  $ confluent iam certificate-authority update op-123456 --certificate-chain $CERTIFICATE_CHAIN --certificate-chain-filename certificate.pem

**Optional Parameters:**

- `certificate-chain`: A base64 encoded string containing the signing certificate chain.
- `certificate-chain-filename`: The name of the certificate file.
- `context`: CLI context name.
- `crl-chain`: A base64 encoded string containing the CRL for this certificate authority.
- `crl-url`: The URL from which to fetch the CRL (Certificate Revocation List) for the certificate authority.
- `description`: Description of the certificate authority.
- `name`: Name of the certificate authority.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_iam_iam_certificate_pool_create`

**Description:** Create a certificate pool.

Example:
Create a certificate pool named "pool-123".

  $ confluent iam certificate-pool create pool-123 --provider provider-123 --description "new description"

**Required Parameters:**

- `provider`: ID of this pool's certificate authority.

**Optional Parameters:**

- `context`: CLI context name.
- `description`: Description of the certificate pool.
- `external-identifier`: External Identifier for this pool.
- `filter`: A supported Common Expression Language (CEL) filter expression. (default: `true`)
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `resource-owner`: The resource ID of the principal who will be assigned resource owner on the created resource. Principal can be a "user", "group-mapping", "service-account", or "identity-pool".

**Priority:** medium

---

### `confluent_iam_iam_certificate_pool_delete`

**Description:** Delete one or more certificate pools.

Example:
Delete certificate pool "pool-123":

  $ confluent iam certificate-pool delete pool-123 --provider provider-123

**Required Parameters:**

- `provider`: ID of this pool's certificate authority.

**Optional Parameters:**

- `context`: CLI context name.
- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_iam_iam_certificate_pool_describe`

**Description:** Describe a certificate pool.

Example:
Describe a certificate pool with ID "pool-123".

  $ confluent iam certificate-pool describe pool-123 --provider provider-123

**Required Parameters:**

- `provider`: ID of this pool's certificate authority.

**Optional Parameters:**

- `context`: CLI context name.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_iam_iam_certificate_pool_list`

**Description:** List certificate pools.

**Required Parameters:**

- `provider`: ID of this pool's certificate authority.

**Optional Parameters:**

- `context`: CLI context name.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_iam_iam_certificate_pool_update`

**Description:** Update a certificate pool.

Example:
Update a certificate pool named "pool-123".

  $ confluent iam certificate-pool update pool-123 --provider provider-123 --description "update pool"

**Required Parameters:**

- `provider`: ID of this pool's certificate authority.

**Optional Parameters:**

- `context`: CLI context name.
- `description`: Description of the certificate pool.
- `external-identifier`: External Identifier for this pool.
- `filter`: A supported Common Expression Language (CEL) filter expression. (default: `true`)
- `name`: Name of the certificate pool.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_iam_iam_group_mapping_create`

**Description:** Create a group mapping.

Example:
Create a group mapping named "demo-group-mapping".

  $ confluent iam group-mapping create demo-group-mapping --description "new description" --filter "\"demo\" in groups"

**Optional Parameters:**

- `context`: CLI context name.
- `description`: Description of the group mapping.
- `filter`: A supported Common Expression Language (CEL) filter expression. (default: `true`)
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_iam_iam_group_mapping_delete`

**Description:** Delete one or more group mappings.

Example:
Delete group mapping "group-123456":

  $ confluent iam group-mapping delete group-123456

**Optional Parameters:**

- `context`: CLI context name.
- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_iam_iam_group_mapping_describe`

**Description:** Describe a group mapping.

**Optional Parameters:**

- `context`: CLI context name.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_iam_iam_group_mapping_list`

**Description:** List group mappings.

**Optional Parameters:**

- `context`: CLI context name.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_iam_iam_group_mapping_update`

**Description:** Update a group mapping.

Example:
Update the description of group mapping "group-123456".

  $ confluent iam group-mapping update group-123456 --description "updated description"

**Optional Parameters:**

- `context`: CLI context name.
- `description`: Description of the group mapping.
- `filter`: A supported Common Expression Language (CEL) filter expression. (default: `true`)
- `name`: Name of the group mapping.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_iam_iam_ip_filter_create`

**Description:** Create an IP filter.

Example:
Create an IP filter named "demo-ip-filter" with operation group "management" and IP groups "ipg-12345" and "ipg-67890":

  $ confluent iam ip-filter create demo-ip-filter --operations management --ip-groups ipg-12345,ipg-67890

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Identifier of the environment for which this filter applies. Without this flag, applies only to the organization.
- `ip-groups`: A comma-separated list of IP group IDs.
- `no-public-networks`: Use in place of ip-groups to reference the no public networks IP Group.
- `operations`: A comma-separated list of operation groups: "MANAGEMENT", "SCHEMA", "FLINK", "KAFKA_MANAGEMENT", "KAFKA_DATA", "KAFKA_DISCOVERY", or "KSQL".
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `resource-group`: Name of resource group: "management" or "multiple". (default: `multiple`)

**Priority:** medium

---

### `confluent_iam_iam_ip_filter_delete`

**Description:** Delete an IP filter.

Example:
Delete IP filter "ipf-12345":

  $ confluent iam ip-filter delete ipf-12345

**Optional Parameters:**

- `context`: CLI context name.
- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_iam_iam_ip_filter_describe`

**Description:** Describe an IP filter.

**Optional Parameters:**

- `context`: CLI context name.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_iam_iam_ip_filter_list`

**Description:** List IP filters.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Identifier of the environment for which this filter applies. Without this flag, applies only to the organization.
- `include-parent-scopes`: Include organization scoped filters when listing filters in an environment.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_iam_iam_ip_filter_update`

**Description:** Update an IP filter.

Example:
Update the name and add an IP group and operation group to IP filter "ipf-abcde":

  $ confluent iam ip-filter update ipf-abcde --name "New Filter Name" --add-ip-groups ipg-12345 --add-operation-groups SCHEMA,FLINK,KAFKA_MANAGEMENT,KAFKA_DATA,KAFKA_DISCOVERY,KSQL

**Optional Parameters:**

- `add-ip-groups`: A comma-separated list of IP groups to add.
- `add-operation-groups`: A comma-separated list of operation groups to add.
- `context`: CLI context name.
- `name`: Updated name of the IP filter.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `remove-ip-groups`: A comma-separated list of IP groups to remove.
- `remove-operation-groups`: A comma-separated list of operation groups to remove.
- `resource-group`: Name of resource group: "management" or "multiple". (default: `multiple`)

**Priority:** medium

---

### `confluent_iam_iam_ip_group_create`

**Description:** Create an IP group.

Example:
Create an IP group named "demo-ip-group" with CIDR blocks "168.150.200.0/24" and "147.150.200.0/24":

  $ confluent iam ip-group create demo-ip-group --cidr-blocks 168.150.200.0/24,147.150.200.0/24

**Required Parameters:**

- `cidr-blocks`: A comma-separated list of CIDR blocks in IP group.

**Optional Parameters:**

- `context`: CLI context name.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_iam_iam_ip_group_delete`

**Description:** Delete an IP group.

Example:
Delete IP group "ipg-12345":

  $ confluent iam ip-group delete ipg-12345

**Optional Parameters:**

- `context`: CLI context name.
- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_iam_iam_ip_group_describe`

**Description:** Describe an IP group.

**Optional Parameters:**

- `context`: CLI context name.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_iam_iam_ip_group_list`

**Description:** List IP groups.

**Optional Parameters:**

- `context`: CLI context name.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_iam_iam_ip_group_update`

**Description:** Update an IP group.

Example:
Update the name and add a CIDR block to IP group "ipg-12345":

  $ confluent iam ip-group update ipg-12345 --name "New Group Name" --add-cidr-blocks 123.234.0.0/16

**Optional Parameters:**

- `add-cidr-blocks`: A comma-separated list of CIDR blocks to add.
- `context`: CLI context name.
- `name`: Updated name of the IP group.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `remove-cidr-blocks`: A comma-separated list of CIDR blocks to remove.

**Priority:** medium

---

### `confluent_iam_iam_pool_create`

**Description:** Create an identity pool.

Example:
Create an identity pool named "demo-identity-pool" with identity provider "op-12345":

  $ confluent iam pool create demo-identity-pool --provider op-12345 --description "new description" --identity-claim claims.sub --filter 'claims.iss=="https://my.issuer.com"'

**Required Parameters:**

- `provider`: ID of this pool's identity provider.
- `identity-claim`: Claim specifying the external identity using this identity pool.

**Optional Parameters:**

- `context`: CLI context name.
- `description`: Description of the identity pool.
- `filter`: A supported Common Expression Language (CEL) filter expression. (default: `true`)
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `resource-owner`: The resource ID of the principal who will be assigned resource owner on the created resource. Principal can be a "user", "group-mapping", "service-account", or "identity-pool".

**Priority:** medium

---

### `confluent_iam_iam_pool_delete`

**Description:** Delete one or more identity pools.

Example:
Delete identity pool "pool-12345":

  $ confluent iam pool delete pool-12345 --provider op-12345

**Required Parameters:**

- `provider`: ID of this pool's identity provider.

**Optional Parameters:**

- `context`: CLI context name.
- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_iam_iam_pool_describe`

**Description:** Describe an identity pool.

**Required Parameters:**

- `provider`: ID of this pool's identity provider.

**Optional Parameters:**

- `context`: CLI context name.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_iam_iam_pool_list`

**Description:** List identity pools.

**Required Parameters:**

- `provider`: ID of this pool's identity provider.

**Optional Parameters:**

- `context`: CLI context name.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_iam_iam_pool_update`

**Description:** Update an identity pool.

Example:
Update the description of identity pool "pool-123456":

  $ confluent iam pool update pool-123456 --provider op-12345 --description "updated description"

**Required Parameters:**

- `provider`: ID of this pool's identity provider.

**Optional Parameters:**

- `context`: CLI context name.
- `description`: Description of the identity pool.
- `filter`: A supported Common Expression Language (CEL) filter expression. (default: `true`)
- `identity-claim`: Claim specifying the external identity using this identity pool.
- `name`: Name of the identity pool.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_iam_iam_provider_create`

**Description:** Create an identity provider.

Example:
Create an identity provider named "demo-identity-provider".

  $ confluent iam provider create demo-identity-provider --description "new description" --jwks-uri https://company.provider.com/oauth2/v1/keys --issuer-uri https://company.provider.com

**Required Parameters:**

- `issuer-uri`: URI of the identity provider issuer.
- `jwks-uri`: JWKS (JSON Web Key Set) URI of the identity provider.

**Optional Parameters:**

- `context`: CLI context name.
- `description`: Description of the identity provider.
- `identity-claim`: The JSON Web Token (JWT) claim to extract the authenticating identity to Confluent resources from Registered Claim Names.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_iam_iam_provider_delete`

**Description:** Delete one or more identity providers.

Example:
Delete identity provider "op-12345":

  $ confluent iam provider delete op-12345

**Optional Parameters:**

- `context`: CLI context name.
- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_iam_iam_provider_describe`

**Description:** Describe an identity provider.

**Optional Parameters:**

- `context`: CLI context name.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_iam_iam_provider_list`

**Description:** List identity providers.

**Optional Parameters:**

- `context`: CLI context name.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_iam_iam_provider_update`

**Description:** Update an identity provider.

Example:
Update the description of identity provider "op-123456".

  $ confluent iam provider update op-123456 --description "updated description"

**Optional Parameters:**

- `context`: CLI context name.
- `description`: Description of the identity provider.
- `identity-claim`: The JSON Web Token (JWT) claim to extract the authenticating identity to Confluent resources from Registered Claim Names.
- `name`: Name of the identity provider.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_iam_iam_rbac_role_binding_create`

**Description:** Create a role binding.

Example:
Grant the role "CloudClusterAdmin" to the principal "User:u-123456" in the environment "env-123456" for the cloud cluster "lkc-123456":

  $ confluent iam rbac role-binding create --principal User:u-123456 --role CloudClusterAdmin --environment env-123456 --cloud-cluster lkc-123456

Grant the role "ResourceOwner" to the principal "User:u-123456", in the environment "env-123456" for the Kafka cluster "lkc-123456" on the resource "Topic:my-topic":

  $ confluent iam rbac role-binding create --principal User:u-123456 --role ResourceOwner --resource Topic:my-topic --environment env-123456 --cloud-cluster lkc-123456 --kafka-cluster lkc-123456

Grant the role "MetricsViewer" to service account "sa-123456":

  $ confluent iam rbac role-binding create --principal User:sa-123456 --role MetricsViewer

Grant the "ResourceOwner" role to principal "User:u-123456" and all subjects for Schema Registry cluster "lsrc-123456" in environment "env-123456":

  $ confluent iam rbac role-binding create --principal User:u-123456 --role ResourceOwner --environment env-123456 --schema-registry-cluster lsrc-123456 --resource "Subject:*"

Grant the "ResourceOwner" role to principal "User:u-123456" and subject "test" for the Schema Registry cluster "lsrc-123456" in the environment "env-123456":

  $ confluent iam rbac role-binding create --principal User:u-123456 --role ResourceOwner --environment env-123456 --schema-registry-cluster lsrc-123456 --resource "Subject:test"

Grant the "ResourceOwner" role to principal "User:u-123456" and all subjects in schema context "schema_context" for Schema Registry cluster "lsrc-123456" in the environment "env-123456":

  $ confluent iam rbac role-binding create --principal User:u-123456 --role ResourceOwner --environment env-123456 --schema-registry-cluster lsrc-123456 --resource "Subject::.schema_context:*"

Grant the "ResourceOwner" role to principal "User:u-123456" and subject "test" in schema context "schema_context" for Schema Registry "lsrc-123456" in the environment "env-123456":

  $ confluent iam rbac role-binding create --principal User:u-123456 --role ResourceOwner --environment env-123456 --schema-registry-cluster lsrc-123456 --resource "Subject::.schema_context:test"

Grant the "FlinkDeveloper" role to principal "User:u-123456" in environment "env-123456":

  $ confluent iam rbac role-binding create --principal User:u-123456 --role FlinkDeveloper --environment env-123456

Grant the "FlinkDeveloper" scoped to Flink compute pool "lfcp-123456" in AWS us-east-1 to principal "User:u-123456":

  $ confluent iam rbac role-binding create --principal User:u-123456 --role FlinkDeveloper --environment env-123456 --flink-region aws.us-east-1 --resource ComputePool:lfcp-123456

**Required Parameters:**

- `role`: Role name of the new role binding.
- `principal`: Principal type and identifier using "Prefix:ID" format.

**Optional Parameters:**

- `cloud-cluster`: Cloud cluster ID for the role binding.
- `current-environment`: Use current environment ID for scope.
- `environment`: Environment ID for scope of role-binding operation.
- `flink-region`: Flink region for the role binding, formatted as "cloud.region".
- `kafka-cluster`: Kafka cluster ID for the role binding.
- `ksql-cluster`: ksqlDB cluster name for the role binding.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `prefix`: Whether the provided resource name is treated as a prefix pattern.
- `resource`: Resource type and identifier using "Prefix:ID" format.
- `schema-registry-cluster`: Schema Registry cluster ID for the role binding.

**Priority:** medium

---

### `confluent_iam_iam_rbac_role_binding_delete`

**Description:** Delete a role binding.

Example:
Delete the role "ResourceOwner" for the resource "Topic:my-topic" on the Kafka cluster "lkc-123456":

  $ confluent iam rbac role-binding delete --principal User:u-123456 --role ResourceOwner --environment env-123456 --kafka-cluster lkc-123456 --resource Topic:my-topic

**Required Parameters:**

- `role`: Role name of the existing role binding.
- `principal`: Principal type and identifier using "Prefix:ID" format.

**Optional Parameters:**

- `cloud-cluster`: Cloud cluster ID for the role binding.
- `current-environment`: Use current environment ID for scope.
- `environment`: Environment ID for scope of role-binding operation.
- `flink-region`: Flink region for the role binding, formatted as "cloud.region".
- `force`: Skip the deletion confirmation prompt.
- `kafka-cluster`: Kafka cluster ID for the role binding.
- `ksql-cluster`: ksqlDB cluster name for the role binding.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `prefix`: Whether the provided resource name is treated as a prefix pattern.
- `resource`: Resource type and identifier using "Prefix:ID" format.
- `schema-registry-cluster`: Schema Registry cluster ID for the role binding.

**Priority:** medium

---

### `confluent_iam_iam_rbac_role_binding_list`

**Description:** List role bindings.

Example:
List the role bindings for the current user:

  $ confluent iam rbac role-binding list --current-user

List the role bindings for user "u-123456":

  $ confluent iam rbac role-binding list --principal User:u-123456

List the role bindings for principals with role "CloudClusterAdmin":

  $ confluent iam rbac role-binding list --role CloudClusterAdmin --current-environment --cloud-cluster lkc-123456

List the role bindings for user "u-123456" with role "CloudClusterAdmin":

  $ confluent iam rbac role-binding list --principal User:u-123456 --role CloudClusterAdmin --environment env-123456 --cloud-cluster lkc-123456

List the role bindings for user "u-123456" for all scopes:

  $ confluent iam rbac role-binding list --principal User:u-123456 --inclusive

List the role bindings for the current user with the environment scope and nested scopes:

  $ confluent iam rbac role-binding list --current-user --environment env-123456 --inclusive

**Optional Parameters:**

- `cloud-cluster`: Cloud cluster ID, which specifies the cloud cluster scope.
- `current-environment`: Use current environment ID for the environment scope.
- `current-user`: List role bindings assigned to the current user.
- `environment`: Environment ID, which specifies the environment scope.
- `flink-region`: Flink region for the role binding, formatted as "cloud.region".
- `inclusive`: List role bindings for specified scopes and nested scopes. Otherwise, list role bindings for the specified scopes. If scopes are unspecified, list only organization-scoped role bindings.
- `kafka-cluster`: Kafka cluster ID, which specifies the Kafka cluster scope.
- `ksql-cluster`: ksqlDB cluster name, which specifies the ksqlDB cluster scope.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `principal`: Principal ID, which limits role bindings to this principal. If unspecified, list all principals and role bindings.
- `resource`: Resource type and identifier using "Prefix:ID" format. If specified with "--role" and no principals, list all principals and role bindings.
- `role`: Predefined role assigned to "--principal". If "--principal" is unspecified, list all principals assigned the role.
- `schema-registry-cluster`: Schema Registry cluster ID, which specifies the Schema Registry cluster scope.

**Priority:** high

---

### `confluent_iam_iam_rbac_role_describe`

**Description:** Describe the resources and operations allowed for a role.

**Optional Parameters:**

- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_iam_iam_rbac_role_list`

**Description:** List the available RBAC roles.

**Optional Parameters:**

- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_iam_iam_service_account_create`

**Description:** Create a service account.

Example:
Create a service account named "my-service-account".

  $ confluent iam service-account create my-service-account --description "new description"

**Required Parameters:**

- `description`: Description of the service account.

**Optional Parameters:**

- `context`: CLI context name.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `resource-owner`: The resource ID of the principal who will be assigned resource owner on the created resource. Principal can be a "user", "group-mapping", "service-account", or "identity-pool".

**Priority:** medium

---

### `confluent_iam_iam_service_account_delete`

**Description:** Delete one or more service accounts.

Example:
Delete service account "sa-123456".

  $ confluent iam service-account delete sa-123456

**Optional Parameters:**

- `context`: CLI context name.
- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_iam_iam_service_account_describe`

**Description:** Describe a service account.

**Optional Parameters:**

- `context`: CLI context name.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_iam_iam_service_account_list`

**Description:** List service accounts.

**Optional Parameters:**

- `context`: CLI context name.
- `display-name`: A comma-separated list of service account display names.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_iam_iam_service_account_other`

**Description:** Unset the current service account.

**Optional Parameters:**

- `context`: CLI context name.

**Priority:** low

---

### `confluent_iam_iam_service_account_update`

**Description:** Update a service account.

Example:
Update the description of service account "sa-123456".

  $ confluent iam service-account update sa-123456 --description "updated description"

**Required Parameters:**

- `description`: Description of the service account.

**Optional Parameters:**

- `context`: CLI context name.

**Priority:** medium

---

### `confluent_iam_iam_service_account_use`

**Description:** Choose a service account to be used in subsequent commands.

**Priority:** low

---

### `confluent_iam_iam_user_delete`

**Description:** Delete one or more users from your organization.

**Optional Parameters:**

- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_iam_iam_user_describe`

**Description:** Describe a user.

**Optional Parameters:**

- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_iam_iam_user_invitation_create`

**Description:** Invite a user to join your organization.

**Priority:** medium

---

### `confluent_iam_iam_user_invitation_list`

**Description:** List the organization's invitations.

**Optional Parameters:**

- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_iam_iam_user_list`

**Description:** List an organization's users.

**Optional Parameters:**

- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_iam_iam_user_update`

**Description:** Update a user.

**Required Parameters:**

- `full-name`: The user's full name.

**Priority:** medium

---

## Kafka Skills

### `confluent_kafka_kafka_acl_create`

**Description:** Create a Kafka ACL.

Example:
You can specify only one of the following flags per command invocation: `--cluster-scope`, `--consumer-group`, `--topic`, or `--transactional-id`. For example, for a consumer to read a topic, you need to grant "read" and "describe" both on the `--consumer-group` and the `--topic` resources, issuing two separate commands:

  $ confluent kafka acl create --allow --service-account sa-55555 --operations read,describe --consumer-group java_example_group_1

  $ confluent kafka acl create --allow --service-account sa-55555 --operations read,describe --topic "*"

**Required Parameters:**

- `operations`: A comma-separated list of ACL operations: (alter, alter-configs, cluster-action, create, delete, describe, describe-configs, idempotent-write, read, write).

**Optional Parameters:**

- `allow`: Access to the resource is allowed.
- `cluster`: Kafka cluster ID.
- `cluster-scope`: Modify ACLs for the cluster.
- `consumer-group`: Modify ACLs for the specified consumer group resource.
- `context`: CLI context name.
- `deny`: Access to the resource is denied.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `prefix`: When this flag is set, the specified resource name is interpreted as a prefix.
- `principal`: Principal for this operation, prefixed with "User:".
- `service-account`: The service account ID.
- `topic`: Modify ACLs for the specified topic resource.
- `transactional-id`: Modify ACLs for the specified TransactionalID resource.

**Priority:** medium

---

### `confluent_kafka_kafka_acl_delete`

**Description:** Delete a Kafka ACL.

**Required Parameters:**

- `operations`: A comma-separated list of ACL operations: (alter, alter-configs, cluster-action, create, delete, describe, describe-configs, idempotent-write, read, write).

**Optional Parameters:**

- `allow`: Access to the resource is allowed.
- `cluster`: Kafka cluster ID.
- `cluster-scope`: Modify ACLs for the cluster.
- `consumer-group`: Modify ACLs for the specified consumer group resource.
- `context`: CLI context name.
- `deny`: Access to the resource is denied.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `prefix`: When this flag is set, the specified resource name is interpreted as a prefix.
- `principal`: Principal for this operation, prefixed with "User:".
- `service-account`: The service account ID.
- `topic`: Modify ACLs for the specified topic resource.
- `transactional-id`: Modify ACLs for the specified TransactionalID resource.

**Priority:** medium

---

### `confluent_kafka_kafka_acl_list`

**Description:** List Kafka ACLs for a resource.

**Optional Parameters:**

- `all`: Include ACLs for deleted principals with integer IDs.
- `cluster`: Kafka cluster ID.
- `cluster-scope`: Modify ACLs for the cluster.
- `consumer-group`: Modify ACLs for the specified consumer group resource.
- `context`: CLI context name.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `prefix`: When this flag is set, the specified resource name is interpreted as a prefix.
- `principal`: Principal for this operation, prefixed with "User:".
- `service-account`: Service account ID.
- `topic`: Modify ACLs for the specified topic resource.
- `transactional-id`: Modify ACLs for the specified TransactionalID resource.

**Priority:** high

---

### `confluent_kafka_kafka_broker_configuration_list`

**Description:** List Kafka broker configurations.

Example:
List all configurations for broker 1.

  $ confluent kafka broker configuration list 1

Describe the "min.insync.replicas" configuration for broker 1.

  $ confluent kafka broker configuration list 1 --config min.insync.replicas

**Optional Parameters:**

- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent REST Proxy.
- `client-cert-path`: Path to client cert to be verified by Confluent REST Proxy. Include for mTLS authentication.
- `client-key-path`: Path to client private key, include for mTLS authentication.
- `config`: Get a specific configuration value.
- `no-authentication`: Include if requests should be made without authentication headers and user will not be prompted for credentials.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `prompt`: Bypass use of available login credentials and prompt for Kafka Rest credentials.
- `url`: Base URL of REST Proxy Endpoint of Kafka Cluster (include "/kafka" for embedded Rest Proxy). Must set flag or CONFLUENT_REST_URL.

**Priority:** high

---

### `confluent_kafka_kafka_broker_configuration_update`

**Description:** Update Kafka broker configurations.

Example:
Update configuration values for broker 1.

  $ confluent kafka broker configuration update 1 --config min.insync.replicas=2,num.partitions=2

**Required Parameters:**

- `config`: A comma-separated list of "key=value" pairs, or path to a configuration file containing a newline-separated list of "key=value" pairs.

**Optional Parameters:**

- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent REST Proxy.
- `client-cert-path`: Path to client cert to be verified by Confluent REST Proxy. Include for mTLS authentication.
- `client-key-path`: Path to client private key, include for mTLS authentication.
- `no-authentication`: Include if requests should be made without authentication headers and user will not be prompted for credentials.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `prompt`: Bypass use of available login credentials and prompt for Kafka Rest credentials.
- `url`: Base URL of REST Proxy Endpoint of Kafka Cluster (include "/kafka" for embedded Rest Proxy). Must set flag or CONFLUENT_REST_URL.

**Priority:** medium

---

### `confluent_kafka_kafka_broker_delete`

**Description:** Delete one or more Kafka brokers.

**Optional Parameters:**

- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent REST Proxy.
- `client-cert-path`: Path to client cert to be verified by Confluent REST Proxy. Include for mTLS authentication.
- `client-key-path`: Path to client private key, include for mTLS authentication.
- `force`: Skip the deletion confirmation prompt.
- `no-authentication`: Include if requests should be made without authentication headers and user will not be prompted for credentials.
- `prompt`: Bypass use of available login credentials and prompt for Kafka Rest credentials.
- `url`: Base URL of REST Proxy Endpoint of Kafka Cluster (include "/kafka" for embedded Rest Proxy). Must set flag or CONFLUENT_REST_URL.

**Priority:** medium

---

### `confluent_kafka_kafka_broker_describe`

**Description:** Describe a Kafka broker.

**Optional Parameters:**

- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent REST Proxy.
- `client-cert-path`: Path to client cert to be verified by Confluent REST Proxy. Include for mTLS authentication.
- `client-key-path`: Path to client private key, include for mTLS authentication.
- `no-authentication`: Include if requests should be made without authentication headers and user will not be prompted for credentials.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `prompt`: Bypass use of available login credentials and prompt for Kafka Rest credentials.
- `url`: Base URL of REST Proxy Endpoint of Kafka Cluster (include "/kafka" for embedded Rest Proxy). Must set flag or CONFLUENT_REST_URL.

**Priority:** high

---

### `confluent_kafka_kafka_broker_list`

**Description:** List Kafka brokers.

**Optional Parameters:**

- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent REST Proxy.
- `client-cert-path`: Path to client cert to be verified by Confluent REST Proxy. Include for mTLS authentication.
- `client-key-path`: Path to client private key, include for mTLS authentication.
- `no-authentication`: Include if requests should be made without authentication headers and user will not be prompted for credentials.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `prompt`: Bypass use of available login credentials and prompt for Kafka Rest credentials.
- `url`: Base URL of REST Proxy Endpoint of Kafka Cluster (include "/kafka" for embedded Rest Proxy). Must set flag or CONFLUENT_REST_URL.

**Priority:** high

---

### `confluent_kafka_kafka_broker_task_list`

**Description:** List broker tasks.

Example:
List remove-broker tasks for broker 1.

  $ confluent kafka broker task list 1 --task-type remove-broker

List broker tasks for all brokers in the cluster

  $ confluent kafka broker task list

**Optional Parameters:**

- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent REST Proxy.
- `client-cert-path`: Path to client cert to be verified by Confluent REST Proxy. Include for mTLS authentication.
- `client-key-path`: Path to client private key, include for mTLS authentication.
- `no-authentication`: Include if requests should be made without authentication headers and user will not be prompted for credentials.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `prompt`: Bypass use of available login credentials and prompt for Kafka Rest credentials.
- `task-type`: Search by task type (add-broker or remove-broker).
- `url`: Base URL of REST Proxy Endpoint of Kafka Cluster (include "/kafka" for embedded Rest Proxy). Must set flag or CONFLUENT_REST_URL.

**Priority:** high

---

### `confluent_kafka_kafka_client_config_create_other`

**Description:** Create a Clojure client configuration file.

Example:
Create a Clojure client configuration file.

  $ confluent kafka client-config create clojure

Create a Clojure client configuration file with arguments.

  $ confluent kafka client-config create clojure --environment env-123 --cluster lkc-123456 --api-key my-key --api-secret my-secret

Create a Clojure client configuration file, redirecting the configuration to a file and the warnings to a separate file.

  $ confluent kafka client-config create clojure 1> my-client-config-file.config 2> my-warnings-file

Create a Clojure client configuration file, redirecting the configuration to a file and keeping the warnings in the console.

  $ confluent kafka client-config create clojure 1> my-client-config-file.config 2>&1

**Optional Parameters:**

- `api-key`: API key.
- `api-secret`: API secret.
- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.

**Priority:** low

---

### `confluent_kafka_kafka_cluster_configuration_describe`

**Description:** Describe a Kafka cluster configuration.

Example:
Describe Kafka cluster configuration "auto.create.topics.enable".

  $ confluent kafka cluster configuration describe auto.create.topics.enable

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_kafka_kafka_cluster_configuration_list`

**Description:** List updated Kafka cluster configurations.

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_kafka_kafka_cluster_configuration_update`

**Description:** Update Kafka cluster configurations.

Example:
Update Kafka cluster configuration "auto.create.topics.enable" to "true".

  $ confluent kafka cluster configuration update --config auto.create.topics.enable=true

**Required Parameters:**

- `config`: A comma-separated list of "key=value" pairs, or path to a configuration file containing a newline-separated list of "key=value" pairs.

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.

**Priority:** medium

---

### `confluent_kafka_kafka_cluster_create`

**Description:** Create a Kafka cluster.

Example:
Create a new dedicated cluster that uses a customer-managed encryption key in GCP:

  $ confluent kafka cluster create sales092020 --cloud gcp --region asia-southeast1 --type dedicated --cku 1 --byok cck-a123z

Create a new dedicated cluster that uses a customer-managed encryption key in AWS:

  $ confluent kafka cluster create my-cluster --cloud aws --region us-west-2 --type dedicated --cku 1 --byok cck-a123z

Create a new Freight cluster that uses a customer-managed encryption key in AWS:

  $ confluent kafka cluster create my-cluster --cloud aws --region us-west-2 --type freight --cku 1 --byok cck-a123z --availability high

For more information, see https://docs.confluent.io/current/cloud/clusters/byok-encrypted-clusters.html.

**Required Parameters:**

- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `region`: Cloud region for Kafka (use "confluent kafka region list" to see all).

**Optional Parameters:**

- `availability`: Specify the availability of the cluster as "single-zone", "multi-zone", "low", or "high". (default: `single-zone`)
- `byok`: Confluent Cloud Key ID of a registered encryption key (use "confluent byok create" to register a key).
- `cku`: Number of Confluent Kafka Units (non-negative). Required for Kafka clusters of type "dedicated".
- `context`: CLI context name.
- `environment`: Environment ID.
- `max-ecku`: Maximum number of Elastic Confluent Kafka Units (eCKUs) that Kafka clusters should auto-scale to. Kafka clusters with "HIGH" availability must have at least two eCKUs.
- `network`: Network ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `type`: Specify the type of the Kafka cluster as "basic", "standard", "enterprise", "freight", or "dedicated". (default: `basic`)

**Priority:** medium

---

### `confluent_kafka_kafka_cluster_delete`

**Description:** Delete one or more Kafka clusters.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_kafka_kafka_cluster_describe`

**Description:** Describe a Kafka cluster.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_kafka_kafka_cluster_endpoint_list`

**Description:** List Kafka cluster endpoints.

Example:
List the available Kafka cluster endpoints for cluster "lkc-123456".

  $ confluent kafka cluster endpoint list --cluster lkc-123456

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_kafka_kafka_cluster_endpoint_use`

**Description:** Use a Kafka cluster endpoint.

Example:
Use "https://lkc-s1232.us-west-2.aws.private.confluent.cloud:443" for subsequent Kafka cluster commands.

  $ confluent kafka cluster endpoint use "https://lkc-s1232.us-west-2.aws.private.confluent.cloud:443"

**Priority:** low

---

### `confluent_kafka_kafka_cluster_list`

**Description:** List Kafka clusters.

**Optional Parameters:**

- `all`: List clusters across all environments.
- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_kafka_kafka_cluster_update`

**Description:** Update a Kafka cluster.

Example:
Update the name and CKU count of a Kafka cluster:

  $ confluent kafka cluster update lkc-123456 --name "New Cluster Name" --cku 3

Update the type of a Kafka cluster from "Basic" to "Standard":

  $ confluent kafka cluster update lkc-123456 --type "standard"

Update the Max eCKU count of a Kafka cluster:

  $ confluent kafka cluster update lkc-123456 --max-ecku 5

**Optional Parameters:**

- `cku`: Number of Confluent Kafka Units. For Kafka clusters of type "dedicated" only. When shrinking a cluster, you must reduce capacity one CKU at a time.
- `context`: CLI context name.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `max-ecku`: Maximum number of Elastic Confluent Kafka Units (eCKUs) that Kafka clusters should auto-scale to. Kafka clusters with "HIGH" availability must have at least two eCKUs.
- `name`: Name of the Kafka cluster.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `type`: Type of the Kafka cluster. Only supports upgrading from "Basic" to "Standard".

**Priority:** medium

---

### `confluent_kafka_kafka_cluster_use`

**Description:** Use a Kafka cluster in subsequent commands.

**Priority:** low

---

### `confluent_kafka_kafka_consumer_group_describe`

**Description:** Describe a Kafka consumer group.

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_kafka_kafka_consumer_group_lag_describe`

**Description:** Describe consumer lag for a Kafka topic partition.

Example:
Describe the consumer lag for topic "my-topic" partition "0" consumed by consumer group "my-consumer-group".

  $ confluent kafka consumer group lag describe my-consumer-group --topic my-topic --partition 0

**Required Parameters:**

- `topic`: Topic name.
- `partition`: Partition ID.

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_kafka_kafka_consumer_group_lag_list`

**Description:** List consumer lags for a Kafka consumer group.

Example:
List consumer lags in consumer group "my-consumer-group".

  $ confluent kafka consumer group lag list my-consumer-group

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_kafka_kafka_consumer_group_lag_other`

**Description:** Summarize consumer lag for a Kafka consumer group.

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** low

---

### `confluent_kafka_kafka_consumer_group_list`

**Description:** List Kafka consumer groups.

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_kafka_kafka_consumer_list`

**Description:** List Kafka consumers.

Example:
List all consumers in consumer group "my-consumer-group".

  $ confluent kafka consumer list --group my-consumer-group

**Required Parameters:**

- `group`: Consumer group ID.

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_kafka_kafka_link_configuration_list`

**Description:** List cluster link configurations.

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_kafka_kafka_link_configuration_update`

**Description:** Update cluster link configurations.

Example:
Update configuration values for the cluster link "my-link".

  $ confluent kafka link configuration update my-link --config my-config.txt

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `config`: A comma-separated list of "key=value" pairs, or path to a configuration file containing a newline-separated list of "key=value" pairs.
- `config-file`: Name of the file containing link configuration overrides. Each property key-value pair should have the format of key=value. Properties are separated by new-line characters.
- `context`: CLI context name.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.

**Priority:** medium

---

### `confluent_kafka_kafka_link_create`

**Description:** Create a new cluster link.

Example:
Create a cluster link, using a configuration file.

  $ confluent kafka link create my-link --source-cluster lkc-123456 --config config.txt

Create a cluster link using command line flags.

  $ confluent kafka link create my-link --source-cluster lkc-123456 --source-bootstrap-server my-host:1234 --source-api-key my-key --source-api-secret my-secret

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `config`: A comma-separated list of "key=value" pairs, or path to a configuration file containing a newline-separated list of "key=value" pairs.
- `config-file`: Name of the file containing link configuration. Each property key-value pair should have the format of key=value. Properties are separated by new-line characters.
- `context`: CLI context name.
- `destination-api-key`: An API key for the destination cluster. This is used for remote cluster authentication links at the source cluster. If specified, the cluster will use SASL_SSL with PLAIN SASL as its mechanism for authentication. If you wish to use another authentication mechanism, do not specify this flag, and add the security configurations in the configuration file.
- `destination-api-secret`: An API secret for the destination cluster. This is used for remote cluster authentication for links at the source cluster. If specified, the cluster will use SASL_SSL with PLAIN SASL as its mechanism for authentication. If you wish to use another authentication mechanism, do not specify this flag, and add the security configurations in the configuration file.
- `destination-bootstrap-server`: Bootstrap server address of the destination cluster for source initiated cluster links. Can alternatively be set in the configuration file using key "bootstrap.servers".
- `destination-cluster`: Destination cluster ID for source initiated cluster links.
- `dry-run`: Validate a link, but do not create it.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `local-api-key`: An API key for the local cluster for bidirectional links. This is used for local cluster authentication if remote link's connection mode is Inbound. If specified, the cluster will use SASL_SSL with PLAIN SASL as its mechanism for authentication. If you wish to use another authentication mechanism, do not specify this flag, and add the security configurations in the configuration file.
- `local-api-secret`: An API secret for the local cluster for bidirectional links. This is used for local cluster authentication if remote link's connection mode is Inbound. If specified, the cluster will use SASL_SSL with PLAIN SASL as its mechanism for authentication. If you wish to use another authentication mechanism, do not specify this flag, and add the security configurations in the configuration file.
- `no-validate`: Create a link even if the source cluster cannot be reached.
- `remote-api-key`: An API key for the remote cluster for bidirectional links. This is used for remote cluster authentication. If specified, the cluster will use SASL_SSL with PLAIN SASL as its mechanism for authentication. If you wish to use another authentication mechanism, do not specify this flag, and add the security configurations in the configuration file.
- `remote-api-secret`: An API secret for the remote cluster for bidirectional links. This is used for remote cluster authentication. If specified, the cluster will use SASL_SSL with PLAIN SASL as its mechanism for authentication. If you wish to use another authentication mechanism, do not specify this flag, and add the security configurations in the configuration file.
- `remote-bootstrap-server`: Bootstrap server address of the remote cluster for bidirectional links. Can alternatively be set in the configuration file using key "bootstrap.servers".
- `remote-cluster`: Remote cluster ID for bidirectional cluster links.
- `source-api-key`: An API key for the source cluster. For links at destination cluster this is used for remote cluster authentication. For links at source cluster this is used for local cluster authentication. If specified, the cluster will use SASL_SSL with PLAIN SASL as its mechanism for authentication. If you wish to use another authentication mechanism, do not specify this flag, and add the security configurations in the configuration file.
- `source-api-secret`: An API secret for the source cluster. For links at destination cluster this is used for remote cluster authentication. For links at source cluster this is used for local cluster authentication. If specified, the cluster will use SASL_SSL with PLAIN SASL as its mechanism for authentication. If you wish to use another authentication mechanism, do not specify this flag, and add the security configurations in the configuration file.
- `source-bootstrap-server`: Bootstrap server address of the source cluster. Can alternatively be set in the configuration file using key "bootstrap.servers".
- `source-cluster`: Source cluster ID.

**Priority:** medium

---

### `confluent_kafka_kafka_link_delete`

**Description:** Delete one or more cluster links.

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.

**Priority:** medium

---

### `confluent_kafka_kafka_link_describe`

**Description:** Describe a cluster link.

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_kafka_kafka_link_list`

**Description:** List cluster links.

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `include-topics`: If set, will list mirrored topics for the links returned.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_kafka_kafka_link_task_list`

**Description:** List a cluster link's tasks.

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_kafka_kafka_mirror_create`

**Description:** Create a mirror topic under the link.

Example:
Create a mirror topic "my-topic" under cluster link "my-link":

  $ confluent kafka mirror create my-topic --link my-link

Create a mirror topic with a custom replication factor and configuration file:

  $ confluent kafka mirror create my-topic --link my-link --replication-factor 5 --config my-config.txt

Create a mirror topic "src_my-topic" where "src_" is the prefix configured on the link:

  $ confluent kafka mirror create src_my-topic --link my-link --source-topic my-topic

**Required Parameters:**

- `link`: Name of cluster link.

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `config`: A comma-separated list of "key=value" pairs, or path to a configuration file containing a newline-separated list of "key=value" pairs.
- `config-file`: Name of a file with additional topic configuration. Each property should be on its own line with the format: key=value.
- `context`: CLI context name.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `replication-factor`: Replication factor. (default: `3`)
- `source-topic`: Name of the source topic to be mirrored over the cluster link. Only required when there is a prefix configured on the link.

**Priority:** medium

---

### `confluent_kafka_kafka_mirror_describe`

**Description:** Describe a mirror topic.

Example:
Describe mirror topic "my-topic" under the link "my-link":

  $ confluent kafka mirror describe my-topic --link my-link

**Required Parameters:**

- `link`: Name of cluster link.

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_kafka_kafka_mirror_list`

**Description:** List mirror topics in a cluster or under a cluster link.

Example:
List all mirror topics in the cluster:

  $ confluent kafka mirror list --cluster lkc-1234

List all active mirror topics under "my-link":

  $ confluent kafka mirror list --link my-link --mirror-status active

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `link`: Name of cluster link.
- `mirror-status`: Mirror topic status. Can be one of "active", "failed", "paused", "stopped", or "pending_stopped". If not specified, list all mirror topics.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_kafka_kafka_mirror_other`

**Description:** Failover mirror topics.

Example:
Failover mirror topics "my-topic-1" and "my-topic-2":

  $ confluent kafka mirror failover my-topic-1 my-topic-2 --link my-link

**Required Parameters:**

- `link`: Name of cluster link.

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `dry-run`: If set, does not actually failover the mirror topic, but simply validates it.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** low

---

### `confluent_kafka_kafka_mirror_state_transition_error_list`

**Description:** Lists the mirror topic's state transition errors.

Example:
Lists mirror topic "my-topic" state transition errors under the link "my-link":

  $ confluent kafka mirror state-transition-error list my-topic --link my-link

**Required Parameters:**

- `link`: Name of cluster link.

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_kafka_kafka_partition_describe`

**Description:** Describe a Kafka partition.

Example:
Describe partition "1" for topic "my_topic".

  $ confluent kafka partition describe 1 --topic my_topic

**Required Parameters:**

- `topic`: Topic name to describe a partition of.

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_kafka_kafka_partition_list`

**Description:** List Kafka partitions.

Example:
List the partitions of topic "my_topic".

  $ confluent kafka partition list --topic my_topic

**Required Parameters:**

- `topic`: Topic name to list partitions of.

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_kafka_kafka_partition_reassignment_list`

**Description:** List ongoing partition reassignments.

Example:
List all partition reassignments for the Kafka cluster.

  $ confluent kafka partition reassignment list

List partition reassignments for topic "my_topic".

  $ confluent kafka partition reassignment list --topic my_topic

List partition reassignments for partition "1" of topic "my_topic".

  $ confluent kafka partition reassignment list 1 --topic my_topic

**Optional Parameters:**

- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent REST Proxy.
- `client-cert-path`: Path to client cert to be verified by Confluent REST Proxy. Include for mTLS authentication.
- `client-key-path`: Path to client private key, include for mTLS authentication.
- `no-authentication`: Include if requests should be made without authentication headers and user will not be prompted for credentials.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `prompt`: Bypass use of available login credentials and prompt for Kafka Rest credentials.
- `topic`: Topic name to search by.
- `url`: Base URL of REST Proxy Endpoint of Kafka Cluster (include "/kafka" for embedded Rest Proxy). Must set flag or CONFLUENT_REST_URL.

**Priority:** high

---

### `confluent_kafka_kafka_quota_create`

**Description:** Create a Kafka client quota.

Example:
Create client quotas for service accounts "sa-1234" and "sa-5678" on cluster "lkc-1234".

  $ confluent kafka quota create my-client-quota --ingress 500 --egress 100 --principals sa-1234,sa-5678 --cluster lkc-1234

Create a default client quota for all principals without an explicit quota assignment.

  $ confluent kafka quota create my-default-quota --ingress 500 --egress 500 --principals "<default>" --cluster lkc-1234

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `description`: Description of quota.
- `egress`: Egress throughput limit for client (bytes/second).
- `environment`: Environment ID.
- `ingress`: Ingress throughput limit for client (bytes/second).
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `principals`: A comma-separated list of service accounts to apply the quota to. Use "<default>" to apply the quota to all service accounts.

**Priority:** medium

---

### `confluent_kafka_kafka_quota_delete`

**Description:** Delete one or more Kafka client quotas.

**Optional Parameters:**

- `force`: Skip the deletion confirmation prompt.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_kafka_kafka_quota_describe`

**Description:** Describe a Kafka client quota.

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_kafka_kafka_quota_list`

**Description:** List client quotas for given cluster.

Example:
List client quotas for cluster "lkc-12345".

  $ confluent kafka quota list --cluster lkc-12345

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `principal`: Principal ID.

**Priority:** high

---

### `confluent_kafka_kafka_quota_update`

**Description:** Update a Kafka client quota.

Example:
Add "sa-12345" to an existing quota and remove "sa-67890".

  $ confluent kafka quota update cq-123ab --add-principals sa-12345 --remove-principals sa-67890

**Optional Parameters:**

- `add-principals`: A comma-separated list of service accounts to add to the quota.
- `context`: CLI context name.
- `description`: Update description.
- `egress`: Update egress limit for quota.
- `ingress`: Update ingress limit for quota.
- `name`: Update name.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `remove-principals`: A comma-separated list of service accounts to remove from the quota.

**Priority:** medium

---

### `confluent_kafka_kafka_region_list`

**Description:** List cloud provider regions.

**Optional Parameters:**

- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_kafka_kafka_replica_list`

**Description:** List Kafka replicas.

Example:
List the replicas for partition 1 of topic "my-topic".

  $ confluent kafka replica list --topic my-topic --partition 1

List the replicas of topic "my-topic".

  $ confluent kafka replica list --topic my-topic

**Required Parameters:**

- `topic`: Topic name.

**Optional Parameters:**

- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent REST Proxy.
- `client-cert-path`: Path to client cert to be verified by Confluent REST Proxy. Include for mTLS authentication.
- `client-key-path`: Path to client private key, include for mTLS authentication.
- `no-authentication`: Include if requests should be made without authentication headers and user will not be prompted for credentials.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `partition`: Partition ID.
- `prompt`: Bypass use of available login credentials and prompt for Kafka Rest credentials.
- `url`: Base URL of REST Proxy Endpoint of Kafka Cluster (include "/kafka" for embedded Rest Proxy). Must set flag or CONFLUENT_REST_URL.

**Priority:** high

---

### `confluent_kafka_kafka_replica_status_list`

**Description:** List Kafka replica statuses.

Example:
List the replica statuses for partition 1 of topic "my-topic".

  $ confluent kafka replica status list --topic my-topic --partition 1

List the replica statuses for topic "my-topic".

  $ confluent kafka replica status list --topic my-topic

**Required Parameters:**

- `topic`: Topic name.

**Optional Parameters:**

- `certificate-authority-path`: Path to a PEM-encoded Certificate Authority to verify the Confluent REST Proxy.
- `client-cert-path`: Path to client cert to be verified by Confluent REST Proxy. Include for mTLS authentication.
- `client-key-path`: Path to client private key, include for mTLS authentication.
- `no-authentication`: Include if requests should be made without authentication headers and user will not be prompted for credentials.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `partition`: Partition ID.
- `prompt`: Bypass use of available login credentials and prompt for Kafka Rest credentials.
- `url`: Base URL of REST Proxy Endpoint of Kafka Cluster (include "/kafka" for embedded Rest Proxy). Must set flag or CONFLUENT_REST_URL.

**Priority:** high

---

### `confluent_kafka_kafka_share_group_consumer_list`

**Description:** List Kafka share group consumers.

Example:
List all consumers in share group "my-share-group".

  $ confluent kafka share-group consumer list --group my-share-group

**Required Parameters:**

- `group`: Share group ID.

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_kafka_kafka_share_group_describe`

**Description:** Describe a Kafka share group.

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_kafka_kafka_share_group_list`

**Description:** List Kafka share groups.

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_kafka_kafka_topic_configuration_list`

**Description:** List Kafka topic configurations.

Example:
List configurations for topic "my-topic".

  $ confluent kafka topic configuration list my-topic

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_kafka_kafka_topic_create`

**Description:** Create a Kafka topic.

Example:
Create a topic named "my_topic" with default options.

  $ confluent kafka topic create my_topic

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `config`: A comma-separated list of configuration overrides ("key=value") for the topic being created.
- `context`: CLI context name.
- `dry-run`: Run the command without committing changes.
- `environment`: Environment ID.
- `if-not-exists`: Exit gracefully if topic already exists.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `partitions`: Number of topic partitions.

**Priority:** medium

---

### `confluent_kafka_kafka_topic_delete`

**Description:** Delete one or more Kafka topics.

Example:
Delete the topics "my_topic" and "my_topic_avro". Use this command carefully as data loss can occur.

  $ confluent kafka topic delete my_topic
  $ confluent kafka topic delete my_topic_avro

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.

**Priority:** medium

---

### `confluent_kafka_kafka_topic_describe`

**Description:** Describe a Kafka topic.

Example:
Describe the "my_topic" topic.

  $ confluent kafka topic describe my_topic

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_kafka_kafka_topic_list`

**Description:** List Kafka topics.

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_kafka_kafka_topic_other`

**Description:** Consume messages from a Kafka topic.

Example:
Consume items from topic "my-topic" and press "Ctrl-C" to exit.

  $ confluent kafka topic consume my-topic --from-beginning

Consume from a cloud Kafka topic named "my_topic" without logging in to Confluent Cloud.

  $ confluent kafka topic consume my_topic --api-key 0000000000000000 --api-secret <API_SECRET> --bootstrap SASL_SSL://pkc-12345.us-west-2.aws.confluent.cloud:9092 --value-format avro --schema-registry-endpoint https://psrc-12345.us-west-2.aws.confluent.cloud --schema-registry-api-key 0000000000000000 --schema-registry-api-secret <SCHEMA_REGISTRY_API_SECRET>

**Optional Parameters:**

- `api-key`: API key.
- `api-secret`: API secret.
- `bootstrap`: Kafka cluster endpoint (Confluent Cloud) or a comma-separated list of broker hosts, each formatted as "host" or "host:port" (Confluent Platform).
- `cert-location`: Path to client's public key (PEM) used for SSL authentication.
- `certificate-authority-path`: File or directory path to one or more Certificate Authority certificates for verifying the broker's key with SSL.
- `client-cert-path`: File or directory path to client certificate to authenticate the Schema Registry client.
- `client-key-path`: File or directory path to client key to authenticate the Schema Registry client.
- `cluster`: Kafka cluster ID.
- `config`: A comma-separated list of configuration overrides ("key=value") for the consumer client. For a full list, see https://docs.confluent.io/platform/current/clients/librdkafka/html/md_CONFIGURATION.html
- `config-file`: The path to the configuration file for the consumer client, in JSON or Avro format.
- `context`: CLI context name.
- `delimiter`: The delimiter separating each key and value. (default: `	`)
- `environment`: Environment ID.
- `from-beginning`: Consume from beginning of the topic.
- `full-header`: Print complete content of message headers.
- `group`: Consumer group ID. (default: `confluent_cli_consumer_<randomly-generated-id>`)
- `key-format`: Format of message key as "string", "avro", "double", "integer", "jsonschema", or "protobuf". Note that schema references are not supported for Avro. (default: `string`)
- `key-location`: Path to client's private key (PEM) used for SSL authentication.
- `key-password`: Private key passphrase for SSL authentication.
- `offset`: The offset from the beginning to consume from.
- `partition`: The partition to consume from. (default: `-1`)
- `password`: SASL_SSL password for use with PLAIN mechanism.
- `print-key`: Print key of the message.
- `print-offset`: Print partition number and offset of the message.
- `protocol`: Specify the broker communication protocol as "PLAINTEXT", "SASL_SSL", or "SSL". (default: `SSL`)
- `sasl-mechanism`: SASL_SSL mechanism used for authentication. (default: `PLAIN`)
- `schema-registry-api-key`: Schema registry API key.
- `schema-registry-api-secret`: Schema registry API secret.
- `schema-registry-context`: The Schema Registry context under which to look up schema ID.
- `schema-registry-endpoint`: Endpoint for Schema Registry cluster.
- `timestamp`: Print message timestamp in milliseconds.
- `username`: SASL_SSL username for use with PLAIN mechanism.
- `value-format`: Format message value as "string", "avro", "double", "integer", "jsonschema", or "protobuf". Note that schema references are not supported for Avro. (default: `string`)

**Priority:** low

---

### `confluent_kafka_kafka_topic_update`

**Description:** Update a Kafka topic.

Example:
Modify the "my_topic" topic to have a retention period of 3 days (259200000 milliseconds).

  $ confluent kafka topic update my_topic --config retention.ms=259200000

**Required Parameters:**

- `config`: A comma-separated list of "key=value" pairs, or path to a configuration file containing a newline-separated list of "key=value" pairs.

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `dry-run`: Run the command without committing changes.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

## Ksql Skills

### `confluent_ksql`

**Description:** Configure ACLs for a ksqlDB cluster.

Example:
Configure ACLs for ksqlDB cluster "lksqlc-12345" for topics "topic_1" and "topic_2":

  $ confluent ksql cluster configure-acls lksqlc-12345 topic_1 topic_2

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `dry-run`: Run the command without committing changes.
- `environment`: Environment ID.
- `kafka-endpoint`: Endpoint to be used for this Kafka cluster.

**Priority:** low

---

## Local Skills

### `confluent_local_local_kafka_broker_configuration_list`

**Description:** List local Kafka broker configurations.

Example:
Describe the "min.insync.replicas" configuration for broker 1.

  $ confluent local broker configuration list 1 --config min.insync.replicas

**Optional Parameters:**

- `config`: Get a specific configuration value.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_local_local_kafka_broker_configuration_update`

**Description:** Update local Kafka broker configurations.

Example:
Update configuration values for broker 1.

  $ confluent kafka broker configuration update 1 --config min.insync.replicas=2,num.partitions=2

**Required Parameters:**

- `config`: A comma-separated list of "key=value" pairs, or path to a configuration file containing a newline-separated list of "key=value" pairs.

**Optional Parameters:**

- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_local_local_kafka_broker_describe`

**Description:** Describe a local Kafka broker.

**Optional Parameters:**

- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_local_local_kafka_broker_list`

**Description:** List local Kafka brokers.

**Optional Parameters:**

- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_local_local_kafka_cluster_configuration_list`

**Description:** List local Kafka cluster configurations.

Example:
List configuration values for the Kafka cluster.

  $ confluent local kafka cluster configuration list

**Optional Parameters:**

- `config`: Get a specific configuration value.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_local_local_kafka_cluster_configuration_update`

**Description:** Update local Kafka cluster configurations.

Example:
Update configuration values for the Kafka cluster.

  $ confluent local kafka cluster configuration update --config min.insync.replicas=2,num.partitions=2

**Optional Parameters:**

- `config`: A comma-separated list of "key=value" pairs, or path to a configuration file containing a newline-separated list of "key=value" pairs.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_local_local_kafka_start`

**Description:** Start a local Apache Kafka instance.

**Optional Parameters:**

- `brokers`: Number of brokers (between 1 and 4, inclusive) in the Confluent Local Kafka cluster. (default: `1`)
- `kafka-rest-port`: Kafka REST port number. (default: `8082`)
- `plaintext-ports`: A comma-separated list of port numbers for plaintext producer and consumer clients for brokers. If not specified, a random free port is used.

**Priority:** low

---

### `confluent_local_local_kafka_stop`

**Description:** Stop the local Apache Kafka service.

**Priority:** low

---

### `confluent_local_local_kafka_topic_configuration_list`

**Description:** List Kafka topic configurations.

Example:
List configurations for topic "test".

  $ confluent local kafka topic configuration list test

**Optional Parameters:**

- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_local_local_kafka_topic_create`

**Description:** Create a Kafka topic.

Example:
Create a topic named "test" with specified configuration parameters.

  $ confluent local kafka topic create test --config cleanup.policy=compact,compression.type=gzip

**Optional Parameters:**

- `config`: A comma-separated list of "key=value" pairs, or path to a configuration file containing a newline-separated list of "key=value" pairs.
- `if-not-exists`: Exit gracefully if topic already exists.
- `partitions`: Number of topic partitions.
- `replication-factor`: Number of replicas.

**Priority:** medium

---

### `confluent_local_local_kafka_topic_delete`

**Description:** Delete one or more Kafka topics.

Example:
Delete the topic "test". Use this command carefully as data loss can occur.

  $ confluent local kafka topic delete test

**Optional Parameters:**

- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_local_local_kafka_topic_describe`

**Description:** Describe a Kafka topic.

Example:
Describe the "test" topic.

  $ confluent local kafka topic describe test

**Optional Parameters:**

- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_local_local_kafka_topic_list`

**Description:** List local Kafka topics.

**Optional Parameters:**

- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_local_local_kafka_topic_other`

**Description:** Consume messages from a Kafka topic.

Example:
Consume message from topic "test" from the beginning and with keys printed.

  $ confluent local kafka topic consume test --from-beginning --print-key

**Optional Parameters:**

- `config`: A comma-separated list of configuration overrides ("key=value") for the consumer client. For a full list, see https://docs.confluent.io/platform/current/clients/librdkafka/html/md_CONFIGURATION.html
- `config-file`: The path to the configuration file for the consumer client, in JSON or Avro format.
- `delimiter`: The delimiter separating each key and value. (default: `	`)
- `from-beginning`: Consume from beginning of the topic.
- `group`: Consumer group ID.
- `offset`: The offset from the beginning to consume from.
- `partition`: The partition to consume from. (default: `-1`)
- `print-key`: Print key of the message.
- `timestamp`: Print message timestamp in milliseconds.

**Priority:** low

---

### `confluent_local_local_kafka_topic_update`

**Description:** Update a Kafka topic.

Example:
Describe the "test" topic.

  $ confluent kafka topic describe test

**Optional Parameters:**

- `config`: A comma-separated list of topics configuration ("key=value") overrides for the topic being created.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_local_local_other`

**Description:** Get the path of the current Confluent run.

Example:
In Linux, running `confluent local current` should resemble the following:

  /tmp/confluent.SpBP4fQi

In macOS, running `confluent local current` should resemble the following:

  /var/folders/cs/1rndf6593qb3kb6r89h50vgr0000gp/T/confluent.000000

**Priority:** low

---

### `confluent_local_local_services_connect_connector_list`

**Description:** List all bundled connectors.

**Priority:** high

---

### `confluent_local_local_services_connect_connector_other`

**Description:** View or set connector configurations.

Example:
Print the current configuration of a connector named `s3-sink`:

  $ confluent local services connect connector config s3-sink

Configure a connector named `wikipedia-file-source` by passing its configuration properties in JSON format.

  $ confluent local services connect connector config wikipedia-file-source --config <path-to-connector>/wikipedia-file-source.json

Configure a connector named `wikipedia-file-source` by passing its configuration properties as Java properties.

  $ confluent local services connect connector config wikipedia-file-source --config <path-to-connector>/wikipedia-file-source.properties

**Optional Parameters:**

- `config`: Configuration file for a connector.

**Priority:** low

---

### `confluent_local_local_services_connect_other`

**Description:** Print logs showing Connect output.

**Optional Parameters:**

- `follow`: Log additional output until the command is interrupted.

**Priority:** low

---

### `confluent_local_local_services_connect_plugin_list`

**Description:** List available Connect plugins.

**Priority:** high

---

### `confluent_local_local_services_connect_start`

**Description:** Start Connect.

**Optional Parameters:**

- `config`: Configure Connect with a specific properties file.

**Priority:** low

---

### `confluent_local_local_services_connect_stop`

**Description:** Stop Connect.

**Priority:** low

---

### `confluent_local_local_services_control_center_other`

**Description:** Print logs showing Control Center output.

**Optional Parameters:**

- `follow`: Log additional output until the command is interrupted.

**Priority:** low

---

### `confluent_local_local_services_control_center_start`

**Description:** Start Control Center.

**Optional Parameters:**

- `config`: Configure Control Center with a specific properties file.

**Priority:** low

---

### `confluent_local_local_services_control_center_stop`

**Description:** Stop Control Center.

**Priority:** low

---

### `confluent_local_local_services_kafka_other`

**Description:** Consume from a Kafka topic.

Example:
Consume Avro data from the beginning of topic called `mytopic1` on a development Kafka cluster on localhost. Assumes Confluent Schema Registry is listening at `http://localhost:8081`.

  $ confluent local services kafka consume mytopic1 --value-format avro --from-beginning

Consume newly arriving non-Avro data from a topic called `mytopic2` on a development Kafka cluster on localhost.

  $ confluent local services kafka consume mytopic2

**Optional Parameters:**

- `bootstrap-server`: The server(s) to connect to. The broker list string has the form HOST1:PORT1,HOST2:PORT2.
- `cloud`: Consume from Confluent Cloud.
- `config`: Change the Confluent Cloud configuration file. (default: `/Users/slateef/.confluent/config.json`)
- `consumer-property`: A mechanism to pass user-defined properties in the form key=value to the consumer.
- `consumer.config`: Consumer config properties file. Note that [consumer-property] takes precedence over this config.
- `enable-systest-events`: Log lifecycle events of the consumer in addition to logging consumed messages. (This is specific for system tests.)
- `formatter`: The name of a class to use for formatting kafka messages for display. (default "kafka.tools.DefaultMessageFormatter")
- `from-beginning`: If the consumer does not already have an established offset to consume from, start with the earliest message present in the log rather than the latest message.
- `group`: The consumer group id of the consumer.
- `isolation-level`: Set to read_committed in order to filter out transactional messages which are not committed. Set to read_uncommitted to read all messages. (default "read_uncommitted")
- `key-deserializer`: 
- `max-messages`: The maximum number of messages to consume before exiting. If not set, consumption is continual.
- `offset`: The offset id to consume from (a non-negative number), or "earliest" which means from beginning, or "latest" which means from end. (default "latest")
- `partition`: The partition to consume from. Consumption starts from the end of the partition unless "--offset" is specified.
- `property`: The properties to initialize the message formatter. Default properties include:
	print.timestamp=true|false
	print.key=true|false
	print.value=true|false
	key.separator=<key.separator>
	line.separator=<line.separator>
	key.deserializer=<key.deserializer>
	value.deserializer=<value.deserializer>
Users can also pass in customized properties for their formatter; more specifically, users can pass in properties keyed with "key.deserializer." and "value.deserializer." prefixes to configure their deserializers.
- `skip-message-on-error`: If there is an error when processing a message, skip it instead of halting.
- `timeout-ms`: If specified, exit if no messages are available for consumption for the specified interval.
- `value-deserializer`: 
- `value-format`: Format output data: avro, json, or protobuf.

- `whitelist`: Regular expression specifying whitelist of topics to include for consumption.

**Priority:** low

---

### `confluent_local_local_services_kafka_rest_other`

**Description:** Print logs showing Kafka REST output.

**Optional Parameters:**

- `follow`: Log additional output until the command is interrupted.

**Priority:** low

---

### `confluent_local_local_services_kafka_rest_start`

**Description:** Start Kafka REST.

**Optional Parameters:**

- `config`: Configure Kafka REST with a specific properties file.

**Priority:** low

---

### `confluent_local_local_services_kafka_rest_stop`

**Description:** Stop Kafka REST.

**Priority:** low

---

### `confluent_local_local_services_kafka_start`

**Description:** Start Apache Kafka®.

**Optional Parameters:**

- `config`: Configure Apache Kafka® with a specific properties file.

**Priority:** low

---

### `confluent_local_local_services_kafka_stop`

**Description:** Stop Apache Kafka®.

**Priority:** low

---

### `confluent_local_local_services_kraft_controller_other`

**Description:** Print logs showing KRaft Controller output.

**Optional Parameters:**

- `follow`: Log additional output until the command is interrupted.

**Priority:** low

---

### `confluent_local_local_services_kraft_controller_start`

**Description:** Start KRaft Controller.

**Optional Parameters:**

- `config`: Configure KRaft Controller with a specific properties file.

**Priority:** low

---

### `confluent_local_local_services_kraft_controller_stop`

**Description:** Stop KRaft Controller.

**Priority:** low

---

### `confluent_local_local_services_ksql_server_other`

**Description:** Print logs showing ksqlDB Server output.

**Optional Parameters:**

- `follow`: Log additional output until the command is interrupted.

**Priority:** low

---

### `confluent_local_local_services_ksql_server_start`

**Description:** Start ksqlDB Server.

**Optional Parameters:**

- `config`: Configure ksqlDB Server with a specific properties file.

**Priority:** low

---

### `confluent_local_local_services_ksql_server_stop`

**Description:** Stop ksqlDB Server.

**Priority:** low

---

### `confluent_local_local_services_list`

**Description:** List all Confluent Platform services.

**Priority:** high

---

### `confluent_local_local_services_other`

**Description:** Check the status of all Confluent Platform services.

**Priority:** low

---

### `confluent_local_local_services_schema_registry_other`

**Description:** Specify an ACL for Schema Registry.

**Optional Parameters:**

- `add`: Indicates you are trying to add ACLs.
- `list`: List all the current ACLs.
- `operation`: Operation that is being authorized. Valid operation names are SUBJECT_READ, SUBJECT_WRITE, SUBJECT_DELETE, SUBJECT_COMPATIBILITY_READ, SUBJECT_COMPATIBILITY_WRITE, GLOBAL_COMPATIBILITY_READ, GLOBAL_COMPATIBILITY_WRITE, and GLOBAL_SUBJECTS_READ.
- `principal`: Principal to which the ACL is being applied to. Use * to apply to all principals.
- `remove`: Indicates you are trying to remove ACLs.
- `subject`: Subject to which the ACL is being applied to. Only applicable for SUBJECT operations. Use * to apply to all subjects.
- `topic`: Topic to which the ACL is being applied to. The corresponding subjects would be topic-key and topic-value. Only applicable for SUBJECT operations. Use * to apply to all subjects.

**Priority:** low

---

### `confluent_local_local_services_schema_registry_start`

**Description:** Start Schema Registry.

**Optional Parameters:**

- `config`: Configure Schema Registry with a specific properties file.

**Priority:** low

---

### `confluent_local_local_services_schema_registry_stop`

**Description:** Stop Schema Registry.

**Priority:** low

---

### `confluent_local_local_services_start`

**Description:** Start all Confluent Platform services.

Example:
Start all available services:

  $ confluent local services start

Start Apache Kafka® and its dependency:

  $ confluent local services kafka start

**Priority:** low

---

### `confluent_local_local_services_stop`

**Description:** Stop all Confluent Platform services.

Example:
Stop all running services:

  $ confluent local services stop

Stop Apache Kafka® and its dependent services.

  $ confluent local services kafka stop

**Priority:** low

---

### `confluent_local_local_services_zookeeper_other`

**Description:** Print logs showing Apache ZooKeeper™ output.

**Optional Parameters:**

- `follow`: Log additional output until the command is interrupted.

**Priority:** low

---

### `confluent_local_local_services_zookeeper_start`

**Description:** Start Apache ZooKeeper™.

**Optional Parameters:**

- `config`: Configure Apache ZooKeeper™ with a specific properties file.

**Priority:** low

---

### `confluent_local_local_services_zookeeper_stop`

**Description:** Stop Apache ZooKeeper™.

**Priority:** low

---

## Login Skills

### `confluent_login`

**Description:** Log in to Confluent Cloud or Confluent Platform.

Example:
Log in to Confluent Cloud.

  $ confluent login

Log in to a specific organization in Confluent Cloud.

  $ confluent login --organization 00000000-0000-0000-0000-000000000000

Log in to Confluent Platform with a MDS URL.

  $ confluent login --url http://localhost:8090

Log in to Confluent Platform with a MDS URL and Certification Authority certificate.

  $ confluent login --url https://localhost:8090 --certificate-authority-path certs/my-cert.crt

Log in to Confluent Platform with SSO even if `CONFLUENT_PLATFORM_USERNAME` and `CONFLUENT_PLATFORM_PASSWORD` are set.

  CONFLUENT_PLATFORM_SSO=true confluent login --url https://localhost:8090 --certificate-authority-path certs/my-cert.crt

**Optional Parameters:**

- `certificate-authority-path`: Self-signed certificate chain in PEM format, for on-premises deployments.
- `certificate-only`: Authenticate using mTLS certificate and key without SSO or username/password.
- `client-cert-path`: Path to client cert to be verified by MDS. Include for mTLS authentication.
- `client-key-path`: Path to client private key, include for mTLS authentication.
- `no-browser`: Do not open a browser window when authenticating using Single Sign-On (SSO).
- `organization`: The Confluent Cloud organization to log in to. If empty, log in to the default organization.
- `prompt`: Bypass non-interactive login and prompt for login credentials.
- `save`: Save username and encrypted password (non-SSO credentials) to the configuration file in your $HOME directory, and to macOS keychain if applicable. You will be logged back in when your token expires, after one hour for Confluent Cloud, or after six hours for Confluent Platform.
- `url`: Metadata Service (MDS) URL, for on-premises deployments.
- `us-gov`: Log in to the Confluent Cloud US Gov environment.

**Priority:** low

---

## Logout Skills

### `confluent_logout`

**Description:** Log out of Confluent Cloud.

**Priority:** low

---

## Network Skills

### `confluent_network_network_access_point_private_link_egr_describe`

**Description:** Describe an egress endpoint.

Example:
Describe egress endpoint "ap-123456".

  $ confluent network access-point private-link egress-endpoint describe ap-123456

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_network_network_access_point_private_link_egres_create`

**Description:** Create an egress endpoint.

Example:
Create an AWS PrivateLink egress endpoint with high availability.

  $ confluent network access-point private-link egress-endpoint create --cloud aws --gateway gw-123456 --service com.amazonaws.vpce.us-west-2.vpce-svc-00000000000000000 --high-availability

Create an Azure Private Link egress endpoint named "my-egress-endpoint".

  $ confluent network access-point private-link egress-endpoint create my-egress-endpoint --cloud azure --gateway gw-123456 --service /subscriptions/0000000/resourceGroups/plsRgName/providers/Microsoft.Network/privateLinkServices/privateLinkServiceName

Create a GCP Private Service Connect egress endpoint named "my-egress-endpoint".

  $ confluent network access-point private-link egress-endpoint create my-egress-endpoint --cloud gcp --gateway gw-123456 --service projects/projectName/regions/us-central1/serviceAttachments/serviceAttachmentName

Create a GCP Private Service Connect egress endpoint named "my-egress-endpoint" for endpoints that connect to Global Google APIs.

  $ confluent network access-point private-link egress-endpoint create my-egress-endpoint --cloud gcp --gateway gw-123456 --service all-google-apis

**Required Parameters:**

- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `service`: Name of an AWS VPC endpoint service, ID of an Azure Private Link service, URI of a GCP Private Service Connect Published Service, or all-google-apis or ALL_GOOGLE_APIS for endpoints that connect to Global Google APIs.
- `gateway`: Gateway ID.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `high-availability`: Enable high availability for AWS egress endpoint.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `subresource`: Name of an Azure Private Link subresource.

**Priority:** medium

---

### `confluent_network_network_access_point_private_link_egres_delete`

**Description:** Delete one or more egress endpoints.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_network_network_access_point_private_link_egres_update`

**Description:** Update an existing egress endpoint.

Example:
Update the name of egress endpoint "ap-123456".

  $ confluent network access-point private-link egress-endpoint update ap-123456 --name my-new-egress-endpoint

**Required Parameters:**

- `name`: Name of the egress endpoint.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_network_network_access_point_private_link_egress__list`

**Description:** List egress endpoints.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `names`: A comma-separated list of display names.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_network_network_access_point_private_link_ing_describe`

**Description:** Describe an ingress endpoint.

Example:
Describe ingress endpoint "ap-123456".

  $ confluent network access-point private-link ingress-endpoint describe ap-123456

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_network_network_access_point_private_link_ingre_create`

**Description:** Create an ingress endpoint.

Example:
Create an AWS PrivateLink ingress endpoint.

  $ confluent network access-point private-link ingress-endpoint create --cloud aws --gateway gw-123456 --vpc-endpoint-id vpce-00000000000000000

**Required Parameters:**

- `cloud`: Specify the cloud provider as "aws".
- `vpc-endpoint-id`: ID of an AWS VPC endpoint.
- `gateway`: Gateway ID.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_network_network_access_point_private_link_ingre_delete`

**Description:** Delete one or more ingress endpoints.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_network_network_access_point_private_link_ingre_update`

**Description:** Update an existing ingress endpoint.

Example:
Update the name of ingress endpoint "ap-123456".

  $ confluent network access-point private-link ingress-endpoint update ap-123456 --name my-new-ingress-endpoint

**Required Parameters:**

- `name`: Name of the ingress endpoint.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_network_network_access_point_private_link_ingress_list`

**Description:** List ingress endpoints.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `names`: A comma-separated list of display names.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_network_network_access_point_private_network__describe`

**Description:** Describe a private network interface.

Example:
Describe private network interface "ap-123456".

  $ confluent network access-point private-network-interface describe ap-123456

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_network_network_access_point_private_network_in_create`

**Description:** Create a private network interface.

Example:
Create an AWS private network interface access point.

  $ confluent network access-point private-network-interface create --cloud aws --gateway gw-123456 --network-interfaces usw2-az1,usw2-az2,usw2-az3 --account 000000000000

**Required Parameters:**

- `cloud`: Specify the cloud provider as "aws".
- `gateway`: Gateway ID.
- `network-interfaces`: A comma-separated list of the IDs of the Elastic Network Interfaces.
- `account`: The AWS account ID associated with the Elastic Network Interfaces.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_network_network_access_point_private_network_in_delete`

**Description:** Delete one or more private network interfaces.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_network_network_access_point_private_network_in_update`

**Description:** Update an existing private network interface.

Example:
Update the name of private network interface "ap-123456".

  $ confluent network access-point private-network-interface update ap-123456 --name my-new-private-network-interface

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `name`: Name of the private network interface.
- `network-interfaces`: A comma-separated list of the IDs of the Elastic Network Interfaces.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_network_network_access_point_private_network_inte_list`

**Description:** List private network interfaces.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `names`: A comma-separated list of display names.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_network_network_create`

**Description:** Create a network.

Example:
Create a Confluent network in AWS with connection type "transitgateway" by specifying zones and CIDR.

  $ confluent network create --cloud aws --region us-west-2 --connection-types transitgateway --zones usw2-az1,usw2-az2,usw2-az4 --cidr 10.1.0.0/16

Create a named Confluent network in AWS with connection type "transitgateway" by specifying zones and CIDR.

  $ confluent network create aws-tgw-network --cloud aws --region us-west-2 --connection-types transitgateway --zones usw2-az1,usw2-az2,usw2-az4 --cidr 10.1.0.0/16

Create a named Confluent network in AWS with connection types "transitgateway" and "peering" by specifying zones and CIDR.

  $ confluent network create aws-tgw-peering-network --cloud aws --region us-west-2 --connection-types transitgateway,peering --zones usw2-az1,usw2-az2,usw2-az4 --cidr 10.1.0.0/16

Create a named Confluent network in AWS with connection type "peering" by specifying zone info.

  $ confluent network create aws-peering-network --cloud aws --region us-west-2 --connection-types peering --zone-info usw2-az1=10.10.0.0/27,usw2-az3=10.10.0.32/27,usw2-az4=10.10.0.64/27

Create a named Confluent network in GCP with connection type "peering" by specifying zones and CIDR.

  $ confluent network create gcp-peering-network --cloud gcp --region us-central1 --connection-types peering --zones us-central1-a,us-central1-b,us-central1-c --cidr 10.1.0.0/16

Create a named Confluent network in Azure with connection type "privatelink" by specifying DNS resolution.

  $ confluent network create azure-pl-network --cloud azure --region eastus2 --connection-types privatelink --dns-resolution chased-private

**Required Parameters:**

- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `region`: Cloud region ID for this network.
- `connection-types`: A comma-separated list of network access types: "privatelink", "peering", or "transitgateway".

**Optional Parameters:**

- `cidr`: A /16 IPv4 CIDR block. Required for networks of connection type "peering" and "transitgateway".
- `context`: CLI context name.
- `dns-resolution`: Specify the DNS resolution as "private" or "chased-private".
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `reserved-cidr`: A /24 IPv4 CIDR block. Can be used for AWS networks of connection type "peering" and "transitgateway".
- `zone-info`: A comma-separated list of "zone=cidr" pairs or CIDR blocks. Each CIDR must be a /27 IPv4 CIDR block.
- `zones`: A comma-separated list of availability zones for this network.

**Priority:** medium

---

### `confluent_network_network_delete`

**Description:** Delete one or more networks.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_network_network_describe`

**Description:** Describe a network.

Example:
Describe Confluent network "n-abcde1".

  $ confluent network describe n-abcde1

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_network_network_dns_forwarder_create`

**Description:** Create a DNS forwarder.

Example:
Create a DNS forwarder.

  $ confluent network dns forwarder create --dns-server-ips 10.200.0.0,10.201.0.0 --gateway gw-123456 --domains abc.com,def.com

Create a named DNS forwarder.

  $ confluent network dns forwarder create my-dns-forwarder --dns-server-ips 10.200.0.0,10.201.0.0 --gateway gw-123456 --domains abc.com,def.com

Create a named DNS forwarder using domain-mapping. This option reads the list of "domainName=zoneName,projectName" mapping from a local file.

  network dns forwarder create my-dns-forwarder-file --gateway gateway-1 --domains example.com --domain-mapping filename

**Required Parameters:**

- `gateway`: Gateway ID.
- `domains`: A comma-separated list of domains for the DNS forwarder to use.

**Optional Parameters:**

- `context`: CLI context name.
- `dns-server-ips`: A comma-separated list of IP addresses for the DNS server.
- `domain-mapping`: Path to a domain mapping file containing domain mappings. Each mapping should have the format of domain=zone,project. Mappings are separated by new-line characters.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_network_network_dns_forwarder_delete`

**Description:** Delete one or more DNS forwarders.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_network_network_dns_forwarder_describe`

**Description:** Describe a DNS forwarder.

Example:
Describe DNS forwarder "dnsf-123456".

  $ confluent network dns forwarder describe dnsf-123456

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_network_network_dns_forwarder_list`

**Description:** List DNS forwarders.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_network_network_dns_forwarder_update`

**Description:** Update an existing DNS forwarder.

Example:
Update the name of DNS forwarder "dnsf-123456".

  $ confluent network dns forwarder update dnsf-123456 --name my-new-dns-forwarder

Update the DNS server IPs and domains of DNS forwarder "dnsf-123456".

  $ confluent network dns forwarder update dnsf-123456 --dns-server-ips 10.200.0.0,10.201.0.0 --domains abc.com,def.com

**Optional Parameters:**

- `context`: CLI context name.
- `dns-server-ips`: A comma-separated list of IP addresses for the DNS server.
- `domain-mapping`: Path to a domain mapping file containing domain mappings. Each mapping should have the format of domain=zone,project. Mappings are separated by new-line characters.
- `domains`: A comma-separated list of domains for the DNS forwarder to use.
- `environment`: Environment ID.
- `name`: Name of the DNS forwarder.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_network_network_dns_record_create`

**Description:** Create a DNS record.

Example:
Create a DNS record.

  $ confluent network dns record create --gateway gw-123456 --private-link-access-point ap-123456 --domain www.example.com

Create a named DNS record.

  $ confluent network dns record create my-dns-record --gateway gw-123456 --private-link-access-point ap-123456 --domain www.example.com

**Required Parameters:**

- `private-link-access-point`: ID of associated PrivateLink Access Point.
- `gateway`: Gateway ID.
- `domain`: Fully qualified domain name of the DNS record.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_network_network_dns_record_delete`

**Description:** Delete one or more DNS records.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_network_network_dns_record_describe`

**Description:** Describe a DNS record.

Example:
Describe DNS recorder "dnsrec-123456".

  $ confluent network dns recorder describe dnsrec-123456

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_network_network_dns_record_list`

**Description:** List DNS records.

Example:
List DNS records with display names "my-dns-record-1" or "my-dns-record-2.

  $ confluent network dns record list --names my-dns-record-1,my-dns-record-2

**Optional Parameters:**

- `context`: CLI context name.
- `domains`: A comma-separated list of fully qualified domain names.
- `environment`: Environment ID.
- `gateways`: A comma-separated list of gateway IDs.
- `names`: A comma-separated list of display names.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `resources`: A comma-separated list of resource IDs.

**Priority:** high

---

### `confluent_network_network_dns_record_update`

**Description:** Update an existing DNS record.

Example:
Update the name of DNS record "dnsrec-123456".

  $ confluent network dns record update dnsrec-123456 --name my-new-dns-record

Update the Privatelink access point of DNS record "dnsrec-123456".

  $ confluent network dns record update dnsrec-123456 --private-link-access-point ap-123456

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `name`: Name of the DNS record.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `private-link-access-point`: ID of associated PrivateLink Access Point.

**Priority:** medium

---

### `confluent_network_network_gateway_create`

**Description:** Create a network gateway.

Example:
Create AWS ingress private link gateway "my-ingress-gateway".

  $ confluent network gateway create my-ingress-gateway --cloud aws --region us-east-1 --type ingress-privatelink

Create AWS egress private link gateway "my-egress-gateway".

  $ confluent network gateway create my-egress-gateway --cloud aws --region us-east-1 --type egress-privatelink

Create AWS private network interface gateway "my-pni-gateway".

  $ confluent network gateway create my-pni-gateway --cloud aws --region us-east-1 --type private-network-interface

**Required Parameters:**

- `cloud`: Specify the cloud provider as "aws" or "azure".
- `type`: Specify the gateway type as "egress-privatelink", "ingress-privatelink", or "private-network-interface".
- `region`: AWS or Azure region of the gateway.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `zones`: A comma-separated list of availability zones for this gateway.

**Priority:** medium

---

### `confluent_network_network_gateway_delete`

**Description:** Delete one or more gateways.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_network_network_gateway_describe`

**Description:** Describe a gateway.

Example:
Describe gateway "gw-123456".

  $ confluent network gateway describe gw-123456

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_network_network_gateway_list`

**Description:** List gateways.

**Optional Parameters:**

- `context`: CLI context name.
- `display-name`: A comma-separated list of display names.
- `environment`: Environment ID.
- `id`: A comma-separated list of gateway IDs.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `phase`: A comma-separated list of phases.
- `region`: A comma-separated list of regions.
- `types`: A comma-separated list of gateway types: "aws-egress-privatelink", "aws-ingress-privatelink", "azure-egress-privatelink", or "gcp-egress-private-service-connect".

**Priority:** high

---

### `confluent_network_network_gateway_update`

**Description:** Update a gateway.

Example:
Update the name of gateway "gw-abc123".

  $ confluent network gateway update gw-abc123 --name new-name

**Required Parameters:**

- `name`: Name of the gateway.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_network_network_ip_address_list`

**Description:** List Confluent Cloud egress public IP addresses.

**Optional Parameters:**

- `address-type`: A comma-separated list of address-types.
- `cloud`: A comma-separated list of cloud providers.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `region`: A comma-separated list of cloud region IDs.
- `services`: A comma-separated list of services.

**Priority:** high

---

### `confluent_network_network_link_endpoint_create`

**Description:** Create a network link endpoint.

Example:
Create a network link endpoint for network "n-123456" and network link service "nls-abcde1".

  $ confluent network link endpoint create --network n-123456 --description "example network link endpoint" --network-link-service nls-abcde1

Create a named network link endpoint for network "n-123456" and network link service "nls-abcde1".

  $ confluent network link endpoint create my-network-link-endpoint --network n-123456 --description "example network link endpoint" --network-link-service nls-abcde1

**Required Parameters:**

- `network`: Network ID.
- `network-link-service`: Network link service ID.

**Optional Parameters:**

- `context`: CLI context name.
- `description`: Network link endpoint description.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_network_network_link_endpoint_delete`

**Description:** Delete one or more network link endpoints.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_network_network_link_endpoint_describe`

**Description:** Describe a network link endpoint.

Example:
Describe network link endpoint "nle-123456".

  $ confluent network link endpoint describe nle-123456

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_network_network_link_endpoint_list`

**Description:** List network link endpoints.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `name`: A comma-separated list of network link endpoint names.
- `network`: A comma-separated list of network IDs.
- `network-link-service`: A comma-separated list of network link service IDs.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `phase`: A comma-separated list of phases.

**Priority:** high

---

### `confluent_network_network_link_endpoint_update`

**Description:** Update an existing network link endpoint.

Example:
Update the name and description of network link endpoint "nle-123456".

  $ confluent network link endpoint update nle-123456 --name my-network-link-endpoint --description "example network link endpoint"

**Optional Parameters:**

- `context`: CLI context name.
- `description`: Description of the network link endpoint.
- `environment`: Environment ID.
- `name`: Name of the network link endpoint.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_network_network_link_service_association_describe`

**Description:** Describe a network link service association.

**Required Parameters:**

- `network-link-service`: Network link service ID.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_network_network_link_service_association_list`

**Description:** List associations for a network link service.

Example:
List associations for network link service "nls-123456".

  $ confluent network link service association list --network-link-service nls-123456

**Required Parameters:**

- `network-link-service`: Network link service ID.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `phase`: A comma-separated list of phases.

**Priority:** high

---

### `confluent_network_network_link_service_create`

**Description:** Create a network link service.

Example:
Create a network link service for network "n-123456" with accepted environments "env-111111" and "env-222222".

  $ confluent network link service create --network n-123456 --description "example network link service" --accepted-environments env-111111,env-222222

Create a named network link service for network "n-123456" with accepted networks "n-abced1" and "n-abcde2".

  $ confluent network link service create my-network-link-service --network n-123456 --description "example network link service" --accepted-networks n-abcde1,n-abcde2

**Required Parameters:**

- `network`: Network ID.

**Optional Parameters:**

- `accepted-environments`: A comma-separated list of environments from which connections can be accepted.
- `accepted-networks`: A comma-separated list of networks from which connections can be accepted.
- `context`: CLI context name.
- `description`: Network link service description.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_network_network_link_service_delete`

**Description:** Delete one or more network link services.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_network_network_link_service_describe`

**Description:** Describe a network link service.

Example:
Describe network link service "nls-123456".

  $ confluent network link service describe nls-123456

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_network_network_link_service_list`

**Description:** List network link services.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `name`: A comma-separated list of network link service names.
- `network`: A comma-separated list of network IDs.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `phase`: A comma-separated list of phases.

**Priority:** high

---

### `confluent_network_network_link_service_update`

**Description:** Update an existing network link service.

Example:
Update the name and description of network link service "nls-123456".

  $ confluent network link service update nls-123456 --name my-network-link-service --description "example network link service"

Update the accepted environments and accepted networks of network link service "nls-123456".

  $ confluent network link service update nls-123456 --description "example network link service" --accepted-environments env-111111 --accepted-networks n-111111,n-222222

**Optional Parameters:**

- `accepted-environments`: A comma-separated list of environments from which connections can be accepted.
- `accepted-networks`: A comma-separated list of networks from which connections can be accepted.
- `context`: CLI context name.
- `description`: Description of the network link service.
- `environment`: Environment ID.
- `name`: Name of the network link service.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_network_network_list`

**Description:** List networks.

**Optional Parameters:**

- `cidr`: A comma-separated list of /16 IPv4 CIDR blocks.
- `cloud`: A comma-separated list of cloud providers.
- `connection-types`: A comma-separated list of network access types: "privatelink", "peering", or "transitgateway".
- `context`: CLI context name.
- `environment`: Environment ID.
- `name`: A comma-separated list of network names.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `phase`: A comma-separated list of phases.
- `region`: A comma-separated list of cloud region IDs.

**Priority:** high

---

### `confluent_network_network_peering_create`

**Description:** Create a peering.

Example:
Create an AWS VPC peering.

  $ confluent network peering create --network n-123456 --cloud aws --cloud-account 123456789012 --virtual-network vpc-1234567890abcdef0 --aws-routes 172.31.0.0/16,10.108.16.0/21

Create a named AWS VPC peering.

  $ confluent network peering create aws-peering --network n-123456 --cloud aws --cloud-account 123456789012 --virtual-network vpc-1234567890abcdef0 --aws-routes 172.31.0.0/16,10.108.16.0/21

Create a named GCP VPC peering.

  $ confluent network peering create gcp-peering --network n-123456 --cloud gcp --cloud-account temp-123456 --virtual-network customer-test-vpc-network --gcp-routes

Create a named Azure VNet peering.

  $ confluent network peering create azure-peering --network n-123456 --cloud azure --cloud-account 1111tttt-1111-1111-1111-111111tttttt --virtual-network /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet --customer-region centralus

**Required Parameters:**

- `network`: Network ID.
- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `cloud-account`: AWS account ID or Google Cloud project ID associated with the VPC that you are peering with Confluent Cloud network or Azure Tenant ID in which your Azure Subscription exists.
- `virtual-network`: AWS VPC ID, name of the Google Cloud VPC, or Azure Resource ID of the VNet that you are peering with Confluent Cloud network.

**Optional Parameters:**

- `aws-routes`: A comma-separated list of CIDR blocks of the AWS VPC that you are peering with Confluent Cloud network. Required for AWS VPC Peering.
- `context`: CLI context name.
- `customer-region`: Cloud region ID of the AWS VPC or Azure VNet that you are peering with Confluent Cloud network.
- `environment`: Environment ID.
- `gcp-routes`: Enable customer route import for Google Cloud VPC Peering.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_network_network_peering_delete`

**Description:** Delete one or more peerings.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_network_network_peering_describe`

**Description:** Describe a peering.

Example:
Describe peering "peer-123456".

  $ confluent network peering describe peer-123456

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_network_network_peering_list`

**Description:** List peerings.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `name`: A comma-separated list of peering names.
- `network`: A comma-separated list of network IDs.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `phase`: A comma-separated list of phases.

**Priority:** high

---

### `confluent_network_network_peering_update`

**Description:** Update an existing peering.

Example:
Update the name of peering "peer-123456".

  $ confluent network peering update peer-123456 --name "new name"

**Required Parameters:**

- `name`: Name of the peering.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_network_network_private_link_access_create`

**Description:** Create a private link access.

Example:
Create an AWS PrivateLink access.

  $ confluent network private-link access create --network n-123456 --cloud aws --cloud-account 123456789012

Create a named AWS PrivateLink access.

  $ confluent network private-link access create aws-private-link-access --network n-123456 --cloud aws --cloud-account 123456789012

Create a named GCP Private Service Connect access.

  $ confluent network private-link access create gcp-private-link-access --network n-123456 --cloud gcp --cloud-account temp-123456

Create a named Azure Private Link access.

  $ confluent network private-link access create azure-private-link-access --network n-123456 --cloud azure --cloud-account 1234abcd-12ab-34cd-1234-123456abcdef

**Required Parameters:**

- `network`: Network ID.
- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `cloud-account`: AWS account ID for the account containing the VPCs you want to connect from using AWS PrivateLink. GCP project ID for the account containing the VPCs that you want to connect from using Private Service Connect. Azure subscription ID for the account containing the VNets you want to connect from using Azure Private Link.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_network_network_private_link_access_delete`

**Description:** Delete one or more private link accesses.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_network_network_private_link_access_describe`

**Description:** Describe a private link access.

Example:
Describe private link access "pla-123456".

  $ confluent network private-link access describe pla-123456

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_network_network_private_link_access_list`

**Description:** List private link accesses.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `name`: A comma-separated list of private link access names.
- `network`: A comma-separated list of network IDs.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `phase`: A comma-separated list of phases.

**Priority:** high

---

### `confluent_network_network_private_link_access_update`

**Description:** Update an existing private link access.

Example:
Update the name of private link access "pla-123456".

  $ confluent network private-link access update pla-123456 --name "new name"

**Required Parameters:**

- `name`: Name of the private link access.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_network_network_private_link_attachment_conne_describe`

**Description:** Describe a private link attachment connection.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_network_network_private_link_attachment_connect_create`

**Description:** Create a private link attachment connection.

Example:
Create a Private Link attachment connection named "aws-private-link-attachment-connection".

  $ confluent network private-link attachment connection create aws-private-link-attachment-connection --cloud aws --endpoint vpce-1234567890abcdef0 --attachment platt-123456

Create a Private Link attachment connection named "gcp-private-link-attachment-connection".

  $ confluent network private-link attachment connection create gcp-private-link-attachment-connection --cloud gcp --endpoint 1234567890123456 --attachment platt-123456

**Required Parameters:**

- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `endpoint`: ID of an endpoint that is connected to either AWS VPC endpoint service, Azure PrivateLink service, or GCP Private Service Connect service.
- `attachment`: Private link attachment ID.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_network_network_private_link_attachment_connect_delete`

**Description:** Delete one or more private link attachment connections.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_network_network_private_link_attachment_connect_update`

**Description:** Update an existing private link attachment connection.

Example:
Update the name of private link attachment connection "plattc-123456".

  $ confluent network private-link attachment connection update plattc-123456 --name "new name"

**Required Parameters:**

- `name`: Name of the private link attachment connection.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_network_network_private_link_attachment_connectio_list`

**Description:** List connections for a private link attachment.

Example:
List connections for private link attachment "platt-123456".

  $ confluent network private-link attachment connection list --attachment platt-123456

**Required Parameters:**

- `attachment`: Private link attachment ID.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_network_network_private_link_attachment_create`

**Description:** Create a private link attachment.

Example:
Create a named Private Link attachment.

  $ confluent network private-link attachment create private-link-attachment --cloud aws --region us-west-2

**Required Parameters:**

- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `region`: Cloud service provider region where the resources to be accessed using the private link attachment are located.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_network_network_private_link_attachment_delete`

**Description:** Delete one or more private link attachments.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_network_network_private_link_attachment_describe`

**Description:** Describe a private link attachment.

Example:
Describe Private Link attachment "platt-123456".

  $ confluent network private-link attachment describe platt-123456

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_network_network_private_link_attachment_list`

**Description:** List private link attachments.

**Optional Parameters:**

- `cloud`: A comma-separated list of cloud providers.
- `context`: CLI context name.
- `environment`: Environment ID.
- `name`: A comma-separated list of private link attachment names.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `phase`: A comma-separated list of phases.
- `region`: A comma-separated list of cloud region IDs.

**Priority:** high

---

### `confluent_network_network_private_link_attachment_update`

**Description:** Update an existing private link attachment.

Example:
Update the name of private link attachment "platt-123456".

  $ confluent network private-link attachment update platt-123456 --name "new name"

**Required Parameters:**

- `name`: Name of the private link attachment.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_network_network_region_list`

**Description:** List cloud provider regions for networking.

**Optional Parameters:**

- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_network_network_transit_gateway_attachment_create`

**Description:** Create a transit gateway attachment.

Example:
Create a transit gateway attachment in AWS.

  $ confluent network transit-gateway-attachment create --network n-123456 --aws-ram-share-arn arn:aws:ram:us-west-2:123456789012:resource-share/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx --aws-transit-gateway tgw-xxxxxxxxxxxxxxxxx --routes 10.0.0.0/16,100.64.0.0/10

Create a named transit gateway attachment in AWS.

  $ confluent network transit-gateway-attachment create my-tgw-attachment --network n-123456 --aws-ram-share-arn arn:aws:ram:us-west-2:123456789012:resource-share/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx --aws-transit-gateway tgw-xxxxxxxxxxxxxxxxx --routes 10.0.0.0/16,100.64.0.0/10

**Required Parameters:**

- `network`: Network ID.
- `aws-ram-share-arn`: AWS Resource Name (ARN) for the AWS Resource Access Manager (RAM) Share of the AWS Transit Gateway that you want Confluent Cloud to be attached to.
- `aws-transit-gateway`: ID of the AWS Transit Gateway that you want Confluent Cloud to be attached to.
- `routes`: A comma-separated list of CIDRs.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_network_network_transit_gateway_attachment_delete`

**Description:** Delete one or more transit gateway attachments.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_network_network_transit_gateway_attachment_describe`

**Description:** Describe a transit gateway attachment.

Example:
Describe transit gateway attachment "tgwa-123456".

  $ confluent network transit-gateway-attachment describe tgwa-123456

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_network_network_transit_gateway_attachment_list`

**Description:** List transit gateway attachments.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `name`: A comma-separated list of transit gateway attachment names.
- `network`: A comma-separated list of network IDs.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `phase`: A comma-separated list of phases.

**Priority:** high

---

### `confluent_network_network_transit_gateway_attachment_update`

**Description:** Update an existing transit gateway attachment.

Example:
Update the name of transit gateway attachment "tgwa-123456".

  $ confluent network transit-gateway-attachment update tgwa-123456 --name "new name"

**Required Parameters:**

- `name`: Name of the transit gateway attachment.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_network_network_update`

**Description:** Update an existing network.

Example:
Update the name of network "n-123456".

  $ confluent network update n-123456 --name "new name"

**Required Parameters:**

- `name`: Name of the network.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

## Organization Skills

### `confluent_organization`

**Description:** Describe the current Confluent Cloud organization.

**Optional Parameters:**

- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

## Plugin Skills

### `confluent_plugin`

**Description:** Install or update official Confluent CLI plugins.

**Priority:** low

---

## Prompt Skills

### `confluent_prompt`

**Description:** Add Confluent CLI context to your terminal prompt.

**Optional Parameters:**

- `format`: The format string to use. See the help for details. (default: `(confluent|%C)`)
- `timeout`: The maximum execution time in milliseconds. (default: `200`)

**Priority:** low

---

## Provider Skills

### `confluent_provider_integration_create`

**Description:** Create a provider integration.

Example:
Create provider integration "s3-provider-integration" associated with AWS IAM role ARN "arn:aws:iam::000000000000:role/my-test-aws-role" in the current environment.

  $ confluent provider-integration create s3-provider-integration --cloud aws --customer-role-arn arn:aws:iam::000000000000:role/my-test-aws-role

Create provider integration "s3-provider-integration" associated with AWS IAM role ARN "arn:aws:iam::000000000000:role/my-test-aws-role" in environment "env-abcdef".

  $ confluent provider-integration create s3-provider-integration --cloud aws --customer-role-arn arn:aws:iam::000000000000:role/my-test-aws-role --environment env-abcdef

**Required Parameters:**

- `customer-role-arn`: Amazon Resource Name (ARN) that identifies the AWS Identity and Access Management (IAM) role that Confluent Cloud assumes when it accesses resources in your AWS account, and must be unique in the same environment.
- `cloud`: Specify the cloud provider as "aws".

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_provider_integration_delete`

**Description:** Delete one or more provider integrations.

Example:
Delete the provider integration "cspi-12345" in the current environment.

  $ confluent provider-integration delete cspi-12345

Delete the provider integrations "cspi-12345" and "cspi-67890" in environment "env-abcdef".

  $ confluent provider-integration delete cspi-12345 cspi-67890 --environment env-abcdef

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_provider_integration_describe`

**Description:** Describe a provider integration.

Example:
Describe provider integration "cspi-12345" in the current environment.

  $ confluent provider-integration describe cspi-12345

Describe provider integration "cspi-12345" in environment "env-abcdef".

  $ confluent provider-integration describe cspi-12345 --environment env-abcdef

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_provider_integration_list`

**Description:** List provider integrations.

Example:
List provider integrations in the current environment.

  $ confluent provider-integration list

List provider integrations in environment "env-abcdef".

  $ confluent provider-integration list --environment env-abcdef

**Optional Parameters:**

- `cloud`: Specify the cloud provider as "aws", "azure", or "gcp".
- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_provider_integration_other`

**Description:** Validate a provider integration.

Example:
Validate Azure provider integration "cspi-123456".

  $ confluent provider-integration v2 validate cspi-123456

Validate GCP provider integration "cspi-789012".

  $ confluent provider-integration v2 validate cspi-789012

**Optional Parameters:**

- `azure-tenant-id`: Customer Azure Tenant ID (for validating Azure provider before update).
- `context`: CLI context name.
- `environment`: Environment ID.
- `gcp-service-account`: Customer Google Service Account (for validating GCP provider before update).

**Priority:** low

---

### `confluent_provider_integration_update`

**Description:** Update a provider integration.

Example:
Update Azure provider integration "cspi-123456" with customer tenant ID.

  $ confluent provider-integration v2 update cspi-123456 --azure-tenant-id 00000000-0000-0000-0000-000000000000

Update GCP provider integration "cspi-789012" with customer service account.

  $ confluent provider-integration v2 update cspi-789012 --gcp-service-account my-sa@my-project.iam.gserviceaccount.com

**Optional Parameters:**

- `azure-tenant-id`: Customer Azure Tenant ID (required for Azure provider).
- `context`: CLI context name.
- `environment`: Environment ID.
- `gcp-service-account`: Customer Google Service Account (required for GCP provider).
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

## Schema Skills

### `confluent_schema_registry_schema_registry_cluster_describe`

**Description:** Describe the Schema Registry cluster for this environment.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.

**Priority:** high

---

### `confluent_schema_registry_schema_registry_cluster_list`

**Description:** List registered Schema Registry clusters.

**Optional Parameters:**

- `client-cert-path`: Path to client cert to be verified by MDS. Include for mTLS authentication.
- `client-key-path`: Path to client private key, include for mTLS authentication.
- `context`: CLI context name.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_schema_registry_schema_registry_cluster_update`

**Description:** Update global mode or compatibility of Schema Registry.

Example:
Update top-level compatibility of Schema Registry.

  $ confluent schema-registry cluster update --compatibility backward

Update the top-level compatibility of Schema Registry and set the compatibility group to "application.version".

  $ confluent schema-registry cluster update --compatibility backward --compatibility-group application.version

Update top-level mode of Schema Registry.

  $ confluent schema-registry cluster update --mode readwrite

**Optional Parameters:**

- `compatibility`: Can be "backward", "backward_transitive", "forward", "forward_transitive", "full", "full_transitive", or "none".
- `compatibility-group`: The name of the compatibility group.
- `context`: CLI context name.
- `environment`: Environment ID.
- `metadata-defaults`: The path to the schema metadata defaults file.
- `metadata-overrides`: The path to the schema metadata overrides file.
- `mode`: Can be "readwrite", "readonly", or "import".
- `ruleset-defaults`: The path to the schema ruleset defaults file.
- `ruleset-overrides`: The path to the schema ruleset overrides file.
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.

**Priority:** medium

---

### `confluent_schema_registry_schema_registry_configuration_delete`

**Description:** Delete top-level or subject-level schema configuration.

Example:
Delete the top-level configuration.

  $ confluent schema-registry configuration delete

Delete the subject-level configuration of subject "payments".

  $ confluent schema-registry configuration delete --subject payments

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.
- `subject`: Subject of the schema.

**Priority:** medium

---

### `confluent_schema_registry_schema_registry_configuration_describe`

**Description:** Describe top-level or subject-level schema configuration.

Example:
Describe the configuration of subject "payments".

  $ confluent schema-registry configuration describe --subject payments

Describe the top-level configuration.

  $ confluent schema-registry configuration describe

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.
- `subject`: Subject of the schema.

**Priority:** high

---

### `confluent_schema_registry_schema_registry_dek_create`

**Description:** Create a Data Encryption Key (DEK).

Example:
Create a DEK with KEK "test", and subject "test-value":

  $ confluent schema-registry dek create --kek-name test --subject test-value --version 1

**Required Parameters:**

- `kek-name`: Name of the Key Encryption Key (KEK).
- `subject`: Subject of the Data Encryption Key (DEK).
- `version`: Version of the Data Encryption Key (DEK).

**Optional Parameters:**

- `algorithm`: Use algorithm "AES128_GCM", "AES256_GCM", or "AES256_SIV" for the Data Encryption Key (DEK).
- `context`: CLI context name.
- `encrypted-key-material`: The encrypted key material for the Data Encryption Key (DEK).
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.

**Priority:** medium

---

### `confluent_schema_registry_schema_registry_dek_delete`

**Description:** Delete a Data Encryption Key (DEK).

**Required Parameters:**

- `kek-name`: Name of the Key Encryption Key (KEK).
- `subject`: Subject of the Data Encryption Key (DEK).

**Optional Parameters:**

- `algorithm`: Use algorithm "AES128_GCM", "AES256_GCM", or "AES256_SIV" for the Data Encryption Key (DEK).
- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.
- `permanent`: Delete the Data Encryption Key (DEK) permanently.
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.
- `version`: Version of the Data Encryption Key (DEK). When not specified, all versions of the Data Encryption Key (DEK) will be deleted.

**Priority:** medium

---

### `confluent_schema_registry_schema_registry_dek_describe`

**Description:** Describe a Data Encryption Key (DEK).

**Required Parameters:**

- `kek-name`: Name of the Key Encryption Key (KEK).
- `subject`: Subject of the Data Encryption Key (DEK).

**Optional Parameters:**

- `algorithm`: Use algorithm "AES128_GCM", "AES256_GCM", or "AES256_SIV" for the Data Encryption Key (DEK).
- `all`: Include soft-deleted Data Encryption Key (DEK).
- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.
- `version`: Version of the Data Encryption Key (DEK). (default: `1`)

**Priority:** high

---

### `confluent_schema_registry_schema_registry_dek_other`

**Description:** Undelete a Data Encryption Key (DEK).

**Required Parameters:**

- `kek-name`: Name of the Key Encryption Key (KEK).
- `subject`: Subject of the Data Encryption Key (DEK).

**Optional Parameters:**

- `algorithm`: Use algorithm "AES128_GCM", "AES256_GCM", or "AES256_SIV" for the Data Encryption Key (DEK).
- `context`: CLI context name.
- `environment`: Environment ID.
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.
- `version`: Version of the Data Encryption Key (DEK). When not specified, all versions of the Data Encryption Key (DEK) will be undeleted.

**Priority:** low

---

### `confluent_schema_registry_schema_registry_dek_subject_list`

**Description:** List Schema Registry Data Encryption Key (DEK) subjects.

Example:
List subjects for the Data Encryption Key (DEK) created with a KEK named "test":

  $ confluent schema-registry dek subject list --kek-name test

**Required Parameters:**

- `kek-name`: Name of the Key Encryption Key (KEK).

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.

**Priority:** high

---

### `confluent_schema_registry_schema_registry_dek_version_list`

**Description:** List Schema Registry Data Encryption Key (DEK) versions.

**Optional Parameters:**

- `algorithm`: Use algorithm "AES128_GCM", "AES256_GCM", or "AES256_SIV" for the Data Encryption Key (DEK).
- `all`: Include soft-deleted Data Encryption Key (DEK).
- `context`: CLI context name.
- `environment`: Environment ID.
- `kek-name`: Name of the Key Encryption Key (KEK).
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.
- `subject`: Subject of the Data Encryption Key (DEK).

**Priority:** high

---

### `confluent_schema_registry_schema_registry_endpoint_list`

**Description:** List all schema registry endpoints.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_schema_registry_schema_registry_exporter_conf_describe`

**Description:** Describe the schema exporter configuration.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `json`)
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.

**Priority:** high

---

### `confluent_schema_registry_schema_registry_exporter_create`

**Description:** Create a new schema exporter.

Example:
Create a new schema exporter.

  $ confluent schema-registry exporter create my-exporter --config config.txt --subjects my-subject1,my-subject2 --subject-format my-\${subject} --context-type custom --context-name my-context

**Optional Parameters:**

- `config`: A comma-separated list of "key=value" pairs, or path to a configuration file containing a newline-separated list of "key=value" pairs.
- `config-file`: Exporter configuration file.
- `context`: CLI context name.
- `context-name`: Exporter context name.
- `context-type`: Exporter context type. One of "auto", "custom", or "none". (default: `auto`)
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.
- `subject-format`: Exporter subject rename format. The format string can contain ${subject}, which will be replaced with the default subject name. (default: `${subject}`)
- `subjects`: A comma-separated list of exporter subjects.

**Priority:** medium

---

### `confluent_schema_registry_schema_registry_exporter_delete`

**Description:** Delete one or more schema exporters.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.

**Priority:** medium

---

### `confluent_schema_registry_schema_registry_exporter_describe`

**Description:** Describe a schema exporter.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.

**Priority:** high

---

### `confluent_schema_registry_schema_registry_exporter_list`

**Description:** List all schema exporters.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.

**Priority:** high

---

### `confluent_schema_registry_schema_registry_exporter_other`

**Description:** Pause schema exporter.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.

**Priority:** low

---

### `confluent_schema_registry_schema_registry_exporter_stat_describe`

**Description:** Describe the schema exporter status.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.

**Priority:** high

---

### `confluent_schema_registry_schema_registry_exporter_update`

**Description:** Update schema exporter.

Example:
Update schema exporter information.

  $ confluent schema-registry exporter update my-exporter --subjects my-subject1,my-subject2 --subject-format my-\${subject} --context-type custom --context-name my-context

Update schema exporter configuration.

  $ confluent schema-registry exporter update my-exporter --config config.txt

**Optional Parameters:**

- `config`: A comma-separated list of "key=value" pairs, or path to a configuration file containing a newline-separated list of "key=value" pairs.
- `config-file`: Exporter configuration file.
- `context`: CLI context name.
- `context-name`: Exporter context name.
- `context-type`: Exporter context type. One of "auto", "custom", or "none". (default: `auto`)
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.
- `subject-format`: Exporter subject rename format. The format string can contain ${subject}, which will be replaced with the default subject name. (default: `${subject}`)
- `subjects`: A comma-separated list of exporter subjects.

**Priority:** medium

---

### `confluent_schema_registry_schema_registry_kek_create`

**Description:** Create a Key Encryption Key (KEK).

Example:
Create a KEK with an AWS KMS key:

  $ confluent schema-registry kek create my-kek --kms-type aws-kms --kms-key arn:aws:kms:us-west-2:037502941121:key/a1231e22-1n78-4l0d-9d50-9pww5faedb54 --kms-properties KeyUsage=ENCRYPT_DECRYPT,KeyState=Enabled

**Required Parameters:**

- `kms-type`: The type of Key Management Service (KMS), typically one of "aws-kms", "azure-kms", or "gcp-kms".
- `kms-key`: The key ID of the Key Management Service (KMS).

**Optional Parameters:**

- `context`: CLI context name.
- `doc`: An optional user-friendly description for the Key Encryption Key (KEK).
- `environment`: Environment ID.
- `kms-properties`: A comma-separated list of additional properties (key=value) used to access the Key Management Service (KMS).
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.
- `shared`: If the DEK Registry has shared access to the Key Management Service (KMS).

**Priority:** medium

---

### `confluent_schema_registry_schema_registry_kek_delete`

**Description:** Delete one or more Key Encryption Keys (KEKs).

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.
- `permanent`: Delete the Key Encryption Key (KEK) permanently.
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.

**Priority:** medium

---

### `confluent_schema_registry_schema_registry_kek_describe`

**Description:** Describe a Key Encryption Key (KEK).

**Optional Parameters:**

- `all`: Include soft-deleted Key Encryption Keys (KEKs).
- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.

**Priority:** high

---

### `confluent_schema_registry_schema_registry_kek_list`

**Description:** List Key Encryption Keys (KEKs).

**Optional Parameters:**

- `all`: Include soft-deleted Key Encryption Keys (KEKs).
- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.

**Priority:** high

---

### `confluent_schema_registry_schema_registry_kek_other`

**Description:** Undelete one or more Key Encryption Keys (KEKs).

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.

**Priority:** low

---

### `confluent_schema_registry_schema_registry_kek_update`

**Description:** Update a Key Encryption Key (KEK).

**Optional Parameters:**

- `context`: CLI context name.
- `doc`: An optional user-friendly description for the Key Encryption Key (KEK).
- `environment`: Environment ID.
- `kms-properties`: A comma-separated list of additional properties (key=value) used to access the Key Management Service (KMS).
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.
- `shared`: If the DEK Registry has shared access to the Key Management Service (KMS).

**Priority:** medium

---

### `confluent_schema_registry_schema_registry_schema_compatibi_other`

**Description:** Validate a schema with a subject version.

Example:
Validate the compatibility of schema "payments" against the latest version of subject "records".

  $ confluent schema-registry schema compatibility validate payments.avsc --type avro --subject records --version latest

**Required Parameters:**

- `subject`: Subject of the schema.
- `type`: Specify the schema type as "avro", "json", or "protobuf".

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `references`: The path to the references file.
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.
- `version`: Version of the schema. Can be a specific version or "latest".

**Priority:** low

---

### `confluent_schema_registry_schema_registry_schema_create`

**Description:** Create a schema.

Example:
Register a new Avro schema.

  $ confluent schema-registry schema create --subject employee --schema employee.avsc --type avro

Where "employee.avsc" may include the following content:

  {
  	"type" : "record",
  	"namespace" : "Example",
  	"name" : "Employee",
  	"fields" : [
  		{ "name" : "Name" , "type" : "string" },
  		{ "name" : "Age" , "type" : "int" }
  	]
  }

For more information on schema types and references, see https://docs.confluent.io/platform/current/schema-registry/fundamentals/serdes-develop/index.html.

**Required Parameters:**

- `schema`: The path to the schema file.
- `subject`: Subject of the schema.

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `metadata`: The path to metadata file.
- `normalize`: Alphabetize the list of schema fields.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `references`: The path to the references file.
- `ruleset`: The path to schema ruleset file.
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.
- `type`: Specify the schema type as "avro", "json", or "protobuf".

**Priority:** medium

---

### `confluent_schema_registry_schema_registry_schema_delete`

**Description:** Delete one or more schema versions.

Example:
Soft delete the latest version of subject "payments".

  $ confluent schema-registry schema delete --subject payments --version latest

Soft delete version "2" of subject "payments".

  $ confluent schema-registry schema delete --subject payments --version 2

Permanently delete version "2" of subject "payments", which must be soft deleted first.

  $ confluent schema-registry schema delete --subject payments --version 2 --permanent

**Required Parameters:**

- `subject`: Subject of the schema.
- `version`: Version of the schema. Can be a specific version, "all", or "latest".

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.
- `permanent`: Permanently delete the schema. You must first soft delete the schema by deleting the schema without this flag.
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.

**Priority:** medium

---

### `confluent_schema_registry_schema_registry_schema_describe`

**Description:** Get schema by ID, or by subject and version.

Example:
Describe the schema with ID "1337".

  $ confluent schema-registry schema describe 1337

Describe the schema with subject "payments" and version "latest".

  $ confluent schema-registry schema describe --subject payments --version latest

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.
- `show-references`: Display the entire schema graph, including references.
- `subject`: Subject of the schema.
- `version`: Version of the schema. Can be a specific version or "latest".

**Priority:** high

---

### `confluent_schema_registry_schema_registry_schema_list`

**Description:** List schemas for a given subject prefix.

Example:
List all schemas for subjects with prefix "my-subject".

  $ confluent schema-registry schema list --subject-prefix my-subject

List all schemas for all subjects in context ":.mycontext:".

  $ confluent schema-registry schema list --subject-prefix :.mycontext:

List all schemas in the default context.

  $ confluent schema-registry schema list

**Optional Parameters:**

- `all`: Include soft-deleted schemas.
- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.
- `subject-prefix`: List schemas for subjects with a given prefix.

**Priority:** high

---

### `confluent_schema_registry_schema_registry_subject_describe`

**Description:** Describe subject versions.

Example:
Retrieve all versions registered under subject "payments" and its compatibility level.

  $ confluent schema-registry subject describe payments

**Optional Parameters:**

- `all`: Include deleted versions.
- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.

**Priority:** high

---

### `confluent_schema_registry_schema_registry_subject_list`

**Description:** List subjects.

**Optional Parameters:**

- `all`: Include deleted subjects.
- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `prefix`: Subject prefix. (default: `:*:`)
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.

**Priority:** high

---

### `confluent_schema_registry_schema_registry_subject_update`

**Description:** Update subject compatibility or mode.

Example:
Update subject-level compatibility of subject "payments".

  $ confluent schema-registry subject update payments --compatibility backward

Update subject-level compatibility of subject "payments" and set compatibility group to "application.version".

  $ confluent schema-registry subject update payments --compatibility backward --compatibility-group application.version

Update subject-level mode of subject "payments".

  $ confluent schema-registry subject update payments --mode readwrite

**Optional Parameters:**

- `compatibility`: Can be "backward", "backward_transitive", "forward", "forward_transitive", "full", "full_transitive", or "none".
- `compatibility-group`: The name of the compatibility group.
- `context`: CLI context name.
- `environment`: Environment ID.
- `metadata-defaults`: The path to the schema metadata defaults file.
- `metadata-overrides`: The path to the schema metadata overrides file.
- `mode`: Can be "readwrite", "readonly", or "import".
- `ruleset-defaults`: The path to the schema ruleset defaults file.
- `ruleset-overrides`: The path to the schema ruleset overrides file.
- `schema-registry-endpoint`: The URL of the Schema Registry cluster.

**Priority:** medium

---

## Secret Skills

### `confluent_secret_other`

**Description:** Add secrets to a configuration properties file.

**Required Parameters:**

- `config-file`: Path to the configuration properties file. File extension must be one of ".json" or ".properties" (key=value pairs).
- `local-secrets-file`: Path to the local encrypted configuration properties file.
- `remote-secrets-file`: Path to the remote encrypted configuration properties file.
- `config`: List of key/value pairs of configuration properties.

**Priority:** low

---

### `confluent_secret_update`

**Description:** Update secrets in a configuration properties file.

**Required Parameters:**

- `config-file`: Path to the configuration properties file. File extension must be one of ".json" or ".properties" (key=value pairs).
- `local-secrets-file`: Path to the local encrypted configuration properties file.
- `remote-secrets-file`: Path to the remote encrypted configuration properties file.
- `config`: List of key/value pairs of configuration properties.

**Priority:** medium

---

## Service Skills

### `confluent_service_quota`

**Description:** List Confluent Cloud service quota values by a scope.

Example:
List Confluent Cloud service quota values for scope "organization".

  $ confluent service-quota list organization

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `environment`: Environment ID.
- `network`: Filter the result by network ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `quota-code`: Filter the result by quota code.

**Priority:** high

---

## Shell Skills

### `confluent_shell`

**Description:** Start an interactive shell.

**Priority:** low

---

## Stream Skills

### `confluent_stream_share_create`

**Description:** Invite a consumer with email.

Example:
Invite a user with email "user@example.com":

  $ confluent stream-share provider invite create --email user@example.com --topic topic-12345 --environment env-123456 --cluster lkc-12345

**Required Parameters:**

- `email`: Email of the user with whom to share the topic.
- `topic`: Topic to be shared.
- `environment`: Environment ID.
- `cluster`: Kafka cluster ID.

**Optional Parameters:**

- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `schema-registry-subjects`: A comma-separated list of Schema Registry subjects.

**Priority:** medium

---

### `confluent_stream_share_delete`

**Description:** Delete one or more consumer shares.

Example:
Delete consumer share "ss-12345":

  $ confluent stream-share consumer share delete ss-12345

**Optional Parameters:**

- `force`: Skip the deletion confirmation prompt.

**Priority:** medium

---

### `confluent_stream_share_describe`

**Description:** Describe a consumer share.

Example:
Describe consumer share "ss-12345":

  $ confluent stream-share consumer share describe ss-12345

**Optional Parameters:**

- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_stream_share_list`

**Description:** List consumer shares.

Example:
List consumer shares for shared resource "sr-12345":

  $ confluent stream-share consumer share list --shared-resource sr-12345

**Optional Parameters:**

- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `shared-resource`: Filter the results by a shared resource.

**Priority:** high

---

### `confluent_stream_share_other`

**Description:** Redeem a stream share token.

Example:
Redeem a stream share token:

  $ confluent stream-share consumer redeem DBBG8xGRfh85ePuk4x5BaENvb25vaGsydXdhejRVNp-pOzCWOLF85LzqcZCq1lVe8OQxSJqQo8XgUMRbtVs5fqbpM5BUKhnHAUcd3C5ip_yWfd3BFRlMVxGQwYo75aSQDb44ACdoAcgjwLH_9YVbk4GJoK-BtZtlpjYSTAIBbhvbFWWOU1bcFyW3HetlyzTIlIjG_UkSKFfDZ_5YNNuw0CBLZQf14J36b4QpSLe05jx9s695tINCm-dyPLX8_pUIqA2ekEZyf86pE7Azh7NBZz00uGZ0FrRl_ir9UvHF1uZ9sID6aZc=

**Optional Parameters:**

- `aws-account`: Consumer's AWS account ID for PrivateLink access.
- `azure-subscription`: Consumer's Azure subscription ID for PrivateLink access.
- `gcp-project`: Consumer's GCP project ID for Private Service Connect access.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** low

---

## Tableflow Skills

### `confluent_tableflow_create`

**Description:** Create a catalog integration.

Example:
Create an Aws Glue catalog integration.

  $ confluent tableflow catalog-integration create my-catalog-integration --type aws --provider-integration cspi-stgce89r7

Create a Snowflake catalog integration.

  $ confluent tableflow catalog-integration create my-catalog-integration --type snowflake --endpoint https://vuser1_polaris.snowflakecomputing.com/ --warehouse catalog-name --allowed-scope session:role:R1 --client-id $CLIENT_ID --client-secret $CLIENT_SECRET

Create a Unity catalog integration.

  $ confluent tableflow catalog-integration create my-catalog-integration --type unity --workspace-endpoint https://dbc-1.cloud.databricks.com --catalog-name tableflow-quickstart-catalog --unity-client-id $CLIENT_ID --unity-client-secret $CLIENT_SECRET

**Required Parameters:**

- `type`: Specify the catalog integration type as "aws", "snowflake", or "unity".

**Optional Parameters:**

- `allowed-scope`: Specify the allowed scope of the Snowflake Open Catalog.
- `catalog-name`: Specify the name of the catalog.
- `client-id`: Specify the client id.
- `client-secret`: Specify the client secret.
- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `endpoint`: Specify the The catalog integration connection endpoint for Snowflake Open Catalog.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `provider-integration`: Specify the provider integration id.
- `unity-client-id`: Specify the Unity client id.
- `unity-client-secret`: Specify the Unity client secret.
- `warehouse`: Specify the warehouse name of the Snowflake Open Catalog.
- `workspace-endpoint`: Specify the Databricks workspace URL associated with the Unity Catalog.

**Priority:** medium

---

### `confluent_tableflow_delete`

**Description:** Delete catalog integrations.

Example:
Delete a catalog integration.

  $ confluent tableflow catalog-integration delete tci-abc123

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** medium

---

### `confluent_tableflow_describe`

**Description:** Describe a catalog integration.

Example:
Describe a catalog integration.

  $ confluent tableflow catalog-integration describe tci-abc123

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_tableflow_list`

**Description:** List catalog integrations.

Example:
List catalog integrations.

  $ confluent tableflow catalog-integration list

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_tableflow_other`

**Description:** Disable topics.

Example:
Disable a Tableflow topic "my-tableflow-topic" related to a Kafka cluster "lkc-123456".

  $ confluent tableflow topic disable my-tableflow-topic --cluster lkc-123456

**Optional Parameters:**

- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.

**Priority:** low

---

### `confluent_tableflow_update`

**Description:** Update a catalog integration.

Example:
Update a catalog integration name.

  $ confluent tableflow catalog-integration update tci-abc123 --name new-name

Create a Snowflake catalog integration.

  $ confluent tableflow catalog-integration update tc-abc123 --endpoint https://vuser1_polaris.snowflakecomputing.com/ --warehouse catalog-name --allowed-scope session:role:R1 --client-id $CLIENT_ID --client-secret $CLIENT_SECRET

**Optional Parameters:**

- `allowed-scope`: Specify the allowed scope of the Snowflake Open Catalog.
- `client-id`: Specify the client id.
- `client-secret`: Specify the client secret.
- `cluster`: Kafka cluster ID.
- `context`: CLI context name.
- `endpoint`: Specify the The catalog integration connection endpoint for Snowflake Open Catalog.
- `environment`: Environment ID.
- `name`: Name of the catalog integration.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)
- `warehouse`: Specify the warehouse name of the Snowflake Open Catalog.

**Priority:** medium

---

## Unified Skills

### `confluent_unified_stream_manager_describe`

**Description:** Describe a Connect cluster.

Example:
Describe a Confluent Platform Connect cluster with the USM ID usmcc-abc123.

  $ confluent unified-stream-manager connect describe usmcc-abc123

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_unified_stream_manager_list`

**Description:** List Connect clusters.

Example:
List Connect clusters.

  $ confluent unified-stream-manager connect list

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `output`: Specify the output format as "human", "json", or "yaml". (default: `human`)

**Priority:** high

---

### `confluent_unified_stream_manager_other`

**Description:** Deregister Connect clusters.

Example:
Deregister a Confluent Platform Connect cluster.

  $ confluent unified-stream-manager connect deregister usmcc-abc123

**Optional Parameters:**

- `context`: CLI context name.
- `environment`: Environment ID.
- `force`: Skip the deletion confirmation prompt.

**Priority:** low

---

## Update Skills

### `confluent_update`

**Description:** Update the Confluent CLI.

**Optional Parameters:**

- `major`: Allow major version updates.
- `no-verify`: Skip checksum verification of new binary.
- `yes`: Update without prompting.

**Priority:** medium

---

## Version Skills

### `confluent_version`

**Description:** Show version of the Confluent CLI.

**Priority:** low

---

