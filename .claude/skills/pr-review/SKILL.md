---
name: pr-review
description:
  Reviews pull requests for the Confluent CLI. Use when reviewing PRs, doing self-review before
  sharing with the team, or when the user mentions "review PR", "help with PR", "review changes",
  "self-review", "review local changes", or "check my PR". Focuses on output/serialization design,
  backend-first validation, error-handling semantics, backward compatibility, generator alignment,
  and failure-path integration tests.
allowed-tools:
  - Read
  - Bash
  - Grep
  - Glob
  - Agent
---

# PR Review Skill

Reviews pull requests for the Confluent CLI (Go + Cobra), focusing on project-specific patterns and
the failure modes most likely to slip past `golangci-lint` and `lint-cli`.

The authoritative conventions live in `AGENTS.md` (symlinked as `CLAUDE.md`) and the per-package
`README.md` files under `pkg/` (`pkg/errors/README.md`, `pkg/output/README.md`,
`pkg/deletion/README.md`, `pkg/cmd/ANNOTATIONS.md`). This skill applies them to a diff; when a
convention is ambiguous, read the source of truth rather than re-deriving it from the diff.

The same review conventions are mirrored, in review-framed form, as path-scoped rules in
`.claude/rules/` (e.g. `output-and-compatibility.md`, `validation-and-errors.md`). Those are what the
automated R2 code review consumes, since it runs headless and read-only in CI and cannot invoke this
skill (which relies on `gh` and subagents). Keep the two in sync when a convention changes.

## Two Review Modes

The mode is selected by the invocation context, not by the user. If the user supplies a PR
number/URL or asks about someone else's PR, run **Formal Review Mode**. Otherwise (no PR number,
working from a local branch, phrases like "self-review" or "check my PR") run **Self-Review Mode**.
When ambiguous, ask which mode to use.

### Self-Review Mode (for PR authors)

Use when: the author wants to check their own changes before sharing with the team. Typically on a
draft PR or against local changes before pushing.

Goals:

- Catch issues early, before formal review
- Check that `*Out` structs match the API/spec and behave consistently across human/json/yaml
- Confirm new validation is deferred to the backend rather than hand-rolled client-side
- Confirm every new command/flag has integration tests covering the failure paths, not just success
- Catch backward-compatibility breaks (serialized-field and stdout/stderr changes) before a reviewer does
- Verify Cobra command wiring is complete (registration, `Args`, annotations, completion)

### Formal Review Mode (for reviewers)

Use when: a reviewer needs to evaluate a PR from another team member.

Goals:

- Quickly understand the scope and purpose of changes
- Identify potential issues or concerns
- Provide constructive feedback grounded in CLI conventions
- Verify the PR template checklist is honored

## Review Process

### Step 1: Gather Information

**For local changes (self-review):**

```bash
# files changed since divergence from main
git diff main --name-only

# overview and full diff
git diff main --stat
git diff main
```

**For GitHub PRs:**

```bash
# PR metadata
gh pr view <PR_NUMBER> --json number,title,body,author,baseRefName,headRefName,additions,deletions,changedFiles,state,reviewDecision

# if no PR number is given, try the current branch
gh pr view --json number,title,body,author,baseRefName,headRefName,additions,deletions,changedFiles,state,reviewDecision

# diff
gh pr diff <PR_NUMBER>

# existing reviews and inline comments
gh pr view <PR_NUMBER> --json reviews,comments

# referenced issues
gh issue view <ISSUE_NUMBER> --json body,comments
```

### Step 2: Filter Files for Review

**SKIP these paths entirely** (generated or infrastructure):

- `mock/**` — regenerated, not hand-edited
- `pkg/version/**` — stamped at build time
- `docs/**` — generated from command help text (review the `Short`/`Long`/`Example` source instead)
- `dist/**`, `vendor/**`
- `.semaphore/**`, `.goreleaser.yml`, `debian/**`, `packaging/**`, `docker/Dockerfile*` — release/CI

**DO review carefully** (small file, big blast radius):

- `internal/command.go` (root command registration)
- `go.mod` / `go.sum` (new dependencies)
- `.golangci.yml`, `Makefile`, `.pre-commit-config.yaml` (cross-cutting; changes here should be
  intentional and called out)
- `test/fixtures/output/**/*.golden` (the user-visible output contract)

### Step 3: Categorize the Changes

