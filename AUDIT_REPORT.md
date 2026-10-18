# 代码与业务链路审计报告

审计日期：2026-10-05 至 2026-10-06  
审计对象：本地 `dujiao-next` 工作树与交接文档指向的云端运行环境，重点覆盖普通购买、代理操作、邀请注册/归因、支付履约及生产部署边界。  
方式：检查本地代码/文档/部署脚本并追踪关键 API 与持久化调用；先对云主机执行只读服务/配置/权限检查，随后在用户明确授权下清理指定日志、创建并验证备份、部署代码及执行配置权限加固。未输出账户凭据或订单个人信息，也未创建/修改业务记录。

## 结论

**本轮已把认证/邀请、CORS 和自动履约恢复修复部署到云端，并完成运行账号、文件权限与 systemd 隔离加固；但生产购买仍未达到可验收状态。**复核发现线上没有任何支付渠道配置（`payment_channels=0`），订单/支付数据也为 0，因此不能完成或证明真实付款、回调、交付与代理结算闭环。真实支付渠道凭据、邮件投递、供应商履约、退款冲正、异地备份恢复与安全响应头策略仍需验收；不能承诺“无风险”。

普通购物与代理推广下单的后端主路径已有较完整实现：订单服务在预览和正式创建时重新解析代理定价，并保存订单分销快照；这可以避免仅信任前端价格。现有测试支持这些基础行为，但不足以证明真实支付渠道回调、自动履约、退款/拒付与代理利润冲正全链路安全，也不等于生产配置与部署已验证。

## 初次本地审计发现（修复前快照，后续修复见下文）

### P0 — 登录接口允许客户端自行提权为代理

