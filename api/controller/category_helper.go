package controller

import "strings"

// normalizeCategories trims entries, removes empties, and deduplicates case-insensitively.
func normalizeCategories(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, len(values))
	categories := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}

		key := strings.ToLower(trimmed)
		if _, exists := seen[key]; exists {
			continue
		}

		seen[key] = struct{}{}
		categories = append(categories, trimmed)
	}

	return categories
}
