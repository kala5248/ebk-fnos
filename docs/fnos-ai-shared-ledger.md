# 飞牛 fnOS 适配 / 每用户 AI 模型 / 共享账本

本次改造在 ezBookkeeping 原有架构上新增三块能力，全部随主程序编译，无独立服务：

1. **飞牛（fnOS）授权登录**：用飞牛 NAS 账号登录 ezBookkeeping，支持首次自动建号、老账号绑定。
2. **每用户大模型配置**：AI 一句话记账、票据图片记账使用的模型由每个用户在设置页自行配置，服务端加密保存；未配置时回退到全局 `conf/ezbookkeeping.ini` 的 LLM 配置。
3. **共享账本**：多个用户共用同一个账本记账，账本内有所有者 / 编辑者 / 只读成员三种角色，通过邀请码加入。

## 一、飞牛授权登录

### 配置

```ini
[auth]
enable_fnos_auth = true
fnos_base_url = http://192.168.1.10:5666
fnos_client_id = <在飞牛侧注册/申请到的 client_id>
fnos_client_secret =
fnos_scope = trim.user.all
fnos_auto_register = true
fnos_request_timeout = 10000
fnos_skip_tls_verify = false
```

`fnos_base_url` 填飞牛 Web UI 的地址（协议 + 主机 + 端口，不带路径），服务端会以
`<fnos_base_url>/oauthapi/...` 访问授权接口。

> `fnos_client_id` 必须在飞牛侧注册获得。仓库里不内置任何真实 client_id（官方 trim-cli 使用的
> `YJNMPJUGA9` 属于飞牛官方客户端，不应复用）。

### 流程（与标准 OAuth2 回调不同）

飞牛第三方授权没有 redirect_uri，授权完成后由用户在页面上复制一次性 code，再粘贴回应用：

1. 登录页点击「使用飞牛账号登录」；
2. 前端调用 `GET /api/fnos/auth_url.json`，拿到授权链接与 `state`（服务端同时保存 PKCE verifier 与 state，10 分钟有效）；
3. 浏览器打开该授权链接（`/oauthapi/authorize`），用户在飞牛页面登录；
4. 用户复制页面显示的一次性 code，粘贴到应用弹窗；
5. 前端调用 `POST /api/fnos/login.json`（`{state, code}`），服务端用 PKCE verifier 调用
   `POST /oauthapi/third-part/token` 换取 token 与用户信息；
6. 若该飞牛账号（按 uid 映射，`ExternalEmail = fnos-<uid>@ezbookkeeping.local`）尚未绑定，
   在 `enable_register && fnos_auto_register` 时自动建号（用户名 `fnos_<飞牛用户名>`，邮箱已验证，币种 CNY）；
7. 返回 ezBookkeeping 自己的登录 token，后续请求与密码登录完全一致。

已登录用户可在设置页把当前账号与飞牛账号绑定：`GET /api/v1/fnos/link_url.json` +
`POST /api/v1/fnos/link.json`。

### 限制

- 已启用两步验证（2FA）的账号暂不支持飞牛登录，会提示改用密码登录（避免绕过 2FA）。
- 飞牛账号被禁用（`disabled`）或命中 OAuth2 登录限制时拒绝登录。

## 二、每用户 AI 模型配置

### 数据与加密

新增表 `user_llm_config`：

| 字段 | 说明 |
| --- | --- |
| `uid` | 用户，主键，即“每个用户一份配置” |
| `provider` | `openai` / `anthropic` / `googleai` / `ollama` / `lmstudio` |
| `base_url` | 自定义 API 地址，可留空使用默认 |
| `api_key` | 使用 `secret_key` 派生的密钥加密后存储 |
| `text_model_id` / `image_model_id` | 文本识别模型 / 票据图片识别模型 |
| `enable_text_recognition` / `enable_image_recognition` | 两个独立开关 |

接口返回的 `apiKeyMasked` 只显示末 4 位（`****abcd`），保存时 `apiKey` 传空表示沿用旧密钥。

### 接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/v1/llm/user_config/get.json` | 读取当前登录用户的配置（无则返回 `false`） |
| POST | `/api/v1/llm/user_config/save.json` | 保存配置 |
| POST | `/api/v1/llm/user_config/delete.json` | 删除配置 |
| POST | `/api/v1/llm/user_config/test.json` | 用已保存配置做一次真实调用测试 |

以上 4 个配置接口与 `recognize_text` / `recognize_receipt_image` 两个识别接口**始终注册**，不再依赖全局
LLM 配置是否启用；是否真正可用由 handler 在请求时按“用户配置优先、全局回退”判断。

### 解析顺序

文本记账 / 图片记账时：

1. 优先使用**当前登录用户**的配置（需对应开关打开且配置可用）；
2. 否则回退到全局 `[llm]` 配置（`text_recognition_llm_provider` 等）与全局开关。

因此即使全局关闭，用户自己配置后也能使用 AI 记账；前端入口开关由服务端下发的
`llmtr` / `llmir` 与本机缓存的用户配置共同决定。

