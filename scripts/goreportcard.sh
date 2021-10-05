#!/bin/bash
set -e

if ! command -v goreportcard-cli; then
    PATH="$(go env GOPATH)/bin:$PATH"
    if ! command -v goreportcard-cli; then
        rm -rf /tmp/goreportcard || true
        git clone https://github.com/gojp/goreportcard.git /tmp/goreportcard
        (cd /tmp/goreportcard && make install && go install ./cmd/goreportcard-cli)
        rm -rf /tmp/goreportcard || true
        # Install a somewhat recent ineffassign over the totally outdated one
        # that goreportcard still insists of infecting the system with.
        (cd /tmp && go get -u github.com/gordonklaus/ineffassign)
        # Install the missing misspell, oh well...
        (cd /tmp && go get -u github.com/client9/misspell/cmd/misspell)
    fi
fi

goreportcard-cli -v
