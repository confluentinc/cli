package plural

import "strings"

// Plural returns the plural form of a word
func Plural(word string) string {
	if word == "" {
		return ""
	}

	// Words ending in `-ty` generally replace it with `-ties` in their plural forms
	if strings.HasSuffix(word, "ty") {
		return strings.TrimSuffix(word, "y") + "ies"
	}

	// Singular words ending w/ these suffixes generally add an extra -es syllable in their plural forms
	suffixes := map[string]bool{
		"s":  true,
		"x":  true,
		"z":  true,
		"ch": true,
		"sh": true,
	}

	for suffix := range suffixes {
		if strings.HasSuffix(word, suffix) {
			return word + "es"
		}
	}

	return word + "s"
}
