# magicCommon/foundation/pool 模块说明

## 概述

`foundation/pool` 提供一个泛型资源池，用于复用昂贵资源并限制资源总量。

典型场景包括：

- 数据库连接包装对象
- 外部客户端实例
- 带初始化成本的临时资源

## 核心接口

```go
pool, err := pool.New(factory,
    pool.WithInitialCapacity(5),
    pool.WithMaxSize(20),
)

resource, err := pool.Get()
err = pool.Put(resource)
pool.Close(releaseFunc)
```

## 核心语义

### New

- `factory` 不能为空。
- `initialCapacity` 不能小于 0。
- `maxSize` 必须大于 0。
- `initialCapacity` 不能超过 `maxSize`。

### Get

- 如果有空闲资源，优先直接返回空闲资源。
- 如果没有空闲资源且池未达到 `maxSize`，会同步创建新资源。
- 如果资源已满，则等待其他调用方 `Put()` 归还资源。
- 当空闲队列被取空且总量未到上限时，池会尝试异步预创建一个资源，减少下一次 `Get()` 的延迟。

### Put

- 归还资源到空闲队列。
- 如果池已经关闭，会返回错误。

### Close

- `Close()` 会将池标记为关闭，并释放当前空闲资源。
- `Close()` 是幂等的。
- `releaseFunc` 允许为 `nil`；此时会跳过资源回收回调。
- 对于已经借出但未归还的资源，当前实现只会记录告警，不会强制回收。

## 并发模型

- 资源池内部通过 `Mutex + Cond` 协调并发 `Get/Put`。
- 预创建资源使用单独 goroutine 异步完成，并在锁内维护 `preCreating` 状态，避免无限并发预创建。

## 当前限制

- 没有借出资源的超时回收能力。
- 没有资源健康检查能力。
- `Close()` 不会等待借出的资源归还，只处理当前空闲资源。

如果后续要扩展到更重的资源管理场景，建议补充：

1. 借出资源超时检测
2. 可选的健康检查
3. 显式的关闭等待策略
