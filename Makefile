# Where to install the CLI tool binaries to
PREFIX ?= /usr/local
GOPATH = $(shell go env GOPATH)

# Go version to use when building containers
goversion = 1.15

tools := dumpns lsallns lspidns lsuns nscaps pidtree

# Location of the test image Docker compose project, and its project name.
ci_test_dir := test/ci-test-image
ci_test_projectname := lxknstest

.PHONY: clean coverage help install test

help:
	@echo "available targets: clean, coverage, install, test"

clean:
	rm -f $(tools)
	rm -f coverage.html coverage.out coverage-root.out
	rm -f coverage.txt coverage-root.txt

coverage:
	scripts/cov.sh

install:
	go install -v ./cmd/... ./examples/lsallns
	install -t $(PREFIX)/bin $(addprefix $(GOPATH)/bin/,$(tools))

test: # runs all tests in a container
	docker-compose -p $(ci_test_projectname) -f $(ci_test_dir)/docker-compose.yaml build --build-arg GOVERSION=$(goversion)
	docker-compose -p $(ci_test_projectname) -f $(ci_test_dir)/docker-compose.yaml up
