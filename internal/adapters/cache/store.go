package cache

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// Entry is the standard cache envelope: {found, data}. A stored entry with
// Found=false records a negative lookup.
type Entry[T any] struct {
	Found bool `json:"found"`
	Data  *T   `json:"data,omitempty"`
}

// Store is a generic Redis-backed cache. It knows nothing about databases or
// domain services — it only reads and writes cache entries.
type Store[T any] struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewStore[T any](rdb *redis.Client, ttl time.Duration) *Store[T] {
	return &Store[T]{rdb: rdb, ttl: ttl}
}

// Get reads a key. hit=false means cache miss. When hit=true, found signals
// whether the entry is positive (data present) or a cached negative lookup.
func (s *Store[T]) Get(ctx context.Context, key string) (data *T, found, hit bool) {
	raw, err := s.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false, false
	}
	var entry Entry[T]
	if err := json.Unmarshal(raw, &entry); err != nil {
		return nil, false, false
	}
	if !entry.Found {
		return nil, false, true
	}
	if entry.Data == nil {
		return nil, false, false
	}
	return entry.Data, true, true
}

func (s *Store[T]) Set(ctx context.Context, key string, data *T) {
	s.write(ctx, key, Entry[T]{Found: true, Data: data})
}

func (s *Store[T]) SetNotFound(ctx context.Context, key string) {
	s.write(ctx, key, Entry[T]{Found: false})
}

func (s *Store[T]) Delete(ctx context.Context, key string) {
	if err := s.rdb.Del(ctx, key).Err(); err != nil {
		log.Printf("cache: del %s: %v", key, err)
	}
}

func (s *Store[T]) write(ctx context.Context, key string, entry Entry[T]) {
	raw, err := json.Marshal(entry)
	if err != nil {
		log.Printf("cache: marshal %s: %v", key, err)
		return
	}
	if err := s.rdb.Set(ctx, key, raw, s.ttl).Err(); err != nil {
		log.Printf("cache: set %s: %v", key, err)
	}
}
