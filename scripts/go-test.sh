#!/bin/sh
set -eu

go list ./... | grep -v '/web/node_modules/' | xargs go test
