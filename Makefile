# Where to install the CLI tool binaries to
PREFIX ?= /usr/local
GOPATH = $(shell go env GOPATH)
GOGEN = go generate .

GET_SEMVERSION = awk '/const SemVersion/ { sub(/.*= *"/,""); sub(/".*/,""); print; exit }' defs_version.go

export GOTOOLCHAIN=local

.PHONY: clean deploy undeploy help buildapp startapp scan dist grype yarnsetup

help: ## list available targets
	@# Shamelessly stolen from Gomega's Makefile
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}'

clean: ## cleans up build and testing artefacts
	rm -f $(tools)
	rm -f coverage.html coverage.out coverage-root.out
	rm -f coverage.txt coverage-root.txt

deploy: ## deploys lxkns service on host port 5010
	$(GOGEN)
	$(eval GIT_VERSION := $(shell $(GET_SEMVERSION)))
	echo $(GIT_VERSION)
	scripts/docker-build.sh deployments/lxkns/Dockerfile \
		-t lxkns \
		--build-arg REACT_APP_GIT_VERSION="$(GIT_VERSION)" \
		--build-context webappsrc=./web/lxkns
	docker compose -p lxkns -f deployments/lxkns/docker-compose.yaml up

undeploy: ## removes any deployed lxkns service
	docker compose -p lxkns -f deployments/lxkns/docker-compose.yaml down

buildapp: ## builds web UI app
	$(GOGEN)
	$(eval GIT_VERSION := $(shell $(GET_SEMVERSION)))
	@echo "building version" $(GIT_VERSION)
	@cd web/lxkns && sed -i "s/^VITE_REACT_APP_GIT_VERSION=.*/VITE_REACT_APP_GIT_VERSION=$$GIT_VERSION/" .env && yarn build

startapp: ## starts web UI app for development
	$(GOGEN)
	$(eval GIT_VERSION := $(shell $(GET_SEMVERSION)))
	@echo "starting version" $(GIT_VERSION)
	@cd web/lxkns && sed -i "s/^VITE_REACT_APP_GIT_VERSION=.*/VITE_REACT_APP_GIT_VERSION=$$GIT_VERSION/" .env && yarn start
