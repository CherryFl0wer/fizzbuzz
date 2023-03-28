VERSION = $(shell git describe --tag --always --dirty)
PROJECT := fizzbuzz
MOCKS := $(shell grep -l 'go:generate' repository/*.go | sed -e "s?repository/\(.*\).go?repository/mock/\1.go?") $(shell grep -l 'go:generate' service/*.go | sed -e "s?service/\(.*\).go?service/mock/\1.go?")
GOLANGCI := .deps/golangci-lint
TESTTIMEOUT := 60s

$(GOLANGCI): ## Install linter
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./.deps v1.49.0

.DEFAULT_GOAL := build

.PHONY: setup
setup: ## Install all the build and lint dependencies
	go get -u golang.org/x/tools/cmd/cover

.PHONY: fmt
fmt: ## Run goimports on all go files
	find . -name '*.go' -not -wholename './vendor/*' | while read -r file; do goimports -w "$$file"; done

.PHONY: lint
lint: $(GOLANGCI) version.txt ## Run the linter
	$(GOLANGCI) run

.PHONY: main
main:
	go build -v ./cmd/main

run-main:
	go run ./cmd/main

.PHONY: build
build: version.txt main ## Build a version


.PHONY: staticbuild
staticbuild: version.txt ## Build a static version
	CGO_ENABLED=0 go build -ldflags '-extldflags "-static"' -v ./cmd/...

.PHONY: clean
clean: ## Remove temporary files
	go clean

.PHONY: install
install: ## install project and it's dependencies, useful for autocompletion feature
	go install -i

.PHONY: version
version: ## display version
	@echo $(VERSION)

.PHONY: docker
docker: build ## build docker image
	docker build -t $(PROJECT):$(VERSION) .

.PHONY: monitoring
monitoring: ## run monitoring stack prometheus
	docker-compose up -d prometheus

.PHONY: cache
cache: ## run cache stack
	docker-compose up -d redis


.PHONY: start
start: $(GOLANGCI) monitoring cache ## Start the whole project
	docker-compose up $(PROJECT)


install-gomock:
	GOBIN=$(PWD)/.deps go install github.com/golang/mock/mockgen@v1.6.0

repository/mock/%.go: install-gomock repository/%.go
	go generate ./...

generate-mocks: $(MOCKS)

.PHONY: tests
tests: $(MOCKS)
	go test -covermode=atomic -coverprofile=coverage.txt -race -timeout=$(TESTTIMEOUT) ./...

version.txt: get_version.sh
	go generate

# Absolutely awesome: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
