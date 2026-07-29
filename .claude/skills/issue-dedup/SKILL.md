---
name: issue-dedup
description: Finds and resolves duplicate GitHub issues in confluentinc/cli, with a lightweight check for related issues in the sibling confluentinc/terraform-provider-confluent repo. Scans open issues, clusters likely duplicates with a suggested canonical issue, and — only after explicit confirmation — comments, labels, and closes confirmed duplicates. Use when asked to find, flag, merge, close, or clean up duplicate issues in this repo. Scanning is read-only; resolving mutates GitHub and always confirms first.
---

# Issue Dedup — confluentinc/cli

Two phases in one skill: **Scan** (read-only, always safe to run) and **Resolve** (mutates GitHub, requires explicit confirmation). Do not skip straight to Resolve — always scan (or accept a prior scan's output) first.

## Phase 1 — Scan (read-only)

1. **Scope**: default to all open issues in `confluentinc/cli`. If the user narrows it (a label, a time window, a keyword), use that instead. If there are more than ~500 open issues, report the count and ask whether to filter before pulling everything.

   ```
   gh issue list --repo confluentinc/cli --state open \
     --json number,title,body,labels,createdAt,updatedAt,comments,url --limit 500
   ```

2. **Normalize**: strip issue-template boilerplate (fixed headers, checkbox scaffolding) before comparing bodies — it inflates similarity between unrelated issues that just used the same template. Focus on what varies: error messages, exact commands, stack traces, version strings, repro steps.

3. **Cluster** (don't do a naive O(n²) full-text pass):
   - Block by shared signals first: same error string/exit code, same subcommand (e.g. `confluent kafka topic create`), overlapping title tokens, same label.
   - Within a block, actually read the normalized bodies and judge whether they're the same bug/request, not just the same topic. "Topic creation fails" with two different root causes is not a duplicate pair.

4. **Cross-repo check**: for clusters that look like they might stem from a platform/API behavior rather than a CLI-specific bug, do one targeted search against the sibling repo before concluding:

   ```
   gh issue list --repo confluentinc/terraform-provider-confluent --state open --search "<keyword>"
   ```

   If you find a match, flag it as **cross-repo related** rather than a same-repo duplicate — see classification below.

5. **Pick a canonical issue** per group: prefer most complete repro/description, then oldest by `createdAt`, then most comments/reactions. State which rule decided it.

6. **Classify**:
   - **Same-repo duplicate** (High/Medium/Low confidence) — candidate for closing in Phase 2.
   - **Cross-repo related** (High/Medium/Low confidence) — same underlying platform behavior surfacing in both repos. Default action is cross-linking, not closing either side.

   For every group, give one or two sentences of *concrete* reasoning (shared error text, same command, same field) — never just "these look similar."

7. **Output**: write a markdown report (default `./issue-dedup-report-<YYYY-MM-DD>.md`) with one entry per group:

   ```markdown
   ## Group 1 — [same-repo duplicate | cross-repo related] — confidence: High/Medium/Low
   Canonical: #123 <title> (<url>)
   Duplicates:
   - #456 <title> (<url>)
   Reasoning: <concrete, specific>
   ```

   Then summarize in chat: issues scanned, groups found, top few by confidence. Stop here unless the user asks you to resolve.

## Phase 2 — Resolve (mutates GitHub, confirm first)

Only enter this phase against a scan you (or a prior run) just produced, or a group the user hands you directly. Never invent duplicate judgments from scratch here.

1. **Re-verify** each issue in each group before trusting the report — reports go stale:

   ```
   gh issue view <number> --repo confluentinc/cli --json state,title,labels,comments,body
   ```

   Drop/flag a group if an issue is already closed, has picked up substantial new discussion diverging from the canonical thread (treat as "needs manual review"), or its body has changed enough to undercut the original judgment.

2. **Show the exact plan before acting** — literal comment text and commands, not a summary:

   ```
   Canonical: #123 <title>
   Will close as duplicate: #456 <title>
     Comment: "Closing as a duplicate of #123. If this doesn't fully capture your issue, please reopen or comment on the original with details."
     Label: duplicate
     Close reason: not planned
   ```

   For **cross-repo related** groups, default to cross-linking only (a comment on each issue pointing at the other) — do not close either side unless the user explicitly says to.

3. **Wait for explicit confirmation.** Confirm per group unless the user explicitly approves the whole batch — and even then, show every group's plan first.

4. **Execute only confirmed groups:**

   ```
   gh issue comment <dup> --repo confluentinc/cli --body "<comment>"
   gh issue edit <dup> --repo confluentinc/cli --add-label duplicate
   gh issue close <dup> --repo confluentinc/cli --reason "not planned"
   ```

   For a cross-repo cross-link, the corresponding comment goes on the `confluentinc/terraform-provider-confluent` issue via the same `gh issue comment` pattern with that repo. If `gh` reports a permissions error, stop and tell the user — don't retry or route around it.

5. **Log and summarize**: append each executed action (issue, action, timestamp) to `./issue-dedup-actions-<YYYY-MM-DD>.md`, then report groups resolved, groups skipped (why), and groups left for manual review.

## Defaults

- Treat every run as dry-run (show the plan, don't execute) unless the user's message already contains clear approval to act ("go ahead", "do it", "apply these").
- Never bulk-close a group with materially divergent discussion — surface it for manual review instead.
