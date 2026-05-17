#!/usr/bin/env sh
# Derive chain example: external信息 ingest → candidate → 衍生新 candidate
# 演示"取代 ≠ 改"：父 Pattern 永存于台账，衍生新 Pattern 携带 modify_type 标签
set -eu
ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
BIN=${BIN:-"$ROOT/bin/p-cfl"}
STORE=${STORE:-"$ROOT/.p-cfl-state/derive"}
GOCACHE=${GOCACHE:-/private/tmp/p-cfl-go-cache} "$ROOT/scripts/build.sh" "$BIN" >/dev/null
rm -rf "$STORE" && mkdir -p "$STORE"

# Step 1: external 信息 ingest（外部 GitHub 代码片段，起始 candidate per SCOPE.md v1）
PARENT=$(printf '%s\n' '{"api":"encode","store_dir":"'$STORE'","input":{"snippet":"function handler(req,res){return res.json({ok:true})}","language":"javascript","origin_url":"https://github.com/example/repo/blob/main/handler.js"},"from":{"type":"model","model_id":"external-fetch","prompt_hash":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","call_ts":1700000000},"where":{"project":"prvse-demo","session":"derive-chain"}}' | "$BIN")
echo "=== step 1: external ingest ==="
echo "$PARENT"
PARENT_REF=$(printf '%s' "$PARENT" | sed -n 's/.*"pattern_ref":"\([^"]*\)".*/\1/p')

# Step 2: structuring 衍生（外部 snippet 结构化为本系统 schema）
echo ""
echo "=== step 2: derive (modify_type=structuring) ==="
printf '%s\n' '{"api":"derive","store_dir":"'$STORE'","parent_ref":"'$PARENT_REF'","new_content":{"handler_name":"handler","params":["req","res"],"returns":{"ok":true}},"modify_type":"structuring","from":{"type":"cfl","cfl_id":"p-cfl-demo","layer":"L1","cert":"sha256:demo-cert"}}' | "$BIN"

# Step 3: 父 Pattern 仍可查（"取代 ≠ 改"）
echo ""
echo "=== step 3: parent still exists ==="
printf '%s\n' '{"api":"decode","store_dir":"'$STORE'","pattern_ref":"'$PARENT_REF'"}' | "$BIN"
