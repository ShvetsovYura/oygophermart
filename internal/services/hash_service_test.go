package services

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	hs := NewHashService()
	r := rand.Int()
	token, err := hs.GenerateToken(uint64(r))
	assert.NoError(t, err)

	res, err := hs.ValidateSign(token)

	assert.True(t, res)
	assert.NoError(t, err)
}
