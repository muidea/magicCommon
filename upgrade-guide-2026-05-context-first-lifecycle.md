# magicCommon 升级迁移指南（2026-05）

## 目标

本轮升级将 `magicCommon` 的应用生命周期统一收敛为 `context-first` 模式。

核心原则只有一条：

- 启动、运行、关闭、终止、teardown 都必须显式接收 `context.Context`

这是一轮破坏性升级，不保留旧签名兼容层。

## 为什么要改

之前的生命周期实现混用了三种模式：

1. 无参 `Run()` / `Shutdown()` / `Terminate()`
2. 内部自行 `context.Background()`
3. 调用方外层再套一个超时 goroutine 兜底

这会导致几个确定性问题：

- shutdown 无法被上层取消，只能无限等待
- goroutine、event hub、background task 的退出预算不一致
- HTTP server 看起来已经退出，但进程仍卡在 teardown
- 关闭链路的阻塞点无法归因，最终只能靠外层超时强杀

这次升级的目标就是把“关闭预算”收敛成一条链路上传递的同一个 `ctx`。

## 新的生命周期合约

### application

旧：

```go
application.Startup(service)
application.Run()
application.Shutdown()
```

新：

```go
application.Startup(ctx, service)
application.Run(ctx)
application.Shutdown(ctx)
```

要求：

- `Startup` / `Run` / `Shutdown` 都必须传入调用方 context
- 不允许在应用入口继续依赖无参 shutdown
- shutdown budget 必须由最外层调用方决定

### service

旧：

```go
type Service interface {
    Startup(serviceName string, eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) *cd.Error
    Run() *cd.Error
    Shutdown()
}
```

新：

```go
type Service interface {
    Startup(ctx context.Context, serviceName string, eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) *cd.Error
    Run(ctx context.Context) *cd.Error
    Shutdown(ctx context.Context)
}
```

要求：

- `Run(ctx)` 必须响应 `ctx.Done()`
- `Shutdown(ctx)` 只能使用上传入的 `ctx`
- 不允许在 service 内部重新创建 `context.Background()` 作为关闭预算

### plugin / initiator / module

旧：

```go
Setup(eventHub event.Hub, backgroundRoutine task.BackgroundRoutine)
Run()
Teardown()
```

新：

```go
Setup(ctx context.Context, eventHub event.Hub, backgroundRoutine task.BackgroundRoutine)
Run(ctx context.Context)
Teardown(ctx context.Context)
```

要求：

- `Setup` 只做接线和初始化，不做无法取消的长期阻塞
- `Run` 中启动的 listener、timer、loop 必须可被 `ctx` 或其派生 ctx 终止
- `Teardown(ctx)` 不允许无视传入 ctx 再走 `Background()`

### event.Hub

旧：

```go
hub.Terminate()
```

新：

```go
hub.Terminate(ctx)
```

要求：

- `Terminate(ctx)` 的等待预算完全由 `ctx` 决定
- 不再允许内部写死 5s、10ms、无限等待作为最终关闭语义
- terminate 需要做到幂等；重复调用直接返回

### task.BackgroundRoutine

旧：

```go
routine.Shutdown(timeout)
```

新：

```go
routine.Shutdown(ctx)
```

要求：

- 不再将 timeout 作为独立关闭协议
- 是否超时、是否取消，一律由 `ctx.Done()` 决定
- 调用方如需 30s 预算，应在外层构造：

```go
ctx, cancel := context.WithTimeout(parent, 30*time.Second)
defer cancel()
routine.Shutdown(ctx)
```

## 迁移规则

### 1. 最外层先建立 shutdown context

推荐模式：

```go
runCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
defer stop()

if err := application.Startup(runCtx, service.DefaultService()); err != nil {
    return err
}

if err := application.Run(runCtx); err != nil {
    return err
}

shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
application.Shutdown(shutdownCtx)
```

重点：

- `runCtx` 和 `shutdownCtx` 可以不同
- 关闭预算必须由调用方显式给出
- 不要把 `Run(ctx)` 使用的长期 context 直接当作 shutdown timeout

### 2. 删除所有外层 goroutine 超时兜底

反模式：

```go
done := make(chan struct{})
go func() {
    shutdown()
    close(done)
}()

select {
case <-done:
case <-time.After(30 * time.Second):
}
```

