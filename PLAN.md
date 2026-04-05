# SubManager - 代理订阅自助管理平台 实施计划

## Context
构建一个自部署的代理订阅管理系统，用户可通过网站自助获取订阅链接（支持 singbox/clash/v2ray 等格式），管理员通过后台管理订阅源（SubStore 聚合 + 自建 VPS Agent 上报）组成服务群，为用户分配套餐。

## 技术栈
- **后端**: Go + Gin + GORM
- **前端**: React 18 + Vite + TypeScript + shadcn/ui + Tailwind CSS（使用 frontend-design skill 生成高质量 UI）
- **数据库**: SQLite（开发阶段）→ PostgreSQL（生产 Docker 部署）
- **测试**: 后端 Go 单元测试 + 前端 Vitest 单元测试 + 集成测试
- **部署**: Docker Compose
- **认证**: 前端 JWT 登录 + 订阅链接 Token 认证

---

## 项目结构

```
submanager/
├── docker-compose.yml
├── backend/
│   ├── Dockerfile
│   ├── go.mod
│   ├── main.go
│   ├── config/
│   │   └── config.go          # 配置加载（环境变量，支持 SQLite/PG 切换）
│   ├── model/
│   │   └── models.go          # 数据模型（GORM）
│   ├── handler/
│   │   ├── auth.go            # 登录注册
│   │   ├── user.go            # 用户订阅面板 API
│   │   ├── admin_user.go      # 管理员-用户管理
│   │   ├── admin_plan.go      # 管理员-套餐管理
│   │   ├── admin_sub.go       # 管理员-订阅源管理
│   │   ├── admin_group.go     # 管理员-服务群管理
│   │   ├── admin_agent.go     # 管理员-Agent管理
│   │   ├── subscribe.go       # 订阅链接输出（/sub/:token）
│   │   └── agent.go           # Agent上报接口
│   ├── service/
│   │   ├── subscription.go    # 订阅解析转换（vless/vmess/ss/trojan → 统一格式）
│   │   ├── converter_clash.go # Clash YAML 输出
│   │   ├── converter_singbox.go # Sing-box JSON 输出
│   │   ├── converter_base64.go  # Base64 节点列表输出
│   │   └── node_filter.go     # 节点过滤/套餐匹配
│   ├── middleware/
│   │   ├── auth.go            # JWT 认证中间件
│   │   ├── admin.go           # 管理员权限中间件
│   │   └── token.go           # 订阅 Token 认证中间件
│   ├── router/
│   │   └── router.go          # 路由注册
│   └── tests/
│       └── router_test.go     # 路由集成测试
├── frontend/
│   ├── Dockerfile
│   ├── package.json
│   ├── vite.config.ts
│   ├── tailwind.config.js
│   ├── tsconfig.json
│   ├── components.json           # shadcn/ui 配置
│   ├── src/
│   │   ├── App.tsx
│   │   ├── main.tsx
│   │   ├── index.css             # Tailwind 入口
│   │   ├── lib/
│   │   │   └── utils.ts          # cn() 等工具函数
│   │   ├── api/
│   │   │   ├── client.ts         # axios 实例 + 拦截器
│   │   │   ├── auth.ts           # 认证 API
│   │   │   ├── user.ts           # 用户 API
│   │   │   └── admin.ts          # 管理员 API
│   │   ├── hooks/
│   │   │   ├── useAuth.ts        # 认证状态 hook
│   │   │   └── useApi.ts         # 通用请求 hook
│   │   ├── pages/
│   │   │   ├── Login.tsx         # 登录页
│   │   │   ├── Dashboard.tsx     # 用户面板
│   │   │   ├── Subscribe.tsx     # 订阅链接页
│   │   │   └── admin/
│   │   │       ├── Layout.tsx    # 管理员布局（侧边栏）
│   │   │       ├── Users.tsx     # 用户管理
│   │   │       ├── Plans.tsx     # 套餐管理
│   │   │       ├── Subscriptions.tsx  # 订阅源管理
│   │   │       ├── ServiceGroups.tsx  # 服务群管理
│   │   │       └── Agents.tsx    # Agent管理
│   │   └── components/
│   │       ├── ui/               # shadcn/ui 组件（button, card, table, dialog...）
│   │       ├── Layout.tsx        # 用户端布局
│   │       ├── PrivateRoute.tsx  # 路由守卫
│   │       └── TrafficChart.tsx  # 流量可视化组件
│   └── index.html
├── nginx/
│   └── default.conf             # Nginx 反向代理配置
└── README.md
```

