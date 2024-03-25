package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLuhn(t *testing.T) {
	tests := []struct {
		input    []uint8
		expected bool
	}{
		{
			input:    []uint8{3, 3, 4, 3, 4, 4, 2, 5, 2, 2, 3, 4, 5, 2, 3, 5, 2, 3, 2, 8},
			expected: true,
		}, {
			input:    []uint8{3, 4, 5, 4, 4, 5, 5, 6, 6, 4, 6, 4, 6, 8},
			expected: true,
		}, {
			input:    []uint8{3, 4, 5, 4, 4, 5, 5, 6, 6, 4, 6, 4},
			expected: false,
		},
	}

	for _, test := range tests {
		isValid := CheckLuhn(test.input)
		assert.Equal(t, test.expected, isValid)
	}

}

func TestLuhnFromStr(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{
			input:    "33434425223452352328",
			expected: true,
		}, {
			input:    "34544556646468",
			expected: true,
		}, {
			input:    "345445566464",
			expected: false,
		},
	}

	for _, test := range tests {
		isValid, _ := CheckLuhnFromStr(test.input)
		assert.Equal(t, test.expected, isValid)
	}

}
