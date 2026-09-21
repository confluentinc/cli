# CLI Agent Eval Harness

The eval harness measures the impact of state isolation on CLI robustness under concurrent multi-agent usage. It proves that isolating each agent's `~/.confluent/` configuration directory eliminates the state-corruption failures that occur under shared state.

## What It Proves

In multi-agent scenarios where concurrent sessions share a single `~/.confluent/config.json`, a last-writer-wins race condition causes state damage: collisions (an agent intends one outcome - e.g. active environment `env-596` - but reads another written by a concurrent peer), lost updates (one session's write silently overwrites another's), and torn-write config corruption. Isolation - giving each session its own home directory - eliminates these failure modes and their downstream impact on correctness.

Five scenarios exercise this, all driven by the same `Scenarios` registry in `scenario.go`:

- **environment-crosstalk** - concurrent `login` + `environment use` sessions racing over the active environment.
- **selection-cross-field** - two sessions set different `current_*` fields (`kafka cluster use` vs `flink compute-pool use`) on one shared context.
- **crud-lost-update** - two sessions each `api-key create --resource global`; a lockless whole-file save drops one.
- **auth-race** - one session keeps working while another `logout`s on the shared context.
- **mixed-workload** - three sessions doing unrelated real work (select an env; create an api-key; use a cluster then log out) on one shared context.

**Observed result** (2-3 concurrent sessions, 10 trials each): every scenario's shared cell shows
damage (`pass^k = 0`) while its isolated cell is perfectly clean (`pass^k = 1.0`). Representative
sample from the `environment-crosstalk` scenario: shared - 50% collision rate, 0% corruption rate,
`pass^k = 0`; isolated - 0% collision rate, 0% corruption rate, `pass^k = 1.0`. The exact split
between collisions and corruption varies by scenario and scheduling; `pass^k = 0` under shared state
is the only invariant across runs.

This demonstrates the value of the state-isolation effort (APIE-1515 multi-agent infrastructure) on the CLI v4 codebase today.

## How to Run

```bash
go test -tags eval ./test/eval/... -run TestEval -v
```

The test:

1. Builds the CLI with coverage instrumentation via `make build-for-integration-test`
2. Starts a mock Confluent Cloud backend on a fixed port
3. For each scenario in the `Scenarios` registry, runs both the `shared` and `isolated` provisioners across 10 trials
4. Each trial runs every session's `Setup` steps (e.g. login, environment selection) to completion, holds all sessions at a barrier until every session has finished setup, then releases them together to run their `Contend` steps (the racing writes) - maximizing the overlap between concurrent writes
5. Grades each session (invocation error, config corruption, or collision against the scenario's expected outcome, in that priority) via `assertRedSharedGreenIsolated` in `metrics.go`
6. Writes `test/eval/results/report.json` plus a multi-page HTML report: `test/eval/results/index.html` (one row per scenario) and one `test/eval/results/<slugified-scenario-name>.html` drill-down page per scenario (e.g. `environment-crosstalk.html`)

The outcome comes from grading each session's final on-disk `config.json` after all writes complete. In shared mode, concurrent sessions writing to one file race: depending on scheduling, that race surfaces as a collision (one session's write wins outright, leaving another pointed at the wrong state) or as corruption (an interleaved/torn write leaves the file unparseable or structurally incomplete) - either way, every trial fails (`pass^k = 0`). Isolated mode leaves each session's own file untouched by the others, so neither failure mode is reachable.

## Why Behind the `eval` Build Tag

The harness is excluded from `make test` and runs only with `-tags eval` because:

- The mock backend (`test/test-server`) binds fixed ports and requires a `*testing.T`, so it cannot run as a standalone binary or alongside `make integration-test` (which rebuilds the CLI during execution)
- Concurrent sessions writing to shared files create race conditions that are only observable under genuine concurrency, not in isolation
- These scenarios are specialized proofs that don't belong in the general test suite; they graduate to CI only once the real agent layer lands

## How to Read the Report

