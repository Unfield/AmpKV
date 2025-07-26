package auth

import (
	"fmt"
	"time"

	"github.com/Unfield/AmpKV/pkg/embedded"
	"github.com/Unfield/AmpKV/utils"
)

const (
	apiKeyKeyPrefix                 = "internal::api::key::"
	apiKeyMaxCreationAttempts uint8 = 10
	apiKeyCost                      = 1
)

type ApiKeyManager struct {
	ampKV *embedded.AmpKV
}

func NewApiKeyManager(ampKVptr *embedded.AmpKV) (*ApiKeyManager, error) {
	if ampKVptr == nil {
		return nil, fmt.Errorf("Failed to create new Api Key Manager: ampKVptr must not be empty")
	}
	return &ApiKeyManager{
		ampKV: ampKVptr,
	}, nil
}

func (m *ApiKeyManager) CreateAPIKey(name string, perms []Permission, disabled bool, ttl *time.Duration) (*ApiKey, error) {
	if len(name) < 5 {
		return nil, NewKeyError(KeyMalformed, "name must be at least 5 characters long")
	}
	if len(perms) < 1 {
		return nil, NewKeyError(KeyMalformed, "perms must contain at least 1 permission")
	}

	keyID, err := utils.NewID()
	if err != nil {
		return nil, NewKeyError(InternalError, "failed to create a unique id")
	}

	var expirationDate *time.Time
	if ttl != nil && *ttl > 0 {
		exp := time.Now().Add(*ttl)
		expirationDate = &exp
	}

	apiKey := ApiKey{
		ID:          keyID,
		Key:         "",
		Name:        name,
		Permissions: perms,
		CreatedAt:   time.Now(),
		ExpiresAt:   expirationDate,
		Disabled:    disabled,
	}

	for currentAttempt := range apiKeyMaxCreationAttempts {
		key := utils.GenerateKey()
		if _, found := m.ampKV.Get(apiKeyKeyPrefix + key); !found {
			apiKey.Key = key
			break
		}
		time.Sleep(time.Duration(1<<currentAttempt) * 5 * time.Millisecond)
	}

	if apiKey.Key == "" {
		return nil, &KeyError{Kind: InternalError, Message: "failed to create a unique key"}
	}

	apiKeyByteSlice, err := apiKey.ToByteSlice()
	if err != nil {
		return nil, NewKeyErrorWithCause(InternalError, "failed to create byte slice from ApiKey", err)
	}

	if ttl != nil {
		err = m.ampKV.SetWithTTL(apiKeyKeyPrefix+apiKey.Key, apiKeyByteSlice, apiKeyCost, *ttl)
		if err != nil {
			return nil, NewKeyErrorWithCause(InternalError, "failed to save key to storage", err)
		}
	} else {
		err = m.ampKV.Set(apiKeyKeyPrefix+apiKey.Key, apiKeyByteSlice, apiKeyCost)
		if err != nil {
			return nil, NewKeyErrorWithCause(InternalError, "failed to save key to storage", err)
		}
	}

	return &apiKey, nil
}

func (m *ApiKeyManager) GetApiKey(key string) (*ApiKey, error) {
	if key == "" {
		return nil, ErrKeyMalformed
	}

	apiKeyValue, found := m.ampKV.Get(apiKeyKeyPrefix + key)
	if !found {
		return nil, ErrKeyNotFound
	}

	apiKey, err := ApiKeyFromBuffer(apiKeyValue.Data)
	if err != nil {
		return nil, NewKeyErrorWithCause(InternalError, "failed to convert byte slice into ApiKey", err)
	}

	if !apiKey.IsValid() {
		return nil, ErrKeyExpired
	}

	return apiKey, nil
}

