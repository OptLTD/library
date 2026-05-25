package schema

import (
	"context"
	"encoding/json"
	"search/consts"
	"search/support"
	"strings"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
)

// ICache 定义缓存接口
type ICache interface {
	Del(key string)
	Get(key string) *Source
	Set(key string, source *Source)
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Provider string // mem, redis
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

// redisCache Redis 缓存实现
type redisCache struct {
	client *redis.Client
	prefix string

	ctx context.Context
	ttl time.Duration
}

func (c *redisCache) getKey(key string) string {
	return c.prefix + key
}

func (c *redisCache) Get(key string) *Source {
	fullKey := c.getKey(key)
	data, err := c.client.Get(c.ctx, fullKey).Bytes()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return nil
	}

	var value *Source
	if err := json.Unmarshal(data, &value); err != nil {
		return nil
	}
	return value
}

func (c *redisCache) Set(key string, value *Source) {
	data, err := json.Marshal(value)
	if err != nil {
		return
	}

	fullKey := c.getKey(key)
	c.client.Set(c.ctx, fullKey, data, c.ttl)
}

func (c *redisCache) Del(key string) {
	fullKey := c.getKey(key)
	c.client.Del(c.ctx, fullKey)
}

var defaultCache ICache

// NewCache 创建缓存实例（类似 NewEngine 的工厂模式）
func NewCache(provider string, ttl time.Duration) ICache {
	provider = strings.ToLower(provider)
	switch provider {
	case "redis":
		// 否则从 registry 获取
		redisClient, ok := support.GetValue(consts.DATABASE_REDIS)
		if ok && redisClient != nil {
			ctx := context.Background()
			return &redisCache{
				client: redisClient.(*redis.Client),
				prefix: "cache:", ttl: ttl, ctx: ctx,
			}
		}
		// 如果都没有，返回默认的内存缓存
		return newMemCache(ttl, 8*time.Minute)
	default:
		// 默认使用内存缓存
		return newMemCache(ttl, 8*time.Minute)
	}
}

// GetCache 获取全局缓存实例
func GetCache() ICache {
	if defaultCache == nil {
		// 默认使用内存缓存
		defaultCache = newMemCache(5*time.Minute, 8*time.Minute)
	}
	return defaultCache
}

// SetCache 设置全局缓存实例（用于测试或自定义实现）
func SetCache(cache ICache) {
	defaultCache = cache
}
