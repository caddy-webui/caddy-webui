[中文文档](README_CN.md)

# Caddy WebUI

A lightweight Caddy web management panel designed for low-memory VPS (minimum 64MB), providing an intuitive web interface to manage Caddy websites, reverse proxies, SSL certificates, and more.

## Features

- **Dashboard** — Real-time system status (CPU, memory, disk), Caddy running status, site count, certificate overview
- **Site Management** — Create, edit, delete, enable/disable sites with domain validation and uniqueness checks
- **Reverse Proxy Configuration** — Proxy target URL, path routing, load balancing (multiple backends), custom headers, WebSocket support
- **SSL Certificate Management** — Two modes: auto-provision (Let's Encrypt ACME) and custom upload, with certificate update and mode switching
- **Caddyfile Editor** — Online editing with syntax validation and automatic backup/rollback
- **File Management** — Upload static files to site directories
- **Global Settings** — Modify WebUI port, admin password, log level
- **Caddy Service Control** — Start, stop, restart, and reload Caddy via web interface
- **IPv6 Support** — Automatically adds `bind ::` directive for all sites
- **One-Click Install** — Supports Debian 12+, Ubuntu 18.04+, CentOS 7, CentOS Stream 8+, AlmaLinux 8+, Rocky Linux 8+, RHEL 7+

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Backend | Go 1.21+, native `net/http` standard library |
| Database | SQLite3 (embedded, CGO) |
| Frontend | Native HTML5 + CSS3 + JavaScript (no framework) |
| Caddy | v2, managed via Caddyfile + Admin API |
| Auth | JWT (HS256) + bcrypt |
| Install | Shell script, compatible with major Linux distros |

## Project Structure

```
caddy-webui/
├── main.go                          # Entry point, route registration
├── internal/
│   ├── config/                      # Config loading & logging
│   ├── database/                    # SQLite3 database operations
│   ├── models/                      # Data models
│   ├── handlers/                    # HTTP request handlers
│   ├── middleware/                   # Middleware (auth, logging, recovery, etc.)
│   ├── auth/                        # JWT + bcrypt + account lockout
│   ├── caddy/                       # Caddyfile generation & service control
│   └── system/                      # System status monitoring
├── static/                          # Frontend static files (embedded via Go embed)
│   ├── css/
│   ├── js/
│   └── index.html
├── scripts/
│   └── install.sh                   # One-click install script
├── config/
│   └── config.toml                  # Default config template
├── Makefile
├── LICENSE
└── README.md
```

## Quick Start

### One-Click Install (Recommended)

Run on your target server:

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/caddy-webui/caddy-webui/main/scripts/install.sh)
```

Or download the script manually:

```bash
wget https://raw.githubusercontent.com/caddy-webui/caddy-webui/main/scripts/install.sh
bash install.sh
```

After installation, visit `http://<SERVER_IP>:8729` to set up the admin account.

### Build from Source

**Prerequisites**: Go 1.21+, GCC (required for CGO)

```bash
git clone https://github.com/caddy-webui/caddy-webui.git
cd caddy-webui
go mod tidy
CGO_ENABLED=1 go build -ldflags="-s -w" -o bin/caddy-webui .
```

### Manual Deployment

```bash
# Create directories
mkdir -p /opt/caddy-webui/{bin,config,data,sites,ssl,www,log}

# Copy binary and config
cp bin/caddy-webui /opt/caddy-webui/bin/
cp config/config.toml /opt/caddy-webui/config/

# Register systemd service
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

## Configuration

Config file path: `/opt/caddy-webui/config/config.toml`

```toml
[server]
port = 8729          # WebUI listening port
host = "0.0.0.0"     # Listen address (default: localhost only)

[log]
level = "INFO"       # Log level: DEBUG/INFO/WARN/ERROR
dir = "/opt/caddy-webui/log/"

[caddy]
binary_path = "/usr/bin/caddy"
config_path = "/opt/caddy-webui/config/Caddyfile"
service_name = "caddy"
admin_api = "http://localhost:2019"
```

## API Reference

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /api/auth/setup | System initialization |
| POST | /api/auth/login | Admin login |
| PUT | /api/auth/password | Change password |
| GET | /api/auth/status | Check initialization status |
| GET | /api/dashboard | Dashboard data |
| GET | /api/sites | List sites |
| POST | /api/sites | Create site |
| GET/PUT/DELETE | /api/sites/:id | Get/update/delete site |
| PUT | /api/sites/:id/toggle | Enable/disable site |
| GET/POST | /api/caddy/status\|start\|stop\|restart\|reload | Caddy service control |
| GET | /api/certificates | List certificates |
| POST | /api/certificates/:id/renew | Renew certificate |
| POST | /api/certificates/:id/upload | Upload custom certificate |
| PUT | /api/certificates/:id/update | Update certificate files |
| PUT | /api/certificates/:id/mode | Switch certificate mode |
| GET/PUT | /api/settings | Global settings |
| GET/PUT | /api/files/caddyfile | Caddyfile editor |
| POST | /api/files/upload | Upload static files |

## Performance Targets

| Metric | Target |
|--------|--------|
| Panel memory usage | < 30MB |
| Caddy memory usage | < 20MB |
| API response time | < 3s |
| Config change apply time | < 5s |

## Security Features

- Admin password stored with bcrypt encryption
- JWT (HS256) token authentication with 24-hour expiry
- Account lockout after 5 consecutive failed login attempts (15 minutes)
- Automatic backup and rollback for Caddyfile operations
- PEM format and key-pair matching validation on custom certificate upload
- Executable file upload restriction

## Supported Operating Systems

| OS | Minimum Version |
|----|----------------|
| Debian | 12+ |
| Ubuntu | 18.04+ |
| CentOS | 7 |
| CentOS Stream | 8+ |
| AlmaLinux | 8+ |
| Rocky Linux | 8+ |
| RHEL | 7+ |

## License

[MIT](LICENSE)
