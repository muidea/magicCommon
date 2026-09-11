# magicCommon v1.5.16

本次发布收口 EventHub、后台任务和 framework 生命周期的可靠完成协议。

## 升级注意

本次沿用仓库现有 v1 发布序列和 module 路径，但**包含源码接口不兼容变更**，不是无须适配的补丁升级：

- `event.Hub` 和 `event.SimpleObserver` 的 `Subscribe` / `Unsubscribe` 返回 `*def.Error`；自定义实现、mock 和共享 Base Biz 必须同步更新，调用方不能忽略错误。
- `application.Application` 增加 `ShutdownChecked(context.Context) *def.Error`。
- `task.BackgroundRoutine` 增加 `AsyncTaskContext(context.Context, Task) error`。
- 自定义 Application、Hub、BackgroundRoutine 的实现和测试替身需重新编译验证。
- 插件在 Setup 失败后不再由局部 PluginMgr 提前释放。直接使用 Service/PluginMgr 的调用方必须执行统一检查式清理；组件需支持部分 Setup 后清理。
- 停机失败保持 `stopping` 和未释放依赖，不能立即重新启动；已成功停机的 Application 自有注入 runtime 在重启前需换成新实例。

## EventHub

- 控制队列拒绝和 Hub 关闭返回明确错误；已经入队的订阅操作等待真实完成，回执不再因超时被丢弃。
- SimpleObserver 的本地回调与 Hub 订阅状态串行更新，失败保留原状态，可重试；修复旧匹配结果覆盖新订阅缓存的竞态。
- Send 在执行前取消时不运行 handler；执行已开始时等待其真正结束。
- 同步调用链支持同 Hub 的活动祖先 lane 重入，独立调用链形成循环等待时拒绝；Post 始终排队。
- `event.DrainingHub` 提供 `Drain` / `TerminateChecked`。关闭超时保留资源供重试；取消订阅不等于在途通知排空。

## 后台任务

- SyncTask/SyncFunction 返回提交失败、任务 panic 和完成结果，不再吞掉错误或让 panic 后的等待永久阻塞。
- 超时返回 `def.Timeout`，但不取消已接受的任务；超时预算从提交成功后开始。
- 容量为 1 的队列正常调度；定时器及排队提交支持关闭/取消，Shutdown 等待真实退出和排空。

## Framework

- 新增 `application.Execute`，统一 Startup/Run 后的检查式清理。每次停机使用独立预算，未完成则重试；不会强杀未配合取消的回调。
- 支持 `BeginShutdown`、`Quiesce`、检查式 Teardown/Shutdown；先停止输入、排空在途任务和事件，再释放依赖。
- 按已进入的 Setup 阶段清理，失败遇错即停；重试跳过已完成阶段，传播清理错误和 panic。
- 修复停机后误建新 runtime、重启复用已关闭自有注入 runtime 的问题。

## 验证范围

- EventHub/Execute/Task/Application/Service/Plugin 受影响测试通过连续 3 轮 race 检查；订阅专项通过 20 轮重复验证。
- magicBase、magicCas、magicModulesRepo、magicRunner 的受影响调用方完成源码依赖编译/回归及当前 vendor 验证；这些仓库不随本库一并发布。
- 本地发布检查跳过 `TestDatabase`、`TestInsert`、`TestQuery` 三个会创建/删除数据库的集成测试，避免触及本机现有 PostgreSQL。标签 CI 的数据库测试使用流水线独立服务。
- MySQL 模式的完整本地检查因既有 `TestMySQLConnectionPoolInFetch` 连接 `localhost:3306` 超时而未通过；本次未改动 DAO。无数据库环境的 MySQL 验证排除 DAO 包，真实数据库覆盖以 CI 结果为准。

## 发布方式

发布 `master` 提交和注释标签 `v1.5.16`，供 Go module 按版本引用。当前标签流水线运行构建、格式/静态检查及 PostgreSQL/MySQL 测试，不自动创建 GitHub Release 条目或多平台二进制产物。
