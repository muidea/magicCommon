# magicCommon 发布说明（2026-03）

## 本轮重点

这轮主要完成了四件事：

1. 收敛 `execute / task / event` 的生命周期语义
2. 为 `foundation/cache` 增加容量限制和统计能力
3. 刷新 `monitoring` 的接入和测试说明
4. 补齐面向知识库和发布沟通的文档

## 核心变化

### 1. execute / task / event

- `execute.Execute` 新增：
  - `Idle()`
  - `WaitContext(ctx)`
- `task.BackgroundRoutine` 新增：
  - `TimerWithContext(ctx, ...)`
  - `Shutdown(timeout)`
- `framework/application.Shutdown()` 现在会：
  - 先执行 service shutdown
  - 再关闭默认 `BackgroundRoutine`
  - 再终止默认 `EventHub`
  - 最后重建新的默认组件实例，避免单例在测试或重复启动场景中复用已关闭对象

### 2. foundation/cache

- 保留原有构造器：
  - `NewCache`
  - `NewKVCache`
  - `NewGenericKVCache`
- 新增带 options 的构造器：
  - `NewCacheWithOptions`
  - `NewKVCacheWithOptions`
  - `NewGenericKVCacheWithOptions`
- 新增 `CacheOptions`：
  - `Capacity`
  - `CleanupInterval`
- 三类缓存实现都增加 `Stats()`，可查看：
  - `Entries`
  - `Capacity`
  - `Puts`
  - `Hits`
  - `Misses`
  - `Evictions`
  - `Expirations`

### 3. monitoring 文档与测试说明

- 推荐优先使用实例级 `Manager`
- 明确推荐初始化顺序：
  - `NewManager`
  - `Initialize`
  - `RegisterProvider`
  - `Start`
  - `Shutdown`
- 补充 exporter 依赖本地端口监听的环境说明
- 明确受限环境下应允许 exporter 集成测试跳过

### 4. 文档沉淀

本轮同步刷新了：

- `execute/README.md`
- `task/README.md`
- `event/README.md`
- `foundation/cache/README.md`
- `monitoring/README.md`
- `monitoring/QUICK_START.md`
- `technical-note-infra-hardening-2026-03.md`

## 验证

已验证：

```bash
cd /home/rangh/codespace/magicCommon
GOCACHE=/tmp/magiccommon-gocache GOFLAGS=-mod=vendor go test ./... -count 1
```

通过。
