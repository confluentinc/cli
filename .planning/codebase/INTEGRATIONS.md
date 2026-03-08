# External Integrations

**Analysis Date:** 2026-03-08

## APIs & External Services

**Confluent Cloud API:**
- Service: Confluent Cloud control plane and data plane APIs
  - SDK/Client: github.com/confluentinc/ccloud-sdk-go-v2/* (30+ service-specific packages)
  - Auth: OAuth2 with API keys, email/password, SSO
  - Base URL: https://confluent.cloud
  - Services: Kafka clusters (CMK), Flink, ksqlDB, Connect, Schema Registry, Stream Designer, networking, IAM, billing, organizations, service quotas, BYOK, certificate authority

**Confluent Platform (On-Prem):**
- Service: Confluent Platform MDS (Metadata Service) for RBAC
  - SDK/Client: github.com/confluentinc/mds-sdk-go-public/mdsv1, mdsv2alpha1
  - Auth: Username/password, SSO, mTLS certificates
  - Env var: CONFLUENT_PLATFORM_MDS_URL
  - Certificate support: CA cert path, client cert/key paths

**Kafka REST API:**
- Service: Kafka REST Proxy
  - SDK/Client: github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3
  - Auth: API keys, Basic auth
  - Used for: Kafka cluster operations, topic management, consumer groups

**Schema Registry API:**
- Service: Confluent Schema Registry
  - SDK/Client: github.com/confluentinc/schema-registry-sdk-go
  - Auth: API keys
  - Used for: Schema management, subject operations, compatibility checks

**LaunchDarkly:**
- Service: Feature flag management
  - SDK/Client: gopkg.in/launchdarkly/go-sdk-common.v2
  - Client IDs: Separate for prod/test/stag/devel environments
  - Used for: Feature flags, announcements, deprecation notices
  - Base URL pattern: ldapi/sdk/eval endpoint

**Stripe:**
- Service: Payment processing
  - SDK/Client: github.com/stripe/stripe-go/v76
  - Used for: Billing operations and payment management

**Docker:**
- Service: Local Docker daemon
  - SDK/Client: github.com/docker/docker v28
  - Used for: Local Kafka cluster management (`confluent local kafka` commands)
  - Requires: Docker daemon running locally

## Data Storage

**Databases:**
- None - CLI is stateless except for local config

**File Storage:**
- Local filesystem only
  - Config: `~/.confluent/config.json`
  - Credentials: macOS Keychain (darwin), Windows DPAPI (windows), or encrypted file (other)
  - Logs: Local log files (location configurable)
  - Test fixtures: `test/fixtures/` directory

**Caching:**
- In-memory caching of API clients during command execution
- No persistent cache

## Authentication & Identity

**Auth Providers:**
- Confluent Cloud
  - Email/password authentication
  - API key authentication (cloud and resource-scoped)
  - OAuth2 SSO (browser-based flow)
  - MFA support via `pkg/auth/mfa`
  - Environment variables: CONFLUENT_CLOUD_EMAIL, CONFLUENT_CLOUD_PASSWORD, CONFLUENT_CLOUD_ORGANIZATION_ID

- Confluent Platform
  - Username/password authentication
  - LDAP/AD integration (via MDS)
  - SSO (browser-based OAuth flow)
  - mTLS certificate authentication
  - Environment variables: CONFLUENT_PLATFORM_USERNAME, CONFLUENT_PLATFORM_PASSWORD, CONFLUENT_PLATFORM_MDS_URL, CONFLUENT_PLATFORM_SSO

**Credential Storage:**
- macOS: Keychain via github.com/keybase/go-keychain
- Windows: DPAPI via github.com/billgraziano/dpapi
- Linux: Encrypted file with master key in CONFLUENT_SECURITY_MASTER_KEY env var
- JWT token handling via github.com/go-jose/go-jose/v3

**SSO Implementation:**
- Local HTTP server for OAuth callback (`pkg/auth/sso/auth_server.go`)
- Browser launch via github.com/pkg/browser
- State management for CSRF protection
- Both Cloud and Platform SSO supported

## Monitoring & Observability

**Error Tracking:**
- None - Errors logged locally only

**Logs:**
- Local logging via `pkg/log` package
- Structured logging with log levels
- No external log aggregation

**Telemetry:**
- LaunchDarkly for feature flag evaluation (includes user context)
- No explicit analytics or crash reporting

## CI/CD & Deployment

**Hosting:**
- Distributed as standalone binary
- Installation methods:
  - Direct download (install.sh script)
  - Package managers (deb, rpm, brew - inferred from packaging/ directory)
  - Docker images (docker/ directory)

**CI Pipeline:**
- Semaphore CI (.semaphore/semaphore.yml)
  - Linux amd64 builds and tests
  - Linux arm64 builds
  - macOS builds (separate workflow)
  - Coverage reporting to SonarQube
- GitHub Actions (.github/workflows/)
  - Inactive issues management
  - Live tests workflow

**Build Process:**
- GoReleaser for cross-platform builds
- Cross-compilation from macOS to Linux/Windows
- Alpine Linux support (musl builds)
- FIPS builds via BoringCrypto

**Secrets Management:**
- Vault integration via vault-sem-get-secret (CI only)
- SonarQube token stored in Vault

## Environment Configuration

**Required env vars:**

Cloud login:
- CONFLUENT_CLOUD_EMAIL
- CONFLUENT_CLOUD_PASSWORD
- CONFLUENT_CLOUD_ORGANIZATION_ID (optional)

Platform login:
- CONFLUENT_PLATFORM_USERNAME
- CONFLUENT_PLATFORM_PASSWORD
- CONFLUENT_PLATFORM_MDS_URL

Certificates (optional):
- CONFLUENT_PLATFORM_CERTIFICATE_AUTHORITY_PATH
- CONFLUENT_PLATFORM_CLIENT_CERT_PATH
- CONFLUENT_PLATFORM_CLIENT_KEY_PATH

CMF (Confluent Metadata Framework):
- CONFLUENT_CMF_URL
- CONFLUENT_CMF_CLIENT_KEY_PATH
- CONFLUENT_CMF_CLIENT_CERT_PATH
- CONFLUENT_CMF_CERTIFICATE_AUTHORITY_PATH

Other:
- CONFLUENT_HOME - On-prem installation directory
- CONFLUENT_CURRENT - On-prem runtime state
- CONFLUENT_SECURITY_MASTER_KEY - Master key for credential encryption
- CLI_LIVE_TEST - Enable live tests
- CLI_LIVE_TEST_GROUPS - Live test group selection

**Secrets location:**
- macOS: Keychain (`pkg/keychain/keychain_darwin.go`)
- Windows: DPAPI encrypted credentials
- Linux: File-based with encryption
- Environment variables (not recommended for production)

## Webhooks & Callbacks

**Incoming:**
- OAuth callback endpoint: `http://localhost:<random-port>/callback`
  - Used for SSO flows (both Cloud and Platform)
  - Ephemeral server started during authentication
  - Automatic port selection via github.com/phayes/freeport

**Outgoing:**
- None - CLI does not send webhooks

## Cloud Provider Integrations

**AWS:**
- Service: Key Management Service (KMS)
  - SDK: github.com/aws/aws-sdk-go
  - Used for: BYOK (Bring Your Own Key) operations
  - Integration: `internal/byok/` package

**Azure:**
- Service: Key Vault
  - SDK: github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azkeys
  - Used for: BYOK operations
  - Requires: Azure credentials and key vault access

**Google Cloud Platform:**
- Service: Cloud KMS
  - SDK: github.com/tink-crypto/tink-go-gcpkms/v2
  - Used for: BYOK operations
  - Requires: GCP credentials and KMS access

**HashiCorp Vault:**
- Service: Vault KMS
  - SDK: github.com/hashicorp/vault/api, github.com/tink-crypto/tink-go-hcvault/v2
  - Used for: Secret storage and BYOK
  - AppRole authentication supported

## Protocol Support

**REST APIs:**
- Primary communication method
- JSON payloads
- OAuth2 Bearer token authentication
- Retry logic via github.com/hashicorp/go-retryablehttp

**gRPC:**
- Not directly used (SDKs may use internally)

**WebSocket:**
- github.com/gorilla/websocket present (likely for streaming APIs)

**Protobuf:**
- Serialization/deserialization support via `pkg/serdes/`
- Used for Kafka message handling

**Avro:**
- Schema serialization via github.com/linkedin/goavro/v2
- Used for Kafka message handling with Schema Registry

---

*Integration audit: 2026-03-08*