---

## 数据模型

> **设计原则**：关联关系使用外键 + 关联表，避免 JSON 数组存储 ID 列表（确保引用完整性和级联删除）。用户流量限额实时从关联套餐计算，避免双重存储不一致。

### users（用户）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| username | string(64) | 用户名（唯一索引）|
| password_hash | string(256) | bcrypt 加密密码 |
| plan_id | *uint | 关联套餐（外键 → plans.id，ON DELETE SET NULL）|
| sub_token | string(32) | 订阅令牌（唯一索引，自动生成）|
| traffic_used | int64 | 已用流量（字节）|
| expire_at | *time | 到期时间 |
| traffic_reset_at | *time | 流量周期重置时间（套餐 duration_days 到期后重置 traffic_used）|
| is_admin | bool | 是否管理员（默认 false）|
| enabled | bool | 是否启用（默认 true）|
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

> **流量限额**：`traffic_limit` 不存储在 users 表中，而是通过 `user.plan_id → plan.traffic_limit` 实时查询。好处：修改套餐限额立即生效，无需批量同步。

### plans（套餐）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| name | string(64) | 套餐名称（如"基础版"、"高级版"）|
| description | string(256) | 套餐描述 |
| traffic_limit | int64 | 流量限额（字节）|
| duration_days | int | 有效天数（0=永久）|
| price | string | 价格展示（如"¥10/月"，仅展示用）|
| enabled | bool | 是否启用 |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

### subscription_sources（订阅源）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| name | string(64) | 源名称 |
| type | string(16) | substore / url / raw |
| url | string(512) | 订阅 URL（SubStore 或原始链接）|
| headers | json(map) | 请求头（User-Agent 等）|
| refresh_interval | int | 自动刷新间隔（分钟，0=不自动）|
| node_count | int | 节点数量（自动更新）|
| last_fetch_at | *time | 最后拉取时间 |
| fetch_error | string | 最近拉取错误信息 |
| enabled | bool | 是否启用 |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

> **SSRF 防护**：后端在拉取订阅 URL 时，需校验目标 IP 不在内网范围（10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, 127.0.0.0/8, 169.254.0.0/16, [::1]/128），拒绝指向本地/链路层/云元数据地址的请求。

> **raw 类型说明**：当 `type=raw` 时，`url` 为空，节点数据通过管理后台手动粘贴 Base64 编码的节点列表，后端直接解析写入 node_cache。

### service_groups（服务群）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| name | string(64) | 群名称 |
| description | string(256) | 群描述 |
| sort_order | int | 排序权重 |
| enabled | bool | 是否启用 |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

### 关联表（替代 JSON 数组）

#### plan_service_groups（套餐 ↔ 服务群）
| 字段 | 类型 | 说明 |
|------|------|------|
| plan_id | uint | 外键 → plans.id（ON DELETE CASCADE）|
| service_group_id | uint | 外键 → service_groups.id（ON DELETE CASCADE）|

**联合主键**: `(plan_id, service_group_id)`

#### group_subscription_sources（服务群 ↔ 订阅源）
| 字段 | 类型 | 说明 |
|------|------|------|
| service_group_id | uint | 外键 → service_groups.id（ON DELETE CASCADE）|
| subscription_source_id | uint | 外键 → subscription_sources.id（ON DELETE CASCADE）|

**联合主键**: `(service_group_id, subscription_source_id)`

