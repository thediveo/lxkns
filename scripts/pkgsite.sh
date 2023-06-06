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
npm list --depth=0 browser-sync &>/dev/null || \
    (cd $HOME && npm install browser-sync)

npm list --depth=0 nodemon &>/dev/null || \
    (cd $HOME && npm install nodemon)

# https://stackoverflow.com/a/2173421
trap "trap - SIGTERM && kill -- -$$" SIGINT SIGTERM EXIT

# https://mdaverde.com/posts/golang-local-docs
npm exec -- browser-sync start --port 6060 --proxy localhost:6061 --reload-delay 2000 --reload-debounce 5000 --no-ui --no-open &
PKGSITE=$(which pkgsite)
npm exec -- nodemon --signal SIGTERM --watch './**/*' -e go --exec "browser-sync --port 6060 reload && $PKGSITE -http=localhost:6061 ."
