package auth

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"slices"
	"time"
)

type Permission string

const (
	PermRead   Permission = "read"
	PermWrite  Permission = "write"
	PermDelete Permission = "delete"

	PermAdmin Permission = "admin"
)

type ApiKey struct {
	ID          string
	Key         string
	Name        string
	Permissions []Permission
	CreatedAt   time.Time
	ExpiresAt   *time.Time
	Disabled    bool
}

func (apr *ApiKey) HasPermission(p Permission) bool {
	if slices.Contains(apr.Permissions, p) {
		return true
	}

	if p != PermAdmin && apr.HasPermission(PermAdmin) {
		return true
	}

	return false
}

func (apr *ApiKey) IsValid() bool {
	if apr.Disabled {
		return false
	}
	if apr.ExpiresAt != nil && time.Now().After(*apr.ExpiresAt) {
		return false
	}
	return true
}

func (apr *ApiKey) ToByteSlice() ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	err := encoder.Encode(apr)
	if err != nil {
		return nil, fmt.Errorf("Failed to encode ApiKey to byte slice: %w", err)
	}

	return buffer.Bytes(), nil
}

func ApiKeyFromBuffer(buffer []byte) (*ApiKey, error) {
	decoder := gob.NewDecoder(bytes.NewReader(buffer))
	var decodedApiKey ApiKey
	err := decoder.Decode(&decodedApiKey)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode ApiKey from buffer: %w", err)
	}

	return &decodedApiKey, nil
}
