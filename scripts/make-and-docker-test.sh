#!/bin/bash

REPO_DIR=$(git rev-parse --show-toplevel)
cd "$REPO_DIR" || exit 1

DOCKER_DIR="docker"

make all

cd "$DOCKER_DIR" || exit 1

docker compose restart
#docker exec -it rsyslox-test /opt/rsyslox/log-generator.sh

