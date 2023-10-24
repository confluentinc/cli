package utils

import (
	"context"
	"sync"
	"time"
)

func NewDebouncer[T any](returnValOnCancel T, timeout time.Duration) Debouncer[T] {
	return Debouncer[T]{
		defaultVal: returnValOnCancel,
		timeout:    timeout,
	}
}

type Debouncer[T any] struct {
	timeout    time.Duration
	defaultVal T

	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
}

func (c *Debouncer[T]) Debounce(f func() T) T {
	c.mu.Lock()
	if c.cancel != nil {
		c.cancel()
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.mu.Unlock()

	select {
	case <-c.ctx.Done():
		return c.defaultVal
	case <-time.After(c.timeout):
		return f()
	}
}
