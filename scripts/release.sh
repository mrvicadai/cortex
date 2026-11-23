#!/usr/bin/env bash
set -euo pipefail

# Release script for @mrvicadai/cortex.
# Usage:
#   ./scripts/release.sh [patch|minor|major|X.Y.Z] [--dry-run]
#
# Defaults to "patch" (auto-incrementing from the latest v* git tag).

usage() {
  cat <<'EOF'
Usage: ./scripts/release.sh [patch|minor|major|X.Y.Z] [--dry-run]

Bumps package.json and internal/version/version.go, runs tests,
commits the release, creates an annotated vX.Y.Z tag, and pushes
to origin/main (triggering .github/workflows/release.yml).

Arguments:
  patch       Increment patch from latest v* tag (default)
  minor       Increment minor and reset patch to 0
  major       Increment major and reset minor/patch to 0
  X.Y.Z       Set an explicit semver version (leading 'v' optional)
  --dry-run   Print the computed version without modifying files or git
EOF
}

BUMP="patch"
DRY_RUN=false

for arg in "$@"; do
  case "$arg" in
    -h|--help)
      usage
      exit 0
      ;;
    -n|--dry-run)
      DRY_RUN=true
      ;;
    *)
      if [[ "$BUMP" != "patch" && "$arg" != "patch" ]]; then
        echo "Error: unexpected argument '$arg'" >&2
        usage >&2
        exit 1
      fi
      BUMP="$arg"
      ;;
  esac
done

ROOT_DIR="$(git rev-parse --show-toplevel)"
cd "$ROOT_DIR"

if [[ "$DRY_RUN" == false ]]; then
  if [[ -n "$(git status --porcelain)" ]]; then
    echo "Error: working tree is not clean. Commit or stash changes first." >&2
    exit 1
  fi

  CURRENT_BRANCH="$(git rev-parse --abbrev-ref HEAD)"
  if [[ "$CURRENT_BRANCH" != "main" ]]; then
    echo "Error: releases must be cut from 'main' (currently on '$CURRENT_BRANCH')." >&2
    exit 1
  fi

  git fetch origin main --tags --quiet
  BEHIND_COUNT="$(git rev-list --count HEAD..origin/main)"
  if [[ "$BEHIND_COUNT" -gt 0 ]]; then
    echo "Error: local 'main' is behind 'origin/main' by $BEHIND_COUNT commit(s). Pull first." >&2
    exit 1
  fi
fi

LATEST_TAG="$(git tag --list 'v*' --sort=-v:refname | head -n 1)"
if [[ -n "$LATEST_TAG" ]]; then
  CURRENT_VERSION="${LATEST_TAG#v}"
else
  CURRENT_VERSION="$(node -p "require('./package.json').version")"
fi

if [[ ! "$CURRENT_VERSION" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
  echo "Error: current version '$CURRENT_VERSION' is not valid semver (X.Y.Z)." >&2
  exit 1
fi

MAJOR="${BASH_REMATCH[1]}"
MINOR="${BASH_REMATCH[2]}"
PATCH="${BASH_REMATCH[3]}"

case "$BUMP" in
  patch)
    NEXT_VERSION="${MAJOR}.${MINOR}.$((PATCH + 1))"
    ;;
  minor)
    NEXT_VERSION="${MAJOR}.$((MINOR + 1)).0"
    ;;
  major)
    NEXT_VERSION="$((MAJOR + 1)).0.0"
    ;;
  *)
    NEXT_VERSION="${BUMP#v}"
    if [[ ! "$NEXT_VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
      echo "Error: invalid version '$BUMP'. Expected patch, minor, major, or X.Y.Z." >&2
      exit 1
    fi
    ;;
esac

NEXT_TAG="v${NEXT_VERSION}"

if git rev-parse "$NEXT_TAG" >/dev/null 2>&1; then
  echo "Error: tag '$NEXT_TAG' already exists." >&2
  exit 1
fi

if [[ "$DRY_RUN" == true ]]; then
  echo "Dry run: current=${CURRENT_VERSION} -> next=${NEXT_VERSION} (tag ${NEXT_TAG})"
  exit 0
fi

echo "Releasing ${NEXT_TAG} (previous: v${CURRENT_VERSION})..."

node -e '
const fs = require("fs");
const pkgPath = "package.json";
const pkg = JSON.parse(fs.readFileSync(pkgPath, "utf8"));
pkg.version = process.argv[1];
fs.writeFileSync(pkgPath, JSON.stringify(pkg, null, 2) + "\n");
' "$NEXT_VERSION"

cat > internal/version/version.go <<EOF
package version

// Version is the semantic version of cortex (without a leading "v").
// It is updated by scripts/release.sh and can also be overridden at link time
// via -ldflags "-X github.com/mrvicadai/cortex/internal/version.Version=X.Y.Z".
var Version = "${NEXT_VERSION}"
EOF

echo "Running go test ./..."
go test ./...

git add package.json internal/version/version.go
git commit -m "chore: release ${NEXT_TAG}"
git tag -a "${NEXT_TAG}" -m "Release ${NEXT_TAG}"
git push origin main --follow-tags

echo "Released ${NEXT_TAG}! GitHub Actions will build and publish the release."
