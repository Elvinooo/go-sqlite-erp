# 项目结构说明

本文档描述仓库中主要目录和文件的职责。运行时依赖、本地配置、数据库、日志和构建产物不属于源码仓库。

## 总体架构

```text
cmd/                    可执行程序入口
internal/
  application/          应用服务和请求/响应 DTO
  bootstrap/            配置、数据库、仓储、服务和路由装配
  config/               YAML 配置加载
  domain/               领域模型和仓储接口
  infrastructure/       数据库、日志、JWT 和 GORM 仓储实现
  interfaces/http/      HTTP 控制器、中间件和路由
  shared/               通用错误、分页、响应和工具
migrations/             SQLite 增量迁移
tools/frontend-server/  前端静态服务及 /api 反向代理
web/                    Vue 3 单页应用
packaging/              Windows 和 Linux 运行包脚本
docs/                   项目与 API 文档
.github/                CI、Dependabot、Issue 和 PR 模板
```

后端主要调用链：

```text
cmd/server
  -> internal/bootstrap
  -> interfaces/http/router
  -> controller
  -> application service
  -> domain repository interface
  -> infrastructure/persistence/gorm
  -> SQLite
```

前端主要调用链：

```text
web/src/main.js
  -> router / stores / views
  -> api modules
  -> api/request.js
  -> /api/v1
```

## 根目录文件

| 文件 | 用途 |
| --- | --- |
| `README.md` | 项目介绍、本地运行、测试、部署和安全入口 |
| `CONTRIBUTING.md` | 开发、测试和 Pull Request 规范 |
| `SECURITY.md` | 漏洞报告与生产部署安全要求 |
| `LICENSE` | MIT 开源许可证 |
| `config.example.yaml` | 可公开提交的后端配置模板 |
| `go.mod` / `go.sum` | Go 模块、工具链与依赖锁定 |
| `.gitignore` | 排除本地配置、数据、日志、依赖和构建产物 |
| `.gitattributes` | 统一跨平台文本换行规则 |

业务流程、状态、金额公式及数据副作用集中记录在 `docs/BUSINESS_LOGIC.md`。

## 后端入口

| 路径 | 用途 |
| --- | --- |
| `cmd/server` | 启动 HTTP API 服务 |
| `cmd/seed-cn` | 初始化中文演示数据 |
| `cmd/seed-statement` | 初始化客户对账示例数据 |
| `cmd/flow-test` | 执行业务主流程验收 |
| `cmd/finance-flow-test` | 执行财务流程验收 |

`internal/bootstrap/app.go` 是后端装配入口，负责读取 `config.yaml`、初始化日志和 SQLite、执行自动迁移、创建仓储与服务，并注册 Gin 路由。

## 后端分层

| 目录 | 职责 |
| --- | --- |
| `internal/domain` | 用户、客户、供应商、业务、看板和系统设置模型，以及仓储接口 |
| `internal/application` | 认证、客户、供应商、业务、看板和系统设置的业务编排 |
| `internal/infrastructure/database` | SQLite 连接、自动迁移和基础数据初始化 |
| `internal/infrastructure/persistence/gorm` | 仓储接口的 GORM 实现及核心业务测试 |
| `internal/infrastructure/security` | JWT 生成和校验 |
| `internal/interfaces/http/controller` | HTTP 参数绑定、调用服务和统一响应 |
| `internal/interfaces/http/middleware` | JWT、RBAC、数据范围、审计、CORS、恢复和请求日志 |
| `internal/interfaces/http/router` | 健康检查、认证和受保护 API 路由注册 |
| `internal/shared` | 共享错误、模型、分页、响应和密码工具 |

## 前端结构

| 路径 | 职责 |
| --- | --- |
| `web/src/main.js` | Vue 应用入口，注册 Pinia、路由和 Element Plus |
| `web/src/App.vue` | PC/移动端应用外壳、导航和用户操作 |
| `web/src/router` | 页面路由和登录守卫 |
| `web/src/stores` | 用户认证和应用状态 |
| `web/src/api` | Axios 客户端与各业务 API 封装 |
| `web/src/views` | 登录、看板、动态业务、客户、供应商和系统页面 |
| `web/src/composables` | 可复用查询和分页逻辑 |
| `web/src/utils` | 日期、移动端、打印、扫码和状态显示工具 |
| `web/public` | PWA manifest、图标和 Service Worker |

## 数据与迁移

- `data/` 保存 SQLite 数据库和上传文件，仅属于本地或部署环境，不提交到 Git。
- `migrations/` 保存可审查的增量 SQL；已经应用的迁移不应直接改写。
- `internal/infrastructure/database/migrate.go` 负责 GORM 自动迁移、索引与基础数据初始化。
- `docs/swagger.json` 保存 API 描述；接口变化时应同步更新。
- `docs/BUSINESS_LOGIC.md` 保存采购、销售、维修、工程、财务、对账和删除反处理规则。

## 部署与仓库自动化

- `packaging/runtime/`：Windows 运行包的启动、停止、状态和开机启动脚本。
- `packaging/linux/`：Linux 运行包及 systemd 安装脚本。
- `.github/workflows/ci.yml`：后端测试、`go vet` 和前端生产构建。
- `.github/dependabot.yml`：Go、npm 和 GitHub Actions 依赖更新检查。
- `docs/GITHUB_UPLOAD.md`：首次创建远程仓库并推送的完整步骤。

## 不应提交的内容

以下内容由 `.gitignore` 排除：

- `config.yaml`、`.env`、`web/.env.local`
- `data/`、数据库文件和上传文件
- `logs/`、`.run/` 和临时文件
- `node_modules/`、`dist/`、`build/` 和运行包输出
- 本地 IDE 设置和操作系统缓存文件
