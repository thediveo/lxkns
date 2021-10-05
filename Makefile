# Where to install the CLI tool binaries to
PREFIX ?= /usr/local
GOPATH = $(shell go env GOPATH)
GIT_VERSION = $(shell git describe 2>/dev/null || echo "v0.0.0")
GOGEN = go generate .

# Go version to use when building the test containers; start with a version
# 1.14+ first to get better testbasher diagnosis in case a test script runs
# into trouble.
goversion = 1.16 1.15

tools := dumpns lsallns lspidns lsuns nscaps pidtree lxkns

# To suckessfully run the tests, we need CRAP_SYS_ADMIN capabilities.
testcontaineropts := \
	--pid host \
	--cap-drop ALL \
	--cap-add CAP_SYS_ADMIN \
	--cap-add CAP_SYS_CHROOT \
	--cap-add CAP_SYS_PTRACE \
	--cap-add CAP_DAC_READ_SEARCH \
	--cap-add CAP_DAC_OVERRIDE \
	--cap-add CAP_SETUID \
	--cap-add CAP_SETGID \
	--security-opt systempaths=unconfined \
	--security-opt apparmor=unconfined \
	--security-opt seccomp=unconfined \
	-v /sys/fs/cgroup:/sys/fs/cgroup:rw

.PHONY: clean coverage deploy undeploy help install test report buildapp startapp docsify

help: ## list available targets
	@# Shamelessly stolen from Gomega's Makefile
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}'

clean: ## cleans up build and testing artefacts
	rm -f $(tools)
	rm -f coverage.html coverage.out coverage-root.out
	rm -f coverage.txt coverage-root.txt

coverage: ## runs tests with code coverage
	scripts/cov.sh

deploy: ## deploys lxkns service on host port 5010
	@echo "deploying version" $${GIT_VERSION}
	$(GOGEN)
	docker-compose -p lxkns -f deployments/lxkns/docker-compose.yaml build --build-arg GIT_VERSION=$(GIT_VERSION)
	docker-compose -p lxkns -f deployments/lxkns/docker-compose.yaml up

undeploy: ## removes any deployed lxkns service
	docker-compose -p lxkns -f deployments/lxkns/docker-compose.yaml down

install: ## installs lxkns commands
	$(GOGEN)
	go install -v ./cmd/... ./examples/lsallns
	install -t $(PREFIX)/bin $(addprefix $(GOPATH)/bin/,$(tools))

test: ## runs all tests in test containers
	$(GOGEN)
	@set -e; for GOVERSION in $(goversion); do \
		echo "🧪 🧪 🧪 Testing on Go $${GOVERSION}"; \
		docker build -t lxknstest:$${GOVERSION} --build-arg GOVERSION=$${GOVERSION} -f deployments/test/Dockerfile .;  \
		docker run -it --rm --name lxknstest_$${GOVERSION} $(testcontaineropts) lxknstest:$${GOVERSION}; \
	done; \
	echo "🎉 🎉 🎉 All tests passed"

# builds a static webapp and the lxkns service, then runs the service and the
# webapp unit and end-to-end tests, and finally stops the lxkns service after
# the tests have run.
citestapp: ## builds and tests lxkns with static web UI
	@sudo /bin/true
	@$(GOGEN)
	@cd web/lxkns && yarn build:dev
	@go build -v ./cmd/lxkns
	@TMPPIDFILE=$$(mktemp -p /tmp lxkns.service.pid.XXXXXXXXXX) && \
	sudo chown root $$TMPPIDFILE && \
	sudo bash -c "./lxkns --debug --http localhost:5100 & echo \$$! > $$TMPPIDFILE" && \
	sudo bash -c "chown \$$SUDO_USER $$TMPPIDFILE" && \
	ls -l $$TMPPIDFILE && \
	LXKNSPID=$$(cat $$TMPPIDFILE) && \
	rm $$TMPPIDFILE && \
	echo "lxkns background service PID:" $$LXKNSPID && \
	(cd web/lxkns && yarn cypress:run --config baseUrl=http://localhost:5100,screenshotOnRunFailure=false); STATUS=$$? ; \
	sleep 1s && \
	echo "stopping lxkns background service and waiting for it to exit..." && \
	sudo kill $$LXKNSPID && \
	timeout 10s tail --pid=$$LXKNSPID -f /dev/null && \
	exit $$STATUS

report: ## runs goreportcard
	@./scripts/goreportcard.sh

buildapp: ## builds web UI app
	@echo "building version" $${GIT_VERSION}
	@cd web/lxkns && yarn build

startapp: ## starts web UI app for development
	@echo "starting version" $${GIT_VERSION}
	@cd web/lxkns && yarn start

docsify: ## serves docsified docs on host port(s) 3030 and 3031
	@docsify serve -p 3030 -P 3031 docs
