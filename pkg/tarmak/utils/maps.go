package utils

func MergeMaps(base map[string]interface{}, m ...map[string]interface{}) map[string]interface{} {
	for _, vars := range m {
		for key, value := range vars {
			base[key] = value
		}
	}
	return base
}
