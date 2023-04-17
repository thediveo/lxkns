# Where to install the CLI tool binaries to
PREFIX ?= /usr/local
GOPATH = $(shell go env GOPATH)
GOGEN = go generate .

GET_SEMVERSION = awk '{match($$0,/const\s+SemVersion\s+=\s+"(.*)"/,m);if (m[1]!="") print m[1]}' defs_version.go

# Go version to use when building the test containers; see README.md for
# supported versions strategy.
goversion = 1.19 1.18

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

.PHONY: clean vuln coverage deploy undeploy help install test report manual pkgsite buildapp startapp scan systempodman systempodman-down

help: ## list available targets
	@# Shamelessly stolen from Gomega's Makefile
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}'

clean: ## cleans up build and testing artefacts
	rm -f $(tools)
	rm -f coverage.html coverage.out coverage-root.out
	rm -f coverage.txt coverage-root.txt

coverage: ## gathers coverage and updates README badge
	@scripts/cov.sh

manual: ## start docsify server for manual
	@scripts/docsify.sh ./docs

pkgsite: ## serves Go documentation on port 6060
	@echo "navigate to: http://localhost:6060/github.com/thediveo/lxkns"
	@scripts/pkgsite.sh

deploy: ## deploys lxkns service on host port 5010
	$(GOGEN)
	$(eval GIT_VERSION := $(shell $(GET_SEMVERSION)))
	docker buildx build -t lxkns --build-arg GIT_VERSION=$(GIT_VERSION) -f deployments/lxkns/Dockerfile .
	docker compose -p lxkns -f deployments/lxkns/docker-compose.yaml up

undeploy: ## removes any deployed lxkns service
	docker compose -p lxkns -f deployments/lxkns/docker-compose.yaml down

install: ## installs lxkns commands
	$(GOGEN)
	go install -v ./cmd/... ./examples/lsallns
	install -t $(PREFIX)/bin $(addprefix $(GOPATH)/bin/,$(tools))

test: ## runs all tests
	go test -v -p=1 -count=1 -exec sudo ./... && go test -v -p=1 -count=1 ./...

testc: ## runs all tests in test containers
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
	@scripts/goreportcard.sh

buildapp: ## builds web UI app
	$(GOGEN)
	$(eval GIT_VERSION := $(shell $(GET_SEMVERSION)))
	@echo "building version" $(GIT_VERSION)
	@cd web/lxkns && GIT_VERSION=$(GIT_VERSION) yarn build

startapp: ## starts web UI app for development
	$(GOGEN)
	$(eval GIT_VERSION := $(shell $(GET_SEMVERSION)))
	@echo "starting version" $(GIT_VERSION)
	@cd web/lxkns && GIT_VERSION=$(GIT_VERSION) yarn start

scan: ## scans the repository for CVEs
	@scripts/scan.sh

systempodman: ## builds lxkns using podman system service
	$(GOGEN)
	$(eval GIT_VERSION := $(shell $(GET_SEMVERSION)))
	sudo podman build -t lxkns --build-arg GIT_VERSION=$(GIT_VERSION) -f deployments/podman/Dockerfile .
	# podman-compose doesn't support "pid:host" which we absolutely need.
	sudo docker --host unix:///run/podman/podman.sock compose -p lxkns -f deployments/podman/docker-compose.yaml up

systempodman-down: ## removes any deployed lxkns service
	sudo docker --host unix:///run/podman/podman.sock compose -p lxkns -f deployments/podman/docker-compose.yaml down

userpodman: ## builds lxkns using podman system service
	$(GOGEN)
	$(eval GIT_VERSION := $(shell $(GET_SEMVERSION)))
	podman build -t userlxkns --build-arg GIT_VERSION=$(GIT_VERSION) -f deployments/podman/Dockerfile .
	$(eval UID := $(shell id -u))
	UID=$(UID) docker --host unix:///run/user/$(UID)/podman/podman.sock compose -p userlxkns -f deployments/userpodman/docker-compose.yaml up

vuln: ## run go vulnerabilities check
	@scripts/vuln.sh
