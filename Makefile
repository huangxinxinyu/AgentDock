SHELL := /bin/sh
GOCACHE ?= $(CURDIR)/.cache/go-build
GOMODCACHE ?= $(CURDIR)/.cache/go-mod

.PHONY: install deps-up deps-down backend frontend go-test frontend-test test go-build frontend-build build format format-check lint typecheck migrations-check ci

install:
	cd web && npm install

deps-up:
	docker compose -f deploy/docker-compose.yml up -d

deps-down:
	docker compose -f deploy/docker-compose.yml down

backend:
	env GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) sh scripts/run-backend.sh

frontend:
	cd web && npm run dev

go-test:
	env GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) sh scripts/go-test.sh

frontend-test:
	cd web && npm test

test: go-test frontend-test

go-build:
	mkdir -p .cache/bin
	env GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) go build -o .cache/bin/agentdock-api ./cmd/agentdock-api

frontend-build:
	cd web && npm run build

build: go-build frontend-build

format:
	gofmt -w cmd internal
	cd web && npm run format

format-check:
	sh scripts/check-gofmt.sh
	cd web && npm run format:check

lint:
	cd web && npm run lint

typecheck:
	cd web && npm run typecheck

migrations-check:
	sh scripts/check-migrations.sh

ci: format-check lint typecheck migrations-check test build
