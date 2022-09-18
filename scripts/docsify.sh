#!/bin/bash
set -e

if ! command -v pkgsite &>/dev/null; then
    export PATH="$(go env GOPATH)/bin:$PATH"
    if ! command -v pkgsite &>/dev/null; then
        go install golang.org/x/pkgsite/cmd/pkgsite@master
    fi
fi

# In case the user hasn't set an explicit installation location, avoid polluting
# our own project...
NPMBIN=$(cd $HOME && npm bin)
export PATH="$NPMBIN:$PATH"
if ! command -v docsify &>/dev/null; then
    (cd $HOME && npm install docsify-cli)
fi

echo "starting docsify on port 3300 (and 3301)..."
docsify serve -p 3300 -P 3301 "$@"
