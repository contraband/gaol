#!/usr/bin/env bash

set -e

if [ -z "$GAOL_TARGET" ]; then
  echo 'please set $GAOL_TARGET to an empty garden server'
  exit 1
fi

GAOL_TMPDIR=$(mktemp -d)
trap 'rm -r "$GAOL_TMPDIR"' EXIT

go build -o $GAOL_TMPDIR/gaol github.com/contraband/gaol
export PATH=$BATS_TMPDIR/bin:$PATH

bats test/
