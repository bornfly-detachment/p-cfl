package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Store struct {
	dir  string
	path string
	mu   sync.Mutex
}

func openStore(dir string) (*Store, error) {
	if dir == "" {
		dir = ".p-cfl-state"
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &Store{dir: dir, path: filepath.Join(dir, "patterns.jsonl")}, nil
}

func (s *Store) append(p StoredPattern) (StoredPattern, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	last, err := s.lastRecordHash()
	if err != nil {
		return p, err
	}
	p.PrevHash = last
	if p.When == 0 {
		p.When = time.Now().UTC().UnixNano()
	}
	p.RecordHash = recordHash(p)
	f, err := os.OpenFile(s.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return p, err
	}
	defer f.Close()
	b, _ := json.Marshal(canonical(p))
	if _, err := f.Write(append(b, '\n')); err != nil {
		return p, err
	}
	return p, nil
}

func (s *Store) all() ([]StoredPattern, error) {
	f, err := os.Open(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []StoredPattern
	scan := bufio.NewScanner(f)
	scan.Buffer(make([]byte, 64<<10), 8<<20)
	for scan.Scan() {
		var p StoredPattern
		if err := json.Unmarshal(scan.Bytes(), &p); err != nil {
			return nil, fmt.Errorf("decode store: %w", err)
		}
		out = append(out, p)
	}
	return out, scan.Err()
}

func (s *Store) find(ref string) (*StoredPattern, bool, error) {
	items, err := s.all()
	if err != nil {
		return nil, false, err
	}
	for i := len(items) - 1; i >= 0; i-- {
		if items[i].ContentHash == ref {
			return &items[i], true, nil
		}
	}
	return nil, false, nil
}

func (s *Store) lastRecordHash() (string, error) {
	items, err := s.all()
	if err != nil || len(items) == 0 {
		return "", err
	}
	return items[len(items)-1].RecordHash, nil
}

func (s *Store) auditChain() (bool, []string, error) {
	items, err := s.all()
	if err != nil {
		return false, nil, err
	}
	var errs []string
	prev := ""
	for i, p := range items {
		if p.PrevHash != prev {
			errs = append(errs, fmt.Sprintf("record_%d_prev_hash", i))
		}
		if recordHash(p) != p.RecordHash {
			errs = append(errs, fmt.Sprintf("record_%d_record_hash", i))
		}
		prev = p.RecordHash
	}
	return len(errs) == 0, errs, nil
}
