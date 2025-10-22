#!/usr/bin/env bash
set -euo pipefail

# Build script for "next" images on main branch
# Usage: ./scripts/build-docker-next.sh [TAG] [PULL]
# Example: ./scripts/build-docker-next.sh arcane:next --pull

# Get the current git revision (short commit hash)
REVISION=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# For next builds, version is always "next"
VERSION="next"

# Parse optional arguments
TAG="${1:-arcane:next}"
PULL="${2:---pull}"

echo "Building Docker image: ${TAG}"
echo "  VERSION: ${VERSION}"
echo "  REVISION: ${REVISION}"
echo ""

docker build ${PULL} --rm \
  -f 'docker/Dockerfile' \
  --build-arg VERSION="${VERSION}" \
  --build-arg REVISION="${REVISION}" \
  -t "${TAG}" \
  .

echo ""
echo "✓ Build complete: ${TAG}"
