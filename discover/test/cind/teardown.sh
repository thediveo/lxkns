#!/bin/bash
TESTDIR=$(dirname "$0")
. "${TESTDIR}/common.sh"

docker rm -f ${CNTR_NAME}

exit 0
