#!/usr/bin/env zsh

set -euo pipefail
set -o pipefail

# set git pager to cat instead of less
export GIT_PAGER=cat
export PAGER=cat

print_err() { printf "%s\n" "$*" >&2; }

usage() {
  cat <<'EOF'
Usage: create-release.zsh [--major|--minor|--patch] [--yes] [--dry-run]

Options:
  --major        Force a major version bump
  --minor        Force a minor version bump
  --patch        Force a patch version bump
  --yes          Skip confirmation prompt
  --dry-run      Show planned actions without applying changes
  -h, --help     Show this help

Logic (when no explicit bump flag is given):
  - If no previous tag exists: minor
  - If commits contain feat: minor
  - Else if commits contain fix: patch
  - Else abort (no release)

Examples:
  ./create-release.zsh
  ./create-release.zsh --major --yes
  ./create-release.zsh --dry-run

EOF
}

# ---------- Parse arguments ----------
FORCE_BUMP=""
AUTO_YES=false
DRY_RUN=false

for arg in "$@"; do
  case "$arg" in
    --major|--minor|--patch)
      FORCE_BUMP="${arg#--}"
      ;;
    --yes)
      AUTO_YES=true
      ;;
    --dry-run)
      DRY_RUN=true
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      print_err "Unknown argument: $arg"
      usage
      exit 1
      ;;
  esac
done

# ---------- Project root checks ----------
if [[ ! -f .version || ! -f frontend/package.json || ! -f CHANGELOG.md ]]; then
  print_err "Error: Must run from project root (need .version, frontend/package.json, CHANGELOG.md)"
  exit 1
fi

# ---------- Tool checks ----------
if ! command -v conventional-changelog &>/dev/null; then
  echo "conventional-changelog not found, installing..."
  if ! npm install -g conventional-changelog-cli; then
    print_err "Error: Failed to install conventional-changelog-cli."
    exit 1
  fi
fi

if ! command -v gh &>/dev/null; then
  print_err "Error: GitHub CLI (gh) not installed or not in PATH. Run: gh auth login"
  exit 1
fi

if ! command -v jq &>/dev/null; then
  print_err "Error: jq is required."
  exit 1
fi

# ---------- Git state ----------
git fetch --tags --quiet || true

current_branch=$(git rev-parse --abbrev-ref HEAD)
if [[ "$current_branch" != "main" ]]; then
  print_err "Error: Must be on main branch (current: $current_branch)."
  exit 1
fi

VERSION=$(<.version)

increment_version() {
  local version part
  version=$1
  part=$2
  local -a parts
  IFS='.' read -rA parts <<<"$version"
  case "$part" in
    major)
      parts[0]=$((parts[0] + 1))
      parts[1]=0
      parts[2]=0
      ;;
    minor)
      parts[1]=$((parts[1] + 1))
      parts[2]=0
      ;;
    patch)
      parts[2]=$((parts[2] + 1))
      ;;
    *)
      print_err "Unknown bump part: $part"
      return 1
      ;;
  esac
  echo "${parts[0]}.${parts[1]}.${parts[2]}"
}

# ---------- Determine release type ----------
RELEASE_TYPE=""
if [[ -n "$FORCE_BUMP" ]]; then
  RELEASE_TYPE="$FORCE_BUMP"
else
  LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
  if [[ -z "$LATEST_TAG" ]]; then
    RELEASE_TYPE="minor"
  else
    SUBJECTS=$(git log --no-merges --format=%s "${LATEST_TAG}..HEAD")
    if echo "$SUBJECTS" | grep -Eiq '^feat(\([^)]+\))?: '; then
      RELEASE_TYPE="minor"
    elif echo "$SUBJECTS" | grep -Eiq '^fix(\([^)]+\))?: '; then
      RELEASE_TYPE="patch"
    else
      echo "No 'fix' or 'feat' commits since ${LATEST_TAG}. No release created."
      echo "Commits since ${LATEST_TAG}:"
      git log --oneline --no-merges "${LATEST_TAG}..HEAD" || true
      exit 0
    fi
  fi
fi

# ---------- Compute new version ----------
NEW_VERSION=$(increment_version "$VERSION" "$RELEASE_TYPE") || exit 1
echo "Proposed ${RELEASE_TYPE} release: $VERSION -> $NEW_VERSION"

if ! $AUTO_YES; then
  if $DRY_RUN; then
    echo "(dry-run) Skipping confirmation prompt."
  else
    read "?Proceed with release v$NEW_VERSION? (y/n) " CONFIRM
    if [[ "$CONFIRM" != "y" ]]; then
      echo "Aborted."
      exit 1
    fi
  fi
