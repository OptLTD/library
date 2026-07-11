package redis

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/OptLTD/library/search/schema"
	"github.com/redis/go-redis/v9"
)

type cache struct {
	client *redis.Client
	prefix string
	ctx    context.Context
	ttl    time.Duration
}

func NewCache(client *redis.Client, prefix string, ttl time.Duration) schema.ICache {
	if prefix == "" {
		prefix = "cache:"
	}
	return &cache{
		client: client,
		prefix: prefix,
		ctx:    context.Background(),
		ttl:    ttl,
	}
}

func (c *cache) getKey(key string) string {
	return c.prefix + key
}

func (c *cache) Get(key string) *schema.Source {
	fullKey := c.getKey(key)
	data, err := c.client.Get(c.ctx, fullKey).Bytes()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return nil
	}

	var value *schema.Source
	if err := json.Unmarshal(data, &value); err != nil {
		return nil
	}
	return value
}

func (c *cache) Set(key string, value *schema.Source) {
	data, err := json.Marshal(value)
	if err != nil {
		return
	}

	fullKey := c.getKey(key)
	c.client.Set(c.ctx, fullKey, data, c.ttl)
}

func (c *cache) Del(key string) {
	fullKey := c.getKey(key)
	c.client.Del(c.ctx, fullKey)
}

// SetGlobalCache 将 Redis 缓存设为 schema 全局缓存。
func SetGlobalCache(client *redis.Client, prefix string, ttl time.Duration) {
	schema.SetCache(NewCache(client, prefix, ttl))
}

// CacheKey 规范化缓存 key 前缀。
func CacheKey(prefix string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return "cache:"
	}
	if !strings.HasSuffix(prefix, ":") {
		prefix += ":"
	}
	return prefix
}
