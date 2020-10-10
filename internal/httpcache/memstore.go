package httpcache

import (
	"context"
	"sync"
)

// Memstore implements an in-memory store.
type Memstore struct {
	c sync.Map
}

// Store implementation.
func (m *Memstore) Store(ctx context.Context, key uint64, value []byte) error {
	m.c.Store(key, value)
	return nil
}

// Load implementation.
func (m *Memstore) Load(ctx context.Context, key uint64) ([]byte, error) {
	if v, ok := m.c.Load(key); ok {
		return v.([]byte), nil
	}
	return nil, nil
}
