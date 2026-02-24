#!/bin/bash

REPO_DIR=$(git rev-parse --show-toplevel)
cd "$REPO_DIR" || exit 1
DOCKER_DIR="docker"
[[ $1 == "make" ]] && make all
cd "$DOCKER_DIR" || exit 1
docker compose restart

