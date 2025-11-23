# SwiftLog v0.1.0 Release Guide

## 快速发布指南 (Quick Release Guide)

本文档提供SwiftLog v0.1.0版本的发布流程和CI/CD配置指南。

## 📋 发布前检查清单

在创建发布之前，请确保：

- [ ] 所有功能已完成并测试
- [ ] 所有测试通过 (`make test`)
- [ ] 文档已更新（README, CLI docs, API docs）
- [ ] CHANGELOG.md已包含本版本的所有变更
- [ ] VERSION文件已更新为正确版本号

## 🚀 发布流程

### 方式一：使用Makefile（推荐）

```bash
# 1. 确认版本号
make version

# 2. 创建发布（会构建所有平台的CLI并创建git tag）
make release

# 3. 推送tag到GitHub（触发自动化流程）
git push origin v0.1.0
```

### 方式二：手动发布

```bash
# 1. 更新版本
echo "0.1.0" > VERSION

# 2. 更新CHANGELOG
nano CHANGELOG.md

# 3. 构建CLI（可选，GitHub Actions会自动构建）
make cli-all

# 4. 创建并推送tag
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0
```

## 🤖 GitHub Actions自动化流程

推送tag后，GitHub Actions会自动执行：

### 1. Release Workflow (`.github/workflows/release.yml`)

**自动完成：**
- ✅ 构建CLI二进制文件（Linux, macOS, Windows - amd64/arm64）
- ✅ 生成SHA256校验和
- ✅ 创建GitHub Release
- ✅ 上传所有二进制文件到Release
- ✅ 从CHANGELOG.md提取Release Notes

**产出物：**
- `swiftlog-linux-amd64`
- `swiftlog-linux-arm64`
- `swiftlog-darwin-amd64`
- `swiftlog-darwin-arm64`
- `swiftlog-windows-amd64.exe`
- 各文件的`.sha256`校验和文件

### 2. Docker Build Workflow

**自动完成：**
- ✅ 构建所有服务的Docker镜像（多平台：amd64, arm64）
- ✅ 推送到GitHub Container Registry (ghcr.io)
- ✅ 标记版本号和latest标签

**Docker镜像：**
```
ghcr.io/aliancn/swiftlog/api:v0.1.0
ghcr.io/aliancn/swiftlog/ingestor:v0.1.0
ghcr.io/aliancn/swiftlog/websocket:v0.1.0
ghcr.io/aliancn/swiftlog/ai-worker:v0.1.0
ghcr.io/aliancn/swiftlog/frontend:v0.1.0
```

### 3. Deploy Workflow (`.github/workflows/deploy.yml`)

**自动完成（如果配置了部署密钥）：**
- ✅ 准备部署文件
- ✅ 通过SCP复制到服务器
- ✅ 执行部署脚本
- ✅ 运行健康检查
- ✅ 验证部署

## ⚙️ GitHub Actions配置

### 必需的Secrets

在GitHub仓库设置中配置（Settings → Secrets and variables → Actions）：

**部署相关（如需自动部署）：**
```
DEPLOY_HOST          - 服务器地址（IP或域名）
DEPLOY_USER          - SSH用户名
DEPLOY_SSH_KEY       - SSH私钥（完整内容，包含头尾）
DEPLOY_PATH          - 部署路径（可选，默认：/opt/swiftlog）
```

**前端配置（可选）：**
```
NEXT_PUBLIC_API_URL  - API地址
NEXT_PUBLIC_WS_URL   - WebSocket地址
```

### 配置步骤

详细配置步骤请参见：[`.github/SETUP.md`](.github/SETUP.md)

**快速配置：**

1. 生成SSH密钥对：
   ```bash
   ssh-keygen -t ed25519 -f ~/.ssh/swiftlog_deploy
   ```

2. 将公钥添加到服务器：
   ```bash
   ssh-copy-id -i ~/.ssh/swiftlog_deploy.pub user@server
   ```

3. 将私钥添加到GitHub Secrets：
   - 复制`~/.ssh/swiftlog_deploy`的完整内容
   - 在GitHub添加为`DEPLOY_SSH_KEY` secret