func (m *ApiKeyManager) DisabledApiKey(key string) error {
	apiKeyValue, found := m.ampKV.Get(apiKeyKeyPrefix + key)
	if !found {
		return ErrKeyNotFound
	}

	apiKey, err := ApiKeyFromBuffer(apiKeyValue.Data)
	if err != nil {
		return NewKeyErrorWithCause(InternalError, "failed to convert byte slice into ApiKey", err)
	}

	if !apiKey.IsValid() {
		return ErrKeyExpired
	}

	apiKey.Disabled = true

	apiKeyBytes, err := apiKey.ToByteSlice()
	if err != nil {
		return NewKeyErrorWithCause(InternalError, "failed to convert ApiKey to byte slice", err)
	}

	if apiKey.ExpiresAt != nil {
		err := m.ampKV.SetWithTTL(apiKeyKeyPrefix+key, apiKeyBytes, apiKeyCost, time.Until(*apiKey.ExpiresAt))
		if err != nil {
			return NewKeyErrorWithCause(InternalError, "failed to save key to store", err)
		}
	} else {
		err := m.ampKV.Set(apiKeyKeyPrefix+key, apiKeyBytes, apiKeyCost)
		if err != nil {
			return NewKeyErrorWithCause(InternalError, "failed to save key to store", err)
		}
	}
	return nil
}

func (m *ApiKeyManager) EnableApiKey(key string) error {
	apiKeyValue, found := m.ampKV.Get(apiKeyKeyPrefix + key)
	if !found {
		return ErrKeyNotFound
	}

	apiKey, err := ApiKeyFromBuffer(apiKeyValue.Data)
	if err != nil {
		return NewKeyErrorWithCause(InternalError, "failed to convert byte slice into ApiKey", err)
	}

	if !apiKey.IsValid() {
		return ErrKeyExpired
	}

	apiKey.Disabled = false

	apiKeyBytes, err := apiKey.ToByteSlice()
	if err != nil {
		return NewKeyErrorWithCause(InternalError, "failed to convert ApiKey to byte slice", err)
	}

	if apiKey.ExpiresAt != nil {
		err := m.ampKV.SetWithTTL(apiKeyKeyPrefix+key, apiKeyBytes, apiKeyCost, time.Until(*apiKey.ExpiresAt))
		if err != nil {
			return NewKeyErrorWithCause(InternalError, "failed to save key to store", err)
		}
	} else {
		err := m.ampKV.Set(apiKeyKeyPrefix+key, apiKeyBytes, apiKeyCost)
		if err != nil {
			return NewKeyErrorWithCause(InternalError, "failed to save key to store", err)
		}
	}
	return nil
}

func (m *ApiKeyManager) SetExpiration(key string, newExpiration time.Time) error {
	apiKeyValue, found := m.ampKV.Get(apiKeyKeyPrefix + key)
	if !found {
		return ErrKeyNotFound
	}

	if newExpiration.Before(time.Now()) {
		return NewKeyError(KeyMalformed, "newExpriration can not be in the past")
	}

	apiKey, err := ApiKeyFromBuffer(apiKeyValue.Data)
	if err != nil {
		return NewKeyErrorWithCause(InternalError, "failed to convert byte slice into ApiKey", err)
	}

	if !apiKey.IsValid() {
		return ErrKeyExpired
	}

	apiKey.ExpiresAt = &newExpiration

	apiKeyBytes, err := apiKey.ToByteSlice()
	if err != nil {
		return NewKeyErrorWithCause(InternalError, "failed to convert ApiKey to byte slice", err)
	}

	if apiKey.ExpiresAt != nil {
		err := m.ampKV.SetWithTTL(apiKeyKeyPrefix+key, apiKeyBytes, apiKeyCost, time.Until(*apiKey.ExpiresAt))
		if err != nil {
			return NewKeyErrorWithCause(InternalError, "failed to save key to store", err)
		}
	} else {
		err := m.ampKV.Set(apiKeyKeyPrefix+key, apiKeyBytes, apiKeyCost)
		if err != nil {
			return NewKeyErrorWithCause(InternalError, "failed to save key to store", err)
		}
	}
	return nil
}

func (m *ApiKeyManager) DeleteKey(key string) error {
	m.ampKV.Delete(apiKeyKeyPrefix + key)
	return nil
}
