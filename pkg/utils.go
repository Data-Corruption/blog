package blog

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
)

// CopyNotNil copies the value of src to dst if src is not nil.
func CopyNotNil[T any](dst, src *T) {
	if src != nil {
		*dst = *src
	}
}

func Ternary[T any](condition bool, a, b T) T {
	if condition {
		return a
	}
	return b
}

func PadString(s string, length int) string {
	if len(s) < length {
		return s + strings.Repeat(" ", length-len(s))
	}
	return s
}

func GenRandomString(size int) (string, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
