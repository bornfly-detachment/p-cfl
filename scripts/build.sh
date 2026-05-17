#!/usr/bin/env sh
set -eu
ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
OUT=${1:-"$ROOT/bin/p-cfl"}
mkdir -p "$(dirname "$OUT")"
(cd "$ROOT" && go build -o "$OUT" .)
printf 'built %s\n' "$OUT"
