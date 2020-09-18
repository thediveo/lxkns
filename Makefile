# Where to install the CLI tool binaries to
PREFIX ?= /usr/local
GOPATH = $(shell go env GOPATH)

# Go version to use when building containers
goversion = 1.13 1.15

tools := dumpns lsallns lspidns lsuns nscaps pidtree

testcontaineropts := --privileged --pid host

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
	@set -e; for GOVERSION in $(goversion); do \
		echo "🧪 🧪 🧪 Testing on Go $${GOVERSION}"; \
		docker build -t lxknstest:$${GOVERSION} --build-arg GOVERSION=$${GOVERSION} -f test/image/Dockerfile .;  \
		docker run -it --rm --name lxknstest_$${GOVERSION} $(testcontaineropts) lxknstest:$${GOVERSION}; \
	done; \
	echo "🎉 🎉 🎉 All tests passed"
