package strutil

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
)

// Pad returns a string of length `length` by padding `s` with spaces.
// If `s` is already longer than `length`, it is returned as-is.
func Pad(s string, length int) string {
	if len(s) < length {
		return s + strings.Repeat(" ", length-len(s))
	}
	return s
}

// Random returns a random url and filepath safe string of length `length`.
func Random(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
