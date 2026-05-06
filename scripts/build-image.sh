#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

if ! docker info > /dev/null 2>&1; then
    echo "Error: Docker is not running."
    exit 1
fi

PLATFORM="${PLATFORM:-linux/amd64,linux/arm64}"
IMAGE="${IMAGE:-upvote-service}"

# Multi-platform builds require --push to a registry; single-platform can load locally.
if [[ "$PLATFORM" == *","* ]]; then
    if [[ "$IMAGE" != *"/"* ]]; then
        echo "Error: multi-platform builds require a registry image name (e.g. IMAGE=ghcr.io/user/upvote-service)."
        exit 1
    fi
    OUTPUT="--push"
else
    OUTPUT="--load"
fi

echo "Building Docker image: $IMAGE (platforms: $PLATFORM)..."
docker buildx build --platform "$PLATFORM" -t "$IMAGE" $OUTPUT .
echo "Image built successfully: $IMAGE"
