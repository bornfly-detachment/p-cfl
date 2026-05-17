#!/usr/bin/env sh
# Tamper detection example: 改 JSONL 文件后 decode 仍能取出 but chain_valid=false
# 演示矛盾 ④：append-only + hash 链对篡改可检测（虽然 MVP 期不可阻挡物理修改）
set -eu
ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
BIN=${BIN:-"$ROOT/bin/p-cfl"}
STORE=${STORE:-"$ROOT/.p-cfl-state/tamper"}
GOCACHE=${GOCACHE:-/private/tmp/p-cfl-go-cache} "$ROOT/scripts/build.sh" "$BIN" >/dev/null
rm -rf "$STORE" && mkdir -p "$STORE"

# Step 1: 写一个 Pattern
ENC=$(printf '%s\n' '{"api":"encode","store_dir":"'$STORE'","input":{"statement":"信息时空一致性和不可篡改性是 Pattern CFL 第四主要矛盾"},"from":{"type":"model","model_id":"demo","prompt_hash":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","call_ts":1},"where":{"project":"tamper-demo"}}' | "$BIN")
echo "=== step 1: encode ==="
echo "$ENC"
REF=$(printf '%s' "$ENC" | sed -n 's/.*"pattern_ref":"\([^"]*\)".*/\1/p')

# Step 2: 篡改 JSONL 文件（模拟攻击者直接改存储）
echo ""
echo "=== step 2: tamper jsonl directly ==="
sed -i.bak 's/Pattern CFL/Pattern XXX/' "$STORE/patterns.jsonl"
echo "(patched 'Pattern CFL' -> 'Pattern XXX')"

# Step 3: decode — verdict 仍 pass 但 chain_valid=false（hash 链不一致被检出）
echo ""
echo "=== step 3: decode after tamper (expect chain_valid=false) ==="
printf '%s\n' '{"api":"decode","store_dir":"'$STORE'","pattern_ref":"'$REF'"}' | "$BIN" || true
