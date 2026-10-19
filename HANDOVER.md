# Dujiao-Next 代理分销与商城系统重构交接文档

> **文档版本**：v1.0.0  
> **编写日期**：2026-10-05  
> **交接目标**：供后续工程团队及 Codex 对本轮系统重构与业务改造进行全盘检查、测试审计与持续迭代。

---

## 1. 项目基本信息与运行环境

### 1.1 系统架构与技术栈
- **后端架构**：Go 1.24+、Gin Web 框架、GORM ORM、SQLite 3（WAL 模式）、Redis、Asynq 队列系统。
- **前端架构**：Vue 3、Vite、TypeScript、Pinia 状态管理、Tailwind CSS、Lucide Vue Next 图标库。
- **邮件服务**：Brevo SMTP 网关（`smtp-relay.brevo.com:587`，发件人 `service@ooxxoo.cn`）。
- **网络与反向代理**：Caddy 2 + Xray (TLS Reality/Fallback) 反向代理至本地 `127.0.0.1:8080`。

### 1.2 部署环境与目录规范
- **生产服务器 IP**：`66.94.122.149` (Ubuntu x86_64)
- **SSH 连接方式**：`ssh -i ~/.ssh/node1_vnc_ed25519 root@66.94.122.149`
- **云端源码目录**：`/opt/dujiao-next-src`
- **云端运行目录**：`/opt/dujiao-next`
- **数据库路径**：`/opt/dujiao-next/db/dujiao.db`
- **Systemd 进程服务**：`dujiao-next.service`（主二进制：`/opt/dujiao-next/dujiao-next -mode api`）
- **主要域名**：
  - 官网与商城前台：`https://ai.ooxxoo.cn`
  - 管理后台入口：`https://ai.ooxxoo.cn/admin`

---

## 2. 本轮核心业务需求与实现汇总

### 2.1 代理邀请码与准入体系
1. **管理后台邀请码模块**：
   - 路径：`frontend/admin/src/views/admin/ResellerInviteCodes.vue`
   - 功能：支持单张或批量生成 16 位大写邀请码（格式：`XXXX-XXXX-XXXX-XXXX`），可配置最大使用次数、有效期、绑定备注。
   - 修复：修复了 RBAC 策略注册，确保站长与管理员拥有完整的生成、停用与删除权限。
2. **代理注册开通流程（极简规范化）**：
   - 路径：`frontend/user/src/views/AgentCenter.vue`
   - **必填项**：专属邀请码、账号（必须为邮箱）、密码、二次确认密码。
   - **选填项**：QQ、Telegram 联系方式。
   - **开通机制**：代理注册需提供有效邀请码并明确同意条款；邮箱验证码豁免仅适用于该代理注册流程。服务端通过稳定的 `agent` 等级 slug 查找会员等级 ID，在同一事务内创建代理用户并核销邀请码；登录请求中的 `role` 不作为授权依据。

### 2.2 代理价格体系与分销利润机制
1. **价格分层**：
   - **主站零售指导价**：如 Plus 168.00 元、Pro 5x 899.00 元、Pro 20x 1688.00 元、Pro 50x 3899.00 元。
   - **代理拿货底价**：如 Plus 底价 114.00 元、Pro 5x 底价 660.00 元、Pro 20x 底价 1140.00 元、Pro 50x 底价 3520.00 元。
   - **代理自主销售价**：代理在控制台自行设定的对外售价（如 Plus 设定为 135.00 元）。
2. **定价校验规则修复**：
   - 修正了原本必须 $\ge$ 零售指导价的限制，调整为 **底价 $\le$ 自主售价 $\le$ 允许上限**。
3. **分销佣金与结算**：
   - 客户通过代理推广链接下单时，订单金额即为代理设定的自主销售价；
   - 代理获得的差价利润自动计算为：$\text{代理利润} = \text{自主销售价} - \text{代理拿货底价}$（例如：$135 - 114 = 21$ 元）。

### 2.3 前台报价展示与推广闭环
1. **商城首页价格展示**：
   - 无论是游客访问还是代理账号登入后，在首页卡片和弹窗中统一展示站点的**标准零售指导价**（168 / 899 / 1688 / 3899 元），不再直接展示代理拿货底价，移除“已享代理特惠价”字样。
2. **专属轻量化推广链接（`https://ai.ooxxoo.cn/?agent=XXXXXX`）**：
   - 客户通过专属链接访问时，系统后端自动解析租户并下发代理所配置的自定义销售价；
   - 支持在「销售链接与定价」中一键复制链接、生成带专属推广码的高清二维码弹窗，并提供**保存图片**与**转发分享**功能。

### 2.4 代理密码找回功能
- 在代理登录卡片中增加了「忘记密码？」入口；
- 采用全站统一的 **Brevo SMTP 邮件 6 位验证码** 找回机制（`POST /api/v1/auth/verify-code` + `POST /api/v1/auth/forgot-password`），校验成功后即刻重置密码并返回登录。

---

## 3. 核心源码文件改动索引

### 3.1 后端核心文件 (Golang)

| 文件路径 | 核心职责与改动说明 |
| :--- | :--- |
| `internal/modules/reseller/application/domain_resolver.go` | **核心路由与租户解析器**：修复了白名单机制，将 `/api/v1/reseller`、`/api/v1/user`、`/api/v1/auth` 纳入豁免，防止代理后台操作被判定为分销子站租户导致 403。 |
| `internal/modules/reseller/application/pricing_rules.go` | **定价规则校验器**：修正 `ValidateUnitAmount` 逻辑，允许代理设定低于主站零售价、高于拿货底价的销售价。 |
| `internal/modules/reseller/application/product_setting.go` | **分销商品配置用例**：实现代理自定义售价的保存、预览与重置，以及底价上限校验。 |
| `internal/modules/order/application/reseller_pricing.go` | **展示价与利润计算**：处理分销站/推广链接的商品展示价覆盖，以及结算时的利润差价计算。 |
| `internal/modules/identity/userauth/transport/http/user_login_handler.go` | **注册与登录端点**：支持代理使用邮箱一键凭邀请码开通，自动绑定代理身份与生成专属代理码。 |
| `internal/modules/reseller/infrastructure/gormstore/profile.go` | **代理 Profile 仓储**：实现了 6 位随机大写字母+数字的 AgentCode 生成策略，替代旧的流水号。 |
| `internal/authz/bootstrap.go` | **RBAC 鉴权策略**：补充了代理邀请码管理相关 API 的角色权限策略。 |
| `internal/upstream/zovocard.go` | **上游 ZovoCard 供应商适配器**：对接外部卡密与充值接口。 |

### 3.2 前端核心文件 (Vue 3 / TypeScript)

| 文件路径 | 核心职责与改动说明 |
| :--- | :--- |
| `frontend/user/src/views/AgentCenter.vue` | **代理中心综合页面**：<br>1. 整合开通账号、代理登录、找回密码三态表单；<br>2. 代理拿货、商品自主定价、订单与提现结算；<br>3. 推广二维码生成、下载与分享弹窗。 |
| `frontend/user/src/views/Home.vue` | **商城首页**：移除针对登录代理的前端强行改价逻辑，展示建议零售价，适配推广参数。 |
| `frontend/user/src/components/ProductQuickBuy.vue` | **极速下单购买弹窗**：清理冗余代理标识，统一价格呈现机制。 |
| `frontend/admin/src/views/admin/ResellerInviteCodes.vue` | **管理后台邀请码管理**：邀请码列表、批量生成、状态切换与使用记录。 |
| `frontend/user/src/api/reseller.ts` | **代理前台 API 客户端**：封装代理资料、商品定价、财务报表与提现等接口。 |

---

## 4. 关键排查与踩坑记录 (Troubleshooting Log)

### 踩坑 1：修改商品售价时报错“无权限访问” (403 Forbidden)
- **现象**：代理在 `/agent` 保存商品自定义售价时，前端弹窗提示 `ai.ooxxoo.cn 显示: 无权限访问`。
- **原因**：当代理访问过 `?agent=XXX` 链接后，浏览器设置了 `dj_agent` Cookie；而后端 `routes_storefront.go` 中的代理管理路由挂载了 `RequireMainTenantForResellerConsole` 中间件（要求必须在主站租户上下文操作）；而 `domain_resolver.go` 的路径白名单中仅包含 `/api/v1/user/reseller`，但前端实际请求路径为 `/api/v1/reseller/product-settings/*`，导致该请求被识别为“分销子站租户”，从而被中间件直接拦截返回 403。
- **解决方案**：在 `domain_resolver.go` 白名单中补充 `strings.HasPrefix(req.URL.Path, "/api/v1/reseller")` 等接口前缀，彻底解决拦截问题。

### 踩坑 2：自主售价保存时报错低于基础价
- **现象**：代理尝试将售价从 168 降至 135 时，后端拒绝保存。
- **原因**：历史代码中将 `BasePriceAmount`（零售指导价 168 元）误作为下限下限校验，导致无法设置低于主站零售价的优惠售价。
- **解决方案**：在 `pricing_rules.go` 中重构下限判断，将拿货底价 `CostPriceAmount`（114 元）作为硬性下限。

---

## 5. 常用编译、构建与部署指令

### 5.1 本地编译与构建

```bash
# 1. 前端用户端打包构建
npm --prefix frontend/user run build

# 2. 前端管理端打包构建 (如有改动)
npm --prefix frontend/admin run build

# 3. 后端自动化单元测试与架构规约测试
go test ./...
```

### 5.2 生产发布安全门（旧整目录发布方式已禁用）

禁止把本地仓库根目录 rsync 到生产源码目录，也禁止直接在脏的 `/opt/dujiao-next-src` 上编译发布。工作区可能包含未提交功能、旧于生产的页面或临时文件，这会把整站前后端一起替换。

生产 UI 基线与校验值见 `ops/production-ui-baseline-20261012/README.md`（清单于 2026-10-19 按线上当前运行版本刷新）。线上版本是本地、GitHub 与云服务器源码的共同 UI 基准；线上更新后须同步刷新清单。发布必须从干净隔离工作树构建，并先运行 `scripts/assert-production-release-safe.sh`。它校验 UI 源码和资源清单、HEAD 祖先锚点、工作树状态及线上二进制哈希。只有用户明确授权 UI 改动，才可设置 `APPROVE_UI_CHANGE=YES`。

每次发布先制作并验证候选二进制，再原子替换；发布前备份生产二进制、配置、数据库和旧资源；新资源增量添加并保留旧哈希文件；发布后检查公网 `/health`、首页和涉及的前端资源。校验失败立即停止，回退到上一版二进制与资源。

生产配置、数据库、上传内容只保留在 `/opt/dujiao-next`，禁止从源码工作树覆盖。当前共享本地工作区含有未提交改动，不得 reset、clean 或 stash 来绕过安全门。
## 7. 生产发布记录与接手事项（2026-10-06，服务器时区）

本次已在用户授权下完成 GPT 旧日志清理和线上修复发布。**仅删除了可从采样识别为 GPT 的 19 个旧日志文件（约 9.27 GB）；Claude 日志和无法识别模型的日志均保留。**根盘使用率由约 94% 降至约 90%，可用空间约 21–22 GB。

发布前创建并校验了备份目录 `/var/backups/dujiao-next/maintenance-20261005T190124Z`，包含配置、旧二进制、API/worker systemd 单元和 SQLite 在线备份；目录/文件权限为 0700/0600，SHA-256 清单及数据库完整性检查通过。此备份仍在同一块云盘，尚未形成异地备份。

本地代码与内嵌前端资源已同步到源码目录，并在云端构建、原子替换运行二进制。认证/代理身份、邀请码事务、CORS 默认策略和支付后履约恢复扫描修复均已上线。API 与 worker 运行账号为 `dujiao`，`server.mode` 为 `release`，systemd 已启用只读系统目录与受限写路径等隔离；公网 `GET /health`、主页、管理页、代理 SPA、公开商品/分类/配置 API 均返回成功。恶意 Origin 预检不再收到 `Access-Control-Allow-Origin`。

