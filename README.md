# SubManager

自部署的代理订阅自助管理平台。支持 Sing-box / Clash / V2Ray (VLESS/VMess/Shadowsocks/Trojan) 多格式订阅输出，管理员通过后台组合 SubStore 聚合订阅和自建 VPS Agent 上报节点形成服务群，为用户分配套餐。

## 功能特性

- **多格式订阅输出** — Clash YAML / Sing-box JSON / Base64，支持 User-Agent 自动识别
- **SubStore 聚合** — 导入 SubStore 或任意订阅链接，自动解析同步节点
- **VPS Agent 上报** — 自建服务器通过 Agent API 上报状态、流量和节点信息
- **服务群组合** — 管理员灵活组合订阅源和 Agent 节点形成服务群
- **套餐管理** — 管理员手动为用户分配套餐，控制流量限额和有效期
- **自助订阅** — 用户登录后一键获取订阅链接，支持多客户端格式
- **Docker 一键部署** — 提供部署脚本，交互式配置参数后一键拉起

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go + Gin + GORM |
| 前端 | React 18 + Vite + TypeScript + shadcn/ui + Tailwind CSS |
| 数据库 | SQLite（开发）/ PostgreSQL（生产）|
| 认证 | JWT（前端登录）+ Token（订阅链接）|
| 部署 | Docker Compose + Nginx |

## 快速开始

### 开发环境

```bash
# 后端
cd backend
go mod tidy
go run main.go

# 前端
cd frontend
npm install
npm run dev
```

### 生产部署

```bash
chmod +x deploy.sh
./deploy.sh
```

部署脚本会交互式引导配置以下参数：

| 参数 | 说明 | 默认值 |
|------|------|--------|
| HTTP_PORT | 外部 HTTP 端口 | 80 |
| APP_SECRET | JWT 密钥 | 自动生成 |
| DB_PASSWORD | 数据库密码 | 自动生成 |
| SUB_BASE_URL | 订阅基础 URL | 自动检测 |
| ADMIN_USERNAME | 管理员用户名 | admin |
| ADMIN_PASSWORD | 管理员密码 | 自动生成 |
| ALLOW_REGISTER | 允许用户注册 | true |

所有容器使用 `submgr-` 前缀隔离：
- `submgr-nginx` — Nginx 反向代理
- `submgr-app` — Go 后端
- `submgr-postgres` — PostgreSQL 数据库

### 常用管理命令

```bash
docker compose ps                           # 查看状态
docker compose logs -f submgr-app           # 后端日志
docker compose down                         # 停止（保留数据）
docker compose build --no-cache && docker compose up -d  # 重建更新
```

## 项目结构

```
submanager/
├── deploy.sh                  # 一键部署脚本
├── docker-compose.yml         # 容器编排
├── nginx/default.conf         # Nginx 反向代理配置
├── backend/                   # Go 后端
├── frontend/                  # React 前端
├── PLAN.md                    # 详细实施计划
└── README.md
```

## API 概览

### 用户端
- `POST /api/auth/login` — 登录
- `POST /api/auth/register` — 注册
- `GET /api/user/profile` — 用户信息
- `GET /api/user/subscription` — 订阅详情

### 订阅输出
- `GET /sub/:token?format=clash` — Clash YAML
- `GET /sub/:token?format=singbox` — Sing-box JSON
- `GET /sub/:token?format=base64` — Base64 节点列表

### Agent 上报
- `POST /api/agent/report` — 心跳 + 流量 + 节点上报
- `GET /api/agent/config/:token` — 拉取 Agent 配置

### 管理员
- `/api/admin/users` — 用户管理
- `/api/admin/plans` — 套餐管理
- `/api/admin/subscriptions` — 订阅源管理
- `/api/admin/groups` — 服务群管理
- `/api/admin/agents` — Agent 管理

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| APP_ENV | 环境 | development |
| APP_PORT | 后端端口 | 8080 |
| APP_SECRET | JWT 密钥 | - |
| DB_DRIVER | 数据库驱动 | sqlite |
| DB_DSN | 数据库连接串 | submanager.db |
| SUB_BASE_URL | 订阅基础 URL | http://localhost:8080 |
| ALLOW_REGISTER | 允许注册 | true |
| AGENT_REPORT_INTERVAL | Agent 上报间隔(秒) | 60 |

## 许可证

MIT
