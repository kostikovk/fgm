APP := fgm
BIN_DIR := tmp
BIN := $(BIN_DIR)/$(APP)
CMD ?= --help

.PHONY: help build run cmd test fmt fix update-lint-compat clean

help:
	@echo "Available targets:"
	@echo "  make build  - build $(APP) into $(BIN)"
	@echo "  make run    - run the CLI with --help"
	@echo "  make cmd    - run the built CLI, override with CMD='current --chdir .'"
	@echo "  make test   - run all tests"
	@echo "  make fmt    - format Go files"
	@echo "  make fix    - apply Go fixes, format Go files, and run tests"
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

update-lint-compat:
	go run ./scripts/update-lint-compatibility

clean:
	rm -rf $(BIN_DIR)
