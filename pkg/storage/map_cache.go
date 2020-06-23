package storage

import (
	"sync"

	"github.com/yurykabanov/scraper/pkg/domain"
)

type MapCache struct {
	mu     *sync.RWMutex
	values map[string]*domain.Result
}

func NewMapCache() *MapCache {
	return &MapCache{
		mu:     new(sync.RWMutex),
		values: make(map[string]*domain.Result),
	}
}

func (c *MapCache) Get(url string) (*domain.Result, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.values[url], nil
}

func (c *MapCache) Put(url string, resp *domain.Result) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[url] = resp
	return nil
}
