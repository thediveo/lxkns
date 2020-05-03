#!/bin/bash
DIR=$(dirname "$0")

# Updates the list of capabilities
LIBCAP_VERSION="2.33"

wget -O - "https://git.kernel.org/pub/scm/libs/libcap/libcap.git/snapshot/libcap-${LIBCAP_VERSION}.tar.gz" \
    | tar -xz -C /tmp "libcap-${LIBCAP_VERSION}/libcap/include/uapi/linux/capability.h"

OUTFILE="${DIR}/capnames.go"
echo "// DO NOT EDIT
// automatically generated file using \"update-caplist.sh\"

package main

var CapNames = map[int]string{" > ${OUTFILE}
sed -n -e 's/^#define\s\(CAP_[A-Z_]\+\)\s\+\([[:digit:]]\+\)/    \2: "\L\1",/p' \
    "/tmp/libcap-${LIBCAP_VERSION}/libcap/include/uapi/linux/capability.h" \
    >> ${OUTFILE}
echo "}" >> ${OUTFILE}
