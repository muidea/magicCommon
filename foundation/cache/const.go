package cache

// ForeverAge 最大存放期限，无限期
const ForeverAgeValue = -1
const OneMinuteAgeValue = 60
const TenMinutesAgeValue = 600
const HalfAnHourAgeValue = 1800

type commandAction int

const (
	putIn        commandAction = iota // 存放数据
	fetchOut                          // 获取数据
	search                            // 搜索数据
	remove                            // 删除指定数据
	getAll                            // 获取全部
	clearAll                          // 清除全部数据
	checkTimeOut                      // 检查超过生命周期的数据
	end                               // 停止Cache
)

// ConcurrentGoroutines 并发执行的协程数量
const ConcurrentGoroutines = 2

type SearchOpr func(val any) bool

type commandData struct {
	action commandAction
	value  any
	result chan any //单向Channel
}

type ExpiredCleanCallBackFunc func(string)
