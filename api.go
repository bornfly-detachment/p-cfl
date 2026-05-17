package main

import (
	"errors"
	"fmt"
	"time"
)

var allowedModify = map[string]bool{"standalone-modify": true, "structuring": true, "adaptation": true, "mod": true}

func handle(req Request, started time.Time) Response {
	store, err := openStore(req.StoreDir)
	if err != nil {
		return failure(started, "fail", err)
	}
	chainValid, chainErrs, err := store.auditChain()
	if err != nil {
		return failure(started, "fail", err)
	}
	_ = chainErrs
	switch req.API {
	case "encode":
		return encodeAPI(store, req, started, chainValid)
	case "decode":
		return decodeAPI(store, req, started, chainValid)
	case "derive":
		return deriveAPI(store, req, started, chainValid)
	case "measure_uncertainty":
		return uncertaintyAPI(store, req, started, chainValid)
	case "measure_info_density":
		return densityAPI(store, req, started, chainValid)
	case "update", "delete", "lock", "unlock":
		return failure(started, "fail", fmt.Errorf("%s API does not exist; use derive", req.API))
	default:
		return failure(started, "fail", fmt.Errorf("unknown api %q", req.API))
	}
}

func encodeAPI(store *Store, req Request, started time.Time, chainValid bool) Response {
	if req.Input == nil {
		return failure(started, "fail", errors.New("input/what required"))
	}
	if req.Where.Project == "" {
		return failure(started, "fail", errors.New("where.project required"))
	}
	if err := validateFrom(req.From, req.Input); err != nil {
		return failure(started, "fail", err)
	}
	meta := patternMetadata(req.Input, req.From)
	for k, v := range req.Metadata {
		meta[k] = v
	}
	threshold := req.EntropyThreshold
	if threshold == 0 {
		threshold = 2.0
	}
	entropy, density := infoDensity(req.Input)
	if entropy < threshold {
		return metricFailure(started, "fail", fmt.Errorf("information entropy %.6f below threshold %.6f", entropy, threshold), entropy, density, threshold)
	}
	p := StoredPattern{From: req.From, When: time.Now().UTC().UnixNano(), Where: req.Where, What: req.Input, Metadata: meta}
	p.ContentHash = contentHash(p.From, p.Where, p.What, p.Metadata)
	p, err := store.append(p)
	if err != nil {
		return failure(started, "fail", err)
	}
	return success(started, p, chainValid, map[string]any{"api": "encode", "pattern_ref": p.ContentHash, "record_hash": p.RecordHash})
}

func decodeAPI(store *Store, req Request, started time.Time, chainValid bool) Response {
	if req.PatternRef == "" {
		return failure(started, "fail", errors.New("pattern_ref required"))
	}
	p, ok, err := store.find(req.PatternRef)
	if err != nil {
		return failure(started, "fail", err)
	}
	if !ok {
		return failure(started, "n/a", errors.New("pattern_ref not found"))
	}
	resp := success(started, *p, chainValid, map[string]any{"api": "decode", "pattern_ref": p.ContentHash, "record_hash": p.RecordHash})
	resp.Pattern = p
	resp.Output = p.What
	return resp
}

func deriveAPI(store *Store, req Request, started time.Time, chainValid bool) Response {
	if req.ParentRef == "" {
		return failure(started, "fail", errors.New("parent_ref required"))
	}
	if req.NewContent == nil {
		return failure(started, "fail", errors.New("new_content required"))
	}
	if !allowedModify[req.ModifyType] {
		return failure(started, "fail", errors.New("modify_type must be standalone-modify|structuring|adaptation|mod"))
	}
	parent, ok, err := store.find(req.ParentRef)
	if err != nil {
		return failure(started, "fail", err)
	}
	if !ok {
		return failure(started, "n/a", errors.New("parent_ref not found"))
	}
	if err := validateFrom(req.From, req.NewContent); err != nil {
		return failure(started, "fail", err)
	}
	where := req.Where
	if where.Project == "" {
		where = parent.Where
	}
	meta := patternMetadata(req.NewContent, req.From)
	for k, v := range req.Metadata {
		meta[k] = v
	}
	meta["parent_ref"] = parent.ContentHash
	meta["modify_type"] = req.ModifyType
	meta["derive_from_record_hash"] = parent.RecordHash
	threshold := req.EntropyThreshold
	if threshold == 0 {
		threshold = 2.0
	}
	entropy, density := infoDensity(req.NewContent)
	if entropy < threshold {
		return metricFailure(started, "fail", fmt.Errorf("derived information entropy %.6f below threshold %.6f", entropy, threshold), entropy, density, threshold)
	}
	p := StoredPattern{From: req.From, When: time.Now().UTC().UnixNano(), Where: where, What: req.NewContent, Metadata: meta}
	p.ContentHash = contentHash(p.From, p.Where, p.What, p.Metadata)
	p, err = store.append(p)
	if err != nil {
		return failure(started, "fail", err)
	}
	return success(started, p, chainValid, map[string]any{"api": "derive", "pattern_ref": p.ContentHash, "parent_ref": parent.ContentHash, "record_hash": p.RecordHash})
}

