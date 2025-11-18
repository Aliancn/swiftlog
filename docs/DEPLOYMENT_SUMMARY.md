# SwiftLog 部署总结

## ✅ 完成的重构

已成功将SwiftLog重构为一键启动架构：

### 1. 统一的Docker Compose配置

**文件**: `docker compose.yaml`（项目根目录）

**包含的服务**:
- ✅ PostgreSQL 16（自动迁移数据库）
- ✅ Grafana Loki 2.9（日志存储）
- ✅ Redis 7（消息队列）
- ✅ Ingestor（gRPC服务，端口50051）
- ✅ API（REST服务，端口8080）
- ✅ WebSocket（实时服务，端口8081）
- ✅ AI Worker（后台分析服务）
- ✅ Frontend（Next.js，端口3000）

**特性**:
- 健康检查配置完善
- 服务依赖关系明确
- 数据持久化（volumes）
- 网络隔离（swiftlog网络）

### 2. 自动数据库迁移

**实现方式**: PostgreSQL的`docker-entrypoint-initdb.d`机制

**迁移文件**（已修改为幂等）:
```
backend/migrations/
├── 000_init_extensions.sql      # PostgreSQL扩展
├── 001_create_users_table.sql   # 用户表
├── 002_create_api_tokens_table.sql  # API令牌表
├── 003_create_projects_table.sql    # 项目表
├── 004_create_log_groups_table.sql  # 日志分组表
└── 005_create_log_runs_table.sql    # 日志运行表
```

**幂等性保证**:
- 所有表: `CREATE TABLE IF NOT EXISTS`
- 所有索引: `CREATE INDEX IF NOT EXISTS`
- 触发器: `DROP TRIGGER IF EXISTS` + `CREATE TRIGGER`
- 函数: `CREATE OR REPLACE FUNCTION`

### 3. 简化的启动方式

#### 方式一: Makefile（推荐）

```bash
make start    # 启动所有服务
make stop     # 停止所有服务
make restart  # 重启服务
make logs     # 查看日志
make cli      # 编译CLI工具
make clean    # 清理所有数据
make dev-up   # 仅启动基础设施（开发模式）
```

#### 方式二: 启动脚本

```bash
./start.sh    # 带颜色输出的交互式启动
```

#### 方式三: Docker Compose

```bash
docker compose up -d      # 启动
docker compose down       # 停止
docker compose logs -f    # 查看日志
```

### 4. 环境配置

**文件**: `.env.example` → `.env`

**必需配置**:
```bash
POSTGRES_PASSWORD=your-secure-password
OPENAI_API_KEY=sk-your-openai-key
JWT_SECRET=your-random-secret
```

**可选配置**:
```bash
ENVIRONMENT=production
LOG_LEVEL=info

# OpenAI配置
OPENAI_MODEL=gpt-4o-mini
OPENAI_BASE_URL=https://api.openai.com/v1  # 支持Azure OpenAI、LocalAI等兼容端点
```

### 5. 网络安全配置

**仅暴露用户交互端口**:
- ✅ Frontend: 3000
- ✅ API: 8080
- ✅ WebSocket: 8081
- ✅ gRPC Ingestor: 50051

**内部服务（Docker网络内）**:
- ❌ PostgreSQL: 5432（不暴露）
- ❌ Loki: 3100（不暴露）
- ❌ Redis: 6379（不暴露）

**开发调试**: 如需访问内部服务，取消`docker compose.yaml`中相应端口映射的注释。

### 6. CLI工具独立编译

**编译方式**:
```bash
make cli
# 或
cd cli && go build -o swiftlog
```

**安装**:
```bash
sudo cp cli/swiftlog /usr/local/bin/
```

**配置**:
```bash
swiftlog config set --token YOUR_TOKEN --server localhost:50051
```

## 🎯 一键启动流程

### 完整启动步骤（3分钟）

```bash
# 1. 克隆项目
git clone <repository-url>
cd swiftlog

# 2. 配置环境
cp .env.example .env
nano .env  # 设置POSTGRES_PASSWORD、OPENAI_API_KEY、JWT_SECRET

# 3. 一键启动（自动构建镜像、创建网络、启动服务、迁移数据库）
make start

# 4. 编译CLI
make cli

# 5. 测试
./cli/swiftlog run --project test -- echo "Hello SwiftLog"
```

## 📊 服务端口映射

| 服务 | 容器端口 | 主机端口 | 暴露 | 用途 |
|------|---------|---------|------|------|
| Frontend | 3000 | 3000 | ✅ | Web界面 |
| API | 8080 | 8080 | ✅ | REST API |
| WebSocket | 8081 | 8081 | ✅ | 实时流 |
| Ingestor | 50051 | 50051 | ✅ | gRPC（CLI连接） |
| PostgreSQL | 5432 | - | ❌ | 数据库（内部） |
| Loki | 3100 | - | ❌ | 日志存储（内部） |
| Redis | 6379 | - | ❌ | 消息队列（内部） |

