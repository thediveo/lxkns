#!/bin/bash
set -e

export BUILDTAGS="podman,exclude_graphdriver_btrfs,exclude_graphdriver_devicemapper,libdm_no_deferred_remove"

if ! command -v gobadge &>/dev/null; then
    export PATH="$(go env GOPATH)/bin:$PATH"
    if ! command -v gobadge &>/dev/null; then
        go install github.com/AlexBeauchemin/gobadge@latest
    fi
fi

# As of Go 1.20 (and later) we now use the new coverage profiling support that
# also *cough* covers integration tests (even if this particular project might
# not use the latter). The benefit of this slightly more involved approach is
# that we don't need external coverage profile file processing tools anymore,
# but can achieve our goal with just using the standard Go toolchain.

# First, we set up a temporary directory to receive the coverage (binary)
# files...
GOCOVERTMPDIR="$(mktemp -d)"
trap 'rm -rf -- "$GOCOVERTMPDIR"' EXIT
# Now run the (unit) tests with coverage, but don't use the existing textual
# format and instead tell "go test" to produce the new binary coverage data file
# format. This way we can easily run multiple coverage (integration) tests, as
# needed, without worrying about how to aggregate the coverage data later. The
# new Go toolchain already does this for us. 
go test -cover -v -exec sudo -tags ${BUILDTAGS} -p=1 -count=1 -race ./... -args -test.gocoverdir="$GOCOVERTMPDIR"
go test -cover -v -tags ${BUILDTAGS} -p=1 -count=1 -race ./... -args -test.gocoverdir="$GOCOVERTMPDIR"
# Finally transform the coverage information collected in potentially multiple
# runs into the well-proven textual format so we can process it as we have come
# to learn and love.
go tool covdata textfmt -i="$GOCOVERTMPDIR" -o=coverage.out
go tool cover -html=coverage.out -o=coverage.html
go tool cover -func=coverage.out -o=coverage.out
gobadge -filename=coverage.out -green=80 -yellow=50
