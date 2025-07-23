package embedded

import (
	"fmt"
	"time"

	"github.com/Unfield/AmpKV/internal/storage"
	"github.com/Unfield/AmpKV/pkg/common"
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

func (ampkv *AmpKV) Get(key string) (*common.AmpKVValue, bool) {
	var rawVal []byte
	var found bool

	cacheVal, cacheHit := ampkv.cache.Get(key)
	if cacheHit {
		rawVal = cacheVal
		found = true
	} else {
		storeVal, storeFound := ampkv.store.Get(key)
		if storeFound {
			rawVal = storeVal
			found = true
			if ampkv.defaultTTL > 0 {
				ampkv.cache.SetWithTTL(key, rawVal, ampkv.defaultCost, ampkv.defaultTTL)
			} else {
				ampkv.cache.Set(key, rawVal, ampkv.defaultCost)
			}
		}
	}

	if found {
		ampKVValue, err := common.AmpKVValueFrom(rawVal)
		if err != nil {
			fmt.Printf("Error decoding AmpKVValue for key '%s': %v\n", key, err)
			return nil, false
		}
		return ampKVValue, true
	}

	return nil, false
}

func (ampkv *AmpKV) Set(key string, value any, cost int64) error {
	ampKVData, err := common.NewAmpKVValue(value)
	if err != nil {
		return err
	}
	ampKVDataByteSlice, err := ampKVData.ToByteSlice()
	if err != nil {
		return err
	}

	err = ampkv.cache.Set(key, ampKVDataByteSlice, cost)
	if err != nil {
		return fmt.Errorf("Failed to set value to Cache: %w", err)
	}
	err = ampkv.store.Set(key, ampKVDataByteSlice, cost)
	if err != nil {
		return fmt.Errorf("Failed to set value to Store: %w", err)
	}
	return nil
}

func (ampkv *AmpKV) SetWithTTL(key string, value any, cost int64, ttl time.Duration) error {
	ampKVData, err := common.NewAmpKVValue(value)
	if err != nil {
		return err
	}
	ampKVDataByteSlice, err := ampKVData.ToByteSlice()
	if err != nil {
		return err
	}

	err = ampkv.cache.SetWithTTL(key, ampKVDataByteSlice, cost, ttl)
	if err != nil {
		return fmt.Errorf("Failed to set value to Cache: %w", err)
	}
	err = ampkv.store.SetWithTTL(key, ampKVDataByteSlice, cost, ttl)
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
