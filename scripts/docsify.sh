#!/bin/bash
set -e

# In case the user hasn't set an explicit installation location, avoid polluting
# our own project...
NPMBIN=$(cd $HOME && npm root)/.bin
export PATH="$NPMBIN:$PATH"
if ! command -v docsify &>/dev/null; then
    (cd $HOME && npm install docsify-cli)
fi

echo "starting docsify on port 3300 (and 3301)..."
docsify serve -p 3300 -P 3301 "$@"
