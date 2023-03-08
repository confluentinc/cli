package utils

func Contains(haystack []string, needle string) bool {
	for _, x := range haystack {
		if x == needle {
			return true
		}
	}
	return false
}

func Remove(haystack []string, needle string) []string {
	for i, x := range haystack {
		if x == needle {
			return append(haystack[:i], haystack[i+1:]...)
		}
	}
	return haystack
}

func RemoveDuplicates(s []string) []string {
	check := make(map[string]bool)
	for _, v := range s {
		check[v] = true
	}
	return GetKeys(check)
}
