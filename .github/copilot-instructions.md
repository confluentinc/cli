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
  a Cobra annotation, never a runtime `if` (see checkpoint 5).
- **Machine-readable output (`-o json` / `-o yaml`) is a compatibility surface.** Serialized field
  names and formats are part of the customer contract; changing them is a breaking change. Output
  structs are also the highest-traffic review surface (see checkpoint 1).
- **Much of the CMF/CP surface is generated.** Command code, goldens, and live tests for newer
  areas are emitted by `cli-terraform-generator`; reviewers push to keep hand-written code minimal
  and spec-aligned so it can be generated (see checkpoint 7).

## Code Review Guidelines (GitHub PR Reviews)

When reviewing pull requests for this project, focus on the checkpoints below. Let `golangci-lint`
(`make lint-go`), the custom `lint-cli` spell/naming checker, and the pre-commit hooks handle
formatting and style — don't comment on anything the tooling already enforces (gofmt, goimports,
import ordering, naked returns, misspellings, etc.).

The checkpoints are ordered by how often they actually decide a review in this repo. Output-shape,
validation, and compatibility questions dominate; argument arity and error-message casing are
near-noise (convention and `lint-cli` own them). Weight comments accordingly.

### 1. Output and serialization design (the highest-traffic review surface)

An `*Out` struct and its `human:` / `serialized:` (or `json:`/`yaml:`) tags **are** the command's
user contract, and reviewers scrutinize their design more than anything else. Per
`pkg/output/README.md` and observed convention:

- Serialized field names should **mirror the API/spec**, not an invented CLI name — matching the
  spec is what allows these commands to be generated later. Flag a divergent name.
- `omitempty` belongs on **optional** response fields only. A field the API always returns should
  not carry it; an optional one should. (Note the compatibility caveat in checkpoint 3: _adding_
  `omitempty` to a field that already ships without it is a breaking change.)
- Use `human:"-"` to hide a field from the human table while keeping it in `-o json`/`-o yaml`,
  rather than dropping the field from the struct.
- Field sets should be **consistent across human/json/yaml**, or an intentional divergence should be
  explained in the PR — "a different set of fields per format" is a legitimate thing to question.
- New table columns should be chosen deliberately (narrow is easy to widen later; removing a column
  is breaking). Long/unbounded values may warrant `human:"-"` over a wrapped column.

```go
// Avoid: invented serialized name; omitempty on a field the API always returns;
// a long/unbounded value forced into the human table.
type licenseOut struct {
    Type      string `human:"Type" serialized:"type"`                 // spec calls it "license_type"
    ExpiresAt string `human:"Expires At,omitempty" serialized:"expires_at,omitempty"` // always returned
    RawJwt    string `human:"Raw JWT" serialized:"raw_jwt"`           // long; clutters the table
}

// Prefer: serialized names match the spec; omitempty only on optional fields;
// keep the long value in json/yaml but out of the human table with human:"-".
type licenseOut struct {
    LicenseType string `human:"License Type" serialized:"license_type"`
    ExpiresAt   string `human:"Expires At" serialized:"expires_at"`
    RawJwt      string `human:"-" serialized:"raw_jwt"`
}
```

### 2. Prefer backend validation over client-side checks

The house rule is to **let the backend validate** and surface its error, not duplicate spec
constraints in the CLI:

- Question new hand-rolled validation of a spec-recorded constraint (allowed enum values,
  required-ness, formats) — the backend already enforces it and returns a field-named error.
- Flag **relationships** should use Cobra built-ins — `cmd.MarkFlagsMutuallyExclusive(...)` for
  `oneOf` groups, `cmd.MarkFlagsRequiredTogether(...)` for co-required flags — not a bespoke
  `if flagA != "" && flagB != "" { ... }` ladder.
- Case normalization (`strings.ToUpper` on an enum) should be a conscious choice; for
  `x-extensible-enum` fields a client-side allow-list breaks forward-compat.
- Do **not** branch on backend error wording — string-matching a server message is brittle and is
  usually a sign validation should have stayed server-side.

### 3. Backward compatibility (enterprise binary)

Breaking changes require a major release. Treat these as blockers outside a major bump:

- Removing or renaming a command, subcommand, or flag.
- Changing serialized field names or output format under `-o json` / `-o yaml`.
- Removing a serialized field, or adding `omitempty` to a field that didn't have it.
- Printing a status line (e.g. `Updated "x".`) to **stdout** on a command that also emits
  `-o json`/`-o yaml` — it breaks `| jq`. Status messages belong on stderr, or are omitted for
  resources that print a payload.

