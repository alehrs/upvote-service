#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

if ! docker info > /dev/null 2>&1; then
    echo "Error: Docker is not running."
    exit 1
fi

echo "Building and starting upvote-service..."
docker compose up --build
