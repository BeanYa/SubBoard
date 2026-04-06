#!/bin/bash
set -e

COMPOSE_FILE="docker-compose.yml"
ENV_FILE=".env"

# Color definitions
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
RED='\033[0;31m'
NC='\033[0m'

info() { echo -e "${CYAN}[INFO]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
success() { echo -e "${GREEN}[OK]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; }

echo -e ""
echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}  SubManager 代理订阅管理平台 部署脚本   ${NC}"
echo -e "${CYAN}========================================${NC}"
echo -e ""

# 1. Check environment
check_docker() {
    info "检查 Docker 环境..."
    if ! command -v docker &> /dev/null; then
        error "Docker 未安装，请先安装 Docker"
        exit 1
    fi
    if ! docker compose version &> /dev/null; then
        error "Docker Compose 未安装"
        exit 1
    fi
    success "Docker 环境检查通过"
}

# 2. Load existing config or generate new
load_or_generate_config() {
    if [ -f "$ENV_FILE" ]; then
        info "检测到已有 .env 配置文件"
        read -p "是否使用现有配置？(y/n) [n]: " use_existing
        if [ "$use_existing" != "y" ] && [ "$use_existing" != "Y" ]; then
            backup_env
            generate_new_config
        else
            info "使用现有配置"
            source "$ENV_FILE"
        fi
    else
        generate_new_config
    fi
}

backup_env() {
    local backup="${ENV_FILE}.backup.$(date +%Y%m%d%H%M%S)"
    info "备份现有配置到 $backup"
    cp "$ENV_FILE" "$backup"
}

# 3. Generate new configuration
generate_new_config() {
    echo -e ""
    echo -e "${CYAN}--- 配置参数 ---${NC}"

    read -p "外部 HTTP 端口 [80]: " HTTP_PORT
    HTTP_PORT=${HTTP_PORT:-80}

    read -p "外部 HTTPS 端口 [443]: " HTTPS_PORT
    HTTPS_PORT=${HTTPS_PORT:-443}

    APP_SECRET=$(openssl rand -hex 32 2>/dev/null || head -c 32 /dev/urandom | xxd -p)
    echo "生成的 APP_SECRET: ${APP_SECRET:0:20}..."

    DB_PASSWORD=$(openssl rand -hex 16 2>/dev/null || head -c 16 /dev/urandom | xxd -p)
    echo "生成的 DB_PASSWORD: ${DB_PASSWORD:0:16}..."

    DEFAULT_IP=$(hostname -I 2>/dev/null | awk '{print $1}' || echo "localhost")
    read -p "订阅基础 URL [http://${DEFAULT_IP}:${HTTP_PORT}]: " SUB_BASE_URL
    SUB_BASE_URL=${SUB_BASE_URL:-http://${DEFAULT_IP}:${HTTP_PORT}}

    read -p "管理员用户名 [admin]: " ADMIN_USERNAME
    ADMIN_USERNAME=${ADMIN_USERNAME:-admin}

    read -p "管理员密码 [admin]: " ADMIN_PASSWORD
    ADMIN_PASSWORD=${ADMIN_PASSWORD:-admin}
    if [ "$ADMIN_PASSWORD" == "admin" ]; then
        warn "使用默认密码 admin，建议生产环境修改"
    else
        echo "设置的 ADMIN_PASSWORD: ${ADMIN_PASSWORD}"
    fi

    INIT_TOKEN=$(openssl rand -hex 16 2>/dev/null || head -c 16 /dev/urandom | xxd -p)

    read -p "允许用户注册 (true/false) [true]: " ALLOW_REGISTER
    ALLOW_REGISTER=${ALLOW_REGISTER:-true}

    read -p "Agent 上报间隔(秒) [60]: " AGENT_REPORT_INTERVAL
    AGENT_REPORT_INTERVAL=${AGENT_REPORT_INTERVAL:-60}

    read -p "Agent 离线判定超时(秒) [180]: " AGENT_OFFLINE_TIMEOUT
    AGENT_OFFLINE_TIMEOUT=${AGENT_OFFLINE_TIMEOUT:-180}

    read -p "订阅源自动刷新间隔(分钟) [30]: " SUB_REFRESH_INTERVAL
    SUB_REFRESH_INTERVAL=${SUB_REFRESH_INTERVAL:-30}

    echo -e ""
    echo -e "${CYAN}--- 配置摘要 ---${NC}"
    echo "  HTTP 端口:         $HTTP_PORT"
    echo "  HTTPS 端口:         $HTTPS_PORT"
    echo "  订阅基础 URL:       $SUB_BASE_URL"
    echo "  管理员用户名:       $ADMIN_USERNAME"
    echo "  管理员密码:         $ADMIN_PASSWORD"
    echo "  允许注册:           $ALLOW_REGISTER"
    echo -e "${CYAN}------------------${NC}"

    read -p "确认配置无误？(y/n) [y]: " confirm
    if [ "$confirm" != "" ] && [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
        error "取消部署"
        exit 1
    fi
}

# 4. Write .env file
write_env() {
    cat > "$ENV_FILE" <<EOF
# SubManager Configuration
HTTP_PORT=${HTTP_PORT}
HTTPS_PORT=${HTTPS_PORT}
APP_SECRET=${APP_SECRET}
INIT_TOKEN=${INIT_TOKEN}
DB_PASSWORD=${DB_PASSWORD}
DB_NAME=submanager
DB_USER=submgr
SUB_BASE_URL=${SUB_BASE_URL}
ADMIN_USERNAME=${ADMIN_USERNAME}
ADMIN_PASSWORD=${ADMIN_PASSWORD}
ALLOW_REGISTER=${ALLOW_REGISTER}
AGENT_REPORT_INTERVAL=${AGENT_REPORT_INTERVAL}
AGENT_OFFLINE_TIMEOUT=${AGENT_OFFLINE_TIMEOUT}
SUB_REFRESH_INTERVAL=${SUB_REFRESH_INTERVAL}

# SSL (optional - leave empty for HTTP only)
SSL_CERT_PATH=
SSL_KEY_PATH=
EOF
    success "配置文件已写入 $ENV_FILE"
}

# 5. Generate self-signed SSL certificate (optional)
generate_ssl() {
    read -p "是否生成自签名 SSL 证书？(y/n) [n]: " generate_ssl_cert
    if [ "$generate_ssl_cert" == "y" ] || [ "$generate_ssl_cert" == "Y" ]; then
        mkdir -p nginx/ssl
        openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
            -keyout nginx/ssl/key.pem \
            -out nginx/ssl/cert.pem \
            -subj "/C=CN/ST=Beijing/L=Beijing/O=SubManager/CN=submanager" 2>/dev/null
        sed -i "s|SSL_CERT_PATH=|SSL_CERT_PATH=./nginx/ssl/cert.pem|" "$ENV_FILE"
        sed -i "s|SSL_KEY_PATH=|SSL_KEY_PATH=./nginx/ssl/key.pem|" "$ENV_FILE"
        success "SSL 证书已生成"
    fi
}

# 6. Build and deploy
build_and_deploy() {
    echo -e ""
    info "构建 Docker 镜像..."
    docker compose build --no-cache

    info "启动服务..."
    docker compose up -d

    info "等待服务就绪..."
    local retries=0
    while [ $retries -lt 60 ]; do
        if curl -sf "http://localhost:${HTTP_PORT}/api/health" &> /dev/null; then
            success "后端服务已就绪"
            break
        fi
        retries=$((retries + 1))
        sleep 2
        if [ $((retries % 10)) -eq 0 ]; then
            warn "等待中... ($retries/60)"
        fi
    done

    if [ $retries -ge 60 ]; then
        error "服务启动超时，请检查日志: docker compose logs -f"
        exit 1
    fi
}

# 7. Initialize admin
init_admin() {
    echo -e ""
    info "初始化管理员账户..."
    local result
    result=$(curl -sf -X POST "http://localhost:${HTTP_PORT}/api/system/init" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"${ADMIN_USERNAME}\",\"password\":\"${ADMIN_PASSWORD}\",\"init_token\":\"${INIT_TOKEN}\"}" 2>&1) || true

    if echo "$result" | grep -q "success\|created\|ok"; then
        success "管理员初始化成功"
    else
        warn "管理员可能已存在或初始化失败: $result"
    fi
}

# 8. Show info
show_info() {
    local protocol="http"
    if [ -n "$SSL_CERT_PATH" ] && [ -f "$SSL_CERT_PATH" ]; then
        protocol="https"
    fi

    echo -e ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  部署完成!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo -e ""
    echo -e "  访问地址:   ${CYAN}${protocol}://localhost:${HTTP_PORT}${NC}"
    echo -e "  管理后台:   ${CYAN}${protocol}://localhost:${HTTP_PORT}/admin${NC}"
    echo -e "  管理员:     ${CYAN}${ADMIN_USERNAME}${NC}"
    echo -e "  密码:       ${CYAN}${ADMIN_PASSWORD}${NC}"
    echo -e ""
    echo -e "  常用命令:"
    echo -e "    查看日志:     ${CYAN}docker compose logs -f${NC}"
    echo -e "    查看状态:     ${CYAN}docker compose ps${NC}"
    echo -e "    停止服务:     ${CYAN}docker compose down${NC}"
    echo -e "    重启服务:     ${CYAN}docker compose restart${NC}"
    echo -e "    重新构建:     ${CYAN}docker compose build && docker compose up -d${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo -e ""
    echo -e "${YELLOW}首次登录后请前往 管理后台 -> 用户管理 创建普通用户${NC}"
    echo -e "${YELLOW}然后在 管理后台 -> 套餐管理 创建套餐并分配给用户${NC}"
    echo -e ""
}

# Main execution
check_docker
load_or_generate_config
write_env
generate_ssl
build_and_deploy
init_admin
show_info