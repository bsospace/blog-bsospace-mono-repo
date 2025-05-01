package utils

import (
	"crypto/sha256"
	"encoding/binary"
)

// MockEmbed generates a fake embedding (vector) for a text by hashing
func MockEmbed(text string) []float64 {
	hash := sha256.Sum256([]byte(text))
	vec := make([]float64, 8)
	for i := 0; i < 8; i++ {
		vec[i] = float64(binary.BigEndian.Uint32(hash[i*4 : (i+1)*4]))
	}
	return vec
}
