### query

Runs a bounded ("snapshot") Flink SQL statement to completion and returns the whole
result set, synchronously, from the client. Backs `confluent query`
(`internal/query/command.go`).

The verb, the flags and the result shape are all expected to move.

#### Why this exists

The Flink gateway has no synchronous execute endpoint. A statement is submitted, polled
until it leaves `PENDING`, and then its result pages are pulled one at a time following
`metadata.next`. `Run` performs that handshake behind a single call so a non-interactive
command can behave like an ordinary database query.

```go
result, err := query.Run(ctx, query.Options{
	Client:         gatewayClient,
	EnvironmentId:  environmentId,
	OrganizationId: organizationId,
	RequireBounded: true,
}, statementName)
```

The statement must already exist — submitting it is the caller's job, so the caller keeps
ownership of naming, properties and cleanup.

#### Why not reuse the shell

`Store` + `ResultFetcher` + `MaterializedStatementResults` already does submit, poll and
page. It is not reused here, because it was written for a scrolling viewer where being
wrong degrades to "the user presses refresh again". Three of those degradations become
silent data corruption once a script is reading stdout:

| Shell behavior | Consequence for a synchronous command |
| --- | --- |
| `MaterializedStatementResults.cleanup()` evicts from the front past `MaxResultsCapacity` (10,000) | a 50k-row `SELECT` prints the **last** 10k and exits 0 |
| `Append` skips rows whose field count ≠ header count, returning a bool that `fetchNextPageAndUpdateState` discards | short result set, no signal |
| `updateState` sets `Completed` on `PageToken == ""` without checking the phase | exit 0 with a partial result set |

This package handles each explicitly: no cap unless `Options.MaxRows` is set (and then
`Result.Truncated` says so), a hard error from `ConvertToInternalResults` on a schema
mismatch, and termination only when the page token is gone **and** the statement has
reached a terminal phase.

It also skips the shell's table-mode materialization. `Result.Rows` is the raw changelog
as the gateway delivered it. For a bounded append-only snapshot the changelog and the
materialized table are identical; for anything else the caller decides.

#### The drain loop

`page_token` is a positional offset into the collect-sink protocol, not an opaque cursor.
There is therefore no token that advances past a page which did not supply one.

The gateway's own foreground/streaming endpoint (`GetStatementResultEndpoint` in
`cc-flink-gateway-service-v2`, `internal/service/sql/v1/service.go`) only leaves `next`
empty when the JobManager itself reports `IsFinished == true` — every other case, even an
empty page, gets a fresh `next` token. So an empty `next` is not ambiguous at the
protocol level; it means the JobManager is genuinely done producing rows.

What can still go wrong is a race between two *separately updated* signals: the
JobManager's own `IsFinished` (which drives whether `next` is populated) versus the
statement's `Status.Phase`, read here via a separate `GetStatement` call reconciled by a
different subsystem. If that call lands just before the JobManager flips to finished,
`terminalBeforeFetch` comes back `false` even though the page fetched right after is
already the last one. `drain()` handles this by re-reading the phase once more instead of
conceding immediately:

- the page was **empty** — treated as "nothing yet". Re-requesting the same offset is
  harmless, so the loop backs off and retries.
- the page carried **rows** and the phase read before the fetch wasn't terminal — the
  loop re-reads the phase once more before giving up. Only if that second read is still
  non-terminal does it set `Result.Incomplete`, which the command surfaces as a warning.

This is confirmed from gateway source for the one execution path this package exercises
(foreground, JobManager-backed snapshot statements) — not verified for other paths (e.g.
any legacy/batch branch this package never hits), so the phase re-check stays as
defense-in-depth rather than being narrowed or removed.

#### Known limitations

- **Cloud only.** `Options.Client` is a `ccloudv2.GatewayClientInterface`. On-prem goes
  through CMF (`store_onprem.go`, itself a near-copy of `store.go`), so parity means a
  second implementation and a second test surface.
- **Token refresh is best-effort, not retry-aware.** `Options.RefreshToken` is invoked
  before each gateway call (see the command's `refreshGatewayToken`), unlike the shell's
  `synchronizedTokenRefresh`, which wraps every call including mid-flight retries. In
  practice this rarely matters: the command's default 10-minute `--timeout` is on the
  same order as the dataplane token's own lifetime, so a run is unlikely to still be
  going when a refresh would be needed. It only bites if `--timeout` is raised well past
  the default.
- **Expired-result handling lives in the command, not here.** This package just returns
  `ResultsFetchError` on any failed page fetch. `internal/query/command.go`'s
  `handleQueryError` is what distinguishes a 404 (statement deleted or mistyped) from a
  408 (the snapshot result window has closed) and gives each a targeted suggestion.
  Confirmed from gateway source: for `sql.snapshot.mode=now` statements, results are
  retained for exactly one hour after statement creation
  (`cc-flink-gateway-service-v2` `internal/service/sql/v1/service.go`,
  `GetStatementResultEndpoint`), after which the gateway returns 408 with an explicit
  message instead of continuing to page.
- **The whole result set is held in memory** as `[]types.StatementResultRow`. There is no
  streaming-to-stdout path, so the peak footprint scales with the result.
- **Ops are dropped from serialized output.** The command's `-o json` / `-o yaml` rows are
  keyed by column name and carry no `op`, so a non-append-only statement loses its
  update/delete markers. Human output grows an `Operation` column instead. Real gap if
  non-append-only ever needs supporting.
- **No statement cleanup on success.** Each query leaves a terminal statement behind
  against the 50K-per-environment pool.
- **`--unsafe-trace` dumps customer rows**, and this is the surface most likely to run in
  CI with retained logs.

#### Working on this

```bash
go test ./pkg/flink/query/                                              # unit tests, mocked gateway
```

The unit tests drive `pkg/flink/test/mock.MockGatewayClientInterface` and inject
`Options.sleep`, so backoff costs no wall time.

Two unrelated failures reproduce on a clean `main` and are not caused by changes here:
`pkg/flink/internal/controller` and `TestFlinkShell`/`TestFlinkShellOnPrem` panic without
a TTY. If `make lint-go` reports `unsupported version of the configuration`, a
golangci-lint v2 on `PATH` is shadowing the v1.64.8 the Makefile pins — run
`$(go env GOPATH)/bin/golangci-lint run` directly.
