package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"
)

func sha256Ref(b []byte) string {
	h := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(h[:])
}

func evidenceHash(v any) string { return sha256Ref(canonicalBytes(v)) }

func canonicalBytes(v any) []byte {
	b, _ := json.Marshal(canonical(v))
	return b
}

func canonical(v any) any {
	if v == nil {
		return nil
	}
	rv := reflect.ValueOf(v)
	return canonicalValue(rv)
}

func canonicalValue(v reflect.Value) any {
	if !v.IsValid() {
		return nil
	}
	if v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		return canonicalValue(v.Elem())
	}
	switch v.Kind() {
	case reflect.Map:
		out := map[string]any{}
		keys := v.MapKeys()
		sort.Slice(keys, func(i, j int) bool { return fmt.Sprint(keys[i].Interface()) < fmt.Sprint(keys[j].Interface()) })
		for _, k := range keys {
			out[fmt.Sprint(k.Interface())] = canonicalValue(v.MapIndex(k))
		}
		return out
	case reflect.Slice, reflect.Array:
		out := make([]any, v.Len())
		for i := 0; i < v.Len(); i++ {
			out[i] = canonicalValue(v.Index(i))
		}
		return out
	case reflect.Struct:
		out := map[string]any{}
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			f := t.Field(i)
			if f.PkgPath != "" {
				continue
			}
			name := f.Name
			if tag := f.Tag.Get("json"); tag != "" {
				name = strings.Split(tag, ",")[0]
				if name == "-" {
					continue
				}
			}
			out[name] = canonicalValue(v.Field(i))
		}
		return out
	case reflect.Float32, reflect.Float64:
		f := v.Float()
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return nil
		}
		return f
	default:
		return v.Interface()
	}
}

func contentHash(from FromSource, where SpaceContext, what any, metadata map[string]any) string {
	stableMeta := map[string]any{}
	for _, k := range []string{"parent_ref", "modify_type"} {
		if metadata != nil && metadata[k] != nil {
			stableMeta[k] = metadata[k]
		}
	}
	return sha256Ref(canonicalBytes(map[string]any{"from": from, "where": where, "what": what, "derive": stableMeta}))
}

func recordHash(p StoredPattern) string {
	p.RecordHash = ""
	return sha256Ref(canonicalBytes(p))
}
