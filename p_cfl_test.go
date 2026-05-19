package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEncodeDecodeDeriveRoundTrip(t *testing.T) {
	store := t.TempDir()
	from := modelSource()
	where := SpaceContext{Project: "prvse", Branch: "main", Session: "test"}
	what := map[string]any{"statement": "Pattern CFL preserves information density and hash-chain consistency", "kind": "pattern"}

	enc := handle(Request{API: "encode", StoreDir: store, Input: what, From: from, Where: where}, time.Now())
	if enc.Verdict != "pass" || enc.PatternRef == "" || enc.RecordHash == "" {
		t.Fatalf("encode failed: %#v", enc)
	}

	dec := handle(Request{API: "decode", StoreDir: store, PatternRef: enc.PatternRef}, time.Now())
	if dec.Verdict != "pass" || !dec.ChainValid {
		t.Fatalf("decode failed: %#v", dec)
	}
	got, _ := json.Marshal(canonical(dec.Output))
	want, _ := json.Marshal(canonical(what))
	if !bytes.Equal(got, want) {
		t.Fatalf("decode mismatch got=%s want=%s", got, want)
	}

	derivedWhat := map[string]any{"statement": "Derived Pattern structures the original without replacing it", "kind": "derived-pattern"}
	der := handle(Request{API: "derive", StoreDir: store, ParentRef: enc.PatternRef, NewContent: derivedWhat, ModifyType: "structuring", From: from}, time.Now())
	if der.Verdict != "pass" || der.PatternRef == enc.PatternRef {
		t.Fatalf("derive failed: %#v", der)
	}
	old := handle(Request{API: "decode", StoreDir: store, PatternRef: enc.PatternRef}, time.Now())
	if old.Verdict != "pass" {
		t.Fatalf("old pattern not preserved: %#v", old)
	}
	newDec := handle(Request{API: "decode", StoreDir: store, PatternRef: der.PatternRef}, time.Now())
	if newDec.Metadata["parent_ref"] != enc.PatternRef || newDec.Metadata["modify_type"] != "structuring" {
		t.Fatalf("derive metadata missing: %#v", newDec.Metadata)
	}
}

func TestFourElementsAndNoMutationAPI(t *testing.T) {
	store := t.TempDir()
	from := modelSource()
	what := map[string]any{"statement": "Enough information to pass density threshold"}
	missingWhere := handle(Request{API: "encode", StoreDir: store, Input: what, From: from}, time.Now())
	if missingWhere.Verdict != "fail" || !strings.Contains(missingWhere.Error, "where.project") {
		t.Fatalf("expected where failure: %#v", missingWhere)
	}
	missingFrom := handle(Request{API: "encode", StoreDir: store, Input: what, Where: SpaceContext{Project: "p"}}, time.Now())
	if missingFrom.Verdict != "fail" {
		t.Fatalf("expected from failure: %#v", missingFrom)
	}
	update := handle(Request{API: "update", StoreDir: store}, time.Now())
	if update.Verdict != "fail" || !strings.Contains(update.Error, "does not exist") {
		t.Fatalf("update should not exist: %#v", update)
	}
}

func TestInformationMetrics(t *testing.T) {
	store := t.TempDir()
	from := modelSource()
	what := map[string]any{"alpha": "Pattern information carries alpha beta gamma delta", "beta": "uncertainty reduction"}
	enc := handle(Request{API: "encode", StoreDir: store, Input: what, From: from, Where: SpaceContext{Project: "metric"}}, time.Now())
	if enc.Verdict != "pass" {
		t.Fatalf("encode failed: %#v", enc)
	}
	density := handle(Request{API: "measure_info_density", StoreDir: store, PatternRef: enc.PatternRef, EntropyThreshold: 2.0}, time.Now())
	if density.Verdict != "pass" || density.Entropy < 2.0 || !density.Passed {
		t.Fatalf("density failed: %#v", density)
	}
	low := handle(Request{API: "measure_info_density", Input: "aaaaaa", EntropyThreshold: 1.0}, time.Now())
	if low.Verdict != "fail" || low.Passed {
		t.Fatalf("low entropy should fail: %#v", low)
	}
	uncertainty := handle(Request{API: "measure_uncertainty", StoreDir: store, PatternRef: enc.PatternRef, ReceiverPrior: map[string]float64{"pattern": 0.9, "information": 0.1}}, time.Now())
	if uncertainty.Verdict != "pass" || uncertainty.KLDiv <= 0 {
		t.Fatalf("uncertainty failed: %#v", uncertainty)
	}
}

func TestFromAuthenticationModes(t *testing.T) {
	store := t.TempDir()
	where := SpaceContext{Project: "auth"}
	what := map[string]any{"statement": "Authentication mode has sufficient information density for Pattern CFL"}
	sources := []FromSource{bornflySource(t, what), {Type: "human"}, {Type: "cfl", CFLID: "p-cfl-test", Layer: "L0", Cert: "sha256:cert"}, modelSource(), {Type: "v-check", VCFLID: "l0-v-hash-cfl", RecordHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}
	for _, src := range sources {
		resp := handle(Request{API: "encode", StoreDir: store, Input: what, From: src, Where: where}, time.Now())
		if resp.Verdict != "pass" {
			t.Fatalf("source %s failed: %#v", src.Type, resp)
		}
	}
	fakeBornfly := FromSource{Type: "bornfly", PubKey: sources[0].PubKey, Signature: strings.Repeat("0", 128)}
	resp := handle(Request{API: "encode", StoreDir: store, Input: what, From: fakeBornfly, Where: where}, time.Now())
	if resp.Verdict != "fail" {
		t.Fatalf("fake bornfly accepted: %#v", resp)
	}
}

func TestHashChainTamperDetection(t *testing.T) {
	store := t.TempDir()
	from := modelSource()
	where := SpaceContext{Project: "tamper"}
	what := map[string]any{"statement": "Hash chain detects tampering with stored Pattern content"}
	enc := handle(Request{API: "encode", StoreDir: store, Input: what, From: from, Where: where}, time.Now())
	if enc.Verdict != "pass" {
		t.Fatal(enc)
	}
	path := filepath.Join(store, "patterns.jsonl")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	b = bytes.Replace(b, []byte("tampering"), []byte("altering"), 1)
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
	dec := handle(Request{API: "decode", StoreDir: store, PatternRef: enc.PatternRef}, time.Now())
	if dec.Verdict != "pass" || dec.ChainValid {
		t.Fatalf("tamper not detected: %#v", dec)
	}
}

func TestCLIJSON(t *testing.T) {
	store := t.TempDir()
	payload := map[string]any{"api": "measure_info_density", "store_dir": store, "input": "Pattern CFL command line emits deterministic JSON", "entropy_threshold": 1.0}
	b, _ := json.Marshal(payload)
	cmd := exec.Command("go", "run", ".")
	cmd.Env = append(os.Environ(), "GOCACHE=/private/tmp/p-cfl-go-cache")
	cmd.Stdin = bytes.NewReader(b)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go run failed: %v output=%s", err, out)
	}
	var resp Response
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("invalid json: %s", out)
	}
	if resp.Verdict != "pass" || resp.EvidenceHash == "" {
		t.Fatalf("bad cli response: %#v", resp)
	}
}

func modelSource() FromSource {
	return FromSource{Type: "model", ModelID: "deterministic-model", PromptHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", CallTS: 1}
}

func bornflySource(t *testing.T, what any) FromSource {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	sig := ed25519.Sign(priv, canonicalBytes(what))
	return FromSource{Type: "bornfly", PubKey: hex.EncodeToString(pub), Signature: hex.EncodeToString(sig)}
}
