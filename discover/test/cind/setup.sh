#!/bin/bash
TESTDIR=$(dirname "$0")
. "${TESTDIR}/common.sh"

# Stop and get rid of any old running instance, as we want to start anew.
docker rm -f ${CNTR_NAME}

# Build the testing image, if not already cached.
docker build --build-arg "KINDBASE_IMAGE=${KINDBASE_IMAGE}" -t ${CIND_IMAGE} ${TESTDIR}

# Fire it up...
docker run -d -it --rm --name ${CNTR_NAME} --privileged ${CIND_IMAGE} || exit 1

# Now wait for the test container to have spun up and be ready...
exec 3< <(docker logs -f ${CNTR_NAME})
DOCKERLOGS_PID=$!
NEXTDOT=0
echo -n "waiting for test containers"
while IFS= read -r <&3 LINE; do
    if echo "$LINE" | grep -q "SLEEPY READY"; then
        break
    fi
    ((NEXTDOT--))
    if (( ${NEXTDOT} <= 0 )); then
        echo -n "."
        NEXTDOT=${LINESPERDOT}
    fi
done
echo
if [[ $(docker inspect -f '{{.State.Running}}' "${CNTR_NAME}") == "true" ]]; then
    kill $DOCKERLOGS_PID
    echo "container in container spun up"
else
    echo "ERROR: failed to correctly start testing container"
    exit 1
fi

exit 0
