# Where to install the CLI tool binaries to
PREFIX ?= /usr/local
GOPATH = $(shell go env GOPATH)
GOGEN = go generate .

GET_SEMVERSION = awk '{match($$0,/const\s+SemVersion\s+=\s+"(.*)"/,m);if (m[1]!="") print m[1]}' defs_version.go

# Go version to use when building the test containers; see README.md for
# supported versions strategy.
goversion = 1.21 1.20

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

.PHONY: clean vuln coverage deploy undeploy help install test report manual pkgsite buildapp startapp scan dist grype yarnsetup

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

dist: ## build multi-arch image (amd64, arm64) and push to local running registry on port 5000.
	$(GOGEN)
	$(eval GIT_VERSION := $(shell $(GET_SEMVERSION)))
	scripts/multiarch-builder.sh \
		--build-arg REACT_APP_GIT_VERSION=$(GIT_VERSION) \
		--build-context webappsrc=./web/lxkns

deploy: ## deploys lxkns service on host port 5010
	$(GOGEN)
	$(eval GIT_VERSION := $(shell $(GET_SEMVERSION)))
	scripts/docker-build.sh deployments/lxkns/Dockerfile \
		-t lxkns \
		--build-arg REACT_APP_GIT_VERSION=$(GIT_VERSION) \
		--build-context webappsrc=./web/lxkns
	docker compose -p lxkns -f deployments/lxkns/docker-compose.yaml up

undeploy: ## removes any deployed lxkns service
	docker compose -p lxkns -f deployments/lxkns/docker-compose.yaml down

install: ## installs lxkns commands
	$(GOGEN)
	go install -v ./cmd/... ./examples/lsallns
	install -t $(PREFIX)/bin $(addprefix $(GOPATH)/bin/,$(tools))

test: ## runs all tests
	go test -v -p=1 -count=1 -exec sudo ./... && go test -v -p=1 -count=1 ./...

report: ## runs goreportcard
	@scripts/goreportcard.sh

buildapp: ## builds web UI app
	$(GOGEN)
	$(eval GIT_VERSION := $(shell $(GET_SEMVERSION)))
	@echo "building version" $(GIT_VERSION)
	@cd web/lxkns && sed -i "s/^VITE_REACT_APP_GIT_VERSION=.*/VITE_REACT_APP_GIT_VERSION=$$GIT_VERSION/" .env && yarn build

startapp: ## starts web UI app for development
	$(GOGEN)
	$(eval GIT_VERSION := $(shell $(GET_SEMVERSION)))
	@echo "starting version" $(GIT_VERSION)
	@cd web/lxkns && sed -i "s/^VITE_REACT_APP_GIT_VERSION=.*/VITE_REACT_APP_GIT_VERSION=$$GIT_VERSION/" .env && yarn build

scan: ## scans the repository for CVEs
	@scripts/scan.sh

vuln: ## run go vulnerabilities check
	@scripts/vuln.sh

grype: ## run grype vul scan on sources
	@scripts/grype.sh

yarnsetup: ## set up yarn v4 correctly
	cd web/lxkns && \
	rm -f .yarnrc.yml && \
	rm -rf .yarn/ && \
	rm -rf node_modules && \
	yarn set version berry && \
	yarn config set nodeLinker node-modules && \
	yarn install && \
	yarn eslint --init
