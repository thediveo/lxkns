#!/bin/bash
set -e

if ! command -v goreportcard-cli; then
    PATH="$(go env GOPATH)/bin:$PATH"
    if ! command -v goreportcard-cli; then
        # Don't touch our local module dependencies, so run installation from
        # somewhere else...
        (cd /tmp && go get github.com/gojp/goreportcard)
        (cd $(go env GOPATH)/src/github.com/gojp/goreportcard && make install)
        (cd /tmp && go get github.com/gojp/goreportcard/cmd/goreportcard-cli)
    fi
fi

goreportcard-cli -v
