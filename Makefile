APP := fgm
BIN_DIR := tmp
BIN := $(BIN_DIR)/$(APP)
CMD ?= --help
COVERPROFILE := $(BIN_DIR)/coverage.out
COVERHTML := $(BIN_DIR)/coverage.html
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w -X main.buildVersion=$(VERSION) -X main.buildCommit=$(COMMIT) -X main.buildDate=$(DATE)

.PHONY: help build run cmd test cover cover-html fmt fix tidy pre-commit hook-install update-lint-compat clean

help:
	@echo "Available targets:"
	@echo "  make build  - build $(APP) into $(BIN)"
	@echo "  make run    - run the CLI with --help"
	@echo "  make cmd    - run the built CLI, override with CMD='current --chdir .'"
	@echo "  make test   - run all tests"
	@echo "  make cover  - run all tests with a coverage profile"
	@echo "  make cover-html - generate an HTML coverage report at $(COVERHTML)"
	@echo "  make fmt    - format Go files"
	@echo "  make fix    - apply Go fixes and format Go files"
	@echo "  make tidy   - clean up go.mod and go.sum"
	@echo "  make pre-commit - run the local pre-commit checks"
	@echo "  make hook-install - sync Hooky hooks into .git/hooks"
	@echo "  make update-lint-compat - regenerate golangci-lint compatibility.json"
	@echo "  make clean  - remove local build artifacts"

build:
	@mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BIN) .

run:
	go run . --help

cmd: build
	$(BIN) $(CMD)

test:
	go test ./...

cover:
	@mkdir -p $(BIN_DIR)
	go test ./... -coverprofile=$(COVERPROFILE)
	go tool cover -func=$(COVERPROFILE)

cover-html: cover
	go tool cover -html=$(COVERPROFILE) -o $(COVERHTML)
	@echo "HTML coverage report written to $(COVERHTML)"

fmt:
	go fmt ./...

fix:
	go fix ./...
	go fmt ./...

tidy:
	go mod tidy

pre-commit: fix test tidy build

hook-install:
	hooky init

update-lint-compat:
	go run ./scripts/update-lint-compatibility

clean:
	rm -rf $(BIN_DIR)
