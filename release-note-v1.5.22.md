# magicCommon v1.5.22 发布说明

## 变更与兼容性

新增 `foundation/net.HTTPSourceResolver`，按显式可信 CIDR 或受控 DNS 名称解析 HTTP 代理链，分别保留规范化来源 IP 与连接对端。解析器默认不信任代理，拒绝无限制 CIDR；代理 DNS 按请求解析，失败不扩大信任。来源元数据绑定请求上下文，转发时清理原来源头，仅传递服务端已确认的来源。

`GetHTTPRemoteAddress` 仍保留原签名，但返回纯 IP：有入口绑定时读取可信来源，没有绑定时只采用连接对端 IP，不再直接接受任意 `X-Forwarded-For`。依赖此前隐式信任行为的入口必须显式接入解析中间件。通用库不内置平台服务名、网段或业务策略。

## 验证与发布

来源解析和 Session 回归已通过 `go test -race ./foundation/net ./session -count=1`，`go vet ./...` 通过。覆盖伪造头、多级代理、IPv6、非法链、DNS 更换与失败、请求传播及后台请求无来源。

发布使用不可覆写的注释标签 `v1.5.22`；标签流水线另行执行构建与 PostgreSQL/MySQL 全量回归，结果以对应 Actions 为准。本版本没有 Schema 变更；平台真实来源 IP 是否正确仍须配套部署后专项验收，参见[跨项目来源合同](../docs/request-source-ip-contract.md)。
