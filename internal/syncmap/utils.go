package syncmap

import "sync"

type SyncMap[U comparable, T interface{}] struct {
	sync.Map
}

func (m *SyncMap[K, V]) Load(key K) (value V, ok bool) {
	v, ok := m.Map.Load(key)
	if !ok {
		return value, ok
	}
	return v.(V), ok
}

func (m *SyncMap[U, T]) Store(key U, value T) {
	m.Map.Store(key, value)
}

func (m *SyncMap[U, T]) Delete(key U) {
	m.Map.Delete(key)
}

func (m *SyncMap[U, T]) Range(f func(key U, value T) bool) {
	m.Map.Range(func(key, value any) bool {
		return f(key.(U), value.(T))
	})
}