| Category             | File patterns                                       | What to check                                                        |
| -------------------- | --------------------------------------------------- | -------------------------------------------------------------------- |
| Command definitions  | `internal/<command>/command*.go`                    | Registration, `Args` validation, annotations, `ValidArgsFunction`    |
| On-prem variants     | `internal/<command>/command_*_onprem.go`            | Mode divergence handled by sibling file, not runtime `if`            |
| Shared command glue  | `pkg/cmd/**`, `internal/command.go`                 | PreRunner client access; root registration                           |
| Error handling       | `pkg/errors/**`, any `*ErrorMsg` / `*Suggestions`   | Casing rules, `NewErrorWithSuggestions`                              |
| Output               | `pkg/output/**`, printers in command files          | Backticks for commands/flags, quotes for IDs; serialized-field churn |
| Deletion             | anything calling delete/confirm                     | Uses `deletion.ValidateAndConfirm` + `deletion.Delete`               |
| Unit tests           | `internal/**/*_test.go`, `pkg/**/*_test.go`         | Co-located, testify suite style                                      |
| Integration tests    | `test/**`, `test/fixtures/output/**`                | New command/flag covered; golden diffs intended                      |
| Docs / help text     | `Short` / `Long` / `Example` strings, root markdown | Accuracy; EA/OP callouts; run the `docs-drift` skill                 |
| Project rules/skills | `.claude/rules/**`, `.claude/skills/**`             | Frontmatter, path globs, trigger phrases                             |

### Step 4: Check Critical Requirements

**IMPORTANT: only review lines that were actually changed in the PR diff.** Context lines from the
diff are for understanding, not for review. Do not flag pre-existing issues in unchanged code.

The checkpoints below are ordered by how often they actually decide a review here — output-shape
and validation questions dominate; arity and error casing are near-noise (the linter owns them).
Weight your attention accordingly.

#### 1. Output and serialization design (the highest-traffic review surface)

The `*Out` struct and its `human:` / `serialized:` (or `json:`/`yaml:`) tags **are** the command's
user contract. Reviewers scrutinize their design far more than any other single thing:

- [ ] Serialized field names **mirror the API/spec**, not an invented CLI name. Matching the spec is
      what lets these commands be generated later; a divergent name is a standing question.
- [ ] `omitempty` is on **optional** response fields only. A field the API always returns should not
      carry it; an optional one should (so it drops out cleanly when absent).
- [ ] `human:"-"` is the right tool to hide a field from the human table while keeping it in
      `-o json`/`-o yaml` — prefer it over dropping the field from the struct entirely.
- [ ] The field set is **consistent across human/json/yaml**, or an intentional divergence is
      explained. "Is it OK for customers to see a different set of fields in each format?" is a real
      review question here, not a nitpick.
- [ ] New table columns are chosen deliberately (start narrow; adding later is easy, removing is a
      breaking change). Long/unbounded values may warrant `human:"-"` rather than a wrapped column.

#### 2. Prefer backend validation over client-side checks

The house rule (stated repeatedly by senior reviewers) is to **let the backend validate** and
surface its error, rather than duplicating spec constraints in the CLI:

- [ ] New hand-rolled validation of a spec-recorded constraint (allowed enum values, required-ness,
      formats) is questioned — the backend already enforces it and returns a field-named error.
- [ ] Flag **relationships** use Cobra's built-ins — `cmd.MarkFlagsMutuallyExclusive(...)` for
      `oneOf` groups, `cmd.MarkFlagsRequiredTogether(...)` for co-required flags — not a bespoke
      `if flagA != "" && flagB != "" { return err }` ladder.
- [ ] Case normalization (`strings.ToUpper` on an enum before sending) is a conscious choice, not a
      reflex; for `x-extensible-enum` fields a client-side allow-list breaks forward-compat.

Red flag: string-matching against backend error wording to branch behavior — brittle the moment the
backend rewords, and usually a sign validation should have stayed server-side.

#### 3. Backward compatibility

- [ ] No renamed/removed command, subcommand, or flag outside a major release
- [ ] No changed/removed serialized field name under `-o json` / `-o yaml`; no new `omitempty` on a
      previously-always-emitted field; no output-format change
- [ ] Renaming a `human:` column header is **not** breaking; changing a `serialized:` key **is**.
      Keep that distinction straight before flagging.
- [ ] Printing a status line (e.g. `Updated "x".`) to **stdout** on a command that also has
      `-o json`/`-o yaml` output breaks `| jq` — status messages belong on stderr, or are omitted
      for resources that print a payload.
- [ ] EA/OP instability is disclosed in the command's `Short`/`Long`

If a `*.golden` file for a `json`/`yaml` test changed, confirm the field-name/shape delta is
intentional and non-breaking (or gated on a major).

#### 4. Error-handling semantics (not casing)

Casing rules (lowercase, no trailing period, `...ErrorMsg`/`...Suggestions` suffixes,
`NewErrorWithSuggestions`) are real but **owned by convention + `lint-cli`** — mention them only in
passing. Spend the review on what the linter can't see:

- [ ] An `existenceFunc` / precheck doesn't collapse **every** failure (401, 500, timeout,
      connection-refused) into "not found" — only a true 404 should map to "not found"; other errors
      must preserve their cause.