替换为：

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
shutdown(ctx)
```

说明：

- 超时控制必须进入被关闭对象内部
- 外层 watcher 只会掩盖真正卡住的位置

### 3. 删除 teardown 内部的 `context.Background()`

反模式：

```go
func (p *plugin) Teardown() {
    _ = p.module.Teardown(context.Background())
}
```

替换为：

```go
func (p *plugin) Teardown(ctx context.Context) {
    _ = p.module.Teardown(ctx)
}
```

### 4. 删除 framework 内部私造关闭预算

反模式：

- `EventHub.Terminate()` 内部自己写死等待 5s
- `BackgroundRoutine.Shutdown()` 内部自己决定 timeout
- `Service.Shutdown()` 内部自己 `context.WithTimeout(context.Background(), ...)`

替换原则：

- framework 只消费调用方传入的 `ctx`
- framework 可以记录 `ctx.Err()`，但不能擅自替换 budget

## 必须禁止的反模式

以下写法在升级后都应视为错误：

### 1. 关闭路径使用 `context.Background()`

```go
func (s *Service) Shutdown(ctx context.Context) {
    server.Shutdown(context.Background())
}
```

问题：

- 调用方预算被直接丢弃

### 2. teardown 中重新创建 timeout

```go
func (s *Service) Shutdown(ctx context.Context) {
    innerCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    _ = s.http.Shutdown(innerCtx)
}
```

问题：

- 关闭预算分叉
- 外层 30s，内层 5s，语义冲突

### 3. 继续保留旧接口的兼容包装

```go
func Shutdown() {
    Shutdown(context.Background())
}
```

问题：

- 调用方很容易继续误用无界关闭

## 推荐实现模式

### application shutdown

```go
func (s *appImpl) Shutdown(ctx context.Context) {
    if ctx == nil {
        ctx = context.Background()
    }

    s.service.Shutdown(ctx)
    s.backgroundRoutine.Shutdown(ctx)
    s.eventHub.Terminate(ctx)
}
```

### hold service

```go
func (s *holdService) Run(ctx context.Context) *cd.Error {
    if err := s.defaultService.Run(ctx); err != nil {
        return err
    }

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    defer signal.Stop(sigChan)

    select {
    case <-sigChan:
    case <-ctx.Done():
    }
    return nil
}
```

### background routine shutdown

```go
func (s *backgroundRoutine) Shutdown(ctx context.Context) bool {
    if ctx == nil {
        ctx = context.Background()
    }

    // close queue once
    // wait loop done
    // wait execute drain with WaitContext(ctx)
}
```

### event hub terminate

```go
func (s *hubImpl) Terminate(ctx context.Context) {
    if ctx == nil {
        ctx = context.Background()
    }

    // broadcast terminate
    // wait by ctx
    // drain execute by ctx
    // cleanup maps/channels
}
```

## 升级检查清单

调用方迁移时，至少完成下面这些检查：

1. 所有 `application.Startup/Run/Shutdown` 都已传入 ctx
2. 所有 `service.Startup/Run/Shutdown` 实现都已改成 ctx 签名
3. 所有 plugin / initiator / module 的 `Setup/Run/Teardown` 都已改成 ctx 签名
4. 所有 `hub.Terminate()` 已改为 `hub.Terminate(ctx)`
5. 所有 `routine.Shutdown(timeout)` 已改为 `routine.Shutdown(ctx)`
6. 关闭链路里不再出现 `context.Background()` 取代上传入 ctx
7. 删除旧的 shutdown goroutine + timer 外层兜底逻辑
8. 测试覆盖：
   - shutdown 正常完成
   - shutdown 超时返回
   - repeated shutdown 幂等
   - signal cancel 能结束 `Run(ctx)`

## 建议验证

框架层：

```bash
cd /home/rangh/codespace/magicCommon
go test ./event ./task ./framework/application ./framework/service ./framework/plugin/... -count=1
```

接入方：

```bash
go test ./... -count=1
```

并额外验证：

- 正常启动 -> 正常关闭
- HTTP server 停止后进程能退出
- shutdown timeout 生效时能返回而不是卡死
- 无残留 goroutine / channel / queue drain 阻塞

## 本次升级结论

这次升级不是“给旧生命周期再加一个 timeout 兜底”，而是把生命周期协议本身改成：

- context 传入
- context 透传
- context 决定关闭预算

只有这样，应用、service、event hub、background routine、plugin teardown 才会共享同一条可取消、可观测、可验证的退出链路。
