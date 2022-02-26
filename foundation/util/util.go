package util

import "regexp"

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

// SameIntArray 是否存在数组中
func SameIntArray(val []int, array []int) bool {
	if len(val) != len(array) {
		return false
	}

	for _, v := range array {
		if !ExistIntArray(v, val) {
			return false
		}
	}

	return true
}

var telephonePattern = "1(3\\d|4[5-9]|5[0-35-9]|6[567]|7[0-8]|8\\d|9[0-35-9])\\d{8}"
var telephoneReg = regexp.MustCompile(telephonePattern)

func ExtractTelephone(val string) (ret string) {
	ret = telephoneReg.FindString(val)
	return
}