Results are written to `test/eval/results/report.json` and a multi-page, theme-aware HTML report
(follows the OS light/dark preference, no JavaScript):

- `test/eval/results/index.html` - one row per scenario: sessions x trials, "clean runs" (N of
  total trials where every session graded ok) for the shared and isolated cells, a plain-text
  summary of what damage the shared cell showed, and a link to that scenario's own page. A
  collapsible glossary at the top defines each metric.
- `test/eval/results/<slug>.html` per scenario (e.g. `environment-crosstalk.html`, from the
  scenario name run through the same slugify rule the report uses to link to it) - the full
  cell-summary table plus the `<details>` drill-down: trial → session → invocation, with each
  invocation's captured stdout/stderr.

**Captured transcripts:** each invocation's full stdout/stderr is captured verbatim into both
`report.json` and the HTML pages. That's safe today - the eval only runs against the mock backend,
`test/eval/results/` is gitignored, and this harness doesn't run in CI - but before pointing it at a
live backend (a later phase) or archiving `results/` as a CI artifact, add redaction of any auth
material from captured output first.

The JSON nests trial- and session-level detail under each cell, so a collision or corruption can be
traced back to the exact invocation that caused it:

```json
{
  "build": "a1b2c3d",
  "generated_at": "2026-09-16T00:00:00Z",
  "scenarios": [
    {
      "name": "environment-crosstalk",
      "description": "...",
      "sessions": 2,
      "trials": 10,
      "cells": [
        {
          "name": "shared",
          "metrics": {
            "trials": 10,
            "collision_rate": 0.5,
            "corruption_rate": 0,
            "error_rate": 0,
            "pass_caret_k": 0
          },
          "trials": [
            {
              "trial": 0,
              "all_passed": false,
              "sessions": [
                {
                  "session": 1,
                  "intent": "env-595",
                  "observed": "env-596",
                  "verdict": "collision",
                  "detail": "acted on \"env-596\", intended \"env-595\" (clobbered by a concurrent session)",
                  "invocations": [
                    {
                      "command": "login --url http://...",
                      "exit_code": 0,
                      "duration_ms": 12
                    },
                    {
                      "command": "environment use env-595",
                      "exit_code": 0,
                      "duration_ms": 8
                    }
                  ]
                }
              ]
            }
          ]
        }
      ]
    }
  ]
}
```

**Metrics (per cell):**

- **trials:** number of concurrent trials (independent test runs) in the cell
- **collision_rate:** fraction of sessions whose observed state didn't match what they intended (intended ≠ observed) - e.g. the wrong active environment, a dropped api-key, or resurrected credentials, depending on the scenario
- **corruption_rate:** fraction of sessions where `~/.confluent/config.json` failed validation (e.g., torn write)
- **error_rate:** fraction of sessions where a `confluent` invocation itself failed (nonzero exit, timeout, or spawn error) - graded separately from collisions so a broken run isn't miscounted as state damage. `assertRedSharedGreenIsolated`'s headline assertion checks collision/corruption rates only: an error-only shared run does not prove crosstalk was demonstrated, so it fails the eval rather than passing by coincidence
- **pass_caret_k (pass^k):** fraction of trials where _every_ session graded "ok". The HTML report's index page calls this "clean runs" and shows it as "N / total" - same number, plainer name; the glossary on that page notes the pass^k formalism.

**Per-session verdicts** (in priority order - the first that applies wins): `error` (an invocation
failed), `corruption` (config unparseable/incomplete), `collision` (observed state ≠ intended state),
`ok`. Each session's `detail` explains the verdict in plain text, and `invocations` carries the full
transcript (command, exit code, stdout, stderr, duration) for that session's run.

Each "session" within a trial represents one concurrent agent's config directory. Each "trial" is a full concurrent run under one provisioner. Rates are aggregated across all trials in a cell.

## Running Against Another Build

