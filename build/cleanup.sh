#!/bin/bash

# Remove GPU Tracker files
GPU_TRACKER_FILE=/var/log/gpu-tracker.json
if [ -f "$GPU_TRACKER_FILE" ]; then
    rm -f "$GPU_TRACKER_FILE"
fi
GPU_TRACKER_LOCK_FILE=/var/log/gpu-tracker.lock
if [ -f "$GPU_TRACKER_LOCK_FILE" ]; then
    rm -f "$GPU_TRACKER_LOCK_FILE"
fi

# Remove default CDI directory if present
CDI_DIR="/etc/cdi"
if [ -d "$CDI_DIR" ]; then
    rm -rf "$CDI_DIR"
fi

# Default Docker config file
DOCKER_CONFIG="/etc/docker/daemon.json"

# Exit if Docker config file is not available
if [[ ! -f "$DOCKER_CONFIG" ]]; then
    exit 0
fi

# Exit if jq is not available
if ! type "jq" > /dev/null; then
    echo "jq is not available. Docker config file cannot be cleaned up."
    exit 0
fi

update=false
out=$(cat "$DOCKER_CONFIG")

# Unset AMD default runtime
if $(echo "$out" | jq 'has("default-runtime") and ."default-runtime" == "amd"'); then
    out=$(jq 'del(.["default-runtime"])' "$DOCKER_CONFIG")
    update=true
fi

# Remove AMD runtime
if $(echo "$out" | jq '.runtimes | has("amd")'); then
    out=$(echo "$out" | jq -e 'del(.runtimes["amd"])')
    if $(echo "$out" | jq '.runtimes | length == 0'); then
        out=$(echo "$out" | jq -e 'del(.["runtimes"])')
    fi
    update=true
fi

# Remove CDI feature
if $(echo "$out" | jq '.features | has("cdi")'); then
    out=$(echo "$out" | jq -e 'del(.features["cdi"])')
    if $(echo "$out" | jq '.features | length == 0'); then
        out=$(echo "$out" | jq -e 'del(.["features"])')
    fi
    update=true
fi

# Update Docker config file
if [ "$update" == true ]; then
    echo "$out" > "$DOCKER_CONFIG"
    echo "Updated the docker config file"
    echo "Please restart docker daemon"
fi