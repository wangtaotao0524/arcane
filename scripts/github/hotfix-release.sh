#!/usr/bin/env bash
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if the script is being run from the root of the project
if [ ! -f .version ] || [ ! -f frontend/package.json ] || [ ! -f CHANGELOG.md ]; then
    echo -e "${RED}Error: This script must be run from the root of the project.${NC}"
    exit 1
fi

# Check if git cliff is installed
if ! command -v git cliff &>/dev/null; then
    echo "Error: git cliff is not installed. Please install it from https://git-cliff.org/docs/installation."
    exit 1
fi

# Check if GitHub CLI is installed
if ! command -v gh &>/dev/null; then
    echo -e "${RED}Error: GitHub CLI (gh) is not installed. Please install it and authenticate using 'gh auth login'.${NC}"
    exit 1
fi

# Ensure local tags are up to date
git fetch --tags --quiet || true

# Get the latest release tag
LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
if [ -z "$LATEST_TAG" ]; then
    echo -e "${RED}Error: No previous release tag found.${NC}"
    exit 1
fi

echo -e "${GREEN}Latest release tag: ${LATEST_TAG}${NC}"

# Extract version components from tag (format: v1.2.3)
VERSION=${LATEST_TAG#v}
IFS='.' read -r MAJOR MINOR PATCH <<< "$VERSION"

# Increment patch version for hotfix
NEW_PATCH=$((PATCH + 1))
NEW_VERSION="${MAJOR}.${MINOR}.${NEW_PATCH}"
NEW_TAG="v${NEW_VERSION}"

echo -e "${YELLOW}Proposed hotfix version: ${NEW_TAG}${NC}"

# Create release branch name
RELEASE_BRANCH="release/v${MAJOR}.${MINOR}"

# Check if release branch already exists
if git show-ref --verify --quiet "refs/heads/${RELEASE_BRANCH}"; then
    echo -e "${YELLOW}Release branch ${RELEASE_BRANCH} already exists.${NC}"
    read -p "Do you want to use the existing branch? (y/n) " USE_EXISTING
    if [[ "$USE_EXISTING" != "y" ]]; then
        echo -e "${RED}Aborting.${NC}"
        exit 1
    fi
    git checkout "${RELEASE_BRANCH}"
else
    echo -e "${BLUE}Creating new release branch from ${LATEST_TAG}...${NC}"
    git checkout -b "${RELEASE_BRANCH}" "${LATEST_TAG}"
fi

echo ""
echo -e "${YELLOW}=== Available fix commits since ${LATEST_TAG} ===${NC}"
echo ""

# List all fix commits since the last tag
FIX_COMMITS=$(git log "${LATEST_TAG}..main" \
    --oneline \
    --no-merges \
    --grep="^fix:" \
    --grep="^hotfix:" \
    --regexp-ignore-case \
    --pretty=format:"%C(yellow)%h%Creset %s %C(dim)(%an)%Creset" || echo "")

if [ -z "$FIX_COMMITS" ]; then
    echo -e "${RED}No fix commits found since ${LATEST_TAG}${NC}"
    echo ""
    read -p "Do you want to continue anyway? (y/n) " CONTINUE
    if [[ "$CONTINUE" != "y" ]]; then
        git checkout main
        git branch -D "${RELEASE_BRANCH}" 2>/dev/null || true
        exit 0
    fi
else
    echo "$FIX_COMMITS"
    echo ""
fi

echo -e "${BLUE}=== Cherry-pick Instructions ===${NC}"
echo ""
echo "You can now cherry-pick the commits you want to include in this hotfix."
echo ""
echo "Example:"
echo -e "  ${GREEN}git cherry-pick abc123 def456${NC}"
echo ""
echo "Or pick them one at a time and continue when done."
echo ""

# Interactive cherry-pick loop
while true; do
    echo ""
    read -p "Enter commit hash to cherry-pick (or 'done' to finish, 'list' to see commits again, 'quit' to abort): " INPUT
    
    case "$INPUT" in
        done)
            break
            ;;
        quit)
            echo -e "${YELLOW}Aborting hotfix release...${NC}"
            git checkout main
            read -p "Delete release branch ${RELEASE_BRANCH}? (y/n) " DELETE_BRANCH
            if [[ "$DELETE_BRANCH" == "y" ]]; then
                git branch -D "${RELEASE_BRANCH}" 2>/dev/null || true
            fi
            exit 0
            ;;
        list)
            echo ""
            echo -e "${YELLOW}=== Available fix commits ===${NC}"
            echo "$FIX_COMMITS"
            continue
            ;;
        "")
            continue
            ;;
        *)
            # Try to cherry-pick the commit(s)
            if git cherry-pick $INPUT; then
                echo -e "${GREEN}✓ Successfully cherry-picked: $INPUT${NC}"
            else
                echo -e "${RED}✗ Cherry-pick failed. Resolve conflicts and run:${NC}"
                echo -e "  ${YELLOW}git add <resolved-files>${NC}"
                echo -e "  ${YELLOW}git cherry-pick --continue${NC}"
                echo ""
                read -p "Press Enter when resolved (or 'abort' to skip this commit): " RESOLVE
                if [[ "$RESOLVE" == "abort" ]]; then
                    git cherry-pick --abort
                    echo -e "${YELLOW}Cherry-pick aborted${NC}"
                fi
            fi
            ;;
    esac
