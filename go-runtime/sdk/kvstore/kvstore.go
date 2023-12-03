// Package kvstore provides a generic key-value store.
package kvstore

import (
	"sync"
)

type Interface[V any] interface {
	Put(key string, value V) error
	Get(key string) (V, bool, error)
	Delete(key string) error
	Upsert(key string, upsert func(v V, created bool) (V, error)) (V, error)
}

type KV[V any] struct {
	lock  sync.Mutex
	store map[string]V
}

// Require ensures that a configued and provisioned KV store is available.
func Require[V any]() Interface[V] {
	return &KV[V]{store: map[string]V{}}
}

func (k *KV[V]) Put(key string, value V) error {
	k.lock.Lock()
	defer k.lock.Unlock()
	k.store[key] = value
	return nil
}

func (k *KV[V]) Get(key string) (V, bool, error) {
	k.lock.Lock()
	defer k.lock.Unlock()
	v, ok := k.store[key]
	return v, ok, nil
}

func (k *KV[V]) Delete(key string) error {
	k.lock.Lock()
	defer k.lock.Unlock()
	delete(k.store, key)
	return nil
}

// Upsert atomically inserts or updates an entry.
func (k *KV[V]) Upsert(key string, upsert func(v V, created bool) (V, error)) (V, error) {
	k.lock.Lock()
	defer k.lock.Unlock()
	v, ok := k.store[key]
	newValue, err := upsert(v, !ok)
	if err != nil {
		return v, err
	}
	k.store[key] = newValue
	return newValue, nil
}
