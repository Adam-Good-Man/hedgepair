package cache

import (
	"sync"
	"time"
)

var GlobalCache *Cache

// CacheItem 代表缓存中的单个项目
type CacheItem struct {
	Value      interface{}
	Expiration int64
}

func InitGlobalCache() {
	GlobalCache = New()
}

// Cache 代表一个简单的内存缓存系统
type Cache struct {
	items map[string]*CacheItem
	mu    sync.RWMutex
}

// New 创建并返回一个新的Cache实例
func New() *Cache {
	return &Cache{
		items: make(map[string]*CacheItem),
	}
}

// Set 向缓存中添加一个值，带有过期时间（单位：秒）
func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &CacheItem{
		Value:      value,
		Expiration: time.Now().Add(duration).Unix(),
	}
}

// Get 从缓存中检索一个值
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	if time.Now().Unix() > item.Expiration {
		delete(c.items, key)
		return nil, false
	}

	return item.Value, true
}

// Delete 从缓存中删除一个值
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}
