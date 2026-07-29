---
name: docs-drift
description: Validates confluentinc/cli's in-code documentation (Short/Long/Example text on Cobra commands, plus root markdown like README.md/CONTRIBUTING.md) against the live command tree to catch stale references to deprecated or hidden commands/flags, wrong/missing/extra flags in examples, typos, and broken links. Use when asked to check for documentation drift, validate CLI examples/help text, or audit docs for staleness. Read-only by default — only edits a file when explicitly asked to fix something.
---

# Docs Drift — confluentinc/cli

The CLI has no separate hand-maintained `docs/` tree. Its "docs" are the `Short`/`Long`/`Example` fields embedded directly in each Cobra command (e.g. `internal/kafka/command_topic_create.go`), plus root-level markdown. Ground truth is the live command/flag tree, not any static file.

## 1. Reuse the existing linter — don't re-derive it

```
make lint-cli
```

This runs `pkg/linter.RequireValidExamples()` (`pkg/linter/command_rules.go:220`) over the live command tree: every `IsFlagRequired` flag must appear in the example, every `--flag` token in an example must exist on that command's `pflag.FlagSet`, and `--flag=value` syntax is rejected outright. It also runs hunspell-backed `CommandRule`/`FlagRule` checks (naming, capitalization, punctuation) on `Use`/`Short`/`Long`/flag usage strings. Capture and categorize this output as doc-drift findings rather than reimplementing the flag-matching logic by hand.

## 2. Find deprecated/hidden surface still referenced in prose

Commands and flags are hidden or disabled via:
- `cmd.Flags().MarkHidden("flag-name")` (scattered across command files, e.g. `internal/kafka/command_topic_produce.go`)
- `command.Hidden = true` (see `internal/command.go:229`)
- `pkg/featureflags.DisableHelpText(cmd, flags)` — LaunchDarkly-driven hiding (`pkg/featureflags/disable.go`)

LaunchDarkly can also deprecate visible surface area via `featureflags.DeprecateCommandTree` and `featureflags.DeprecateFlags`, which prefix command `Short`/`Long` text or flag `Usage` with `DEPRECATED:`. Grep for all hiding, disabling, and deprecation patterns repo-wide to build a list of affected commands and flags, then cross-reference each name against every `Example:`/`Long:` string that still mentions it by exact name. A hidden/disabled command or flag shown in another example is stale; a deprecated one shown without an appropriate warning is a drift candidate. Report the file:line of both the mechanism and the stale reference.

## 3. Cross-check subcommands mentioned in prose

Some `Long:` fields reference other subcommands by name (e.g. "run `confluent kafka topic list` first"). Build the live command tree the same way `cmd/docs/main.go` and `cmd/lint/main.go` do (`internal.NewConfluentCommand(cfg)`), then verify every `confluent <subcommand path>` mentioned in prose still resolves to a real, non-hidden command. A renamed or removed subcommand referenced this way is a direct hit — this is the CLI-side analog of a doc pointing at a resource/attribute that no longer exists.

## 4. Root markdown: typos and broken links

For `README.md`, `CONTRIBUTING.md`, and any other repo-root markdown:
- Extract `[text](url)` links. For relative links, verify the target file exists. For `http(s)://` links, follow redirects during a quick HEAD check (`curl -sSIL -o /dev/null -w '%{http_code}' <url>`). If the server rejects HEAD with `405` or `501`, retry with a lightweight GET that discards the body (`curl -sSL -o /dev/null -w '%{http_code}' <url>`). Treat other final 4xx/5xx responses as broken, and skip anything pointing at internal/non-public infrastructure.
- Cross-reference command/flag names mentioned in prose against the real command surface built in step 3 — a misspelled command (`confluent kafak topic`) won't resolve and is a giveaway.

## 5. Optional: regenerate the full reference tree for a manual pass

`go run cmd/docs/main.go` renders the entire live command tree to `.rst` under `docs/` (gitignored, not committed, and not wired into `make`/CI). Regenerating and skimming it is useful for a broader manual read beyond the automated checks above, but there's no prior committed version to diff against — don't treat it as a source of drift by itself, just as an on-demand snapshot.

## Output

Report findings grouped by category — `lint-cli` failures, stale deprecated/hidden reference, broken subcommand reference, typo, broken link — each with a file:line and a one-line concrete explanation (not "looks wrong"). Stay read-only: only edit a file if the user explicitly asks you to fix something, and show the diff before/after.
