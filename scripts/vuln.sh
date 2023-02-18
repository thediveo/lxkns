#!/bin/bash
set -e

if ! command -v govulncheck &>/dev/null; then
    export PATH="$(go env GOPATH)/bin:$PATH"
    if ! command -v govulncheck &>/dev/null; then
        go install golang.org/x/vuln/cmd/govulncheck@latest
    fi
fi

govulncheck ./...
