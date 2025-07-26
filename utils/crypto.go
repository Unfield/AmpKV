package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	KEY_LENGTH  = 32
	ID_LENGTH   = 32
	ID_ALPHABET = "23456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
)

func GenerateKey() string {
	bytes := make([]byte, KEY_LENGTH)
	rand.Read(bytes)
	return base64.RawURLEncoding.EncodeToString(bytes)
}

func NewID() (string, error) {
	return GenerateID(ID_LENGTH)
}

func GenerateID(length int) (string, error) {
	if length <= 0 {
		length = ID_LENGTH
	}
	id, err := gonanoid.Generate(ID_ALPHABET, length)
	if err != nil {
		return "", fmt.Errorf("Failed to genearet id: %w", err)
	}
	return id, nil
}