#### group_agents（服务群 ↔ Agent）
| 字段 | 类型 | 说明 |
|------|------|------|
| service_group_id | uint | 外键 → service_groups.id（ON DELETE CASCADE）|
| agent_id | uint | 外键 → agents.id（ON DELETE CASCADE）|

**联合主键**: `(service_group_id, agent_id)`

> **设计优势**：删除套餐/服务群/订阅源/Agent 时，数据库自动清理关联关系，不会产生孤儿数据。

### agents（VPS Agent）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| name | string(64) | Agent 名称（如"HK-Node-01"）|
| token | string(32) | 上报认证 Token（唯一，自动生成）|
| server_addr | string(128) | 服务器地址 |
| port | int | 端口 |
| protocol | string(16) | vless / vmess / shadowsocks / snell / trojan |
| protocol_config | json(map) | 协议参数（见下方协议参数 Schema）|
| traffic_used | int64 | 已用流量 |
| traffic_total | int64 | 总流量 |
| cpu_usage | float64 | CPU 使用率（Agent 上报）|
| mem_usage | float64 | 内存使用率（Agent 上报）|
| status | string(16) | online / offline / unknown |
| last_report_at | *time | 最后上报时间 |
| enabled | bool | 是否启用 |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

### protocol_config Schema（按协议）

```jsonc
// vless
{ "uuid": "string(必填)", "transport": "ws|grpc|tcp(必填)",
  "path": "string", "host": "string", "sni": "string",
  "tls": "bool", "flow": "xtls-rprx-vision|空" }

// vmess
{ "uuid": "string(必填)", "alterId": "int(默认0)",
  "security": "auto|aes-128-gcm|chacha20-poly1305(必填)",
  "transport": "ws|grpc|tcp(必填)", "path": "string",
  "host": "string", "sni": "string", "tls": "bool" }

// shadowsocks
{ "method": "aes-256-gcm|chacha20-ietf-poly1305|...(必填)",
  "password": "string(必填)", "plugin": "obfs-local|v2ray-plugin|空",
  "plugin_opts": "string" }

// trojan
{ "password": "string(必填)", "sni": "string", "transport": "ws|grpc|tcp",
  "path": "string", "host": "string" }

// snell
{ "psk": "string(必填)", "version": "1|2|3(必填)",
  "obfs": "http|tls|空", "obfs_host": "string" }
```

### node_cache（节点缓存）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| source_type | string(16) | subscription / agent |
| source_id | uint | 来源 ID（订阅源 ID 或 Agent ID）|
| name | string(128) | 节点名称 |
| raw_link | string(1024) | 原始完整链接 |
| protocol | string(16) | 协议类型 |
| address | string(128) | 服务器地址 |
| port | int | 端口 |
| extra | json(map) | 协议扩展参数（UUID/加密/传输层等）|
| updated_at | time | 更新时间 |

**索引**: `(source_type, source_id)` 组合索引

> **缓存清理策略**：
> 1. 每次订阅源刷新时，先删除 `source_type="subscription", source_id=<该源ID>` 的全部旧缓存，再写入新数据（全量替换）
> 2. Agent 上报时同理，全量替换该 Agent 的缓存节点
> 3. 删除订阅源或 Agent 时，CASCADE 清理对应缓存
> 4. 无需 TTL——节点缓存始终与来源同步，不存在"过期但未清理"的中间态

---

## API 设计

### 用户端
```
POST   /api/auth/login              # 登录 → JWT
POST   /api/auth/register           # 注册（可通过环境变量关闭）
GET    /api/user/profile            # 当前用户信息（JWT）
GET    /api/user/subscription       # 订阅详情（套餐、流量、节点数预览）
PUT    /api/user/password           # 修改密码
```

### 系统初始化（仅首次部署，自动执行）
```
POST   /api/system/init            # 初始化管理员（仅无管理员时可用）
   Body: { "username": "admin", "password": "xxx", "init_token": "<INIT_TOKEN>" }
   说明: init_token 从环境变量 INIT_TOKEN 读取，部署脚本自动注入后清除。
         若数据库已存在管理员账户，此接口返回 409 拒绝执行。
```