文件：[user_login_handler.go](internal/modules/identity/userauth/transport/http/user_login_handler.go#L256)

登录成功后，服务端只检查请求体中的 `req.Role == "agent"`，随即调用 `SetAgentLevel`，不检查用户当前是否为代理、是否有代理档案或邀请码。该字段由客户端控制，攻击者可直接提交：

```json
{"email":"普通用户邮箱","password":"正确密码","role":"agent"}
```

随后会将会员等级更新为固定 ID `2`。这使普通用户能取得代理控制台能力以及相关代理配置/财务功能访问资格（具体可用范围取决于代理资料的懒创建和路由鉴权）。这是明确的服务端授权缺陷；前端是否显示“代理登录”不构成安全控制。

### P1 — 注册完成后忽略邀请码核销失败，且核验与核销不是原子操作

文件：[user_login_handler.go](internal/modules/identity/userauth/transport/http/user_login_handler.go#L177)、[invite_code_store.go](internal/modules/reseller/infrastructure/gormstore/invite_code_store.go#L84)

注册前先调用 `VerifyInviteCode`，创建用户后才单独调用 `UseInviteCode`；核销错误被丢弃，仍返回注册成功和登录令牌。并发请求可同时通过核验，之后超过 `MaxUses` 的账号仍可能创建；数据库/存储错误同样会造成“代理注册成功但邀请码未消耗”。邀请码使用计数和新账号/代理等级变更也没有共同事务。因此“最大使用次数”不是严格配额，审计记录也可能与实际开通数不一致。

### P1 — 邀请码字段会触发邮件验证豁免，但是否创建代理取决于另一个可选字段

文件：[user_login_handler.go](internal/modules/identity/userauth/transport/http/user_login_handler.go#L98)、[user_login_handler.go](internal/modules/identity/userauth/transport/http/user_login_handler.go#L126)

`Role == "agent"` **或**只要提供任意非空 `InviteCode`，都会进入代理注册分支；该分支设置 `AgreementAccepted = true`，有效邀请码且未填验证码时还会关闭邮箱验证码要求。可是最终升级和核销只在 `Role == "agent"` 时执行。于是请求可带一个有效邀请码但不带 `role=agent`，注册出已标记邮箱验证通过的普通账户，而且邀请码不消耗。邀请链接/邀请码若泄露，便成为邮件验证和协议勾选的通用绕过入口。代理开通免验证码是否符合业务预期，不代表普通账户注册也应因此绕过。

### P2 — 邀请码有效期在管理界面/API 未实现，交接描述不符

交接文档称可配置有效期；但管理接口请求结构只有 `Code/Prefix/Count/Memo/MaxUses`，没有 `ExpiresAt`，生成记录时也没有设置有效期：[admin_invite_code_handler.go](internal/modules/reseller/transport/http/admin/admin_invite_code_handler.go#L61)。管理 UI 只配置最大使用次数与备注，没有过期时间输入（`frontend/admin/src/views/admin/ResellerInviteCodes.vue`）。数据库模型和校验器虽支持 `ExpiresAt`，但当前提供给管理员的创建流程无法设置它。容易造成管理员误以为邀请码会过期，实际邀请码长期有效。

### P2 — 代理身份使用硬编码会员等级 ID

文件：[service.go](internal/modules/identity/userauth/application/service.go#L59)

`SetAgentLevel` 将等级写死为 `2`，用户资料响应也直接回填 `MemberLevelID = 2`。若某环境的等级种子、迁移或人工调整后 ID 2 不再是代理等级，会升级到错误等级或数据库更新失败；调用端忽略写入错误仍回成功。应按稳定的等级代码/配置查找，并将失败作为开通失败处理。这个问题也放大 P0 提权的影响。

### P2 — 代理邀请生成接口对批量部分失败没有告警

文件：[admin_invite_code_handler.go](internal/modules/reseller/transport/http/admin/admin_invite_code_handler.go#L108)

随机码生成失败或唯一键冲突都会被 `continue` 忽略，最后仍返回成功，只在 `count` 中给出实际数量。管理员批量生成时可能拿到少于请求数量的邀请码而没有失败原因；这会造成操作误判与重复操作。

## 已检查链路与正向证据

- **普通用户购买**：商品展示/结账以 API 预览和服务端正式创建结果为准；客户端仅提交商品、数量等输入。普通路径的支付处理不在本次范围内作生产级认证结论。
- **推广链接归因**：前端捕获 `?agent=...` 并保存到本地存储/Cookie，后续请求以 `X-Agent-Code` 提交；后端只接受活动代理档案。请注意归因可由用户主动覆盖，这是推广归属策略，不应当承担授权功能。
- **代理定价**：订单正式创建时再次使用代理设置计算单价，校验商品/SKU上下架和底价/上限，并写订单分销快照；预览和正式创建均经过该解析器。相关实现位于 `internal/modules/order/application/reseller_pricing.go` 与 `order_service.go`。
- **安全边界**：推广码选择代理租户本身不能代替登录授权；P0 说明代理身份判定仍错误地依赖请求字段。
- **交接文档差异**：文档名为 `HANDOVER.md`（而非根目录 `HANDOVERR.md`）。其代理邀请有效期描述与当前创建 UI/API 不一致；“登录/注册”描述也没有披露服务端根据登录请求 `Role` 改级的授权风险。

## 验证结果

- `go test ./...`：通过。
- `frontend/user` 的 `npm run build`：通过；有依赖注释、过期 Browserslist 数据和 Node 弃用提示，不影响本次构建结果。
- `frontend/admin` 的 `npm run build`：通过；有过期 Browserslist 数据提示。
- 初次审计时测试未覆盖登录提权、邀请码并发核销、注册字段组合绕过和管理员有效期配置；对应服务端回归测试已在后续本地修复中补充。仍没有执行真实支付、供应商履约、邮件送达、提现线下核销或生产部署演练。

## 最初建议（已实施于本地；云端未上线）

1. 移除登录请求中的角色升级行为；代理资格只从可信数据库状态确认，客户端 `role` 不参与授权。
2. 将邀请码核销、用户创建、代理等级/档案开通设计为事务或可恢复的一致性流程；核销失败不能返回成功；对最大次数使用数据库条件更新/锁避免并发超用。
3. 明确注册合同：只有代理注册模式才接受邀请码并授权相应验证豁免；普通注册无论是否附带邀请字段都必须走普通验证与协议规则。
4. 使用等级代码而非 ID `2`，并对升级、资料创建失败回滚/报错。
5. 补齐邀请码过期时间管理，或修正文档明确未提供该功能；批量生成返回明确的成功/失败数量和原因。
6. 增加上述边界的服务端回归测试，再按支付渠道逐项演练付款通知幂等、金额核验、退款、自动交付失败重试和代理账务冲正。

## 范围限制

本次以 `dujiao-next` 作为目标业务系统审查。工作区中的 `zovo_card_cdk_auto` 是独立仓库，本次未对其全部源码/部署脚本作全面审计；仅检查了 Dujiao 侧代理/订单相关调用索引。若它参与实际充值交付，应另做供应商接口、凭据、回调签名、重放和补单专项审计。云端只做只读健康/安全核查，没有用真实支付、邮件或供应商凭据进行端到端联调，也没有执行生产交易或部署。

## 修复跟进（2026-10-05）

针对上面的发现，本轮已实施以下更改：

- 登录流程彻底忽略客户端 `role`，移除“登录时升级等级”的副作用；代理注册使用单独的 `RegisterAgent` 路径，按 `agent` 等级 slug 查找 ID 并在用户首次写入时赋等级。
- 普通注册携带邀请码会被拒绝；代理注册必须同时指定代理注册模式、有效邀请码并明确同意条款。前端移除了本地伪造代理资料/一键激活，并改为从后端代理档案判断身份。
- 邀请码核销与用户创建现已放入同一个 `AuthUnitOfWork` 数据库事务：核销用带状态、有效期、使用上限条件的原子更新，任何用户创建失败都会回滚邀请码计数。新增并发核销与事务提交/回滚测试。令牌签发发生在提交后；若 JWT 配置/签发异常，账号和邀请码会保持已提交状态，用户可重试登录，但生产配置仍需验证。
- 邀请码有效期已连通管理 UI 和 API；过期时间必须晚于当前时间，列表显示过期状态。批量生成会重试码冲突并返回成功/失败数量，前端明确提示部分失败。
- 修复 `MaxUses=0` 被 ORM schema 默认值覆盖的问题；后台未传该字段时默认单次，明确传 0 时为无限使用。
- 新增代理等级 slug 映射、普通账户不自动获得档案、邀请注册与角色伪造等测试。

修复后的验证结果：`go test ./...` 通过；`frontend/user` 与 `frontend/admin` 的 `npm run build` 均通过。前端构建仍会输出第三方依赖注释、Browserslist 数据过期及 Node 弃用提示，不影响构建成功。

## 本轮作业总结（供后续接手）

- 已处理本报告列出的授权提权、代理注册路径、邀请码并发配额/过期、代理等级硬编码、批量生成反馈及前端假代理状态等问题；核心修改涉及 `internal/modules/identity/userauth`、`internal/modules/reseller`、`internal/modules/memberlevel`、`internal/bootstrap/userauth` 与两套前端。
- 回归证据：全仓 `go test ./...`、`go vet ./...`、后台 `npm run build`、用户端 `npm run build` 均成功；支付后自动交付恢复查询有 GORM 测试覆盖。
- 仍需业务方在真实环境核对 `agent` 等级 slug/种子、代理注册条款和邮箱验证策略；联调真实支付渠道回调、邮件投递、供应商履约/失败重试、退款与代理账务冲正、备份恢复及部署回滚。
- 尚未全量审计独立仓库 `zovo_card_cdk_auto`，也未对生产支付、邮件、供应商履约执行真实交易联调。云端仅做只读核查；不能据此承诺“无风险”或声称已通过生产上线验收。
- 本地还修复了凭据型 CORS 默认放行和支付成功后自动交付入队失败缺乏数据库恢复扫描的问题；这些补丁目前尚未部署到云端。

## 部署前云端只读复核快照（2026-10-06）

### P0 — 登录请求曾允许提权（修复后状态见部署后复核）

交接文档所指主机可访问，API 与 worker 的 systemd 服务均为 active。云端 Git HEAD 与本地同为 `d2e44618`，但这只代表共同基线相同：本地有 57 个跟踪文件改动、云端有 60 个；只读 `rsync --dry-run` 显示关键认证/邀请码/前端修复文件仍与云端不同。云端源码 `internal/modules/identity/userauth/transport/http/user_login_handler.go` 仍含客户端 `req.Role == "agent"` 后调用 `SetAgentLevel` 的登录逻辑，并且仍先验码、创建用户、忽略邀请码核销错误。不能把“同 commit”或“服务 active”当作修复已上线的证据；线上二进制也未通过运行时源码指纹确认包含本地补丁。

### P1 — 线上 CORS 曾对任意来源开放凭据（现已修复并验证）

对公网 API 发出的只读 OPTIONS 请求，使用不受信任的 `Origin: https://evil.example`，线上返回该 Origin 的 `Access-Control-Allow-Origin`、`Access-Control-Allow-Credentials: true`。本地仓库的 CORS 默认值和 `config.yml.example` 也原先配置了 `*` 与 credentials，middleware 会反射请求 Origin。项目还为 Google 重定向认证设置跨站 `SameSite=None` HttpOnly cookie，因此这一组合扩大了跨站读取认证流程响应的风险面。

本轮已在代码中修复并部署：默认不允许跨域来源、不允许凭据；通配符与凭据同时配置时拒绝来源；只有显式精确来源允许时才发送凭据头。云端公网 OPTIONS 复核显示不受信任 Origin 未获得 `Access-Control-Allow-Origin`。若有跨源 API 客户端，应在配置中提供精确可信 Origin。

### P1 — 云端服务曾以 root 运行且数据库/配置文件对所有本地用户可读（现已加固）

部署后 API 与 worker 均由 `dujiao` 专用账号运行；生产配置权限为 `0640 root:dujiao`，SQLite、上传、日志目录仅授予服务账号访问。systemd 使用 `ProtectSystem=strict`、`NoNewPrivileges`、独立临时目录、限制写路径等隔离项；`systemd-analyze security` 当前整体暴露评分为 `4.2 OK`。

### P1 — 云端此前没有 SQLite 备份；现有同盘可验证备份，仍缺异地副本

GPT 旧日志清理前根盘使用率约 94%、可用约 13 GB。按用户授权，仅删除了从日志采样可识别为 GPT 模型的 19 个旧文件（2026-08-17 至 2026-08-21），释放 `9,268,706,183` 字节；所有 Claude 日志及 2 个无法识别模型的文件均保留。清理后根盘约 90%、可用约 21–22 GB。已在 `/var/backups/dujiao-next/maintenance-20261005T190124Z` 建立配置、二进制、systemd 单元和通过 SQLite 在线备份 API 创建的数据库快照；权限为 0600/0700，SHA-256 清单通过校验，快照及独立恢复副本的 SQLite `quick_check` 均为 `ok`。这是同盘备份，尚未复制到异地，仍需监控磁盘并建立异地备份。

### P1 — 交接中的 rsync 命令会覆盖线上配置和运行数据

原 `HANDOVER.md` 同步命令没有排除 `config.yml`、`db/` 和 `uploads/`。仓库根目录存在被 Git 忽略的本地 `config.yml`；照抄命令会把它覆盖到云端，可能覆盖线上密钥、运行模式、数据库配置及第三方配置，且也可能同步本地数据库/上传数据。本轮已修正交接命令，明确排除配置、数据库、上传、日志与 `.env*`，并注明部署前备份要求。不要继续使用旧命令。

### P1 — 支付成功但 Redis 暂时不可用时，自动交付任务可能永久丢失

原支付回调先提交成功支付/订单状态，再向 Asynq 入队自动交付；入队失败仅记日志并仍确认回调。后续重复成功回调走幂等分支，不重新入队，因此短暂 Redis 故障可能留下已付款未交付订单。

本轮新增的每分钟恢复扫描已随 worker 部署：查询数据库中已付款、所有商品均为自动交付且没有 fulfillment 记录的订单，再走原有事务化 `CreateAuto` 逻辑；测试覆盖混合/人工/取消/已完成和拆单子订单。这降低 Redis 故障造成永久漏单的风险，但不替代完整事务 outbox；仍需监控积压、库存不足及队列状态。

### P1 — 当前没有支付渠道配置；P2 公网安全响应头不完整

云端 `payment_channels` 表有 0 条记录，虽然商品/SKU、代理配置和邀请码数据存在，但当前不能据此验证用户可选择支付方式或完成购买。已将 `server.mode` 切为 `release` 并启动成功。公网首页有 HSTS 和 `X-Content-Type-Options: nosniff`，但未观察到 CSP、`X-Frame-Options` 或 `Referrer-Policy`；需确定由边缘还是应用统一负责这些策略。

公网 `GET /health` 当前返回 200；`/api/v1/health` 不属于有效路径。现有 liveness 不覆盖数据库/队列 readiness。

### 当前云端观察值与边界

- API、worker active；Caddy/Cloudflare 可达，HTTPS 首页和管理 SPA 返回 200；证书有效期到 2027-01-02。
- SQLite 只读 `quick_check` 为 `ok`；观察时用户表 3 条、订单与支付表均 0 条、代理档案 2 条、邀请码 53 条。这只说明当时数据量，不是业务链路成功率证明。
- 未创建生产用户/订单、未尝试支付、未发送邮件、未核销邀请码；只读检查确认订单与支付记录为 0、支付渠道为 0。API 与 worker 已在备份后发布并重启，当前均 active 且由 `dujiao` 运行；首页、`/health`、代理前台 SPA 与公开商品/分类/配置 API 返回 200。管理 SPA `/admin/` 返回 200。曾误用 `-version` 参数探测线上二进制，之后确认未留额外进程；该探测不作为版本查询方式，无法证明它是否触发幂等启动迁移检查。

## 云端发布后作业总结（2026-10-06，服务器时区）

- 按用户授权清理旧 GPT 日志；未删除 Claude 或模型无法识别的日志。
- 先备份并验证 SQLite 与配置/二进制，再将本地修复源码同步到 `/opt/dujiao-next-src`，在云端构建候选并以原子替换部署；生产配置、数据库、上传和日志目录未被 rsync 覆盖。
- 已部署登录角色授权、邀请码事务与等级查找、CORS、自动交付恢复扫描等修复；API 和 worker 都已重启并以 release 模式、专用低权限账号稳定运行。
- 当前检查通过：`go test ./...`、`go vet ./...`、前后端构建、生产健康与静态路由检查、SQLite `quick_check`、备份校验、恶意 Origin CORS 检查。
- **生产业务仍被支付配置阻塞**：`payment_channels=0`，且未做真实付款/回调/供应商/邮件/退款联调。不能声称购买/代理分账全链路已上线验收。建议先配置至少一个真实支付渠道并用受控小额测试单完成支付、回调幂等、自动/人工履约和代理账务核对；同时配置异地备份与 CSP/点击劫持/Referrer 策略。
- 尚未全量审计独立仓库 `zovo_card_cdk_auto`；本次也没有真实提现打款。不能承诺“无风险”。

## 金额跨页签串价问题复核（2026-10-06）

### 根因

推广代理码原来写入共享的 `localStorage.dj_agent_code` 和 30 天 Cookie `dj_agent`。同一浏览器的其他页签会读取同一 localStorage，API 客户端因此继续发送旧 `X-Agent-Code`；未带代理参数的 URL 也不会清除此值。后端还会在 query/header 缺失时回退读取共享 Cookie。这会让普通主站页签、其他代理链接页签错误继承先前代理的租户和售价。截图中代理链接商品为 138、主站后台售价为 168，本身符合代理可配置售价低于零售价的规则；应修复的是代理上下文跨页签泄漏，而不是要求代理售价等于 168。

### 修复与验证

- 前端代理价格上下文改为 `sessionStorage`，按浏览器页签隔离；每次应用首次加载时，未带代理参数的 URL 会清除该页签可能继承的代理状态，并清理旧版本遗留的共享 localStorage/Cookie。相同页签内继续结账或访问站内其他页面时会保留该次推广归因；新页签加载无代理 URL 则按主站价格请求。通过其他代理 URL 加载时，当前页签会立即替换为新代理码。
- API 客户端仅从 `sessionStorage` 注入代理码；后端不再从 `dj_agent` 或 `agent_code` Cookie 推断价格租户。后端回归测试覆盖携带旧 Cookie 的主站请求仍解析为主站。
- 验证通过：代理租户解析包测试；订单代理价格快照集成测试；前端 62 项测试；用户前端生产构建。
- 该修复已重新构建前端嵌入资源和后端二进制，并在创建新备份后部署，API 与 worker 重启后均 active。生产只读 API 对比两个活动代理身份：未标记主站、代理甲、代理乙均返回各自不同的商品价格；携带旧代理 Cookie 的主站 API 结果与无标记主站完全相同。首页新 JS bundle 已可访问，包含 sessionStorage 逻辑；公网 `/health` 返回 200。后端定价/订单预览集成测试也通过。
- 尚未通过真实图形浏览器手动操作多个页签，也未创建真实订单；建议再用 A=代理甲、B=无代理、C=代理乙三个页签走到下单预览，并比较预览金额与正式创建订单响应。生产无支付渠道，真实收款仍不能验收。
- 本次价格修复发布备份位于 `/var/backups/dujiao-next/pricing-context-20261005T192714Z`；含生产配置、旧二进制、完整 systemd unit/drop-in 与 SQLite 在线备份，备份清单校验通过。

## 兑换桥与 ZovoCard 连通性复核（2026-10-06）

- 云端独立服务 `zovo-cdk.service` 为 active，`127.0.0.1:8082/health` 返回 200。其 SQLite 配置中生产 API host 为 `zovocard.com`、API Key 已配置；没有读取或输出密钥值。
- 使用该服务已有配置只读调用 ZovoCard：`GET /openapi/v1/balance` 与 `GET /openapi/v1/gpt-direct/plans?product=gpt` 均返回 HTTP 200、业务 code 0，套餐 registry 返回 12 项。未创建 CDK、兑换、下单或读取并输出余额。
- 但 Dujiao 生产 SQLite 的 `site_connections` 当前为 0 条。因此 **“兑换桥 → ZovoCard”只读 API 已打通；“Dujiao 采购履约 → 兑换桥/ZovoCard”生产链路未打通**。Dujiao 有 `zovocard` 适配器和 mock HTTP 测试（`go test ./internal/upstream` 通过），但没有生产连接/映射，也没有真实发码和兑换联调。
- 另发现 Dujiao adapter 的真实下单风险：创建请求只传数值 `SKUID`，没有把上游 `SKUCode` 传入适配器；适配器只从表单 `manual_form_data.plan` 读取套餐，否则默认 plus。当前商品表单 schema 仅有 session 字段，因此 Pro 商品有可能误发 Plus CDK（见 `internal/upstream/zovocard.go`）。下单请求即使 `count>1` 也仅处理 `issued[0]`；自动兑换失败时日志会写入完整 CDK；`CancelOrder` 当前为空操作。这些问题解决并测试前不应开启生产采购履约。
- 本轮只读线上探测只证明余额/套餐 API 可鉴权访问；未测试发码、CDK 预览/预检/兑换、异步状态轮询、失败补偿或取消。独立 `zovo_card_cdk_auto` 的完整源码/部署安全审计也不在本次范围。

## 购买链路逐角色说明（2026-10-06）

### 结论

代码有普通商城订单、服务端报价、create-and-pay、支付回调状态机、订单履约派发和代理利润核算路径；但**当前生产不能宣称购买链路已可用或无风险**：生产支付渠道为 0，真实付款/回调没有验收；Dujiao 上游连接也为 0，自动采购交付当前无生产连接。假设仅补齐支付接口，付款部分的预期流程存在，但支付后是否交付取决于商品的履约类型、卡密库存/手工处理准备，或正确的上游连接及 SKU 映射。因此“支付接口完整”本身不足以保证用户最终收到商品。

### 普通用户：主站直接访问

1. 管理员先上架商品/SKU，设置主站售价、库存/履约类型，并配置支付渠道。若商品走上游自动履约，还须配置上游 site connection、SKU 对应关系和可用供应商凭据；若走卡密或人工交付，须确保库存及处理人员可用。
2. 用户直接打开 `https://ai.ooxxoo.cn`（无代理参数），挑商品/SKU/数量，加入购物车或立即购买，进入结账。游客按页面要求填写邮箱、订单密码及验证码；登录用户使用自己的账号。
3. checkout 调用订单预览接口显示服务端核算金额；提交时调用 authenticated 或 guest `create-and-pay`。前端 payload 包含商品、SKU、数量、优惠码和支付选择等，不以浏览器展示价格作为实际收费依据。
4. 服务端创建订单并调用支付创建器；返回支付 URL/二维码等后跳转 `/pay`。用户按支付页面指引完成付款，等待支付平台回调及页面状态刷新。
5. 支付回调检查支付状态、订单状态、支付金额是否覆盖剩余在线应付金额，并处理已关闭/重复/过期支付；通过后事务性标记订单 paid。
6. Worker 按商品履约方式处理交付：自动履约取库存或走已连通上游，人工履约等待管理员操作。用户在订单详情查看/下载交付内容；若仍处理中需等待或联系客服。

### 普通用户：通过代理推广链接

1. 代理账号先获准启用，设置允许销售的商品/SKU 和售价，确保价格符合底价及限价约束，并取得已经启用的推广链接/店铺域名。
2. 用户点击该链接访问商城；链接中的代理标识建立当前页签的代理租户价格上下文。无代理标识的另一个新页签应恢复主站价；不同代理链接应重新加载对应代理价格。
3. 用户选品、预览并结账，流程与主站用户相同。报价和创建订单时服务端根据当前代理上下文重算并保存订单价格/代理利润快照，不能只依赖页面金额。
4. 用户完成支付；回调通过金额及订单状态校验后标记 paid，并在同一事务生成代理待确认利润记录。确认期后利润进入可用余额；退款则按账务流程冲减。商品履约仍依赖前述库存、人工流程或已配置上游。

### 代理本人：在“自己的代理价”购买

代理中心 `/reseller` 提供店铺/价格配置、销售订单、余额/流水和提现管理；代码中没有独立的“代理按拿货底价买货”结账流程。代理若打开自己的有效推广链接或店铺域名，像顾客一样从商城下单，则适用其公开销售价，并走同一 create-and-pay 与履约流程。这会被当作该代理店铺订单，系统可能记录“售价 - 成本”的代理利润；若业务真正需要代理按成本价批发自购，需要单独的采购订单/余额扣款/库存入账业务路径，当前不能把代理中心当作已有此能力。

代理本人操作顺序：登录代理中心 → 配好商品售价/上架范围 → 打开“预览店铺”或自己的推广链接 → 作为买家选品并结账 → 选择已配置支付方式并付款 → 在商城订单页查看交付；回到代理中心的销售订单/账务页查看归属订单和利润状态。若其希望按拿货价自购，则目前无对应操作入口。

### 尚未验证和主要阻塞

- 生产 `payment_channels=0`，支付 UI/真实支付渠道、金额一致性、回调签名/幂等和退款闭环未做真实小额交易验收；代码路径通过测试不能替代实收款验证。
- 生产 `site_connections=0`；兑换桥只读接口连通不代表 Dujiao 已能发起上游采购。ZovoCard adapter 另有套餐识别、批量 CDK 处理、敏感码日志与取消语义缺陷（见上节）。自动履约不应上线启用。
- 没有生产订单/支付数据，故三类路径均无真实购买样本；没有浏览器端到端对支付、库存锁定、并发售罄、邮件送达、异常补偿、退款和账务最终余额做验收。
- 对外承诺应区分“订单可以创建”“款项已核验”“商品已交付”“代理利润可提现”四种状态，不能将支付成功直接说成购买完成。当前代理利润先进入待确认状态，非支付成功即刻可提现。

本次仅做代码与既有生产状态审阅并更新交接文档；没有创建订单、支付、退款、发码、兑换或修改线上业务数据。

## ZovoCard 线上只读复核与进货路径补充

- 最新只读检查时 `zovo-cdk.service`、`dujiao-next.service` 均为 active，兑换桥本机 `/health` 返回正常。使用桥服务数据库中已有 host/API key 在服务器进程内调用 ZovoCard，只打印 HTTP 状态和业务 code、不打印 key 或余额数值：`GET /openapi/v1/balance`、`GET /openapi/v1/gpt-direct/plans?product=gpt` 都是 HTTP 200、业务 code 0。匿名公网请求 plans 返回 401，符合鉴权要求。此结果只证明当前保存的 API 凭据可访问查询接口，未覆盖任何写操作。
- Dujiao 生产 SQLite 只读计数：`site_connections=0`、`sku_mappings=0`、`procurement_orders=0`。因此 Dujiao 购买履约端到 ZovoCard 的实际自动采购链路尚未配置/打通。不要将独立兑换桥健康或它对 ZovoCard 的查询成功等同于 Dujiao 自动进货完成。
- **人工进货**：ZovoCard 帮助文档描述的 CDK issuer 流程为登录符合资格的账号，准备资金，进入 CDK 管理，选择产品/套餐和数量、核对费用，创建 CDK 并安全保管；然后按本项目人工库存/人工交付流程录入和发放。具体入口、金额和费用以登录后账号实际页面为准。本项目当前尚未核验手工导入/库存全链路。
- **自动采购**：Dujiao 管理后台先添加 ZovoCard upstream site connection（API URL/Key、回调/重试），为本地商品配置 product mapping 和本地 SKU 到上游 SKU 的映射，并将商品设为 upstream 履约。用户付款成功后，系统创建采购单、Worker 调上游发码/兑换、回调或轮询更新采购状态和交付。当前连接与映射计数皆为零；并且 Dujiao adapter 套餐识别、批量处理、CDK 日志泄露、取消语义缺陷未修复，因此不应开启自动采购。
- 官方来源：ZovoCard [API 集成帮助](https://help.zovocard.com/en/help/api/)要求使用授权 API key、配置权限/限额并依赖幂等；[CDK 创建/兑换帮助](https://help.zovocard.com/en/help/cdk/)说明签发与兑换是不同操作并给出人工发码步骤；[CDK 页面](https://zovocard.com/home)当前条款声明中国大陆居民/所在地人士不得注册或使用，且禁止以技术手段绕过地域限制。此资格限制必须在任何实际采购前确认，不得建议规避。
- 本轮未发起任何写请求，没有采购、发码、兑换或更改生产业务数据。

## USDT 人工收款链路（本地代码复核，未上线）

### 设计

- 收款渠道使用固定钱包地址而非 BEpusdt/EPusdt 商户网关。用户确认由管理员人工核账，因此提交交易 hash 仅设为待审核，不会直接把订单改成已付款。
- 按用户方向使用 Binance P2P `USDT/CNY`、`SELL`、`ALIPAY` 公告报价（用户卖 USDT，收到 CNY）。服务端筛选广告限额覆盖订单 CNY 金额、卖家剩余 USDT 足够、支付方式包含 Alipay 的在线商家，选可用最高报价；不是 spot `USDT/CNY` 交易对。当前官方 Binance Skills Hub 文档列出无需 API Key 的公开广告查询。**公开广告并不锁价**；创建支付时快照后，网站固定该单需转的 USDT 数量。
- 数量按 `CNY 金额 / 报价(CNY per USDT)` 计算并向上舍入 2 位小数，ProviderPayload 保存该单实际有效 USDT/CNY rate，使现有 callback 覆盖金额检查可回算原订单 CNY 金额。
- 访问失败、异常 JSON、无匹配广告以及非 CNY 订单均拒绝创建支付，禁止静默退回固定汇率。

### 防重和审核

- 人工 proof API 仅允许所属用户/提供 Guest Authorization 凭证的游客提交 64 位 hash；提交成功仍不改订单支付状态。
- 新增 `manual_crypto_tx_claims` 唯一 hash 索引，限制 hash 重复用于多个 payment。拒绝证明后保留 hash claim；同一个交易不允许再次使用。
- Finance 管理员 POST 审核路由由后台授权保护。后台对 `manual_crypto` 仅下发允许展示的 payload 字段，避免高级 JSON 任意配置字段泄露。审核通过后经原 callback 事务处理支付和履约；用户不能自助确认到账。
- 仅校验 EVM 地址和 TRON 地址的基本格式，不验证收款钱包的链上拥有权或 Binance deposit 支持状态。用户 BEP20 与 ERC20 地址相同，属于地址格式上有效的 EVM 地址；上线前仍需核对 Binance 每条网络的 USDT 充值页面。
- 管理员需核实真实链上交易成功、正确网络和接收地址、到账不小于本单金额、以及足够确认数。当前没有自动链上 RPC 验证，也没有构造交易证明；不能将 UI/编译通过说成已实收款验收。

### 验证和上线状态

- `go build ./cmd/server`、用户端 `pnpm run build`、管理端 `pnpm run build:fullstack` 与 `git diff --check` 通过；本轮按用户指令没有运行测试。
- Binance 公开广告查询做过只读调用并确认接口当前返回限额、余额、支付方式、商家类型和价格字段。
- 以上是本地代码变更：当前没有发布这些变更，生产 `payment_channels` 仍按此前只读检查为 0。还没有实际配置三个链渠道、创建订单、提交哈希、人工核账、处理真实支付或履约。因此当前生产尚不能通过 USDT 收款；后续部署与支付渠道启用后须小额受控端到端验收。

### 自动采购适配器修复与发布记录

- Plan ID 由 registry 的排列序号派生，原请求又没有携带可恢复 plan 的 SKU 数据，导致计划顺序变化可能错购；未知 ID 原先静默默认为 Plus。现在采用 plan key 的稳定 ID，并在任何发码写请求前重新读取 registry 解析 plan，未知 SKU 失败关闭。
- 原适配器 `count > 1` 只处理首个 CDK。现在收集并交付所有完整 `code`，汇总实际服务费，通过采购 application 的即时交付回调完成本地订单。若用户提交 session 直充凭据，一次限制 quantity=1；自动兑换失败时仍回退交付 CDK，但日志和异常不包含完整 CDK 或上游响应 body。供应商取消 API 未实现时显式返回 unsupported，不再假报取消成功。移除了此前硬编码的 `payment_country=PH`，避免伪报地域。
- 新增/更新测试覆盖正确 plan 和多 CDK 交付、未知 SKU 不触发发码、即刻交付状态推进。`go test ./internal/upstream ./internal/modules/procurement/...` 通过，`git diff --check` 通过。
- 修复二进制已部署云端；发布前备份为 `/var/backups/dujiao-next/zovo-adapter-20261005T222554Z`。服务 active，本机 health 和公网首页返回 200。部署后只读确认线上 `site_connections=0`、`sku_mappings=0`、`procurement_orders=0`、`payment_channels=0`；本次没有调用 CDK 创建/兑换写 API。
- **采购尚未端到端跑通**：部署只修好实现，不会创建生产连接或映射。待账户服务条款资格确认后，需在 Dujiao 的「站点连接」添加 `zovocard` 连接（Base URL `https://zovocard.com`、具有 CDK issue/redeem 权限及合理限额的 API key），在商品映射导入并映射每个 SKU；验证对应商品成本/售价币种、履约类型 `upstream` 和支付渠道后，先用一笔用户授权的小额真实采购检查供应商扣费、上游结果、完整交付 payload、订单状态和异常补偿。此单会产生真实供应商写操作/费用，不能由本次只读核查代替。
- ZovoCard [API 接入帮助](https://help.zovocard.com/en/help/api/)要求 API key scope、IP/操作限额和幂等调用；[CDK 帮助](https://help.zovocard.com/en/help/cdk/)把签发与兑换分成不同操作；[当前服务条款](https://zovocard.com/home)有地域/用户资格限制。需要先确认账户资格；不可绕过限制。

### ZovoCard 连接可配置性补充

生产已部署 ZovoCard 可不填 API Secret 的兼容改动，因为协议 adapter 只消费 API Key；部署前备份生产数据库/配置/二进制，SQLite 完整性通过，部署后 API 健康。此次只部署代码，没有创建连接或任何上游映射。生产有 4 个既有商品及各 1 个 SKU；现有售价和成本未改，履约方式保持 `manual`。线上仍是 `site_connections=0`、`product_mappings=0`、`sku_mappings=0`、`procurement_orders=0`。

不能把计划 registry 的 `$0.15 service_fee_usd_minor` 当成完整采购价：这不足以重算当前 CNY 商品成本/售价。自动同步价格必须保持关闭，待供应商可核对的实际采购报价/成本语义明确后再设置映射成本。ZovoCard 官方 API 指南要求本人获批 API Key，并建议先做只读权限验证、配置限额和幂等；地域资格仍须合规确认。真实发码、采购、兑换和验收均未执行，后续须等待用户明确指令。

### ZovoCard 生产连接和 SKU 映射完成状态

用户已明确确认合法合作伙伴账户和 API 权限。现已用兑换桥存储的 API Key 建立 Dujiao 生产连接 `ZovoCard production`（ID 1），在服务器进程内执行只读余额/plans 验证，管理后台 Ping 后连接显示已激活；API Key 未输出。协议无须 API Secret。4 个本地商品/default SKU 已分别映射 `plus / pro_5x / pro_20x / pro_50x`，稳定上游 ID 为 `1152239221 / 1101237304 / 470958661 / 331537736`。后台当前显示 4 条已启用映射。

安全边界：本地商品售价/成本未动，全部履约仍为 `manual` 且 `is_mapped=false`，避免后台编辑商品时自动切换为 `upstream`；`auto_sync_price=false`，因此只完成可复用的连接和 SKU 绑定，没有因本次配置激活自动采购。plans registry 的 `$0.15` 字段只代表 service fee 元数据，不足以核对完整采购成本；映射库存显示的 `9999` 是 adapter 合成占位数，不是上游保证的真实库存。须在采购验收前核实实际扣费、余额和数量限制。线上 procurement orders 仍为 0，payment channels 仍为 0；没有调用 CDK 创建、兑换、采购写 API。真实采购验收必须等待用户明确指令。

连接 PING 的只读请求在浏览器后台通过，云端服务健康。adapter 还修复了 provider `FulfillmentType` 分类：ZovoCard 自动发码计划作为 provider `auto` 上报，避免映射导入层因把 `upstream` 误判为非 `auto` 而规范化成 `manual`。本地对应 upstream、mapping、siteconnection、procurement 测试通过，管理端类型检查和 Vite build 通过。上线前发现旧版连接 API 会通过 JSON 序列化返回 API Key；现已把 `ApiKey` 从 JSON 排除，管理端编辑时不再填充凭据，留空更新保持既有 Key。上线后已检查连接列表、编辑表单 Key 空白和 Ping 成功。由于旧版本曾向登录后的管理 API 客户端返回 Key，若存在非受信任的管理账号/终端，建议在 ZovoCard 管理端轮换该 Key，再同步更新 Dujiao 和兑换桥；不要在聊天中粘贴 Key。

### 2026-10-06 管理后台静态资源 MIME 错误复核

用户提供的 DevTools 截图显示后台模块脚本与样式请求收到 `text/html`，导致浏览器 MIME 检查拒绝加载。复现路径与根因：生产后台入口来自全栈嵌入资源，但实际内嵌 `index.html` 是普通 `vite build` 产物，使用 `/assets/...` 绝对根路径；同源 fullstack 路由中这并非 `/admin` 的静态目录，落入用户 SPA fallback 后返回 HTML。`frontend/admin/vite.config.ts` 中已有 `VITE_FULLSTACK=1` base 与 admin placeholder 插件，需调用 `pnpm run build:fullstack`。

已按 fullstack 构建重建管理端、同步云端内嵌资源并发布。备份在 `/var/backups/dujiao-next/admin-static-base-20261006T000217Z/`。云端 loopback 验证后台深链接返回 base `/admin/`，JS 路径 `/admin/assets/...` 返回 `text/javascript`；服务 active。浏览器强制刷新后的 UI 交互以及公网 CDN 请求仍待用户浏览器验收；当前执行环境对公网域名请求返回 403，无法据此断定 CDN 状态。没有动数据库或支付设置。
