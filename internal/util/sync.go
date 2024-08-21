package util

import (
	"sync"
	"time"
)

const (
	SEND_TIMEOUT time.Duration = 10 * time.Second
)

type SyncMapTypeSafe[K any, V any] struct {
	m sync.Map
}

func (m *SyncMapTypeSafe[K, V]) Delete(key K) { m.m.Delete(key) }
func (m *SyncMapTypeSafe[K, V]) Load(key K) (value V, ok bool) {
	v, ok := m.m.Load(key)
	if !ok {
		return value, ok
	}
	return v.(V), ok
}
func (m *SyncMapTypeSafe[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	v, loaded := m.m.LoadAndDelete(key)
	if !loaded {
		return value, loaded
	}
	return v.(V), loaded
}
func (m *SyncMapTypeSafe[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	a, loaded := m.m.LoadOrStore(key, value)
	return a.(V), loaded
}
func (m *SyncMapTypeSafe[K, V]) Range(f func(key K, value V) bool) {
	m.m.Range(func(key, value any) bool { return f(key.(K), value.(V)) })
}
func (m *SyncMapTypeSafe[K, V]) Store(key K, value V) { m.m.Store(key, value) }
