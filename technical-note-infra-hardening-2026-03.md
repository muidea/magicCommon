# magicCommon 基础设施加固说明（2026-03）

## 概览

这轮工作集中在 `magicCommon` 的基础设施模块，目标是修复确定性的并发、关闭和生命周期问题，并把当前稳定语义沉淀成可查阅文档。

涉及模块：

- `event`
- `execute`
- `task`
- `session`
- `framework/configuration`
- `framework/plugin/common`
- `framework/service`
- `foundation/cache`
- `foundation/pool`
- `foundation/path`
- `foundation/signal`
- `foundation/net`
- `monitoring`
- `foundation/dao`

## 关键修复

### event

- 修复 `Subscribe` / `Unsubscribe` / `Send` 在内部 channel 超时后仍等待 replay 的死锁路径
- 修复匹配缓存只按 `eventID` 复用导致不同 destination 漏收事件的问题
- 为匹配缓存增加独立锁，消除并发写风险
- 将 `Terminate()` 的终止状态改为原子控制，保证并发幂等
- 修复 `Terminate()` 在 hub action send 超时后仍死等 replay 的路径

### execute / task

- 为 `execute` 增加 `WaitTimeout(timeout)`，让调用方能显式判断是否真正等到空闲
- 为 `execute` 增加 `Idle()` 和 `WaitContext(ctx)`，让关闭流程和上层组件能按状态或上下文判断排空结果
- 保留 `Wait()` 历史兼容行为，但内部转为默认超时等待
- `event.Terminate()` 开始显式处理执行器等待超时
- 修复 `task.SyncTaskWithTimeOut()` 超时后任务完成再发送结果时的 panic 风险
- 为 `BackgroundRoutine` 增加 `TimerWithContext()` 和 `Shutdown(timeout)`，使定时任务和后台队列具备显式取消/排空能力
- `framework/application.Shutdown()` 现在会同步关闭默认 `BackgroundRoutine` 和 `EventHub`，并重建干净实例供后续复用

### monitoring

- `Manager.RegisterProvider` 改为使用实例自己的 registry，而不是误用全局 registry
- `UpdateConfig()` 在导出仍开启但配置变化时会正确重建 exporter

### session / configuration / service

- 修复 `session` 中多处读锁下写状态、observer 遍历与修改并发的问题
- 修复 `session.Registry.CountSession(nil)` 会意外终止内部 worker 的问题
- `session.Registry.Release()` 改为幂等
- 修复 `SimpleFileWatcher.Stop()` 重复调用导致的二次关闭 panic
- 修复 `configuration.Unwatch*()` 实际无法注销 handler 的问题
- 配置变更通知改为基于 watcher 快照异步分发，避免和注销/注册并发竞争
- 修复 `framework/service` 中无效的 `recover()` 写法，改为真正的 `defer recover`
- `HoldService` 增加信号订阅清理

### foundation/cache / pool / path / signal / net

- `cache`：
  - 修复释放后过期清理仍向命令通道发送消息的 panic 风险
  - 将过期清理改为 worker 内同步删除
  - 统一 `Fetch/Search/GetAll/checkTimeOut` 的锁语义
  - `Release()` 改为幂等
  - 增加带 `CacheOptions` 的构造器，可配置容量上限和清理周期
  - 增加 `Stats()` 统计快照，导出 entries / hits / misses / evictions / expirations
- `pool`：
  - `Close(nil)` 不再 panic
- `path`：
  - `Monitor.Start()` / `Stop()` 补充明确状态和 goroutine 回收
  - observer 通知改为快照，避免和增删观察者并发竞争
- `signal`：
  - 将 `Gard` 改为显式状态管理，不再依赖 `recover` 吞掉关闭竞态
  - `TriggerSignal()` 在关闭后返回明确错误
- `net`：
  - `NewDNSCacheHttpClient()` 改为克隆默认 transport，不再污染全局 `http.DefaultTransport`

### foundation/dao

- PostgreSQL / MySQL 集成测试在数据库不可用时改为干净跳过或显式失败，不再把环境问题放大成空对象错误或 panic

## 新增回归测试

本轮新增或补强的回归测试主要包括：

- `event/hub_regression_test.go`
- `execute/execute_test.go`
- `task/background_regression_test.go`
- `session/session_regression_test.go`
- `framework/configuration/file_watcher_test.go`
- `framework/plugin/common/util_regression_test.go`
- `foundation/signal/signal_test.go`
- `foundation/net/httpClient_test.go`
- `foundation/path/monitor_test.go` 中的生命周期补充用例
- `foundation/cache/*_test.go` 中的 release/timeout cleanup 边界用例
- `foundation/cache/*_test.go` 中的 capacity/stats 回归用例
- `monitoring/manager_test.go`

## 文档补充

本轮同步新增或刷新了这些说明文档：

- `execute/README.md`
- `task/README.md`
- `foundation/cache/README.md`
- `foundation/pool/README.md`
- `foundation/path/README.md`
- `foundation/signal/README.md`
- `foundation/net/README.md`
- `event/README.md`
- `README.md`

这些文档已经覆盖当前实现中的关键运行语义，尤其是：

- 超时等待
- 显式取消与排空
- 关闭与释放
- 并发访问
- 组件幂等性
- 全局状态污染边界

## 当前验证结果

已验证通过的关键命令包括：

```bash
env GOCACHE=/tmp/magiccommon-gocache GOFLAGS=-mod=vendor \
go test ./event ./execute ./session ./framework/configuration \
  ./foundation/cache ./foundation/pool ./foundation/path \
  ./foundation/system ./foundation/signal ./foundation/net \
  ./framework/plugin/common ./framework/application \
  ./framework/service ./task ./monitoring ./monitoring/core \
  ./foundation/dao -count 1
```

## 当前仍保留的限制

- `execute.Wait()` 仍保留历史兼容语义，不保证无限等待到真正空闲
- `Monitor` 停止后不能再次启动
- `foundation/net` 仍以 helper 函数为主，没有统一 client 生命周期抽象
- `foundation/cache` 仍是轻量内存缓存，当前只提供简单容量上限和“最旧条目淘汰”，没有更复杂的策略和命中率驱动调优

## 建议的下一步

如果继续演进，优先级建议如下：

1. 继续把 `execute / task / event` 的关闭、排空和取消语义收敛成更统一的上层约定
2. 对剩余未系统审过的基础模块做最后一轮全仓代码评审
3. 把这轮变更整理成更面向团队协作的 release note 或 changelog
