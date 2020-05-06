#!/bin/sh
if ! which dlv ; then
	PATH="${GOPATH}/bin:$PATH"
fi
if [ "$DEBUG_AS_ROOT" = "true" ]; then
	DLV=$(which dlv)
	exec sudo "$DLV" --only-same-user=false "$@"
else
	exec dlv "$@"
fi