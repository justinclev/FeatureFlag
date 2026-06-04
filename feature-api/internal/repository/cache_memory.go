package repository

import (
	"hash/fnv"
	"sync"
	"time"

	"github.com/featureflags/feature-api/internal/models"
)

const shardCount = 64

type cacheItem struct {
	flag      *models.Flag
	expiresAt time.Time
}

type shard struct {
	sync.RWMutex
	data map[string]cacheItem
}

// ShardedL1Cache is a high-concurrency in-memory cache that minimizes lock contention.
type ShardedL1Cache struct {
	shards [shardCount]*shard
}

func newShardedL1Cache() *ShardedL1Cache {
	c := &ShardedL1Cache{}
	for i := 0; i < shardCount; i++ {
		c.shards[i] = &shard{data: make(map[string]cacheItem)}
	}
	go c.janitor()
	return c
}

func (c *ShardedL1Cache) janitor() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		c.Cleanup()
	}
}

// Cleanup removes all expired items from the cache.
func (c *ShardedL1Cache) Cleanup() {
	for i := 0; i < shardCount; i++ {
		s := c.shards[i]

		var expired []string
		s.RLock()
		now := time.Now()
		for k, v := range s.data {
			if now.After(v.expiresAt) {
				expired = append(expired, k)
			}
		}
		s.RUnlock()

		if len(expired) == 0 {
			continue
		}

		s.Lock()
		for _, k := range expired {
			if v, ok := s.data[k]; ok && time.Now().After(v.expiresAt) {
				delete(s.data, k)
			}
		}
		s.Unlock()
	}
}

func (c *ShardedL1Cache) getShard(key string) *shard {
	h := fnv.New32a()
	h.Write([]byte(key))
	return c.shards[h.Sum32()%shardCount]
}

// Get retrieves a flag from the cache if it exists and has not expired.
func (c *ShardedL1Cache) Get(key string) (*models.Flag, bool) {
	s := c.getShard(key)
	s.RLock()
	item, ok := s.data[key]
	s.RUnlock()

	if !ok || time.Now().After(item.expiresAt) {
		return nil, false
	}
	return item.flag, true
}

// Set adds or updates a flag in the cache with the given TTL.
func (c *ShardedL1Cache) Set(key string, flag *models.Flag, ttl time.Duration) {
	s := c.getShard(key)
	s.Lock()
	s.data[key] = cacheItem{
		flag:      flag,
		expiresAt: time.Now().Add(ttl),
	}
	s.Unlock()
}

// Remove deletes a flag from the cache.
func (c *ShardedL1Cache) Remove(key string) {
	s := c.getShard(key)
	s.Lock()
	delete(s.data, key)
	s.Unlock()
}
