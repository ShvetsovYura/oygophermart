package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

type HashService struct {
}

func NewHashService() *HashService {
	return &HashService{}
}

func (s *HashService) Hash(val string) string {
	hash := sha256.Sum256([]byte(val))
	return hex.EncodeToString(hash[:])
}

func (s *HashService) GetToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
