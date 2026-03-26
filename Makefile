BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo Unknown)
VERSION    := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo Unknown)

BINARY     := notmuch2maildir
BUILD_FILE := ./cmd/notmuch2maildir/
LDFLAGS    := -w -s \
	-X main.buildDate=$(BUILD_DATE) \
	-X main.gitCommit=$(GIT_COMMIT) \
	-X main.version=$(VERSION)

GOFILES := $(shell find . -name "*.go" -not -path "./.git/*")

.PHONY: all build install test fmt clean ci

all: build

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) $(BUILD_FILE)

install:
	go install -ldflags "$(LDFLAGS)" $(BUILD_FILE)

test:
	go test ./...

fmt:
	@out=$$(gofmt -l $(GOFILES)); if [ -n "$$out" ]; then echo "gofmt issues in:"; echo "$$out"; exit 1; fi

clean:
	rm -f $(BINARY)
	go clean

ci: fmt test build clean
