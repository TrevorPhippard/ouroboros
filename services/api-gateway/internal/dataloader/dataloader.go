// Package dataloader provides a small, dependency-free batching loader.
//
// A Loader collects every key requested within a short window (or until the
// batch is full), resolves them all with ONE call to the batch function, and
// fans the results back out to the callers. Loaders are intended to be created
// per GraphQL operation so that their cache can never leak data between users.
//
// Semantics worth knowing:
//   - The batch function returns a map. A key that is absent from the map is
//     "not found": Load returns the zero value of V and a nil error. With
//     pointer value types this means a nil result, which maps naturally onto
//     nullable GraphQL fields.
//   - If the batch function fails (or panics) every key in that batch receives
//     the error. Errors are never cached, so a later Load retries the key.
//   - Duplicate keys inside a batch are collapsed even when caching is off.
package dataloader

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"
)

// BatchFunc resolves many keys at once. It must only return entries for keys
// it was given. Missing keys are treated as "not found".
type BatchFunc[K comparable, V any] func(ctx context.Context, keys []K) (map[K]V, error)

// Config tunes a Loader. The zero value is valid.
type Config struct {
	// Wait is how long the loader waits for more keys before dispatching a
	// batch. Defaults to 2ms, which is enough to coalesce sibling resolvers.
	Wait time.Duration
	// MaxBatch dispatches immediately once this many unique keys are queued.
	// Defaults to 100. A value < 0 means unlimited.
	MaxBatch int
	// DisableCache turns off the per-loader result cache (batching and
	// in-batch de-duplication still apply). Use it for long-lived loaders such
	// as the one backing a subscription.
	DisableCache bool
}

const (
	defaultWait     = 2 * time.Millisecond
	defaultMaxBatch = 100
)

// Loader batches and caches lookups of V by K.
type Loader[K comparable, V any] struct {
	ctx   context.Context
	fetch BatchFunc[K, V]
	cfg   Config

	mu      sync.Mutex
	cache   map[K]*entry[V]
	pending *batch[K, V]
}

type entry[V any] struct {
	done  chan struct{}
	value V
	err   error
}

type batch[K comparable, V any] struct {
	keys    []K
	entries map[K]*entry[V]
	timer   *time.Timer
	closed  bool
}

// New creates a Loader. ctx is handed to fetch for every batch, so it should
// be the request-scoped context (carrying auth metadata, deadlines, ...), not
// the context of whichever caller happened to trigger the batch.
func New[K comparable, V any](ctx context.Context, cfg Config, fetch BatchFunc[K, V]) *Loader[K, V] {
	if cfg.Wait <= 0 {
		cfg.Wait = defaultWait
	}
	if cfg.MaxBatch == 0 {
		cfg.MaxBatch = defaultMaxBatch
	}
	l := &Loader[K, V]{ctx: ctx, fetch: fetch, cfg: cfg}
	if !cfg.DisableCache {
		l.cache = make(map[K]*entry[V])
	}
	return l
}

// Thunk waits for a previously queued key.
type Thunk[V any] func(ctx context.Context) (V, error)

// Load returns the value for key, batching it with concurrent requests.
func (l *Loader[K, V]) Load(ctx context.Context, key K) (V, error) {
	return l.LoadThunk(key)(ctx)
}

// LoadThunk queues key and returns immediately. Calling the thunk blocks until
// the batch completes or the supplied context is cancelled. Queue several keys
// before calling any thunk to batch them without spawning goroutines.
func (l *Loader[K, V]) LoadThunk(key K) Thunk[V] {
	e := l.enqueue(key)
	return func(ctx context.Context) (V, error) {
		select {
		case <-e.done:
			return e.value, e.err
		case <-ctx.Done():
			var zero V
			return zero, ctx.Err()
		}
	}
}

// LoadMany loads every key and returns the values keyed by input key. A key
// that was not found maps to the zero value of V (nil for pointer types), so
// callers can distinguish "asked for it, nothing there" from "never asked".
// The first error encountered is returned.
func (l *Loader[K, V]) LoadMany(ctx context.Context, keys []K) (map[K]V, error) {
	thunks := make(map[K]Thunk[V], len(keys))
	for _, key := range keys {
		if _, ok := thunks[key]; !ok {
			thunks[key] = l.LoadThunk(key)
		}
	}
	out := make(map[K]V, len(thunks))
	var firstErr error
	for key, thunk := range thunks {
		v, err := thunk(ctx)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		out[key] = v
	}
	return out, firstErr
}

// Prime stores value for key unless the key is already cached or in flight.
// It is a no-op when caching is disabled. Use it to seed the loader with data
// you already fetched (for example after a mutation returned a full object).
func (l *Loader[K, V]) Prime(key K, value V) {
	if l.cache == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.cache[key]; ok {
		return
	}
	e := &entry[V]{done: make(chan struct{}), value: value}
	close(e.done)
	l.cache[key] = e
}

// Clear evicts key from the cache so the next Load refetches it.
func (l *Loader[K, V]) Clear(key K) {
	if l.cache == nil {
		return
	}
	l.mu.Lock()
	delete(l.cache, key)
	l.mu.Unlock()
}

func (l *Loader[K, V]) enqueue(key K) *entry[V] {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.cache != nil {
		if e, ok := l.cache[key]; ok {
			return e
		}
	}

	b := l.pending
	if b == nil {
		b = &batch[K, V]{entries: make(map[K]*entry[V])}
		l.pending = b
		b.timer = time.AfterFunc(l.cfg.Wait, func() { l.dispatch(b) })
	}

	// De-duplicate inside the batch even when the cache is disabled.
	if e, ok := b.entries[key]; ok {
		return e
	}
	e := &entry[V]{done: make(chan struct{})}
	b.entries[key] = e
	b.keys = append(b.keys, key)
	if l.cache != nil {
		l.cache[key] = e
	}

	if l.cfg.MaxBatch > 0 && len(b.keys) >= l.cfg.MaxBatch {
		// Detach the full batch now; the next key starts a fresh one.
		b.timer.Stop()
		l.pending = nil
		b.closed = true
		go l.run(b)
	}
	return e
}

// dispatch is called by the batch timer.
func (l *Loader[K, V]) dispatch(b *batch[K, V]) {
	l.mu.Lock()
	if b.closed {
		l.mu.Unlock()
		return
	}
	b.closed = true
	if l.pending == b {
		l.pending = nil
	}
	l.mu.Unlock()
	l.run(b)
}

func (l *Loader[K, V]) run(b *batch[K, V]) {
	results, err := l.safeFetch(b.keys)

	// Settle every entry first, evicting failures from the cache, so waiters
	// never observe a half-filled batch.
	if err != nil && l.cache != nil {
		l.mu.Lock()
		for _, key := range b.keys {
			delete(l.cache, key)
		}
		l.mu.Unlock()
	}
	for _, key := range b.keys {
		e := b.entries[key]
		if err != nil {
			e.err = err
		} else {
			e.value = results[key] // zero value when not found
		}
		close(e.done)
	}
}

func (l *Loader[K, V]) safeFetch(keys []K) (res map[K]V, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("dataloader: batch function panicked: %v\n%s", r, debug.Stack())
			res = nil
		}
	}()
	return l.fetch(l.ctx, keys)
}
