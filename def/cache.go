package def

// QueryCacheResult 查询缓存结果
type QueryCacheResult struct {
	Result
	Cache interface{} `json:"cache"`
}

// CreateCacheParam 新建Cache参数
type CreateCacheParam struct {
	Value string `json:"value"`
	Age   int    `json:"age"`
}

// CreateCacheResult 新建Cache结果
type CreateCacheResult struct {
	Result
	Token string `json:"token"`
}

// DestroyCacheResult 删除Cache结果
type DestroyCacheResult Result