> **安全设计**：不再使用无保护的 `/api/admin/init` 接口。改为：
> 1. 部署脚本生成随机 `INIT_TOKEN` 写入 .env
> 2. 后端启动时读取 `INIT_TOKEN`，初始化完成后从内存中清除
> 3. `POST /api/system/init` 必须携带正确的 init_token 且数据库无管理员时才生效
> 4. 成功后该接口永久返回 409（已有管理员）

### 订阅输出（Token 认证，无需登录）
```
GET    /sub/:token                  # 根据 ?format= 参数返回（默认 clash）
GET    /sub/:token?format=clash     # Clash YAML 格式
GET    /sub/:token?format=singbox   # Sing-box JSON 格式
GET    /sub/:token?format=base64    # Base64 节点链接列表
GET    /sub/:token?format=raw       # 原始节点链接（每行一个）
```

**User-Agent 智能识别**: 根据 UA 自动返回对应格式
- clash → Clash YAML
- sing-box → Sing-box JSON
- 其他 → Base64

### 管理员（需 JWT + admin 权限）
```
# 用户管理
GET    /api/admin/users             # 用户列表（分页/搜索）
POST   /api/admin/users             # 创建用户
PUT    /api/admin/users/:id         # 更新用户
DELETE /api/admin/users/:id         # 删除用户
POST   /api/admin/users/:id/plan   # 分配套餐（body: plan_id）
POST   /api/admin/users/:id/reset  # 重置流量
POST   /api/admin/users/:id/toggle # 启用/禁用用户

# 套餐管理
GET    /api/admin/plans
POST   /api/admin/plans
PUT    /api/admin/plans/:id
DELETE /api/admin/plans/:id

# 订阅源管理
GET    /api/admin/subscriptions
POST   /api/admin/subscriptions
PUT    /api/admin/subscriptions/:id
DELETE /api/admin/subscriptions/:id
POST   /api/admin/subscriptions/:id/refresh  # 手动刷新节点

# 服务群管理
GET    /api/admin/groups
POST   /api/admin/groups
PUT    /api/admin/groups/:id
DELETE /api/admin/groups/:id

# Agent 管理
GET    /api/admin/agents
POST   /api/admin/agents
PUT    /api/admin/agents/:id
DELETE /api/admin/agents/:id
```

### Agent 上报接口（Token 认证）
```
POST   /api/agent/report
  Headers: Authorization: Bearer <agent_token>
  Body: {
    "traffic_used": 1234567890,
    "traffic_total": 9999999999,
    "cpu_usage": 23.5,
    "mem_usage": 45.2,
    "status": "online",
    "nodes": [
      { "name": "HK-01", "protocol": "vless", "address": "1.2.3.4",
        "port": 443, "extra": { "uuid": "xxx", "transport": "ws", ... } }
    ]
  }
  Response: { "ok": true, "interval": 60 }  # interval 建议下次上报间隔（秒）

GET    /api/agent/config/:token
  Response: {
    "name": "HK-Node-01",
    "protocol": "vless",
    "server_addr": "0.0.0.0",
    "port": 443,
    "protocol_config": { ... },
    "report_interval": 60
  }
```

---

## 核心流程

### 1. 订阅生成流程
```
用户客户端 → GET /sub/:token?format=clash
    ↓
Token 中间件 → 查询 users 表 → 验证 enabled/expire_at/traffic
    ↓
查询 user.plan_id → plans.traffic_limit（实时计算流量限额，无需同步）
    ↓
查询 plan_service_groups → 获取该套餐关联的 service_group 列表
    ↓
查询 group_subscription_sources + group_agents → 获取各服务群关联的订阅源和 Agent
    ↓
从 node_cache 聚合所有节点
    ↓
按 format 参数调用对应 converter
    ↓
返回 Clash YAML / Sing-box JSON / Base64
```

