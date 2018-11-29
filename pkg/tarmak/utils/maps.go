// Copyright Jetstack Ltd. See LICENSE for details.
package utils

func MergeMaps(base map[string]interface{}, m ...map[string]interface{}) map[string]interface{} {
	for _, vars := range m {
		for key, value := range vars {
			base[key] = value
		}
	}
	return base
}

func MergeMapsBool(base map[string]bool, m ...map[string]bool) map[string]bool {
	for _, vars := range m {
		for key, value := range vars {
			base[key] = value
		}
	}
	return base
}

func DuplicateMapBool(base map[string]bool) map[string]bool {
	newMap := make(map[string]bool)
	for key, value := range base {
		newMap[key] = value
	}
	return newMap
}
