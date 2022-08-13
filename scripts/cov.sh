#!/bin/bash
set -e

if ! command -v go-acc; then
    PATH="$(go env GOPATH)/bin:$PATH"
    if ! command -v go-acc; then
        go install github.com/ory/go-acc@latest
    fi
fi

if ! command -v gobadge &>/dev/null; then
    export PATH="$(go env GOPATH)/bin:$PATH"
    if ! command -v gobadge &>/dev/null; then
        go install github.com/AlexBeauchemin/gobadge@latest
    fi
fi

# First, run the tests as root; this will skip a few tests where we explicitly
# will need root, but running as root is the broader scope.
go-acc --covermode atomic -o $(pwd)/coverage-root.out ./... -- -v -p 1 -exec sudo
# Second, rerun the tests as non-root in order to cover the tests skipped
# on the first root run.
go-acc --covermode atomic -o coverage.out ./... -- -v -p 1
tail -n +2 coverage-root.out >> coverage.out
go tool cover -func=coverage.out -o=coverage.out
gobadge -filename=coverage.out -green=80 -yellow=50