### 2. 订阅源同步流程
```
管理员添加 SubStore URL → 保存 subscription_sources
    ↓
定时任务 / 手动触发 → 校验 URL（SSRF 防护：拒绝内网/链路层/元数据地址）
    ↓
HTTP GET 拉取内容 → Base64 解码 → 逐行解析 vless:// vmess:// ss:// trojan:// 链接
    ↓
解析为统一 Node 结构 → 删除该源旧缓存（DELETE node_cache WHERE source_type="subscription" AND source_id=X）
    ↓
写入新缓存到 node_cache
    ↓
更新 subscription_sources.node_count / last_fetch_at
```

### 3. Agent 上报流程
```
Agent 进程 → POST /api/agent/report（带 token）
    ↓
Token 中间件 → 查询 agents 表 → 验证 token
    ↓
更新 agents 表（traffic_used, cpu_usage, status, last_report_at）
    ↓
同步 agent 节点到 node_cache（source_type="agent"）
    ↓
返回 { ok: true, interval: 60 }
```

---

## 测试策略

### 后端单元测试
- **框架**: Go 标准 testing + testify/assert
- **数据库**: SQLite `:memory:` 每个测试独立建表
- **覆盖层**:
  - `model/` → GORM 模型 CRUD 测试
  - `handler/` → HTTP handler 测试（httptest.NewRecorder）
  - `service/` → 协议解析、格式转换测试（核心逻辑）
  - `middleware/` → JWT/Token 中间件测试
- **运行**: `go test ./... -v -cover`

### 后端集成测试
- 完整 Gin Engine + SQLite 内存数据库
- 模拟完整用户流程：注册 → 登录 → 管理员创建套餐 → 分配 → 获取订阅
- Agent 上报完整流程测试

### 前端单元测试
- **框架**: Vitest + React Testing Library
- **API Mock**: MSW (Mock Service Worker)
- **覆盖层**:
  - 组件渲染测试
  - 用户交互测试（表单提交、按钮点击）
  - API 调用测试
- **运行**: `npm test`

### 前端集成测试
- MSW mock 完整 API
- 端到端用户流程：登录 → 查看订阅 → 复制链接
- 管理员流程：创建套餐 → 添加订阅源 → 创建服务群 → 分配用户

---

## 实施步骤

### Phase 1: 后端（Step 1-5）
1. **基础框架** — 项目初始化、配置、数据模型、SQLite 迁移
2. **认证 + 管理员 CRUD** — JWT、中间件、全部管理接口 + 单元测试
3. **订阅源解析** — 协议解析器、节点缓存 + 单元测试
4. **订阅输出** — 多格式转换器 + 单元测试
5. **Agent 上报** — 上报接口 + 单元测试

### Phase 2: 前端（Step 6-8）
6. **基础搭建** — Vite + React + shadcn/ui + 路由 + API 封装 + 测试框架
7. **用户端页面** — 使用 frontend-design skill 生成高质量 UI
8. **管理员页面** — 使用 frontend-design skill 生成高质量 UI

### Phase 3: 集成与部署（Step 9-10）
9. **集成测试** — 前后端联调、完整流程验证
10. **Docker 部署** — Dockerfile、docker-compose.yml、Nginx 配置

---

## 环境变量配置

```env
# 通用
APP_ENV=development          # development / production
APP_PORT=8080
APP_SECRET=your-secret-key   # JWT 签名密钥
INIT_TOKEN=                  # 管理员初始化令牌（部署时随机生成，初始化后失效）

# 数据库
DB_DRIVER=sqlite             # sqlite / postgres
DB_DSN=submanager.db         # SQLite: 文件路径; PG: postgres://user:pass@host:5432/db

# 订阅
SUB_BASE_URL=http://localhost:8080  # 订阅链接基础 URL
SUB_REFRESH_INTERVAL=30      # 订阅源自动刷新间隔（分钟）
ALLOW_REGISTER=true          # 是否允许用户自行注册

# Agent
AGENT_REPORT_INTERVAL=60     # 默认建议 Agent 上报间隔（秒）
AGENT_OFFLINE_TIMEOUT=180    # Agent 离线判定超时（秒）
```

