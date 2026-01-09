GO ?= go

DIST_DIR := dist
WASM_OUT := $(DIST_DIR)/renderer.wasm
WASM_EXEC_OUT := $(DIST_DIR)/wasm_exec.js


.PHONY: all docs test test-go test-wasm dist wasm

all: test docs wasm

clean:
	rm -rf $(DIST_DIR)

docs:
	go run ./cmd/docs

test: test-go test-wasm

test-go:
	go test ./pkg/...

test-wasm: wasm
	node node_wasm_test.js

dist:
	mkdir -p $(DIST_DIR)

wasm: dist
	GOOS=js GOARCH=wasm $(GO) build -o $(WASM_OUT) ./cmd/wasm
	cp "$$($(GO) env GOROOT)/lib/wasm/wasm_exec.js" $(WASM_EXEC_OUT)
