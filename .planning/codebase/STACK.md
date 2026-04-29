# Technology Stack

**Analysis Date:** 2026-03-08

## Languages

**Primary:**
- Go 1.25.7 - Core CLI implementation

**Secondary:**
- Shell/Bash - Build scripts, installer (`install.sh`, `test-installer.sh`)

## Runtime

**Environment:**
- Go 1.25.7 (specified in `.go-version`)
- CGO_ENABLED=1 (required for native dependencies)
- GOEXPERIMENT=boringcrypto (FIPS compliance support)

**Package Manager:**
- Go modules (`go.mod`)
- Lockfile: `go.sum` present

## Frameworks

**Core:**
- Cobra v1.8.1 - CLI command framework
- Sling v1.4.2 - HTTP client library for API interactions
- Viper (via Cobra) - Configuration management

**Testing:**
- Go testing package - Unit tests
- Testify v1.10.0 - Assertion library
- go-mock v0.4.0 - Mocking framework
- Cupaloy v2.8.0 - Snapshot testing
- Gotestsum v1.13.0 - Test runner with JUnit reporting (CI only)

**Build/Dev:**
- GoReleaser v2.13.3 - Cross-platform binary building
- golangci-lint v1.64.8 - Linting
- pre-commit - Git hooks for code quality

## Key Dependencies

**Critical:**
- github.com/confluentinc/ccloud-sdk-go-v2/* (30+ packages) - Confluent Cloud API clients for all services (Kafka, Flink, Connect, Schema Registry, IAM, networking, etc.)
- github.com/confluentinc/ccloud-sdk-go-v1-public - Legacy Confluent Cloud v1 API
- github.com/confluentinc/mds-sdk-go-public/* - Confluent Platform MDS (RBAC) clients
- github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3 v0.3.18 - Kafka REST API client
- github.com/confluentinc/schema-registry-sdk-go - Schema Registry API client
- github.com/confluentinc/confluent-kafka-go/v2 v2.13.0 - Kafka client library

**Infrastructure:**
- golang.org/x/oauth2 v0.27.0 - OAuth2 authentication
- github.com/go-jose/go-jose/v3 v3.0.4 - JWT handling
- github.com/hashicorp/go-retryablehttp v0.7.7 - Retryable HTTP client
- github.com/hashicorp/vault/api v1.15.0 - HashiCorp Vault integration (for secret storage)

**UI/Terminal:**
- github.com/charmbracelet/bubbletea v0.26.6 - TUI framework
- github.com/charmbracelet/bubbles v0.18.0 - TUI components
- github.com/charmbracelet/lipgloss v0.11.0 - Terminal styling
- github.com/charmbracelet/glamour v0.7.0 - Markdown rendering
- github.com/rivo/tview - Terminal UI widgets
- github.com/olekukonko/tablewriter v0.0.5 - ASCII table formatting
- github.com/fatih/color v1.17.0 - Terminal colors

**Platform-specific:**
- github.com/keybase/go-keychain - macOS Keychain integration
- github.com/billgraziano/dpapi v0.5.0 - Windows DPAPI for credential storage
- github.com/panta/machineid v1.0.2 - Machine identification

**Utilities:**
- github.com/google/uuid v1.6.0 - UUID generation
- github.com/samber/lo v1.44.0 - Functional programming helpers
- github.com/iancoleman/strcase v0.3.0 - String case conversion
- github.com/tidwall/gjson v1.17.1 - JSON parsing
- github.com/tidwall/sjson v1.2.5 - JSON modification
- gopkg.in/yaml.v3 v3.0.1 - YAML parsing

**Cloud Provider SDKs:**
- github.com/aws/aws-sdk-go v1.54.15 - AWS SDK (for BYOK/KMS)
- github.com/Azure/azure-sdk-for-go/sdk/azcore v1.11.1 - Azure SDK
- cloud.google.com/go/* - Google Cloud SDK
- github.com/tink-crypto/tink-go-gcpkms/v2 v2.1.0 - Google Cloud KMS
- github.com/tink-crypto/tink-go-hcvault/v2 v2.1.0 - HashiCorp Vault KMS

**Docker:**
- github.com/docker/docker v28.0.0 - Docker client (for local Kafka commands)
- github.com/docker/go-connections v0.5.0 - Docker networking

**Serialization:**
- github.com/linkedin/goavro/v2 v2.13.0 - Avro serialization
- github.com/gogo/protobuf v1.3.2 - Protocol Buffers
- google.golang.org/protobuf v1.34.2 - Protocol Buffers v2
- github.com/bufbuild/protocompile v0.14.1 - Protobuf compiler

**Feature Flags:**
- gopkg.in/launchdarkly/go-sdk-common.v2 v2.5.1 - LaunchDarkly SDK

**Other:**
- github.com/inconshreveable/go-update - Self-update mechanism
- github.com/stripe/stripe-go/v76 v76.25.0 - Stripe API (for billing operations)
- github.com/go-git/go-git/v5 v5.16.5 - Git operations
- github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c - Browser launching (for SSO)

## Configuration

**Environment:**
- Authentication via environment variables (see INTEGRATIONS.md for details)
- Config file: `~/.confluent/config.json` (inferred from codebase)
- Supports both Confluent Cloud and Confluent Platform modes
- CONFLUENT_HOME - On-prem installation directory
- CONFLUENT_CURRENT - On-prem runtime state directory

**Build:**
- `.goreleaser.yml` - Cross-platform build configuration
- `.golangci.yml` - Linter configuration
- `Makefile` - Build automation
- `.pre-commit-config.yaml` - Pre-commit hooks
- Build tags: `musl` for Alpine Linux builds, `live_test` for live testing

**Version Control:**
- `.go-version` - Go version pinning (for goenv)
- Version injection via ldflags at build time

## Platform Requirements

**Development:**
- Go 1.25.7
- goenv recommended for Go version management
- Make
- Pre-commit hooks installed (`pre-commit install`)
- macOS: Increased file descriptor limits for integration tests
- Cross-compilation support for Linux/Windows (requires appropriate toolchains)

**Production:**
- Deployment targets: Linux (amd64/arm64, glibc/musl), macOS (amd64/arm64), Windows (amd64)
- Single static binary with CGO dependencies
- FIPS mode supported via BoringCrypto

---

*Stack analysis: 2026-03-08*
