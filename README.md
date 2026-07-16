# Go SQLite ERP

一个面向中小型办公设备、维修与工程服务场景的轻量 ERP。后端使用 Go、Gin、GORM 和 SQLite，前端使用 Vue 3、Vite、Pinia 与 Element Plus。

> 项目当前适合作为演示、二次开发和内部部署基础。正式用于生产环境前，请完成权限、安全、备份和业务规则审计。

## 主要功能

- JWT 登录、刷新令牌、修改密码和强制改密
- 用户、角色、权限、菜单和操作审计
- 客户、供应商和商品资料管理
- 采购、销售、库存批次与库存流水
- 维修、工程、签到和业务图片
- 应收、应付、资金账户、客户对账单和利润报表
- 老板驾驶舱、价格利润分析和移动端页面
- Excel 导入导出、单据打印和 PWA 基础支持
- SQLite 单文件数据库，无需单独安装数据库服务

## 技术栈

| 部分 | 技术 |
| --- | --- |
| 后端 | Go 1.24、Gin、GORM、SQLite、Zap、JWT |
| 前端 | Vue 3、Vite 8、Pinia、Element Plus、Axios、ECharts |
| 数据 | SQLite，默认文件为 `data/erp.db` |
| 接口 | REST API，统一前缀 `/api/v1` |

## 环境要求

- Go 1.24 或兼容 `go.mod` 中 toolchain 配置的 Go 版本
- Node.js 20.19+，推荐 Node.js 22.12+ LTS
- npm 10+
- Git

## 本地运行

### 1. 创建本地配置

Windows PowerShell：

```powershell
Copy-Item config.example.yaml config.yaml
Copy-Item web/.env.example web/.env.local
```

Linux 或 macOS：

```bash
cp config.example.yaml config.yaml
cp web/.env.example web/.env.local
```

首次启动前至少修改 `config.yaml` 中的 `jwt.secret` 和 `admin.password`。`config.yaml`、数据库、日志和本地环境变量均已加入 `.gitignore`，不要强制提交这些文件。

### 2. 启动后端

```bash
go mod download
go run ./cmd/server
```

后端默认监听 [http://127.0.0.1:18080](http://127.0.0.1:18080)，健康检查地址为 [http://127.0.0.1:18080/healthz](http://127.0.0.1:18080/healthz)。首次启动会自动创建数据库和基础数据。

### 3. 启动前端

在另一个终端中运行：

```bash
cd web
npm ci
npm run dev
```

浏览器打开 [http://127.0.0.1:5173/login](http://127.0.0.1:5173/login)。开发服务器会将 `/api` 请求代理到本地后端。

### 4. 登录演示账号

示例配置默认使用：

```text
用户名：admin
密码：admin888
```

该账号只用于本地演示。首次登录后应立即修改密码，公网部署前必须替换示例密码和 JWT 密钥。

## 配置说明

配置模板见 [`config.example.yaml`](config.example.yaml)：

| 配置项 | 作用 | 默认值 |
| --- | --- | --- |
| `server.host` | 后端监听地址 | `127.0.0.1` |
| `server.port` | 后端端口 | `18080` |
| `database.dsn` | SQLite 文件和连接参数 | `data/erp.db?...` |
| `jwt.secret` | JWT 签名密钥 | 仅示例占位值 |
| `admin.*` | 首次初始化的管理员信息 | 演示账号 |
| `app.demo_mode` | 限制危险写操作的演示模式 | `true` |

前端配置模板见 [`web/.env.example`](web/.env.example)。同域部署时保持 `VITE_API_BASE_URL=/api/v1` 即可。

## 测试与构建

```bash
go test ./...
go vet ./...
cd web
npm ci
npm run build
```

GitHub Actions 会对每次推送和 Pull Request 执行相同的后端测试、静态检查与前端构建。

## 项目结构

```text
cmd/                    服务入口、种子数据和流程验收命令
internal/               后端领域、应用、基础设施和 HTTP 层
migrations/             SQLite 增量迁移脚本
web/                    Vue 3 前端
tools/frontend-server/  生产构建的静态文件服务与 API 代理
packaging/              Windows 和 Linux 运行包脚本
docs/                   API 与项目文档
```

更详细的文件职责见 [`docs/PROJECT_FILES.md`](docs/PROJECT_FILES.md)，采购、销售、维修、财务、对账和删除反处理规则见 [`docs/BUSINESS_LOGIC.md`](docs/BUSINESS_LOGIC.md)。

## 部署

- Windows 运行包说明：[`packaging/runtime/README.txt`](packaging/runtime/README.txt)
- Linux 运行包说明：[`packaging/linux/README-LINUX.txt`](packaging/linux/README-LINUX.txt)
- 首次上传 GitHub：[`docs/GITHUB_UPLOAD.md`](docs/GITHUB_UPLOAD.md)

源码仓库不包含已经构建好的运行包。部署前需要分别构建后端、前端和 `tools/frontend-server`，或通过自己的发布流水线生成制品。

## 安全提示

- 不要提交 `config.yaml`、`.env`、数据库、上传文件、日志或令牌
- 公网部署应使用 HTTPS 和反向代理，并只开放必要端口
- 当前 CORS 默认仅允许本机开发地址；域名部署需按实际来源修改
- 定期备份 `data/`，并验证备份能够恢复
- 漏洞报告请参阅 [`SECURITY.md`](SECURITY.md)，不要在公开 Issue 中披露敏感细节

## 参与贡献

提交代码前请阅读 [`CONTRIBUTING.md`](CONTRIBUTING.md)。Issue 和 Pull Request 模板已放在 `.github/` 目录。

## License

本项目使用 [MIT License](LICENSE)。
