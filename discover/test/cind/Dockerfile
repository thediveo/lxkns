# Our test harness image is based off the kindest/base Kubernetes-in-Docker base
# image, all built locally, so we can also build for architectures where there
# are no pre-built kindest/base images available.

# Thanks to https://github.com/moby/moby/pull/31352, it is possible to
# parameterize the FROM statement, yay!
ARG KINDBASE_IMAGE
FROM ${KINDBASE_IMAGE}
COPY files/ /
RUN mkdir -p /kind \
    && echo "Installing packages..." \
        && apt-get update \
        && apt-get install -y socat \
    && echo "Enabling systemd testing service..." \
        && systemctl enable testing
ENTRYPOINT [ "/usr/local/bin/entrypoint", "/sbin/init" ]