func uncertaintyAPI(store *Store, req Request, started time.Time, chainValid bool) Response {
	if req.PatternRef == "" {
		return failure(started, "fail", errors.New("pattern_ref required"))
	}
	if len(req.ReceiverPrior) == 0 {
		return failure(started, "fail", errors.New("receiver_prior required"))
	}
	p, ok, err := store.find(req.PatternRef)
	if err != nil {
		return failure(started, "fail", err)
	}
	if !ok {
		return failure(started, "n/a", errors.New("pattern_ref not found"))
	}
	posterior := tokenDistribution(p.What)
	kl := klDiv(req.ReceiverPrior, posterior)
	priorH := shannonDist(req.ReceiverPrior)
	postH := shannonDist(posterior)
	resp := success(started, *p, chainValid, map[string]any{"api": "measure_uncertainty", "pattern_ref": p.ContentHash, "kl": kl, "posterior": posterior})
	resp.KLDiv = kl
	resp.PriorEntropy = priorH
	resp.PosteriorEntropy = postH
	resp.UncertaintyReduction = round(priorH-postH, 6)
	resp.Passed = kl > 0
	return resp
}

func densityAPI(store *Store, req Request, started time.Time, chainValid bool) Response {
	var what any
	ref := req.PatternRef
	if ref != "" {
		p, ok, err := store.find(ref)
		if err != nil {
			return failure(started, "fail", err)
		}
		if !ok {
			return failure(started, "n/a", errors.New("pattern_ref not found"))
		}
		what = p.What
	} else {
		what = req.Input
	}
	if what == nil {
		return failure(started, "fail", errors.New("pattern_ref or input required"))
	}
	threshold := req.EntropyThreshold
	if threshold == 0 {
		threshold = 2.0
	}
	entropy, density := infoDensity(what)
	resp := Response{Verdict: "pass", ChainValid: chainValid, Entropy: entropy, Density: density, Threshold: threshold, Passed: entropy >= threshold, LatencyMs: time.Since(started).Milliseconds()}
	if !resp.Passed {
		resp.Verdict = "fail"
	}
	resp.EvidenceHash = evidenceHash(map[string]any{"api": "measure_info_density", "pattern_ref": ref, "entropy": entropy, "density": density, "threshold": threshold, "passed": resp.Passed})
	return resp
}

func success(started time.Time, p StoredPattern, chainValid bool, ev map[string]any) Response {
	return Response{Verdict: "pass", PatternRef: p.ContentHash, RecordHash: p.RecordHash, PrevHash: p.PrevHash, ChainValid: chainValid, Metadata: p.Metadata, EvidenceHash: evidenceHash(ev), LatencyMs: time.Since(started).Milliseconds()}
}

func failure(started time.Time, verdict string, err error) Response {
	return Response{Verdict: verdict, Error: err.Error(), ChainValid: false, EvidenceHash: evidenceHash(map[string]any{"verdict": verdict, "error": err.Error()}), LatencyMs: time.Since(started).Milliseconds()}
}

func metricFailure(started time.Time, verdict string, err error, entropy, density, threshold float64) Response {
	resp := failure(started, verdict, err)
	resp.Entropy = entropy
	resp.Density = density
	resp.Threshold = threshold
	resp.Passed = false
	return resp
}
