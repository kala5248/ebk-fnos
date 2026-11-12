# ezBookkeeping · 飞牛 fnOS 定制版（ebk-fnos）

[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Upstream](https://img.shields.io/badge/upstream-mayswind%2Fezbookkeeping-blue.svg)](https://github.com/mayswind/ezbookkeeping)
[![Version](https://img.shields.io/badge/version-2.1.1--fnos-orange.svg)](https://github.com/kala5248/ebk-fnos/releases)

本仓库是 [mayswind/ezbookkeeping](https://github.com/mayswind/ezbookkeeping) 的定制分支，在上游 **v2.1.1** 发布基线上新增四块能力：**飞牛 fnOS 授权登录**、**AI 一句话 / 图片记账**、**每用户独立的大模型配置**、**多用户共享账本**。上游原有功能（记账、导入导出、图表、多币种等）全部保留。

- 上游英文 README：[README.upstream.md](./README.upstream.md)
- 改造说明（配置项、接口表、角色权限、部署与限制）：[docs/fnos-ai-shared-ledger.md](./docs/fnos-ai-shared-ledger.md)
- 本分支不向上游提交 PR，仅作为个人 / 家庭自托管发行版维护。

## 相对上游的差异

| 模块 | 上游 | 本分支 |
|---|---|---|
| 登录方式 | 账号密码、OIDC、2FA | 额外支持 **飞牛 fnOS 授权登录**（PKCE，无需 redirect_uri，粘贴一次性 code 完成登录 / 绑定） |
| AI 识别配置 | 全实例共用一组 INI 配置 | 支持 **每个用户配置自己的 Provider / Base URL / API Key / 文本模型 / 视觉模型**，API Key 加密入库，未配置时自动回退全局配置 |
| AI 记账 | 文本 / 小票图片识别 | 保留并贯通到 **桌面与移动端**，识别结果回填记账表单，用户确认后才落库 |
| 账本模型 | 单用户私有账本 | 新增 **共享账本**：邀请码加入、成员角色（读写 / 只读）、影子用户隔离数据、一键切换当前账本 |

## 功能

### 1. 飞牛 fnOS 授权登录

- 登录页提供「Log in with fnOS」；桌面端弹窗、移动端走系统浏览器打开授权页。
- 采用 PKCE（`code_challenge` / `code_verifier`，S256），**不依赖 redirect_uri**：授权页复制一次性 code，粘贴回应用完成登录。
- 已登录用户可在设置页把飞牛账号绑定到当前账号；首次登录按 `fnos-<uid>@ezbookkeeping.local` 查找绑定，未注册且允许注册时自动建号。
- 开启 2FA 的账号拒绝通过飞牛登录绕过二次验证。

配置项（`conf/ezbookkeeping.ini`）：

```ini
enable_fnos_auth = true
fnos_base_url = http://192.168.1.10:5666   # 飞牛 Web UI 地址
fnos_client_id =                            # 飞牛侧注册的 OAuth client id
fnos_client_secret =                        # 公共客户端（PKCE）可留空
fnos_scope = trim.user.all
fnos_auto_register = true
```

> `fnos_client_id` 需在飞牛侧申请注册，仓库不内置真实值。

### 2. AI 一句话 / 图片记账

- 一句话记账：输入「今天中午吃饭花了 20」「昨天晚上打车花了 30」「20 号发工资 5000」，模型抽取 `类型 / 时间 / 金额 / 账户 / 分类 / 标签 / 备注`，服务端再按当前账本的账户与分类字典做**名称 → ID** 映射。
- 图片记账：上传小票图片，交给视觉模型识别，解析流程与文本一致。
- **AI 只做识别，不直接写库**：结果回填记账表单，用户确认或修改后走原有 `transactions/add` 接口，金额精度、余额、权限校验全部沿用原逻辑。
- 入口：桌面交易列表「AI 文本识别 / AI 图片识别」、记账弹窗内识别；移动端首页「+」长按菜单、记账页内识别与剪贴板识别。

### 3. 每用户模型配置

设置页 **Settings → AI & SHARING → AI Accounting Model**（移动端：设置 → AI Accounting Model）：

- 可填 Provider、Base URL、API Key、文本模型 ID、视觉模型 ID，分别开关「文本识别」「图片识别」，并提供 Test / Remove。
- API Key 使用实例 `secret_key` 加密后入库，接口只返回掩码（`*` + 末 4 位）；留空保存表示沿用旧值。
- 解析优先级：**当前登录用户配置 → 全局 INI 配置**；共享账本下始终使用登录用户自己的 Key。

配置项：

```ini
enable_user_llm_config = true
```

### 4. 共享账本

- 创建账本自动生成 **8 位邀请码**，其他用户凭邀请码加入；成员分为**读写 / 只读**。
- 每个共享账本对应一个无密码影子用户（`ledger_<ledgerId>`），成员以影子身份读写数据，**个人账本数据天然隔离**。
- 只读成员的非 `GET/HEAD/OPTIONS` 请求直接 403（错误码 `222012`）。
- 切换账本：桌面顶栏账本切换器；移动端首页导航栏账本按钮 + 动作面板。
- 身份类接口（`/tokens`、`/users`、`/ledgers`、`/llm/user_config`、`/fnos`）不参与账本作用域切换。

## 快速开始

### 从源码构建

需要 Golang（`go.mod` 要求 1.27.x，首次编译会自动下载 toolchain）、GCC、Node.js、NPM：

```bash
./build.sh package -o ebk-fnos.tar.gz
```

构建产物版本取自 `package.json`，本分支为 `2.1.1-fnos`。

### 运行

```bash
./ezbookkeeping server run --conf-path conf/ezbookkeeping.ini
```

默认监听 `8080`，浏览器访问 `http://<host>:8080/`。首次部署需在 `conf/ezbookkeeping.ini` 中至少确认：`db_path`、`local_filesystem_path`、`static_root_path`（需绝对路径）、`secret_key`。

> 注意：二进制需在源码根目录（或包含 `templates/prompt/*.tmpl` 的发行包目录）下启动，否则 AI 识别会因找不到提示词模板而失败。

### 打包成飞牛应用（.fpk）

一条命令完成「编译后端 + 构建前端 + 组装 fpk 目录 + 打包」：

```bash
./build-fnos-fpk.sh
# 产物：build/ebk-fnos-2.1.1-fnos-x86.fpk
```

可选参数：

| 参数 | 说明 |
|---|---|
| `--skip-frontend` | 复用已有 `dist/`，CI 或二次构建时加快速度 |
| `--no-pack` | 只组装 `build/fnos/` 目录，不生成 `.fpk` |

打包依赖 Go、GCC、Node.js/npm；若安装了官方 `fnpack` 工具则用它产出标准 `.fpk`，否则回退为 `tar.gz` 兼容格式（文件名仍为 `.fpk`）。前端构建建议至少 8 GB 内存。

**安装向导**（`fnos/wizard/install`）在飞牛应用中心手动安装时弹出两步：

1. **服务设置**：服务端口，默认 `8080`，可改（改完会自动同步桌面入口端口）。
2. **数据库设置**：内置 SQLite（默认）/ MySQL / PostgreSQL，选择外部数据库时填写 host:port、库名、用户名、密码。
3. **飞牛授权登录（可选）**：开启后登录页出现「使用飞牛账号登录」，需填写 fnOS 地址（默认 `http://127.0.0.1:5666`）与在 fnOS 注册的 OAuth 客户端 ID / 密钥；未填客户端 ID 时自动关闭，不影响账号密码登录。

安装后由 `cmd/install_callback` 生成 `$TRIM_PKGETC/ezbookkeeping.ini`：SQLite 数据库落在 `$TRIM_PKGVAR/data/`，附件落在 `$TRIM_PKGVAR/files/`，日志落在 `$TRIM_PKGVAR/log/`，`secret_key` 首次安装随机生成。安装完成后可在应用中心「设置」中通过配置向导（`fnos/wizard/config`）改端口或数据库，应用会自动重启。

## 新增接口一览

| 接口 | 说明 |
|---|---|
| `GET /api/fnos/auth_url.json` | 获取飞牛授权地址（含 PKCE 参数与 state） |
| `POST /api/fnos/login.json` | 用授权码登录，返回 token |
| `GET /api/v1/fnos/link_url.json` / `POST /api/v1/fnos/link.json` | 绑定飞牛账号 |
| `POST /api/v1/llm/transactions/recognize_text.json` | 一句话记账识别（`{ "text": "..." }`） |
| `POST /api/v1/llm/transactions/recognize_receipt_image.json` | 小票图片识别（multipart `image`） |
| `GET/POST/DELETE /api/v1/llm/user_config/{get,save,delete,test}.json` | 用户模型配置读写与连通性测试 |
| `GET/POST /api/v1/ledgers/{list,create,join,rename,invoice/regenerate}.json` | 账本列表 / 创建 / 加入 / 重命名 / 邀请码重置 |
| `GET/POST /api/v1/ledgers/{members,members/role,members/remove,leave,delete}.json` | 成员管理、角色调整、退出与删除 |

账本相关请求通过 `X-Ledger-Id` 头指定当前账本；前端已在 axios 拦截器中统一注入。

## 实现要点

| 文件 | 作用 |
|---|---|
| `pkg/auth/fnos/fnos_authentication.go` | PKCE 生成、授权 URL、授权码换令牌 |
| `pkg/api/fnos_authentications.go` | state 表（10 分钟 TTL）、登录 / 绑定、自动建号 |
| `pkg/services/user_llm_configs.go` | 用户配置加密存储、掩码、`ResolveLLMConfig` 回退 |
| `pkg/api/large_language_models.go` | 文本 / 图片识别入口、名称→ID 映射 |
| `pkg/middlewares/ledger_scope.go` | `X-Ledger-Id` → 影子 uid、只读拦截、身份路径豁免 |
| `pkg/services/ledgers.go` | 创建账本、影子用户、复制分类树、邀请码、成员管理 |
| `src/components/desktop/LedgerSwitcher.vue` | 桌面端账本切换 |
| `src/views/{desktop,mobile}/settings/AiModelSettingsPage.vue` | 用户模型配置页 |
| `src/views/{desktop,mobile}/settings/SharedLedgerPage.vue` | 共享账本管理与成员页 |

## 测试情况

- 后端接口冒烟：15 项全部通过（授权链接、建账本、邀请码加入、成员角色、只读 403、数据隔离、密钥掩码、全局关闭时仍走用户配置、伪造 code 拒绝等）。
- Playwright 界面测试：桌面端核心链路（飞牛入口、设置导航、模型保存、UI 建账本）通过；移动端登录与设置入口通过，其余用例受沙箱内存限制（Vite dev server OOM）未跑完，待本地内存充足环境补测。

## 许可

沿用上游 MIT License，见 [LICENSE](LICENSE)。上游项目与版权归 [mayswind/ezbookkeeping](https://github.com/mayswind/ezbookkeeping) 所有。
