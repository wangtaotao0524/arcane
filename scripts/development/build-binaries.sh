#!/usr/bin/env bash
set -euo pipefail

cd backend
mkdir -p .bin

# Read version from repo root .version (trim whitespace), fallback to "dev"
VERSION=$(cat ../.version | sed 's/^\s*\|\s*$//g')
LDFLAGS="-w -s -buildid=${VERSION}"

DOCKER_ONLY=false
for arg in "${@:-}"; do
  if [ "$arg" = "--docker" ]; then
    DOCKER_ONLY=true
    break
  fi
done

build_platform() {
  local target="$1" os="$2" arch="$3" arm_version="${4:-}"

  local ext=""
  [ "$os" = "windows" ] && ext=".exe"
  local output_path=".bin/arcane-${target}${ext}"

  printf "Building %s/%s%s ... " "$os" "$arch" "${arm_version:+ GOARM=$arm_version}"

  if [ -n "$arm_version" ]; then
    GOARM="$arm_version" CGO_ENABLED=1 GOOS="$os" GOARCH="$arch" \
      go build -ldflags="$LDFLAGS" -trimpath -o "$output_path" ./cmd/main.go
  else
    CGO_ENABLED=1 GOOS="$os" GOARCH="$arch" \
      go build -ldflags="$LDFLAGS" -trimpath -o "$output_path" ./cmd/main.go
  fi

  echo "Done"
}

if [ "$DOCKER_ONLY" = true ]; then
  echo "Building for Docker platforms only (linux/amd64, linux/arm64)..."
  build_platform "linux-amd64" "linux" "amd64"
  build_platform "linux-arm64" "linux" "arm64"
else
  echo "Version: ${VERSION}"
  echo "Building for all platforms..."
  build_platform "linux-amd64" "linux" "amd64"
  build_platform "linux-386"   "linux" "386"
  build_platform "linux-arm64" "linux" "arm64"
  build_platform "linux-armv7" "linux" "arm" "7"
  build_platform "macos-x64"   "darwin" "amd64"
  build_platform "macos-arm64" "darwin" "arm64"
fi

echo "Compilation done"