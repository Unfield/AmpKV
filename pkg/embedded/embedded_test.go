package embedded_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Unfield/AmpKV/drivers/cache/ristretto"
	"github.com/Unfield/AmpKV/drivers/store/badger"
	"github.com/Unfield/AmpKV/pkg/embedded"
)

func setupTestAmpKV(t *testing.T) *embedded.AmpKV {
	tempDir, err := os.MkdirTemp("", "ampkv-test-db-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory for BadgerDB: %v", err)
	}

	t.Cleanup(func() {
		time.Sleep(200 * time.Millisecond)
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to clean up temp directory %s: %v", tempDir, err)
		}
	})

	cache, err := ristretto.NewRistrettoCache(1e7, 1<<30, 64)
	if err != nil {
		t.Fatalf("Failed to initialize Cache: %v", err)
	}

	store, err := badger.NewBadgerStore(filepath.Join(tempDir, "ampkv.db"))
	if err != nil {
		t.Fatalf("Failed to initialize Store: %v", err)
	}

	ampkv := embedded.NewAmpKV(cache, store, embedded.AmpKVOptions{
		DefaultTTL: 60 * time.Second,
	})

	t.Cleanup(func() {
		if err := ampkv.Close(); err != nil {
			t.Errorf("Failed to close AmpKV: %v", err)
		}
	})

	return ampkv
}

func TestAmpKVOperations(t *testing.T) {
	ampkv := setupTestAmpKV(t)

	t.Run("Get on non-existent key", func(t *testing.T) {
		_, foundVal := ampkv.Get("nonexistentkey")
		if foundVal {
			t.Errorf("Expected key 'nonexistentkey' not to be found, but it was.")
		}
	})

	t.Run("Set and Get", func(t *testing.T) {
		key := "testkey"
		value := []byte("testvalue")
		err := ampkv.Set(key, value, 1)
		if err != nil {
			t.Fatalf("Error setting key '%s'", key)
		}

		retrievedValue, foundVal := ampkv.Get(key)
		if !foundVal {
			t.Fatalf("Expected key '%s' to be found after Set, but it was not.", key)
		}
		if string(retrievedValue) != string(value) {
			t.Errorf("Retrieved value for '%s' was '%s', expected '%s'", key, retrievedValue, value)
		}
	})

	t.Run("SetWithTTL and Get before expiry", func(t *testing.T) {
		key := "testkeywithttl"
		value := []byte("testvaluewithttl")
		ampkv.SetWithTTL(key, value, 1, 2*time.Second)

		retrievedValue, foundVal := ampkv.Get(key)
		if !foundVal {
			t.Fatalf("Expected key '%s' to be found before TTL expiry, but it was not.", key)
		}
		if string(retrievedValue) != string(value) {
			t.Errorf("Retrieved value for '%s' was '%s', expected '%s'", key, retrievedValue, value)
		}
	})

	t.Run("SetWithTTL and Get after expiry", func(t *testing.T) {
		key := "testkeywithttl_expired"
		value := []byte("testvaluewithttl_expired")
		ampkv.SetWithTTL(key, value, 1, 2*time.Second)

		time.Sleep(1 * time.Second)

		_, foundVal := ampkv.Get(key)
		if foundVal {
			t.Errorf("Expected key '%s' to be gone after TTL expiry, but it was still found.", key)
		}
	})

	t.Run("Delete existing key", func(t *testing.T) {
		key := "keyToDelete"
		ampkv.Set(key, []byte("valueToDelete"), 1)

		_, foundVal := ampkv.Get(key)
		if !foundVal {
			t.Fatalf("Pre-condition failed: '%s' not found before deletion attempt.", key)
		}

		ampkv.Delete(key)

		_, foundVal = ampkv.Get(key)
		if foundVal {
			t.Errorf("Expected key '%s' to be gone after Delete, but it was still found.", key)
		}
	})

	t.Run("Delete non-existent key", func(t *testing.T) {
		key := "nonExistentKeyToDelete"
		ampkv.Delete(key)

		_, foundVal := ampkv.Get(key)
		if foundVal {
			t.Errorf("Expected non-existent key '%s' to remain non-existent after Delete, but it was found.", key)
		}
	})
}
