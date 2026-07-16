Go SQLite ERP Linux 运行包说明

选择与服务器 CPU 匹配的运行包：
- amd64 / x86_64：erp-linux-amd64-runtime
- arm64 / aarch64：erp-linux-arm64-runtime

运行包使用 SQLite，不需要安装额外数据库服务。

首次运行
========

1. 解压完整运行包并进入目录。
2. 编辑 config.yaml，至少修改 jwt.secret 和 admin.password。
3. 执行：

   chmod +x *.sh erp-server erp-frontend
   ./start-all.sh

4. 浏览器打开：http://SERVER_IP:5173/login

默认演示账号
============

用户名：admin
密码：admin888

默认配置启用 app.demo_mode: true。正式使用前请修改密码，完成安全检查后再按需要设置：

app:
  demo_mode: false

然后执行：

./stop-all.sh
./start-all.sh

服务管理
========

启动：./start-all.sh
停止：./stop-all.sh
状态：./status.sh
安装 systemd：sudo ./install-systemd.sh
卸载 systemd：sudo ./uninstall-systemd.sh

目录与端口
==========

erp-server       后端服务
erp-frontend     前端静态服务和 /api 代理
config.yaml      本机配置，不应公开上传
web/             已构建的前端文件
data/            SQLite 数据库和上传文件
logs/            后端和前端日志

后端端口：18080
前端端口：5173
默认数据库：data/erp.db

数据与安全
==========

- 定期备份整个 data/ 目录，并验证备份可恢复。
- 不要公开分享 config.yaml、data/、logs/ 或包含真实信息的截图。
- 公网部署应通过 HTTPS 反向代理访问，不建议直接暴露后端 18080 端口。
- 使用防火墙限制端口，并更换示例密码和 JWT 密钥。
- 建议使用专用的非 root 系统用户运行服务，并限制运行目录权限。
- 升级运行包前先停止服务并备份 data/ 和 config.yaml。
