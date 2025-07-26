package embedded

import (
	"errors"
	"fmt"
	"time"

	nilCacheDriver "github.com/Unfield/AmpKV/drivers/cache/nil"
	nilStoreDriver "github.com/Unfield/AmpKV/drivers/store/nil"
	"github.com/Unfield/AmpKV/internal/storage"
	"github.com/Unfield/AmpKV/pkg/common"
)

type AmpKV struct {
	cache       storage.ICache
	store       storage.KVStore
	defaultTTL  time.Duration
	defaultCost int64
}

type AmpKVStorageMode uint8

const (
	AmpKVStorageModeDefault AmpKVStorageMode = iota
	AmpKVStorageModeCacheOnly
	AmpKVStorageModeStoreOnly
)

func (m AmpKVStorageMode) ToString() string {
	switch m {
	case AmpKVStorageModeDefault:
		return "Default"
	case AmpKVStorageModeCacheOnly:
		return "CacheOnly"
	case AmpKVStorageModeStoreOnly:
		return "StoreOnly"
	default:
		return fmt.Sprintf("Unknown AmpKVStorageMode: %d", m)
	}
}

type AmpKVOptions struct {
	DefaultTTL  time.Duration
	DefaultCost int64
	Mode        AmpKVStorageMode
}

func NewAmpKV(cacheDriver, storeDriver any, options AmpKVOptions) (*AmpKV, error) {
	var (
		finalCache storage.ICache
		finalStore storage.KVStore
		ok         bool
	)

	switch options.Mode {
	case AmpKVStorageModeDefault:
		if cacheDriver == nil {
			return nil, errors.New("cache driver cannot be nil for Default mode")
		}
		if storeDriver == nil {
			return nil, errors.New("store driver cannot be nil for Default mode")
		}

		finalCache, ok = cacheDriver.(storage.ICache)
		if !ok {
			return nil, fmt.Errorf("cache driver does not implement storage.ICache interface for Default mode. Type: %T", cacheDriver)
		}

		finalStore, ok = storeDriver.(storage.KVStore)
		if !ok {
			return nil, fmt.Errorf("store driver does not implement storage.KVStore interface for Default mode. Type: %T", storeDriver)
		}

		return createAmpKV(finalCache, finalStore, options)

	case AmpKVStorageModeStoreOnly:
		if storeDriver == nil {
			return nil, errors.New("store driver cannot be nil for StoreOnly mode")
		}

		finalStore, ok = storeDriver.(storage.KVStore)
		if !ok {
			return nil, fmt.Errorf("store driver does not implement storage.KVStore interface for StoreOnly mode. Type: %T", storeDriver)
		}

		return createAmpKV(&nilCacheDriver.NilCache{}, finalStore, options)

	case AmpKVStorageModeCacheOnly:
		if cacheDriver == nil {
			return nil, errors.New("cache driver cannot be nil for CacheOnly mode")
		}

		finalCache, ok = cacheDriver.(storage.ICache)
		if !ok {
			return nil, fmt.Errorf("cache driver does not implement storage.ICache interface for CacheOnly mode. Type: %T", cacheDriver)
		}

		return createAmpKV(finalCache, &nilStoreDriver.NilStore{}, options)

	default:
		return nil, fmt.Errorf("unknown storage mode: %s", options.Mode.ToString())
	}
}

func createAmpKV(cache storage.ICache, store storage.KVStore, options AmpKVOptions) (*AmpKV, error) {
	if options.DefaultCost == 0 {
		options.DefaultCost = 1
	}

	if options.DefaultTTL < 0 {
		options.DefaultTTL = 0
	}

	if cache.IsNil() && store.IsNil() {
		return nil, fmt.Errorf("cache and store can not be nil at the same time")
	}

	return &AmpKV{
		cache:       cache,
		store:       store,
		defaultTTL:  options.DefaultTTL,
		defaultCost: options.DefaultCost,
	}, nil
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

	if ttl > 0 {
		err = ampkv.cache.SetWithTTL(key, ampKVDataByteSlice, cost, ttl)
		if err != nil {
			return fmt.Errorf("Failed to set value to Cache: %w", err)
		}
		err = ampkv.store.SetWithTTL(key, ampKVDataByteSlice, cost, ttl)
		if err != nil {
			return fmt.Errorf("Failed to set value to Store: %w", err)
		}
		return nil
	} else {
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
