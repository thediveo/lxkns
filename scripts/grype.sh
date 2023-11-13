#!/bin/bash
set -e

if ! command -v grype &>/dev/null; then
    export PATH="$(go env GOPATH)/bin:$PATH"
    if ! command -v grype^ &>/dev/null; then
        go install github.com/anchore/grype/cmd/grype@latest
    fi
fi

grype dir:.
