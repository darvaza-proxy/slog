package internal

import "sort"

// SortedKeys returns a sorted list of non-empty field keys
func SortedKeys(fields map[string]any) []string {
	keys := make([]string, 0, len(fields))
	for key := range fields {
		if key != "" {
			keys = append(keys, key)
		}
	}

	sort.Strings(keys)
	return keys
}