Safe in minor/patch: adding commands/flags/output fields, **renaming a human-readable (`human:`)
column header**, hiding a flag. Renaming a `serialized:` key is breaking; renaming its `human:`
label is not — keep that distinction straight. Features marked **EA** (Early Access) or **OP** (Open
Preview) may break across minor versions — those should say so in the command's `Short`/`Long`.

### 4. Error-handling semantics (not casing)

The casing rules (lowercase error, no trailing period, `ErrorMsg`/`Suggestions` suffixes, combined
via `errors.NewErrorWithSuggestions`, per `pkg/errors/README.md`) are real but owned by convention
and `lint-cli` — don't spend review budget on them. Spend it on what tooling can't see:

- An existence/precheck helper must not collapse **every** failure (401, 500, timeout,
  connection-refused) into "not found" — only a true 404 should map to "not found".
- Status-code gating needs the right boundary: `< 400` treats a 3xx redirect as success and
  unmarshals into a zero-value struct (printing a blank "success"). Prefer an explicit 2xx check,
  noting `201` on create.
- Errors must not be silently swallowed — the underlying cause should reach the user rather than a
  generic substitute.
- Auth/configuration errors should name the missing or invalid setting.
- User-facing strings wrap commands/flags in **backticks** and names/IDs in **double quotes**
  (`pkg/output/README.md`).

### 5. Cobra wiring and Cloud/On-Prem gating

- A subcommand must be registered on its parent (new top-level commands in `internal/command.go`);
  an unregistered command file silently does not exist.
- `Args` should match arity (`cobra.ExactArgs(N)`, or `cobra.MinimumNArgs(1)` for variadic/`delete`),
  with `ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs)` where completion applies. Worth a
  glance, but low-frequency — don't over-index on it.
- Mode-specific commands gate via the `annotations.RunRequirement` annotation
  (`RequireCloudLogin` / `RequireOnPremLogin` / etc.) with the **correct** requirement; meaningfully
  divergent Cloud vs. on-prem behavior belongs in a `command_<sub>_onprem.go` sibling, not an
  `if cloudLogin { ... } else { ... }` branch. See `pkg/cmd/ANNOTATIONS.md`.
- **Multi-resource delete** uses `deletion.ValidateAndConfirm` + `deletion.Delete`
  (`pkg/deletion/README.md`) — not a hand-rolled confirmation prompt.
- **Auth and SDK clients** are built in the shared `PreRunner`; inside `RunE` they're accessed
  lazily via `c.V2Client`, `c.MDSClient`, `c.GetKafkaREST()`. Flag any handler that authenticates or
  constructs a client itself.

### 6. Testing — cover failure paths, not just "a golden exists"

- **Every new command and flag needs an integration test.** Integration tests live in `test/`,
  build the CLI against the mock server in `mock/`, and diff stdout/stderr against golden files in
  `test/fixtures/output/<command>/<test>.golden`. A new command/flag with no `test/` coverage is a
  review blocker (global flags like `-v` / `--unsafe-trace` are exempt).
- **Each new command needs a failure-path case**, not just the happy path — e.g. not-found and
  missing-required-flag. A green happy-path golden hides an untested error branch.
- **Mock handlers should assert the request they received** (forwarded query params, method, body),
  not just return a canned response — otherwise a wrong-value-sent-to-server bug (including
  data-loss paths like `delete --version`) still passes its golden.
- Golden files are regenerated with `-update` **only** when the diff is the intended user-visible
  change. Golden churn unrelated to the PR's purpose is a red flag.
- Unit tests are co-located with source in `internal/` and `pkg/` (testify, suite-based); new
  frequently-reused helpers deserve one.

### 7. Generator alignment and internal-reference hygiene

- Much of the CMF/CP command surface is emitted by `cli-terraform-generator`. Flag hand-written Go
  that only re-derives a value already present in the OpenAPI spec — e.g. a hardcoded pending/terminal
  phase set that duplicates the status enum, or a default hardcoded in the command that the spec
  already defines. Flag a one-off helper (map-field extraction, output conversion) that duplicates a
  sibling instead of being generalized into `pkg/flink` (or similar). (Deeper "should this whole thing
  be generated?" judgment is for human reviewers, not a per-diff check.)
- No internal identifiers (JIRA/APIE tickets, RFC links, internal service names like `cc-api`)
  should leak into **user-facing** strings — `Short`/`Long`/`Example`, errors, suggestions. They may
  stay in engineer-facing doc comments. Also flag PR-introduced comments that carry a stale ticket
  reference or sit above the wrong function.

### 8. PR description

- Flag if the PR description leaves the applicable sections of `.github/pull_request_template.md`
  unfilled — release notes, blast radius, tests, and breaking-change / feature-flag status. A blank
  template on a non-trivial change is worth a comment.

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
