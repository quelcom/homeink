#!/bin/bash

CONTAINER_IMAGE="chromedp/headless-shell"

docker container inspect -f '{{.State.Running}}' $(docker ps -aqf "ancestor=$CONTAINER_IMAGE") &>/dev/null

retVal=$?
if [ $retVal -ne 0 ]; then
    echo "Headless shell container not running. Starting now..."
    docker run -d -p 9222:9222 --rm --name headless-shell "$CONTAINER_IMAGE"
fi

echo "Running tests..."
go test -v ./...

echo "Stop running container..."
docker container stop  $(docker ps -aqf "ancestor=$CONTAINER_IMAGE")

echo "Everything done."