By default the eval builds the CLI from the current checkout. To run it against a pre-built binary
instead (e.g. one built from `origin/main`, or later a v5 branch), set `EVAL_CLI_BIN` to that
binary's path and `EVAL_BUILD_LABEL` to the ref it represents, so the report is labeled with the
binary's origin instead of the checkout's `HEAD`:

```bash
EVAL_CLI_BIN=/path/to/confluent EVAL_BUILD_LABEL=origin/main \
  go test -tags eval ./test/eval/ -run TestEval
```

The binary must be a `build-for-integration-test` build (coverage-instrumented, `isTest=true`) -
a plain `make build` binary routes `login --url` differently and won't take the mock backend's
cloud path.

## Designed-for Growth

The harness is structured to grow into three orthogonal dimensions:

1. **Build** (git ref → binary)
   - Today: single build = current tree (`HEAD`), or an explicit `EVAL_CLI_BIN` override
   - Future: multi-ref builds comparing v4 vs. v5

2. **Provisioner** (state isolation strategy)
   - Today: `shared` (all sessions → same HOME) and `isolated` (each session → unique HOME)
   - Future: provider-specific isolation strategies as agent layers land

3. **Scenario** (concurrent behavior under test)
   - Today: five scenarios in the `Scenarios` registry (environment-crosstalk, selection-cross-field, crud-lost-update, auth-race, mixed-workload) - see "What It Proves" above
   - Future: a real headless agent layer (`claude -p` + `PostToolUse` hook), live-CCloud smoke tests

`TestEval` iterates over `Scenarios`, and within each scenario over both provisioners; `Report` nests
`ScenarioReport` → `CellReport` (keyed by cell name, e.g., `"shared"`, `"isolated"`) → `TrialResult` →
`SessionOutcome` to support arbitrary build x provisioner x scenario combinations, and per-session
drill-down, as the harness grows.

## Architecture

- `provisioner.go`: `Provisioner` interface—each session gets a HOME dir via `HomeDir(session int)`
- `scenario.go`: `Invocation` (one captured `confluent` run), `SessionScript` (a session's `Setup`/`Contend` steps plus its descriptive `Label`), `RunScenario` executes sessions concurrently with a barrier and returns each session's captured `Invocations` in order, `Scenario`/`Scenarios` the registry of the five scenarios (workload + grading) driving `TestEval`
- `grader.go`: `ReadActiveEnvironment`, `GradeConfigIntegrity` - low-level config-file checks
- `metrics.go`: `GradeSession(r, intent, observe)` grades one session into a `Verdict` (`ok`/`collision`/`corruption`/`error`) with a `Detail` (`observe` reads the scenario-relevant state and is the only scenario-specific part); `GradeTrial` grades a trial's sessions; `Aggregate` rolls graded trials into `CellMetrics`; `assertRedSharedGreenIsolated` is the headline assertion applied per-scenario
- `report.go`: `CellReport`/`ScenarioReport`/`Report` nested types, `WriteJSON`, `WriteHTMLReport` (renders the embedded `templates/*.tmpl` set into `index.html` + one per-scenario page), `Summary()` for human-readable output. Template helpers: `slugify` (scenario name → filename), `codeify` (renders backtick-delimited spans as `<code>`, HTML-escaping everything else first), `cleanRuns`/`sharedDamage` (index-page summaries), `cellByName`.
- `templates/`: the committed HTML/CSS source - `styles.tmpl` (shared theme-aware CSS, light + `prefers-color-scheme: dark`), `index.tmpl` (scenario list + glossary), `scenario.tmpl` (per-scenario drill-down, including each session's role label). Parsed once via `embed.FS` at package init; a malformed template fails at import time, and `TestReportTemplatesParseAndResolveNames` guards that all three names resolve.
- `eval_test.go`: `TestEval` the flagship end-to-end eval, looping over `Scenarios` x provisioners x `evalTrials`; `realRun` captures each invocation's full transcript
- `*_test.go`: unit tests for each module (grader, scenario, provisioner, metrics, report)

All files carry `//go:build eval` except `doc.go`.
