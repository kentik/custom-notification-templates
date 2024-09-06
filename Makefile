.PHONY: all docs test

all: test docs

docs:
	go run ./cmd/docs

test:
	go test ./...

