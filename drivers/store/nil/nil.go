package nil

import (
	"time"
)

type NilStore struct {
}

func NewNilStore() (*NilStore, error) {
	return &NilStore{}, nil
}

func (s *NilStore) Get(key string) ([]byte, bool) {
	return nil, false
}

func (s *NilStore) Set(key string, value []byte, cost int64) error {
	return nil
}

func (s *NilStore) SetWithTTL(key string, value []byte, cost int64, ttl time.Duration) error {
	return nil
}

func (s *NilStore) Delete(key string) {
}

func (s *NilStore) Close() error {
	return nil
}

func (r *NilStore) IsNil() bool {
	return true
}
