package main

import (
	"encoding/json"
	"os"
	"sync"
)

type offsetStore struct {
	mu    sync.Mutex
	path  string
	state map[string]int64
}

func loadOffsets(path string) *offsetStore {
	store := &offsetStore{path: path, state: map[string]int64{}}
	data, err := os.ReadFile(path)
	if err != nil {
		return store
	}
	_ = json.Unmarshal(data, &store.state)
	if store.state == nil {
		store.state = map[string]int64{}
	}
	return store
}

func (s *offsetStore) get(name string) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state[name]
}

func (s *offsetStore) set(name string, offset int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state[name] = offset
	data, err := json.Marshal(s.state)
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
