package main

import "encoding/json"

type Request struct {
	API              string             `json:"api"`
	StoreDir         string             `json:"store_dir"`
	Input            any                `json:"input"`
	From             FromSource         `json:"from"`
	Where            SpaceContext       `json:"where"`
	Metadata         map[string]any     `json:"metadata"`
	ParentRef        string             `json:"parent_ref"`
	NewContent       any                `json:"new_content"`
	ModifyType       string             `json:"modify_type"`
	PatternRef       string             `json:"pattern_ref"`
	ReceiverPrior    map[string]float64 `json:"receiver_prior"`
	EntropyThreshold float64            `json:"entropy_threshold"`
}

type FromSource struct {
	Type       string `json:"type"`
	Signature  string `json:"signature,omitempty"`
	PubKey     string `json:"pub_key,omitempty"`
	CFLID      string `json:"cfl_id,omitempty"`
	Layer      string `json:"layer,omitempty"`
	Cert       string `json:"cert,omitempty"`
	ModelID    string `json:"model_id,omitempty"`
	PromptHash string `json:"prompt_hash,omitempty"`
	CallTS     int64  `json:"call_ts,omitempty"`
	VCFLID     string `json:"v_cfl_id,omitempty"`
	RecordHash string `json:"record_hash,omitempty"`
}

type SpaceContext struct {
	Project  string `json:"project"`
	Branch   string `json:"branch,omitempty"`
	Session  string `json:"session,omitempty"`
	FilePath string `json:"file_path,omitempty"`
}

type StoredPattern struct {
	From        FromSource     `json:"from"`
	When        int64          `json:"when"`
	Where       SpaceContext   `json:"where"`
	What        any            `json:"what"`
	ContentHash string         `json:"content_hash"`
	PrevHash    string         `json:"prev_hash,omitempty"`
	RecordHash  string         `json:"record_hash"`
	Metadata    map[string]any `json:"metadata"`
}

type Response struct {
	Verdict              string          `json:"verdict"`
	Error                string          `json:"error,omitempty"`
	PatternRef           string          `json:"pattern_ref,omitempty"`
	RecordHash           string          `json:"record_hash,omitempty"`
	PrevHash             string          `json:"prev_hash,omitempty"`
	Pattern              *StoredPattern  `json:"pattern,omitempty"`
	Output               any             `json:"output,omitempty"`
	ChainValid           bool            `json:"chain_valid"`
	KLDiv                float64         `json:"kl_divergence,omitempty"`
	PriorEntropy         float64         `json:"prior_entropy,omitempty"`
	PosteriorEntropy     float64         `json:"posterior_entropy,omitempty"`
	UncertaintyReduction float64         `json:"uncertainty_reduction,omitempty"`
	Entropy              float64         `json:"entropy,omitempty"`
	Density              float64         `json:"density,omitempty"`
	Threshold            float64         `json:"threshold,omitempty"`
	Passed               bool            `json:"passed,omitempty"`
	Metadata             map[string]any  `json:"metadata,omitempty"`
	EvidenceHash         string          `json:"evidence_hash"`
	LatencyMs            int64           `json:"latency_ms"`
	Raw                  json.RawMessage `json:"-"`
}

func exitCode(verdict string) int {
	switch verdict {
	case "pass":
		return 0
	case "n/a":
		return 2
	default:
		return 1
	}
}
