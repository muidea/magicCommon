# magicCommon/foundation/signal 模块说明

## 概述

`foundation/signal` 提供一个轻量的进程内信号协调器，适合用于：

- 单次事件通知
- 请求与响应之间的简单等待配对
- 按 ID 协调异步结果

它不是操作系统信号封装，而是基于内存 map 和 channel 的内部同步工具。

## 核心接口

```go
type Gard struct {}

func (s *Gard) PutSignal(id int) error
func (s *Gard) WaitSignal(id, timeOut int) (interface{}, error)
func (s *Gard) TriggerSignal(id int, val interface{}) error
func (s *Gard) CleanSignal(id int)
func (s *Gard) Reset()
```

## 运行语义

### PutSignal

- 为指定 `id` 创建一个信号槽位。
- 重复创建同一个 `id` 会返回错误。

### WaitSignal

- 等待指定 `id` 的信号值。
- `timeOut < 0` 时，当前实现按 1 小时处理。
- 等待结束后，无论是收到值还是超时，都会移除并关闭对应信号槽位。

### TriggerSignal

- 向指定 `id` 发送信号值。
- 如果 `id` 不存在，或对应信号已经被清理/关闭，会返回错误。
- 当前实现不再依赖 `recover` 吞掉关闭竞态，而是显式返回失败结果。

### CleanSignal / Reset

- `CleanSignal(id)` 会清理并关闭单个信号槽位。
- `Reset()` 会清理并关闭全部信号槽位。
- 两者都会保证重复关闭是安全的。

## 当前限制

- 每个 `id` 当前只适合一发一收的简单场景。
- 没有广播、多等待者或取消语义。
- `timeOut` 参数单位是秒，不是 `time.Duration`。
