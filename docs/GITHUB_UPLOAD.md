# 上传到 GitHub

本文档用于首次将本地项目发布到 GitHub。当前仓库如果尚无提交或远程地址，请按下面顺序操作。

## 1. 上传前检查

确认本地敏感文件已被忽略：

```bash
git status --short
git status --ignored
git check-ignore -v config.yaml data logs web/.env.local
```

预期 `config.yaml`、`data/`、`logs/` 和 `web/.env.local` 都显示为 ignored。不要使用 `git add -f` 强制加入这些文件。

进一步检查待提交文件：

```bash
git diff --cached --stat
git diff --cached
```

提交前运行项目检查：

```bash
go test ./...
go vet ./...
cd web
npm ci
npm run build
cd ..
```

Windows PowerShell 如果无法执行 `npm.ps1`，请改用 `npm.cmd ci` 和 `npm.cmd run build`。

## 2. 在 GitHub 创建空仓库

在 GitHub 新建名为 `go-sqlite-erp` 的仓库。由于本地已经包含 README、`.gitignore` 和 License，创建时不要勾选自动生成这些文件，避免首次推送产生无关冲突。

公开仓库会让任何人看到全部提交历史。推送前请确认历史中没有真实密码、令牌、客户数据或数据库文件；仅从当前文件删除秘密并不能清除 Git 历史。

## 3. 创建首次提交

```bash
git branch -M main
git add .
git commit -m "chore: initial open-source release"
```

提交后检查：

```bash
git status
git show --stat --oneline HEAD
```

## 4. 关联并推送远程仓库

HTTPS 方式：

```bash
git remote add origin https://github.com/YOUR_ACCOUNT/go-sqlite-erp.git
git push -u origin main
```

SSH 方式：

```bash
git remote add origin git@github.com:YOUR_ACCOUNT/go-sqlite-erp.git
git push -u origin main
```

把 `YOUR_ACCOUNT` 替换为你的 GitHub 用户名或组织名。如果已经存在 `origin`，先查看当前地址：

```bash
git remote -v
```

需要更换时使用：

```bash
git remote set-url origin https://github.com/YOUR_ACCOUNT/go-sqlite-erp.git
```

## 5. 使用 GitHub CLI（可选）

已经安装并登录 `gh` 时，可以创建仓库并推送：

```bash
gh auth status
gh repo create YOUR_ACCOUNT/go-sqlite-erp --public --source=. --remote=origin --push
```

如项目只用于内部部署，将 `--public` 改为 `--private`。

## 6. 上传后的仓库设置

建议在 GitHub 仓库设置中完成：

- 将默认分支设为 `main`
- 为 `main` 启用分支保护，要求 Pull Request 和 CI 通过
- 开启 Dependabot alerts 和 secret scanning（可用时）
- 在 About 中补充项目描述、主题标签和主页地址
- 使用 Releases 发布运行包，不要把二进制和数据库直接提交到源码分支

仓库内的 GitHub Actions 会在推送和 Pull Request 时检查后端测试、`go vet` 和前端构建。

## 常见问题

### 推送提示 `remote origin already exists`

执行 `git remote -v` 检查地址，然后使用 `git remote set-url origin ...` 修改，不要重复添加。

### 远程仓库已有 README，推送被拒绝

最简单的处理方式是删除刚创建的远程仓库并重新创建一个空仓库。若远程已有重要内容，应先正常拉取并合并，不要使用 `--force` 覆盖。

### 错误提交了密钥或数据库

立即撤销或轮换相关密钥，并在公开前清理整个 Git 历史。仅添加 `.gitignore` 或删除当前文件不能移除历史内容。已经推送时，还需要通知所有协作者重新同步干净历史。

### Go 模块名为什么是 `erp`

本项目当前是应用程序，短模块名可正常构建。如果未来希望作为可导入的 Go 模块发布，可在确定 GitHub 地址后将模块名改为 `github.com/YOUR_ACCOUNT/go-sqlite-erp`，并同步更新源码中的内部导入路径。
