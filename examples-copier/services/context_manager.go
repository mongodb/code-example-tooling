package services

import (
	"context"
	"sync"
)

// ContextManager manages a global context with thread-safe access.
// Allows setting and getting the context safely across multiple goroutines using a read-write mutex.
type ContextManager struct {
	ctx context.Context
	mu  sync.RWMutex
}

var GlobalContext = &ContextManager{
	ctx: context.Background(), // Default context
}

// GetContext returns the current context in a thread-safe manner
func (c *ContextManager) GetContext() context.Context {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ctx
}

// SetContext sets a new context in a thread-safe manner. If the provided context is nil, it defaults to context.Background()
func (c *ContextManager) SetContext(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ctx = ctx
}
