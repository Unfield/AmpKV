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

type ApiKeyRecord struct {
	ID          string
	Key         string
	Name        string
	Permissions []Permission
	CreatedAt   time.Time
	ExpiresAt   *time.Time
	Disabled    bool
}

func (apr *ApiKeyRecord) HasPermission(p Permission) bool {
	if slices.Contains(apr.Permissions, p) {
		return true
	}

	if p != PermAdmin && apr.HasPermission(PermAdmin) {
		return true
	}

	return false
}

func (apr *ApiKeyRecord) IsValid() bool {
	if apr.Disabled {
		return false
	}
	if apr.ExpiresAt != nil && time.Now().After(*apr.ExpiresAt) {
		return false
	}
	return true
}

func (apr *ApiKeyRecord) ToByteSlice() ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	err := encoder.Encode(apr)
	if err != nil {
		return nil, fmt.Errorf("Failed to encode ApiKeyRecord to byte slice: %w", err)
	}

	return buffer.Bytes(), nil
}

func ApiKeyRecordFromBuffer(buffer []byte) (*ApiKeyRecord, error) {
	decoder := gob.NewDecoder(bytes.NewReader(buffer))
	var decodedApiKeyRecord ApiKeyRecord
	err := decoder.Decode(&decodedApiKeyRecord)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode ApiKeyRecord from buffer: %w", err)
	}

	return &decodedApiKeyRecord, nil
}
