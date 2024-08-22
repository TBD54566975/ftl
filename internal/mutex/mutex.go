package mutex

import "sync"

// Mutex is a simple mutex that can be used to protect a value.
//
// The zero value is safe to use if the zero value of T is safe to use.
//
// Example:
//
//	var m mutex.Mutex[*string]
//	s := m.Lock()
//	defer m.Unlock()
//	*s = "hello"
type Mutex[T any] struct {
	m sync.Mutex
	v T
}

func New[T any](v T) *Mutex[T] {
	return &Mutex[T]{v: v}
}

// Lock the Mutex and return its protected value.
func (l *Mutex[T]) Lock() T {
	l.m.Lock()
	return l.v
}

// Unlock the Mutex. The value returned by Lock is no longer valid.
func (l *Mutex[T]) Unlock() {
	l.m.Unlock()
}
