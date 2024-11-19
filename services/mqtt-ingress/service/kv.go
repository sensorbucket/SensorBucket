package service

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type TTLValue[T any] struct {
	Expiry time.Time
	Value  T
}

type KV[T any] struct {
	ctx    context.Context
	lock   sync.RWMutex
	values map[string]TTLValue[T]
}

func NewKV[T any](ctx context.Context) *KV[T] {
	return &KV[T]{
		ctx:    ctx,
		lock:   sync.RWMutex{},
		values: map[string]TTLValue[T]{},
	}
}

func (kv *KV[T]) StartCleaner(interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-kv.ctx.Done():
			return
		case <-t.C:
			kv.Clean()
		}
	}
}

func (kv *KV[T]) Clean() {
	now := time.Now()
	count := uint(0)
	kv.lock.Lock()
	for key, value := range kv.values {
		if now.After(value.Expiry) {
			delete(kv.values, key)
			count++
		}
	}
	kv.lock.Unlock()
	tooketh := time.Now().Sub(now)
	fmt.Printf("KV Clean: took %d ms and %d items cleared\n", tooketh.Milliseconds(), count)
}

func (kv *KV[T]) Set(key string, value T, expiry time.Time) {
	kv.lock.Lock()
	kv.values[key] = TTLValue[T]{
		Expiry: expiry,
		Value:  value,
	}
	kv.lock.Unlock()
}

func (kv *KV[T]) Get(key string) (T, bool) {
	kv.lock.RLock()
	defer kv.lock.RUnlock()
	v, ok := kv.values[key]
	if !ok {
		return *new(T), false
	}
	return v.Value, true
}

func (kv *KV[T]) Delete(key string) {
	kv.lock.Lock()
	delete(kv.values, key)
	kv.lock.Unlock()
}
