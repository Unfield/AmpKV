package auth

import (
	"fmt"
	"time"

	"github.com/Unfield/AmpKV/pkg/embedded"
	"github.com/Unfield/AmpKV/utils"
)

type ApiKeyManager struct {
	ampKV *embedded.AmpKV
}

func NewApiKeyManager(ampKVptr *embedded.AmpKV) (ApiKeyManager, error) {
	if ampKVptr == nil {
		return ApiKeyManager{}, fmt.Errorf("Failed to create new Api Key Manager: ampKVptr must not be empty")
	}
	return ApiKeyManager{
		ampKV: ampKVptr,
	}, nil
}

func (m ApiKeyManager) CreateApiKey(name string, perms []Permission, disabled bool, ttl *time.Duration) (*ApiKeyRecord, error) {
	if name == "" || len(name) < 5 {
		return nil, fmt.Errorf("Failed to create Api Key: name must be at least 5 characters long")
	}
	if len(perms) < 1 {
		return nil, fmt.Errorf("Failed to create Api Key: perms must contain at least 1 permission")
	}

	keyID, err := utils.NewID()
	if err != nil {
		return nil, fmt.Errorf("Failed to create Api Key: %w", err)
	}

	var expireationDate *time.Time
	if ttl != nil && *ttl > 0 {
		exp := time.Now().Add(*ttl)
		expireationDate = &exp
	}

	apiKeyRecord := ApiKeyRecord{
		ID:          keyID,
		Key:         utils.GenerateKey(),
		Name:        name,
		Permissions: perms,
		CreatedAt:   time.Now(),
		ExpiresAt:   expireationDate,
		Disabled:    disabled,
	}

	apiKeyRecordByteSlice, err := apiKeyRecord.ToByteSlice()
	if err != nil {
		return nil, fmt.Errorf("Failed to create Api Key: %w", err)
	}

	if ttl != nil && *ttl > 0 {
		err = m.ampKV.SetWithTTL(fmt.Sprintf("internal::api::key::%s", apiKeyRecord.Key), apiKeyRecordByteSlice, 1, *ttl)
		if err != nil {
			return nil, fmt.Errorf("Failed to create Api Key: %w", err)
		}
	} else {
		err = m.ampKV.Set(fmt.Sprintf("internal::api::key::%s", apiKeyRecord.Key), apiKeyRecordByteSlice, 1)
		if err != nil {
			return nil, fmt.Errorf("Failed to create Api Key: %w", err)
		}
	}

	return &apiKeyRecord, nil
}
