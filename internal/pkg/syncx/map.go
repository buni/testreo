package syncx

import "sync"

// SyncMap is a type safe wrapper around sync.Map.
type SyncMap[K comparable, V any] struct {
	m sync.Map
}

// Store ...
func (sm *SyncMap[K, V]) Store(key K, value V) {
	sm.m.Store(key, value)
}

// Load ...
func (sm *SyncMap[K, V]) Load(key K) (value V, ok bool) { //nolint:ireturn
	v, ok := sm.m.Load(key)
	if !ok {
		return value, false //nolint:revive,forcetypeassert
	}
	return v.(V), true //nolint:revive,forcetypeassert
}

// LoadOrStore ...
func (sm *SyncMap[K, V]) LoadOrStore(key K, value V) (V, bool) { //nolint:ireturn
	v, ok := sm.m.LoadOrStore(key, value)
	if !ok {
		return value, false //nolint:revive,forcetypeassert
	}
	return v.(V), true //nolint:revive,forcetypeassert
}

// Delete ...
func (sm *SyncMap[K, V]) Delete(key K) {
	sm.m.Delete(key)
}

// Range ...
func (sm *SyncMap[K, V]) Range(f func(key K, value V) bool) {
	sm.m.Range(func(k, v any) bool {
		return f(k.(K), v.(V)) //nolint:forcetypeassert
	})
}

// Values ...
func (sm *SyncMap[K, V]) Values() []V {
	var values []V
	sm.Range(func(_ K, value V) bool {
		values = append(values, value)
		return true
	})
	return values
}
