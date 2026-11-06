# Network connections made by the Confluent CLI

> This page lists what the Confluent CLI connects to for Confluent Cloud. Unless noted, the CLI makes these connections itself. Optional integrations configured separately, such as key management for field-level encryption, aren't listed. Confluent Cloud for Government uses different hostnames and isn't covered here.

Contents:

- [Confluent-operated services](#confluent-operated-services)
- [Confluent Cloud clusters](#confluent-cloud-clusters)
- [Third-party services](#third-party-services)
- [Notes](#notes)

## Confluent-operated services

| Destination | Used for | When | Authentication |
|---|---|---|---|
| `api.confluent.cloud`, `confluent.cloud/api/` | Management of control plane resources such as environments and API keys, login and token refresh, billing, usage limits, data plane token exchange, and user feedback | Resource management, login, billing and feedback commands, and any command that calls a cluster endpoint | Username and password at login, then authentication token |
| `api.confluent.cloud/cli/v1/usages` | CLI usage telemetry | After each Confluent Cloud command while logged in, and after a crash | Authentication token |
| `confluent.cloud/ldapi/` | Feature flags, evaluated for the user's organization, environment and CLI version | Before each Confluent Cloud command if cached flags are stale or missing | None |
| `api.telemetry.confluent.cloud` | Reading resource metrics | Commands that use resource metrics, e.g. `confluent schema-registry cluster describe` | Data plane token |
| `api.logging.confluent.cloud` | Searching connector logs | `confluent connect logs` | Data plane token |
| `login.confluent.io` | Confluent Cloud SSO and MFA login. Sign-in happens in the browser; the CLI listens on `127.0.0.1:26635` (local only) to receive the login code | Cloud login with SSO or MFA; token refresh | Login code (PKCE), then refresh token |
| `packages.confluent.io` | Checking for and downloading / verifying new CLI versions & release notes | About once a day before a command (except `update` or `prompt`) and on `confluent update` | None |
| `api.hub.confluent.io` | Finding and installing Connect plugins | `confluent connect plugin install` | None |
| `www.confluent.io/get-started` | Opening the sign-up page in a browser | `confluent cloud-signup` | None |

## Confluent Cloud clusters

By default, the CLI looks up each address from Confluent Cloud. The Override column shows how to use a different one.

Hostnames vary with a cluster's type and networking, such as public, private link or peering. The examples below show one common form. Ports are 443 for HTTPS endpoints and 9092 for Kafka. To see a cluster's endpoints, run `confluent endpoint endpoint list`.

| Destination | Used for | When | Override | Authentication |
|---|---|---|---|---|
| Kafka REST endpoint, for example `<pkc-id>.<region>.<cloud>.confluent.cloud:443` | Topics, ACLs, consumer groups and similar | Kafka commands | `--kafka-endpoint` for one command, or `confluent kafka cluster endpoint use` for subsequent commands | Data plane token |
| Kafka bootstrap server, for example `<pkc-id>.<region>.<cloud>.confluent.cloud:9092` | Producing and consuming messages | Commands like `confluent kafka topic produce` and `consume` | `--bootstrap`, or `confluent kafka cluster endpoint use`, which sets both addresses | Kafka API key |
| Schema Registry endpoint, for example `<psrc-id>.<region>.<cloud>.confluent.cloud:443` | Schemas | Schema commands; produce or consume with schemas | `--schema-registry-endpoint` | Schema Registry API key or data plane token |
| Flink endpoint, for example `flink.<region>.<cloud>.confluent.cloud:443` | Flink statements and the Flink shell | `confluent flink` commands | `confluent flink endpoint use` | Data plane token |

## Third-party services

| Destination | Used for | When | Notes |
|---|---|---|---|
| `api.stripe.com` | Turning a card into a token | `confluent billing payment update` | Card details are entered at prompts, not as arguments, and go directly to Stripe and are not logged in `--unsafe-trace`; the CLI sends Confluent only the resulting token |
| `github.com/confluentinc/cli-plugins` | Browsing and installing CLI plugins | `confluent plugin install`, `confluent plugin search` | No credentials required |
| `raw.githubusercontent.com/confluentinc/examples` | Downloading client configuration templates | `confluent kafka client-config create` | No credentials required |
| Go module proxy (default `proxy.golang.org`) | Downloading Go plugins (made by the `go` tool) | `confluent plugin install` for Go plugins specifically | Uses the local Go settings (such as `GOPROXY`); requires Go |
| Docker Hub | Pulling the `confluentinc/confluent-local` image (made by Docker) | `confluent local kafka start` | The CLI talks only to the local Docker daemon, which pulls the image |
| Cloud storage (AWS, Azure or GCP) | Uploading artifact and plugin files through a presigned URL from the API, on the cloud selected with `--cloud` | Commands like `confluent flink artifact create` and `confluent connect artifact create` | No CLI credentials; authorized by the presigned URL |

## Notes

- Update checks (`packages.confluent.io`) can be stopped with `confluent configuration update disable_update_check true`.
- Feature flags (`confluent.cloud/ldapi/`) can be stopped with `confluent configuration update disable_feature_flags true`.
- Cluster endpoints can be selected. See the Override column in the Confluent Cloud clusters table, and run `confluent endpoint endpoint list` to see a cluster's endpoints.
