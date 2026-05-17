#!/usr/bin/env sh
# Information-theoretic metrics example: 矛盾 ② (KL) + ③ (Shannon entropy)
# 演示信息密度阈值（低熵 input 直接拒收）+ KL 散度计算
set -eu
ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
BIN=${BIN:-"$ROOT/bin/p-cfl"}
STORE=${STORE:-"$ROOT/.p-cfl-state/metrics"}
GOCACHE=${GOCACHE:-/private/tmp/p-cfl-go-cache} "$ROOT/scripts/build.sh" "$BIN" >/dev/null
rm -rf "$STORE" && mkdir -p "$STORE"

# Step 1: 低熵 input（噪声/空话）→ encode 物理拒收（矛盾 ③）
echo "=== step 1: low-entropy encode (expect fail) ==="
printf '%s\n' '{"api":"encode","store_dir":"'$STORE'","input":"aaaaaaaaaa","from":{"type":"model","model_id":"demo","prompt_hash":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","call_ts":1},"where":{"project":"metrics-demo"}}' | "$BIN" || true

# Step 2: 正常 input 通过
echo ""
echo "=== step 2: healthy encode (expect pass) ==="
ENC=$(printf '%s\n' '{"api":"encode","store_dir":"'$STORE'","input":{"statement":"Pattern CFL 通过 Shannon 熵和 KL 散度度量信息论 KPI"},"from":{"type":"model","model_id":"demo","prompt_hash":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","call_ts":1},"where":{"project":"metrics-demo"}}' | "$BIN")
echo "$ENC"
REF=$(printf '%s' "$ENC" | sed -n 's/.*"pattern_ref":"\([^"]*\)".*/\1/p')

# Step 3: 信息密度度量（矛盾 ③）
echo ""
echo "=== step 3: measure_info_density ==="
printf '%s\n' '{"api":"measure_info_density","store_dir":"'$STORE'","pattern_ref":"'$REF'"}' | "$BIN"

# Step 4: KL 散度（矛盾 ②）— receiver 显式传 prior
echo ""
echo "=== step 4: measure_uncertainty (KL divergence) ==="
printf '%s\n' '{"api":"measure_uncertainty","store_dir":"'$STORE'","pattern_ref":"'$REF'","receiver_prior":{"pattern":0.6,"cfl":0.3,"unknown":0.1}}' | "$BIN"
