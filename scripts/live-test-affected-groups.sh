#!/usr/bin/env bash
#
# Prints the live-test groups relevant to this PR's diff, as ONE line:
#   NONE          nothing with live coverage changed -> skip
#   all           a shared change -> run everything
#   kafka,flink   run just these groups
#
# It maps each changed command folder (internal/<dir>) to its live-test group
# (the table below). Shared code (pkg/, cmd/, go.mod) runs everything; folders
# with no live tests run nothing. You only touch the table when a brand-new
# live-test GROUP is added.
#
# Env: BASE_REF (default origin/main), CHANGED (override the diff, for tests).

set -euo pipefail
BASE_REF="${BASE_REF:-origin/main}"

# internal/<dir> -> its live-test group(s); empty = no live coverage.
group_for_dir() {
  case "$1" in
    kafka)                            echo kafka ;;
    schema-registry)                  echo schema_registry ;;
    connect)                          echo connect ;;
    flink)                            echo flink ;;
    iam)                              echo iam,core ;;   # rbac(iam) + service-account(core)
    login|logout)                     echo auth ;;
    api-key|environment|organization) echo core ;;
    rtce)                             echo rtce ;;
    *)                                echo "" ;;
  esac
}

changed="${CHANGED:-}"
if [ -z "$changed" ]; then
  git fetch -q origin "${BASE_REF#origin/}" 2>/dev/null || true
  base=$(git merge-base HEAD "$BASE_REF" 2>/dev/null || echo "$BASE_REF")
  changed=$(git diff --name-only "$base" HEAD 2>/dev/null || true)
fi
[ -n "${changed// }" ] || { echo NONE; exit 0; }

groups=""
add() {
  local x
  for x in $(echo "$1" | tr ',' ' '); do
    case ",$groups," in *",$x,"*) ;; *) groups="${groups:+$groups,}$x";; esac
  done
}

while IFS= read -r f; do
  [ -n "$f" ] || continue
  case "$f" in
    go.mod|go.sum|pkg/*|cmd/*) echo all; exit 0 ;;   # shared code -> everything
    internal/*)
      d=${f#internal/}; d=${d%%/*}
      add "$(group_for_dir "$d")" ;;
    *) : ;;                                          # docs, mock, ci -> ignore
  esac
done <<EOF
$changed
EOF

[ -n "$groups" ] && echo "$groups" || echo NONE
