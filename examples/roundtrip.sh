#!/usr/bin/env sh
set -eu
ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
BIN=${BIN:-"$ROOT/bin/p-cfl"}
STORE=${STORE:-"$ROOT/.p-cfl-state/example"}
GOCACHE=${GOCACHE:-/private/tmp/p-cfl-go-cache} "$ROOT/scripts/build.sh" "$BIN" >/dev/null
mkdir -p "$STORE"
ENC=$(printf '%s\n' '{"api":"encode","store_dir":"'$STORE'","input":{"statement":"Pattern CFL preserves information density and hash-chain consistency","kind":"example"},"from":{"type":"model","model_id":"demo","prompt_hash":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","call_ts":1},"where":{"project":"demo","session":"roundtrip"}}' | "$BIN")
echo "$ENC"
REF=$(printf '%s' "$ENC" | sed -n 's/.*"pattern_ref":"\([^"]*\)".*/\1/p')
printf '%s\n' '{"api":"decode","store_dir":"'$STORE'","pattern_ref":"'$REF'"}' | "$BIN"
printf '%s\n' '{"api":"measure_info_density","store_dir":"'$STORE'","pattern_ref":"'$REF'"}' | "$BIN"
