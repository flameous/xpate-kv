package kv

import (
	"sync"
	"time"
)

const defaultTTL = 1 * 1000 * 1000 * 1000 // 60 seconds

type value struct {
	val         string
	createdTime int64
	TTL         int64
}

type Cacher interface {
	Set(key, value string, ttl *int64)
	Read(key string) (string, bool)
	Delete(key string)
}

func NewCacher() Cacher {
	return &cache{
		container: make(map[string]value),
		mu:        &sync.RWMutex{},
	}
}

type cache struct {
	container map[string]value
	mu        *sync.RWMutex
}

func (c *cache) Set(k, v string, ttl *int64) {
	c.mu.Lock()

	innerValue := value{
		val:         v,
		createdTime: time.Now().UnixNano(),
		TTL:         defaultTTL,
	}
	if ttl != nil {
		innerValue.TTL = *ttl
	}

	c.container[k] = innerValue
	c.mu.Unlock()
}

func (c *cache) Read(key string) (string, bool) {
	c.mu.RLock()
	v, ok := c.container[key]
	c.mu.RUnlock()

	if !ok {
		return "", false
	}

	// data was expired
	if time.Now().UnixNano() > v.createdTime+v.TTL {
		c.Delete(key)
		return "", false
	}
	return v.val, true
}

func (c *cache) Delete(key string) {
	c.mu.Lock()
	delete(c.container, key)
	c.mu.Unlock()
}
