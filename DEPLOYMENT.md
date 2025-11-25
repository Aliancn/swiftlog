# SwiftLog Deployment Guide

本文档说明如何部署SwiftLog到生产环境。

## 前置要求

### 服务器要求
- Ubuntu 20.04+ 或其他支持Docker的Linux发行版
- Docker 24.0+
- Docker Compose v2+
- 至少 2GB RAM
- 至少 10GB 磁盘空间

### GitHub Secrets 配置

所有环境配置通过GitHub Secrets管理。可以使用提供的脚本自动配置，或手动在GitHub仓库的Settings -> Secrets and variables -> Actions中配置。

#### 方式1：使用自动配置脚本（推荐）

```bash
./scripts/setup-gh-secrets.sh
```

该脚本会交互式地配置所有必需的secrets，并自动生成安全的随机密钥。

#### 方式2：手动配置Secrets

在GitHub仓库的Settings -> Secrets and variables -> Actions中配置以下secrets：

##### 必需的Secrets

| Secret名称 | 说明 | 示例/生成方式 |
|-----------|------|--------------|
| `DEPLOY_HOST` | 部署服务器的IP或域名 | `123.45.67.89` |
| `DEPLOY_USER` | SSH用户名 | `ubuntu` |
| `DEPLOY_SSH_KEY` | SSH私钥（完整内容） | `-----BEGIN RSA PRIVATE KEY-----\n...` |
| `POSTGRES_PASSWORD` | PostgreSQL数据库密码 | `openssl rand -base64 32` |
| `JWT_SECRET` | JWT签名密钥 | `openssl rand -base64 32` |
| `ENCRYPTION_KEY` | API密钥加密密钥 | `openssl rand -base64 32` |
| `ADMIN_PASSWORD` | 初始管理员密码 | `openssl rand -base64 16` |

##### 可选但强烈推荐的Secrets

| Secret名称 | 说明 | 默认值 | 重要性 |
|-----------|------|--------|--------|
| `PUBLIC_URL` | 公网访问URL | - | **强烈推荐**,用于设置CORS |
| `CORS_ORIGINS` | CORS允许的源（多个用逗号分隔） | 使用`PUBLIC_URL`值 | **强烈推荐** |

> ⚠️ **重要**: 在生产环境中,强烈建议设置 `PUBLIC_URL` 或 `CORS_ORIGINS`,否则会使用默认值 `http://localhost,http://127.0.0.1`,这可能导致跨域访问问题。

##### 其他可选Secrets

| Secret名称 | 说明 | 默认值 |
|-----------|------|--------|
| `DEPLOY_PATH` | 服务器上的部署路径 | `/opt/swiftlog` |
| `ENVIRONMENT` | 运行环境 | `production` |
| `LOG_LEVEL` | 日志级别 | `info` |
| **数据库配置** | | |
| `POSTGRES_USER` | PostgreSQL用户名 | `swiftlog` |
| `POSTGRES_DB` | PostgreSQL数据库名 | `swiftlog` |
| **端口配置** | | |
| `NGINX_PORT` | Nginx映射到主机的端口 | `80` |
| `GRPC_PORT` | gRPC服务端口 | `50051` |
| `API_PORT` | API服务内部端口 | `8080` |
| `WS_PORT` | WebSocket服务内部端口 | `8081` |
| **安全配置** | | |
| `ADMIN_USERNAME` | 初始管理员用户名 | `admin` |
| `JWT_EXPIRATION` | JWT过期时间 | `2h` |

## 部署流程

### 1. Release流程（构建和发布）

当你push一个版本tag时，会自动触发release流程：

```bash
# 创建一个新版本tag
git tag v1.0.0
git push origin v1.0.0
```

Release流程会自动：
1. 构建CLI工具的多平台二进制文件（Linux, macOS, Windows）
2. 创建GitHub Release并上传CLI二进制文件
3. 构建所有Docker镜像（backend服务 + frontend）
4. 推送Docker镜像到 GitHub Container Registry (ghcr.io)
5. 为镜像打上版本标签（`v1.0.0`, `1.0`, `latest`）

### 2. Deploy流程（部署到服务器）

