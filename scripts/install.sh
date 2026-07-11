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

    # 优先使用脚本同目录下的预编译二进制文件
    if [ -f "$SCRIPT_DIR/caddy-webui" ]; then
        cp "$SCRIPT_DIR/caddy-webui" /opt/caddy-webui/bin/caddy-webui
        chmod 755 /opt/caddy-webui/bin/caddy-webui
        info "已部署本地二进制文件"
        return
    fi

    # 检查是否已有编译好的二进制文件（如重复安装）
    if [ -f /opt/caddy-webui/bin/caddy-webui ]; then
        info "检测到已存在的二进制文件，跳过部署"
        return
    fi

    # 尝试从源码自动编译
    info "未找到本地二进制文件，尝试从源码编译..."

    # 安装编译依赖（CGO 需要 gcc 和 libc-dev）
    info "安装编译依赖..."
    if [ "$PKG_MANAGER" = "apt" ]; then
        apt-get install -y golang-go build-essential
    elif [ "$PKG_MANAGER" = "yum" ]; then
        yum install -y golang gcc make
    elif [ "$PKG_MANAGER" = "dnf" ]; then
        dnf install -y golang gcc make
    fi

    # 检查 Go 是否可用
    if ! command -v go &> /dev/null; then
        # 系统包管理器的 Go 版本可能过低，尝试安装官方版本
        info "系统 Go 版本不满足要求，安装官方 Go 环境..."
        local GO_VERSION="1.21.13"
        local GO_ARCH="$(uname -m)"
        case "$GO_ARCH" in
            x86_64)  GO_ARCH="amd64" ;;
            aarch64) GO_ARCH="arm64" ;;
            *)       error "不支持的系统架构: $GO_ARCH" ;;
        esac

        wget -q "https://go.dev/dl/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz" -O /tmp/go.tar.gz
        rm -rf /usr/local/go
        tar -C /usr/local -xzf /tmp/go.tar.gz
        rm -f /tmp/go.tar.gz
        export PATH="/usr/local/go/bin:$PATH"

        if ! command -v go &> /dev/null; then
            error "Go 环境安装失败，请手动安装 Go 1.21+ 后重新运行"
        fi
    fi

    info "Go 版本: $(go version)"

    # 查找项目源码目录
    local SOURCE_DIR=""
    if [ -f "$SCRIPT_DIR/../main.go" ] && [ -f "$SCRIPT_DIR/../go.mod" ]; then
        SOURCE_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
    elif [ -f "$SCRIPT_DIR/../../main.go" ] && [ -f "$SCRIPT_DIR/../../go.mod" ]; then
        SOURCE_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
    fi

    if [ -z "$SOURCE_DIR" ]; then
        error "未找到项目源码目录（需要 main.go 和 go.mod），请将编译后的 caddy-webui 放到脚本同目录后重新运行"
    fi

    info "从源码编译: $SOURCE_DIR"
    cd "$SOURCE_DIR"

    # 下载依赖并编译
    go mod download
    CGO_ENABLED=1 go build -ldflags="-s -w" -o /opt/caddy-webui/bin/caddy-webui .

    if [ -f /opt/caddy-webui/bin/caddy-webui ]; then
        chmod 755 /opt/caddy-webui/bin/caddy-webui
        info "编译成功"
    else
        error "编译失败，请检查错误信息或手动编译后重新运行"
    fi
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

    if command -v ufw &> /dev/null && ufw status 2>/dev/null | grep -q "active"; then
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
    elif command -v nft &> /dev/null && systemctl is-active nftables &>/dev/null; then
        # Debian 12 默认使用 nftables
        local NFT_TABLE="caddy-webui"
        if ! nft list table inet "$NFT_TABLE" &>/dev/null; then
            nft add table inet "$NFT_TABLE"
        fi
        local NFT_CHAIN="${NFT_TABLE}_input"
        if ! nft list chain inet "$NFT_TABLE" "$NFT_CHAIN" &>/dev/null; then
            nft add chain inet "$NFT_TABLE" "$NFT_CHAIN" '{ type filter hook input priority 0 ; policy accept ; }'
        fi
        nft add rule inet "$NFT_TABLE" "$NFT_CHAIN" tcp dport { 80, 443, 8729 } accept
        # 持久化规则
        if [ -d /etc/nftables.d ]; then
            nft list ruleset > /etc/nftables.d/caddy-webui.nft 2>/dev/null
        elif [ -f /etc/nftables.conf ]; then
            nft list ruleset > /etc/nftables.conf
        fi
        info "已通过 nftables 开放端口 8729、80、443"
    elif command -v iptables &> /dev/null; then
        # 回退到 iptables
        iptables -A INPUT -p tcp --dport 8729 -j ACCEPT
        iptables -A INPUT -p tcp --dport 80 -j ACCEPT
        iptables -A INPUT -p tcp --dport 443 -j ACCEPT
        # 持久化规则
        if command -v iptables-save &> /dev/null; then
            if [ -d /etc/iptables ]; then
                iptables-save > /etc/iptables/rules.v4
            elif command -v netfilter-persistent &> /dev/null; then
                netfilter-persistent save
            fi
        fi
        info "已通过 iptables 开放端口 8729、80、443"
    else
        warn "未检测到防火墙工具（ufw/firewalld/nftables/iptables），请手动开放端口 8729、80、443"
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
    systemctl start caddy-webui

    # 等待服务就绪（最多等待 10 秒）
    local RETRY=0
    local MAX_RETRY=10
    while [ $RETRY -lt $MAX_RETRY ]; do
        if systemctl is-active --quiet caddy-webui; then
            break
        fi
        RETRY=$((RETRY + 1))
        sleep 1
    done

    if systemctl is-active --quiet caddy-webui; then
        info "caddy-webui 服务启动成功"
    else
        error "caddy-webui 服务启动失败，请检查日志: journalctl -u caddy-webui -n 50"
    fi
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
