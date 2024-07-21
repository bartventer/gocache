.SHELLFLAGS = -ecuo pipefail
SHELL = /usr/bin/env bash

.PHONY: test
test:
	./scripts/test.sh

.PHONY: update
update:
	./scripts/update.sh

.PHONY: lint
lint:
	golangci-lint run ./... --timeout 5m --fix --verbose

.PHONY: release
release:
	./scripts/release.sh