**业务验收阻塞项：**生产 `payment_channels` 表当前为 0 条，生产订单/支付记录也为 0。因此真实用户尚不能完成已验证的支付购买闭环；不能以代码测试或页面可打开代替支付、回调、交付和代理利润结算验收。接手时应先在后台配置至少一个真实支付渠道，再用受控小额订单核验金额签名/回调幂等、自动与人工交付、退款与代理账务；还需配置异地备份、邮件送达和供应商真实凭据联调。公网当前未观察到 CSP、`X-Frame-Options`、`Referrer-Policy`，应明确边缘或应用的安全响应头策略。独立仓库 `zovo_card_cdk_auto` 尚未全量审计。

详细发现、验证与边界见同目录 `AUDIT_REPORT.md`。本次生产变更没有创建用户、订单、支付或核销邀请码，也没有发送邮件或执行真实供应商履约。

### 7.1 金额串价修复（2026-10-06）

发现推广代理码此前存放在跨页签共享的 `localStorage` 与 Cookie，导致同一浏览器其他页签沿用旧代理售价。最新代码将价格上下文改为按页签隔离的 `sessionStorage`；应用首次加载无代理参数的链接时重置代理上下文，并清理旧版遗留状态；后端忽略共享代理 Cookie。相同页签里的推广访问和结账仍保留当前代理归因。修复影响 `frontend/user/src/utils/affiliate.ts`、`frontend/user/src/api/client.ts` 和 `internal/modules/reseller/application/domain_resolver.go`。

验证记录：`go test ./internal/modules/reseller/application`、`go test ./internal/modules/order/integrationtest/application`、用户端 62 项测试、`npm run build` 均通过。修复已在新备份后部署到云端，API 和 worker 重启后均 active。线上只读 API 验证未标记主站、两个活动代理返回不同配置价格，旧代理 Cookie 不会改变主站价格；新前端 bundle 包含 sessionStorage 逻辑。真实浏览器多页签交互仍建议人工验收；真实收款由于生产尚未配置支付渠道仍无法验收。

本次发布备份：`/var/backups/dujiao-next/pricing-context-20261005T192714Z`。包含上一版运行二进制、生产配置、API/worker unit 与 drop-in、在线数据库快照及通过 SHA-256 校验的清单。

### 7.2 兑换桥与 ZovoCard 连通状态（2026-10-06）

线上 `zovo-cdk.service` active，桥服务健康检查返回 200；已配置生产 ZovoCard host 和 API Key。使用该配置进行的只读 `GET /balance`、`GET /gpt-direct/plans?product=gpt` 均 HTTP 200 且业务 code 为 0。没有发 CDK 或执行兑换。

**区分两段链路：桥服务到 ZovoCard 的只读鉴权调用已通；Dujiao 采购履约到上游目前未接通。**Dujiao 生产库 `site_connections` 为 0 条；适配器虽存在且 mock 测试通过，但真实订单未联调。接入前先修复：上游套餐依赖 `manual_form_data.plan`，缺失时默认 Plus，Dujiao 请求没有传 SKUCode；数量大于 1 时只处理第一张 CDK；自动兑换失败日志包含完整 CDK；取消方法为空操作。详见 `AUDIT_REPORT.md` 连通性复核节。不要在修复 SKU/plan、批量交付、凭据日志和取消语义前启用生产采购连接。

### 7.3 三类购买路径代码审阅（2026-10-06）

审阅结论：普通主站用户和推广链接访客共用商城 checkout；服务端报价和创建订单时重算金额，不信任前端单价。支付成功回调在校验金额覆盖订单在线应付额、订单仍可支付后，以事务标记订单 paid 并写入代理利润待确认账本；确认期届满才转为可用余额。支付创建后前端进入 `/pay` 等待支付/显示结果。

代理中心 `/reseller` 用于管理销售价格、店铺、订单和账务，没有独立的“代理按拿货价采购”checkout。代理本人可以打开自己的有效店铺/推广链接，按自己配置的零售价走商城购买；该操作仍属于其店铺销售订单，可能会给本人生成利润记录，不等于批发采购。

现状不能回答为“已具备可商用的完整购买”。生产无支付渠道（`payment_channels=0`），收款闭环未验收；生产 Dujiao `site_connections=0`，自动上游履约链路未连接。配置支付后，只有已备妥库存、映射/交付方式的商品才可能完成；真实回调、交付、退款和代理结算仍需受控订单端到端验收。详细用户操作步骤及各角色前置条件见 `AUDIT_REPORT.md` 中“购买链路逐角色说明”。

### 7.4 ZovoCard 当前连通复核与采购说明

本次再次只读检查云端：`zovo-cdk.service` 与 `dujiao-next.service` 均 active；兑换桥 `/health` 正常。通过兑换桥已配置的凭据在服务器内存中执行 ZovoCard 只读 API：`/openapi/v1/balance` 和 `/openapi/v1/gpt-direct/plans?product=gpt` 均 HTTP 200、业务 code 0；未输出余额/凭据，也未发码、下单或兑换。无认证的公网 plans 请求返回 401，符合接口认证要求。**这证明兑换桥配置可鉴权查询 ZovoCard，不证明 Dujiao 采购集成可用。**Dujiao 生产 `site_connections`、`sku_mappings`、`procurement_orders` 当前均为 0。

业务有两种不同的进货方式：在 ZovoCard 控制台为符合其服务条款的账户入金后，进入 CDK 管理选择产品/套餐与数量、核对费用并生成 CDK，再安全保管/人工导入交付；或者启用 Dujiao 自动采购——后台创建 ZovoCard 连接、商品映射和每个本地 SKU 的上游 SKU 映射，商品履约类型设为 upstream，客户订单支付后 Worker 自动调 ZovoCard 发 CDK/执行兑换并同步结果。后一种目前不能启用：Dujiao 侧无任何连接/映射，且 `internal/upstream/zovocard.go` 有套餐字段未从 SKU 传递、数量大于 1 只处理首张 CDK、失败日志泄露完整 CDK、取消为空操作等上线阻断缺陷。ZovoCard 当前条款页面另明确限制中国大陆居民/所在地用户注册和使用，并禁止规避地域限制；运营前应自行确认账户和业务资格。

### 7.5 自动采购适配器加固发布

修复并部署了 ZovoCard 自动采购适配器：上游 plan ID 改为按 plan key 稳定生成；采购前用当前上游 registry 解析 plan，未知 SKU 失败关闭，不再默认发 Plus；多数量发码把所有完整 CDK 作为交付内容并汇总供应商费用；CDK 即时交付通过采购状态回调直接落单；自动兑换错误响应改为不回显敏感 body，失败日志不再包含 CDK；取消不支持时明确报错而非返回假成功；移除固定 `payment_country=PH`，不虚报用户所在地区。增加了 plan 解析、多 CDK 交付、未知 SKU 拒绝和即时完成落单测试。

验证：本地 `go test ./internal/upstream ./internal/modules/procurement/...` 全通过，`git diff --check` 通过；云端 release/fullstack 二进制已构建原子替换，API/worker 与桥服务均 active，`/health` 和公网首页 HTTP 200。发布前数据库完整性及 SHA-256 备份位于 `/var/backups/dujiao-next/zovo-adapter-20261005T222554Z`。**本次没有配置生产上游连接或 SKU 映射，线上 `site_connections=0`、`sku_mappings=0`、`procurement_orders=0`，且 `payment_channels=0`；没有触发任何采购写接口。**仍需账户地域资格确认、添加合法使用的 API connection/映射，以及受控的一笔真实采购验收后才能宣称自动进货跑通。

---

## 6. 建议后续持续关注事项
1. **Brevo SMTP 发送配额**：关注找回密码及订单邮件通知的日发送限额及投递率。
2. **代理提现审核流程**：代理在「提现结算」提交申请后，需在管理后台进行打款核销；可在后续迭代中增加 Telegram Bot 提现申请即时推送。
3. **上游供应商接口状态监控**：关注 `internal/upstream/zovocard.go` 针对订单自动发货的履约回调日志。

### 7.6 ZovoCard 连接配置准备（2026-10-06）

为 ZovoCard 协议补充了“API Secret 可为空”的后端连接创建/读取处理：该协议 adapter 仅使用 API Key，要求虚构一个 Secret 会阻断合法配置。修改 `internal/modules/siteconnection/application/service.go` 已单文件同步到云端源码目录，云端使用 `/usr/local/go/bin/go` 构建并原子替换 API 二进制；API 服务 active，`/health` 返回 `{"status":"ok"}`。发布前备份 `/var/backups/dujiao-next/siteconnection-20261005T230702Z` 的 SQLite `PRAGMA integrity_check` 为 `ok`。

生产只读复核：商品与 SKU 各 4 个；本地售价/成本分别为 Plus 168/114、Pro 5x 899/660、Pro 20x 1688/1140、Pro 50x 3899/3260 CNY；它们当前均为 `manual` 履约。`site_connections=0`、`product_mappings=0`、`sku_mappings=0`、`procurement_orders=0`，所以本次部署没有创建真实连接或映射，也没有触发采购写 API。仍待账户/API 使用资格确认后，才可将 API Key 写入 Dujiao 生产连接并绑定映射。

价格限制：当前 ZovoCard plans registry 暴露的是 `service_fee_usd_minor`（本次观察为 15，即标称 $0.15 服务费），不是可确认的商品完整采购价。不可据此覆盖本地成本 114/660/1140/3260 CNY，也不能启用自动价格同步或据此承诺毛利。最终映射应保留本地售价和成本；上游成本需从有权限的实际订单报价/供应商规则核实。API 帮助文档要求使用本人获批的 API Key，并建议先以只读接口验证、配置权限/限额和幂等；真实采购验收仍必须等待用户明确指令。

### 7.7 ZovoCard 生产连接与商品 SKU 映射完成（2026-10-06）

用户确认其为 ZovoCard 合作伙伴，具备合法账户/API 权限，实际操作符合服务条款。通过兑换桥服务器内 SQLite 读取既有 API Key，在服务器进程内向 `https://zovocard.com` 执行只读余额与 plans 查询（HTTP 200、业务 code 0），随后把 Key 写入 Dujiao 生产 `site_connections`；未打印或回显凭据。连接 #1 `ZovoCard production` 已在管理后台执行 Ping 并显示“已激活”。ZovoCard 连接 API Secret 留空（协议不使用 Secret），重试策略 3 次、间隔 30/60/120 秒；加价 0、价格自动同步关闭。

已为 4 个既有商品和默认 SKU 建立启用的 product/SKU 映射：`chatgpt-plus-1m → plus (1152239221)`、`chatgpt-pro-5x-1m → pro_5x (1101237304)`、`chatgpt-pro-20x-1m → pro_20x (470958661)`、`chatgpt-pro-50x-1m → pro_50x (331537736)`。上游 plans API 当前报告各计划 `service_fee_usd_minor=15`，映射字段记为 $0.15 USD；这只是服务费元数据，不是已核实的完整供应商采购成本。plans API 不提供数字库存；adapter 现以 -1 表示按需签发/不追踪固定库存，不能视为供应商保证无限可发；采购前须确认资金余额、计划实时可购买性和实际费用。

为避免这次配置启动自动采购，商品 `fulfillment_type` 全部仍为 `manual` 且 `is_mapped=false`；这避免编辑商品时映射逻辑强制将履约改成 `upstream`。product/SKU 映射记录保留启用状态供后续验收，当前不会触发采购。本地 CNY 售价与成本（168/114、899/660、1688/1140、3899/3260）没有改变，连接价格自动同步保持关闭。生产后台映射列表显示 4 条，连接列表显示 active 且 Ping 时间更新；API `/health` 正常；生产采购订单仍 0。当前生产支付渠道仍 0。**连接与映射已配置，但自动履约/真实购买尚未启用或验收；不得把 Ping/只读 plans 成功写成采购链路验收通过。等用户另行发出真实采购验收指令后，才进行任何供应商写操作。**

