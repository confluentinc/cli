# Confluent CLI

## Project Overview

The Confluent CLI (`confluent`) is a Go/Cobra command-line tool for operating Confluent Cloud and
Confluent Platform (on-prem). It ships to enterprise customers via Homebrew, APT, YUM, Docker, and
a Windows ZIP, so backward compatibility is a hard constraint (see the compatibility checkpoint
below).

These instructions exist to keep GitHub Copilot PR reviews focused on the invariants that matter in
this codebase. Author-facing guidance (build/test commands, environment setup, the full convention
reference) lives in `AGENTS.md` (symlinked as `CLAUDE.md`), `CONTRIBUTING.md`, and the per-package
`README.md` files under `pkg/` — refer to those rather than repeating them here.

## Architecture (what reviewers need to know)

- **Commands are Cobra command packages under `internal/<command>/`** (e.g. `internal/kafka/`,
  `internal/flink/`, `internal/iam/`). The shape per package is `command.go` (the top-level
  command) + `command_<subcommand>.go` files, with an optional `command_<sub>_onprem.go` sibling
  for on-prem-only variants. Root registration happens in `internal/command.go`. The binary entry
  point is `cmd/confluent/main.go`; other `cmd/` binaries (`docs`, `lint`, ...) are tooling, not
  the CLI itself.
- **Shared behavior lives in `pkg/`** — notably `pkg/cmd/prerunner.go` (auth + SDK client
  construction), `pkg/errors` (error/suggestion formatting), `pkg/output` (output formatting),
  `pkg/deletion` (multi-resource delete), and `pkg/ccloudv2` (the Cloud v2 client).
- **Two runtime modes.** The CLI runs in either a Cloud or an on-prem (Confluent Platform) mode
  depending on login state, and many commands are valid in only one. Mode gating is expressed with
  a Cobra annotation, never a runtime `if` (see checkpoint 3).
- **Machine-readable output (`-o json` / `-o yaml`) is a compatibility surface.** Serialized field
  names and formats are part of the customer contract; changing them is a breaking change.

## Code Review Guidelines (GitHub PR Reviews)

When reviewing pull requests for this project, focus on the checkpoints below. Let `golangci-lint`
(`make lint-go`), the custom `lint-cli` spell/naming checker, and the pre-commit hooks handle
formatting and style — don't comment on anything the tooling already enforces (gofmt, goimports,
import ordering, naked returns, misspellings, etc.).

### 1. Error messages and suggestions

Per `pkg/errors/README.md`:

- Error messages are **lowercase, no trailing period**; the variable suffix is `ErrorMsg`.
- Suggestions are a **capitalized full sentence with a period**; the variable suffix is
  `Suggestions`.
- The two are combined via `errors.NewErrorWithSuggestions(errMsg, suggestions)`. Flag hand-rolled
  error construction that bypasses this.
- Auth/configuration errors should name the missing or invalid setting so users don't have to read
  source to debug.

### 2. Output formatting

Per `pkg/output/README.md`, in any user-facing string (help text, errors, suggestions):

- CLI commands and flags are wrapped in **backticks**: `` `confluent kafka cluster list` ``.
- Resource names and IDs are wrapped in **double quotes**: `"lkc-123456"`.

### 3. Cloud / On-Prem gating — the non-inferable pattern

Commands valid in only one mode must gate via the Cobra annotation, not a runtime branch:

```go
Annotations: map[string]string{annotations.RunRequirement: annotations.RequireCloudLogin}
```

When Cloud and on-prem implementations diverge meaningfully, the convention is a sibling
`command_<sub>_onprem.go` file rather than `if cloudLogin { ... } else { ... }` inside one handler.
Flag PRs that add mode-conditional logic as a runtime branch. See `pkg/cmd/ANNOTATIONS.md`.

### 4. Cobra command wiring

- Argument validation is explicit: `cobra.ExactArgs(N)` for fixed arity, `cobra.MinimumNArgs(1)`
  for variadic commands (especially `delete`). Missing `Args` on a command that takes positional
  arguments is a review blocker.
- Tab completion uses `ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs)`.
- **Multi-resource delete** uses `deletion.ValidateAndConfirm` + `deletion.Delete`
  (`pkg/deletion/README.md`). Do not accept a hand-rolled confirmation prompt.
- **Auth and SDK clients** are built in the shared `PreRunner` during `PersistentPreRunE`. Inside
  `RunE`, they're accessed lazily via `c.V2Client`, `c.MDSClient`, `c.GetKafkaREST()`. Flag any
  handler that authenticates or constructs a client itself.

### 5. Backward compatibility (enterprise binary)

Breaking changes require a major release. Treat these as blockers outside a major bump:

- Removing or renaming a command, subcommand, or flag.
- Changing serialized field names or output format under `-o json` / `-o yaml`.
- Removing a serialized field, or adding `omitempty` to a field that didn't have it.

Safe in minor/patch: adding commands/flags/output fields, renaming human-readable column headers in
default tabular output, hiding a flag. Features marked **EA** (Early Access) or **OP** (Open
Preview) may break across minor versions — those should say so in the command's `Short`/`Long`.

### 6. Testing

- **Every new command and flag needs an integration test.** Integration tests live in `test/`,
  build the CLI against the mock server in `mock/`, and diff stdout/stderr against golden files in
  `test/fixtures/output/<command>/<test>.golden`. A new command/flag with no `test/` coverage is a
  review blocker (global flags like `-v` / `--unsafe-trace` are exempt).
- Golden files are regenerated with `-update` **only** when the diff is the intended user-visible
  change. A golden churn unrelated to the PR's purpose is a red flag.
- Unit tests are co-located with source in `internal/` and `pkg/` (testify, suite-based).

## Files to skip in reviews (generated or infrastructure)

Do not review these for style, patterns, or best practices:

- `mock/**` — regenerated, not hand-edited.
- `pkg/version/**` — stamped at build time.
- `docs/**` — generated from command help text by `cmd/docs` (review the `Short`/`Long`/`Example`
  strings in the command source instead).
- `.semaphore/**`, `.goreleaser.yml`, `debian/**`, `packaging/**`, `docker/Dockerfile*` — release
  and CI infrastructure managed by platform tooling.

## Style preferences (avoid nitpicking)

- `golangci-lint` owns formatting, import ordering (`gci`: stdlib → default → `confluentinc/` →
  `confluentinc/cli/`), naked returns, unused code, and misspellings. Don't comment on any of it.
- `lint-cli` spell-checks user-facing strings. Don't duplicate its findings by hand.
- Don't flag that a change "might fail the build" — Semaphore runs `make lint && make test` and the
  author fixes failures before merge.
- Focus reviews on logic, the checkpoints above, and the compatibility contract.

## Review checklist

Before approving, confirm:

- [ ] Error messages/suggestions follow the casing + `NewErrorWithSuggestions` convention;
      user-facing strings use backticks for commands/flags and double quotes for names/IDs.
- [ ] Mode-specific commands gate via the `annotations.RunRequirement` annotation (or a
      `_onprem.go` sibling), not a runtime branch.
- [ ] Cobra `Args` validation is present and correct; multi-delete uses the `deletion` helpers;
      clients are accessed via `PreRunner`, not constructed in the handler.
- [ ] No breaking change to a command/flag name or to `-o json`/`-o yaml` serialized output outside
      a major release (EA/OP exceptions called out in help text).
- [ ] Every new command and flag has an integration test with golden files; golden diffs are
      intended, not incidental churn.
- [ ] PR description and applicable checklist items in `.github/pull_request_template.md` are filled
      in (release notes, blast radius, tests, breaking-change and feature-flag status).
