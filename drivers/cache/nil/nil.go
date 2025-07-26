package nil

import (
	"time"
)

type NilCache struct {
}

func NewNilCache() (*NilCache, error) {
	return &NilCache{}, nil
}

func (r *NilCache) Get(key string) ([]byte, bool) {
	return nil, false
}

func (r *NilCache) Set(key string, value []byte, cost int64) error {
	return nil
}

func (r *NilCache) SetWithTTL(key string, value []byte, cost int64, ttl time.Duration) error {
	return nil
}

func (r *NilCache) Delete(key string) {
}

func (r *NilCache) Close() error {
	return nil
}

func (r *NilCache) IsNil() bool {
	return true
}
