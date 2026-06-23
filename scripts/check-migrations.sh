#!/bin/sh
set -eu

for up in db/migrations/*.up.sql; do
  down="${up%.up.sql}.down.sql"
  if [ ! -f "$down" ]; then
    echo "missing down migration for $up" >&2
    exit 1
  fi
done
