package util

// ToPointer ...
func ToPointer[T any](s T) *T {
	return &s
}
