#!/bin/bash
DIR=$(dirname "$0")

# Updates the list of capabilities
LIBCAP_VERSION="2.43"

wget -O - "https://git.kernel.org/pub/scm/libs/libcap/libcap.git/snapshot/libcap-${LIBCAP_VERSION}.tar.gz" \
    | tar -xz -C /tmp "libcap-${LIBCAP_VERSION}/libcap/include/uapi/linux/capability.h"

OUTFILE="${DIR}/capnames.go"
echo "// DO NOT EDIT
// automatically generated file using \"update-caplist.sh\"
// from libcap version ${LIBCAP_VERSION}

package caps

// CapNames maps the defined CAP_xxx constants to suitable capabilities names
// in form of \"cap_xxx\".
var CapNames = map[int]string{" > ${OUTFILE}
sed -n -e 's/^#define\s\(CAP_[A-Z_]\+\)\s\+\([[:digit:]]\+\)/    \2: "\L\1",/p' \
    "/tmp/libcap-${LIBCAP_VERSION}/libcap/include/uapi/linux/capability.h" \
    >> ${OUTFILE}
echo "}" >> ${OUTFILE}

gofmt -s ${OUTFILE} > ${OUTFILE}.new
mv -f ${OUTFILE}.new ${OUTFILE}

echo "success."
