package cache

import (
	"context"
	"sync"
	"time"
)

type memEntry struct {
	val     []byte
	expires time.Time
}

type Memory struct {
	ttl   time.Duration
	mu    sync.RWMutex
	items map[string]memEntry
}

func NewMemory(ttl time.Duration) *Memory {
	m := &Memory{
		ttl:   ttl,
		items: make(map[string]memEntry),
	}
	go m.sweep()
	return m
}

func (m *Memory) Get(_ context.Context, key string, dst any) (bool, error) {
	m.mu.RLock()
	e, ok := m.items[key]
	m.mu.RUnlock()
	if !ok || time.Now().After(e.expires) {
		return false, nil
	}
	if err := unmarshal(e.val, dst); err != nil {
		return false, err
	}
	return true, nil
}

func (m *Memory) Set(_ context.Context, key string, val any) error {
	buf, err := marshal(val)
	if err != nil {
		return err
	}
	m.mu.Lock()
	m.items[key] = memEntry{val: buf, expires: time.Now().Add(m.ttl)}
	m.mu.Unlock()
	return nil
}

func (m *Memory) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	delete(m.items, key)
	m.mu.Unlock()
	return nil
}

func (m *Memory) sweep() {
	t := time.NewTicker(time.Minute)
	defer t.Stop()
	for range t.C {
		now := time.Now()
		m.mu.Lock()
		for k, e := range m.items {
			if now.After(e.expires) {
				delete(m.items, k)
			}
		}
		m.mu.Unlock()
	}
}
