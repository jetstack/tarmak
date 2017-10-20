package utils

func RemoveDuplicateStrings(slice []string) (result []string) {
	seen := make(map[string]bool)

	for _, str := range slice {
		if _, ok := seen[str]; !ok {
			result = append(result, str)
			seen[str] = true
		}
	}

	return result
}

func RemoveDuplicateInts(slice []int) (result []int) {
	seen := make(map[int]bool)

	for _, num := range slice {
		if _, ok := seen[num]; !ok {
			result = append(result, num)
			seen[num] = true
		}
	}

	return result
}
