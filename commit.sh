#!/usr/bin/env bash
# commit.sh — human-gated commit from the agent WIP log.
#
# Your old habit (git add .; git commit -am "msg") works: add stages new files
# in this directory, -a restages tracked edits. From a subdirectory it misses
# the rest of the repo; -am without add misses untracked files.
#
# This script always runs at the git root:
#   git add -A          # new, modified, deleted — whole tree
#   git commit -F log   # message = .agent/COMMITLOG (not -am)
#
# Usage: ./commit.sh   or:  commit <project>
#
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT"
LOG="${ROOT}/.agent/COMMITLOG"
EXAMPLE="${ROOT}/.agent/COMMITLOG.example"

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo >&2 "error: ${ROOT} is not a git repo"
  exit 1
fi

if [ ! -f "$LOG" ]; then
  mkdir -p "${ROOT}/.agent"
  if [ -f "$EXAMPLE" ]; then
    cp "$EXAMPLE" "$LOG"
  else
    printf '%s\n' "committed (empty — write why this commit exists)" >"$LOG"
  fi
  echo >&2 "created ${LOG} — write the commit message there, then re-run"
  exit 1
fi

if ! grep -qvE '^[[:space:]]*$|^committed( |$)' "$LOG"; then
  echo >&2 "error: ${LOG} has no new work (only a committed stamp or blank)"
  echo >&2 "agents: write the why-message into this file before commit"
  exit 1
fi

ask() {
  local prompt=$1
  local reply
  if [ ! -r /dev/tty ]; then
    echo >&2 "error: need a real terminal for y/n (don't run from a non-interactive agent)"
    exit 1
  fi
  printf '%s' "$prompt" >/dev/tty
  IFS= read -r reply </dev/tty
  case "$reply" in
    y|Y|yes|YES) return 0 ;;
    *) return 1 ;;
  esac
}

if [ -z "$(git status --porcelain)" ]; then
  echo >&2 "nothing to commit (working tree clean)"
  exit 1
fi

echo "Repo: ${ROOT}"
echo "Branch: $(git rev-parse --abbrev-ref HEAD)"
echo
echo "===== git status ====="
git status
echo
echo "===== git add -A (dry-run; whole tree, not cwd) ====="
git add -A -n
echo
echo "===== git diff --stat (unstaged tracked) ====="
git diff --stat
echo
echo "===== COMMITLOG (commit message) ====="
cat "$LOG"
echo "====================================="
echo
echo "Will run only if you confirm:"
echo "  git add -A"
echo "  git commit -F .agent/COMMITLOG"
echo

if ! ask "Commit this? [y/N] "; then
  echo "aborted"
  exit 0
fi

git add -A
git commit -F "$LOG"

stamp="$(date '+%Y-%m-%d %H:%M:%S %z')"
printf 'committed %s\n' "$stamp" >"$LOG"
echo "COMMITLOG reset to: committed ${stamp}"

if ! ask "Push to remote? [y/N] "; then
  echo "done (not pushed)"
  exit 0
fi

git push
echo "pushed"
