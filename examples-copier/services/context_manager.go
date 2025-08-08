package services

import (
	"context"
	"sync"
)

type ContextManager struct {
	ctx context.Context
	mu  sync.RWMutex
}

var GlobalContext = &ContextManager{
	ctx: context.Background(), // Default context
}

func (c *ContextManager) GetContext() context.Context {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ctx
}

func (c *ContextManager) SetContext(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ctx = ctx
}
