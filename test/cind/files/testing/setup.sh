#!/bin/bash

CNTR_IMG="docker.io/library/busybox:latest"

# Wait for containerd's API socket to become useable...
until socat -u OPEN:/dev/null UNIX-CONNECT:/var/run/containerd/containerd.sock; do
    echo "waiting for containerd..."
    sleep 0.5
done

# Spin up a test container...
ctr image pull "${CNTR_IMG}"
ctr run \
    --label name=sleepy \
    --read-only \
    --snapshotter=native "${CNTR_IMG}" \
    sleepy /bin/sh -c 'echo "SLEEPY READY"; sleep 1000000'

exit 0
