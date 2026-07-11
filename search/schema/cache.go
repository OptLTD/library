package schema

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

// ICache 定义缓存接口
type ICache interface {
	Del(key string)
	Get(key string) *Source
	Set(key string, source *Source)
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Provider string // mem
}

// memCache 内存缓存实现
type memCache struct {
	cache *gocache.Cache
	ttl   time.Duration
}

func newMemCache(ttl, cleanup time.Duration) ICache {
	cache := gocache.New(ttl, cleanup)
	return &memCache{cache, ttl}
}

func (c *memCache) Get(key string) *Source {
	if value, ok := c.cache.Get(key); !ok {
		return nil
	} else if get, ok := value.(*Source); ok {
		return get
	}
	return nil
}

func (c *memCache) Set(key string, value *Source) {
	c.cache.Set(key, value, c.ttl)
}

func (c *memCache) Del(key string) {
	c.cache.Delete(key)
}

var defaultCache ICache

// NewCache 创建内存缓存实例。Redis 缓存请使用 search/driver/redis。
func NewCache(provider string, ttl time.Duration) ICache {
	return newMemCache(ttl, 8*time.Minute)
}

// GetCache 获取全局缓存实例
func GetCache() ICache {
	if defaultCache == nil {
		defaultCache = newMemCache(5*time.Minute, 8*time.Minute)
	}
	return defaultCache
}

// SetCache 设置全局缓存实例（用于测试或自定义实现）
func SetCache(cache ICache) {
	defaultCache = cache
}
