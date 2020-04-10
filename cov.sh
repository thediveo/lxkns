#!/bin/bash
if ! command -v go-acc; then
    PATH="$(go env GOPATH)/bin:$PATH"
    if ! command -v go-acc; then
        go get github.com/ory/go-acc
    fi
fi

go-acc --covermode atomic -o coverage.txt ./... -- -v \
    && go tool cover -html coverage.txt
