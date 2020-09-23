#!/bin/bash
set -e

if ! command -v godoc; then
    PATH="$(go env GOPATH)/bin:$PATH"
    if ! command -v godoc; then
        # Don't touch our local module dependencies, so run installation from
        # somewhere else...
        (cd /tmp && go get -v golang.org/x/tools/cmd/godoc)
    fi
fi

godoc -http=:6060 &
