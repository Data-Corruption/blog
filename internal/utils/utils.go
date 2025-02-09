package utils

// CopyIfNotNil copies the value of src to dst if src is not nil.
func CopyIfNotNil[T any](dst, src *T) {
	if src != nil {
		*dst = *src
	}
}

// SetDefaultIfNil sets *dst to src if *dst is nil.
func SetDefaultIfNil[T any](dst **T, src *T) {
	if *dst == nil && src != nil {
		*dst = src
	}
}

// Ternary returns a if condition is true, otherwise b.
func Ternary[T any](condition bool, a, b T) T {
	if condition {
		return a
	}
	return b
}
