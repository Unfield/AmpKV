package embedded

import (
	"fmt"
	"time"

	"github.com/Unfield/AmpKV/internal/storage"
)

type AmpKV struct {
	cache       storage.ICache
	store       storage.KVStore
	defaultTTL  time.Duration
	defaultCost int64
}

type AmpKVOptions struct {
	DefaultTTL  time.Duration
	DefaultCost int64
}

func NewAmpKV(cache storage.ICache, store storage.KVStore, options AmpKVOptions) *AmpKV {
	if options.DefaultCost == 0 {
		options.DefaultCost = 1
	}

	if options.DefaultTTL < 0 {
		options.DefaultTTL = 0
	}

	return &AmpKV{
		cache:       cache,
		store:       store,
		defaultTTL:  options.DefaultTTL,
		defaultCost: options.DefaultCost,
	}
}

func (ampkv *AmpKV) Get(key string) ([]byte, bool) {
	val, cacheHit := ampkv.cache.Get(key)
	if cacheHit {
		return val, true
	}
	val, found := ampkv.store.Get(key)
	if found {
		if ampkv.defaultTTL > 0 {
			ampkv.cache.SetWithTTL(key, val, ampkv.defaultCost, ampkv.defaultTTL)
		} else {
			ampkv.cache.Set(key, val, ampkv.defaultCost)
		}
	}
	return val, found
}

func (ampkv *AmpKV) Set(key string, value []byte, cost int64) error {
	err := ampkv.cache.Set(key, value, cost)
	if err != nil {
		return fmt.Errorf("Failed to set value to Cache: %w", err)
	}
	err = ampkv.store.Set(key, value, cost)
	if err != nil {
		return fmt.Errorf("Failed to set value to Store: %w", err)
	}
	return nil
}

func (ampkv *AmpKV) SetWithTTL(key string, value []byte, cost int64, ttl time.Duration) error {
	err := ampkv.cache.SetWithTTL(key, value, cost, ttl)
	if err != nil {
		return fmt.Errorf("Failed to set value to Cache: %w", err)
	}
	err = ampkv.store.SetWithTTL(key, value, cost, ttl)
	if err != nil {
		return fmt.Errorf("Failed to set value to Store: %w", err)
	}
	return nil
}

func (ampkv *AmpKV) Delete(key string) {
	ampkv.cache.Delete(key)
	ampkv.store.Delete(key)
}

func (ampkv *AmpKV) Close() error {
	err := ampkv.cache.Close()
	if err != nil {
		return fmt.Errorf("Failed to close AmpKV Cache: %w", err)
	}
	err = ampkv.store.Close()
	if err != nil {
		return fmt.Errorf("Failed to close AmpKV Store: %w", err)
	}
	return nil
}
