.SHELLFLAGS = -ecuo pipefail
SHELL = /bin/bash
.PHONY: test up down lint build

test:
	go test -v ./...

up:
	docker-compose up -d
	./scripts/check_services.sh || (make down && exit 1)

down:
	docker-compose down --volumes --remove-orphans

lint:
	golangci-lint run ./... --timeout 5m --fix

build:
	go build -o main .