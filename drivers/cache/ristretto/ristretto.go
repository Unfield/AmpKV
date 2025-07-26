package ristretto

import (
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto/v2"
)

type RistrettoCache struct {
	cache *ristretto.Cache[string, []byte]
}

func NewRistrettoCache(NumCounters, MaxCost, BufferItems int64) (*RistrettoCache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config[string, []byte]{
		NumCounters: NumCounters,
		MaxCost:     MaxCost,
		BufferItems: BufferItems,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize Ristretto: %w", err)
	}

	return &RistrettoCache{
		cache,
	}, nil
}

func (r *RistrettoCache) Get(key string) ([]byte, bool) {
	return r.cache.Get(key)
}

func (r *RistrettoCache) Set(key string, value []byte, cost int64) error {
	result := r.cache.Set(key, value, cost)
	r.cache.Wait()
	if !result {
		return fmt.Errorf("Failed to set key: %s", key)
	}
	return nil
}

func (r *RistrettoCache) SetWithTTL(key string, value []byte, cost int64, ttl time.Duration) error {
	result := r.cache.SetWithTTL(key, value, cost, ttl)
	r.cache.Wait()
	if !result {
		return fmt.Errorf("Failed to set key: %s", key)
	}
	return nil
}

func (r *RistrettoCache) Delete(key string) {
	r.cache.Del(key)
}

func (r *RistrettoCache) Close() error {
	r.cache.Close()
	return nil
}

func (r *RistrettoCache) IsNil() bool {
	return false
}
