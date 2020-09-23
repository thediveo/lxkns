# README

Runs all tests in a dedicated Docker container; use `make test` in the
project's top-level directory.

- build variable `GOVERSION` controls the Golang version to use for building a
  test image. It defaults to version `1.15` at the time of this writing.

- all tests are run twice: once without root and once as root.