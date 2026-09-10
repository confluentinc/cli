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

This package handles each explicitly: no cap unless `Options.MaxRows` is set (and then
`Result.Truncated` says so), and a hard error from `ConvertToInternalResults` on a schema
mismatch. Terminating on an empty page token, same as the shell's `updateState`, turns
out to be correct rather than a shell-only shortcut — see the drain loop section below.

It also skips the shell's table-mode materialization. `Result.Rows` is the raw changelog
as the gateway delivered it. For a bounded append-only snapshot the changelog and the
materialized table are identical; for anything else the caller decides.

#### The drain loop

`page_token` is a positional offset into the collect-sink protocol, not an opaque cursor.
There is therefore no token that advances past a page which did not supply one.

An earlier version of this loop treated a token-less page as ambiguous whenever the
statement's own `Status.Phase` wasn't yet terminal, re-checking phase once before
conceding `Result.Incomplete`. That was wrong, and a real, reproducible bug: a bounded,
`LIMIT`-satisfied query delivering every one of the requested rows still printed "may be
incomplete", because a `LIMIT`-bounded read over a streaming source can leave the
underlying job in `RUNNING` indefinitely — the row stream ending is not a promise that
the job itself will ever reach a terminal phase on its own, so waiting for one is waiting
for something that may never happen.

The gateway's protocol guarantee is simpler and doesn't need phase at all: it never omits
the next-page token while more rows remain. A token-less page **is** the complete
signal — done, full stop — regardless of what `Status.Phase` says. `drain()` trusts that:

- no next token → done. `refreshStatement` re-reads the statement once, purely so
  `Result.Statement`/`Phase()` reflect where things actually landed (e.g. `FAILED`), not
  for the completion decision itself.
- a next token with zero rows → "nothing new yet, keep polling this cursor". The loop
  backs off between empty pages so an idle wait doesn't hammer the gateway.

There is no more "incomplete" outcome. If the job hasn't reached a terminal phase once
draining is done — which a `LIMIT`-bounded read over a streaming source routinely
doesn't — `internal/query/command.go`'s deferred cleanup stops it, the same as it does
after a `Truncated` (`--max-rows`) read.

#### Known limitations

- **Cloud only.** `Options.Client` is a `ccloudv2.GatewayClientInterface`. On-prem goes
  through CMF (`store_onprem.go`, itself a near-copy of `store.go`), so parity means a
  second implementation and a second test surface.
- **Token refresh is best-effort, not retry-aware.** `Options.RefreshToken` is invoked
  before each gateway call (see the command's `refreshGatewayToken`), unlike the shell's
  `synchronizedTokenRefresh`, which wraps every call including mid-flight retries. In
  practice this rarely matters: the command's default 10-minute `--wait-timeout` is on
  the same order as the dataplane token's own lifetime, so a run is unlikely to still be
  going when a refresh would be needed. It only bites if `--wait-timeout` is raised well
  past the default, or a single call runs long past it — see the next point.
- **`GatewayClientInterface` takes no context, so `--wait-timeout` can't actually abort an
  in-flight call.** `ccloudv2.FlinkGatewayClient` builds every request from
  `context.Background()` internally; confirmed against real staging, where a single
  `GetStatementResults` call once hung for 49 minutes despite a 2-minute
  `--wait-timeout`. `callWithContext` races each call against `ctx` in a goroutine so
  `Run` still returns once the deadline fires, but it cannot cancel the underlying HTTP
  call — that goroutine keeps running until the transport itself gives up. A proper fix
  means threading a real context through `GatewayClientInterface` and every caller
  (the interactive shell included), which is out of scope for this package alone.
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
`Options.sleep` (drain's retry backoff) and `Options.pollInterval` (await's
wait.PollPhases interval), so waiting costs negligible wall time.

Two unrelated failures reproduce on a clean `main` and are not caused by changes here:
`pkg/flink/internal/controller` and `TestFlinkShell`/`TestFlinkShellOnPrem` panic without
a TTY. If `make lint-go` reports `unsupported version of the configuration`, a
golangci-lint v2 on `PATH` is shadowing the v1.64.8 the Makefile pins — run
`$(go env GOPATH)/bin/golangci-lint run` directly.
