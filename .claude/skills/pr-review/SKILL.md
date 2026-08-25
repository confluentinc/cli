---
name: pr-review
description:
  Reviews pull requests for the Confluent CLI. Use when reviewing PRs, doing self-review before
  sharing with the team, or when the user mentions "review PR", "help with PR", "review changes",
  "self-review", "review local changes", or "check my PR". Focuses on Cobra command wiring,
  Cloud/On-Prem mode gating, error/output formatting, backward compatibility, and golden-file
  integration tests.
allowed-tools:
  - Read
  - Bash
  - Grep
  - Glob
  - Task
---

# PR Review Skill

Reviews pull requests for the Confluent CLI (Go + Cobra), focusing on project-specific patterns and
the failure modes most likely to slip past `golangci-lint` and `lint-cli`.

The authoritative conventions live in `AGENTS.md` (symlinked as `CLAUDE.md`) and the per-package
`README.md` files under `pkg/` (`pkg/errors/README.md`, `pkg/output/README.md`,
`pkg/deletion/README.md`, `pkg/cmd/ANNOTATIONS.md`). This skill applies them to a diff; when a
convention is ambiguous, read the source of truth rather than re-deriving it from the diff.

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
- Verify Cobra command wiring is complete (registration, `Args`, annotations, completion)
- Confirm every new command/flag has a golden-file integration test
- Catch backward-compatibility breaks before they reach a reviewer

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

#### 1. Cobra command wiring (MANDATORY for new commands/subcommands)

- [ ] Command is registered (subcommand added to its parent; new top-level command wired in
      `internal/command.go`)
- [ ] `Args` is set and matches arity: `cobra.ExactArgs(N)`, or `cobra.MinimumNArgs(1)` for
      variadic/`delete` commands
- [ ] `ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs)` present where completion applies
- [ ] Mode gating uses `annotations.RunRequirement` (`RequireCloudLogin` / `RequireOnPremLogin` /
      etc.), not a runtime `if` on login state

Red flags:

- New subcommand file but no registration on the parent → command silently does not exist
- A command that reads positional args with no `Args` validator → panics or misbehaves on bad input
- `if cfg.IsCloudLogin() { ... } else { ... }` splitting behavior that should be an annotation +
  `_onprem.go` sibling

#### 2. Error and output formatting

- [ ] Error strings are lowercase, no trailing period, `...ErrorMsg`; suggestions are full
      sentences, `...Suggestions`; combined with `errors.NewErrorWithSuggestions`
- [ ] User-facing strings use backticks for commands/flags and double quotes for names/IDs
- [ ] Auth/config errors name the missing or invalid setting

#### 3. Client and auth access

- [ ] Handlers access clients lazily via `c.V2Client` / `c.MDSClient` / `c.GetKafkaREST()` — they
      do not authenticate or construct clients themselves (that belongs in `PreRunner`)

#### 4. Backward compatibility

- [ ] No renamed/removed command, subcommand, or flag outside a major release
- [ ] No changed/removed serialized field name under `-o json` / `-o yaml`; no new `omitempty` on a
      previously-always-emitted field; no output-format change
- [ ] EA/OP instability is disclosed in the command's `Short`/`Long`

A useful check for the serialized-output surface: if a `*.golden` file for a `json`/`yaml` test
changed, confirm the field-name/shape delta is intentional and non-breaking (or gated on a major).

#### 5. Testing

- [ ] Every new command and flag has an integration test in `test/` with matching golden files
- [ ] Golden diffs reflect the intended user-visible change only (no incidental churn from an
      unrelated merge)
- [ ] New non-trivial logic in `internal/`/`pkg/` has a co-located unit test

### Step 5: Check Project-Specific Patterns

#### Multi-resource delete

- [ ] Delete commands accept variadic args (`cobra.MinimumNArgs(1)`) and route confirmation +
      deletion through `deletion.ValidateAndConfirm` and `deletion.Delete` rather than a hand-rolled
      prompt/loop

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

- [ ] Cobra wiring (registration + Args + annotations + completion): [status, location of any gap]
- [ ] Error / output formatting: [status]
- [ ] Client access via PreRunner: [status]
- [ ] Backward compatibility (commands/flags, -o json/yaml): [status]
- [ ] Integration tests + golden files for new commands/flags: [status]

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

| Category        | Description                                                              |
| --------------- | ------------------------------------------------------------------------ |
| `cobra-wiring`  | Missing registration, `Args`, annotation gating, or completion hookup    |
| `mode-gating`   | Runtime `if` on login state where an annotation / `_onprem.go` belongs   |
| `errors`        | Error/suggestion casing or `NewErrorWithSuggestions` not used            |
| `output`        | Backtick/quote convention; serialized-field formatting                   |
| `compatibility` | Breaking command/flag rename or `-o json`/`-o yaml` change outside major |
| `deletion`      | Hand-rolled confirmation instead of the `deletion` helpers               |
| `client-access` | Auth/client constructed in a handler instead of via `PreRunner`          |
| `testing`       | Missing integration test/golden for a new command/flag; incidental churn |
| `docs`          | Stale help-text example, missing EA/OP callout                           |
| `secrets`       | Credentials or real IDs in the diff                                      |
| `style`         | Naming/conventions where `golangci-lint` / `lint-cli` do not enforce     |

## Tips

- Start with the PR description and the PR template checklist.
- For a new command, trace it end to end: file exists → registered on parent → `Args` set →
  annotation set → integration test + golden present. A missing link usually means the command is
  dead, panics, or is untested.
- Use `Task` with the Explore agent for blast-radius questions (e.g. "find every caller of this
  serialized struct before renaming a field").
- Companion skills to ground review in current state:
  - `docs-drift` — validates in-code help text and root markdown against the live command tree
  - `cli-design` (user-global) — clig.dev conventions for flag naming, arity, and output when the
    change adds or reshapes a command's surface
- When suggesting a simpler alternative, confirm it exists on the current Cobra / SDK version before
  posting.
