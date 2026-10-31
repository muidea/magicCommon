# magicCommon 开发规则

本文件直接规定本仓的职责、边界和开发要求；单独检出本仓即可阅读执行。仓库内位置均相对于仓库根目录，协作项目通过项目名和 Git 地址识别，不依赖固定工作区或外部规范文件才能理解规则。

## 初期开发的内部协议与数据格式规则

- 当前阶段只维护一套内部协议和数据格式，版本标识固定为各合同的初始值，初始内容以最新代码实现为准，包含当前全部能力。没有用户明确要求，不递增协议／格式／Schema 版本，不引入版本协商、按版本分支、多版本读写、旧字段别名、兼容解码、自动迁移或回填。
- 合同变化直接更新语义 owner、SDK、消费者、夹具、清单及文档，使用配套组件和干净候选数据验证；不为历史开发数据保留过渡路径。已有历史兼容逻辑随改动清理，不以旧测试或旧设计作为继续维护它的理由。
- 本规则针对代码内部的协议识别和数据格式管理；不改变 Git、代码发布版本、构建身份和第三方依赖锁定。防止并发覆盖、撤权失效、重放及错误实例执行的条件修订、摘要、执行代际和持久回执按其实际安全职责保留，不能当作格式版本机械删除。

## 功能定位与边界

- 提供协议无关的基础设施：framework 应用／插件生命周期与配置、EventHub、任务、会话头工具、监控，以及 foundation 通用工具和基础 DAO。
- 提供通用启动、取消、排空、检查式停机和事件完成语义；调用方决定业务是否重试、如何确认远端提交及如何补偿。
- 不认识应用名、租户名、部署阶段、CAS 角色、订阅、Schema 修复或业务补偿策略。平台故障不能通过框架特例解决；通用变更须用与业务无关的用例证明。
- 会话／context 工具只传递事实，不认证用户、不创造权限。框架 handler 已开始执行后须跟踪真实完成，取消通知和 HTTP 结果都不是业务回执。
- 公开 API 与当前消费者配套修改，内部实现使用 internal；新增公共类型或抽象必须有真实消费者，不导出测试状态 setter。

## 协作项目

- magicEngine — https://github.com/muidea/magicEngine.git：消费网络运行所需的通用基础。
- magicOrm — https://github.com/muidea/magicOrm.git：消费通用模型基础工具，持久化语义仍由 ORM 提供。
- magicModulesRepo — https://github.com/muidea/magicModulesRepo.git：装配共享运行能力；业务策略留在业务 owner。
- magicBase — https://github.com/muidea/magicBase.git、magicCas — https://github.com/muidea/magicCas.git、magicFile — https://github.com/muidea/magicFile.git、magicRunner — https://github.com/muidea/magicRunner.git：消费框架，不能反向要求 Common 持有其流程状态。

## 结构、编码与验证

- `def` 为公共定义／错误，`foundation` 为基础工具，`framework` 为运行基础，`event`、`task`、`session`、`monitoring` 各维护自身合同。
- Go 工具链以 go.mod 为准。包名小写，导出项 PascalCase，局部变量 camelCase；沿用 `*cd.Error` 合同，立即处理错误；新日志使用 `log/slog`，保留取消／追踪 context，不输出 token 或秘密。
- 测试置于同目录 `*_test.go`，使用 TestXxx 和表驱动；生命周期修改覆盖启动失败、取消、排空、关闭及真实完成，通用 Provider／DAO 修改验证受影响数据库合同。
- 在仓库根运行 `go build ./...`、`go vet ./...`、`go test -count=1 ./...`；MySQL 专项使用 `-tags=mysql`。先选最小包；数据库测试使用显式一次性环境，不能把 CI 夹具密码当部署默认值，也不把单测当双数据库或 Live 通过。

## 依赖与交付

- 先修改语义 owner，再更新公开合同及直接消费者。工作区内使用 `upgrade_vendor.sh <目标项目...>` 同步依赖；重建前检查工作树和暂存区，保护 vendor 独有修改、原字节及补丁，保护失败即停止。禁止手改 vendor；同步后审查差异，并以默认 vendor 模式验证消费者。本地 replace 模式不能替代此验证；独立检出缺少同步环境时明确记录未验范围。
- 先运行最小受影响测试，再验证直接消费者。仅文档修改不要求格式化未改动的源码或执行部署；代码变化使用 gofmt／goimports、相关测试、go vet 和构建，按风险选择 race／真实数据库验证。源码、vendor、发行包、镜像和 Live 结果分别记录，夹具成功不能替代持久 owner 回读或现场验收。
- 合同或行为变化同时更新本仓使用说明、设计和对应平台材料；本文件的定位和边界直接同步更新，不把必要规则改成外部文档链接。提交按仓库分开使用 Conventional Commits，不提交凭据、密钥、运行数据或构建产物。