done

# Check if any commits were added
BASE_TAG=$(git describe --tags --abbrev=0 "${RELEASE_BRANCH}^" 2>/dev/null || git describe --tags --abbrev=0 2>/dev/null)
COMMITS_ADDED=$(git rev-list "${BASE_TAG}..HEAD" --count)

if [ "$COMMITS_ADDED" -eq 0 ]; then
    echo -e "${RED}No commits were cherry-picked. Aborting hotfix release.${NC}"
    git checkout main
    git branch -D "${RELEASE_BRANCH}" 2>/dev/null || true
    exit 0
fi

echo ""
echo -e "${YELLOW}=== Commits in this hotfix release ===${NC}"
git log "${BASE_TAG}..HEAD" --oneline --no-merges
echo ""

# Confirm release
read -p "Create hotfix release ${NEW_TAG}? (y/n) " CONFIRM
if [[ "$CONFIRM" != "y" ]]; then
    echo -e "${RED}Release canceled.${NC}"
    echo ""
    echo "You are still on branch ${RELEASE_BRANCH}."
    echo "To continue later, run this script again or finalize manually with:"
    echo -e "  ${GREEN}./scripts/github/finalize-hotfix-release.sh ${NEW_TAG}${NC}"
    exit 1
fi

echo ""
echo -e "${BLUE}Finalizing hotfix release ${NEW_TAG}...${NC}"

# Update .version file
echo "$NEW_VERSION" > .version
git add .version

# Update version in frontend/package.json
jq --arg new_version "$NEW_VERSION" '.version = $new_version' frontend/package.json > frontend/package_tmp.json && mv frontend/package_tmp.json frontend/package.json
git add frontend/package.json

# Create/Update .revision file with the latest commit short hash
LATEST_REVISION=$(git rev-parse --short HEAD)
echo "$LATEST_REVISION" > .revision
git add .revision

# Generate changelog for ONLY the fixes in this release branch
echo -e "${BLUE}Generating changelog for hotfix...${NC}"

git-cliff \
    --github-token=$(gh auth token) \
    --prepend CHANGELOG.md \
    --tag "$NEW_TAG" \
    --unreleased

git add CHANGELOG.md

# Commit the version bump and changelog
git commit -m "release: ${NEW_VERSION} (hotfix)

Hotfix release containing critical bug fixes.
Base version: ${BASE_TAG}

See CHANGELOG.md for details."

# Create annotated tag
git tag -a "$NEW_TAG" -m "Release ${NEW_TAG} (Hotfix)

Hotfix release based on ${BASE_TAG}

See CHANGELOG.md for details."

echo ""
echo -e "${GREEN}✅ Hotfix release ${NEW_TAG} created successfully!${NC}"
echo ""
echo -e "${YELLOW}=== Next Steps ===${NC}"
echo ""
echo "1. Review the changes:"
echo -e "   ${BLUE}git show HEAD${NC}"
echo -e "   ${BLUE}git log ${BASE_TAG}..${NEW_TAG}${NC}"
echo ""
echo "2. Push the release branch and tag:"
echo -e "   ${GREEN}git push origin ${RELEASE_BRANCH}${NC}"
echo -e "   ${GREEN}git push origin ${NEW_TAG}${NC}"
echo ""
echo "3. Create GitHub release:"
echo -e "   ${GREEN}gh release create ${NEW_TAG} --title '${NEW_TAG} (Hotfix)' --generate-notes --draft${NC}"
echo ""
echo "   Or with changelog:"
read -p "Create GitHub draft release now? (y/n) " CREATE_RELEASE
if [[ "$CREATE_RELEASE" == "y" ]]; then
    # Extract the changelog content for the latest release
    echo "Extracting changelog content for version $NEW_TAG..."
    CHANGELOG=$(awk '/^## v[0-9]/ { if (found) exit; found=1; next } found' CHANGELOG.md)
    
    if [ -z "$CHANGELOG" ]; then
        echo -e "${YELLOW}Warning: Could not extract changelog. Using generate-notes instead.${NC}"
        gh release create "$NEW_TAG" --title "${NEW_TAG} (Hotfix)" --generate-notes --draft
    else
        gh release create "$NEW_TAG" --title "${NEW_TAG} (Hotfix)" --notes "$CHANGELOG" --draft
    fi
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}GitHub draft release created successfully!${NC}"
        echo ""
        echo "Review the draft at: https://github.com/ofkm/arcane/releases"
    else
        echo -e "${RED}Failed to create GitHub release.${NC}"
    fi
else
    echo ""
    echo "Skipped GitHub release creation."
fi

echo ""
echo "4. (Optional) Merge fixes back to main:"
echo -e "   ${BLUE}git checkout main${NC}"
echo -e "   ${BLUE}git merge ${RELEASE_BRANCH} --no-ff -m 'chore: merge hotfix ${NEW_TAG} to main'${NC}"
echo -e "   ${BLUE}git push origin main${NC}"
echo ""
