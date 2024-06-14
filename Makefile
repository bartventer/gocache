.SHELLFLAGS = -ecuo pipefail
SHELL = /bin/bash
.PHONY: test up down lint build

COVERPROFILE ?= coverage.out

test:
	go test -v -race -coverprofile=$(COVERPROFILE) -covermode=atomic -timeout 15m $(shell go list ./...)

up:
	docker-compose up -d
	./scripts/check_services.sh || (make down && exit 1)

down:
	docker-compose down --volumes --remove-orphans

lint:
	golangci-lint run ./... --timeout 5m --fix --verbose