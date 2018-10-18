// Copyright Jetstack Ltd. See LICENSE for details.
package utils

import (
	"strings"
)

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

func SliceContains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}

	return false
}

func IndexOfString(slice []string, str string) int {
	for i, s := range slice {
		if s == str {
			return i
		}
	}

	return -1
}

func SliceContainsPrefix(slice []string, prefix string) bool {
	for _, s := range slice {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}

	return false
}
