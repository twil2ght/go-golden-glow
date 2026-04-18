package utils

func FilterWithFn[T any](slice []T, fn func(T) bool) []T {
	var result []T
	for _, item := range slice {
		if fn(item) {
			result = append(result, item)
		}
	}
	return result
}
func Filter[T comparable](slice []T, target T) []T {
	var result []T
	for _, item := range slice {
		if item != target {
			result = append(result, item)
		}
	}
	return result
}
func Contain[T comparable](slice []T, target T) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}
	return false
}
