#!/bin/sh
set -eu

files="$(gofmt -l cmd internal)"
if [ -n "$files" ]; then
  printf '%s\n' "$files"
  exit 1
fi
