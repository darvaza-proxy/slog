package internal

// HasFields returns true if the map contains at least one non-empty key
func HasFields(fields map[string]any) bool {
	for k := range fields {
		if k != "" {
			return true
		}
	}
	return false
}