**部署采用手动触发方式**，需要在GitHub Actions页面手动启动：

#### 手动部署步骤

1. 进入GitHub仓库页面
2. 点击 "Actions" 选项卡
3. 在左侧选择 "Deploy" workflow
4. 点击右侧的 "Run workflow" 按钮
5. 在弹出的对话框中配置：
   - **version**: 要部署的Docker镜像版本（例如：`0.1.0`, `latest`）
   - **environment**: 部署环境（`production` 或 `staging`）
6. 点击 "Run workflow" 开始部署

**版本选择说明**：
- 输入版本号时**不要**包含 `v` 前缀（例如：输入 `0.1.0` 而不是 `v0.1.0`）
- 使用 `latest` 可以部署最新构建的镜像
- 版本号必须对应已发布的Docker镜像标签

### 3. 部署步骤详解

Deploy流程执行以下步骤：

1. **准备部署文件**
   - 复制 `docker-compose.yaml`
   - 复制 `nginx/nginx.conf`
   - 复制 `loki-config.yaml`
   - 创建服务更新脚本

2. **修改Docker Compose配置**
   - 注释掉本地构建配置
   - 添加远程镜像引用
   - 使用指定版本的镜像

3. **上传到服务器**
   - 通过SCP上传所有文件到 `/tmp/deploy`

4. **在服务器上执行部署**
   - 创建部署目录（默认`/opt/swiftlog`）
   - 复制新文件到部署目录
   - **从GitHub Secrets自动创建.env文件**（包含所有环境变量）
   - 验证必需的secrets
   - 登录GitHub Container Registry
   - 拉取最新镜像
   - 停止旧服务
   - 启动新服务
   - 等待服务健康

5. **健康检查**
   - 检查所有容器状态
   - 测试Nginx健康端点
   - 测试API健康端点
   - 如果失败，输出日志并退出

**重要说明**：
- 所有环境变量都从GitHub Secrets自动注入，无需手动修改.env文件
- 每次部署都会重新创建.env文件，确保配置与GitHub Secrets同步
- 要更改配置，只需在GitHub Secrets中修改，然后重新部署

## 服务器初始设置

首次部署前，需要在服务器上进行一些初始设置：

### 1. 安装Docker和Docker Compose

```bash
# 安装Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# 将当前用户添加到docker组
sudo usermod -aG docker $USER

# 安装Docker Compose v2
sudo apt-get update
sudo apt-get install docker-compose-plugin

# 验证安装
docker --version
docker compose version
```

### 2. 配置SSH访问

```bash
# 在本地生成SSH密钥（如果还没有）
ssh-keygen -t rsa -b 4096 -C "your_email@example.com"

# 复制公钥到服务器
ssh-copy-id user@your-server-ip

# 将私钥内容添加到GitHub Secrets (DEPLOY_SSH_KEY)
cat ~/.ssh/id_rsa
```

### 3. 创建部署目录

```bash
sudo mkdir -p /opt/swiftlog
sudo chown $USER:$USER /opt/swiftlog
```

### 4. 配置防火墙（可选）

```bash
# 允许HTTP/HTTPS访问
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# 如果需要直接访问gRPC
sudo ufw allow 50051/tcp
```

## 生产环境配置

### 配置管理

所有配置都通过GitHub Secrets管理。要修改配置：

1. 在GitHub仓库的 Settings -> Secrets and variables -> Actions 中修改相应的secret
2. 重新触发部署（推送新tag或手动触发deploy workflow）
3. 部署流程会自动使用新配置创建.env文件

### 推荐的生产环境配置

使用 `scripts/setup-gh-secrets.sh` 脚本时，建议的配置：

```bash
# 部署配置
Server IP/domain: your-server-ip
SSH user: ubuntu
Deploy path: /opt/swiftlog
SSH key path: ~/.ssh/id_ed25519

# 数据库配置
PostgreSQL user: swiftlog
PostgreSQL database name: swiftlog
PostgreSQL password: (auto-generated 32-byte key)

# 端口配置
NGINX port: 80  # 或443（如果使用HTTPS）
gRPC port: 50051  # 保持默认
API port: 8080  # 保持默认
WebSocket port: 8081  # 保持默认

# 应用配置
Public URL: https://logs.yourdomain.com
CORS origins: https://logs.yourdomain.com  # 或留空自动使用PUBLIC_URL
Environment: production
Log level: info  # 或warn（减少日志量）

# 安全配置
Admin username: admin  # 或自定义
Admin password: (auto-generated)
JWT expiration: 2h  # 或更短（1h）提高安全性
JWT secret: (auto-generated)
Encryption key: (auto-generated)
```

