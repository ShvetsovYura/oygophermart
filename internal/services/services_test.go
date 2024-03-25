package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	input := "mysuperawesomepassword"
	exp := "0418d4ef0014dd74e4b3171bc97c58bcbfe8099f0f4459e8b0369152404e8431"
	s := NewHashService()
	res := s.Hash(input)
	assert.Equal(t, exp, res)
}
