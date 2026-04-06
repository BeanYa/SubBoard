#!/bin/bash
set -e

# ============================================
#  SubBoard 一键部署脚本 (1Panel 环境)
#  用法: bash deploy-1panel.sh
# ============================================

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
RED='\033[0;31m'
NC='\033[0m'

info()  { echo -e "${CYAN}[INFO]${NC} $1"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
ok()    { echo -e "${GREEN}[OK]${NC} $1"; }
err()   { echo -e "${RED}[ERROR]${NC} $1"; }

INSTALL_DIR="/opt/subboard"
BACKEND_PORT=8080
REPO_URL="https://github.com/BeanYa/SubBoard.git"

echo ""
echo -e "${CYAN}============================================${NC}"
echo -e "${CYAN}  SubBoard 一键部署 (1Panel 环境)          ${NC}"
echo -e "${CYAN}============================================${NC}"
echo ""

# ---------- 1. 参数收集 ----------
DOMAIN=""
read -p "绑定域名 (留空则用 IP 访问): " DOMAIN

if [ -z "$DOMAIN" ]; then
    SERVER_IP=$(curl -sf ifconfig.me 2>/dev/null || hostname -I | awk '{print $1}')
    SUB_BASE_URL="http://${SERVER_IP}:${BACKEND_PORT}"
    warn "未设置域名，订阅链接将使用 ${SUB_BASE_URL}"
else
    read -p "是否启用 HTTPS? (y/n) [y]: " ENABLE_HTTPS
    ENABLE_HTTPS=${ENABLE_HTTPS:-y}
    if [ "$ENABLE_HTTPS" = "y" ]; then
        SUB_BASE_URL="https://${DOMAIN}"
    else
        SUB_BASE_URL="http://${DOMAIN}"
    fi
fi

read -p "管理员用户名 [admin]: " ADMIN_USER
ADMIN_USER=${ADMIN_USER:-admin}

read -p "管理员密码 [admin]: " ADMIN_PASS
ADMIN_PASS=${ADMIN_PASS:-admin}

APP_SECRET=$(openssl rand -hex 32 2>/dev/null || head -c 32 /dev/urandom | xxd -p)
INIT_TOKEN=$(openssl rand -hex 16 2>/dev/null || head -c 16 /dev/urandom | xxd -p)

read -p "后端监听端口 [8080]: " BACKEND_PORT
BACKEND_PORT=${BACKEND_PORT:-8080}

read -p "允许用户注册 (true/false) [true]: " ALLOW_REGISTER
ALLOW_REGISTER=${ALLOW_REGISTER:-true}

echo ""
echo -e "${CYAN}--- 配置摘要 ---${NC}"
echo "  安装目录:     $INSTALL_DIR"
echo "  绑定域名:     ${DOMAIN:-无(用IP)}"
echo "  订阅基础URL:  $SUB_BASE_URL"
echo "  后端端口:     $BACKEND_PORT"
echo "  管理员:       $ADMIN_USER / $ADMIN_PASS"
echo -e "${CYAN}----------------${NC}"
read -p "确认部署? (y/n) [y]: " CONFIRM
if [ -n "$CONFIRM" ] && [ "$CONFIRM" != "y" ] && [ "$CONFIRM" != "Y" ]; then
    err "已取消"; exit 1
fi

# ---------- 2. 安装依赖 ----------
install_deps() {
    info "检查系统依赖..."

    # Go
    if ! command -v go &>/dev/null; then
        info "安装 Go 1.23..."
        GO_VERSION="1.23.4"
        wget -q "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -O /tmp/go.tar.gz
        rm -rf /usr/local/go
        tar -C /usr/local -xzf /tmp/go.tar.gz
        rm /tmp/go.tar.gz
        export PATH=$PATH:/usr/local/go/bin
        grep -q '/usr/local/go/bin' /etc/profile || echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
        ok "Go $(go version | awk '{print $3}') 已安装"
    else
        ok "Go $(go version | awk '{print $3}')"
    fi

    # Node.js
    if ! command -v node &>/dev/null; then
        info "安装 Node.js 20..."
        curl -fsSL https://deb.nodesource.com/setup_20.x | bash - >/dev/null 2>&1
        apt-get install -y nodejs >/dev/null 2>&1
        ok "Node.js $(node -v) 已安装"
    else
        ok "Node.js $(node -v)"
    fi

    # Git
    if ! command -v git &>/dev/null; then
        info "安装 Git..."
        apt-get install -y git >/dev/null 2>&1 || yum install -y git >/dev/null 2>&1
    fi
    ok "Git $(git --version | awk '{print $3}')"
}

# ---------- 3. 拉取代码 ----------
clone_repo() {
    if [ -d "$INSTALL_DIR/.git" ]; then
        info "更新代码..."
        cd "$INSTALL_DIR"
        git pull --rebase
    else
        info "克隆仓库..."
        rm -rf "$INSTALL_DIR"
        git clone "$REPO_URL" "$INSTALL_DIR"
        cd "$INSTALL_DIR"
    fi
    ok "代码就绪"
}

# ---------- 4. 构建前端 ----------
build_frontend() {
    info "构建前端..."
    cd "$INSTALL_DIR/frontend"
    npm install --registry=https://registry.npmmirror.com 2>&1 | tail -1
    npm run build 2>&1 | tail -3
    if [ ! -d "$INSTALL_DIR/frontend/dist" ] || [ -z "$(ls -A $INSTALL_DIR/frontend/dist)" ]; then
        err "前端构建失败"; exit 1
    fi
    ok "前端构建完成 ($(du -sh $INSTALL_DIR/frontend/dist | awk '{print $1}'))"
}

# ---------- 5. 构建后端 ----------
build_backend() {
    info "构建后端..."
    cd "$INSTALL_DIR/backend"
    export GOPROXY=https://goproxy.cn,direct
    CGO_ENABLED=0 GOOS=linux go build -o submanager .
    ok "后端构建完成"
}

# ---------- 6. 写入配置 ----------
write_config() {
    cat > "$INSTALL_DIR/.env" <<EOF
APP_ENV=production
APP_PORT=${BACKEND_PORT}
APP_SECRET=${APP_SECRET}
INIT_TOKEN=${INIT_TOKEN}
DB_DRIVER=sqlite
DB_DSN=${INSTALL_DIR}/data/submanager.db
SUB_BASE_URL=${SUB_BASE_URL}
ADMIN_USERNAME=${ADMIN_USER}
ADMIN_PASSWORD=${ADMIN_PASS}
ALLOW_REGISTER=${ALLOW_REGISTER}
AGENT_REPORT_INTERVAL=60
AGENT_OFFLINE_TIMEOUT=180
SUB_REFRESH_INTERVAL=30
EOF
    ok "配置文件已写入 $INSTALL_DIR/.env"
}

# ---------- 7. systemd 服务 ----------
install_service() {
    mkdir -p "$INSTALL_DIR/data"

    cat > /etc/systemd/system/subboard.service <<EOF
[Unit]
Description=SubBoard Panel
After=network.target

[Service]
Type=simple
WorkingDirectory=${INSTALL_DIR}/backend
EnvironmentFile=${INSTALL_DIR}/.env
ExecStart=${INSTALL_DIR}/backend/submanager
Restart=always
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable subboard >/dev/null 2>&1
    systemctl restart subboard
    ok "subboard 服务已启动"
}

# ---------- 8. 等待就绪 ----------
wait_ready() {
    info "等待后端就绪..."
    for i in $(seq 1 30); do
        if curl -sf "http://127.0.0.1:${BACKEND_PORT}/api/health" >/dev/null 2>&1; then
            ok "后端已就绪"
            return
        fi
        sleep 1
    done
    warn "后端未在 30s 内就绪，请检查: journalctl -u subboard -f"
}

# ---------- 9. 输出 1Panel 配置指引 ----------
show_1panel_guide() {
    if [ -z "$DOMAIN" ]; then
        ACCESS_URL="http://$(curl -sf ifconfig.me 2>/dev/null || hostname -I | awk '{print $1}')"
    else
        PROTO="http"
        [ "$ENABLE_HTTPS" = "y" ] && PROTO="https"
        ACCESS_URL="${PROTO}://${DOMAIN}"
    fi

    echo ""
    echo -e "${GREEN}============================================${NC}"
    echo -e "${GREEN}  部署完成!${NC}"
    echo -e "${GREEN}============================================${NC}"
    echo ""
    echo -e "  访问地址:    ${CYAN}${ACCESS_URL}${NC}"
    echo -e "  管理后台:    ${CYAN}${ACCESS_URL}/admin${NC}"
    echo -e "  管理员:      ${CYAN}${ADMIN_USER}${NC}"
    echo -e "  密码:        ${CYAN}${ADMIN_PASS}${NC}"
    echo ""
    echo -e "  ${YELLOW}请在 1Panel 中完成以下配置:${NC}"
    echo ""
    echo -e "  ${CYAN}方式一: 网站列表 → 创建网站 → 静态网站${NC}"
    echo "    1. 网站目录设为: ${INSTALL_DIR}/frontend/dist"
    echo "    2. 在网站设置 → 反向代理 中添加:"
    echo "       /api/  →  http://127.0.0.1:${BACKEND_PORT}"
    echo "       /sub/   →  http://127.0.0.1:${BACKEND_PORT}"
    echo ""
    echo -e "  ${CYAN}方式二: 网站列表 → 创建网站 → 反向代理${NC}"
    echo "    1. 代理地址: http://127.0.0.1:${BACKEND_PORT}"
    echo "    2. 在网站设置 → 网站目录 中:"
    echo "       将 ${INSTALL_DIR}/frontend/dist 映射到 /"
    echo ""
    echo -e "  ${CYAN}常用命令:${NC}"
    echo "    查看日志:  journalctl -u subboard -f"
    echo "    重启服务:  systemctl restart subboard"
    echo "    停止服务:  systemctl stop subboard"
    echo "    查看状态:  systemctl status subboard"
    echo ""
    echo -e "  ${CYAN}更新升级:${NC}"
    echo "    cd ${INSTALL_DIR} && git pull"
    echo "    bash deploy-1panel.sh  # 重新构建并部署"
    echo ""
    echo -e "${GREEN}============================================${NC}"
}

# ========== 主流程 ==========
install_deps
clone_repo
build_frontend
build_backend
write_config
install_service
wait_ready
show_1panel_guide
