package utils

// Coalesce returns the first non-zero value if it exists, otherwise returns the zero value.
func Coalesce[T comparable](values ...T) T {
	var zero T
	for _, value := range values {
		if value != zero {
			return value
		}
	}
	return zero
}

// IIF is an inline if statement. Returns v1 if the condition is true, otherwise v2.
func IIF[T any](condition bool, v1, v2 T) T {
	if condition {
		return v1
	}
	return v2
}
