package def

const (
	// InvalidPrivate 无效权限
	InvalidPrivate = iota
	// UnSetPrivate 未设权限
	UnSetPrivate
	// ReadPrivate 只读权限
	ReadPrivate
	// WritePrivate 可写权限
	WritePrivate
	// DeletePrivate 删除权限
	DeletePrivate
	// AllPrivate 全部权限
	AllPrivate
)

// PrivateInfo private info
type PrivateInfo struct {
	Value int    `json:"value"`
	Name  string `json:"name"`
}

var privateInfoList = []*PrivateInfo{
	{Value: InvalidPrivate, Name: "无效权限"},
	{Value: UnSetPrivate, Name: "未设权限"},
	{Value: ReadPrivate, Name: "只读权限"},
	{Value: WritePrivate, Name: "可写权限"},
	{Value: DeletePrivate, Name: "删除权限"},
	{Value: AllPrivate, Name: "全部权限"},
}

// GetPrivateInfoList get private info list
func GetPrivateInfoList() []*PrivateInfo {
	return privateInfoList
}

// GetPrivateInfo get private info
func GetPrivateInfo(value int) (ret *PrivateInfo) {
	switch value {
	case UnSetPrivate,
		ReadPrivate,
		WritePrivate,
		DeletePrivate,
		AllPrivate:
		ret = privateInfoList[value]
	default:
		ret = privateInfoList[0]
	}

	return
}

// PrivateItem 单条配置项
type PrivateItem struct {
	Path  string       `json:"path"`
	Value *PrivateInfo `json:"value"`
}

type Role struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Private     []*PrivateItem `json:"private"`
}
