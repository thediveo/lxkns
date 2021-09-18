#!/bin/bash
set -e

if ! command -v go-acc; then
    PATH="$(go env GOPATH)/bin:$PATH"
    if ! command -v go-acc; then
        # Don't touch our local module dependencies, so run installation from
        # somewhere else...
        (cd /tmp && go get -v github.com/ory/go-acc)
    fi
fi

# First, run tests as non-root; this will need to skip some tests.
go-acc --covermode atomic -o coverage.out ./... -- -v
# Second, run the tests now again, but this tome as root; this will skip some
# other tests, but run the missing ones that need to be run as root.
go-acc --covermode atomic -o $(pwd)/coverage-root.out ./... -- -v -exec sudo
tail -n +2 coverage-root.out >> coverage.out
go tool cover -html coverage.out -o coverage.html
# xdg-open coverage.html
