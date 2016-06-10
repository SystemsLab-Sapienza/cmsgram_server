package auth

import (
	"crypto/rand"
	"math/big"
)

const (
	charset string = "0123456789abcdefghijklmnopqrstuvwxyz"
)

// Returns a newly-allocated random string of 'n' characters, made up from the given charset
func newRandomString(n int, charset string) string {
	newstring := make([]byte, n)
	l := big.NewInt(int64(len(charset)))

	for i := range newstring {
		n, _ := rand.Int(rand.Reader, l)
		index := n.Int64()
		newstring[i] = charset[index]
	}

	return string(newstring)
}

func NewBase36(n int) string {
	return newRandomString(n, charset)
}
