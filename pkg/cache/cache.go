package cache

import (
	"fmt"
	"sync"
)

type entry[V any] struct {
	once  sync.Once
	value V
	err   error
}

type SyncCache[K comparable, V any] struct {
	m sync.Map
}

func NewSyncCache[K comparable, V any]() *SyncCache[K, V] {
	return new(SyncCache[K, V])
}

func (c *SyncCache[K, V]) Get(key K, fn func() (V, error)) (V, error) {
	actual, _ := c.m.LoadOrStore(key, new(entry[V]))

	e := actual.(*entry[V])

	e.once.Do(func() {
		defer func() {
			if r := recover(); r != nil {
				e.err = fmt.Errorf("cache compute panicked: %v", r)
			}
		}()

		e.value, e.err = fn()
	})

	return e.value, e.err
}
