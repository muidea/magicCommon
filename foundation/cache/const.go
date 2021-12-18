package cache

// MaxAgeValue 最大存放期限，无限期
const MaxAgeValue = -1

type commandAction int

const (
	putData      commandAction = iota // 存放数据
	fetchData                         // 获取数据
	remove                            // 删除指定数据
	getAll                            // 获取全部
	clearAll                          // 清除全部数据
	checkTimeOut                      // 检查超过生命周期的数据
	end                               // 停止Cache
)

type commandData struct {
	action commandAction
	value  interface{}
	result chan<- interface{} //单向Channel
}
