# Tinkering

## make Targets

The repository's top-level directory contains a simple `Makefile` featuring the
following targets:

- `test`: builds and runs all tests inside a container; the tests are run twice,
  once as root and once as a non-root user.

- `deploy` and `undeploy`: builds and starts, or stops, the containerized lxkns
  discovery service.

- `coverage`: runs a full coverage on all tests in the module, once as root,
  once as non-root, resulting in a single `coverage.html`.

- `clean`: removes coverage files, as well as any top-level CLI tool binaries
  that happened to end up there instead of `${GOPATH}/bin`.

- `install`: builds and installs the CLI binaries into `${GOPATH}/bin`, then
  installs these binaries into `/usr/local/bin`.

## Automated Tests

All lxkns module tests (including tests for the CLI tools) can be run in a test
container, see the `deployments/test` directory for how the test container is
built.

A Docker engine (including the containerd engine) is required to be installed,
operational, and accessible also from non-root users.

Getting rid of `--privileged` even for the test container was a challenge. The
last missing piece in this puzzle was Docker's CLI flag `--security-opt
systempaths=unconfined`. This finally allows us to successfully pass tests in
child PID namespaces (and even inside child user namespaces to get a better kick
out of it) which require remounting `/proc`. See also the [Docker Engine 19.03
release notes](https://docs.docker.com/engine/release-notes/19.03/), and
[PR&nbsp;#1808: add cli integration for unconfined
systempaths](https://github.com/docker/cli/pull/1808).

It's funny to see how people really get happy when `--privileged` gets dropped,
yet `CRAP_SYS_ADMIN` and `CAP_SYS_PTRACE` doesn't ring any bells – when these
should ring for kingdom come.
