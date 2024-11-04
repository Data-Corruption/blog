package blog

import "strings"

func SetIfNil[T any](ptr **T, value T) {
	if *ptr == nil {
		*ptr = &value
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