## 三、共享账本

### 模型：账本 = 独立数据租户

ezBookkeeping 原有的所有财务查询都按 `uid` 隔离（`pkg/services/transactions.go` 等）。为了不改动
上百处 `uid` 调用，共享账本采用“影子用户”方案：

- 创建共享账本时，同时创建一个无密码的影子用户（`username = ledger_<ledgerId>`，
  邮箱 `ledger+<id>@ezbookkeeping.local`，`EmailVerified = true`），并把创建者的分类树复制过去；
- 该账本的所有账户、流水、标签等都记在影子用户的 `uid` 下，从而与个人账本彻底隔离；
- 请求进入时由中间件按 `X-Ledger-Id` 请求头解析出租户：把 token claims 里的 `Uid` 改写为
  影子用户 `uid`，原有 handler 完全无感；登录 uid 另存于 context，供身份类接口使用。

个人账本 `ledgerId = 0`，数据就是用户自己的 `uid`。

### 角色与权限

| 角色 | 值 | 权限 |
| --- | --- | --- |
| 所有者 Owner | 1 | 全部读写、改名、重置邀请码、调整成员角色、移除成员、删除账本 |
| 编辑者 Editor | 2 | 读写账本数据 |
| 只读 Viewer | 3 | 仅 GET / HEAD / OPTIONS（中间件直接拦截写请求） |

中间件注册在 token 鉴权之后（`cmd/webserver.go` 中 `apiV1Route.Use(middlewares.LedgerDataScope(config))`），
因此在同一个路由组内定义的所有接口都会生效。

身份类接口不切换租户（保持登录 uid）：`/api/v1/tokens/`、`/api/v1/users/`、`/api/v1/ledgers/`、
`/api/v1/llm/user_config/`、`/api/v1/fnos/`。

### 接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/v1/ledgers/list.json` | 我可见的账本（含个人账本） |
| POST | `/api/v1/ledgers/create.json` | 创建共享账本，返回邀请码 |
| POST | `/api/v1/ledgers/join.json` | 用邀请码加入 |
| POST | `/api/v1/ledgers/rename.json` | 重命名 |
| POST | `/api/v1/ledgers/invite/regenerate.json` | 重置邀请码 |
| GET | `/api/v1/ledgers/members.json?ledgerId=` | 成员列表 |
| POST | `/api/v1/ledgers/members/role.json` | 调整成员角色（Owner 才能操作） |
| POST | `/api/v1/ledgers/members/remove.json` | 移除成员（Owner） |
| POST | `/api/v1/ledgers/leave.json` | 退出账本 |
| POST | `/api/v1/ledgers/delete.json` | 删除账本（Owner，删除影子租户数据） |

### 前端

- 顶部导航栏账本切换器（`LedgerSwitcher.vue`）：存在多个账本时出现，切换后 `localStorage` 记录
  当前账本并通过 `X-Ledger-Id` 头随所有请求发送（`services.ts` 拦截器统一注入）；
- 设置页「共享账本」：创建 / 加入 / 切换 / 成员与角色 / 重置邀请码 / 退出 / 删除。

## 四、部署

```bash
# 后端
go build ./...

# 前端
npm install
npm run build

# 运行（示例）
./ezbookkeeping server --conf conf/ezbookkeeping.ini
```

首次启动会自动 `Sync2` 建表（`user_llm_config`、`ledger`、`ledger_member`）。

## 五、已知限制与后续可做

- 飞牛 `client_id` 需在飞牛侧注册，本仓库只提供可配置项。
- 2FA 账号暂不支持飞牛登录（见上）。
- 共享账本采用“账本=租户”隔离：切换账本后，原个人账本的统计/图表数据不会混入共享账本，
  这是设计目标；但也意味着共享账本无法把个人账本的流水“带进来”，需要重新记账或导入。
- 后续可做：飞牛 token 自动续期、共享账本内按成员拆分流水视图、AI 记账结果的人工确认队列。

## 前端页面

桌面端：顶部账本切换器、设置页「共享账本」「AI 记账模型」、登录页飞牛登录。

移动端（本次补齐）：

| 页面 | 路径 | 功能 |
| --- | --- | --- |
| 首页账本切换 | `/` 导航栏右侧按钮 | 列出全部账本，点击切换到对应账本（`localStorage` 记录，刷新生效） |
| 共享账本 | `/settings/shared_ledger` | 账本列表、当前账本标记、创建/加入、改名、重置邀请码、成员管理（改角色/移除）、退出/删除 |
| AI 记账模型 | `/settings/ai_model` | 供应商选择、Base URL、API Key（不回显）、文本/图片模型、两个开关、测试连接、删除配置 |
| 设置入口 | `/settings` | 新增「AI 记账模型」「共享账本」两项 |

移动端沿用 Framework7 的 `f7-actions` 动作面板与 `showPrompt/showConfirm/showToast`，不引入新的交互范式。