本次构建修复了 ZovoCard 上游计划履约类型应为 `auto` 的归类问题、以 `-1` 表达 API 未提供的数字库存，并隐藏 site connection JSON 中的 API Key。管理端编辑表单不再回显 Key，留空会保留原凭据；生产已验证编辑框为空。相关 adapter / mapping / connection / procurement 测试通过，管理端 `vue-tsc`/Vite build 通过。最新包含当前连接和映射数据的发布前备份为 `/var/backups/dujiao-next/zovo-key-redaction-20261005T232858Z`（SQLite integrity_check=`ok`）。本次重新部署后管理后台 Ping 成功。

### 7.7 USDT 收款准备（2026-10-05）

检查了支付渠道管理页、渠道验证/订单创建适配器和异步回调。已有 BEpusdt 与 EPusdt 网关适配器，BEpusdt 支持 USDT-TRC20 transaction 模式，也支持网关 cashier 模式；配置需要网关 URL、认证凭据、链/币种、回调和返回 URL。统一支付回调路由为 `/api/v1/payments/callback`。项目没有“仅录入个人钱包地址即可链上识别到账并自动标记付款”的直连钱包渠道。

当前不能声称 USDT 已配置或可收款：缺少用户选定的网关及其商户配置。不要创建空字段/假凭据的活动生产渠道。继续前确认是已有 BEpusdt/EPusdt 网关商户还是需开发人工核账固定钱包方案，以及接收网络（例如 TRC20）。实际商户密钥仅通过已认证管理后台录入，不要发送到聊天。配置后需要用户授权的真实小额订单验收回调、金额、幂等、履约和退款；未验收前仍不能宣称收款链路可商用。

### 7.8 管理后台静态资源 MIME 错误修复（2026-10-06）

用户截图中的 `/admin/product-mappings` 页面报 JS/CSS MIME 为 `text/html`。根因是内嵌后台 `index.html` 由普通 Vite build 生成，静态资源地址为站点根路径 `/assets/...`；在全栈同源部署中这些请求进入用户 SPA 路由回退并拿到 HTML。仓库已有 `build:fullstack` 机制及 `<base href="__DJ_ADMIN_BASE__/">` 注入要求，但发布时嵌入的后台 dist 未按该流程生成。

已使用 `frontend/admin` 的 `pnpm run build:fullstack` 重建，并同步新 admin dist 到云端源码 `/opt/dujiao-next-src/internal/web/dist/admin`，远程 release/fullstack 二进制构建后原子替换并重启 `dujiao-next.service`。发布前运行二进制备份：`/var/backups/dujiao-next/admin-static-base-20261006T000217Z/dujiao-next`（附 SHA256 清单）。生产本机验证 `/admin/product-mappings` 返回 `<base href="/admin/">`，入口资源变为相对路径；实际 `/admin/assets/index-CH-8BsEh.js` 返回 HTTP 成功、`Content-Type: text/javascript; charset=utf-8` 和 hash 静态资源缓存头，服务 active。未修改数据库/支付渠道。需让浏览器强制刷新后台页面再确认页面交互；外部网络本轮对网站请求返回 403，故未从此执行环境验证公网/CDN层响应。

### 7.9 USDT 固定地址收款（代码已发布，渠道尚未开通）

用户选定人工核账，并提供 USDT 地址：BEP20 与 ERC20 地址均为 `0xaac9a961f425bc4b5d80e359e3c8d37f0bcd74d7`，TRC20 地址为 `TXbPPq88sTWxirGWqxXmvfrmLERhwFCXpK`。系统需创建 3 条独立的 `manual_crypto` / `usdt` 支付渠道，每条分别指定网络和同网地址；代码已经校验 EVM 与 TRON 地址的格式。

报价来源采用 Binance 官方维护的 Binance Skills Hub P2P 文档所列公开只读广告接口：USDT/CNY、用户卖出 USDT（`tradeType=SELL`）、支付宝 (`ALIPAY`)，取报价最高、网络广告限额覆盖本单 CNY 金额且余额足够的商家广告。该公开查询不需要 API Key；用户不确定的 Agent 商户权限不阻塞报价查询。按 CNY/CNY-per-USDT 换算 USDT，向上舍入至 0.01 USDT，固定快照写入该支付单。查询失败、响应异常、无合格广告或非 CNY 金额均失败关闭，不创建付款单。报价没有 Binance 锁价，快照只固定网站订单应付的 USDT 数量。

用户付款页已加入交易哈希提交；证明提交只登记待核验状态，不触发履约。新增唯一索引表 `manual_crypto_tx_claims` 防止同一链上 hash 用于多个支付单；管理员 Payments 详情可按白名单显示人工核账 payload，Finance 角色拥有审核 POST 权限。仅人工点击“确认到账并通过”后，系统通过现有支付回调事务更新支付、订单和代理账务；拒绝则允许再次提交新交易哈希。未接入自动链上验证，管理员必须确认网络、收款地址、金额、交易成功及足够确认数。

验证记录：`go build ./cmd/server`、用户端 `pnpm run build`、管理端 `pnpm run build:fullstack` 和 `git diff --check` 成功；未执行测试。代码、前后端静态资源已在备份后部署到生产，公网 `/health` 返回正常。发布不等于收款功能已启用：线上尚未写入三条生产 payment channel，也没有创建订单或提交链上交易。管理后台 UI 还需核实其是否已加载包含 `manual_crypto` 的最新静态资源，再配置三条地址渠道。真实转账/回调/履约验收需用户另行指令。

### 7.10 Binance P2P 报价权限确认（2026-10-06）

用户不确定 Binance 账户是否具备 P2P Merchant/Agent API 权限。已核对 Binance Skills Hub 官方 P2P 文档：公开广告查询 `GET /bapi/c2c/v1/public/c2c/agent/ad-list` 标注为 Public Queries (No Auth)；订单历史、订单详情、广告发布/管理等个人或商户接口才要求 API Key/Secret 或对应权限。因此，本项目读取公开 USDT/CNY 报价不需要用户提供或开通 Binance API 凭据，也无需先确认账户是否为商户。

