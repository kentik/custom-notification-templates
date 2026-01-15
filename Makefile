GO ?= go

DIST_DIR := dist
WASM_OUT := $(DIST_DIR)/renderer.wasm
WASM_EXEC_OUT := $(DIST_DIR)/wasm_exec.js


.PHONY: all docs test test-go test-wasm dist wasm generate

all: generate test docs wasm

clean:
	rm -rf $(DIST_DIR)

generate:
	@echo "Generating code from doc comments..."
	go generate ./pkg/render
	go fmt ./pkg/render
	@echo "Code generation complete"

docs:
	go run ./cmd/docs

test: test-go test-wasm

test-go: generate
	go test ./...

test-wasm: wasm
	node ./wasm_support/integration_tests.js

dist:
	mkdir -p $(DIST_DIR)

wasm: dist generate
	GOOS=js GOARCH=wasm $(GO) build -o $(WASM_OUT) ./cmd/wasm
	cp wasm_support/wasm_exec.js $(WASM_EXEC_OUT)
