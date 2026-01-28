#!/usr/bin/env sh
set -eu

# Optional: include build tags (e.g., tools) so tagged-only deps are considered by tidy/test.
# Usage: GO_UPDATE_TAGS=tools ./update-go-deps.sh
if [ -n "${GO_UPDATE_TAGS:-}" ]; then
  if [ -n "${GOFLAGS:-}" ]; then
    GOFLAGS="${GOFLAGS} -tags=${GO_UPDATE_TAGS}"
  else
    GOFLAGS="-tags=${GO_UPDATE_TAGS}"
  fi
  export GOFLAGS
fi

# Require jq
if ! command -v jq >/dev/null 2>&1; then
  echo "error: jq is required" >&2
  exit 1
fi

# Ensure we're in a module
modfile="$(go env GOMOD)"
if [ -z "$modfile" ] || [ "$modfile" = "/dev/null" ]; then
  echo "error: not in a Go module (go env GOMOD is /dev/null)" >&2
  exit 1
fi
cd "$(dirname "$modfile")"

## Refuse to run on a dirty git tree (if inside git)
#if command -v git >/dev/null 2>&1 && git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
#  if [ -n "$(git status --porcelain)" ]; then
#    echo "error: git working tree is not clean; commit/stash first" >&2
#    exit 1
#  fi
#fi

# Backup go.mod/go.sum
backup_dir=".depupdate-backup-$(date +%Y%m%d%H%M%S)"
mkdir -p "$backup_dir"
cp -f go.mod "$backup_dir/go.mod"
if [ -f go.sum ]; then
  cp -f go.sum "$backup_dir/go.sum"
fi

# Avoid pipefail dependency: write JSON to a file first.
tmpjson="$(mktemp)"
mods_file="$(mktemp)"
cleanup() { rm -f "$tmpjson" "$mods_file"; }
trap cleanup EXIT INT HUP TERM

echo "Reading requirements from go.mod..."
go mod edit -json > "$tmpjson"

jq -r '
  .Module.Path as $main
  | (.Require // [])
  | map(.Path)
  | map(select(. != $main))
  | unique
  | .[]
  | "\(.)@latest"
' "$tmpjson" > "$mods_file"

count="$(wc -l < "$mods_file" | tr -d " ")"
echo "Updating $count modules to @latest..."

# Batch to avoid command-length limits; -t includes test deps during resolution.
# (BusyBox xargs supports -n)
xargs -n 25 go get -t < "$mods_file"

echo "Tidying and verifying..."
go mod tidy
go mod verify

echo "Running tests..."
go test ./...

echo "Done. Backup saved at: $backup_dir"
if command -v git >/dev/null 2>&1 && git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  git diff --stat
fi
