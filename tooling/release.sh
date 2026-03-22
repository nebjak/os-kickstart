#!/bin/bash
set -euo pipefail

# Release script for OS Kickstart
# Usage: ./tooling/release.sh [version]
# Example: ./tooling/release.sh 1.1.0

VERSION="${1:-}"

if [[ -z "$VERSION" ]]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 1.1.0"
    echo ""
    echo "Current tags:"
    git tag --sort=-v:refname | head -5
    exit 1
fi

# Strip leading 'v' if provided
VERSION="${VERSION#v}"
TAG="v${VERSION}"

echo "==> Releasing ${TAG}"

# Ensure clean working tree
if [[ -n "$(git status --porcelain)" ]]; then
    echo "ERROR: Working tree is not clean. Commit or stash changes first."
    exit 1
fi

# Ensure on main branch
BRANCH=$(git branch --show-current)
if [[ "$BRANCH" != "main" ]]; then
    echo "ERROR: Must be on main branch (currently on ${BRANCH})"
    exit 1
fi

# Ensure up to date with remote
git fetch origin main
LOCAL=$(git rev-parse HEAD)
REMOTE=$(git rev-parse origin/main)
if [[ "$LOCAL" != "$REMOTE" ]]; then
    echo "ERROR: Local main is not up to date with origin. Run: git pull"
    exit 1
fi

# Build and test
echo "==> Building..."
make build

echo "==> Running tests..."
make test

# Check if tag exists
if git tag | grep -q "^${TAG}$"; then
    echo "ERROR: Tag ${TAG} already exists"
    exit 1
fi

# Create and push tag
echo "==> Creating tag ${TAG}..."
git tag -a "${TAG}" -m "Release ${TAG}"
git push origin "${TAG}"

echo ""
echo "==> Tag ${TAG} pushed. GitHub Actions will create the release."
echo "    https://github.com/dpanic/os-kickstart/releases/tag/${TAG}"
echo ""
echo "    Monitor: gh run list --limit 1"
