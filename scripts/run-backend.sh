#!/bin/sh
set -eu

if [ -f .env.local ]; then
  set -a
  . ./.env.local
  set +a
fi

go run ./cmd/agentdock-api
