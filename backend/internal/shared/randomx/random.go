// Package randomx generates cryptographically secure random strings.
package randomx

import (
	"crypto/rand"
	"math/big"
)

func String(length int, alphabet string) string {
	result := make([]byte, length)
	for i := range length {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			result[i] = 'a'
			continue
		}
		result[i] = alphabet[num.Int64()]
	}
	return string(result)
}
