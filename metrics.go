package main

import (
	"math"
	"sort"
	"strings"
	"unicode"
)

func entropyText(s string) float64 {
	if s == "" {
		return 0
	}
	counts := map[rune]float64{}
	total := 0.0
	for _, r := range s {
		counts[r]++
		total++
	}
	return entropyDist(counts, total)
}

func entropyDist(counts map[rune]float64, total float64) float64 {
	if total == 0 {
		return 0
	}
	h := 0.0
	for _, c := range counts {
		p := c / total
		if p > 0 {
			h -= p * math.Log2(p)
		}
	}
	return round(h, 6)
}

func tokenDistribution(v any) map[string]float64 {
	text := strings.ToLower(string(canonicalBytes(v)))
	counts := map[string]float64{}
	var b strings.Builder
	flush := func() {
		if b.Len() > 0 {
			counts[b.String()]++
			b.Reset()
		}
	}
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
			b.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	if len(counts) == 0 {
		counts["<empty>"] = 1
	}
	total := 0.0
	for _, c := range counts {
		total += c
	}
	for k, c := range counts {
		counts[k] = c / total
	}
	return counts
}

func shannonDist(d map[string]float64) float64 {
	h := 0.0
	for _, p := range normalizeDist(d) {
		if p > 0 {
			h -= p * math.Log2(p)
		}
	}
	return round(h, 6)
}

func klDiv(p, q map[string]float64) float64 {
	p = normalizeDist(p)
	q = normalizeDist(q)
	keys := map[string]bool{}
	for k := range p {
		keys[k] = true
	}
	for k := range q {
		keys[k] = true
	}
	eps := 1e-12
	s := 0.0
	for k := range keys {
		pv := p[k]
		if pv <= 0 {
			continue
		}
		qv := q[k]
		if qv <= 0 {
			qv = eps
		}
		s += pv * math.Log2(pv/qv)
	}
	return round(s, 6)
}

func normalizeDist(d map[string]float64) map[string]float64 {
	out := map[string]float64{}
	total := 0.0
	for k, v := range d {
		if v > 0 {
			out[k] = v
			total += v
		}
	}
	if total == 0 {
		out["<empty>"] = 1
		return out
	}
	for k, v := range out {
		out[k] = v / total
	}
	return out
}

func infoDensity(v any) (entropy, density float64) {
	b := canonicalBytes(v)
	entropy = entropyText(string(b))
	if len(b) == 0 {
		return 0, 0
	}
	return entropy, round(entropy/float64(len(b)), 6)
}

func patternMetadata(what any, from FromSource) map[string]any {
	m := map[string]any{"qualification": initialQualification(from), "origin": from.Type, "physical": physicalType(what)}
	if obj, ok := what.(map[string]any); ok {
		keys := make([]string, 0, len(obj))
		for k := range obj {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		m["emergent_keys"] = keys
		m["field_count"] = len(keys)
	}
	ent, density := infoDensity(what)
	m["entropy"] = ent
	m["density"] = density
	m["token_distribution"] = tokenDistribution(what)
	return m
}

func physicalType(v any) string {
	switch v.(type) {
	case map[string]any:
		return "json_object"
	case []any:
		return "json_array"
	case string:
		return "text"
	case float64, int, int64:
		return "number"
	case bool:
		return "bool"
	case nil:
		return "null"
	default:
		return "json_value"
	}
}

func round(v float64, places int) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return v
	}
	m := math.Pow10(places)
	return math.Round(v*m) / m
}
