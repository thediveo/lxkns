ARG GOVERSION=1.16

FROM golang:${GOVERSION}
RUN apt-get update && \
    apt-get install -y sudo && \
    adduser --disabled-password --gecos "" luser && \
    adduser luser sudo && \
    echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers
WORKDIR /home/luser
# The script to kick off the tests should be fairly stable, more so than
# dependencies, so we cache it as early as we can.
COPY --chown=luser deployments/test/runtests /
# Cache dependencies to some extend, to speed up things...
COPY --chown=luser go.mod go.sum ./
#RUN su luser -c "go mod graph | awk '{if (\$1 !~ \"@\") print \$2}' | xargs go get"
RUN su luser -c "go mod download -x"
# Copy in the full lxkns module; unfortunately, tests cannot be prebuild and we
# cannot run tests on namespaces in a build container, so that's all we can do
# here.
COPY --chown=luser . .
USER luser
CMD ["/runtests"]
