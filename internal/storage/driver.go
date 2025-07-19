package storage

import "time"

type ICache interface {
	Get(key string) ([]byte, bool)
	Set(key string, value []byte, cost int64) error
	SetWithTTL(key string, value []byte, cost int64, ttl time.Duration) error
	Delete(key string)
	Close() error
}

type KVStore interface {
	Get(key string) ([]byte, bool)
	Set(key string, value []byte, cost int64) error
	SetWithTTL(key string, value []byte, cost int64, ttl time.Duration) error
	Delete(key string)
	Close() error
}
