# Local CLI build with Switchover support

> Paths below assume this repo (and its sibling repos) are cloned under
> `~/sdk_cli_tf_switchover/`, matching the original workspace this branch was
> built in. Adjust paths if you clone elsewhere.

How to reproduce the local `confluent` CLI build in this workspace, which adds
`confluent switchover pair` and `confluent switchover endpoint` commands on
top of the upstream CLI.

## Workspace layout

```
~/sdk_cli_tf_switchover/
├── cli/                          confluentinc/cli — branch `switchover-cli-local-demo` (PR #3399)
├── ccloud-sdk-go-v2-internal/    confluentinc/ccloud-sdk-go-v2-internal — branch `master`, unmodified
├── cc-switchover/                confluentinc/cc-switchover — branch `master`, unmodified (reference only)
└── terraform-provider-confluent/ confluentinc/terraform-provider-confluent — branch `master`, unmodified (not used here)
```

The switchover Go SDK (`ccloud-sdk-go-v2-internal/switchover/v1`) is already
merged to that repo's `master` (PR #818) and was verified against the current
API spec (`confluentinc/api` PR #2545, `switchover/openapi.yaml`) — a full
regen produced a byte-for-byte identical tree, so `cli` pins directly to
`ccloud-sdk-go-v2-internal` master's current commit rather than a new branch:

```
github.com/confluentinc/ccloud-sdk-go-v2-internal/switchover v0.0.0-20260707163957-4e3e503e8b10
```

(that pseudo-version encodes commit `4e3e503e8b10c538379dedd05fa055b41c14298b`
on `ccloud-sdk-go-v2-internal`'s `master`).

**Branch:** the changes below are on `cli` branch
[`switchover-cli-local-demo`](https://github.com/confluentinc/cli/tree/switchover-cli-local-demo),
pushed as PR **[confluentinc/cli#3399 — "\[local-test-only\] Add switchover
pair/endpoint commands"](https://github.com/confluentinc/cli/pull/3399)**.
This PR is for local-test-only purposes and is not intended to be merged —
`cc-switchover`'s backend doesn't fully implement the API yet, and a fuller,
tested implementation already exists in
[confluentinc/cli#3382](https://github.com/confluentinc/cli/pull/3382)
(branch `ORC-10127-add-switchover-commands`, which also covers `list`/`delete`
and has integration test coverage). To get this branch locally:

```bash
git clone git@github.com:confluentinc/cli.git
cd cli
git checkout switchover-cli-local-demo
```

`confluentinc/cli` is a public repo with an org ruleset that blocks direct
branch pushes — new branches must go through `git push-external` (Confluent's
Airlock proprietary-code-scan wrapper), not plain `git push`, if you need to
push further changes to this branch yourself.

## One-time environment setup

1. **Go toolchain** — `cli/.go-version` pins `1.26.4`. If your `goenv` doesn't
   have it yet, any Go ≥1.21 binary will auto-fetch it via `GOTOOLCHAIN=auto`:
   ```bash
   export GOTOOLCHAIN=auto
   export PATH="$HOME/.goenv/versions/<any-installed-1.2x>/bin:$PATH"
   go version   # should report go1.26.4 once it downloads
   ```
2. **goreleaser** — the CLI's `make build` shells out to it:
   ```bash
   brew install goreleaser
   ```
3. **Docker Desktop**, signed in to an org with access to
   `519856050701.dkr.ecr.us-west-2.amazonaws.com` — only needed if you want to
   regenerate the switchover SDK from `confluentinc/api`, not for building the
   CLI itself.

## Build

```bash
cd ~/sdk_cli_tf_switchover/cli
export GOTOOLCHAIN=auto
export PATH="$HOME/.goenv/versions/<any-installed-1.2x>/bin:$PATH"
make build
```

Binary lands at:

```
~/sdk_cli_tf_switchover/cli/dist/confluent_darwin_arm64_v8.0/confluent
```

(path varies by OS/arch — check `dist/` after the build).

## Alias the binary (recommended)

This binary isn't installed anywhere on `PATH`, so you either type the full
path every time or set up a shortcut. Prefer a distinctly-named **alias**
over adding the `dist/` folder to `PATH`: putting `dist/` on `PATH` would
shadow any real `confluent` CLI you have installed, so every plain
`confluent` invocation in that shell — including unrelated, real work —
would silently run this experimental build instead.

```bash
alias confluent-dev="~/sdk_cli_tf_switchover/cli/dist/confluent_darwin_arm64_v8.0/confluent"
```

Add that line to your shell rc file (e.g. `~/.zshrc` or `~/.aliases` if
sourced from it) to make it permanent, or just run it once per terminal
session. Your regular `confluent` command (if installed) is unaffected
either way. The rest of this README uses `confluent-dev`.

## Usage

Log in first — the `switchover` command requires cloud login
(`RequireNonAPIKeyCloudLogin`):

```bash
confluent-dev login
```

This hits **production** Confluent Cloud (`https://confluent.cloud`) with
real credentials — it's the CLI's stock login flow, not a mock.

### Logging in against staging (`stag`) instead of production

```bash
confluent-dev login --url https://stag.cpdev.cloud
```

Note the hostname is `stag.cpdev.cloud` with **no** `confluent.` prefix
(unlike prod's `confluent.cloud`) — `confluent.stag.cpdev.cloud` doesn't
resolve (`NXDOMAIN`); `stag.cpdev.cloud` does, and `api.stag.cpdev.cloud` is
just a CNAME alias to it. Verify with `host stag.cpdev.cloud` if in doubt.

`--url`'s help text says "for on-premises deployments," but the code path is
generic: `pkg/ccloudv2/utils.go`'s `IsCCloudURL` matches any URL containing
`confluent.cloud`, `cpdev.cloud`, or the gov domains and routes it through
the normal Cloud login (not the on-prem/MDS path), so a staging URL works
here. You'll need:

- A **staging** Confluent Cloud account — separate from your prod account,
  same email won't necessarily carry over. A generic
  "incorrect email, password, or organization ID" error after entering a
  password usually means this account doesn't exist on staging.
- There's no `--sso` flag — SSO is automatic. After you enter your email,
  the CLI asks the backend whether that account is SSO-enabled
  (`pkg/auth/login_credentials_manager.go`'s `isSSOUser`) and transparently
  switches to a browser-based flow instead of prompting for a password if
  so. `--no-browser` exists to do that browser flow headlessly, but it
  doesn't force SSO for an account the backend doesn't already recognize as one.
- If you know your staging org ID, try `--organization <org-id>`.
- Your staging org still needs Switchover Early Access granted before
  `switchover pair` commands return anything but an access/authorization
  error — Early Access isn't bypassed just by using staging.
- `devel.cpdev.cloud` works the same way for the `devel` environment, if needed.

Everything below (`switchover pair`/`switchover endpoint` commands) works
identically once logged in, regardless of which environment you logged into
— the CLI just talks to whichever environment's API you authenticated
against.

### Switchover pairs (backend implemented: create/get/update/trigger-switch)

```bash
# Create a pair between two Kafka clusters
confluent-dev switchover pair create prod-kafka-dr \
  --member name=west,id=lkc-111111 \
  --member name=east,id=lkc-222222 \
  --active-member west \
  --environment env-abcdef

# Describe it
confluent-dev switchover pair describe sw-123456

# Rename it (the only mutable field)
confluent-dev switchover pair update sw-123456 --display-name "prod-kafka-dr-renamed"

# Trigger a failover (prompts for confirmation — moves live traffic)
confluent-dev switchover pair trigger-switch sw-123456 --active-member east --failover-type CLEAN
```

`list` and `delete` are intentionally not implemented in the CLI —
`cc-switchover`'s backend doesn't implement those two RPCs yet (they return
errors regardless of client).

### Switchover endpoints (backend not implemented yet — commands will fail against a live environment)

```bash
confluent-dev switchover endpoint create prod-kafka-dr-endpoint \
  --switchover-pair sw-123456 \
  --endpoint name=west-platt,resource-id=lkc-west,type=PRIVATE \
  --endpoint name=east-platt,resource-id=lkc-east,type=PRIVATE

confluent-dev switchover endpoint describe se-123456
confluent-dev switchover endpoint update se-123456 --display-name "renamed-endpoint"
confluent-dev switchover endpoint activate se-123456
```

These exist to demo the command surface ahead of the backend
(`SwitchoverEndpointService` in `cc-switchover` is still template/stub code).

### Global flags

- `-o json` / `-o yaml` for machine-readable output (default is a human table).
- `--environment env-xxxxx` to override the current context's default environment.
- `--context <name>` to target a specific CLI context/profile.

## Rebuilding after further code changes

Just rerun `make build` from `cli/` — it picks up any local edits. No need to
redo the SDK pinning steps unless `ccloud-sdk-go-v2-internal`'s switchover
module actually changes upstream.
