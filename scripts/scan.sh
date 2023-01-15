#!/bin/bash
set -e

export PATH="$(go env GOPATH)/bin:$PATH"
go install github.com/anchore/syft/cmd/syft@latest
go install github.com/anchore/grype@latest

BOMFILE=$(mktemp "/tmp/lxkns-scan.XXXXXX.json")
trap 'rm -- "$BOMFILE"' EXIT
syft dir:. -o json > $BOMFILE
grype sbom:$BOMFILE
