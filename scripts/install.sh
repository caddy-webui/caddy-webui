#!/bin/bash
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

info()  { echo -e "${GREEN}[INFO]${NC} $1"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

if [ "$(id -u)" -ne 0 ]; then
    error "请使用 root 用户运行此脚本"
fi

info "Caddy WebUI 管理面板安装脚本"

OS_ID=""
OS_VERSION=""
OS_VERSION_ID=""
PKG_MANAGER=""

detect_os() {
    if [ ! -f /etc/os-release ]; then
        error "无法检测操作系统，/etc/os-release 不存在"
    fi

    source /etc/os-release
    OS_ID="$ID"
    OS_VERSION_ID="$VERSION_ID"

    case "$OS_ID" in
        debian)
            OS_VERSION=$(echo "$OS_VERSION_ID" | cut -d. -f1)
            if [ "$OS_VERSION" -lt 12 ]; then
                error "当前操作系统 Debian $OS_VERSION_ID 版本过低，Debian 需要 12 及以上版本，安装终止"
            fi
            PKG_MANAGER="apt"
            ;;
        ubuntu)
            if awk 'BEGIN{exit !('"${OS_VERSION_ID}"' < 18.04)}'; then
                error "当前操作系统 Ubuntu $OS_VERSION_ID 版本过低，Ubuntu 需要 18.04 及以上版本，安装终止"
            fi
            PKG_MANAGER="apt"
            ;;
        centos)
            if [[ "$PRETTY_NAME" == *Stream* ]] || [[ "$VERSION" == *Stream* ]]; then
                OS_VERSION=$(echo "$OS_VERSION_ID" | cut -d. -f1)
                if [ "$OS_VERSION" -lt 8 ]; then
                    error "当前操作系统 CentOS Stream $OS_VERSION_ID 版本过低，CentOS Stream 需要 8 及以上版本，安装终止"
                fi
                PKG_MANAGER="dnf"
            else
                OS_VERSION=$(echo "$OS_VERSION_ID" | cut -d. -f1)
                if [ "$OS_VERSION" -lt 7 ]; then
                    error "当前操作系统 CentOS $OS_VERSION_ID 版本过低，CentOS 需要 7 及以上版本，安装终止"
                fi
                if [ "$OS_VERSION" -le 7 ]; then
                    PKG_MANAGER="yum"
                else
                    PKG_MANAGER="dnf"
                fi
            fi
            ;;
        almalinux)
            OS_VERSION=$(echo "$OS_VERSION_ID" | cut -d. -f1)
            if [ "$OS_VERSION" -lt 8 ]; then
                error "当前操作系统 AlmaLinux $OS_VERSION_ID 版本过低，AlmaLinux 需要 8 及以上版本，安装终止"
            fi
            PKG_MANAGER="dnf"
            ;;
        rocky)
            OS_VERSION=$(echo "$OS_VERSION_ID" | cut -d. -f1)
            if [ "$OS_VERSION" -lt 8 ]; then
                error "当前操作系统 Rocky Linux $OS_VERSION_ID 版本过低，Rocky Linux 需要 8 及以上版本，安装终止"
            fi
            PKG_MANAGER="dnf"
            ;;
        rhel)
            OS_VERSION=$(echo "$OS_VERSION_ID" | cut -d. -f1)
            if [ "$OS_VERSION" -lt 7 ]; then
                error "当前操作系统 RHEL $OS_VERSION_ID 版本过低，RHEL 需要 7 及以上版本，安装终止"
            fi
            if [ "$OS_VERSION" -le 7 ]; then
                PKG_MANAGER="yum"
            else
                PKG_MANAGER="dnf"
            fi
            ;;
        *)
            error "当前操作系统 $OS_ID 不在支持列表中，安装终止。支持的系统：Debian 12+、Ubuntu 18.04+、CentOS 7、CentOS Stream 8+、AlmaLinux 8+、Rocky Linux 8+、RHEL 7+"
            ;;
    esac

    info "检测到操作系统: $PRETTY_NAME"
    info "使用包管理器: $PKG_MANAGER"
}

install_dependencies() {
    info "安装基础依赖..."

    if [ "$PKG_MANAGER" = "apt" ]; then
        apt-get update -y
        apt-get install -y curl wget unzip sqlite3
    elif [ "$PKG_MANAGER" = "yum" ]; then
        yum install -y curl wget unzip sqlite
    elif [ "$PKG_MANAGER" = "dnf" ]; then
        dnf install -y curl wget unzip sqlite
    fi
}

