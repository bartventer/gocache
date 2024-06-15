.SHELLFLAGS = -ecuo pipefail
SHELL = /bin/bash
.PHONY: test lint

test:
	./scripts/test.sh

update:
	./scripts/update.sh

lint:
	golangci-lint run ./... --timeout 5m --fix --verbose