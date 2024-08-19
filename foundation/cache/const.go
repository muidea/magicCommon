package cache

// ForeverAgeValue 最大存放期限，无限期
const ForeverAgeValue = -1
const OneMinuteAgeValue = 1
const TenMinutesAgeValue = 10
const HalfHourAgeValue = 30

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

type SearchOpr func(val interface{}) bool

type commandData struct {
	action commandAction
	value  interface{}
	result chan<- interface{} //单向Channel
}
