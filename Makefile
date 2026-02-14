.PHONY: build test test-integration test-interactive test-automated test-all clean proto proto-force submodule submodule-update lint audit tidy install-lint-tools install-buf

# --- Proto generation ---
XAI_PROTO := xai-proto
BUF := $(shell which buf 2>/dev/null || echo "go run github.com/bufbuild/buf/cmd/buf@latest")
GEN_SENTINEL := proto/.generated
PROTO_SOURCES := $(shell find $(XAI_PROTO)/proto -name '*.proto' 2>/dev/null)

# Sentinel-based generation: only regenerate if sources changed
$(GEN_SENTINEL): $(PROTO_SOURCES) buf.gen.go.yaml | submodule
	@echo "Generating Go code from xai-proto..."
	cd $(XAI_PROTO) && $(BUF) generate --template ../buf.gen.go.yaml -o ..
	@touch $@

proto: $(GEN_SENTINEL)

# Force regeneration regardless of timestamps
proto-force: submodule
	@echo "Generating Go code from xai-proto..."
	cd $(XAI_PROTO) && $(BUF) generate --template ../buf.gen.go.yaml -o ..
	@touch $(GEN_SENTINEL)

# --- Build ---
build: $(GEN_SENTINEL)
	go build ./...

# --- Tests ---
# Harnessed tests (no API key required)
test: $(GEN_SENTINEL)
	go test -v ./tests/...

# Live tests (require XAI_APIKEY; skip via t.Skip when unset)
# -count=1 disables caching to ensure live API calls
test-integration: $(GEN_SENTINEL)
	go test -count=1 -v ./integration/...

# Interactive chat REPL for manual testing
test-interactive: $(GEN_SENTINEL)
	go run ./cmd/minimal-client

# Run automated API verification tests
test-automated: $(GEN_SENTINEL)
	go run ./cmd/minimal-client -test

# Run all tests (harnessed + live)
test-all: test test-integration

# --- Maintenance ---
clean:
	rm -rf proto/

tidy:
	go mod tidy

# --- Submodule ---
submodule:
	@if [ ! -f "$(XAI_PROTO)/buf.yaml" ]; then \
		echo "Initializing xai-proto submodule..."; \
		git submodule update --init --recursive $(XAI_PROTO); \
	fi

submodule-update: submodule
	git submodule update --remote --merge $(XAI_PROTO)
	@echo "Run 'make proto-force' to regenerate after submodule update"

install-buf:
	go install github.com/bufbuild/buf/cmd/buf@latest
	@echo "buf installed. Run 'make proto' to generate."

# --- Code quality ---
GOLANGCI_LINT := $(shell which golangci-lint 2>/dev/null)
GOVULNCHECK := $(shell which govulncheck 2>/dev/null)

install-lint-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest

lint: $(GEN_SENTINEL)
ifndef GOLANGCI_LINT
	@echo "Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
endif
	golangci-lint run ./...

audit: lint
ifndef GOVULNCHECK
	@echo "Installing govulncheck..."
	@go install golang.org/x/vuln/cmd/govulncheck@latest
endif
	govulncheck ./...
