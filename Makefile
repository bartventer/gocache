.SHELLFLAGS = -ecuo pipefail
SHELL = /bin/bash
.PHONY: test lint

COVERPROFILE ?= coverage.out

test:
	go test -v -race -coverprofile=$(COVERPROFILE) -covermode=atomic -timeout 15m $(shell go list ./...)

lint:
	golangci-lint run ./... --timeout 5m --fix --verbose