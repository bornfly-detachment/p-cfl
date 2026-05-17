package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"time"
)

func main() {
	started := time.Now().UTC()
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		write(Response{Verdict: "fail", Error: err.Error(), LatencyMs: time.Since(started).Milliseconds()})
	}
	var req Request
	if err := json.Unmarshal(b, &req); err != nil {
		write(Response{Verdict: "fail", Error: "decode input: " + err.Error(), LatencyMs: time.Since(started).Milliseconds()})
	}
	resp := handle(req, started)
	write(resp)
}

func write(resp Response) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(resp); err != nil {
		slog.Error("encode response", "err", err)
		os.Exit(1)
	}
	os.Exit(exitCode(resp.Verdict))
}