---

## Docker Compose 架构

所有容器名、网络名、卷名均带 `submgr-` 前缀，与其他服务隔离。

```
┌─────────────────────────────────────────────────┐
│              submgr-nginx (:80/:443)             │
│           反向代理 + 前端静态文件托管               │
├────────────────────┬────────────────────────────┤
│  / → Frontend      │  /api/* → Backend :8080    │
│  (React 静态文件)   │  /sub/* → Backend :8080    │
└────────────────────┴────────────────────────────┘
                           │
                    ┌──────┴──────┐
                    │ submgr-app  │
                    │  (Go+Gin)   │
                    └──────┬──────┘
                           │
                  ┌────────┴────────┐
                  │ submgr-postgres │
                  │   (:5432)       │
                  └─────────────────┘
```

### docker-compose.yml 设计

```yaml
services:
  submgr-postgres:
    image: postgres:16-alpine
    container_name: submgr-postgres
    restart: unless-stopped
    environment:
      POSTGRES_DB: ${DB_NAME:-submanager}
      POSTGRES_USER: ${DB_USER:-submgr}
      POSTGRES_PASSWORD: ${DB_PASSWORD:?DB_PASSWORD is required}
    volumes:
      - submgr-pgdata:/var/lib/postgresql/data
    networks:
      - submgr-net
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER:-submgr}"]
      interval: 5s
      timeout: 3s
      retries: 5

  submgr-app:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: submgr-app
    restart: unless-stopped
    depends_on:
      submgr-postgres:
        condition: service_healthy
    environment:
      APP_ENV: production
      APP_PORT: 8080
      APP_SECRET: ${APP_SECRET:?APP_SECRET is required}
      INIT_TOKEN: ${INIT_TOKEN:?INIT_TOKEN is required}
      DB_DRIVER: postgres
      DB_DSN: postgres://${DB_USER:-submgr}:${DB_PASSWORD}@submgr-postgres:5432/${DB_NAME:-submanager}?sslmode=disable
      SUB_BASE_URL: ${SUB_BASE_URL:-http://localhost}
      SUB_REFRESH_INTERVAL: ${SUB_REFRESH_INTERVAL:-30}
      ALLOW_REGISTER: ${ALLOW_REGISTER:-true}
      AGENT_REPORT_INTERVAL: ${AGENT_REPORT_INTERVAL:-60}
      AGENT_OFFLINE_TIMEOUT: ${AGENT_OFFLINE_TIMEOUT:-180}
    networks:
      - submgr-net

  submgr-nginx:
    image: nginx:alpine
    container_name: submgr-nginx
    restart: unless-stopped
    ports:
      - "${HTTP_PORT:-80}:80"
      - "${HTTPS_PORT:-443}:443"
    volumes:
      - ./nginx/default.conf:/etc/nginx/conf.d/default.conf:ro
      - ./frontend/dist:/usr/share/nginx/html:ro
      - ${SSL_CERT_PATH:-/dev/null}:/etc/nginx/ssl/cert.pem:ro
      - ${SSL_KEY_PATH:-/dev/null}:/etc/nginx/ssl/key.pem:ro
    depends_on:
      - submgr-app
    networks:
      - submgr-net

volumes:
  submgr-pgdata:
    name: submgr-pgdata

networks:
  submgr-net:
    name: submgr-net
    driver: bridge
```

### Backend Dockerfile（多阶段构建）

```dockerfile
# 构建阶段
FROM golang:1.22-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o submanager .

# 运行阶段
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /build/submanager .
EXPOSE 8080
CMD ["./submanager"]
```

### Frontend Dockerfile

```dockerfile
# 构建阶段
FROM node:20-alpine AS builder
WORKDIR /build
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# 运行阶段 — 将 dist 复制到固定路径，由 deploy.sh 处理卷挂载
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /dist
COPY --from=builder /build/dist .
```

