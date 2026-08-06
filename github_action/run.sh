#!/bin/bash

# Deployly GitHub Action Runner Script
# Usage: ./run.sh [save|restore] [path] [lockfile] [prefix]

COMMAND=$1
CACHE_PATH=$2
LOCKFILE=$3
PREFIX=$4

set -e

# 1. Download/Ensure CLI binary exists
# In a production scenario, we would download a pre-compiled binary from GitHub Releases.
# For this implementation, we assume the binary is available or built in the environment.
CLI_BIN="deployly"

if ! command -v $CLI_BIN &> /dev/null; then
    echo "Deployly CLI not found in PATH. Attempting to build from source..."
    # Assuming the source is available in the runner's workspace or a known location
    go build -o $CLI_BIN cli/main.go
fi

# 2. Execute Command
if [ "$COMMAND" == "restore" ]; then
    echo "--- Deployly Restore ---"
    $CLI_BIN restore --path "$CACHE_PATH" --lockfile "$LOCKFILE" --prefix "$PREFIX" --url "$DEPLOYLY_URL" --key "$DEPLOYLY_KEY"
    
elif [ "$COMMAND" == "save" ]; then
    echo "--- Deployly Save ---"
    $CLI_BIN save --path "$CACHE_PATH" --lockfile "$LOCKFILE" --prefix "$PREFIX" --url "$DEPLOYLY_URL" --key "$DEPLOYLY_KEY"
    
else
    echo "Error: Unknown command $COMMAND"
    exit 1
fi
