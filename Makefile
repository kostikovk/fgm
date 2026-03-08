APP := fgm
BIN_DIR := tmp
BIN := $(BIN_DIR)/$(APP)
CMD ?= --help

.PHONY: help build run cmd test fmt fix pre-commit hook-install update-lint-compat clean

help:
	@echo "Available targets:"
	@echo "  make build  - build $(APP) into $(BIN)"
	@echo "  make run    - run the CLI with --help"
	@echo "  make cmd    - run the built CLI, override with CMD='current --chdir .'"
	@echo "  make test   - run all tests"
	@echo "  make fmt    - format Go files"
	@echo "  make fix    - apply Go fixes, format Go files, and run tests"
	@echo "  make pre-commit - run the local pre-commit checks"
	@echo "  make hook-install - sync Hooky hooks into .git/hooks"
	@echo "  make update-lint-compat - regenerate golangci-lint compatibility.json"
	@echo "  make clean  - remove local build artifacts"

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN) .

run:
	go run . --help

cmd: build
	$(BIN) $(CMD)

test:
	go test ./...

fmt:
	go fmt ./...

fix:
	go fix ./...
	go fmt ./...
	go test ./...

pre-commit: fix build

hook-install:
	hooky init

update-lint-compat:
	go run ./scripts/update-lint-compatibility

clean:
	rm -rf $(BIN_DIR)
