# magicCommon/foundation/cache 模块说明

## 概述

`foundation/cache` 提供三类内存缓存实现：

- `MemoryCache`: 自动生成字符串 ID 的对象缓存
- `MemoryKVCache`: `string -> any` 的键值缓存
- `GenericKVCache[K, V]`: 泛型键值缓存

三者都采用内部命令通道 + worker 的实现方式，并带有周期性过期清理。

## 构造方式

兼容构造器仍然保留：

- `NewCache(cleanCallback)`
- `NewKVCache(cleanCallback)`
- `NewGenericKVCache(cleanCallback)`

如果需要配置容量和清理周期，可以使用：

- `NewCacheWithOptions(cleanCallback, options)`
- `NewKVCacheWithOptions(cleanCallback, options)`
- `NewGenericKVCacheWithOptions(cleanCallback, options)`

```go
options := &cache.CacheOptions{
    Capacity:        1024,
    CleanupInterval: time.Second,
}
```

## 核心语义

### Put / Fetch / Search

- `Put()` 会写入缓存并刷新最近访问时间。
- `Fetch()` 在命中时返回缓存值；如果条目已过期，会返回空值并删除该条目。
- `Search()` 会遍历缓存并返回首个命中的条目，同时刷新该条目的访问时间。

### 过期清理

- `maxAge` 单位为秒。
- `ForeverAgeValue` 表示永不过期。
- 后台清理协程按固定周期检查过期数据。
- 过期回调会在 worker 内同步执行，然后同步删除条目。

### 容量限制与淘汰

- `Capacity <= 0` 表示不限制条目数量。
- 达到容量上限后，新插入会淘汰当前缓存中“最旧访问时间”的条目。
- 这是轻量级的 oldest-entry eviction，不是完整的 LRU/LFU 策略。

### Stats

三个具体实现都提供 `Stats()`，返回当前快照：

- `Entries`
- `Capacity`
- `Puts`
- `Hits`
- `Misses`
- `Evictions`
- `Expirations`

### Release

- `Release()` 是幂等的，多次调用安全。
- 释放流程会先停止后台超时检查，再向 worker 发送结束命令，最后等待 worker 退出。
- 释放完成后不会再接受新的命令。
- 当前实现已经避免了“释放后仍有过期清理 goroutine 向命令通道发消息”的 panic 风险。

## 并发模型

- 命令型操作通过内部 `commandChannel` 串行进入 worker。
- 缓存内容本身通过 `sync.Map` 保存。
- 对共享缓存条目状态（例如 `cacheTime`）的修改由显式锁保护，避免并发刷新时间戳时发生数据竞争。

## 选择建议

- 新代码优先使用 `GenericKVCache[K, V]`
- 兼容旧接口时使用 `MemoryKVCache`
- 只需要自动分配 ID 的对象缓存时使用 `MemoryCache`

## 当前限制

- 过期清理是周期轮询，不是精确到期触发。
- `Search()` 只返回第一个命中项。
- 当前只提供简单容量上限与 oldest-entry 淘汰，适合作为轻量级进程内缓存，不是通用高阶缓存框架。
