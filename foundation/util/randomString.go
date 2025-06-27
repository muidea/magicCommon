package util

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

var (
	defaultRand = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// RandomIdentifyCode 生成6位随机数验证码
func RandomIdentifyCode() string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	vcode := fmt.Sprintf("%06v", rnd.Int31n(1000000))

	// 避免首字母出现0的情况
	if vcode[0] == '0' {
		vcode = fmt.Sprintf("1%s", vcode[1:])
	}

	return vcode
}

// RandomSpec0 根据各种选项创建随机字符串，使用提供的随机源。
// 如果 start 和 end 都为0，则 start 和 end 被设置为 ' ' 和 'z'，即 ASCII 可打印字符。
// 如果 letters 和 numbers 都为 false，则 start 和 end 被设置为 0 和 math.MaxInt32。
// 如果 set 不为 nil，则从 start 和 end 之间的字符中选择。
// 该方法接受一个用户提供的 rand.Rand 实例作为随机源。
func RandomSpec0(count uint, start, end int, letters, numbers bool,
	chars []rune, rand *rand.Rand) string {
	if count == 0 {
		return ""
	}
	if start == 0 && end == 0 {
		end = 'z' + 1
		start = ' '
		if !letters && !numbers {
			start = 0
			end = math.MaxInt32
		}
	}
	buffer := make([]rune, count)
	gap := end - start
	for count != 0 {
		count--
		var ch rune
		if len(chars) == 0 {
			ch = rune(rand.Intn(gap) + start)
		} else {
			ch = chars[rand.Intn(gap)+start]
		}
		if letters && ((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')) ||
			numbers && (ch >= '0' && ch <= '9') ||
			(!letters && !numbers) {
			if ch >= rune(56320) && ch <= rune(57343) {
				if count == 0 {
					count++
				} else {
					buffer[count] = ch
					count--
					buffer[count] = rune(55296 + rand.Intn(128))
				}
			} else if ch >= rune(55296) && ch <= rune(56191) {
				if count == 0 {
					count++
				} else {
					// 高代理项，在插入之前插入低代理项
					buffer[count] = rune(56320 + rand.Intn(128))
					count--
					buffer[count] = ch
				}
			} else if ch >= rune(56192) && ch <= rune(56319) {
				// 私有高代理项，没有线索，所以跳过
				count++
			} else {
				buffer[count] = ch
			}
		} else {
			count++
		}
	}
	return string(buffer)
}

// RandomSpec1 创建一个指定长度的随机字符串。
// 字符将从参数指示的字母数字字符集中选择。
func RandomSpec1(count uint, start, end int, letters, numbers bool) string {
	return RandomSpec0(count, start, end, letters, numbers, nil, defaultRand)
}

// RandomAlphaOrNumeric 创建一个指定长度的随机字符串。
// 字符将从参数指示的字母数字字符集中选择。
func RandomAlphaOrNumeric(count uint, letters, numbers bool) string {
	return RandomSpec1(count, 0, 0, letters, numbers)
}

// RandomString 创建一个指定长度的随机字符串。
func RandomString(count uint) string {
	return RandomAlphaOrNumeric(count, false, false)
}

// RandomStringSpec0 创建一个指定长度的随机字符串。
func RandomStringSpec0(count uint, set []rune) string {
	return RandomSpec0(count, 0, len(set)-1, false, false, set, defaultRand)
}

// RandomStringSpec1 创建一个指定长度的随机字符串。
func RandomStringSpec1(count uint, set string) string {
	return RandomStringSpec0(count, []rune(set))
}

// RandomAscII 创建一个指定长度的随机字符串。
// 字符将从 ASCII 值在 32 到 126（包括）之间的字符集中选择。
func RandomAscII(count uint) string {
	return RandomSpec1(count, 32, 127, false, false)
}

// RandomAlphabetic 创建一个指定长度的随机字符串。
// 字符将从字母字符集中选择。
func RandomAlphabetic(count uint) string {
	return RandomAlphaOrNumeric(count, true, false)
}

// RandomAlphanumeric 创建一个指定长度的随机字符串。
// 字符将从字母数字字符集中选择。
func RandomAlphanumeric(count uint) string {
	return RandomAlphaOrNumeric(count, true, true)
}

// RandomNumeric 创建一个指定长度的随机字符串。
// 字符将从数字字符集中选择。
func RandomNumeric(count uint) string {
	return RandomAlphaOrNumeric(count, false, true)
}
