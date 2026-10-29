#!/usr/bin/env bash
set -euo pipefail

TAG="${1:-}"
OUT="${2:-build/release/notes.md}"
REPO="${REPO:-chenshiyi007/HttpRepeater}"

if [ -z "$TAG" ]; then
  echo "usage: $0 <tag> [output-file]" >&2
  exit 1
fi

if ! git rev-parse --git-dir >/dev/null 2>&1; then
  echo "not a git repository" >&2
  exit 1
fi

PREV="$(git describe --tags --abbrev=0 --match 'v*' "${TAG}^" 2>/dev/null || true)"
if [ "$PREV" = "$TAG" ]; then
  PREV=""
fi
if [ -z "$PREV" ]; then
  PREV="$(git describe --tags --abbrev=0 --match 'v*' 2>/dev/null || true)"
  if [ "$PREV" = "$TAG" ]; then
    PREV=""
  fi
fi

if [ -n "$PREV" ]; then
  CHANGES="$(git log --no-merges --pretty=format:"- %s (%h)" "${PREV}..HEAD")"
else
  CHANGES="$(git log --no-merges --pretty=format:"- %s (%h)")"
fi

mkdir -p "$(dirname "$OUT")"

{
  echo "## HttpRepeater ${TAG}"
  echo
  echo "Released $(date -u +%Y-%m-%d)"
  echo
  echo "### Changes"
  echo
  if [ -n "$CHANGES" ]; then
    printf '%s\n' "$CHANGES"
  else
    echo "- No code changes."
  fi
  echo
  if [ -n "$PREV" ]; then
    echo "**Full changelog:** https://github.com/${REPO}/compare/${PREV}...${TAG}"
  else
    echo "**Full changelog:** https://github.com/${REPO}/commits/${TAG}"
  fi
  echo
  echo "### Assets"
  echo
  echo "Download the archive for your platform below."
} > "$OUT"

echo "Wrote $OUT"
