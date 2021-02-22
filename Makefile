# Where to install the CLI tool binaries to
PREFIX ?= /usr/local
GOPATH = $(shell go env GOPATH)
GIT_VERSION = $(shell git describe --tags 2>/dev/null || echo "v0.0.0")

# Go version to use when building the test containers; start with a version
# 1.14+ first to get better testbasher diagnosis in case a test script runs
# into trouble.
goversion = 1.15 1.13

tools := dumpns lsallns lspidns lsuns nscaps pidtree

# To suckessfully run the tests, we need CRAP_SYS_ADMIN capabilities.
testcontaineropts := \
	--pid host \
	--cap-drop ALL \
	--cap-add CAP_SYS_ADMIN \
	--cap-add CAP_SYS_CHROOT \
	--cap-add CAP_SYS_PTRACE \
	--cap-add CAP_DAC_READ_SEARCH \
	--cap-add CAP_SETUID \
	--cap-add CAP_SETGID \
	--security-opt systempaths=unconfined \
	--security-opt apparmor=unconfined \
	--security-opt seccomp=unconfined \
	-v /sys/fs/cgroup:/sys/fs/cgroup:rw

.PHONY: clean coverage deploy undeploy help install test report buildapp startapp

help:
	@echo "available targets: clean, coverage, deploy, undeploy, install, test, report, buildapp, startapp"

clean:
	rm -f $(tools)
	rm -f coverage.html coverage.out coverage-root.out
	rm -f coverage.txt coverage-root.txt

coverage:
	scripts/cov.sh

deploy:
	@echo "deploying version" $${GIT_VERSION}
	docker-compose -p lxkns -f deployments/lxkns/docker-compose.yaml build --build-arg GIT_VERSION=$(GIT_VERSION)
	docker-compose -p lxkns -f deployments/lxkns/docker-compose.yaml up

undeploy:
	docker-compose -p lxkns -f deployments/lxkns/docker-compose.yaml down

install:
	go install -v ./cmd/... ./examples/lsallns
	install -t $(PREFIX)/bin $(addprefix $(GOPATH)/bin/,$(tools))

test: # runs all tests in a container
	@set -e; for GOVERSION in $(goversion); do \
		echo "🧪 🧪 🧪 Testing on Go $${GOVERSION}"; \
		docker build -t lxknstest:$${GOVERSION} --build-arg GOVERSION=$${GOVERSION} -f deployments/test/Dockerfile .;  \
		docker run -it --rm --name lxknstest_$${GOVERSION} $(testcontaineropts) lxknstest:$${GOVERSION}; \
	done; \
	echo "🎉 🎉 🎉 All tests passed"

report:
	@./scripts/goreportcard.sh

buildapp:
	@echo "building version" $${GIT_VERSION}
	@cd web/lxkns && yarn build

startapp:
	@echo "starting version" $${GIT_VERSION}
	@cd web/lxkns && yarn start
	