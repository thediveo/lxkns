#!/bin/sh

# Quick and dirty check for lsallns always showing the same information,
# regardless of originating from its own discovery or an imported JSON
# discovery result.

# Of course, this isn't atomic, so it's still prone to spurious failures on a
# system where namespaces tend to come and go quickly and often.

TEMPFILE=$(mktemp)
# Dump all namespaces (and processes) as JSON and then generate an lsallns
# list from it.
sudo go run ./cmd/dumpns -c | go run ./examples/lsallns -i - > ${TEMPFILE}
# Now run a direct lsallns and compare the output of it with the one generated
# only indirectly.
sudo go run ./examples/lsallns | diff -s - ${TEMPFILE}
rm ${TEMPFILE}