## 📦 部署选项

### 选项1：自动部署（GitHub Actions）

配置好Secrets后，推送tag即可自动部署：

```bash
git tag v0.1.0
git push origin v0.1.0
# GitHub Actions会自动部署到配置的服务器
```

### 选项2：手动部署脚本

使用提供的部署脚本：

```bash
./deploy.sh -h your-server.com -u deploy -v v0.1.0
```

参数说明：
- `-h, --host`: 服务器地址
- `-u, --user`: SSH用户
- `-k, --key`: SSH密钥路径（可选）
- `-v, --version`: 版本号
- `-p, --path`: 部署路径（可选）

### 选项3：Docker Compose直接部署

在服务器上：

```bash
# 1. 登录GitHub Container Registry
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# 2. 拉取最新镜像
docker compose pull

# 3. 启动服务
docker compose up -d
```

## 🔍 验证发布

### 1. 检查GitHub Release

访问：`https://github.com/aliancn/swiftlog/releases/tag/v0.1.0`

验证：
- [ ] Release已创建
- [ ] 所有CLI二进制文件已上传
- [ ] Release notes正确显示

### 2. 检查Docker镜像

```bash
# 查看镜像
docker pull ghcr.io/aliancn/swiftlog/api:v0.1.0

# 验证标签
docker images | grep swiftlog
```

### 3. 测试CLI下载

```bash
# Linux
wget https://github.com/aliancn/swiftlog/releases/download/v0.1.0/swiftlog-linux-amd64
chmod +x swiftlog-linux-amd64
./swiftlog-linux-amd64 --version

# macOS
curl -L -o swiftlog https://github.com/aliancn/swiftlog/releases/download/v0.1.0/swiftlog-darwin-arm64
chmod +x swiftlog
./swiftlog --version
```

### 4. 验证部署（如果已自动部署）

```bash
# SSH到服务器
ssh deploy@your-server

# 检查服务状态
cd /opt/swiftlog
docker compose ps

# 测试API
curl http://localhost:8080/health
```

## 📝 发布后任务

- [ ] 在Release页面添加详细的Release Notes
- [ ] 更新README的版本引用
- [ ] 发布公告（如果有）
- [ ] 通知用户升级
- [ ] 监控部署状态和错误日志

## 🔄 版本号规范

SwiftLog遵循[语义化版本](https://semver.org/lang/zh-CN/)：

```
主版本号.次版本号.修订号 (MAJOR.MINOR.PATCH)

v0.1.0 → v0.1.1  - 修复bug
v0.1.1 → v0.2.0  - 新功能（向后兼容）
v0.9.0 → v1.0.0  - 重大变更（不兼容）
```

**v0.1.0说明：**
- 主版本号 0：初始开发阶段，API可能不稳定
- 次版本号 1：第一个功能版本
- 修订号 0：首次发布

## 🐛 问题排查

### 发布失败

**问题：GitHub Actions失败**

检查：
1. Actions标签页查看详细日志
2. 验证tag格式（必须是`v*.*.*`）
3. 确保`VERSION`文件和tag匹配

**问题：Docker镜像推送失败**

检查：
1. GITHUB_TOKEN权限
2. 包是否设为public
3. 网络连接

### 部署失败

**问题：SSH连接失败**

检查：
1. `DEPLOY_SSH_KEY` secret是否完整
2. 服务器上的authorized_keys
3. 防火墙设置

**问题：Docker拉取失败**

检查：
1. 服务器能否访问ghcr.io
2. 是否需要登录（私有镜像）
3. 磁盘空间是否充足

## 📚 参考文档

- [GitHub Actions配置](.github/SETUP.md)
- [部署指南](docs/DEPLOYMENT.md)
- [发布模板](.github/RELEASE_TEMPLATE.md)
- [变更日志](CHANGELOG.md)

## 🆘 获取帮助

- **文档问题**: 查看 `docs/` 目录
- **技术问题**: 提交 GitHub Issue
- **部署问题**: 参考 `docs/DEPLOYMENT.md`

---

**祝发布顺利！** 🎉

如有任何问题，请查看详细文档或提交Issue。