### 安全最佳实践

1. **密钥管理**
   - 所有密钥都由脚本自动生成（32+字符）
   - 保存脚本输出的credentials到安全的密码管理器
   - 定期轮换JWT_SECRET和ENCRYPTION_KEY（建议每季度）

2. **密码强度**
   - POSTGRES_PASSWORD: 32字节base64编码
   - JWT_SECRET: 32字节base64编码
   - ENCRYPTION_KEY: 32字节base64编码
   - ADMIN_PASSWORD: 16字节base64编码（足够安全）

3. **网络安全**
   - 生产环境必须使用HTTPS（配置外部nginx反向代理）
   - 限制SSH访问（仅允许特定IP）
   - 配置防火墙规则（仅开放必要端口）

4. **日志和监控**
   - 使用 `LOG_LEVEL=warn` 或 `info` 平衡安全和调试需求
   - 定期查看容器日志检查异常
   - 设置日志告警（可选）

### 外部Nginx配置（推荐）

如果使用外部Nginx作为反向代理：

```nginx
upstream swiftlog {
    server 127.0.0.1:8080;
}

server {
    listen 80;
    server_name swiftlog.yourdomain.com;

    # Redirect to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name swiftlog.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/swiftlog.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/swiftlog.yourdomain.com/privkey.pem;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;

    client_max_body_size 100M;

    location / {
        proxy_pass http://swiftlog;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # WebSocket support
    location /ws {
        proxy_pass http://swiftlog;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 监控和维护

### 查看服务状态

```bash
cd /opt/swiftlog
docker compose ps
docker compose logs -f
```

### 查看特定服务日志

```bash
docker compose logs -f api
docker compose logs -f frontend
docker compose logs -f websocket
```

### 重启服务

```bash
cd /opt/swiftlog
docker compose restart
```

### 备份数据

```bash
# 备份PostgreSQL数据
docker compose exec -T postgres pg_dump -U swiftlog swiftlog > backup_$(date +%Y%m%d).sql

# 备份环境变量
cp /opt/swiftlog/.env /opt/swiftlog/.env.backup
```

### 回滚到previous版本

```bash
cd /opt/swiftlog

# 编辑docker-compose.yaml，将版本改为之前的版本
# 例如：从 v1.0.1 改回 v1.0.0
sed -i 's/:v1.0.1/:v1.0.0/g' docker-compose.yaml

# 重新部署
docker compose pull
docker compose down
docker compose up -d
```

## 故障排查

### 服务无法启动

1. 检查容器日志
```bash
docker compose logs
```

2. 检查环境变量配置
```bash
cat /opt/swiftlog/.env
```

3. 检查端口占用
```bash
sudo netstat -tlnp | grep -E '(80|8080|50051)'
```

### 数据库连接失败

```bash
# 检查PostgreSQL容器
docker compose logs postgres

# 手动连接测试
docker compose exec postgres psql -U swiftlog -d swiftlog
```

### 镜像拉取失败

```bash
# 检查是否登录到GHCR
docker login ghcr.io

# 手动拉取测试
docker pull ghcr.io/aliancn/swiftlog/api:latest
```

## 性能优化

1. **数据库优化**
   - 定期清理旧日志
   - 为常用查询创建索引
   - 配置PostgreSQL连接池

2. **缓存优化**
   - 使用Redis缓存热点数据
   - 配置合适的缓存过期时间

3. **资源限制**
   - 在docker-compose.yaml中设置资源限制
   - 监控资源使用情况

## 支持

如遇到问题，请：
1. 查看 [GitHub Issues](https://github.com/Aliancn/swiftlog/issues)
2. 提交新issue并附上日志信息
3. 联系维护团队
