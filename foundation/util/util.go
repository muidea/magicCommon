package util

// ExistIntArray 是否存在数组中
func ExistIntArray(val int, array []int) bool {
	found := false
	for _, v := range array {
		if val == v {
			found = true
			break
		}
	}

	return found
}
