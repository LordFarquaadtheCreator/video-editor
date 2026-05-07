package frame

import "strings"

// Replacer performs frame transformations inside byte slices.
type Replacer interface {
	Replace(data []byte, target, replacement string) []byte
}

// StringReplacer operates by converting bytes to string.
type StringReplacer struct{}

func (StringReplacer) Replace(data []byte, target, replacement string) []byte {
	if target == "" {
		return data
	}

	replaced := strings.ReplaceAll(string(data), target, replacement)
	return []byte(replaced)
}

// ReplaceFrames swaps every occurrence of target with replacement.
func ReplaceFrames(input, target, replacement string) string {
	if target == "" {
		return input
	}
	return strings.ReplaceAll(input, target, replacement)
}
