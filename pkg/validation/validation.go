package validation

import (
	"crypto/sha256"
)

// CalculateHashSHA256 подсчет хеша SHA256.
func CalculateHashSHA256(content, key []byte) []byte {
	content = append(content, key...)
	hash := sha256.Sum256(content)
	return hash[:]
}