> **前端产物注入流程**：前端容器仅负责构建。`deploy.sh` 在构建后执行 `docker cp` 将前端容器中的 `/dist` 内容复制到宿主机目录（如 `./frontend/dist/`），然后 nginx 容器通过 bind mount 挂载该目录。不再使用 Docker volume，简化流程。

### Nginx 反向代理配置

```nginx
server {
    listen 80;
    server_name _;

    root /usr/share/nginx/html;
    index index.html;

    # SPA 路由回退
    location / {
        try_files $uri $uri/ /index.html;
    }

    # API 反向代理
    location /api/ {
        proxy_pass http://submgr-app:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # 订阅链接反向代理
    location /sub/ {
        proxy_pass http://submgr-app:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

---

## 部署脚本 (deploy.sh)

交互式引导配置，一键拉起所有容器。

### 脚本功能

1. **检查环境** — 检测 Docker 和 Docker Compose 是否安装
2. **交互配置** — 引导用户填写必要参数
3. **生成 .env** — 将配置写入 `.env` 文件
4. **构建镜像** — 构建前后端 Docker 镜像
5. **拉起服务** — `docker compose up -d`
6. **初始化** — 等待服务就绪，创建默认管理员账户
7. **输出信息** — 显示访问地址、管理员账号等

### 交互参数

| 参数 | 提示语 | 默认值 | 必填 |
|------|--------|--------|------|
| HTTP_PORT | 外部 HTTP 端口 | 80 | 是 |
| HTTPS_PORT | 外部 HTTPS 端口 | 443 | 否 |
| APP_SECRET | JWT 密钥 | 随机生成 | 是 |
| INIT_TOKEN | 管理员初始化令牌 | 随机生成 | 是（自动） |
| DB_PASSWORD | 数据库密码 | 随机生成 | 是 |
| SUB_BASE_URL | 订阅基础 URL | http://当前IP:端口 | 是 |
| ADMIN_USERNAME | 管理员用户名 | admin | 是 |
| ADMIN_PASSWORD | 管理员密码 | 随机生成 | 是 |
| ALLOW_REGISTER | 允许用户注册 | true | 否 |
| SUB_REFRESH_INTERVAL | 订阅源刷新间隔(分钟) | 30 | 否 |
| AGENT_REPORT_INTERVAL | Agent 上报间隔(秒) | 60 | 否 |
| AGENT_OFFLINE_TIMEOUT | Agent 离线判定超时(秒) | 180 | 否 |

### 脚本流程伪代码

```bash
#!/bin/bash
set -e

PREFIX="submgr"
COMPOSE_FILE="docker-compose.yml"
ENV_FILE=".env"

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}  SubManager 代理订阅管理平台 部署脚本   ${NC}"
echo -e "${CYAN}========================================${NC}"

# 1. 环境检查
check_docker() { ... }

# 2. 交互配置
configure() {
    read_input "HTTP_PORT" "外部 HTTP 端口" "80"
    read_input "APP_SECRET" "JWT 密钥（留空自动生成）" ""
    [ -z "$APP_SECRET" ] && APP_SECRET=$(openssl rand -hex 32)
    read_input "DB_PASSWORD" "数据库密码（留空自动生成）" ""
    [ -z "$DB_PASSWORD" ] && DB_PASSWORD=$(openssl rand -hex 16)
    read_input "SUB_BASE_URL" "订阅基础 URL" "http://$(hostname -I | awk '{print $1}'):${HTTP_PORT}"
    read_input "ADMIN_USERNAME" "管理员用户名" "admin"
    read_input "ADMIN_PASSWORD" "管理员密码（留空自动生成）" ""
    [ -z "$ADMIN_PASSWORD" ] && ADMIN_PASSWORD=$(openssl rand -hex 8)
    # 生成管理员初始化令牌
    INIT_TOKEN=$(openssl rand -hex 16)
    read_input "ALLOW_REGISTER" "允许用户注册 (true/false)" "true"
    confirm_config  # 展示配置摘要，确认后继续
}

