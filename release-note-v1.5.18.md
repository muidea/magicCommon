# magicCommon v1.5.18 发布说明

## 变更范围

本版本包含 v1.5.17 之后的以下变更：

- 修复 EventHub 订阅入队到期与队列可写同时就绪时的误拒绝。到期后仅进行一次即时入队探测；成功入队仍等待实际处理结果，拒绝的请求不会延迟执行。10ms 是入队等待窗口，不是整个订阅操作的严格耗时上限。
- 修正 execute 执行队列接近满载的阈值判断。
- 新增 `foundation/profiling` 的有界窗口耗时统计，并接入 HTTP 工具链路，用于按操作分析调用量、错误及耗时分布。
- 新增 `foundation/retention` 的版本化保留策略、文件读写及热加载控制器。业务清理查询和授权仍由调用项目负责。

## 升级与验证

调用方可使用 `go get github.com/muidea/magicCommon@v1.5.18` 更新模块，再通过项目自身的 vendor 同步脚本更新依赖，不直接修改 vendor 文件。

本次没有删除既有公开接口。订阅入队失败仍返回 `ResourceExhausted`，错误描述更准确；调用方应按错误码处理。跨越 v1.5.16 的升级仍须核对该版本的接口与生命周期要求。

发布前已通过 `go test ./... -count=1`、`go test -race ./event -count=1`、`go vet ./...`、`go build ./...` 和 `git diff --check`。PostgreSQL/MySQL 的发布环境验证由标签流水线执行，结果以 CI 为准。

## 发布方式

推送 master 提交和不可覆写的注释标签 `v1.5.18`，触发 `release-build.yml`。当前流程不自动创建 GitHub Release 条目或二进制发布包。
