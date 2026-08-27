#!/usr/bin/env bash
#
# scripts/live-test-affected-groups.sh
#
# Prints the CLI_LIVE_TEST_GROUPS value for the live tests relevant to this
# branch's diff against the base branch.
#
# Unlike the provider, CLI live tests (test/live/*_live_test.go) are NOT
# co-located with command code (internal/<domain>/), so there is no filename
# stem to match on. Instead we map each changed top-level internal/<dir> to its
# live-test group(s) via the explicit DIR table below. Keep that table in sync
# whenever a new live-test group is added (the group vocabulary itself lives in
# the //go:build tags of test/live/*_live_test.go).
#
# Output on stdout is EXACTLY one line, one of:
#   NONE            nothing with live coverage changed -> caller should skip
#   all             a shared / blast-radius file changed -> run everything
#   kafka,core      comma-separated affected groups
#
# Diagnostics go to stderr so stdout stays machine-parseable.
#
# Env knobs:
#   BASE_REF               base to diff against            (default: origin/main)
#   LIVE_DIR               live-test dir                   (default: test/live)
#   BLAST_RADIUS_GROUPS    what to run for a shared change (default: all)
#   CHANGED_FILES_OVERRIDE newline/space list, bypasses git (for tests/dry-runs)

set -euo pipefail

BASE_REF="${BASE_REF:-origin/main}"
LIVE_DIR="${LIVE_DIR:-test/live}"
BLAST_RADIUS_GROUPS="${BLAST_RADIUS_GROUPS:-all}"

log() { echo "[affected-groups] $*" >&2; }

# internal/<dir> -> space-separated live-test group(s).
# Dirs not listed have no live coverage, so changes to them alone don't trigger a run.
dir_to_groups() {
  case "$1" in
    kafka)                                    echo kafka ;;
    schema-registry)                          echo schema_registry ;;
    connect)                                  echo connect ;;
    flink)                                    echo flink ;;
    iam)                                      echo iam core ;;   # rbac(iam) + service-account(core) both live here
    login|logout)                             echo auth ;;
    api-key|environment|organization)         echo core ;;
    rtce)                                     echo rtce ;;
    *)                                        echo "" ;;          # no live coverage
  esac
}

# "//go:build live_test && (all || kafka)" -> lines: kafka
extract_groups() {
  sed -E 's#//go:build##; s#[&|()!]# #g' \
    | tr ' ' '\n' \
    | grep -vE '^(live_test|all)?$' || true
}

# --- changed files ----------------------------------------------------------
if [ -n "${CHANGED_FILES_OVERRIDE:-}" ]; then
  CHANGED=$(printf '%s\n' $CHANGED_FILES_OVERRIDE)
else
  git fetch -q origin "${BASE_REF#origin/}" 2>/dev/null || true
  if MB=$(git merge-base HEAD "$BASE_REF" 2>/dev/null); then
    CHANGED=$(git diff --name-only "$MB" HEAD || true)
  else
    log "no merge-base with $BASE_REF (shallow clone?); diffing against its tip"
    CHANGED=$(git diff --name-only "$BASE_REF" HEAD || true)
  fi
fi

if [ -z "${CHANGED// }" ]; then
  log "no changed files vs $BASE_REF"; echo "NONE"; exit 0
fi
log "changed files:"; printf '%s\n' "$CHANGED" | sed 's/^/  /' >&2

# --- walk the diff ----------------------------------------------------------
RESULT=""; BLAST=0
add_group() {
  local g
  for g in $1; do
    case ",$RESULT," in *",$g,"*) ;; *) RESULT="${RESULT:+$RESULT,}$g";; esac
  done
}

while IFS= read -r f; do
  [ -z "$f" ] && continue
  case "$f" in
    go.mod|go.sum) log "$f -> dependency bump -> blast radius"; BLAST=1;;
    pkg/*|cmd/*)   log "$f -> shared code -> blast radius"; BLAST=1;;
    "$LIVE_DIR"/*_live_test.go)
      tag=$(grep -m1 '//go:build' "$f" 2>/dev/null || true)
      g=$(printf '%s\n' "$tag" | extract_groups | tr '\n' ' ')
      if [ -z "${g// }" ]; then log "$f -> group-less live helper -> blast radius"; BLAST=1
      else log "$f -> groups:$g"; add_group "$g"; fi ;;
    internal/*)
      d=${f#internal/}; d=${d%%/*}
      g=$(dir_to_groups "$d")
      if [ -z "$g" ]; then log "$f -> internal/$d has no live coverage -> ignored"
      else log "$f -> internal/$d -> groups:$g"; add_group "$g"; fi ;;
    *) : ;;  # docs/, mock/, packaging/, ci config, etc. don't gate live tests
  esac
done < <(printf '%s\n' "$CHANGED")

# --- decide -----------------------------------------------------------------
if [ "$BLAST" = "1" ]; then echo "$BLAST_RADIUS_GROUPS"; exit 0; fi
if [ -n "$RESULT" ]; then echo "$RESULT"; exit 0; fi
echo "NONE"
