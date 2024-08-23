#!/bin/bash

# Default values
default_server_entrypoint="./cmd/server/main.go"
default_server_name="server"
default_server_version=""

default_agent_entrypoint="./cmd/agent/main.go"
default_agent_name="agent"
default_agent_version=""

default_binaries_dir="./binaries"

# Initialize variables with default values
SERVER_ENTRYPOINT="$default_server_entrypoint"
SERVER_NAME="$default_server_name"
SERVER_VERSION="$default_server_version"

AGENT_ENTRYPOINT="$default_agent_entrypoint"
AGENT_NAME="$default_agent_name"
AGENT_VERSION="$default_agent_version"

BINARIES_DIR="$default_binaries_dir"

# Process long flags
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --sentry)
            SERVER_ENTRYPOINT=$2
            shift
            ;;
        --sname)
            SERVER_NAME=$2
            shift
            ;;
        --sver)
            SERVER_VERSION=$2
            shift
            ;;
        --aentry)
            AGENT_ENTRYPOINT=$2
            shift
            ;;
        --aname)
            AGENT_NAME=$2
            shift
            ;;
        --aver)
            AGENT_VERSION=$2
            shift
            ;;
        --out)
            BINARIES_DIR=$2
            shift
            ;;
        *)
            echo "Unknown flag: $1" 1>&2
            exit 1
            ;;
    esac
    shift
done

# Get the short hash of the current commit
COMMIT=$(git rev-parse --short HEAD)

# Build function
function build {
    local entrypoint=$1
    local filename=$2
    local buildversion=$3

    local buildtime=$(date -u +'%Y-%m-%d %H:%M:%S')

    echo "Compiling $entrypoint..."
    if ! go build -ldflags "-X main.buildVersion=$buildversion -X main.buildCommit=$COMMIT -X 'main.buildDate=$buildtime'" -o "$BINARIES_DIR/$filename" "$entrypoint"; then
        echo "Failed to build $entrypoint"
        exit 1
    fi
    echo "Done"
}

# Perform builds
build "$SERVER_ENTRYPOINT" "$SERVER_NAME" "$SERVER_VERSION"
build "$AGENT_ENTRYPOINT" "$AGENT_NAME" "$AGENT_VERSION"