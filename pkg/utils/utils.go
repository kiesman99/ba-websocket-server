package utils

func Truncate(s string, maxLength int) string {
	if maxLength > len(s) {
		return s
	}

	return s[:maxLength]
}