fi

# ---------- Dry run early exit path ----------
if $DRY_RUN; then
  # Generate prospective release notes by temporarily bumping package.json
  DRY_RELEASE_NOTES=""
  ORIGINAL_PKG_VERSION=$(jq -r '.version' frontend/package.json || echo "")
  if command -v conventional-changelog &>/dev/null; then
    TMP_UPDATED=false
    if [[ -n "$ORIGINAL_PKG_VERSION" && "$ORIGINAL_PKG_VERSION" != "$NEW_VERSION" ]]; then
      jq --arg new_version "$NEW_VERSION" '.version = $new_version' frontend/package.json > frontend/package_tmp.json
      mv frontend/package_tmp.json frontend/package.json
      TMP_UPDATED=true
    fi

    GENERATED_CHANGELOG=$(conventional-changelog -p conventionalcommits -u -r 0 2>/dev/null || echo "")
    if $TMP_UPDATED; then
      # Revert package.json
      jq --arg orig "$ORIGINAL_PKG_VERSION" '.version = $orig' frontend/package.json > frontend/package_tmp.json || true
      mv frontend/package_tmp.json frontend/package.json || true
    fi
    if [[ -n "$GENERATED_CHANGELOG" ]]; then
      # Extract first section body (skip header line(s) & blank separation similar to post-release logic)
      DRY_RELEASE_NOTES=$(echo "$GENERATED_CHANGELOG" | awk '/^## / {if (NR>1) exit} NR>1 {print}' | awk 'NR>2 || NF {print}')
    fi
  fi
  if [[ -z "$DRY_RELEASE_NOTES" ]]; then
    DRY_RELEASE_NOTES="(No release notes content could be generated – likely no qualifying commits.)"
  fi

  cat <<EOF
Dry run summary:
  Release type : $RELEASE_TYPE
  Old version  : $VERSION
  New version  : $NEW_VERSION
  Actions that would occur:
    - Update .version
    - Update frontend/package.json
    - Update .revision
    - Update ARG VERSION/REVISION in Dockerfiles
    - Regenerate CHANGELOG.md
    - Commit, tag v$NEW_VERSION, push
    - Create draft GitHub release with extracted changelog section

Planned release notes (preview):
$DRY_RELEASE_NOTES
EOF
  exit 0
fi

# ---------- Apply changes ----------
echo "$NEW_VERSION" > .version
git add .version

jq --arg new_version "$NEW_VERSION" '.version = $new_version' frontend/package.json > frontend/package_tmp.json
mv frontend/package_tmp.json frontend/package.json
git add frontend/package.json

LATEST_REVISION=$(git rev-parse --short HEAD)
echo "$LATEST_REVISION" > .revision
git add .revision

if [[ -f docker/Dockerfile ]]; then
  echo "Updating docker/Dockerfile ARGs..."
  sed -i.bak -e "s/^ARG VERSION=.*/ARG VERSION=\"${NEW_VERSION}\"/" \
             -e "s/^ARG REVISION=.*/ARG REVISION=\"${LATEST_REVISION}\"/" docker/Dockerfile
  rm docker/Dockerfile.bak
  git add docker/Dockerfile
fi

if [[ -f docker/Dockerfile-agent ]]; then
  echo "Updating docker/Dockerfile-agent ARGs..."
  sed -i.bak -e "s/^ARG VERSION=.*/ARG VERSION=\"${NEW_VERSION}\"/" \
             -e "s/^ARG REVISION=.*/ARG REVISION=\"${LATEST_REVISION}\"/" docker/Dockerfile-agent
  rm docker/Dockerfile-agent.bak
  git add docker/Dockerfile-agent
fi

echo "Regenerating CHANGELOG.md..."
conventional-changelog -p conventionalcommits -i CHANGELOG.md -s
git add CHANGELOG.md

git commit -m "release: ${NEW_VERSION}"
git tag "v${NEW_VERSION}"
git push
git push --tags

echo "Extracting changelog section for v${NEW_VERSION}..."
# Extract newest section
CHANGELOG=$(awk '/^## / {if (NR>1) exit} NR>1 {print}' CHANGELOG.md | awk 'NR>2 || NF {print}')

if [[ -z "$CHANGELOG" ]]; then
  print_err "Error: Could not extract changelog content for v${NEW_VERSION}"
  exit 1
fi

echo "Creating draft GitHub release..."
if gh release create "v${NEW_VERSION}" --title "v${NEW_VERSION}" --notes "$CHANGELOG" --draft; then
  echo "Draft release created."
else
  print_err "Failed to create GitHub release."
  exit 1
fi

echo "Release complete: v${NEW_VERSION}"