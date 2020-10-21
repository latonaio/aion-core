#!/usr/bin/env bash

PUSH=$1
DATE="$(date "+%Y%m%d%H%M")"
REPOSITORY_NAME="latonaio"
IMAGE_NAME="aion-statuskanban"
DOCKERFILE_DIR="./cmd/kanban-server"
DOCKERFILE_NAME="Dockerfile-kanban-server"

# build servicebroker
DOCKER_BUILDKIT=1 docker build -f ${DOCKERFILE_DIR}/${DOCKERFILE_NAME} -t ${REPOSITORY_NAME}/${IMAGE_NAME}:"${DATE}" .
docker tag ${REPOSITORY_NAME}/${IMAGE_NAME}:"${DATE}" ${REPOSITORY_NAME}/${IMAGE_NAME}:latest

if [[ $PUSH == "push" ]]; then
    docker push ${REPOSITORY_NAME}/${IMAGE_NAME}:"${DATE}"
    docker push ${REPOSITORY_NAME}/${IMAGE_NAME}:latest
fi
