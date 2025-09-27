package util

import "regexp"

// ToPointer ...
func ToPointer[T any](s T) *T {
	return &s
}

// ReplaceFirstRegex ...
func ReplaceFirstRegex(original, search, replacement string) string {
	re := regexp.MustCompile(search)

	// Find the first match
	index := re.FindStringIndex(original)
	if index == nil {
		// Search string not found, return original string
		return original
	}

	// Replace only the first occurrence
	result := original[:index[0]] + replacement + original[index[1]:]

	return result
}
