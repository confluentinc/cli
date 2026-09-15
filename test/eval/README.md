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
7. Writes results to `test/eval/results/environment-crosstalk.json`

The barrier ensures deterministic last-writer-wins collisions in shared mode and deterministic isolation in isolated mode.

## Why Behind the `eval` Build Tag

The harness is excluded from `make test` and runs only with `-tags eval` because:

- The mock backend (`test/test-server`) binds fixed ports and requires a `*testing.T`, so it cannot run as a standalone binary or alongside `make integration-test` (which rebuilds the CLI during execution)
- Concurrent sessions writing to shared files create race conditions that are only observable under genuine concurrency, not in isolation
- Phase-1 scenarios are specialized proofs that don't belong in the general test suite; they graduate to CI only once the agent layer lands (phase 2+)

## How to Read the Report

Results are written to `test/eval/results/environment-crosstalk.json`:

```json
{
  "build": "HEAD",
  "cells": {
    "shared": {
      "Trials": 5,
      "CollisionRate": 0.5,
      "CorruptionRate": 0,
      "PassCaretK": 0
    },
    "isolated": {
      "Trials": 5,
      "CollisionRate": 0,
      "CorruptionRate": 0,
      "PassCaretK": 1
    }
  }
}
```

**Metrics:**

- **Trials:** number of concurrent trials (independent test runs) per cell
- **CollisionRate:** fraction of sessions that read the wrong active environment (intended ≠ observed)
- **CorruptionRate:** fraction of sessions where `~/.confluent/config.json` failed validation (e.g., torn write)
- **PassCaretK (pass^k):** fraction of trials where _all_ sessions passed (no collisions or corruptions within that trial)

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
   - Later: live-CCloud smoke tests, CSV output (JSON only in phase 1)

The test loop in `TestEnvironmentCrosstalkEval` iterates over provisioners and scenarios; the `Report` and `CellMetrics` types are keyed by cell name (e.g., `"shared"`, `"isolated"`) to support arbitrary combinations of builds, provisioners, and scenarios.

## Architecture

- `provisioner.go`: `Provisioner` interface—each session gets a HOME dir via `HomeDir(session int)`
- `scenario.go`: `RunScenario` executes sessions concurrently with a barrier; `CommandFunc` runs one CLI invocation
- `grader.go`: `GradeTrial` grades one concurrent trial for collisions and corruption; `Aggregate` rolls outcomes into rates
- `metrics.go`: metric types: `CellMetrics`, `TrialOutcome`
- `report.go`: `Report` type, JSON serialization, `Summary()` for human-readable output
- `crosstalk_eval_test.go`: `TestEnvironmentCrosstalkEval` the flagship end-to-end eval
- `*_test.go`: unit tests for each module (grader, scenario, provisioner, metrics, report)

All files carry `//go:build eval` except `doc.go`.
