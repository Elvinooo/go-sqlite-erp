Go SQLite ERP Windows 运行包说明

适用环境：Windows 10/11 或 Windows Server 64 位。
运行包使用 SQLite，不需要安装 MySQL、PostgreSQL 或其他数据库服务。

首次运行
========

1. 解压完整运行包，不要只复制单个 exe 文件。
2. 打开 config.yaml，至少修改 jwt.secret 和 admin.password。
3. 双击 start-all.bat。
4. 浏览器打开：http://localhost:5173/login

局域网访问：http://SERVER_IP:5173/login
如需局域网或公网访问，请同时配置 Windows 防火墙、反向代理和 HTTPS。

默认演示账号
============

用户名：admin
密码：admin888

默认配置启用 app.demo_mode: true，会限制部分危险操作。正式使用前请修改密码，完成安全检查后再按需要设置：

app:
  demo_mode: false

修改配置后需要重新启动服务。

目录与文件
==========

erp-server.exe         后端服务
erp-frontend.exe       前端静态服务和 /api 代理
config.yaml            本机配置，不应公开上传
web/                   已构建的前端文件
data/                  SQLite 数据库和上传文件，必须备份
logs/                  后端和前端日志
start-all.bat          启动全部服务
start-backend.bat      仅启动后端
start-frontend.ps1     仅启动前端
stop-all.bat           停止 18080 和 5173 端口上的服务
status.bat             检查服务状态
install-startup.bat    安装开机/登录启动任务
uninstall-startup.bat  删除自动启动任务

自动启动
========

右键 install-startup.bat，选择“以管理员身份运行”。
删除自动启动时，以管理员身份运行 uninstall-startup.bat。

数据与安全
==========

- 默认数据库：data/erp.db
- 默认后端端口：18080
- 默认前端端口：5173
- 定期备份整个 data/ 目录，并验证备份可恢复。
- 不要公开分享 config.yaml、data/、logs/ 或包含真实信息的截图。
- 公网部署必须使用 HTTPS，限制防火墙端口，并更换示例密码和 JWT 密钥。
- 升级运行包前先停止服务并备份 data/ 和 config.yaml。
