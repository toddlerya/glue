package kit

import "encoding/json"

// 使用泛型来进行slice的包含查找
func Contains[T comparable](elems []T, v T) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}

// 使用泛型进行Map操作 ref: https://blog.xintech.co/golang-fan-xing-shi-jian-zhi-map-reduce-filter-han-shu/
func Map[T, M any](a []T, fn func(T) M) []M {
	result := make([]M, len(a))
	for i, e := range a {
		result[i] = fn(e)
	}
	return result
}

// 使用泛型进行Reduce操作 ref: https://blog.xintech.co/golang-fan-xing-shi-jian-zhi-map-reduce-filter-han-shu/
func Reduce[T, M any](s []T, fn func(M, T) M, initValue M) M {
	acc := initValue
	for _, v := range s {
		acc = fn(acc, v)
	}
	return acc
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

// Insert inserts the values v... into s at index i,
// returning the modified slice.
// In the returned slice r, r[i] == v[0].
// Insert panics if i is out of range.
// This function is O(len(s) + len(v)).
func Insert[S ~[]E, E any](s S, i int, v ...E) S {
	tot := len(s) + len(v)
	if tot <= cap(s) {
		s2 := s[:tot]
		copy(s2[i+len(v):], s[i:])
		copy(s2[i:], v)
		return s2
	}
	s2 := make(S, tot)
	copy(s2, s[:i])
	copy(s2[i:], v)
	copy(s2[i+len(v):], s[i:])
	return s2
}

func Struct2JSON(structData interface{}) (map[string]interface{}, error) {
	b, err := json.Marshal(&structData)
	if err != nil {
		return nil, err
	}
	var resultMap map[string]interface{}
	err = json.Unmarshal(b, &resultMap)
	if err != nil {
		return nil, err
	}
	return resultMap, err
}
