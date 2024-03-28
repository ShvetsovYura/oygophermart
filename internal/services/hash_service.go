package services

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
)

type HashService struct {
	key []byte
}

func NewHashService() *HashService {
	return &HashService{key: []byte("mVrADym8hM")}
}

func (s *HashService) Hash(val string) string {
	hash := sha256.Sum256([]byte(val))
	return hex.EncodeToString(hash[:])
}

func (s *HashService) GenerateRnd(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (s *HashService) getSign(src []byte) ([]byte, error) {
	h := hmac.New(sha256.New, s.key)
	h.Write(src)
	sign := h.Sum(nil)
	return sign, nil
}

func (s *HashService) ExtractUserID(token string) (uint64, error) {
	data, err := hex.DecodeString(token)
	if err != nil {
		return 0, err
	}
	idPart := data[:8]
	id := binary.BigEndian.Uint64(idPart)
	return id, nil

}

func (s *HashService) GenerateToken(id uint64) (string, error) {
	var userID = make([]byte, 8)

	binary.BigEndian.PutUint64(userID, id)
	sign, err := s.getSign(userID)
	if err != nil {
		return "", err
	}
	token := append(userID, sign...)
	return hex.EncodeToString(token), nil
}

func (s *HashService) ValidateSign(token string) (bool, error) {

	data, _ := hex.DecodeString(token)
	idPart := data[:8]
	signPart := data[8:]

	sign, err := s.getSign(idPart)
	if err != nil {
		return false, err
	}

	return hmac.Equal(sign, signPart), nil
}