# 3. 写入 .env
write_env() {
    cat > .env <<EOF
HTTP_PORT=${HTTP_PORT}
APP_SECRET=${APP_SECRET}
INIT_TOKEN=${INIT_TOKEN}
DB_PASSWORD=${DB_PASSWORD}
SUB_BASE_URL=${SUB_BASE_URL}
ADMIN_USERNAME=${ADMIN_USERNAME}
ADMIN_PASSWORD=${ADMIN_PASSWORD}
ALLOW_REGISTER=${ALLOW_REGISTER}
EOF
}

# 4. 构建前端
build_frontend() {
    echo -e "${YELLOW}构建前端...${NC}"
    # 构建前端 Docker 镜像并提取产物
    docker build -t submgr-frontend-builder ./frontend
    # 创建临时容器，将 dist 复制到宿主机
    docker create --name submgr-dist-temp submgr-frontend-builder
    mkdir -p ./frontend/dist
    docker cp submgr-dist-temp:/dist/. ./frontend/dist/
    docker rm submgr-dist-temp
}

# 5. 构建并启动
deploy() {
    echo -e "${YELLOW}构建后端镜像...${NC}"
    docker compose build

    echo -e "${YELLOW}启动服务...${NC}"
    docker compose up -d

    echo -e "${YELLOW}等待服务就绪...${NC}"
    wait_for_healthy
}

# 6. 初始化管理员
init_admin() {
    # 等待后端就绪
    retry_count=0
    while [ $retry_count -lt 30 ]; do
        if curl -sf http://localhost:${HTTP_PORT}/api/health > /dev/null 2>&1; then
            break
        fi
        retry_count=$((retry_count + 1))
        sleep 1
    done

    curl -sf -X POST http://localhost:${HTTP_PORT}/api/system/init \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"${ADMIN_USERNAME}\",\"password\":\"${ADMIN_PASSWORD}\",\"init_token\":\"${INIT_TOKEN}\"}"
}

# 7. 输出信息
show_info() {
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  部署完成!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo -e "  访问地址: ${CYAN}${SUB_BASE_URL}${NC}"
    echo -e "  管理员:   ${CYAN}${ADMIN_USERNAME} / ${ADMIN_PASSWORD}${NC}"
    echo -e ""
    echo -e "  管理命令:"
    echo -e "    查看日志: docker compose logs -f"
    echo -e "    停止服务: docker compose down"
    echo -e "    重启服务: docker compose restart"
    echo -e "${GREEN}========================================${NC}"
}

check_docker
configure
write_env
build_frontend
deploy
init_admin
show_info
```

---

## 部署脚本辅助命令

```bash
# 部署
./deploy.sh

# 查看状态
docker compose ps

# 查看日志
docker compose logs -f submgr-app        # 后端日志
docker compose logs -f submgr-nginx      # Nginx 日志
docker compose logs -f submgr-postgres   # 数据库日志

# 停止
docker compose down                      # 停止但保留数据
docker compose down -v                   # 停止并删除数据卷（危险）

# 重建（代码更新后）
docker compose build --no-cache
docker compose up -d

# 备份数据库
docker compose exec submgr-postgres pg_dump -U submgr submanager > backup.sql

# 恢复数据库
docker compose exec -T submgr-postgres psql -U submgr submanager < backup.sql
```

---

## 验证清单

- [ ] `cd backend && go test ./... -v` 全部通过
- [ ] `cd frontend && npm test` 全部通过
- [ ] `cd frontend && npm run build` 构建成功
- [ ] `docker compose up` 一键启动
- [ ] 用户注册 → 登录 → 获取订阅链接
- [ ] 管理员创建套餐 → 添加订阅源 → 创建服务群 → 分配给用户
- [ ] 订阅链接在 Clash/singbox 客户端可正常导入
- [ ] Agent 上报后节点出现在订阅输出
- [ ] 过期/流量耗尽用户订阅返回友好提示
