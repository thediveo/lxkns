ARG GOVERSION=1.15

FROM golang:${GOVERSION}
RUN adduser --disabled-password --gecos "" luser
WORKDIR /home/luser
# The script to kick off the tests should be fairly stable, more so than
# dependencies, so we cache it as early as we can.
COPY --chown=luser deployments/test/runtests /
# Cache dependencies to some extend, to speed up things...
COPY --chown=luser go.mod go.sum ./
#RUN su luser -c "go mod graph | awk '{if (\$1 !~ \"@\") print \$2}' | xargs go get"
RUN su luser -c "go mod download"
# Copy in the full lxkns module; unfortunately, tests cannot be prebuild and we
# cannot run tests on namespaces in a build container, so that's all we can do
# here.
COPY --chown=luser . .
# Tests must be started as root, so we can run the whole test suite once as root
# and another time as non-root.
CMD ["/runtests"]
