BINARY  := notmuch2maildir
CMD_DIR := ./cmd/notmuch2maildir
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -buildid= -X main.version=$(VERSION)
TAGS    := osusergo,netgo

.DEFAULT_GOAL := help

.PHONY: build install fmt fmt-fix tidy vet lint test test-ci clean check help

## build: Build the binary
build:
	CGO_ENABLED=0 \
	go build \
	-trimpath \
	-ldflags '$(LDFLAGS)' \
	-tags '$(TAGS)' \
	-o $(BINARY) $(CMD_DIR)

## install: Install the binary via go install
install:
	go install \
	-trimpath \
	-ldflags '$(LDFLAGS)' \
	-tags '$(TAGS)' \
	$(CMD_DIR)

## fmt: Check formatting (fails on diff)
fmt:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "Unformatted files:"; echo "$$unformatted"; exit 1; \
	fi

## fmt-fix: Fix Go formatting (destructive)
fmt-fix:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "Fixing unformatted files..."; \
		gofmt -s -w .; \
	fi

## tidy: Check that go.mod and go.sum are tidy
tidy:
	go mod tidy -diff

## vet: Run go vet
vet:
	go vet ./...

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## test: Run tests (short mode, race, 2m timeout)
test:
	go test -race -short -timeout 2m ./...

## test-ci: Run tests (full, race, 5m timeout)
test-ci:
	go test -race -timeout 5m ./...

## check: Run all quality gates (fmt tidy vet lint test)
check: fmt tidy vet lint test

## clean: Remove build artifacts
clean:
	rm -f $(BINARY)

## help: Show this help
help:
	@echo "notmuch2maildir Makefile"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //' | column -t -s ':'