install_caddy() {
    if command -v caddy &> /dev/null; then
        info "Caddy 已安装: $(caddy version)"
        return
    fi

    info "安装 Caddy..."

    if [ "$PKG_MANAGER" = "apt" ]; then
        apt-get install -y debian-keyring debian-archive-keyring apt-transport-https curl
        curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
        curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list
        apt-get update
        apt-get install -y caddy
    elif [ "$PKG_MANAGER" = "yum" ]; then
        yum install -y yum-plugin-copr
        yum copr enable -y @caddy/caddy
        yum install -y caddy
    elif [ "$PKG_MANAGER" = "dnf" ]; then
        dnf install -y dnf-plugins-core
        dnf copr enable -y @caddy/caddy
        dnf install -y caddy
    fi

    if command -v caddy &> /dev/null; then
        info "Caddy 安装成功: $(caddy version)"
    else
        error "Caddy 安装失败，请检查网络或手动安装"
    fi
}

create_dirs() {
    info "创建目录结构..."
    mkdir -p /opt/caddy-webui/{bin,config,data,sites,ssl,www,log,scripts}
    chmod 700 /opt/caddy-webui/data
    chmod 700 /opt/caddy-webui/ssl
}

deploy_binary() {
    info "部署面板程序..."

    SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

    if [ -f "$SCRIPT_DIR/caddy-webui" ]; then
        cp "$SCRIPT_DIR/caddy-webui" /opt/caddy-webui/bin/caddy-webui
    elif [ -f "$(dirname "$0")/caddy-webui" ]; then
        cp "$(dirname "$0")/caddy-webui" /opt/caddy-webui/bin/caddy-webui
    else
        warn "未找到本地二进制文件，请手动将编译后的 caddy-webui 复制到 /opt/caddy-webui/bin/"
        warn "编译命令: CGO_ENABLED=1 go build -ldflags=\"-s -w\" -o /opt/caddy-webui/bin/caddy-webui ."
        return
    fi

    chmod 755 /opt/caddy-webui/bin/caddy-webui
}

deploy_config() {
    info "部署默认配置..."

    if [ ! -f /opt/caddy-webui/config/config.toml ]; then
        cat > /opt/caddy-webui/config/config.toml << 'EOF'
[server]
port = 8729
host = "0.0.0.0"

[log]
level = "INFO"
dir = "/opt/caddy-webui/log/"

[caddy]
binary_path = "/usr/bin/caddy"
config_path = "/opt/caddy-webui/config/Caddyfile"
service_name = "caddy"
admin_api = "http://localhost:2019"
EOF
    fi

    if [ ! -f /opt/caddy-webui/config/Caddyfile ]; then
        touch /opt/caddy-webui/config/Caddyfile
    fi
}

register_service() {
    info "注册 systemd 服务..."

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
    systemctl enable caddy-webui
}

configure_firewall() {
    info "配置防火墙..."

    if command -v ufw &> /dev/null && ufw status | grep -q "active"; then
        ufw allow 8729/tcp
        ufw allow 80/tcp
        ufw allow 443/tcp
        info "已通过 ufw 开放端口 8729、80、443"
    elif command -v firewall-cmd &> /dev/null && firewall-cmd --state 2>/dev/null | grep -q "running"; then
        firewall-cmd --permanent --add-port=8729/tcp
        firewall-cmd --permanent --add-port=80/tcp
        firewall-cmd --permanent --add-port=443/tcp
        firewall-cmd --reload
        info "已通过 firewalld 开放端口 8729、80、443"
    else
        warn "未检测到活跃的防火墙，请手动开放端口 8729、80、443"
    fi
}

check_memory() {
    local mem_total=$(awk '/MemTotal/ {print $2}' /proc/meminfo 2>/dev/null || echo "0")
    local mem_mb=$((mem_total / 1024))

    if [ "$mem_mb" -lt 128 ]; then
        warn "系统内存仅 ${mem_mb}MB，建议至少 128MB"
        warn "低内存环境下建议减少 Caddy 缓存并优化 SQLite 配置"
    else
        info "系统内存: ${mem_mb}MB"
    fi
}

start_service() {
    info "启动 caddy-webui 服务..."
    systemctl start caddy-webui || warn "服务启动失败，请检查日志: journalctl -u caddy-webui"
}

show_result() {
    echo ""
    echo "============================================"
    info "Caddy WebUI 管理面板安装完成！"
    echo ""
    info "访问地址: http://<服务器IP>:8729"
    info "首次访问请设置管理员账号"
    echo ""
    info "常用命令:"
    echo "  systemctl status caddy-webui   # 查看状态"
    echo "  systemctl restart caddy-webui  # 重启服务"
    echo "  journalctl -u caddy-webui      # 查看日志"
    echo "============================================"
}

detect_os
check_memory
install_dependencies
install_caddy
create_dirs
deploy_binary
deploy_config
register_service
configure_firewall
start_service
show_result
