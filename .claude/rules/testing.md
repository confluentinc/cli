---
paths:
  - "test/**"
  - "**/*_test.go"
---

# Test coverage: failure paths, not just a golden

- Every new command and flag needs an integration test in `test/` with matching golden files (global
  flags like `-v` / `--unsafe-trace` are exempt).
- Each new command needs a failure-path case, not only the happy path (e.g. not-found and
  missing-required-flag). A green happy-path golden hides an untested error branch.
- Mock handlers should assert the request they received (forwarded query params, method, body), not
  just return a canned response. Otherwise a wrong-value-sent-to-server bug (including data-loss
  paths like `delete --version`) still passes its golden.
- Golden files are regenerated with `-update` only when the diff is the intended user-visible change;
  golden churn unrelated to the PR's purpose is a red flag.
- Unit tests are co-located with source in `internal/` and `pkg/` (testify, suite-based); a new
  frequently-reused helper deserves one.
