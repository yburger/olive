package syncmap

import "sync"

// RWMap a thread safe map protected by RWLock.
type RWMap[K comparable, V any] struct {
	sync.RWMutex
	m map[K]V
}

func NewRWMap[K comparable, V any](n int) *RWMap[K, V] {
	return &RWMap[K, V]{
		m: make(map[K]V, n),
	}
}

func (m *RWMap[K, V]) Get(key K) (V, bool) {
	m.RLock()
	defer m.RUnlock()
	v, existed := m.m[key]
	return v, existed
}

func (m *RWMap[K, V]) Set(key K, value V) {
	m.Lock()
	defer m.Unlock()
	m.m[key] = value
}

func (m *RWMap[K, V]) Delete(key K) {
	m.Lock()
	defer m.Unlock()
	delete(m.m, key)
}

func (m *RWMap[K, V]) Len() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.m)
}

func (m *RWMap[K, V]) Each(f func(key K, value V) bool) {
	m.RLock()
	defer m.RUnlock()

	for k, v := range m.m {
		if !f(k, v) {
			return
		}
	}
}
