BUF ?= $(shell go env GOPATH)/bin/buf
export PATH := $(shell go env GOPATH)/bin:$(PATH)

.PHONY: tools lint generate test ci sync-docs

tools:
	go install github.com/bufbuild/buf/cmd/buf@v1.50.0
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
	go install connectrpc.com/connect/cmd/protoc-gen-connect-go@v1.19.1
	go install github.com/sudorandom/protoc-gen-connect-openapi@v0.25.5

lint:
	$(BUF) lint

generate:
	$(BUF) generate
	./scripts/sync-api-snapshot.sh --finalize-openapi-only

test:
	go test ./...

ci: lint generate test

# Copy the spec into the metalhost-web repo for the /docs/api viewer.
# Usage: make sync-docs DOCS=../metalhost-web
DOCS ?= ../metalhost-web
sync-docs: generate
	@test -d "$(DOCS)/public" || { echo >&2 "$(DOCS)/public not found — set DOCS=path/to/metalhost-web"; exit 1; }
	cp gen/openapi/metalhost.openapi.yaml $(DOCS)/public/openapi.yaml
	@echo "synced openapi.yaml to $(DOCS)/public/openapi.yaml"
