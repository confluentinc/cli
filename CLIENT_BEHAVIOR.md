# How the Confluent CLI handles sessions and saved data

This document outlines what the CLI caches, the TTL (time to live) of cached data, and how expired data is handled.

The values listed below are current, but subject to change.

## Your Confluent Cloud login session

When you run `confluent login`, the CLI stores an authentication token and refresh token (both encrypted) on your machine. The authentication token is used by the CLI to run authenticated commands, while the refresh token is used to automatically refresh the login.

- **TTL:** The exact TTL is determined by the backend, not the CLI.
- **Early renewal:** The CLI treats an authentication token as expired if it is within 60 seconds of the expiration time. This guards against a token expiring mid-command.
- **Expiration handling:** The table below describes the refresh path in descending order of precedence:

  | How you logged in | Refresh mechanism |
  |---|---|
  | Email / password via environment variables `CONFLUENT_CLOUD_EMAIL` and `CONFLUENT_CLOUD_PASSWORD` | Token is refreshed using environment variable values |
  | Email / password with `--save` (on macOS) | Token is refreshed using saved credentials from the keychain |
  | Email / password, SSO, or MFA with valid refresh token | Token is refreshed using the refresh token |

  If the refresh token is expired and you logged in via SSO or MFA, the CLI will initiate the browser login flow to re-authenticate. In all other scenarios, if the CLI is unable to refresh your token, you must log in again.

- **Where it's stored.** Tokens (authentication & refresh) are encrypted and stored in the CLI configuration file in your home directory. For non-SSO/MFA logins with the `--save` flag, username & password are stored in the configuration file (with the password encrypted). On macOS, the username & password are also stored in the keychain.

  Running `confluent logout` removes stored credentials and invalidates the authentication token (for Cloud).

### Access tokens for data plane commands

Some commands (Kafka REST, Schema Registry, Flink, etc.) use a short-lived access token that the CLI requests from Confluent Cloud using the authentication token. The CLI doesn't save these tokens. Each data plane command requests a new token.

## What the CLI remembers

The following values are also stored in the configuration file.

| What | TTL | How to control it |
|---|---|---|
| Feature flag values | 1 hour | Disable with `confluent configuration update disable_feature_flags true` |
| Kafka cluster configuration for clusters you've used (see [below](#kafka-cluster-configuration)) | 7 days or when the active endpoint is updated | Not configurable |
| New version check (for non-package manager installations) | The CLI checks for a new version at most once every 24 hours | Disable with `confluent configuration update disable_update_check true` |

The CLI doesn't cache DNS lookups.

When a cached feature flag value is older than the TTL and the CLI can't fetch a new one, it
prints a warning and uses default values for the rest of the command. It doesn't fall back to the older
cached value.

## Kafka cluster configuration

The CLI keeps an entry in its configuration file for each Kafka cluster you've used.

| What's stored | Set by | Refreshed with the cluster details? |
|---|---|---|
| Cluster ID | Looked up from Confluent Cloud | Yes |
| Cluster display name | Looked up from Confluent Cloud | Yes |
| Bootstrap server | Looked up from Confluent Cloud | Yes |
| REST endpoint | Looked up from Confluent Cloud | Yes |
| API keys and (encrypted) secrets for this cluster | `confluent api-key store` | No (removed) |
| Which stored API key is selected | `confluent api-key use` | No (removed) |
| When the details were last refreshed | The CLI | Yes |
