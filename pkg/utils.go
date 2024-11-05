package blog

import "strings"

func SetIfNil[T any](ptr **T, value T) {
	if *ptr == nil {
		*ptr = &value
	}
}

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
