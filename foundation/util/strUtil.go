package util

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

func MarshalString(val interface{}) string {
	switch val.(type) {
	case string:
		return val.(string)
	default:
	}

	byteVal, err := json.Marshal(val)
	if err != nil {
		return ""
	}

	return string(byteVal)
}

func UnmarshalString(val string) interface{} {
	//val = strings.Trim(val,"\"")
	if val == "" {
		return nil
	}

	var ret interface{}
	for {
		nLen := len(val)
		if val[0] == '{' {
			mVal := map[string]interface{}{}
			err := json.Unmarshal([]byte(val), &mVal)
			if err == nil {
				ret = mVal
				return ret
			}
		}
		if val[0] != '[' {
			if unicode.IsNumber(rune(val[0])) || val[0] == '-' {
				fVal := 0.00
				err := json.Unmarshal([]byte(val), &fVal)
				if err == nil {
					ret = fVal
					break
				}
			}

			// true or false
			if nLen == 4 || nLen == 5 {
				bVal := false
				err := json.Unmarshal([]byte(val), &bVal)
				if err == nil {
					ret = bVal
					break
				}
			}
		}

		if nLen < 2 {
			ret = val
			break
		}

		if unicode.IsNumber(rune(val[1])) || val[1] == '-' {
			fVal := []float64{}
			err := json.Unmarshal([]byte(val), &fVal)
			if err == nil {
				ret = fVal
				break
			}
		}

		bVal := []bool{}
		err := json.Unmarshal([]byte(val), &bVal)
		if err == nil {
			ret = bVal
			break
		}

		strVal := []string{}
		err = json.Unmarshal([]byte(val), &strVal)
		if err == nil {
			ret = strVal
			break
		}

		ret = val
		break
	}

	return ret
}

// ExtractSummary 抽取摘要
func ExtractSummary(content string) string {
	content = strings.TrimLeft(content, "\n")
	offset := strings.Index(content, "\n")
	if offset > 0 {
		return content[:offset]
	}

	return content
}

func cleanStr(str string) string {
	size := len(str)
	if size == 0 {
		return ""
	}

	val := str
	if str[0] == ',' {
		val = str[1:]
	}

	if str[size-1] == ',' {
		val = str[:size-1]
	}

	return strings.TrimSpace(val)
}

// Str2IntArray 字符串转换成数字数组
func Str2IntArray(str string) ([]int, bool) {
	ids := []int{}
	vals := strings.Split(cleanStr(str), ",")

	for _, val := range vals {
		if len(val) == 0 {
			continue
		}

		id, err := strconv.Atoi(val)
		if err != nil {
			return ids, false
		}
		ids = append(ids, id)
	}

	return ids, true
}

// IntArray2Str 数字数组转字符串
func IntArray2Str(ids []int) string {
	if len(ids) == 0 {
		return ""
	}

	val := ""
	for _, v := range ids {
		val = fmt.Sprintf("%s,%d", val, v)
	}

	return val[1:]
}
