package badger

import (
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4"
)

type BadgerStore struct {
	badger *badger.DB
}

func NewBadgerStore(path string) (*BadgerStore, error) {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, fmt.Errorf("Failed to open Badger: %w", err)
	}
	return &BadgerStore{
		badger: db,
	}, nil
}

func (s *BadgerStore) Close() error {
	err := s.badger.Close()
	if err != nil {
		return fmt.Errorf("Failed to close Badger: %w", err)
	}
	return nil
}

func (s *BadgerStore) Get(key string) ([]byte, error) {
	var value []byte
	return value, s.badger.View(
		func(tx *badger.Txn) error {
			item, err := tx.Get([]byte(key))
			if err != nil {
				if err == badger.ErrKeyNotFound {
					return nil
				}
				return fmt.Errorf("Failed to get value from Badger: %w", err)
			}
			err = item.Value(func(val []byte) error {
				value = make([]byte, len(val))
				copy(value, val)
				return nil
			})
			if err != nil {
				return fmt.Errorf("Failed to copy value from Badger: %w", err)
			}
			return nil
		})
}

func (s *BadgerStore) Set(key string, value []byte, cost int64) error {
	err := s.badger.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), value)
	})
	if err != nil {
		return fmt.Errorf("Failed to set key/value pair to Badger: %w", err)
	}
	return nil
}

func (s *BadgerStore) SetWithTTL(key string, value []byte, cost int64, ttl time.Duration) error {
	e := badger.NewEntry([]byte(key), value).WithTTL(ttl)
	err := s.badger.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(e)
	})
	if err != nil {
		return fmt.Errorf("Failed to set key/value pair to Badger: %w", err)
	}
	return nil
}

func (s *BadgerStore) Delete(key string) {
	s.badger.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}