**安全提示**: 基础设施服务（PostgreSQL、Loki、Redis）默认不暴露到主机，仅在Docker网络内可访问，提高安全性。

## 🔧 开发模式

### 本地开发（热重载）

```bash
# 启动基础设施
make dev-up

# 手动运行各个服务（分别在不同终端）
cd backend/cmd/ingestor && go run main.go
cd backend/cmd/api && go run main.go
cd backend/cmd/websocket && go run main.go
cd backend/cmd/ai-worker && go run main.go
cd frontend && npm run dev

# 关闭基础设施
make dev-down
```

## 📁 项目结构变化

### 新增文件

```
swiftlog/
├── docker compose.yaml        # 统一的Docker编排（新）
├── Makefile                   # 构建自动化（新）
├── start.sh                   # 启动脚本（新）
├── .env.example               # 环境变量模板（新）
├── QUICKSTART.md              # 快速开始指南（新）
├── DEPLOYMENT_SUMMARY.md      # 本文档（新）
└── README.md                  # 更新的文档
```

### 删除文件

```
docker/                        # 已删除（整合到根目录docker compose.yaml）
```

### 修改文件

```
backend/migrations/*.sql       # 所有迁移文件改为幂等
README.md                      # 更新为一键启动说明
```

## 🚀 部署检查清单

### 生产环境部署前

- [ ] 修改 `.env` 中的 `POSTGRES_PASSWORD`（使用强密码）
- [ ] 设置安全的 `JWT_SECRET`（随机生成）
- [ ] 配置有效的 `OPENAI_API_KEY`
- [ ] 更新API和WebSocket的CORS设置（在代码中）
- [ ] 配置HTTPS/TLS证书
- [ ] 设置防火墙规则（仅暴露必要端口）
- [ ] 配置备份策略（PostgreSQL数据）
- [ ] 启用监控和日志收集
- [ ] 配置反向代理（nginx/traefik）
- [ ] 设置速率限制

### 启动验证

```bash
# 1. 检查所有服务状态
docker compose ps

# 2. 验证健康检查
docker compose ps | grep healthy

# 3. 测试API
curl http://localhost:8080/health

# 4. 查看日志确认无错误
docker compose logs | grep -i error

# 5. 测试CLI连接
./cli/swiftlog config get
```

## 🐛 常见问题

### 服务启动失败

```bash
# 查看具体错误
docker compose logs <service-name>

# 重建镜像
docker compose build --no-cache
docker compose up -d
```

### 数据库迁移失败

```bash
# 查看PostgreSQL日志
docker compose logs postgres

# 手动运行迁移
docker compose exec postgres psql -U swiftlog -d swiftlog -f /docker-entrypoint-initdb.d/001_create_users_table.sql
```

### 端口冲突

```bash
# 修改docker compose.yaml中的端口映射
# 例如：将 "8080:8080" 改为 "8081:8080"
```

## 📈 性能优化建议

### 生产环境

1. **扩展AI Worker**:
   ```bash
   docker compose up -d --scale ai-worker=3
   ```

2. **PostgreSQL连接池**:
   - 已配置：MaxOpenConns=25, MaxIdleConns=5

3. **Loki数据保留**:
   - 配置Loki的retention策略（默认无限期）

4. **Redis持久化**:
   - 已配置RDB持久化

## 🎉 完成状态

- ✅ 统一Docker Compose配置
- ✅ 自动数据库迁移
- ✅ 幂等SQL脚本
- ✅ Makefile自动化
- ✅ 启动脚本
- ✅ 环境配置模板
- ✅ 更新文档
- ✅ 快速开始指南
- ✅ CLI独立编译
- ✅ 网络安全配置（仅暴露必要端口）
- ✅ 支持自定义OpenAI兼容端点（Azure OpenAI、LocalAI等）

**SwiftLog现在可以一键启动，开箱即用！** 🚀

### 🆕 最新更新

**网络安全优化**:
- 基础设施服务（PostgreSQL、Loki、Redis）不再暴露到主机
- 仅暴露用户交互端口（3000、8080、8081、50051）
- 所有服务间通信通过Docker内部网络

**OpenAI灵活配置**:
- 支持自定义API端点（`OPENAI_BASE_URL`）
- 兼容Azure OpenAI Service
- 支持本地部署的LocalAI、Ollama等
- 可配置任意OpenAI兼容模型

---

**最后更新**: 2025-11-18
