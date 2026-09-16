# CLI Agent Eval Harness

The eval harness measures the impact of state isolation on CLI robustness under concurrent multi-agent usage. Phase 1 proves that isolating each agent's `~/.confluent/` configuration directory eliminates environment-selection collisions that occur under shared state.

## What It Proves

In multi-agent scenarios where concurrent sessions share a single `~/.confluent/config.json`, a last-writer-wins race condition causes environment-selection collisions: an agent intends to use one environment (e.g., `env-596`) but reads another (e.g., `env-595`) written by a concurrent peer. Isolation—giving each session its own home directory—eliminates these collisions and their downstream impact on correctness.

**Phase 1 observed result** (2 concurrent sessions, 5 trials):

- **Shared state:** 50% collision rate, 0% corruption rate, 0.0 pass^k (no trials passed)
- **Isolated state:** 0% collision rate, 0% corruption rate, 1.0 pass^k (all trials passed)

This demonstrates the value of the state-isolation effort (APIE-1515 multi-agent infrastructure) on the CLI v4 codebase today.

## How to Run

```bash
go test -tags eval ./test/eval/... -run TestEnvironmentCrosstalkEval -v
```

The test:

1. Builds the CLI with coverage instrumentation via `make build-for-integration-test`
2. Starts a mock Confluent Cloud backend on a fixed port
3. Runs two concurrent sessions (each targeting a different environment: `env-596` and `env-595`)
4. Executes each session through login and environment selection
5. Holds all sessions at a barrier until every session has written its environment choice
6. Grades the resulting config for collisions and corruption
7. Writes results to `test/eval/results/environment-crosstalk.json` and `test/eval/results/index.html`

The barrier is phase-2 scaffolding for a future post-barrier read/act step; in phase 1 nothing happens after it, so it does not drive the result. The deterministic outcome comes from grading each session's final on-disk `config.json` after all writes complete - in shared mode, two sessions writing distinct environments to one file leave exactly one intended environment surviving (a genuine clobber), and isolated mode leaves each session's own file untouched by the other.

## Why Behind the `eval` Build Tag

The harness is excluded from `make test` and runs only with `-tags eval` because:

- The mock backend (`test/test-server`) binds fixed ports and requires a `*testing.T`, so it cannot run as a standalone binary or alongside `make integration-test` (which rebuilds the CLI during execution)
- Concurrent sessions writing to shared files create race conditions that are only observable under genuine concurrency, not in isolation
- Phase-1 scenarios are specialized proofs that don't belong in the general test suite; they graduate to CI only once the agent layer lands (phase 2+)

## How to Read the Report

Results are written to `test/eval/results/environment-crosstalk.json` and a drill-down
`test/eval/results/index.html` (open the HTML in a browser for the coverage-report view - summary
table per cell, then `<details>` down to each trial, each session, and each invocation's captured
stdout/stderr).

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
      "trials": 5,
      "cells": [
        {
          "name": "shared",
          "metrics": {
            "trials": 5,
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
                  "intended_env": "env-595",
                  "observed_env": "env-596",
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
- **collision_rate:** fraction of sessions that read the wrong active environment (intended ≠ observed)
- **corruption_rate:** fraction of sessions where `~/.confluent/config.json` failed validation (e.g., torn write)
- **error_rate:** fraction of sessions where a `confluent` invocation itself failed (nonzero exit, timeout, or spawn error) - graded separately from collisions so a broken run isn't miscounted as state damage
- **pass_caret_k (pass^k):** fraction of trials where _every_ session graded "ok"

**Per-session verdicts** (in priority order - the first that applies wins): `error` (an invocation
failed), `corruption` (config unparseable/incomplete), `collision` (observed env ≠ intended env),
`ok`. Each session's `detail` explains the verdict in plain text, and `invocations` carries the full
transcript (command, exit code, stdout, stderr, duration) for that session's run.

Each "session" within a trial represents one concurrent agent's config directory. Each "trial" is a full concurrent run under one provisioner. Rates are aggregated across all trials in a cell.

## Designed-for Growth

The harness is structured to grow into three orthogonal dimensions:

1. **Build** (git ref → binary)
   - Phase 1: single build = current tree (`HEAD`)
   - Phase 4+: multi-ref builds comparing v4 vs. v5

2. **Provisioner** (state isolation strategy)
   - Phase 1: `shared` (all sessions → same HOME) and `isolated` (each session → unique HOME)
   - Phase 2+: provider-specific isolation strategies as agent layers land

3. **Scenario** (concurrent behavior under test)
   - Phase 1: `environment-crosstalk` (concurrent `confluent login` + `confluent environment use`)
   - Phase 2+: additional crosstalk scenarios (cluster selection, API key collision, login context)
   - Phase 3+: real headless agent layer (`claude -p` + `PostToolUse` hook)
   - Later: live-CCloud smoke tests, CSV output (JSON + HTML drill-down in phase 1)

The test loop in `TestEnvironmentCrosstalkEval` iterates over provisioners (phase 1 runs a single scenario); `Report` nests `ScenarioReport` → `CellReport` (keyed by cell name, e.g., `"shared"`, `"isolated"`) → `TrialResult` → `SessionOutcome` to support arbitrary build x provisioner x scenario combinations, and per-session drill-down, in later phases.

## Architecture

- `provisioner.go`: `Provisioner` interface—each session gets a HOME dir via `HomeDir(session int)`
- `scenario.go`: `Invocation` (one captured `confluent` run), `RunScenario` executes sessions concurrently with a barrier and returns each session's captured `Invocations` in order
- `grader.go`: `ReadActiveEnvironment`, `GradeConfigIntegrity`, `GradeTargetFidelity` - low-level config-file checks
- `metrics.go`: `GradeSession` grades one session into a `Verdict` (`ok`/`collision`/`corruption`/`error`) with a `Detail`; `GradeTrial` grades a trial's sessions; `Aggregate` rolls graded trials into `CellMetrics`
- `report.go`: `CellReport`/`ScenarioReport`/`Report` nested types, `WriteJSON`, `WriteHTML` (self-contained drill-down page), `Summary()` for human-readable output
- `crosstalk_eval_test.go`: `TestEnvironmentCrosstalkEval` the flagship end-to-end eval; `realRun` captures each invocation's full transcript
- `*_test.go`: unit tests for each module (grader, scenario, provisioner, metrics, report)

All files carry `//go:build eval` except `doc.go`.
