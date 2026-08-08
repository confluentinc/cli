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
├── cli/
├── ccloud-sdk-go-v2-internal/
├── api/
├── cc-switchover/
└── terraform-provider-confluent/
```

- `cli/` — confluentinc/cli, branch `switchover-cli-local-demo`,
  [PR #3399](https://github.com/confluentinc/cli/pull/3399)
- `ccloud-sdk-go-v2-internal/` — confluentinc/ccloud-sdk-go-v2-internal,
  branch `ORC-10130-sync-switchover-sdk`,
  [PR #832](https://github.com/confluentinc/ccloud-sdk-go-v2-internal/pull/832)
- `api/` — confluentinc/api, branch `ORC-10130-switchover-api-spec`,
  [PR #2545](https://github.com/confluentinc/api/pull/2545)
- `cc-switchover/` — confluentinc/cc-switchover, branch `master`
  (reference/source-of-truth)
- `terraform-provider-confluent/` — confluentinc/terraform-provider-confluent,
  branch `master` (not used here)

`cli`'s `go.mod` pins the switchover SDK to
[PR #832](https://github.com/confluentinc/ccloud-sdk-go-v2-internal/pull/832)'s
commit:

```
github.com/confluentinc/ccloud-sdk-go-v2-internal/switchover v0.0.0-20260708221705-f2f8dbb55045
```

To get the `cli` branch locally:

```bash
git clone git@github.com:confluentinc/cli.git
cd cli
git checkout switchover-cli-local-demo
```

`confluentinc/cli` is a public repo with an org ruleset that blocks direct
branch pushes — use `git push-external` (Confluent's Airlock
proprietary-code-scan wrapper), not plain `git push`, to push further changes
to this branch.

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

Prefer a distinctly-named **alias** over adding the `dist/` folder to `PATH`,
so this experimental build doesn't shadow a real installed `confluent` CLI:

```bash
alias confluent-dev="/Users/ewang/sdk_cli_tf_switchover/cli/dist/confluent_darwin_arm64_v8.0/confluent"
```

Already set up in `~/.aliases` (sourced from `~/.zshrc`), alongside the
existing `confluent` alias. New terminals pick it up automatically; in an
already-open terminal run `source ~/.aliases` once. The rest of this README
uses `confluent-dev`.

## Usage

Log in first — the `switchover` command requires cloud login:

```bash
confluent-dev login
```

This hits **production** Confluent Cloud (`https://confluent.cloud`) with
real credentials — it's the CLI's stock login flow, not a mock.

### Logging in against staging (`stag`) instead of production

```bash
confluent-dev login --url https://stag.cpdev.cloud
```

The hostname is `stag.cpdev.cloud` with **no** `confluent.` prefix (unlike
prod's `confluent.cloud`). You'll need a **staging** Confluent Cloud
account — separate from your prod account — and your staging org needs
Switchover Early Access granted before `switchover pair` commands return
anything but an access/authorization error. `devel.cpdev.cloud` works the
same way for the `devel` environment. If you know your staging org ID, add
`--organization <org-id>`.

Everything below (`switchover pair`/`switchover endpoint` commands) works
identically once logged in, regardless of which environment you logged into.

### Bypassing the stag EA gate with a Cloud API key

The Switchover Early Access gate on `stag`/`devel` is only enforced on the
bearer-token (`login`) path — a Cloud API key (Basic auth) sails through it.
If your staging org doesn't have EA granted yet, set both of these and the
`switchover` commands will use Basic auth instead of your login session's
bearer token:

```bash
export CONFLUENT_CLOUD_API_KEY=<key>
export CONFLUENT_CLOUD_API_SECRET=<secret>
```

Create a Cloud API key first (still requires being logged in once, to create
it): `confluent-dev api-key create --resource cloud`. You still need
`confluent-dev login` for non-switchover commands / context; this env var
pair only affects the switchover API calls (see
`pkg/ccloudv2/switchover.go`'s `switchoverApiContext`).

### Switchover pairs

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

### Switchover endpoints

```bash
confluent-dev switchover endpoint create prod-kafka-dr-endpoint \
  --switchover-pair sw-123456 \
  --endpoint name=west-platt,type=private \
  --endpoint name=east-platt,type=private

confluent-dev switchover endpoint describe se-123456
confluent-dev switchover endpoint update se-123456 --display-name "renamed-endpoint"
confluent-dev switchover endpoint activate se-123456
```

### Global flags

- `-o json` / `-o yaml` for machine-readable output (default is a human table).
- `--environment env-xxxxx` to override the current context's default environment.
- `--context <name>` to target a specific CLI context/profile.

## Rebuilding after further code changes

Rerun `make build` from `cli/` — it picks up any local edits. Only redo the
SDK pinning step (`go get github.com/confluentinc/ccloud-sdk-go-v2-internal/switchover@<commit>`
in `cli/`) if `ccloud-sdk-go-v2-internal`'s switchover module changes again
upstream.
