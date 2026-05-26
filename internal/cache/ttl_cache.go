package cache

import (
	"sync"
	"time"

	"github.com/wyzauto/translation-loader/internal/domain"
)

type item struct {
	value   []domain.Translation
	expires time.Time
}

type TTLCache struct {
	mu         sync.RWMutex
	ttl        time.Duration
	items      map[string]item
	entityKeys map[string]map[string]struct{}
}

func NewTTLCache(ttl time.Duration) *TTLCache {
	return &TTLCache{
		ttl:        ttl,
		items:      make(map[string]item),
		entityKeys: make(map[string]map[string]struct{}),
	}
}

func (c *TTLCache) Get(key string) ([]domain.Translation, bool) {
	c.mu.RLock()
	it, ok := c.items[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}

	if time.Now().After(it.expires) {
		c.mu.Lock()
		delete(c.items, key)
		c.removeKeyFromIndexLocked(key)
		c.mu.Unlock()
		return nil, false
	}

	out := make([]domain.Translation, len(it.value))
	copy(out, it.value)
	return out, true
}

func (c *TTLCache) Set(key, entityID string, value []domain.Translation) {
	c.mu.Lock()
	defer c.mu.Unlock()

	cp := make([]domain.Translation, len(value))
	copy(cp, value)

	c.items[key] = item{
		value:   cp,
		expires: time.Now().Add(c.ttl),
	}

	if _, ok := c.entityKeys[entityID]; !ok {
		c.entityKeys[entityID] = make(map[string]struct{})
	}
	c.entityKeys[entityID][key] = struct{}{}
}

func (c *TTLCache) Invalidate(entityID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	keys := c.entityKeys[entityID]
	for key := range keys {
		delete(c.items, key)
	}
	delete(c.entityKeys, entityID)
}

func (c *TTLCache) removeKeyFromIndexLocked(key string) {
	for entityID, keys := range c.entityKeys {
		if _, ok := keys[key]; ok {
			delete(keys, key)
			if len(keys) == 0 {
				delete(c.entityKeys, entityID)
			}
		}
	}
}
