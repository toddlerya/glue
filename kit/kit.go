package kit

// 使用泛型来进行slice的包含查找
func Contains[T comparable](elems []T, v T) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}

// 使用泛型进行slcice的过滤
func Filter[T any](slice []T, fn func(T) bool) []T {
	var result []T
	for _, element := range slice {
		if fn(element) {
			result = append(result, element)
		}
	}
	return result
}
