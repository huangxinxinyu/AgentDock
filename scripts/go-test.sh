#!/bin/sh
set -eu

go test ./cmd/... ./internal/...
