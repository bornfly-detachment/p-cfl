package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

func validateFrom(from FromSource, what any) error {
	switch from.Type {
	case "bornfly":
		if from.Signature == "" || from.PubKey == "" {
			return errors.New("bornfly source requires pub_key and signature")
		}
		pub, err := hex.DecodeString(from.PubKey)
		if err != nil || len(pub) != ed25519.PublicKeySize {
			return errors.New("invalid bornfly pub_key")
		}
		sig, err := hex.DecodeString(from.Signature)
		if err != nil || len(sig) != ed25519.SignatureSize {
			return errors.New("invalid bornfly signature")
		}
		if !ed25519.Verify(ed25519.PublicKey(pub), canonicalBytes(what), sig) {
			return errors.New("bornfly signature verification failed")
		}
	case "cfl":
		if from.CFLID == "" || from.Cert == "" {
			return errors.New("cfl source requires cfl_id and cert")
		}
		if from.Layer != "L0" && from.Layer != "L1" && from.Layer != "L2" {
			return errors.New("cfl layer must be L0|L1|L2")
		}
	case "model":
		if from.ModelID == "" || from.CallTS <= 0 {
			return errors.New("model source requires model_id and call_ts")
		}
		if !strings.HasPrefix(from.PromptHash, "sha256:") {
			return errors.New("model source requires sha256 prompt_hash")
		}
	case "v-check":
		if from.VCFLID == "" || !strings.HasPrefix(from.RecordHash, "sha256:") {
			return errors.New("v-check source requires v_cfl_id and sha256 record_hash")
		}
	default:
		return fmt.Errorf("unknown from.type %q", from.Type)
	}
	return nil
}

func initialQualification(from FromSource) string {
	if from.Type == "bornfly" {
		return "internal"
	}
	return "candidate"
}
