# Caddy WebUI

一个轻量级的 Caddy Web 管理面板，专为小内存 VPS 设计（最低 64MB），提供直观的 Web 界面管理 Caddy 的网站、反向代理、SSL 证书等功能。

## 功能特性

- **仪表盘** — 实时显示系统状态（CPU、内存、磁盘）、Caddy 运行状态、站点数量、证书概览
- **站点管理** — 新增、编辑、删除、启用/禁用站点，支持域名格式校验和唯一性检查
- **反向代理配置** — 代理目标 URL、路径路由、负载均衡（多后端）、自定义请求头、WebSocket 支持
- **SSL 证书管理** — 自动申请（Let's Encrypt ACME）和自定义上传两种模式，支持证书更新和模式切换
- **Caddyfile 编辑器** — 在线编辑 Caddyfile，语法校验，自动备份回滚
- **文件管理** — 上传静态文件到站点目录
- **全局设置** — 修改 WebUI 端口、管理员密码、日志级别
- **Caddy 服务控制** — 通过 Web 界面启动、停止、重启、重载 Caddy
- **IPv6 支持** — 所有站点自动添加 `bind ::` 指令
- **一键安装** — 支持 Debian 12+、Ubuntu 18.04+、CentOS 7、CentOS Stream 8+、AlmaLinux 8+、Rocky Linux 8+、RHEL 7+

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端 | Go 1.21+，原生 `net/http` 标准库 |
| 数据库 | SQLite3（嵌入式，CGO） |
| 前端 | 原生 HTML5 + CSS3 + JavaScript（无框架） |
| Caddy | v2，通过 Caddyfile + Admin API 管理 |
| 认证 | JWT (HS256) + bcrypt |
| 安装 | Shell 脚本，兼容主流 Linux 发行版 |

## 项目结构

```
caddy-webui/
├── main.go                          # 程序入口，路由注册
├── internal/
│   ├── config/                      # 配置加载与日志
│   ├── database/                    # SQLite3 数据库操作
│   ├── models/                      # 数据模型
│   ├── handlers/                    # HTTP 请求处理器
│   ├── middleware/                   # 中间件（认证、日志、恢复等）
│   ├── auth/                        # JWT + bcrypt + 账号锁定
│   ├── caddy/                       # Caddyfile 生成与服务控制
│   └── system/                      # 系统状态监控
├── static/                          # 前端静态文件（Go embed 嵌入）
│   ├── css/
│   ├── js/
│   └── index.html
├── scripts/
│   └── install.sh                   # 一键安装脚本
├── config/
│   └── config.toml                  # 默认配置模板
├── Makefile
├── LICENSE
└── README.md
```

## 快速开始

### 一键安装（推荐）

在目标服务器上执行：

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/caddy-webui/caddy-webui/main/scripts/install.sh)
```

或手动下载脚本：

```bash
wget https://raw.githubusercontent.com/caddy-webui/caddy-webui/main/scripts/install.sh
bash install.sh
```

安装完成后访问 `http://<服务器IP>:8729` 设置管理员账号。

### 从源码编译

**前置条件**：Go 1.21+、GCC（CGO 编译需要）

```bash
git clone https://github.com/caddy-webui/caddy-webui.git
cd caddy-webui
go mod tidy
CGO_ENABLED=1 go build -ldflags="-s -w" -o bin/caddy-webui .
```

### 手动部署

```bash
# 创建目录
mkdir -p /opt/caddy-webui/{bin,config,data,sites,ssl,www,log}

# 复制二进制和配置
cp bin/caddy-webui /opt/caddy-webui/bin/
cp config/config.toml /opt/caddy-webui/config/

# 注册 systemd 服务
cat > /etc/systemd/system/caddy-webui.service << 'EOF'
[Unit]
Description=Caddy WebUI Management Panel
After=network.target caddy.service

[Service]
Type=simple
ExecStart=/opt/caddy-webui/bin/caddy-webui
WorkingDirectory=/opt/caddy-webui
Restart=on-failure
RestartSec=5
Environment=HOME=/opt/caddy-webui

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now caddy-webui
```

## 配置说明

配置文件路径：`/opt/caddy-webui/config/config.toml`

```toml
[server]
port = 8729          # WebUI 监听端口
host = "0.0.0.0"     # 监听地址（默认仅本地访问）

[log]
level = "INFO"       # 日志级别: DEBUG/INFO/WARN/ERROR
dir = "/opt/caddy-webui/log/"

[caddy]
binary_path = "/usr/bin/caddy"
config_path = "/opt/caddy-webui/config/Caddyfile"
service_name = "caddy"
admin_api = "http://localhost:2019"
```

## API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/auth/setup | 系统初始化 |
| POST | /api/auth/login | 管理员登录 |
| PUT | /api/auth/password | 修改密码 |
| GET | /api/auth/status | 检查初始化状态 |
| GET | /api/dashboard | 仪表盘数据 |
| GET | /api/sites | 站点列表 |
| POST | /api/sites | 新增站点 |
| GET/PUT/DELETE | /api/sites/:id | 站点详情/更新/删除 |
| PUT | /api/sites/:id/toggle | 启用/禁用站点 |
| GET/POST | /api/caddy/status\|start\|stop\|restart\|reload | Caddy 服务控制 |
| GET | /api/certificates | 证书列表 |
| POST | /api/certificates/:id/renew | 续期证书 |
| POST | /api/certificates/:id/upload | 上传自定义证书 |
| PUT | /api/certificates/:id/update | 更新证书文件 |
| PUT | /api/certificates/:id/mode | 切换证书模式 |
| GET/PUT | /api/settings | 全局设置 |
| GET/PUT | /api/files/caddyfile | Caddyfile 编辑 |
| POST | /api/files/upload | 上传静态文件 |

## 性能指标

| 指标 | 目标值 |
|------|--------|
| 面板运行内存 | < 30MB |
| Caddy 运行内存 | < 20MB |
| API 响应时间 | < 3s |
| 配置变更生效时间 | < 5s |

## 安全特性

- 管理员密码 bcrypt 加密存储
- JWT (HS256) 令牌认证，24 小时有效期
- 连续 5 次登录失败锁定 15 分钟
- Caddyfile 操作自动备份回滚
- 自定义证书上传时校验 PEM 格式和公私钥匹配性
- 文件上传禁止可执行文件

## 支持的操作系统

| 操作系统 | 最低版本 |
|----------|----------|
| Debian | 12+ |
| Ubuntu | 18.04+ |
| CentOS | 7 |
| CentOS Stream | 8+ |
| AlmaLinux | 8+ |
| Rocky Linux | 8+ |
| RHEL | 7+ |

## License

[MIT](LICENSE)