- [ ] Status-code gating uses the right boundary. `< 400` treats a 3xx redirect as success and
      unmarshals into a zero-value struct (printing a blank "success"); prefer an explicit 2xx check
      (`>= 200 && < 300`), noting `201` on create.
- [ ] Errors aren't silently swallowed — the underlying cause reaches the user rather than a generic
      substitute message. (For a deeper pass, dispatch the user-global `silent-failure-hunter` agent.)
- [ ] Auth/config errors name the missing or invalid setting.
- [ ] User-facing strings use backticks for commands/flags and double quotes for names/IDs.

#### 5. Cobra wiring and client access

- [ ] Command is registered (subcommand added to its parent; new top-level command wired in
      `internal/command.go`) — an unregistered subcommand file silently does not exist.
- [ ] `Args` is set and matches arity (`cobra.ExactArgs(N)`, or `cobra.MinimumNArgs(1)` for
      variadic/`delete`); `ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs)` where completion
      applies. (Low-frequency in practice — check, don't belabor.)
- [ ] Mode gating uses the `annotations.RunRequirement` annotation with the **correct** requirement
      (`RequireCloudLogin` / `RequireOnPremLogin` / etc.); meaningfully divergent Cloud vs. on-prem
      behavior lives in a `command_<sub>_onprem.go` sibling, not an `if cfg.IsCloudLogin()` branch.
- [ ] Handlers access clients lazily via `c.V2Client` / `c.MDSClient` / `c.GetKafkaREST()` — they do
      not authenticate or construct clients themselves (that belongs in `PreRunner`).

#### 6. Testing — cover failure paths, not just "a golden exists"

- [ ] Every new command and flag has an integration test in `test/` with matching golden files.
- [ ] **Each new command has a failure-path case**, not only the happy path — e.g. not-found and
      missing-required-flag cases. A green happy-path golden hides an untested error branch.
- [ ] Mock handlers **assert the request they received** (forwarded query params, method, body), not
      just return a canned response — otherwise a wrong-value-sent-to-server bug (including data-loss
      paths like `delete --version`) still passes its golden.
- [ ] Golden diffs reflect the intended user-visible change only (no incidental churn from an
      unrelated merge).
- [ ] New non-trivial or frequently-reused logic in `internal/`/`pkg/` has a co-located unit test.

### Step 5: Check Project-Specific Patterns

#### Generator alignment (much of this codebase is generated)

Large parts of the CMF/CP command surface are emitted by `cli-terraform-generator`, and the standing
initiative is to hand-write **as little as possible**. When a PR adds custom Go:

- [ ] Ask whether the logic could be **derived from the OpenAPI spec** or **generated** instead of
      hand-maintained (e.g. pending/terminal phase sets from the status enum, defaults from the spec).
- [ ] A one-off helper that will recur (map-field extraction, output conversion) is worth
      generalizing into `pkg/flink` (or similar) so the generator has a single target — reviewers
      actively push for fewer, more generic custom functions.
- [ ] CLI behavior stays unified with the Terraform provider where both consume the same API
      (defaults, timeouts, naming) — flag divergence for discussion rather than silently shipping it.

#### Multi-resource delete

- [ ] Delete commands accept variadic args (`cobra.MinimumNArgs(1)`) and route confirmation +
      deletion through `deletion.ValidateAndConfirm` and `deletion.Delete` rather than a hand-rolled
      prompt/loop

#### Internal-reference hygiene

- [ ] No internal identifiers (JIRA/APIE tickets, RFC links, internal service names like `cc-api`)
      leak into **user-facing** strings — `Short`/`Long`/`Example`, error messages, suggestions. Keep
      them in engineer-facing doc comments if they're useful there.
- [ ] Comments introduced by the PR don't carry stale ticket references or sit above the wrong
      function.

#### Docs drift

- [ ] If `Short`/`Long`/`Example` text or root markdown changed, the examples still reference real
      commands and flags. Invoke the `docs-drift` skill to validate in-code help text and root
      markdown against the live command tree.

### Step 6: Check PR Hygiene

- [ ] PR title is `[<JIRA-TICKET>] <Description>` (plain `chore:`/`docs:` prefixes only for
      non-ticketed work)
- [ ] PR description fills in the template: release notes, What / Blast Radius, tests, breaking-change
      status, feature-flag enablement
- [ ] One logical change per PR; no unrelated changes bundled in
- [ ] No secrets in the diff (`.env`, API keys, real cluster IDs beyond fixtures)

## What NOT to Flag

### Tooling-owned style (avoid nitpicking)

- Formatting, import ordering, naked returns, unused code, misspellings — owned by `golangci-lint`
  and `lint-cli`. Don't hand-flag any of it.
- "This might fail CI" — Semaphore runs `make lint && make test`; the author fixes failures.

### Comment Preservation

- Never suggest deleting existing comments unless they are now actively misleading.
- Comments explain "why", not "what"; they may look redundant but carry context the reviewer lacks.
- When code changes, preserve or update comments rather than removing them.

## Output Format

### For Self-Review

```markdown
## Self-Review Summary

### Changes Overview

[Brief summary of what changed]

### Critical Requirements Checklist

- [ ] Output/serialization design (spec-matched names, `omitempty` on optionals, cross-format field set): [status]
- [ ] Validation deferred to backend; flag relationships via `MarkFlags*`: [status]
- [ ] Backward compatibility (commands/flags, `-o json`/`yaml`, stdout vs stderr): [status]
- [ ] Error-handling semantics (no collapsed/swallowed errors, right status boundary): [status]
- [ ] Cobra wiring + client access via PreRunner: [status, location of any gap]
- [ ] Tests cover failure paths + mock request assertions, not just a happy-path golden: [status]

### Issues to Address Before PR

1. [High-priority issue with file:line]
2. [Medium-priority issue with file:line]

### Suggestions (Optional)

- [Nice-to-have improvements]

### Ready for Review?

[Yes / Not yet, with reasoning]
```

### For Formal Review

```markdown
## PR Review: #{number} - {title}

**Author:** {author}
**Branch:** {headRefName} → {baseRefName}
**Changes:** +{additions} / -{deletions} across {changedFiles} files

### Summary

[2-3 sentence summary of what the PR does and why]

### Changed Components

- [Categorized list of changed files, excluding generated/infra]

### Findings

#### Issues (Must Fix)

- [ ] **[category]**: [description] - `file:line`

#### Suggestions (Consider)

- [ ] **[category]**: [description] - `file:line`

#### Positive Observations

- [Good patterns, thorough tests, well-written code]

### Test Coverage Assessment

- **New tests added:** [Yes/No, list test files + goldens]
- **Coverage gaps:** [Untested commands/flags or edge cases]

### Compatibility Notes

[Any concern about command/flag renames or serialized-output changes]

### Recommendation

**[APPROVE / REQUEST CHANGES / NEEDS DISCUSSION]**

[Brief rationale]
```

## Review Categories

Use these labels in findings:

| Category         | Description                                                                                                  |
| ---------------- | ------------------------------------------------------------------------------------------------------------ |
| `output-design`  | Serialized name not spec-matched; wrong `omitempty`; human/json/yaml field-set drift; column choice          |
| `validation`     | Client-side check that belongs on the backend; flag relationship not via `MarkFlags*`                        |
| `compatibility`  | Breaking command/flag rename, `-o json`/`-o yaml` change, or status line on stdout outside major             |
| `error-handling` | Collapsed/swallowed error, wrong status-code boundary, brittle error-string match                            |
| `generator`      | Hand-written code that could be spec-derived/generated; one-off helper worth generalizing; CLI/TF divergence |
| `testing`        | Missing failure-path case or mock request assertion; missing golden; incidental churn                        |
| `cobra-wiring`   | Missing registration, `Args`, annotation gating, or completion hookup                                        |
| `mode-gating`    | Wrong `RunRequirement`, or runtime `if` on login state where an `_onprem.go` sibling belongs                 |
| `client-access`  | Auth/client constructed in a handler instead of via `PreRunner`                                              |
| `errors-format`  | Error/suggestion casing or `NewErrorWithSuggestions` not used (lint usually owns this)                       |
| `output-format`  | Backtick/quote convention in user-facing strings                                                             |
| `deletion`       | Hand-rolled confirmation instead of the `deletion` helpers                                                   |
| `docs`           | Stale help-text example, missing EA/OP callout                                                               |
| `internal-refs`  | JIRA/RFC/internal service names leaking into user-facing strings or stale comments                           |
| `secrets`        | Credentials or real IDs in the diff                                                                          |
| `style`          | Naming/conventions where `golangci-lint` / `lint-cli` do not enforce                                         |

## Tips

- Start with the PR description and the PR template checklist.
- For a new command, trace it end to end: file exists → registered on parent → `Args` set →
  annotation set → integration test + golden present. A missing link usually means the command is
  dead, panics, or is untested.
- Use the `Agent` tool with the Explore agent for blast-radius questions (e.g. "find every caller of
  this serialized struct before renaming a field").
- Companion skills to ground review in current state:
  - `docs-drift` — validates in-code help text and root markdown against the live command tree
  - `cli-design` (user-global) — clig.dev conventions for flag naming, arity, and output when the
    change adds or reshapes a command's surface
- When suggesting a simpler alternative, confirm it exists on the current Cobra / SDK version before
  posting.
