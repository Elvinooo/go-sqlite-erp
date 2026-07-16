# 贡献指南

感谢参与 Go SQLite ERP。请让每次改动保持范围清晰、能够验证，并避免将真实业务数据或密钥带入仓库。

## 开发准备

1. Fork 仓库并从默认分支创建功能分支。
2. 按 README 创建 `config.yaml` 和 `web/.env.local`。
3. 后端执行 `go mod download`，前端在 `web/` 目录执行 `npm ci`。
4. 不要提交 `data/`、`logs/`、本地配置或构建产物。

建议分支名使用 `feature/简短描述`、`fix/简短描述` 或 `docs/简短描述`。

## 开发约定

- Go 代码提交前执行 `gofmt`，新增行为应补充相应测试。
- 前端保持现有 Vue 3 Composition API、Pinia 和 Element Plus 使用方式。
- 数据库结构变更应新增迁移文件，不要修改已经使用过的迁移。
- 金额计算继续使用 `decimal`，避免使用浮点数处理财务金额。
- API 或配置发生变化时，同步更新 README、Swagger 或相关部署说明。
- 日志、测试数据和截图中不得出现真实客户资料、密码、令牌或其他敏感信息。

## 提交前检查

```bash
gofmt -w ./cmd ./internal ./tools
go test ./...
go vet ./...
cd web
npm ci
npm run build
```

Windows PowerShell 若限制执行 `npm.ps1`，可使用 `npm.cmd ci` 和 `npm.cmd run build`。

## Commit 与 Pull Request

Commit 信息应简短描述结果，例如：

```text
fix: correct customer statement totals
docs: improve GitHub setup guide
```

Pull Request 请说明改动目的、主要变化、验证方式和兼容性影响。涉及界面变化时附上截图；涉及迁移时说明升级和回滚注意事项。请保持一个 Pull Request 只解决一个相对独立的问题。

## 报告问题

普通缺陷和功能建议可使用仓库的 Issue 模板。安全漏洞不要公开提交 Issue，请按 [`SECURITY.md`](SECURITY.md) 的方式报告。
