#!/bin/bash
set -e

# ================================================
# SubManager 一键安装脚本 (含 git clone)
# 用法: bash <(curl -sL https://raw.githubusercontent.com/BeanYa/SubBoard/master/install.sh)
# =============================================

REPO_URL="https://github.com/BeanYa/SubBoard.git"
REPO_BRANCH="master"
INSTALL_DIR="/opt/subboard"

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
echo -e "${CYAN}  SubManager 代理订阅管理平台 一键安装  ${NC}"
echo -e "${CYAN}========================================${NC}"
echo -e ""

# 1. Check Docker
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

# 2. Clone repo
clone_repo() {
    info "克隆仓库..."
    if [ -d "$INSTALL_DIR" ]; then
        warn "检测到已存在安装目录: $INSTALL_DIR"
        read -p "是否重新克隆覆盖？(y/n) [n]: " overwrite
        if [ "$overwrite" != "y" ] && [ "$overwrite" != "Y" ]; then
            info "使用现有目录继续"
            cd "$INSTALL_DIR"
            return
        fi
        warn "删除旧目录..."
        rm -rf "$INSTALL_DIR"
    fi

    info "克隆 SubBoard 仓库..."
    git clone -b "$REPO_BRANCH" --depth=1 "$REPO_URL" "$INSTALL_DIR"
    cd "$INSTALL_DIR"
    success "仓库克隆完成"
}

# 3. Run deploy
run_deploy() {
    chmod +x deploy.sh
    ./deploy.sh
}

# Main
check_docker
clone_repo
run_deploy
