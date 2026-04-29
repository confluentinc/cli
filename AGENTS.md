# AGENTS.md

Operating context for AI coding agents (OpenAI Codex, GitHub Copilot, Cursor, Google Jules, Amp, Claude Code, Aider) working in this repository. Plain markdown per the [agents.md](https://agents.md) standard. `CLAUDE.md` is a symlink to this file — there is one source of truth.

This file documents only the non-inferable details. Architecture, file layout, and call graphs are discoverable by reading the code; do not rely on this file as a substitute.

---

## Stack

- **Language:** Go, version pinned in `.go-version` (currently `1.25.7`). Use `goenv` or any tool that respects `.go-version`.
- **Framework:** Cobra (CLI) + Viper (config).
- **Build/test driver:** `make` — wraps `go build`/`go test` with version stamping and the integration-test harness.
- **Distribution:** Homebrew, APT, YUM, Docker, Windows ZIP. Customer-installed binary; backward compatibility is enforced (see "Compatibility rules" below).

## Setup

```bash
brew install pre-commit          # prerequisite — pre-commit hooks scan for secrets via gitleaks
pre-commit install               # one-time, in repo root
ulimit -n 4096                   # macOS only; persist by appending to /etc/profile (see CONTRIBUTING.md)
```

## Build, test, lint

Place these commands in your immediate context — they're the operational core.

```bash
# Build
make build                                                       # → dist/confluent_<os>_<arch>/confluent
GOARCH=arm64 make build                                          # cross-compile, same OS
GOOS=linux GOARCH=amd64 make cross-build                         # cross-OS (requires musl-cross or mingw-w64; see README)
GOLANG_FIPS=1 make build                                         # FIPS-140 build on macOS (full setup in README.md, NOT CONTRIBUTING.md)

# Test
make test                                                        # unit + integration
make unit-test                                                   # unit only — fastest signal
make unit-test UNIT_TEST_ARGS="-run TestApiTestSuite/TestCreateCloudAPIKey"
make integration-test                                            # rebuilds CLI by default
make integration-test INTEGRATION_TEST_ARGS="-run TestCLI/TestKafka"
make integration-test INTEGRATION_TEST_ARGS="-no-rebuild"        # iterate without rebuild
make integration-test INTEGRATION_TEST_ARGS="-update"            # regenerate golden files (see Testing rules)

# Live tests against real Confluent Cloud (require credentials):
make live-test-kafka                                             # also: -schema-registry, -iam, -auth, -connect, -core, -essential

# Lint and coverage
make lint                                                        # = lint-go + lint-cli
make lint-go                                                     # golangci-lint per .golangci.yml
make lint-cli                                                    # hunspell-based spell check on user-facing strings (cmd/lint/main.go)
make coverage                                                    # merge unit + integration coverage → coverage.txt
```

For an inner loop on a single Go test, `go test ./pkg/foo/ -run TestX -v` is fine and faster than the Make target. Use `make` for the full reproducible build, the integration harness, and CI parity.

## Non-inferable conventions

These rules are not enforced by `lint-cli` (which only spell-checks); they're enforced in code review. Breaking them gets the PR sent back.

**Error messages** (`pkg/errors/README.md`):
- Error message: lowercase, no trailing period; variable suffix `ErrorMsg`.
- Suggestion: capitalized, full sentence with period; variable suffix `Suggestions`.
- Combine with `errors.NewErrorWithSuggestions(errMsg, suggestions)`.

**Output formatting** (`pkg/output/README.md`):
- CLI commands and flags use **backticks**: `` `confluent kafka cluster list` ``.
- Resource names and IDs use **double quotes**: `"lkc-123456"`.
- Combined: `` Update cluster "lkc-123456" with `confluent kafka cluster update lkc-123456`. ``

**Cobra argument validation:**
- Fixed N args: `cobra.ExactArgs(N)`.
- Variadic, especially `delete`: `cobra.MinimumNArgs(1)`.
- Tab completion: `ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs)`.

**Multi-resource delete** (`pkg/deletion/README.md`): use `deletion.ValidateAndConfirm` + `deletion.Delete`. Do not roll your own confirmation prompt.

**Cloud / On-Prem mode gating** (`pkg/cmd/ANNOTATIONS.md`) — *the* non-inferable repo pattern: the CLI runs in two distinct modes based on login state, and many commands are valid in only one. Gate via the Cobra annotation, not a runtime branch:

```go
Annotations: map[string]string{annotations.RunRequirement: annotations.RequireCloudLogin}
```

When implementations diverge meaningfully, write a sibling `command_<sub>_onprem.go` rather than `if cloudLogin { ... } else { ... }`.

**Authentication and SDK clients** are initialized in the shared `PreRunner` (`pkg/cmd/prerunner.go`) during `PersistentPreRunE`. Inside `RunE`, access lazily via `c.V2Client`, `c.MDSClient`, `c.GetKafkaREST()` — do not authenticate or construct clients in your handler.

## Testing rules

- **Unit tests** are co-located with source files in `internal/...` and `pkg/...`.
- **Integration tests** live in `test/`. They build the CLI with coverage instrumentation, run it against the mock server in `mock/`, and diff stdout/stderr against golden files in `test/fixtures/output/<command>/<test>.golden`.
- Update goldens with `-update` **only** when the diff is the intended user-visible change. If goldens become stale from an unrelated merge, that's a separate fix.
- For multi-step state-bearing tests, set `workflow: true` on the `CLITest`.
- Add an integration test for every new command and flag — review will require it.

## Compatibility rules

This binary ships to enterprise customers. Avoid breaking changes outside major releases.

**Breaking** (require a major version):
- Removing or renaming a command, subcommand, or flag.
- Changing serialized field names (`-o json` / `-o yaml`).
- Removing serialized fields.
- Changing serialized output format.

**Non-breaking** (safe in minor/patch):
- Renaming human-readable column headers in default tabular output.
- Hiding a flag (still works; just absent from `--help`).
- Adding new commands, flags, or output fields.

Features marked **EA** (Early Access) or **OP** (Open Preview) may break across minor versions — call this out in the command's `Short`/`Long`.

## PR and commit conventions

- Branch off `main`. PR title format: `[<JIRA-TICKET>] <Description>` — the dominant convention in `git log` (e.g., `[CLI-3708]`, `[APIE-890]`, `[MATRIX-1286]`). Plain prefixes (`chore:`, `docs:`) only for non-ticketed work.
- One logical change per PR.
- Run `make lint && make test` before opening; CI runs both via Semaphore.
- Do not bypass pre-commit hooks with `--no-verify`.

## Edit zones

| | What | Why |
|---|---|---|
| ✅ **Edit freely** | `internal/`, `pkg/`, `cmd/confluent/main.go`, tests, `test/fixtures/output/` (with `-update` only when intended) | Standard development surface |
| ⚠️ **Ask first** | `Makefile`, `.golangci.yml`, `.pre-commit-config.yaml`, `pkg/errors/error_message.go`, `internal/command.go` (root command registration) | Cross-cutting; affects every contributor |
| 🚫 **Don't touch without explicit instructions** | `service.yml`, `.semaphore/`, `.goreleaser.yml`, `debian/`, `packaging/`, `docker/Dockerfile*`, `mock/` (regenerated, not hand-edited), `pkg/version/` (stamped at build time), `LICENSE`, `SECURITY.md` | Release/CI infrastructure managed by platform tooling; manual edits get overwritten or break the release pipeline |

## Deeper references

- [`README.md`](README.md) — install, FIPS-140 setup, cross-compile prerequisites.
- [`CONTRIBUTING.md`](CONTRIBUTING.md) — full contributor guide, environment setup, signing.
- `pkg/cmd/ANNOTATIONS.md` — Cloud/On-Prem command annotations.
- `pkg/cmd/AUTOCOMPLETION.md` — tab-completion architecture.
- `pkg/errors/README.md` — error-message rules.
- `pkg/output/README.md` — output-formatting rules.
- `pkg/deletion/README.md` — multi-delete pattern.

---

*This file is under the 32 KiB cap recommended by OpenAI Codex (`project_doc_max_bytes`). When new context is needed, prefer adding a nested `AGENTS.md` in the relevant subdirectory rather than expanding this file.*
