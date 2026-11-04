# magicCommon EventHub — MQTT Topic 标准化调整设计

> - 文档状态：源码与消费者适配已实施；依赖终验见实施记录，制品及 Live 待验证
> - 版本：1.0 · 2026-11-04
> - 目标仓库：[muidea/magicCommon](https://github.com/muidea/magicCommon)（`master`，以本次源码核查为基线）
> - 变更范围：Common 的 Topic 校验、订阅与派发、独立 Destination 寻址，以及 Base、Runner、ModulesRepo、CAS、File 的消费合同和配套验证
> - **交付要求**：本文件定义 Common 自身改造及已知消费者的完整工作包，不依赖另一份 Rust 设计文档；Common 单仓通过不能代替消费者、制品和 Live 验证。

## 1. 背景和架构决策

### 1.1 目标

magicCommon 的 EventHub 已经采用 `/` 分层 Topic、`+` 和 `#` 通配，但其 `MatchValue` 允许中间 `#`、`:id` 参数通配，并拒绝部分 MQTT 合法的空层/零层匹配。目标是让 **EventHub 的 Topic 相关行为严格符合 MQTT 3.1.1 与 MQTT 5.0 共同的 Topic Name / Topic Filter 规则**。

### 1.2 强制决策 MC-TOPIC-ADR-001

1. **唯一正式规则**：Topic Name、Topic Filter 和匹配逻辑以 OASIS MQTT 为准；不增加 Legacy/Standard 开关、兼容解码或双规则运行。内部合同标识维持初始值，本文版本仍为 1.0。
2. **区分角色**：`Event.ID()` 是**具体 Topic Name**；`Hub.Subscribe(eventID, observer)` 的 `eventID` 是**Topic Filter**，不能把 Filter 当成发布 Topic。
3. **不再将 `:id` 视为通配**：`:id` 是合法的普通字符序列；想匹配任意单层应使用 `+`，参数提取由业务自行处理。
4. **不再允许中间 `#`**：`#` 只能独占末尾层级，或单独作为整个 Filter。
5. **保留非 Topic 能力**：`Source`、`Destination`、`LaneKey`、`Header`、`Context`、`Data`、`Send`/`Post`、Drain/Terminate 继续承担原职责；此项改造不等于实现 MQTT Broker。
6. **采用破坏性变更管理**：历史非标准 Filter 必须由 owner 撤销并按当前规则重新登记；不能静默兼容或自动迁移错误语义。
7. **一次通知**：一次事件向同一 Observer 身份最多通知一次；SimpleObserver 的多个匹配 handler 按第 5 节确定的优先级选择一个，不依赖 map 遍历顺序。
8. **发布拒绝可检查**：Hub 增加 `PostChecked(Event) *cd.Error`，`Post` 委托该入口并记录拒绝；`Send` 对非法 Topic 返回带错误的非 nil Result。入队前校验，无非法事件副作用。
9. **寻址保持独立**：Destination 固定采用当前地址模式的双向匹配，独立私有实现；它不是 MQTT Topic，也不是一套可选的旧 Topic 模式。

### 1.3 非目标

本阶段不增加 MQTT TCP Listener、CONNECT/SUBSCRIBE 报文解码、QoS、Retained、Session、ACL、共享订阅分组或跨进程持久消息服务；不重写 magicCommon EventHub 并发执行模型，也不改变 Go 框架其他包的模块角色。

## 2. 改造前源码核查与差异

源码基线：Common `bb440bfe9c`、Base `2ea38c7cab`、Runner `7356d69056`、ModulesRepo `3232a17d7f`。以下表格保留改造前的核查事实，不描述当前实现。当前源码、依赖和测试证据见[实施记录](~/dataspace/magicPlatform/docs/technical/history/eventhub-topic-implementation-2026-11-04.md)。

- [Hub](~/codespace/magicCommon/event/hub.go)：`MatchValue`、订阅控制面、`findMatchingObservers`、`simpleObserver.Notify`、Destination、缓存与 Lane；
- [Event](~/codespace/magicCommon/event/event.go)：`baseEvent.Match` 直接调用 `MatchValue`；
- [Hub 测试](~/codespace/magicCommon/event/hub_test.go)：包括 `/123/#/+/1212/#` 等扩展语法与空层相关预期；
- [Event 测试](~/codespace/magicCommon/event/event_test.go)、[订阅竞争测试](~/codespace/magicCommon/event/subscription_test.go)、[事件使用说明](~/codespace/magicCommon/event/README.md)。

| 差异 ID | 场景 | MQTT 标准预期 | magicCommon 基线行为 | 实施要求 |
|---|---|---|---|---|
| MC-D01 | Filter `a/+`，Topic `a/` | **匹配**（`+` 可匹配空层） | 不匹配 | 必须修复 |
| MC-D02 | Filter `a/#`，Topic `a` | **匹配**（`#` 匹配零层） | 不匹配 | 必须修复 |
| MC-D03 | Filter `a/#/b` | **非法** | 有自定义中间 `#` 逻辑 | 必须拒绝 |
| MC-D04 | Filter `a/:id`，Topic `a/42` | **不匹配** | 匹配 | 必须修复 |
| MC-D05 | Filter `#`，Topic `$SYS/status` | **不匹配** | 缺少 MQTT `$` 保护 | 必须修复 |
| MC-D06 | Filter `a//b`，Topic `a//b` | **匹配**（空层合法） | 文本匹配可能成立，但规则未完整规范化 | 统一校验与匹配 |
| MC-D07 | 非法 Topic Name `a/+` 用作发布 | **拒绝** | Event 构造/入口缺少明确 MQTT Topic 校验 | 必须添加校验边界 |
| MC-D08 | 匹配规则用于 Destination | Destination 不是 MQTT Topic | 当前 `matchDestination` 对 `MatchValue` 做**双向**调用 | 必须解耦 |
| MC-D09 | 同一 Observer 订阅 `a/b` 和 `a/+`，发布 `a/b` | 通知次数是本库合同，不由 MQTT Topic 规则决定 | 最小程序实测 handler 调用两次；Notify 从 map 首个匹配项选 handler | 固定去重和回调优先级 |
| MC-D10 | Runner 用 `/#/query/+` 分类路由 | HTTP 路由不应使用 MQTT Filter 分类 | 当前匹配成功，新规则将拒绝中间 `#` | 同批修改真实消费者 |

> 重要：上述是具体可见差异，并非断言 `magicCommon` 中所有其他 EventHub 功能不符合要求。保留已有功能须通过回归测试确认。

### 2.1 Owner、正式入口与直接消费者

Common 拥有进程内匹配、准入、通知及排空事实；业务是否重试、授权及持久结果仍归业务 owner。HTTP 路由属于 Engine／Runner；计数触发器的定义、登记和恢复属于 Base。下表列明改造前风险及本轮实现范围；最终状态见第 8 节。Topic、Destination 和 HTTP 路由是三个独立合同，不把平台策略或 REST 匹配器下沉到 Common。

| 位置／owner | 基线用途与已确认风险 | 本批必须完成的产物和验证 |
|---|---|---|
| [Runner 中间件](~/codespace/magicRunner/internal/modules/kernel/runner/service/service.go)、[Base 常量](~/codespace/magicBase/pkg/common/const.go) | 中间件以 `UpdateEventMask`／`DeleteEventMask`／`QueryEventMask` 判断是否填充动态 ID；新 matcher 会使三项全部返回 false，代理路径将保留 `:id` | Runner 在注册 query/update/delete 路由时绑定明确的动态参数处理，不再通过 `event.MatchValue` 分类；filter/insert/count 不注入该参数。Base 删除本工作区无剩余消费者的八个 `*EventMask` 常量，不另导出 REST matcher。经 Engine 正式代理核对上游收到真实 ID、HTTP 方法及查询串，覆盖详情、更新、删除和非动态路由 |
| [Engine 代理](~/codespace/magicEngine/http/proxy.go) | 消费动态参数头以替换目标 URL；其自身不使用 EventHub matcher | 原代理合同不变，复跑代理替换测试，并由 Runner 正式链路验证；不把 MQTT 校验加到 HTTP 路由 |
| [Base 计数器](~/codespace/magicBase/internal/modules/blocks/totalizator/biz/totalizator.go)、[公开 DTO](~/codespace/magicBase/pkg/common/totalizer.go) | `fetchTotalizator` 用注册的 Trigger 匹配通知 Trigger；只改 bool matcher 会把非法注册退化成“未登记” | 注册 Trigger 明确为普通 MQTT Filter，通知 Trigger 为 Topic Name；登记／定义解码、运行目录构造及重启加载在写入／发布目录前校验，通知在计数写入前校验；使用 Common 唯一 matcher；全局声明补建所有匹配且缺失的 Filter／owner／租户身份，已有局部匹配也不能跳过其他缺失的全局声明；已有局部身份的周期集合保持不变，不能通过重复克隆恢复局部已撤销的周期，增加周期必须显式重新登记。撤销按精确注册键执行，不能被新增合法性校验阻断清理。覆盖登记、通知、重复登记、局部子集／部分周期撤销后继续计数、显式重新登记恢复、全局撤销、重启加载及超限反例；不自动转换旧持久数据 |
| [Runner Portal](~/codespace/magicRunner/internal/modules/kernel/portal/biz/profile.go)、[计数生产者](~/codespace/magicRunner/internal/modules/kernel/portal/biz/biz.go) | 已将 `:id` 替换为完整 actorKey，再做精确计数键比较；不需要通配 | Profile 改为与生产端生成的精确键比较，不继续借用 Topic matcher；核对账号身份键、反馈／消息计数、租户及应用隔离 |
| [共享 Base Biz](~/codespace/magicModulesRepo/modules/base/biz/biz.go)、[Base 本仓 Biz](~/codespace/magicBase/internal/modules/base/biz/base.go)、[cron](~/codespace/magicModulesRepo/initiators/cron/cron.go) | 包装 Post／广播；共享 Base 的广播 Destination 为 `/#` | 包装层提供返回准入错误的 `PostEventChecked`；必须获知准入结果的调用方检查返回值，其余通知沿 Post 的日志入口；cron 失败可观测，不自动重发、不把准入当完成。证明广播、启动／停止、排空保持；审查全部 Post 调用并记录选择 |
| [Base timer](~/codespace/magicBase/internal/initiators/timer/timer.go) | 定时窗口以异步通知发布 | 与 cron 一样检查准入；拒绝时记录并保留最近一次接纳时间，串行保护窗口，不自动重放业务 |
| [CAS Biz](~/codespace/magicCas/internal/modules/base/biz/biz.go)、[File Biz](~/codespace/magicFile/internal/modules/base/biz/base.go)、[Runner Biz](~/codespace/magicRunner/internal/modules/base/biz/base.go) | 各宿主封装事件发布，CAS 部分通知发生在持久提交之后 | 增加 `PostEventChecked`，已提交业务的通知继续沿 `Post` 日志入口，不把通知拒绝改成业务失败或重放 |
| Base／CAS／File／Runner 等真实宿主及 Hub 替身 | 消费 Hub 接口、精确事件 ID 和带 MatchID 的 Observer | 适配新增方法、冻结事件声明及替身；保留当前 `:id` 字面量事件，只有已证依赖参数通配的订阅才改为 `+`。构建默认 vendor 消费者并验证公开入口、必需订阅失败和退出排空 |

本轮已再次全工作区检索 `MatchValue`、`Event.Match`、Hub／SimpleObserver 实现、订阅表达式、Post 包装及生成的解释器符号合同；包含 README、配置、样例和测试，排除 vendor 后列出 owner，再通过同步脚本更新 vendor。扫描结果及表达式对应清单见实施记录；表中是本工作区确认的消费链，不覆盖未知的外部调用方。发现额外消费者须补入工作包并重新评审其影响，不能当作实施后的零散修补。

历史保护项：Common `2e3212e` 分离订阅身份与 Destination MatchID，`a8e7d3d` 保留原事件 context，`b227fe0` 跟踪真实完成及停机，`9abaf89` 修复订阅准入仲裁。重构不得恢复身份合并、原事件 context 原地修改、迟到订阅或提前释放依赖。

## 3. 唯一标准规范

规范来源：

- [OASIS MQTT 3.1.1 §4.7、§1.5.3](https://docs.oasis-open.org/mqtt/mqtt/v3.1.1/mqtt-v3.1.1.html)
- [OASIS MQTT 5.0 §4.7（Topic Names and Topic Filters）](https://docs.oasis-open.org/mqtt/mqtt/v5.0/mqtt-v5.0.html)

### 3.1 Topic Name（Event.ID）

- 非空、有效 UTF-8、编码字节长度不超过 **65,535**、不得包含 U+0000。
- **不得包含 `+`、`#`**；不对内容进行 Unicode 归一化或大小写折叠。
- `/` 可出现在任意位置，含前后 `/`、相邻 `//` 和仅 `/`；这些形成的零长度层级是合法的。
- `:id`、空格、Unicode 字符等在满足 UTF-8 要求时均为普通 Topic 文本。
- Go 的 `string` 允许存放非 UTF-8 字节，校验入口必须显式使用 `utf8.ValidString`；`len(s)` 是 UTF-8 **字节数**，不是字符数。
- 其他控制字符、非字符的 SHOULD/ MAY 约束应按对应协议版本单独定义接收策略，不得将建议性约束擅自当成两版都要求的 MUST 拒绝。

### 3.2 Topic Filter（订阅表达式）

- 非空、有效 UTF-8、≤ 65,535 字节、不含 U+0000。
- `+` **独占一个层级**，匹配**恰好一个层级**，包含空层；如 `a/+` 匹配 `a/`，但不匹配 `a`。
- `#` **独占最后一个层级**或独立使用，匹配零层或多层；`a/#` 匹配 `a`、`a/`、`a/x`。
- 含 `+` / `#` 的其他形式（`a+`、`a/#/b`、`a/+x`、`a/b#`）一律非法，必须在**订阅注册前**拒绝。
- 以 `#` 或 `+` 起始的 Filter 不得匹配以 `$` 起始的 Topic Name；`$SYS/#` 可以匹配 `$SYS/status`。这是匹配限制，不是禁止 `$` 名称。
- `:id` 没有通配含义：`a/:id` **只**匹配 `a/:id`。

### 3.3 不属于通用 Topic 匹配的协议特性

MQTT 5 共享订阅 `$share/{ShareName}/{TopicFilter}` 要先解析共享组和内部 Filter；组内择一交付与 Broker Session/订阅状态属于后续 Broker，不进入 `MatchValue`。本次只定义普通 Topic Filter 匹配，不剥离 `$share` 前缀或宣称支持共享订阅。

## 4. 公共接口与发布拒绝合同

下列接口和拒绝合同已在本轮源码中实现。错误沿用现有 `*cd.Error`，不扩张全局错误码或新增协议版本：

```go
// Topic 校验：用于发布和 Event.ID 创建/发送前校验。
func ValidateTopicName(name string) *cd.Error

// Filter 校验：用于 Hub.Subscribe 的注册前校验。
func ValidateTopicFilter(filter string) *cd.Error

// 已验证的 Topic Filter，可以编译缓存，供订阅和派发使用。
type TopicFilter struct { /* private parsed levels */ }
func ParseTopicFilter(raw string) (TopicFilter, *cd.Error)
func (f TopicFilter) MatchTopic(topic string) bool

// 保持函数签名，只执行当前唯一 MQTT 单向匹配合同。
func MatchValue(pattern, val string) bool

type Hub interface {
    Subscribe(eventID string, observer Observer) *cd.Error
    Unsubscribe(eventID string, observer Observer) *cd.Error
    // nil 仅证明已入队，不证明已执行或业务提交。
    PostChecked(event Event) *cd.Error
    Post(event Event)
    Send(event Event) Result
    Terminate(ctx context.Context)
}
```

实施要求：

- `MatchValue` 先调用 `ParseTopicFilter(pattern)`，再调用负责 Topic 校验的 `MatchTopic(val)`；非法任何一端返回 `false`，**不能**在外层“猜测方向”以容忍非法 Filter。
- `Hub.Subscribe` 要显式返回非法 Filter 错误，遵循已有 `*cd.Error` 错误契约，保证失败时**不产生订阅副作用**。
- `TopicFilter` 的零值不可匹配任何 Topic。公共 MatchTopic 自行校验 Topic，Hub 的内部已校验路径可复用同一算法，不能每次派发重复解析所有 Filter。
- `NewEvent`／`NewEventWithContext` 保留构造签名，不 panic，也不借 nil 隐藏错误；是否接受发布统一在 Hub 决定。外部实现同样经过入站校验。
- 发布校验顺序：nil Event → 读取并校验稳定的 Event.ID → context → Hub 准入 → Lane 入队。nil／非法 ID 在 `beginOperation`、创建 Lane、写匹配缓存之前拒绝；不调用 handler，也不改订阅。发送后路由字段必须保持稳定，Hub 和 SimpleObserver 均按 `Event.ID()` 调用包内 matcher，不依赖外部自定义 `Event.Match` 的实现。
- `PostChecked` 沿用 100ms 单次入队等待窗口；过期 context／入队超时返回 `cd.Timeout`，停止接收新根调用返回 `cd.InvalidOperation`。成功仅表示接纳一个异步事件，不改变 Post 的独立调用链和已有排空语义；业务重试与持久证明仍由调用方负责。
- `Post` 只调用一次 `PostChecked`；拒绝时通过 `slog` 记录 `event publish rejected`、错误码及有界原因，不输出 Data、Header、context 或完整非法字节串，不重复投递。返回式入口由调用方负责错误处理，Common 不为其重复记录同一日志。
- `Send` 对 nil／非法 ID 返回含 `cd.IllegalParam` 的非 nil Result；对有效事件的取消、准入、已开始执行等待真实返回及停止期间原有行为保持。不能以本次改造顺带将其他返回路径改成新语义。
- Topic／Filter 的空值、UTF-8、NUL、通配语法或 65,535 字节上限错误均为 `cd.IllegalParam`，Message 使用 `invalid topic name` 或 `invalid topic filter` 加固定原因；长度原因明确为 `too long`。资源容量拒绝使用现有 `cd.ResourceExhausted`，与协议长度错误分开。
- Subscribe 和 Unsubscribe 都验证 Filter，并保留原控制面回执语义；SimpleObserver 在 Hub 成功确认后才提交本地回调变更。纯内存订阅不跨进程恢复，配套切换前先排空旧进程，不为非法旧 Filter 增设取消别名。

## 5. 具体修改范围

### MC-CHG-01：拆出独立 matcher（P0）

**已新增** `~/codespace/magicCommon/event/topic.go`（或包内等价文件），集中维护 Topic Name 校验、Topic Filter 解析和匹配。实现建议为分段编译：`Literal` / `SingleWildcard` / `MultiWildcard`；预编译时拒绝不合法 Filter；匹配只接受已校验的 Topic Name。

匹配伪代码：

```text
parseTopicFilter(filter):
  validate UTF-8, length, NUL
  split by "/" preserving empty levels
  reject non-final/partial #, partial +
  return compiled levels

matches(compiled, topic):
  validate Topic Name
  if topic starts "$" and raw filter starts (+ or #): return false
  walk levels in order:
    Literal: equals one topic level
    +: consumes exactly one level (including "")
    final #: consumes zero or more levels
  success iff topic levels consumed and filter levels consumed
```

注意 `#` 的零层匹配，在 `a/#` 对 `a` 的场景不能先强制要求 Topic 包含 `/`；按父节点路径成功即可匹配。

### MC-CHG-02：规范化 Event.Match / Hub 派发（P0）

- [Event](~/codespace/magicCommon/event/event.go)：`baseEvent.Match` 改为唯一 matcher；保留 Source、Destination、Header、Context 与 Data 不变。
- [Hub](~/codespace/magicCommon/event/hub.go)：`findMatchingObservers` 使用预编译 Filter；注册表、Filter 和缓存在同一订阅读写事务边界内更新，不另建独立漂移的 generation。Observer 结果集按稳定 `ID()` 去重。
- 修改订阅时更新/失效匹配缓存，杜绝“取消订阅已确认，但旧缓存继续命中”；保留既有控制面同步语义。
- 派发遵循单向 `Filter -> Event.ID()`，不把 Event Topic 当成 Filter。

**重叠订阅的固定规则**：

1. 同一事件无论命中同一 Observer 的多少个 Filter，Hub 只调用其 `Notify` 一次；不同 ID 即使共用 MatchID，也各收到一次。不同 Observer 之间的执行顺序不新增保证，同步广播的结果汇总合同也不在本次改变。
2. 同一 Observer 实例的同一 `(Filter, Observer.ID)` 重复 Hub.Subscribe 维持幂等；跨 Filter 注册使用同一身份记录，最后一个 Filter 撤销才释放记录。另一 Observer 实例占用相同 ID 返回 `cd.Duplicated`，不得按 map 顺序选择对象或静默丢失其 handler；同一 ID 的 MatchID 不允许漂移，冲突返回 `cd.IllegalParam`，均无副作用。对象核对必须检查实际值的可比较性，包括 interface 字段、嵌套结构或数组内的切片／map，不能仅按类型的 Comparable 判定，也不得直接使用 `==` 导致 panic；无法证明为同一实例时按身份冲突拒绝，不影响首个合法注册。独立对象必须使用不同 ID；两个 ID 共用 MatchID 保持合法。SimpleObserver 对本地同 Filter 重复 handler 仍返回 `cd.Duplicated`。
3. SimpleObserver 对当前全部匹配 handler 只选一个：完全字面量精确匹配优先；其余逐层比较，第一次不同按“字面量 > `+` > `#`”优先；共同层相同则层级更长优先；仍相同以原 Filter 的 UTF-8 字节升序决定。空字面层也属于字面量，不通过注册顺序或 map 顺序决定。
4. 例：`a` 与 `a/#` 对 Topic `a` 选择 `a`；`a/b` 与 `a/+` 对 `a/b` 选择 `a/b`；`a/+` 与 `a/#` 对 `a/b` 选择 `a/+`；`a/+/c` 与 `a/b/#` 对 `a/b/c` 选择 `a/b/#`。每次仅一个回调，写计数不能翻倍。
5. 退订只移除指定 Filter；若其他 Filter 仍命中，下一次事件按同一优先级选择。全部撤销后不再命中；已选取／执行的通知仍按原生命周期完成，退订不等于排空。
6. Notify 不再调用可被外部 Event 实现覆写的 `ev.Match` 来裁决，避免 Hub 已选中却无 handler；`baseEvent.Match` 和独立 MatchValue 仍复用同一 MQTT 算法。回调执行时不持有订阅锁，保持原 panic 防护。

### MC-CHG-03：Destination 匹配与 Topic 匹配拆分（P0）

改造前 [Hub](~/codespace/magicCommon/event/hub.go) 的 `matchDestination(destination, matchID)` 使用 `MatchValue(destination, matchID) || MatchValue(matchID, destination)`。如果 `MatchValue` 标准化却仍用于这里，会无意改变**非 Topic** 的目标寻址规则。

**固定决策**：将基线 `MatchValue` 的现有地址模式算法原样移入私有 `matchAddressPattern`，`matchDestination(d, m)` 固定为 `matchAddressPattern(d, m) || matchAddressPattern(m, d)`。此算法只有 Destination 消费，不导出、不供 Topic／HTTP／计数复用；不新增开关或过渡模式。本次不修订它的空层、零层及中间 `#` 行为，不对地址套用 Topic 的长度、UTF-8、NUL、`$` 或通配校验。

MatchID 解析规则保持：显式非空 `MatchID()` 优先，否则回退到 `Observer.ID()`；该回退只用于没有显式寻址模式的观察者。Observer.ID 仍是订阅去重／退订身份，不以 MatchID 合并两个 ID。LaneKey 只决定顺序域，不参与匹配。

以下是已用基线算法核对的双向寻址结果；本轮已把原 matcher 的完整输入向量保留为 `TestAddressPattern` 地址回归，不能只复制 Topic 新测试：

| Destination | 有效 MatchID | 是否送达 |
|---|---|---|
| `/modules/a` | `/modules/a` | true |
| `/#` | `/modules/a` | true |
| `/#` | `owner` | false |
| `#` | `owner` | true |
| 空字符串 | `owner` | false |
| `a/:id` | `a/42` | true |
| `a/42` | `a/:id` | true |
| `a/+` | `a/` | false |
| `a/#` | `a` | false |
| `a/#/b` | `a/x/b` | true |
| `#` | `$SYS/status` | true |
| `/#` | `/` | false |

纯地址函数的空串／空串结果为 true；正常 Hub 注册要求非空 Observer.ID，空 MatchID 会回退，空 Destination 不因此成为广播。共享 Base 的 `/#` 广播只覆盖当前地址合同命中的目标，不能自动扩为所有 Observer。上述有意保留的寻址差异不得宣传为 MQTT 匹配能力。

### MC-CHG-04：错误与限制（P0）

- 严格使用第 4 节固定的错误码与原因，不新增 `InvalidTopicName` 等全局 Code；Hub 接口、SDK／包装和测试替身在同一批适配 PostChecked。
- 协议合法性（≤ 65,535 字节）与可选本地资源上限是两种不同错误，不得混同。
- 解析失败不注册、无匹配副作用；若事件本身非法，则拒绝发布或提供可观察失败结果。

### MC-CHG-05：文档、迁移与发布（P1）

- 同步[事件使用说明](~/codespace/magicCommon/event/README.md)、公共 API 注释、[Event 测试](~/codespace/magicCommon/event/event_test.go)、[Hub 测试](~/codespace/magicCommon/event/hub_test.go)、[订阅测试](~/codespace/magicCommon/event/subscription_test.go)、发行说明；全面搜索源码中 `:id`、中间 `#`、空层预期和 `MatchValue` 用法。
- 删除与 MQTT 规则冲突的“成功匹配”测试，替换为“非法 Filter 拒绝”测试；保持原用例作为破坏性变更记录。
- 原地址用例转入 Destination 私有匹配测试继续保护寻址，不能直接删除历史反例。跨仓处理完成前不得发布为“Topic 标准化已交付”；实际行为变化时再同步使用说明，设计文档不能提前改写当前能力声明。

## 6. 最低验收矩阵（magicCommon 必须单独完整通过）

> 格式：`Filter | TopicName | 预期`；若 Filter 或 Topic 非法，优先验证“注册/发布时拒绝”，不以匹配布尔结果代替合法性判断。

| ID | Filter | Topic Name | 预期 |
|---|---|---|---|
| MC-T01 | `device/+` | `device/123` | true |
| MC-T02 | `device/+` | `device/` | true |
| MC-T03 | `device/+` | `device` | false |
| MC-T04 | `device/#` | `device` | true |
| MC-T05 | `device/#` | `device/` | true |
| MC-T06 | `device/#` | `device/a/b` | true |
| MC-T07 | `device/#/status` | `device/a/status` | invalid Filter |
| MC-T08 | `device/:id` | `device/42` | false |
| MC-T09 | `device/:id` | `device/:id` | true |
| MC-T10 | `#` | `$SYS/status` | false |
| MC-T11 | `+/status` | `$SYS/status` | false |
| MC-T12 | `$SYS/#` | `$SYS/status` | true |
| MC-T13 | `device//status` | `device//status` | true |
| MC-T14 | `/` | `/` | true |
| MC-T15 | `a/+` | `a/b/c` | false |
| MC-T16 | `a/b` | `A/b` | false |
| MC-T17 | `a+` | `a+` | invalid Filter（同时 Topic 非法） |
| MC-T18 | `a/+x` | `a/x` | invalid Filter |
| MC-T19 | `a/#` | `a/#` | invalid Topic Name |
| MC-T20 | `+/+` | `/finance` | true |

补充边界：空 Filter/Topic；非法 UTF-8 字节串；U+0000；65,535/65,536 字节；多字节 UTF-8 长度边界；大小写与 Unicode 不归一化；前导/末尾斜线；`$SYS` 与普通 Filter 的组合；并发订阅/取消/发布；重复订阅身份、匹配缓存与 Destinations。

### 6.1 测试分层

1. **纯函数**：`ValidateTopicName`、`ParseTopicFilter`、`TopicFilter.MatchTopic`、`MatchValue` 逐向量通过。
2. **Hub 集成**：真实 `Subscribe` -> `Post` / `Send` -> Observer 收到与预期一致；非标准订阅明确返回错误。
3. **Destination 回归**：第 5 节表全部覆盖，追加两个 ID 共用 MatchID、独立退订、Lane 与路由分离；证明它不影响 Topic 校验。
4. **并发/停止**：Lane 内顺序、不同 Lane 并发、Drain/Terminate、订阅时序与缓存失效保留旧合同。
5. **性能**：预编译 Filter、exact-match 命中和高订阅量场景与修改前基准比较，不能凭功能正确推断性能未退化。

### 6.2 必须补齐的集成反例

| ID | 场景 | 验收结果 |
|---|---|---|
| MC-I01 | 同 ID 订阅 `a`、`a/#`，Send／Post Topic `a` | Notify 与精确 handler 各一次；撤销 `a` 后仅 `a/#` handler 一次；最后撤销后零次 |
| MC-I02 | `a/b`／`a/+`、`a/+`／`a/#`、`a/+/c`／`a/b/#`，反转注册顺序并重复发布 | 按固定优先级选择同一回调，每次计数只增一；不同 ID 不被去重 |
| MC-I03 | 非法 Filter 注册失败、不同对象同 ID、同 ID MatchID 冲突、不可比较的自定义 Observer 值 | 错误码符合固定规则，且无 panic；Hub、SimpleObserver、编译表、缓存全部不变，之后合法注册与退订仍可完成 |
| MC-I04 | nil／非法 Event 经 PostChecked、Post、Send；含外部 Event 实现 | 返回码／Post 拒绝日志符合第 4 节；Lane、缓存、handler 无副作用，合法事件仍工作 |
| MC-I05 | 外部 Event.Match 固定返回 false，但 ID 合法且命中 | Hub／SimpleObserver 按 ID 正常交付一次，不被外部 matcher 改写路由 |
| MC-I06 | PostChecked 遇取消、队列超时、停止新根准入、成功入队 | 拒绝码正确；拒绝不迟到投递，成功不被写作业务完成；在途依赖和排空语义保持 |
| MC-I07 | 非法 Topic 在忙 Lane／满队列时发送 | 仍优先返回 IllegalParam，不能先阻塞或创建新 Lane；Post 不打印完整输入 |
| MC-I08 | Runner 正式代理 query/update/delete 与 filter/insert/count | 上游 ID、方法、查询串和授权边界正确；非动态路由不被误分类，不依赖非法 MQTT mask |
| MC-I09 | Base 计数 Filter 登记／写入／加载、通知、撤销；Portal 反馈／消息计数 | 非法输入在正式写入前报错；合法计数与精确个人摘要一致，多租户／双实例无串扰，撤销清理不被阻断 |
| MC-I10 | Common 接口更新后的共享 Biz、cron 及全部真实宿主 | Hub 替身编译、必需订阅失败、启动、广播、取消、重复停机及真实完成保持；解释器生成符号如受影响须同步生成 |

性能比较使用同机同参数基线：Topic 长度覆盖 65,535 字节，订阅数至少 1／100／1,000，覆盖精确命中、通配、重叠及无命中。记录 ns/op、allocs/op、峰值内存和冷／热缓存，不机械承诺吞吐阈值；若出现随订阅数重复解析／分配或明显退化，须分析并解决后再验收。订阅／匹配缓存现有容量策略不在本次擅自扩张。

建议执行（根据仓库实际环境调整）：

```bash
go test ./event -count=1
go test -race ./event -count=1
go test ./... -count=1
go vet ./...
```

其中 `go test ./...` 若依赖数据库或外部环境，必须在报告中标记环境限制，不能把未运行项写作通过。

Common 通过后，先按工作树／暂存区保护规则，从 `~/codespace` 执行 `~/codespace/upgrade_vendor.sh magicEngine magicOrm magicBase magicCas magicModulesRepo magicFile magicRunner`；审查全部生成差异，再以默认 vendor 模式构建和验证这些消费者，按实际新增依赖调整目标。随后完成 Runner 正式代理和 Base 计数公开入口验证。本轮已保护工作树和暂存区、执行两轮上述同步并审查生成差异；实际命令和默认 vendor 终验见实施记录。源码验证不包含镜像构建或部署。

## 7. 成套切换及历史定义处置

1. 枚举第 2.1 节全部生产／消费入口、配置、README 与测试，生成 `原表达式／所属合同 -> 当前 Filter 或专用处理` 清单。单列 Topic、地址、HTTP 路由及计数定义，不批量字符串替换。
2. 只有确实表达任意单层的 `a/:id` 订阅才改为 `a/+`，在 Handler 提取参数。若生产者和消费者都发布／订阅字面量 `:id` 事件，保留精确事件名，例如 Base `bc.QueryEntity`；HTTP 的 `:id` 路由仍由 Engine 解析。
3. `a/#/status` 无等价 MQTT 单 Filter，改为合法的更宽订阅（如 `a/#`）+ Handler 精确过滤，或重新设计 Topic 树；必须评估范围扩大带来的 ACL 与业务副作用。
4. `a/#` 匹配集合会**扩大**到 `a`，`a/+` 也将匹配 `a/`；不能仅检查原有正向用例，应加负向回归与安全审查。
5. Destination 按第 5 节确定的独立算法保留；重叠订阅按通知去重和优先级清单核对，不能把原有重复写入当成必须兼容的功能。
6. 同批更新 Common、直接消费者、SDK／Hub 替身、夹具／定义、解释器生成物（如涉及）和说明；正常 Git 发布版本按项目约定处理，内部协议／Schema 不递增，不保留双格式、别名、自动回填或迁移。
7. Base 计数表示中的注册 Filter／通知 Topic 使用唯一当前合同；采用配套组件与干净候选定义验证。此次不改变物理 Schema，不承诺将历史计数 Trigger 自动转换为新 Filter；已有环境若有非法定义须先列出 owner 正式撤销／重新登记及业务影响方案，不能在加载时忽略、临时改成宽匹配或修改数据库制造通过。
8. 实施交付先停止外部生产者，等待原调用真实完成并排空旧 Hub，再切换配套宿主。事件 Hub 无持久订阅迁移；不默认混搭或滚动兼容。尚未写入新合同数据时可整体回退成套旧制品；产生新定义／计数数据后须先评估旧组件能否读取，不以单仓回退或清库替代 owner 处置。

## 8. 验收条款与实施任务

| Contract ID（新增） | 验收内容 | 最小测试 |
|---|---|---|
| MC-TOPIC-001 | Topic Name / Filter 区分校验、长度与 UTF-8 | MC-T01—T20 + 边界 |
| MC-TOPIC-002 | `+` 空层、`#` 零层、`$`、`:id` 字面量 | 单元 + Hub 集成 |
| MC-TOPIC-003 | `Event.Match`、`MatchValue`、Hub 路由一致 | 同一向量多路径对照 |
| MC-TOPIC-004 | 非法订阅无副作用，退订后缓存及时失效 | 订阅控制面竞争测试 |
| MC-TOPIC-005 | Destination 不误用 Topic 标准；Lane、同步/异步语义不回退 | 回归测试 |
| MC-TOPIC-006 | 文档、迁移、发行说明和执行证据齐全 | 发布门禁 |
| MC-TOPIC-007 | 重叠订阅一次通知、回调确定、独立退订 | MC-I01—I03 |
| MC-TOPIC-008 | 非法发布检查式拒绝及日志，无入队副作用 | MC-I04—I07 |
| MC-TOPIC-009 | Runner HTTP 代理和 Base／Portal 计数链配套 | MC-I08—I10 + 默认 vendor 终验 |

### 8.1 实施顺序及状态

| 工作包 | 产物／退出标准 | 当前状态 |
|---|---|---|
| MC-01 合同与消费清单 | 本文决策及工作区调用与测试映射 | 已完成本工作区扫描及表达式清单；不覆盖未知外部消费者 |
| MC-02 Common | MQTT matcher、PostChecked、去重与优先级、Destination 私有算法；MC-T／MC-I Common 部分通过 | 已实现；MQTT 向量、拒绝、地址、并发与生命周期测试通过 |
| MC-03 Base／Runner／ModulesRepo／CAS／File | 路由分类、计数准入及摘要、包装与替身、清理及恢复检查 | 已完成适配；真实代理、Portal 摘要及隔离 PostgreSQL 计数链通过，无旧 Topic 模式 |
| MC-04 依赖与源码验证 | vendor 保护／同步、七仓消费者默认模式、race、代理和计数正式链验证 | 已完成两轮依赖同步；七仓全量检查通过，Runner 全仓保留短预算失败，专项复跑通过；性能样本与限制见实施记录 |
| MC-05 制品与 Live | 配套镜像身份、宿主健康、实际代理／计数／广播及排空证据 | 待配套交付和授权环境；不自动部署 |
| MC-06 完成交付 | 实际使用说明、发行记录、逐合同证据和遗留项处置 | 源码说明与实施证据已更新；制品及 Live 门禁仍开放，整体未关闭 |

以上工作包没有持久 Hub 工作流或失败任务需要迁移；本轮没有执行生产任务或修改业务数据，原失败资源处置不适用。若实施验证发现业务残留，按精确原请求及业务 owner 另记处置，不能由 Common 清理业务表。

### 8.2 改造前评审证据与本轮验证范围

以下前四项保留改造前评审证据，不混作新合同的验证结果。

- 已核查第 2.1 节源码调用及列明的四个 Common 历史保护提交，未宣称逐行审计所有项目历史。
- 评审运行 `go test ./event -run '^(TestMatchID|TestSubscriptionRejectionPreservesLocalStateAndAllowsRetry|TestUnsubscribeInvalidatesConcurrentMatchCache)$' -count=1` 通过；其中最后一个选择项当前无同名测试，不计为已覆盖的缓存检查。
- 随后运行 `go test -race ./event -run '^(TestMatchID|TestSubscriptionRejectionPreservesLocalStateAndAllowsRetry|TestUnsubscribeCannotBeUndoneByConcurrentCachePublication|TestHubCacheRespectsDestination)$' -count=1` 通过，覆盖实际缓存竞争与定向匹配基线。
- 临时最小程序实测 `/#/query/+` 对 `/device/query/:id` 当前为 true；同 Observer 的 `a/b`、`a/+` 重叠订阅使单次发布调用 handler 两次。刷新时另对第 5 节寻址边界取基线结果；这些反例证明现状，不证明目标实现。
- 本轮已完成 Common 源码、全部已知直接消费者、Hub 替身和文档适配，已执行 vendor 同步、源码检查、race、正式 Engine 代理及隔离数据库计数链验证。完整命令、逐合同证据、性能数据和非 MQTT 业务测试的初次失败均见[实施记录](~/dataspace/magicPlatform/docs/technical/history/eventhub-topic-implementation-2026-11-04.md)。
- Runner 代理测试隔离了 CAS 认证，Portal 测试的持久 owner 是替身；真实 PostgreSQL 验证的是 Base owner，不证明跨服务 Bearer 授权和 Live 通过。未构建新镜像、发布包或执行部署，MC-05 和最终交付门禁保持开放。

**完成定义（Definition of Done）**：Common、直接消费者、默认 vendor、正式代理／计数链、并发与性能回归、配套制品／Live、切换与清理证明、README、发布说明及实测记录齐全；每条 Contract 分别记录证据与缺口，才能将整个工作包标记完成。**当前源码及已知消费者已实施，制品和 Live 尚未验收，整个工作包不能标记关闭。**
