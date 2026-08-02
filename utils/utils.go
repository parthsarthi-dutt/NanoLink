package utils

import (
	"math/rand"
	"strings"
	"time"
)

const base62Chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// GenerateShortCode generates a random Base62 string of a given length
func GenerateShortCode(length int) string {
	var sb strings.Builder
	for i := 0; i < length; i++ {
		sb.WriteByte(base62Chars[rng.Intn(len(base62Chars))])
	}
	return sb.String()
}
