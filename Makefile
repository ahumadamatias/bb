BINARY := bb
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X main.version=$(VERSION)

.PHONY: build
build:
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./cmd/bb

.PHONY: test
test:
	go test ./...

.PHONY: lint
lint:
	go vet ./...
	@fmtdiff="$$(gofmt -l .)"; \
	if [ -n "$$fmtdiff" ]; then \
		echo "gofmt needs to be run on:"; \
		echo "$$fmtdiff"; \
		exit 1; \
	fi

.PHONY: install
install:
	go install -ldflags "$(LDFLAGS)" ./cmd/bb

.PHONY: clean
clean:
	rm -rf bin dist
