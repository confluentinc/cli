#!/usr/bin/env bash
#
# scripts/live-test-affected-groups.sh
#
# Works out which live tests are relevant to this branch's diff, so a PR can run
# just those instead of the whole nightly suite.
#
# It prints TWO lines to stdout:
#   GROUPS=<NONE | all | comma-list>   the build-tag groups to compile
#   RUN=<regex | empty>                a -run regex naming the exact suite methods,
#                                      or empty to run every Live test in GROUPS
#
# CLI live tests are testify suite methods (`func (s *CLILiveTestSuite)
# TestKafkaTopicCRUDLive()`) run as subtests of the single TestLive, and they are
# NOT co-located with command code (internal/<domain>/). So we:
#   * map each changed internal/<dir> to its live-test group(s) via the DIR table;
#   * within those groups, narrow to the test/live file(s) whose name shares a
#     token with the changed command file, and ask scripts/livetestfuncs for the
#     exact suite methods they declare -> run precisely those
#     (-run '^TestLive$/^(TestKafkaTopicCRUDLive|...)$');
#   * fall back to the whole group when nothing narrows, and to "all" for shared
#     code (pkg/, cmd/, go.mod). It never under-runs; nightly stays the backstop.
#
# Keep the DIR table in sync when a new live-test group is added.
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
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

log() { echo "[affected-groups] $*" >&2; }

extract_groups() {
  sed -E 's#//go:build##; s#[&|()!]# #g' | tr ' ' '\n' | grep -vE '^(live_test|all)?$' || true
}

# internal/<dir> -> space-separated live-test group(s); empty = no live coverage.
dir_to_groups() {
  case "$1" in
    kafka)                            echo kafka ;;
    schema-registry)                  echo schema_registry ;;
    connect)                          echo connect ;;
    flink)                            echo flink ;;
    iam)                              echo iam core ;;   # rbac(iam) + service-account(core)
    login|logout)                     echo auth ;;
    api-key|environment|organization) echo core ;;
    rtce)                             echo rtce ;;
    *)                                echo "" ;;
  esac
}

# space-separated group(s) of a test/live file, or "" if it has no build tag.
file_groups() {
  local tag
  tag=$(grep -m1 '//go:build' "$1" 2>/dev/null || true)
  [ -n "$tag" ] || { echo ""; return; }
  printf '%s\n' "$tag" | extract_groups | tr '\n' ' '
}

# tokens (>=3 chars) of a changed command file's basename, minus the command_ prefix.
tokens_of() {
  local b
  b=$(basename "$1"); b=${b%.go}; b=${b#command_}
  echo "$b" | tr '_' ' '
}

# test/live files in the given groups whose stem shares a token with the change.
match_test_files() {
  local wantgroups="$1" tokens="$2" lf g stem tok wg hit matched=""
  for lf in "$LIVE_DIR"/*_live_test.go; do
    [ -f "$lf" ] || continue
    g=$(file_groups "$lf")
    hit=0
    for wg in $wantgroups; do
      if [[ " $g " == *" $wg "* ]]; then hit=1; fi
    done
    [ "$hit" = "1" ] || continue
    stem=$(basename "$lf"); stem=${stem%_live_test.go}
    for tok in $tokens; do
      [ "${#tok}" -ge 3 ] || continue
      if [[ "$stem" == *"$tok"* ]]; then matched="${matched:+$matched }$lf"; break; fi
    done
  done
  echo "$matched"
}

emit() { echo "GROUPS=$1"; echo "RUN=${2:-}"; }

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
  log "no changed files vs $BASE_REF"; emit "NONE"; exit 0
fi
log "changed files:"; printf '%s\n' "$CHANGED" | sed 's/^/  /' >&2

# --- walk the diff ----------------------------------------------------------
# NB: not named GROUPS -- that is a reserved bash variable.
GRPS=""
FILES=""
BLAST=0
GROUP_LEVEL=0

add_grp() {
  local g
  for g in $1; do
    if [[ ",$GRPS," != *",$g,"* ]]; then GRPS="${GRPS:+$GRPS,}$g"; fi
  done
}
add_file() {
  if [[ " $FILES " != *" $1 "* ]]; then FILES="${FILES:+$FILES }$1"; fi
}

while IFS= read -r f; do
  [ -z "$f" ] && continue
  case "$f" in
    go.mod|go.sum) log "$f -> dependency bump -> blast radius"; BLAST=1; continue;;
    pkg/*|cmd/*)   log "$f -> shared code -> blast radius"; BLAST=1; continue;;
    "$LIVE_DIR"/*_live_test.go)
      g=$(file_groups "$f")
      if [ -n "${g// }" ]; then add_grp "$g"; add_file "$f"; log "$f -> live file changed -> precise"
      else log "$f -> untagged live file -> blast"; BLAST=1; fi
      continue;;
    internal/*) ;;
    *) continue;;
  esac

  d=${f#internal/}; d=${d%%/*}
  wg=$(dir_to_groups "$d")
  if [ -z "$wg" ]; then log "$f -> internal/$d has no live coverage -> ignored"; continue; fi
  mf=$(match_test_files "$wg" "$(tokens_of "$f")")
  if [ -n "${mf// }" ]; then
    for lf in $mf; do add_file "$lf"; add_grp "$(file_groups "$lf")"; done
    log "$f -> internal/$d -> matched: $mf"
  else
    add_grp "$wg"; GROUP_LEVEL=1
    log "$f -> internal/$d -> no token match -> whole group(s): $wg"
  fi
done < <(printf '%s\n' "$CHANGED")

# --- decide -----------------------------------------------------------------
if [ "$BLAST" = "1" ]; then emit "$BLAST_RADIUS_GROUPS"; exit 0; fi
if [ -z "$GRPS" ]; then emit "NONE"; exit 0; fi
if [ "$GROUP_LEVEL" = "1" ] || [ -z "${FILES// }" ]; then
  log "at least one change is group-level; running whole group(s): $GRPS"
  emit "$GRPS"; exit 0
fi

NAMES=$(cd "$SCRIPT_DIR/.." && go run ./scripts/livetestfuncs $FILES 2>/dev/null | sort -u || true)
if [ -z "${NAMES// }" ]; then
  log "could not extract suite methods (falling back to whole group(s): $GRPS)"
  emit "$GRPS"; exit 0
fi
RUN="^TestLive\$/^($(printf '%s' "$NAMES" | paste -sd'|' -))\$"
log "precise run for groups [$GRPS]: $RUN"
emit "$GRPS" "$RUN"