本地使用无需认证的只读 GET 请求实测 Binance 返回 `code=000000`、20 条广告，返回结构含 `price`、`minTransAmount`、`maxTransAmount`、`tradableAmount`、`tradeMethods`、`advertiser.userType`；与适配器字段映射相符。测试的是公共行情读取，没有访问账户或下单。官方文档：[Binance P2P Skill](https://github.com/binance/binance-skills-hub/blob/main/skills/binance/p2p/SKILL.md)。公开网页接口仍可能受限流、地区网络或接口变更影响，生产换算必须在报价不可用时拒绝创建付款单；当前实现按此失败关闭。管理后台实际创建/启用收款渠道仍需单独完成，且未进行真实支付验收。

### 7.11 代理中心 Telegram 邀请卡片（2026-10-06）

用户提供目标群邀请链接 `https://t.me/+PWfyt1g1bJQ3YWRh`。将代理中心“Telegram 官方交流群”卡片整体改为新标签页打开该邀请链接，卡片内的“立即加入”文案保留为提示，避免嵌套链接；卡片同时增加 hover/focus 样式和可访问名称。该改动位于 `frontend/user/src/views/AgentCenter.vue`。本轮需构建用户前端并检查 diff；未获单独部署指令前不覆盖生产静态资源。

### 7.12 USDT 自动核账和自动履约方案评估（2026-10-06）

用户希望 USDT 支付获得第三方支付类似的自动确认和自动采购/交付。源码已有 BEpusdt、EPusdt provider、创建订单和回调验签适配；支付应用成功回调继续执行订单付款及履约队列。建议使用既有自托管支付网关来观察链上交易并回调商城，避免自行从浏览器输入的收款地址实现无订单归属的被动监听。EPusdt 新文档列出 TRC20/ERC20/BEP20 等多链和 HMAC-SHA256 回调；BEpusdt 文档列出订单金额/地址锁定、可指定 address/rate 及异步通知机制。用户希望复用自己的 Binance 收款地址；虽然后端 BEpusdt config 支持 Address，但管理表单没有该字段，且当前适配器创建请求未传 rate。因此现状不能保证可用用户提供的固定钱包，更不能保持此前指定的 Binance P2P SELL+Alipay 实时报价规则。建议采用 BEpusdt 自托管网关，并补上固定接收地址设置与每单 rate 注入；这样用户不必申请 Binance merchant API。网关仍需要在自己的后台启用网络、配置链扫描访问和 API token。生产渠道/网关凭据当前仍须核查，真实交易前不得宣称完成。

自动采购还依赖目标商品的自动履约类型、生产 ZovoCard connection 与完整 SKU 映射、供应商余额/权限和回调/轮询；支付成功只触发异步采购，不保证供应商一定成功，需要失败告警、重试/人工补偿。已参考官方项目材料：[EPusdt API 与回调说明](https://github.com/GMWalletApp/epusdt/blob/master/wiki/API.md)、[BEpusdt API 说明](https://github.com/v03413/BEpusdt/blob/main/docs/api/api.md)。本轮尚未安装或更改生产收款网关，也没有创建订单或发起链上交易。

### 7.13 向用户解释网关部署（2026-10-06）

用户尚未“注册” BEpusdt 或 EPusdt，并询问“部署”的含义。BEpusdt、EPusdt 是可以自行在服务器运行的开源收款网关软件；本项目数据库里的 `provider_type=bepusdt/epusdt` 只是商城对接入口，不能代替运行中的网关。部署指在受控服务器安装并运行网关、设置接收网络和地址、创建商城对接密钥，再让网关监测链上到账并向商城发送经签名的回调。无需先注册 BEpusdt/EPusdt 平台商户账户，也无需 Binance 商户 API 权限；实际运行仍涉及服务器、域名/HTTPS、链节点/API 连接和运维。用户现有的 Binance 充值地址在所选网关版本和链路完成验证前不能假定自动监控已生效。本轮只做概念澄清和查阅项目官方 README，没有安装网关或更改生产配置。

### 7.14 支付渠道选择与手续费说明（2026-10-06）

用户截图为新增渠道表单的 provider 选项，不能视为已经开通。通过只读 SSH 对线上 SQLite 执行 `SELECT COUNT(*), SUM(CASE WHEN is_active = 1 THEN 1 ELSE 0 END) FROM payment_channels;`，结果为 `0|`：截至本次检查，生产没有任何支付渠道。源码里“百分比手续费/固定手续费”是本站为订单计算的渠道费，公式 `fixed_fee + base_amount*fee_rate/100`；全局 `customer_fee_enabled` 决定加到用户实付还是由商家吸收；它不是 BEpusdt/其他网关收费的报价。建议新渠道默认两个字段均为 0，再单独核算链手续费、网关/节点与服务器成本及法币换算差额。

若仅求最快开始接收 USDT，可建 `manual_crypto` 三条网络渠道，指定用户已提供地址并人工核账；交易哈希提交不自动确认。若必须实现自动确认及自动采购，推荐 BEpusdt 自托管 transaction 模式：运行网关服务、配置钱包/链节点、取得 API token、在商城新增相应渠道、验证签名回调与真实小额到账。BEpusdt 官方仓库说明开源自托管，文档支持指定地址及每单 rate；现有商城后台表单未暴露 address，接口未传 rate，因此不能直接保持用户要求的 Binance P2P SELL+Alipay 实时汇率与固定钱包。`OKPAY`/`TokenPay` 等托管/平台类费用需以其当时官方商户报价和合同为准，未查明不写具体百分比。本轮未安装网关、创建渠道或发起支付。

### 7.15 BEpusdt 自托管配置与待验收事项（2026-10-06）

用户要求完成 BEpusdt 配置。生产主机 `66.94.122.149` 已运行自托管 `bepusdt-ai` 容器，镜像固定为 `v03413/bepusdt@sha256:609c795d7d3a48e88b97956b05ad4e0aeb1e53671dfed03d62a8ebf8eacc14e0`（程序 v1.24.2），仅绑定 `127.0.0.1:18090`；数据、日志分别持久化在 `/opt/bepusdt`、`/opt/bepusdt/logs`。网关管理凭据只保存在服务器 root 可读的 `/root/bepusdt-ai-bootstrap.txt`，不得复制到聊天或代码；初始 API token 已轮换。网关已写入 3 个启用的钱包：TRC20 `TXbPPq88sTWxirGWqxXmvfrmLERhwFCXpK`；BEP20、ERC20 均为 `0xaac9a961f425bc4b5d80e359e3c8d37f0bcd74d7`。管理 API 的钱包列表返回 3 条，USDT/CNY 汇率基础行存在，Binance 公共 P2P SELL/ALIPAY 报价接口在生产主机返回 HTTP 200。

商城代码增加 BEpusdt 指定地址和每单 `rate` 参数，管理界面增加地址与 `rate_source`；`binance_p2p_sell_alipay` 使用商家支付宝收款 SELL USDT 报价的中位数作为实时参考 CNY/USDT，订单创建时把该汇率提交 BEpusdt。此前取最高 20 条广告并按每个小订单的 P2P 限额过滤，会使 114、168、660、899 等低价商品没有报价，因此移除“每笔商城订单必须满足 P2P 广告最小成交额”的误用；这里仅取市场参考报价，不代用户实际执行 P2P 交易。网关返回地址与配置地址不符时拒绝创建支付。没有把 CNY/USDT 值写入支付回调的 `exchange_rate` 字段，避免共享结算逻辑误将其当作 USDT/CNY。相关源码位于 `internal/modules/payment/infrastructure/gateway/adapters/manualcrypto/adapter.go`、`adapters/bepusdt/adapter.go`、`gateway/bepusdt/bepusdt.go` 和管理端支付通道弹窗。

生产 SQLite 已创建三条 `provider_type=bepusdt`、`channel_type=bepusdt`、`interaction_mode=redirect` 通道，分别对应 TRC20/BEP20/ERC20，配置固定地址、`fiat=CNY`、实时 Binance P2P SELL/ALIPAY 汇率、内部 `gateway_url=http://127.0.0.1:18090`、签名 API token、商城异步回调与 `/pay` 返回地址。渠道费率和固定费均为 0。**三条 `is_active=0`，尚不能对客户收款。** 网关 `api_app_uri` 设为预定公网收银台 `https://usdt.ooxxoo.cn`。Caddy 已新增该域名的反向代理和公网管理 API 屏蔽规则，并成功验证、重载；Cloudflare DNS `usdt.ooxxoo.cn` 仍待用户明确确认后添加，故证书和公网收银台暂未验收。后台管理仍须使用 SSH 隧道访问本地 18090。禁止公开管理凭据和 API token。

已在 `/var/backups/dujiao-next/bepusdt-20261006T083808Z` 备份商城 SQLite、生产配置和旧二进制，数据库完整性为 `ok`。BEpusdt 范围代码和管理端 bundle 已同步至 `/opt/dujiao-next-src`，生产 release/fullstack 二进制已构建并替换，API/worker active，`127.0.0.1:8080/health` 返回 `{"status":"ok"}`。本轮未创建支付订单、未进行链上转账、未触发真实采购。仍需 DNS 生效、HTTPS 收银台和签名回调的非资金验证、以及用户单独授权后的受控小额真实支付/采购验收；验收前不能开启通道或宣称自动收款已跑通。Binance 入金地址实际入账监控、链节点稳定性和交易手续费仍需以真实交易核对。

### 7.16 BEpusdt 域名与 HTTPS 完成（2026-10-06）

用户称已将 `usdt.ooxxoo.cn` 指向 `66.94.122.149`。检查发现该 A 记录只存在于 DNSPod，而 `ooxxoo.cn` 的权威 NS 是 Cloudflare（`jobs.ns.cloudflare.com`、`nina.ns.cloudflare.com`），因此权威查询一度为 NXDOMAIN。用户明确授权在 Cloudflare 添加后，已在 Cloudflare DNS 新增 `usdt` A 记录，源站 `66.94.122.149`，代理开启、TTL Auto。权威 NS 与 `1.1.1.1` 均已返回 Cloudflare 代理地址 `104.21.50.189` / `172.67.209.98`。DNSPod 的非权威记录未动。

DNS 生效初期公网 HTTPS 返回 525；强制重新加载 Caddy 配置后，Caddy 经 HTTP-01 成功取得 `usdt.ooxxoo.cn` 的 Let's Encrypt 证书。通过 `curl --resolve usdt.ooxxoo.cn:443:104.21.50.189` 绕过本机尚未失效的 NXDOMAIN 缓存核对：收银台首页 HTTP 200、TLS 验证结果 0；公网 `/api/wallet/list` HTTP 403、TLS 验证结果 0。`ai.ooxxoo.cn` 首页 HTTP 200。部分本地递归解析器仍缓存旧 NXDOMAIN，需自然过期后再用普通域名访问复核。

商城支付回调固定路径 `/api/v1/payments/callback` 已在源码注册，生产 `settings` 中没有 `callback_routes_config` 自定义路由。向该路径发送空 JSON 返回 404 是 handler 对无法识别的回调格式的预期行为，不能据此判定真实签名回调成功；未发送伪造有效回调，也未生成支付订单。生产复核：`dujiao-next.service`、`dujiao-next-worker.service` active，`bepusdt-ai` running；三条 BEpusdt 通道全为 `is_active=0`，支付记录 0、采购记录 0。本轮仅完成 DNS、证书和无资金连通性配置，没有转账或上游采购。后续需在用户单独下达真实验收指令后，验证各链小额到账、网关监测、回调验签、订单状态、自动采购、失败补偿和链上手续费；验收前不得启用生产收款通道或宣称自动采购已跑通。

### 7.17 订阅服务前端开放 USDT 通道（2026-10-06）

用户随后明确要求在订阅服务前端配置并开放 USDT 收款通道，该新指令授权启用此前停用的三条 BEpusdt 生产通道。启用前核对：商城 API/worker active、BEpusdt 容器 running，收银台 HTTPS 200 且 TLS 验证通过，Binance P2P 公开报价接口可达；网关 3 个钱包启用且链与地址对应；商城三条通道配置分别为 `usdt.trc20`、`usdt.bep20`、`usdt.erc20`，费率/固定费用均 0，适用角色、等级、支付类型均无限制。4 个在售商品 `fulfillment_type=manual`、`is_mapped=0`；支付和采购记录均为 0。前端结账页已有支付渠道选择组件，读取公开配置或登录态下的可用通道接口。

操作前使用 SQLite 在线备份至 `/var/backups/dujiao-next/usdt-open-20261006T091338Z/dujiao.db`，备份完整性 `ok`。在事务中将 3 条 BEpusdt 通道 `is_active` 从 0 改为 1，清理 Redis `public:config:*` 缓存；未修改商品价格、地址、网关 token、履约类型或上游映射。操作后生产 SQLite 完整性 `ok`；`GET /api/v1/public/config` 经商城原站 API 返回 `status_code=0`，公开 `payment_channels` 包含三条 USDT 通道。公网该配置 API 返回 HTTP 200。商城 API/worker active，网关 running；支付记录 0、采购记录 0。

**当前客户已可在结账页选择 USDT TRC20/BEP20/ERC20 并可能实际付款。** 这次未创建订单、未转账，也未验证真实到账识别、签名回调或上游采购，因此不得称完整自动收款/采购已通过验收。现有商品仍为人工履约，支付成功后不应宣称自动采购。运营方需监控新支付、网关订单和链上到账；真实小额支付与采购验收仍按用户先前要求等待单独指令。

### 7.18 代理资格提示闪现与拿货订单入口修复（2026-10-06）

用户截图显示：已激活代理账号进入 `/agent` 时短暂出现“激活官方代理权限”；拿货订单列表卡片无法点击。根因在 `frontend/user/src/views/AgentCenter.vue`：`refreshAgentIdentity()` 发起异步管理资料查询前先把 `agentIdentityActive` 设为 false，而模板立刻按 `!isAgent` 渲染“未开通”提示；查询失败时也被误判为未开通。拿货订单行原为普通 `div`，没有任何导航事件。后端已有 `/orders/:order_no` 详情路由和 `/pay?order_no=...` 付款路由；订单详情仅在 `pending_payment` 状态提供付款按钮。

已把资格查询改成明确的“查询中 / 查询失败可重试 / 未激活 / 已激活”展示；登录和注册提交时提前置为查询中，避免 auth 状态切换与资料查询之间再闪现错误提示。拿货订单行改为整行可点的 `router-link`：`pending_payment` 进入付款页，其他状态（包括 `canceled`）进入只读订单详情；已取消订单不重新开放付款。新增“继续付款/查看详情”提示和键盘焦点样式。此前用户要求的 Telegram 官方交流群新邀请链接随本次 `AgentCenter.vue` 一并发布。

为了避免把本地另两处尚未发布的 `frontend/user/src/api/client.ts`、`src/utils/affiliate.ts` 差异带入生产，本次用生产前端源码基线加本地新版 `AgentCenter.vue` 在独立临时目录构建；`vue-tsc -b && vite build` 成功。发布前备份生产二进制及旧组件源码至 `/var/backups/dujiao-next/agent-center-20261006T092556Z`，再同步组件源码及前端构建产物到生产源码，重建 release/fullstack 二进制并重启 API。发布后 `dujiao-next.service` active、`/health` 成功、原站和公网 `/agent` 均 HTTP 200，公网新 `AgentCenter` 资源 HTTP 200 且包含新增文案及 Telegram 邀请链接；本地与生产组件 SHA256 一致。没有创建、取消或支付任何订单。未使用真实代理会话逐项点击验收；若有浏览器旧缓存，刷新页面再观察。

### 7.19 BEpusdt 管理登录说明（2026-10-06）

用户看到 `https://usdt.ooxxoo.cn` 的 BEpusdt 公开收银台首页并询问是否需要登录。该公开页面不是管理后台；客户付款无需登录 BEpusdt，网关管理订单/钱包/运行状态才需管理员登录。公网 Caddy 禁止访问网关管理 API/秘密入口。部署时生成的网关管理员账号、密码及秘密后台入口只保存在服务器 root 私有文件 `/root/bepusdt-ai-bootstrap.txt`（权限 `600 root:root`），与商城和 Binance 账号不同；本轮没有在聊天中回显凭据，也未重置。已核实秘密入口在网关内网返回 HTTP 200。

本轮在本机建立了仅绑定 `127.0.0.1:18090` 的 SSH 本地端口转发到服务器同一内网端口，并核实本机 `http://127.0.0.1:18090/` 返回 HTTP 200。用户可用自己持有的 SSH 密钥读取上述私有文件，再在浏览器访问 `http://127.0.0.1:18090` 加文件中的“后台管理入口”路径进行登录。该本地 SSH 通道为本机后台进程，不保证重启/断线后仍在；如失效需重新建立。不要在公网分享管理员密码、秘密入口或 API token。

### 7.20 BEpusdt 管理入口路径拼接说明（2026-10-06）

用户已打开本地 `http://127.0.0.1:18090`，询问“加上后台管理入口路径”的含义。应从服务器 root 私有文件 `/root/bepusdt-ai-bootstrap.txt` 中找到“后台管理入口：”后以 `/` 开头的路径，把它直接接在 `http://127.0.0.1:18090` 后面，**不要再添加末尾 `/` 或 `/#/login`**。例如文件若显示占位路径 `/example123`，浏览器地址写成 `http://127.0.0.1:18090/example123`，网关会设置后台会话并重定向到 `/#/login`。这里的 `example123` 纯属示例，不能用于真实登录。此前误写的 `.../example123/#/login` 会返回 404，已在下一轮纠正。管理员账号和密码也在同一私有文件中。

### 7.21 BEpusdt 精确登录地址纠正（2026-10-06）

用户要求直接给出当前网关的完整地址。已从 root 私有文件读取后台秘密入口并向用户提供本机直达 URL；本机 SSH 隧道仍可访问。核实入口路径本身返回 HTTP 302 并设置后台会话，随后重定向到 `/#/login`，最终 HTTP 200；入口后额外添加 `/` 会返回 404。因此必须直接访问文件中给出的入口路径，不在末尾加斜杠或手工拼接 `/#/login`。此处不重复记录秘密路径值或管理员密码。

### 7.22 三条 USDT 收款通道图标（2026-10-06）

用户要求为订阅服务前端的三条 USDT 通道直接补图标。已在仓库 `assets/payment-icons/` 创建 USDT 绿色主标识加链网络角标的 SVG 源文件和 128×128 PNG：TRC20 红色 TRON、BEP20 黄色 BNB、ERC20 蓝色 Ethereum。SVG 无脚本/外部资源；生产仅上传 PNG，避免浏览器 SVG 下载策略影响图标展示。PNG 已人工查看外观。

生产数据操作前备份 SQLite 至 `/var/backups/dujiao-next/usdt-icons-20261006T095004Z/dujiao.db`，完整性 `ok`。PNG 已安装到持久化上传目录 `/opt/dujiao-next/uploads/common/2026/10/usdt-{trc20,bep20,erc20}.png`，所有权 `dujiao:dujiao`。在事务中更新 `payment_channels` #1/#2/#3 的 `icon` 为对应 `/uploads/common/2026/10/...` 路径，并新增三条 `media` 记录供管理后台素材库显示；清理 Redis 公开配置缓存。未更改渠道地址、汇率、费用、启用状态或订单。

生产复核：SQLite 完整性 `ok`；公网 `/api/v1/public/config` 返回三个通道各自的 icon 路径；三张公网图片均为 HTTP 200、`image/png`。三条 BEpusdt 通道仍启用，商城 API active，支付记录仍 0。前端结账页及管理端渠道列表原本会按 `icon` 字段渲染，无需重新部署前端包；现有已打开页面可刷新获取新公开配置。

## 本轮作业总结（2026-10-06，前端全链路复核）

### 已完成的审查与修复

- 工作区包含大量早先的既有修改与新增文件，未做重置或覆盖。仅凭当前 Git 工作区无法准确把每一项未提交修改归因到 Antigravity；本轮以当前代码行为为准进行了核对。
- 代理售价设置改为**每个商品卡片单独保存**，删除了顶部和底部两个全局保存按钮。保存按钮仅提交该商品的售价，显示独立保存中、成功、失败和未保存状态；配置读取失败时禁止保存，避免默认值覆盖旧售价。加载既有设置后以实际值初始化“已保存”状态。建议价按钮明确为仅修改未保存内容。
- 线上公开商品 API 核对了四个商品与 SKU 映射：商品/SKU ID 为 1–4，与代理售价页面固定映射一致。
- 删除了推广链接中伪造代理码 `AGENT<用户ID>` / `VIP888` 的回退值。代理资料未读取到有效专属码时不再生成可能错误归因的链接，且复制链接、二维码及推广文案操作禁用并提示资料读取问题。
- 代理拿货订单、客户销售订单、提现历史读取失败时现在显示错误状态，不再把接口故障伪装成“暂无记录”。
- QuickBuy、classic/vault 支付结果页新增醒目的 TRC20/TRON、BEP20/BNB Smart Chain、ERC20/Ethereum 网络标识和转错网络提示；QuickBuy 会按商品允许的通道筛选，切换商品时校正不可用的已选通道。
- BEpusdt 官方 checkout 源码另有一份网络提示补丁，位于 `ops/patches/bepusdt-checkout-usdt-network-warning.patch`。补丁只更新收银台界面与前端 JS/CSS；需要在 BEpusdt 源码目录应用、重建其嵌入式静态资源并单独部署，尚未部署到生产。
- 首页购买文案已改为“结算页选择当前可用的支付方式”以及“支付后按订单提示提交 Session，由客服处理开通”，避免把当前人工履约误述成支付宝/微信专属或自动采购交付。首页增加了非精选商品搜索购买区；商品 API 失败或价格缺失时不再展示硬编码旧价，也禁止从首页直接购买。

### 审计发现与仍需处理

- 当前 `Home.vue` 相对基线移除了原先的后台横幅轮播、`template_mode` 列表/卡片切换、分类侧栏和分页商品目录。虽然本轮为额外商品补回搜索购买区（最多请求 100 条），后台横幅与分类/模板配置仍不会完全生效；如果后台仍依赖这些功能，应继续恢复对应控件与分页。
- 首页套餐流程、退款/账号限制条款仍有较多硬编码中文；对英文/繁体界面不一致。固定的“退款扣除 20%”应由业务方确认与当前退款政策、支付渠道扣费规则一致后再继续展示。
- 代理中心其余财务看板/提现申请的整体失败与空状态反馈仍值得逐页复核；当前修复覆盖了订单/销售订单/提现历史和售价配置，未覆盖所有 API 消费点。
- 本轮没有操作生产发布或真实付款、提现、采购。BEpusdt 收银台补丁也未上线；需要另行执行部署和页面视觉验收。

### 验证

- `frontend/user` 与 `frontend/admin` 的 `pnpm build` 均通过。构建中仍有 Node 弃用提示、Browserslist 数据过旧和第三方依赖注释提示，不影响构建。
- BEpusdt checkout 的 JS 通过 `node --check`，补丁 `git diff --check` 通过；未在此工作区构建 BEpusdt 二进制或镜像。
- 本轮未重新执行后端全量测试；本轮未修改后端业务代码。

## 本轮作业总结（2026-10-06，生产发布与支付宝文案校正）

### 用户要求

- 保留“支付宝通道审批通过并开放后也可选择支付宝”的将来时说明；在审批完成、后台实际开放之前，不宣传支付宝当前可用。
- 将本轮其他用户端和 BEpusdt 网络提示改动同步到生产。

### 已发布

- 只把六个用户端文件（`Home.vue`、`AgentCenter.vue`、`ProductQuickBuy.vue`、`usePayment.ts`、classic 与 vault 的 `Payment.vue`）覆盖到云端源码；后台未提交的其他改动、数据库及支付渠道配置没有随本次发布覆盖。
- 以生产现有前端源码为基底构建并嵌入用户端静态包。保留支付宝审批完成后可用的文案。生效内容包括：代理商品卡片独立保存售价、错误/保存状态反馈、推广码无效时不再生成伪造链接、代理订单错误态、购买弹窗与支付页醒目显示 TRC20/TRON、BEP20/BNB Smart Chain、ERC20/Ethereum 网络和转错链提示、按商品允许渠道筛选支付方式，以及此前约定的 Telegram 群入口。
- BEpusdt 使用当前生产容器配置和 v1.24.2 源码构建了带链网络说明的程序，只替换容器内 `/usr/local/bin/bepusdt`，未重建或改变现有端口、卷、环境配置和重启策略。生产程序 SHA256：`d53342534ceee10eebd469683d0c869e71731606e2080c5aa4b4be232d4d2abd`。回退镜像：`local/dujiao-bepusdt:before-network-banner-20261006T141016Z`；当前打补丁镜像：`local/dujiao-bepusdt:network-banner-20261006`。

### 生产备份与验证

- 订阅站发布前备份：`/var/backups/dujiao-next/user-chain-agent-20261006T140702Z`，包含六个目标源文件、旧 `internal/web/dist/user` 和旧服务二进制，并有 `SHA256SUMS`。
- BEpusdt 发布前备份：`/var/backups/dujiao-next/bepusdt-20261006T141016Z`，包含原容器程序及 SHA256 清单。
- `dujiao-next.service` 与 `dujiao-next-worker.service` 均 active；`https://ai.ooxxoo.cn/health` 和首页返回 HTTP 200。
- BEpusdt 容器 running，重启策略保持 `unless-stopped`；容器内新程序哈希与构建产物一致；网关本机 18090 和 `https://usdt.ooxxoo.cn/` 返回 HTTP 200。打补丁镜像已保存，原镜像仍在供回退。
- 未创建订单、未发起真实付款/链上转账，也未触发 ZovoCard 采购。健康检查只验证服务入口；未使用真实支付会话验收收银台交易流程。

### 注意事项与接手

- 前一节“本轮没有操作生产发布、BEpusdt 收银台补丁尚未上线”的描述记录的是发布前状态，现已由本节的生产发布结果更新。
- 首页移除横幅轮播、分类/分页目录和模板模式的遗留问题仍待恢复或由产品确认；错误文案的国际化与代理中心其余财务面板的完整空态审查也未完成。
- 生产构建时发现远端源码中两个不在本次改动内的孤立文件 `frontend/user/src/affiliate.ts`、`frontend/user/src/client.ts` 会导致 TS 检查失败。只在临时构建目录将它们改名以完成构建；生产源码中的文件没有删除或修改。应单独检查这两个文件与当前 API 模块结构。
- BEpusdt 支付页的链提示已随 Go 内嵌资源发布；由于本轮约束不创建真实订单，尚未在真实支付页面视觉走查。可在不发起链上付款的前提下，用受控测试支付会话后续做截图验收。

## 本轮作业总结（2026-10-06，付款流程页面简化评估）

- 用户提供四张连续付款页面截图，询问页面 1、2、3 的重复内容能否在不损害支付安全和完成流程的前提下简化。本轮只做代码与流程评估，未修改代码、未部署。
- `ProductQuickBuy.vue` 当前展示商品、数量、收货/游客查单信息、支付渠道和应付总额；“立即支付”实际调用用户或游客的 `create-and-pay`，并传入 `channel_id`。后端 `CreateHandler.respondCreateAndPay` 随即创建支付单。因此它不是单纯的商品确认页。
- `/pay` 页面由 `Payment.vue` 处理至少两种状态：支付前选渠道/确认，和支付单建立后的待支付结果。渠道、金额、订单详情和订单号在状态间有重复。截图 3 中“更换支付方式”会清理前端支付展示状态，再回到支付前选择；后端 `PaymentService` 按订单及渠道复用并 supersede 支付单，实际改渠道时仍需保留支付状态同步和订单状态校验。
- 建议目标是保留三个清晰阶段，而不是删掉最终付款核验：①购买弹窗只收集商品/数量和游客必要信息，主按钮改为“去结算”或“确认并继续”，去掉第二处重复的支付渠道选择；② `/pay` 统一承担最终价格、渠道/手续费确认与发起支付，并在同一页面切换“待确认”和“等待付款”状态，等待态折叠订单商品和费用明细，只突出订单号、最终应付币种/金额、网络、支付凭证/链接、倒计时、刷新和安全的更换方式；③ BEpusdt 外部收银台保留，因为 QR、地址、链上金额、网络警告和链上转账步骤不可被商城页面替代。
- 改造前需确认 `create-and-pay` 旧流程的所有调用点（quick buy、购物车及其他入口），统一接口语义；不能只隐藏控件而继续在弹窗先创建支付单，否则可能造成重复建单/渠道不一致。更换渠道需以服务端 supersede/失效结果为准，绝不把前端清空视为旧支付失效；已支付、处理中、过期或取消订单需要分别展示明确终态。

## 本轮作业总结（2026-10-06，付款流程简化实施与生产发布）

### 实施内容

- 快速购买弹窗移除支付通道选择和网络提示（网络提示仍在实际选定支付通道后的 `/pay` 页面和 BEpusdt 收银台展示），保留商品、数量、游客邮箱/查单密码。金额明确标为“预估商品金额”，提示最终金额/支付方式在结算确认。
- 快速购买按钮改为“去结算”，只调用用户/游客创建订单 API，不再传 `channel_id` 或提前创建支付单。创建成功后保存游客查单凭证并进入 `/pay?order_no=...`；最终订单金额、手续费、可用支付渠道和支付创建仍由结算页与服务端确认。
- `/pay` 仍使用同一路由在待确认与等待付款状态之间切换；等待付款状态保留支付链接/二维码、链提示、订单状态轮询、有效期/倒计时、刷新和更换支付方式。移除结果区订单摘要中与标题重复的支付方式一行。BEpusdt 独立收银台保持原样。

### 生产发布与核验

- 用户端前端 `pnpm build` 成功；以云端当前源码基底构建、嵌入并原子替换服务二进制。未修改后端支付逻辑、数据库、支付渠道配置或 BEpusdt。
- 发布前备份：`/var/backups/dujiao-next/payment-flow-20261006T142545Z`，包含旧 QuickBuy/Payment 源文件、用户端 dist、服务二进制及 SHA256 清单。
- `dujiao-next.service` 和 `dujiao-next-worker.service` 均 active；生产 `/health` 和首页均 HTTP 200。QuickBuy 生产 JS 资源 HTTP 200，检查确认含结算说明且不再含 `createAndPay` 调用。
- 未创建真实订单、未付款、未触发采购。没有在浏览器中走访客/登录用户两种购买并付款至 BEpusdt 的完整交互；仅进行了构建及只读生产健康/资源核验。

## 本轮作业总结（2026-10-06，结算支付方式与操作区合并）

- 用户要求把付款页左侧“选择支付方式”与右侧“发起支付”合并为单一卡片，卡片放在原右侧操作卡片位置。
- classic 与 vault 两套 Payment 页面都已将钱包余额选项、缓存支付恢复入口、支付通道选择、提示/错误、倒计时和确认/刷新按钮放进右侧支付卡片；左侧仅保留订单、金额和商品明细。倒计时不再在订单摘要重复显示。窄侧栏的支付通道改为单列，避免两列卡片挤压。新增中/英/繁体标题“选择支付方式并发起支付”。
- `frontend/user` 本地构建与生产基线暂存构建均通过。生产仅覆盖以上用户前端页面/组件/语言文件及重新嵌入的 user dist，不改后端、数据库、渠道配置或 BEpusdt。
- 发布前备份：`/var/backups/dujiao-next/payment-card-layout-20261006T143638Z`，含旧源文件、旧用户端静态资源、旧服务二进制与 SHA256 清单。
- `dujiao-next.service`、`dujiao-next-worker.service` 仍 active；生产 `/health` 与首页返回 HTTP 200；云端 classic/vault 支付页及选择组件源文件与本次发布版本一致。
- 未创建订单或支付；没有通过浏览器做视觉验收，用户刷新页面后即可查看布局。若页面保留旧缓存，需强制刷新。

## 本轮作业总结（2026-10-06，首页横幅与代理入口）

- 用户要求缩小首页 GPT 横幅高度，在“查看代充订阅详情”右侧增加“成为代理，稳定副业”按钮并跳转代理中心，同时保持清新淡雅且按钮区分清楚。
- 首页横幅纵向间距、内边距、标题和两枚品牌图标均已收小；主操作采用深青绿色实底，代理入口采用浅薄荷色底和柔和边框。两按钮窄屏自动换行，并有键盘焦点提示。代理按钮使用正式路由 `/agent`。
- 以生产前端源码基线叠加本次 `Home.vue` 变更构建，`vue-tsc -b && vite build` 成功。先部署静态资源，再构建 Go 二进制以确保新资源被嵌入。发布中两次因候选二进制顺序/旧页面缓存核验而触发回滚；最终按正确顺序发布成功，数据库、订单、支付通道和 BEpusdt 均未变更。
- 发布前备份：`/var/backups/dujiao-next/home-agent-cta-20261006T145042Z`，包括 `Home.vue`、用户端 dist、旧服务二进制及 SHA256 清单。
- 最终验证：生产 `dujiao-next.service` 和 `dujiao-next-worker.service` 均 active；`http://127.0.0.1:8080/health` 返回 `{"status":"ok"}`；通过带验证参数的公网首页入口和新首页分包检查按钮文案；生产源文件 SHA256 与本地变更文件一致。浏览器视觉效果尚未自动截图验收；用户刷新首页即可查看，若旧缓存仍在请强制刷新。

## 本轮作业总结（2026-10-06，代理中心隐藏个人中心导航）

- 用户确认去掉代理中心页面顶部“个人中心”，并确认仍需能看业绩和提现。变更限定在 `/agent`：仅当路由名为 `agent-center` 时隐藏顶部个人中心链接；普通用户和其他页面的个人中心入口保持原样，代理中心自身的业绩、客户订单、提现结算栏目和 API 均未更改。移动端底部导航不在本次“顶部”修改范围内。
- 以生产前端基线合入本次 `Navbar.vue` 变更，`vue-tsc -b && vite build` 成功。生产 Go 二进制重新嵌入新静态资源后发布。
- 发布前备份：`/var/backups/dujiao-next/hide-personal-agent-20261006T150300Z`，含旧 `Navbar.vue`、用户端 dist、旧二进制及 SHA256 清单。
- 发布验证：API 与 worker 服务均 active，健康接口返回 `{"status":"ok"}`；公网入口带验证参数返回本次构建的入口 bundle（`index-COHm0xMo.js`），bundle 包含 `agent-center` 路由判断。未用已登录浏览器进行视觉验收；代理可刷新 `/agent` 查看，其他页面的个人中心链接不受影响。

## 本轮作业总结（2026-10-06，游客查单凭据不记忆）

- 用户指出游客订单查询表单会记住历史邮箱与订单密码，要求该区域不留历史信息，避免泄露。Classic 与 vault 两套游客查单页已移除“已保存账号”和清除历史提示，页面打开时输入框为空；邮箱和密码字段设置了 `autocomplete=off` / `new-password` 与专用 name，降低浏览器自动填充。
- `guestOrderAuth` 不再读取或写入 `sessionStorage` / `localStorage`，访问时清除旧版本可能遗留的 `guest_order_auth`，凭据只在当前 SPA 内存中为正在进行的游客查单/支付/订单详情流程短暂传递。离开相关页面时清除；重新载入或关闭标签页后不恢复。游客详情页输入框也增加禁用自动填充属性。访客仍可输入邮箱、订单密码查单；从列表跳到详情或待支付页的当前流程保留必要授权。
- 以生产基线构建，`vue-tsc -b && vite build` 成功；Go 服务重新嵌入用户端静态资源后发布。
- 发布前备份：`/var/backups/dujiao-next/guest-order-no-memory-20261006T153456Z`，含八个目标源文件、旧用户端 dist、旧服务二进制和 SHA256 清单。
- 发布核验：API 与 worker 均 active，健康接口返回 `{"status":"ok"}`；线上 classic/vault 游客查单分包都包含专用查询字段标记，凭据模块 bundle 无 `setItem` 写入调用；新入口 bundle 线上可取。未使用真实游客账号执行订单查询或付款。

## 本轮作业总结（2026-10-06，首页 Telegram 联系入口与页脚收紧）

- 用户要求在首页“精选服务”标题旁添加 Telegram 联系入口，并缩紧页面底部页脚。首页新增“联系我们”浅蓝色 Telegram 图标按钮，链接复用后台站点配置 `contact.telegram`，新标签页打开并带 `noopener noreferrer`；若后台未配置链接则不显示按钮。
- 页脚纵向内边距从 `py-16` 收到 `py-7/md:py-8`，品牌区和联系方式的间距、图标与文字尺寸缩小；Telegram/WhatsApp 联系方式改为紧凑的并排按钮；版权与 GitHub 信息合并一行，隐私政策、服务条款和后台附加链接收拢至同一行。原有链接仍保留。
- 生产源码基线 `vue-tsc -b && vite build` 成功，先切换静态资源后重新构建 Go 嵌入式二进制并发布。发布前备份：`/var/backups/dujiao-next/home-telegram-footer-20261006T154223Z`，包含旧 Home/Footer 源码、旧用户端 dist、旧二进制与 SHA256 清单。
- 发布验证：`dujiao-next.service` 与 `dujiao-next-worker.service` active；健康接口返回 `{"status":"ok"}`；线上新首页分包包含 Telegram 联系入口，新主 bundle 包含紧凑页脚类名，新首页入口 HTML 指向本次构建。没有做浏览器截图视觉验收，刷新首页即可查看。

## 本轮作业总结（2026-10-06，移除全站个人中心导航页签）

- 用户要求在其他页面也移除“个人中心”页签，暂时不通过导航使用个人中心。本次移除桌面商店顶栏的个人中心按钮、登录用户移动端底部“个人中心”项，以及独立代理控制台顶栏和移动抽屉中的个人中心快捷入口。移动端未登录时仍保留“游客查单”；登录后底部只显示“首页”和“代理中心”。
- `/me` 及其资料、订单、钱包、安全设置路由和页面代码保留；支付/认证等业务流程现有的定向导航不会被本次改动删除。代理业绩与提现功能未改，仍从代理中心查看。
- 使用生产前端基线合并三处导航组件的改动，`vue-tsc -b && vite build` 成功。发布用户端静态资源并重新构建嵌入式 Go 二进制。
- 发布前备份：`/var/backups/dujiao-next/remove-personal-center-20261006T155957Z`，包含旧 Navbar、移动底栏、代理控制台顶栏、用户端 dist、旧二进制及 SHA256 清单。
- 发布后 API 与 worker 均 active，健康接口返回 `{"status":"ok"}`；公网返回本次新入口 bundle，并验证 bundle 中保留 `/agent` 路由。未做登录态浏览器截图验收；发布范围未涉及数据库、订单、支付配置或代理财务逻辑。

## 本轮作业总结（2026-10-07，北京时间，后台白屏修复与 Telegram 成交播报）

### 后台白屏

- 根因：线上 `dujiao-next` 二进制嵌入的后台 `index.html` 仍引用 `/assets/...`，而后台实际挂载在 `/admin/`；浏览器因此收到 SPA HTML 作为 JS/CSS，严格 MIME 检查阻止加载。
- 用现有 `frontend/admin` 运行 `pnpm run build:fullstack`，更新服务器嵌入资源后在生产主机重新构建全栈二进制并重启 `dujiao-next.service`。
- 生产 HTML 现在含 `<base href="/admin/">` 并使用 `./assets/...`。HTTPS 线上 JS 返回 `200 text/javascript`、CSS 返回 `200 text/css`；在本机 Chrome 登录并确认仪表盘可见。
- 服务器重建备份位于 `/var/backups/dujiao-next/admin-static-fix-20261006T193516Z`。

### Telegram 成交播报

- 功能源码基于生产代码更新，准备工作树 `/tmp/cdk-telegram-broadcast`，HEAD `fd759d5`；原仓库内既有未提交改动未动。
- 已部署 `zovo-cdk.service` 和前端，服务健康检查通过；部署备份 `/var/backups/zovo-cdk/telegram-broadcast-20261007T034840`。
- 当前 Chrome 已打开 `https://cdkadmin.ooxxoo.cn/ops/webhooks` 的“Telegram 成交播报”弹窗；Chat ID `-1004450863515` 已填，Bot Token 为空。等用户在本机粘贴 Token 并点“保存后测试”，助手不可查看/复制 Token。
- 当前功能仅在卡台 `gpt_direct.completed` 成功回调触发，按订单号唯一去重；付款但充值失败不播报。后续需用户完成 Token 测试，再核验卡台事件订阅并用真实成功订单验证最终文案（商品名、遮蔽邀请码、北京时间）。
- Telegram 功能源码尚未发布到 GitHub Release；生产已直接更新，无需等待 Release。

## 本轮作业总结（2026-10-07，Telegram 链接预览）

- 原因：`https://ai.ooxxoo.cn/` 的静态入口只有空 `<title>`，SEO 元数据由 Vue 在运行时补上；Telegram 等爬虫不执行 JavaScript，因此无法提取标题、简介和图片。
- 在 `frontend/user/index.html` 补上中文标题、description、canonical、站点图标链接、Open Graph 与 Twitter Card 元数据。
- 新增 `frontend/user/public/og-preview.svg` 源图和 1200×630 `og-preview.png`，内容按线上站点品牌“GPT源头代充 / GPT订阅源头 / 用户自己的账号·官方直充”制作。
- 已构建用户前端、替换 fullstack embed 中的用户资源并重建部署 `dujiao-next.service`。备份：`/var/backups/dujiao-next/og-preview-20261007T052731`。
- 线上 `ai.ooxxoo.cn/` 对 TelegramBot/Twitterbot User-Agent 均返回 HTTP 200，含 OG 标题、描述、URL 和图片；`/og-preview.png` 返回 `200 image/png`。
- 用户输入的 `ai.xxooxx.cn` 目前无法解析；本次按已有生产站点域名 `ai.ooxxoo.cn` 修复。如前者是另一正式域名，需先确认 DNS/主机绑定。

## Telegram Android 真实预览复测（2026-10-07）

- 用户在 Telegram Android 发出 `https://ai.ooxxoo.cn` 后仍只显示纯链接。尽管原始 HTML 对 TelegramBot 模拟 UA 返回静态 OG 信息、图片 200 PNG，尚不能据此认定客户端生成了预览。
- 线上根 URL响应 `200`、`Cache-Control: no-cache`，robots.txt 对根页面开放。可能是 Telegram 缓存了更早的无预览结果或 Telegram 侧预览缓存/策略问题。
- 后续先用 Telegram 的 `@WebpageBot` 刷新 URL 缓存，再在聊天重发；若仍失败，测试一次带唯一查询串的主页 URL并检查 Telegram/Cloudflare 抓取记录。不要宣称此项已端到端解决，直到用户确认 Telegram 中出现卡片。

## Telegram 预览用户验收（2026-10-07）

- 用户提供的 Telegram Web 截图确认：`https://ai.ooxxoo.cn/?tg=20261006` 已展示完整标题、描述和 OG 分享图。
- 根网址的旧消息仍为纯文本链接；后续对外分享可用上述带 query 的主页 URL。此链接点击后仍加载首页。

### 7.14 Telegram 分享卡片文案与根网址（2026-10-06/07）

- 管理后台「系统设置 → 站点设置」已保存简中 `brand.site_description`、`seo.title`、`seo.description`。新标题：`GPT订阅源头代充｜用户自己的账号，官方直充`；新描述：`可信赖的源头代充服务，用户使用自己的账号完成官方订阅。代理轻松开启副业，拓展稳定现金流。`
- 静态用户首页 `frontend/user/index.html` 已同步上述文案，并移除 `og:site_name`，避免分享卡显示重复标题。社交图来自 `frontend/user/public/og-preview.png`，仅为 OG/Twitter 分享封面，不是商城站内页面图片。
- 已更新生产嵌入式用户 SPA index 并重建服务二进制；`dujiao-next.service` active。备份 `/var/backups/dujiao-next/share-meta-20261006T214722Z`。
- 生产 localhost 公共配置 API 返回已保存字段；线上根网址对 TelegramBot/Twitterbot UA 为 HTTP 200，新标题/描述生效，OG 图片 `https://ai.ooxxoo.cn/og-preview.png` 为 HTTP 200 PNG。
- 公网 JSON API 在本轮执行主机上返回 Cloudflare 403，故通过生产机本机 API 核对配置；公网根 HTML 和图片可访问。Telegram 对精确根 URL 的最终客户端卡片状态仍需用户重发验证，若缓存旧失败结果应先用 `@WebpageBot` 刷新后重发根网址。

### 7.15 修复分享元数据发布导致的用户端白屏（2026-10-06/07）

- 用户反馈 Telegram @WebpageBot 对根域刷新成功，但 iPhone Safari 打开网站呈整页空白。复核后发现上一节静态 OG 更新部署时，将源码 `frontend/user/index.html`（开发脚本 `/src/main.ts`）直接复制到嵌入 dist，绕过了 Vite 构建。这造成新打开浏览器无法执行 SPA；桌面已有标签保留旧页面，掩盖错误。需特别注意源码 index 不是生产 index。
- 已从 `frontend/user` 执行 `pnpm run build`（`vue-tsc -b && vite build` 成功），完整 dist 共约 4.3 MB，正确引用 `/assets/index-CpgOk-3U.js`、`/assets/index-CqXz-BIk.css`；同步 dist 后在生产重建 Go `release,fullstack` 二进制并重启 `dujiao-next.service`。
- 备份了修复前线上资源和二进制：`/var/backups/dujiao-next/user-build-repair-20261006T215725Z`。
- 验证：systemd active，localhost `/health` 返回正常；localhost 和公网根网址不再含 `/src/main.ts`；公网 JS 为 `200 text/javascript; charset=utf-8`，CSS 为 `200 text/css; charset=utf-8`。iPhone Safari UA 请求验证通过。Chrome 重新加载后首页完整渲染。
- 新访客白屏已修复。待用户在 iPhone 重新打开确认。若 Telegram 分享卡仍未出现在目标会话，请在 WebpageBot 刷新后，另向目标群或私聊发送一次根 URL 验证；Bot 显示“updated successfully”本身只是缓存刷新回执。

### 7.16 Telegram 精确根网址预览仍不出现（2026-10-06/07）

- 用户再测 `https://ai.ooxxoo.cn/` 依然无卡片；同一 Telegram Web 聊天中 `/?tg=20261006` 已有完整卡片。@WebpageBot 向用户报告更新成功并未恢复根网址预览。
- TelegramBot UA 对公网根网址和查询串 URL 请求均返回 HTTP 200、相同 HTML（2610 字节，SHA-256 前 16 位 `089d04959d3eea8c`）；分享图 HTTP 200 PNG（782300 字节）；robots.txt HTTP 200；未见 X-Robots-Tag。生产应用及静态构建正常。
- Telegram 官方开放问题 `https://bugs.telegram.org/c/65828` 描述同一根 URL 失败、查询串成功、@WebpageBot 仍显示成功的现象。本站问题与其吻合，但尚无真实 Telegram 爬虫到达本站的可用 Cloudflare/Caddy 日志，不应把“内部负缓存”写成已证实根因。
- 本轮未再修改业务代码或服务器设置。现阶段已验证可用的分享 URL 为 `https://ai.ooxxoo.cn/?tg=20261006`；精确根 URL 需 Telegram 侧修复或清理其预览状态。若后续继续排查，可在受控时间窗口记录 CDN/源站访问日志，并让用户再次发送精确根 URL 以确认实际爬虫请求链路。

### 7.17 根 URL 抓取链路补充核对（2026-10-06）

- 外网复核 no-slash/root slash/`?tg=20261006` 在 Telegram UA 下均 200，均解析至同一个 `/` 页面；root/query 响应体相同（2243 字节）。生产页面引用已构建 `/assets/...` 文件，静态 OG title/description 均有，未再出现开发入口。
- OG PNG 200 `image/png`（782300 字节）；robots.txt 200，仅封锁后台/账户/支付类路径，没有封锁首页。DNS A/AAAA 均为 Cloudflare 地址。
- 目前不能从源站验证 Telegram 真正的抓取事件：`/var/log/nginx/access.log` 为空，Caddy 的 `ai.ooxxoo.cn` 站点块没有 access logging，当前没有 Cloudflare 边缘请求事件读取途径。请勿声称 Telegram 未请求源站已被证实。
- Telegram 当前表现与官方开放问题 https://bugs.telegram.org/c/65828 一致。站内页面与资源端检查正常；精确根 URL 预览状态仍依赖 Telegram 侧刷新/处理。验证过的 query URL 仍可临时分享。

### 7.18 Cloudflare 真实 TelegramBot 抓取记录核验（2026-10-07）

- 用户授权查看已登录的 Cloudflare 控制台。仅查看 `ooxxoo.cn` 过去 24 小时请求分析与安全分析，未改 DNS、WAF、安全规则或其它线上配置。
- 请求分析按精确 User-Agent `TelegramBot` 筛选后得到 3 次请求（12.01 kB，2 个访客）：主页 `/` 2 次，`/og-preview.png` 1 次；3 次均 HTTP 200。说明 Cloudflare 侧实际观察到 Telegram 抓取页面和 OG 图，且无边缘阻断。
- 全站安全分析显示 34 次被 Cloudflare 缓解，但过滤后的 TelegramBot 请求状态全为 200。展开的一条采样日志实际为 Discordbot，不应误认作 Telegram 请求。安全日志为自适应采样，不能从未命中的采样行推断 Telegram 抓取没有发生。
- 当前更准确结论：根页面与 OG 图片可达、真实 TelegramBot 已经成功取回，当前缺卡片的故障在 Telegram 预览解析/缓存或客户端展示链路；具体环节尚未证实。可用 query URL 的既有预览不代表根 URL 卡片缓存会自动刷新。请求分析为 24 小时聚合，没有单条请求时间，不能与某条聊天消息一一精确关联。

### 7.19 定制 Telegram OG 分享封面（2026-10-07）

- 新封面参考首页 Banner 的浅薄荷绿、深色标题、稳定高效/长期服务与 AI 图形，但沿用 OG 的 1200×630 深蓝绿视觉。右侧为定制的“用户自己的账号 → 官方订阅”卡片插画，以减少与其它 Dujiao-Next 模板站点的通用封面相似。
- 更新源文件 `frontend/user/public/og-preview.svg`、`og-preview.png`，新增 `og-preview-v2.png`；`frontend/user/index.html` 的 Open Graph / Twitter image 改为 `https://ai.ooxxoo.cn/og-preview-v2.png`，更新 alt。用户构建 `npm --prefix frontend/user run build` 成功。
- 为避免顺带发布工作区中既有的其它本地改动，只将构建产物中的 `index.html` 和两张 PNG 同步至生产嵌入资源目录；生产 Go 二进制重建并重启服务。备份 `/var/backups/dujiao-next/og-preview-refresh-20261007T235000`。
- 已核对 `dujiao-next.service` active、`/health` 正常、本机及公网首页均指向 v2 URL，公网 `/og-preview-v2.png` 返回 `200 image/png`（767629 字节）。用户需在 Telegram 用 `@WebpageBot` 更新根 URL 预览并在目标聊天重发，旧消息缓存不会保证自动换图。

### 7.20 代理推广二维码与下载海报改造（2026-10-07）

- `frontend/user/src/views/AgentCenter.vue`：弹窗二维码中心加入圆形动态站点图标，图标取自站点配置 `brand.site_icon`，错误时回退到品牌字母；二维码用 H 级纠错及较宽静区。
- 下载按钮改为“保存推广海报”，导出 1080×1600 PNG，包含站点名称、图标、专属二维码/代理代码、四个套餐价格，以及“用户自己的账号 · 官方直充 / 无需提供账号密码 · 保护隐私”。使用深蓝绿抬头、浅薄荷色内容、价格卡和隐私说明区。
- 通过 `api.get('/public/products', { headers: { 'X-Agent-Code': currentAgentCode } })` 拉取扫码落地页同源公开价格，按 product slug 对应 Plus 1个月、Pro 5x/20x/50x；读不到任一价格则禁用下载并要求重试，未保存的输入草稿不会混入海报。
- 生产端 AgentCenter 原源码除本次改造外与本地一致；以服务器 `frontend/user` 源码为临时构建基线，仅叠加二维码组件、当前站点 index OG 元数据及封面图。该临时基线 `vue-tsc -b && vite build` 通过（临时排除两份已过期且未引用的顶层 `src/client.ts` / `src/affiliate.ts`）。同步构建产物后生产重建 Go embed 服务。
- 备份 `/var/backups/dujiao-next/og-preview-refresh-20261007T235000`，含服务二进制、旧入口/分享图、完整旧 `dist/user`。验证服务 active，`/health` 正常；公网页引用新版 `index-C9BOtC2Z.js`，并映射到新版 AgentCenter chunk；对应 JS 与 OG 图片均 HTTP 200。
- 代理登录后在 `/agent` 强制刷新一次（Mac Chrome `⌘+Shift+R`，Windows Chrome `Ctrl+Shift+R`），再点「推广二维码」。

### 7.21 推广海报站点图标按红框放大（2026-10-07）

- 按用户截图调整 `frontend/user/src/views/AgentCenter.vue`：顶部品牌圆章从直径 112px 放大为 148px，二维码中心圆形图标从二维码画布宽度的 19% 放大至 27%，用于贴合设计预留圆形；二维码主体和海报布局、内容保持不变。
- 生产基线构建通过（`vue-tsc -b && vite build`），部署前备份线上二进制和静态资源：`/var/backups/dujiao-next/agent-qr-icon-size-20261006T235815Z`。完成静态文件同步、Go embed 重建和服务重启。
- 线上验证：`dujiao-next.service` active，源站/公网 `/health` 返回正常；新的 AgentCenter chunk `AgentCenter-COeH0uba.js` 公网 HTTP 200。
- 用户需刷新 `/agent` 并重新保存海报；旧 PNG 是静态文件，不会自行更新。

### 7.22 浏览器标签页 favicon 圆形裁切（2026-10-07）

- 根因：前台将 `brand.site_icon` 原图 URL 直接放入 favicon link，未做圆形蒙版。
- 修改 `frontend/user/src/stores/app.ts`：监听配置站点图标 URL，加载后居中取正方形，在 128×128 canvas 上裁为圆形 PNG data URL，并响应 favicon 配置更新。跨域 canvas 或加载失败时保留原图作为回退。
- 生产源基线构建通过（`vue-tsc -b && vite build`），备份 `/var/backups/dujiao-next/circular-favicon-20261007T000043Z`，已同步 dist、重建 Go embed 二进制并重启。
- 验证线上服务 active、`/health` 正常，生产 index 引用 `index-D6g-m8CK.js`，其中包含 canvas 圆形裁切逻辑。
- 观察 favicon 需强制刷新页面；浏览器可能继续显示之前缓存的 favicon，重新加载页面后稍等即可更新。

### 7.23 代理二维码图标和海报易读性优化（2026-10-07）

- 修改 `frontend/user/src/views/AgentCenter.vue`：站点图标按中心正方形裁切并圆形遮罩，图像铺满预留圆形约 91%，避免方形背景/边角露出；保留少量白色边缘。海报顶部白圆从直径 148px 缩小到 112px，二维码中间图标仍使用二维码宽度 27%。
- 增大海报品牌、副标题、二维码说明、代理码、套餐名/价格、隐私说明和页脚字号/字重，让手机上查看更清楚；其它内容和海报版式维持既有结构。
- 生产源代码基线构建成功（`vue-tsc -b && vite build`）。部署前备份 `/var/backups/dujiao-next/agent-poster-icons-20261007T000734Z`，部署 dist 并重建/重启 Go embed 服务。
- 验证服务 active、源站健康接口正常；新入口 `index-myY-4qV3.js` 与 AgentCenter chunk `AgentCenter-DssfdNAc.js` 公网正常（200）。刷新代理中心重新打开二维码弹窗；若要海报更新，需重新下载生成新 PNG。

### 7.24 海报套餐价格卡左右分栏（2026-10-07）

- 按用户要求改造 `frontend/user/src/views/AgentCenter.vue` 的四个套餐卡：左侧两行分别展示产品名、期限（如“ChatGPT Plus”/“1个月”）；右侧大号展示售价，小号“/月”作为单位。价格仍来自对应代理链接的公开商品接口。
- 卡片尺寸与四卡网格保持，调整文字字号及位置。`vue-tsc -b && vite build` 通过。
- 备份 `/var/backups/dujiao-next/agent-pricing-cards-20261007T001338Z`，同步前端构建产物并重建/重启 Go embed 服务。
- 验证服务 active、健康接口正常；生产入口 `index-H0YLSzCL.js` 引用 AgentCenter chunk `AgentCenter-B4UOZKCt.js`，新 chunk 公网 HTTP 200 且包含“/月”逻辑。
- 海报为客户端下载的静态 PNG，需重新生成下载才能看到本次布局。

### 7.25 推广海报文件名采用用户名+Agent+推广码（2026-10-07）

- 更新 `frontend/user/src/views/AgentCenter.vue` 的下载逻辑，文件名现为 `{用户名}Agent{推广码}.png`。优先昵称，其次 username，再其次邮箱 @ 前缀，最后回退 `user`；非法文件名字符转成下划线，代理码保留安全字符。
- 生产基线 `vue-tsc -b && vite build` 成功；部署前备份 `/var/backups/dujiao-next/agent-poster-filename-20261007T001539Z`，已同步用户端 dist、重建并重启服务。
- 验证服务 active、健康检查正常；入口 `index-By-Dv-tt.js` 指向新 AgentCenter chunk `AgentCenter-MQaoCGEX.js`。重新保存海报后使用新文件名。

### 7.26 核对手动卡台开卡/充值及 Telegram 播报（2026-10-07）

- 只读查看用户当前的 ZovoCard 卡详情页，记录到订单备注 `order#357076`，卡台显示时区 `-04` 的开卡时间为 2026-10-07 14:10:31。一次 OpenAI 消费 15.72 美元状态为“清算中”，卡台显示失败数 0、该金额为冻结额；不能称作已最终清算。
- Zovo CDK Portal `/ops/webhooks` 明确显示“暂无回调”。生产 `cdk_recharge.db` 只读计数：webhook_events / cardplatform_cdk_codes / recharge_tasks / telegram_sale_notifications 均为 0。没有 webhook 到达，也没有发送群播报。
- ZovoCard Developer 页通用 Webhook URL 输入框仍是 `https://你的服务/callback` 默认占位。CDK Portal 展示的本站回调地址为 `https://cdkadmin.ooxxoo.cn/api/v1/webhooks/cardplatform`。当前需在卡台设置正确 callback，并核对 `gpt_direct.completed` 事件订阅；具体断点倾向于卡台未配置/订阅。
- 未编辑卡台/Portal 设置，未重放事件或发消息。向卡台设置回调需要传输 Portal 的 Webhook Secret（签名密钥），需用户明确授权后再执行。只读检查后已将浏览器返回用户原来的卡详情页。

### 7.27 后续进展：配置 ZovoCard API Key 专属成交 Webhook（2026-10-07）

- 用户随后回复“继续”授权完成配置。在 ZovoCard 对应 API Key 的 Webhook 设置中填入本站回调地址 `https://cdkadmin.ooxxoo.cn/api/v1/webhooks/cardplatform`，仅订阅 `gpt_direct.completed`。
- API Key 专属签名密钥已填入 CDK Portal「Webhook Secret」并保存。后台以掩码显示已配置状态；密钥值不记入文档。未改 Bot Token、群组 ID 或旧版通用 Webhook。
- 保存后卡台配置显示正确回调 URL 和成功事件订阅；Portal 页面显示密钥已配置。当前近期事件仍为空，因为尚无新的成功事件。
- 先前手动充值不会自动回放。待下一笔真实成功事件最终核验播报、商品/邀请码映射、北京时间及订单去重；本轮未发测试/成交消息，也未创建交易。

### 7.24 前台 AI 客服部署（2026-10-08）

- 用户端新增可拖动圆形站点 icon 客服入口、轻微浮动/星形旋转动画、滑动对话面板和页面内会话保留；对话由服务端转发至 OpenAI-compatible `/chat/completions`。管理端新增客服配置页（启停、HTTPS API Base URL、模型、只写 API Key、欢迎语、业务知识、连接测试）。
- 后端对聊天/配置 body 限长、聊天消息长度和上下文限长；接口限速；凭证 AES-GCM 加密保存在 settings，后台只返回掩码。管理端客服配置 API 只授权 system_admin。模型系统指令限制为本站面向客户公开业务知识及已鉴权用户本人数据，不提供数据库/后台工具。
- 登录用户提出本人账户/订单/余额问题时，后端以 JWT 中间件注入的本人 user_id 查询本人的基础资料、钱包余额和最多 8 笔本人订单，按 reseller tenant scope 隔离；游客不读用户信息。本人数据只按需发送。公开上下文包括公开商品前 50 项、文章/公告前 20 条和当前白名单公开页面文字；账户、订单、支付、钱包、代理中心页面不会抽取当前页面内容。
- 用户端和管理端从线上源码作隔离构建基线，只叠加 AI 客服相关文件；用户端构建时仅在临时副本排除两份引用缺失模块的过期顶层 `src/client.ts`、`src/affiliate.ts`，线上源文件未动。构建通过。首轮部署校验发现同步路径丢失目录结构，客服路由没有进入二进制；修正源文件层级后重新构建和部署，最终核对成功。
- 发布前备份 `/var/backups/dujiao-next/ai-assistant-20261008T150600Z`：旧二进制、生产配置、SQLite 在线备份（完整性检查 `ok`）、旧用户端/后台静态资源及路由/权限文件，root 私有权限。没有修改数据库内容或支付数据。
- 最终验证：`dujiao-next.service` active，`/health` 返回 `{"status":"ok"}`；`/api/v1/public/ai-assistant/config` 返回成功且 `enabled:false`；用户端主 bundle、管理端主 bundle、`AIAssistantSettings` 分包及 `/admin/ai-assistant` 均返回 HTTP 200。
- **客服代码已线上部署，但尚未配置模型服务密钥/启用**。登录管理后台进入“系统设置 → 智能客服配置”，填写 API Base URL、模型名和 API Key，保存并连接测试通过后再启用。API Key 不要经聊天发送。API Key 加密密钥从生产 JWT secret 派生，JWT secret 轮换后需要重新保存 API Key。尚未运行自动化测试。

### 7.28 更正：AI 客服误加到本项目，现已回滚（2026-10-08）

- 用户确认 AI 客服需求属于另一项目，不属于 AI 代充值的 `dujiao-next`。此前 7.24 的客服部署是错误项目上的操作，已按用户要求回滚。
- 生产已恢复部署前二进制（SHA256 与备份一致）、旧前台/后台 dist、旧后台和前台路由及授权文件。服务 active，`/health` 返回 `{"status":"ok"}`；首页和后台 HTTP 200；`/api/v1/public/ai-assistant/config` 返回 404，表示该功能已撤下。
- 部署前备份位于 `/var/backups/dujiao-next/ai-assistant-20261008T150600Z`；错误版本快照位于 `/var/backups/dujiao-next/wrong-project-ai-assistant-20261008T151048Z`。未修改生产数据库/业务配置，也未录入模型 API Key。
- 本地 `dujiao-next` 中客服模块、前台客服组件/API、后台客服配置页面及相关路由/授权/菜单接线已清理；清理时保留了工作区中与本次客服无关的既有未提交更改。
- 本文 7.24 节保留为历史审计记录，不代表当前部署状态。当前 AI 代充值项目不含 AI 客服。若继续开发，先切换到用户指定的正确项目仓库。
### 7.29 代理订单分类、邮箱查单与 Session 入口修复（2026-10-10）

- 已根据生产订单 `DJ20261010145807741035` 快照核实：114 元订单为 `agent_wholesale` 拿货价，代理账号关联的联系邮箱是 `jamesbond235813@gmail.com`，订单支付成功且仍在 `fulfilling`。
- 后端代理订单摘要增加 `is_wholesale`，由订单价格快照的 `pricing_mode` 判定。代理中心将此类订单显示在“代理拿货订单”，客户订单只显示客户自付；拿货记录会补入代理拿货列表并按订单号去重。
- `fulfilling` 订单卡片增加“继续提交 Session”入口，指向既有订单 Session 页面。游客查单查询改为按联系邮箱匹配游客订单及已登录账号订单；遵循用户选择的仅凭邮箱查询，既有租户范围约束不变。
- 本地 `go build ./...`、前端 `pnpm run build` 均通过；以线上 AgentCenter 源码为前端构建基线，仅叠加本节订单功能，保留线上已有手机布局改动。线上没有 Node/pnpm 时，在本机对线上前端源码快照构建，再将构建资源同步回服务器。
- 曾发布生产并验证：后端全栈二进制重建、前台资源替换，`/health` 返回正常，订单邮箱无代理代码可查到目标订单。
- 用户要求终止升级后，已完整回滚：恢复发布前二进制（SHA256 与备份相同）、原始前台资源和本次改动涉及的五个源码文件；重启后的服务 active，`/health` 返回 `{"status":"ok"}`。因此本节修复当前未在线上启用。
- 回滚备份：`/var/backups/dujiao-next/agent-order-fixes-20261010T075126Z`，包含发布前二进制、原始前台资源、源码及数据库快照；数据库未被本次升级修改。升级期间生成的新静态资源另存于 `internal/web/dist/user-aborted-*`。
