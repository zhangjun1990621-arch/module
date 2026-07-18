# 部署指南

## 1. 部署架构

```
┌──────────────────────────────────────────────────┐
│                 远程服务器                         │
│                                                   │
│  ┌─────────────────────────────────────────────┐ │
│  │  Docker Network: module_net                  │ │
│  │                                               │ │
│  │  ┌──────────┐  ┌──────────┐  ┌───────────┐  │ │
│  │  │ module-  │  │ module-  │  │ module-   │  │ │
│  │  │ postgres │  │ backend  │  │ frontend  │  │ │
│  │  │ :5432    │  │ :8090    │  │ :80       │  │ │
│  │  └──────────┘  └──────────┘  └─────┬─────┘  │ │
│  │                                      │        │ │
│  │  Volume: module_pg_data              │        │ │
│  └──────────────────────────────────────┼────────┘ │
│                                         │          │
│  宿主机端口映射: 8099 → 80 (Nginx)      │          │
│                  8090 → 8090 (后端)     │          │
│                  5433 → 5432 (PG)       │          │
└─────────────────────────────────────────┘──────────┘
```

### 端口规划（与原项目完全隔离）

| 服务 | 容器端口 | 宿主机端口 | 原项目端口 | 说明 |
|------|---------|-----------|-----------|------|
| PostgreSQL | 5432 | **5433** | 5432 | 不同端口 |
| Go 后端 | 8090 | **8090** | 8080 | 不同端口 |
| Nginx 前端 | 80 | **8099** | 8089 | 不同端口 |

---

## 2. Docker 一键部署（推荐）

### 2.1 前置条件

- Docker 20.10+
- Docker Compose 2.0+
- 服务器至少 2GB 内存、10GB 磁盘

### 2.2 部署步骤

```bash
# 1. 克隆代码到服务器
cd /data
git clone https://github.com/zhangjun1990621-arch/module.git
cd module

# 2. 进入部署目录
cd deploy

# 3. 构建并启动（首次约 5-10 分钟）
docker compose up -d --build

# 4. 查看服务状态
docker compose ps

# 5. 查看日志（确认启动成功）
docker compose logs -f backend
# 应看到: "IoT 平台后端服务启动，监听 :8090"

# 6. 健康检查
curl http://localhost:8090/api/health
# 应返回: {"status":"ok"}

# 7. 访问前端
# 浏览器打开 http://服务器IP:8099
# 默认账号: admin / admin123
```

### 2.3 更新部署

```bash
cd /data/module

# 拉取最新代码
git pull origin main

# 重新构建并启动（仅重建变更的服务）
cd deploy
docker compose up -d --build backend frontend

# 如果数据库有变更，需手动执行迁移
docker exec -i module-postgres psql -U postgres -d iot_platform < ../backend/migrations/init.sql
```

---

## 3. 手动部署（无 Docker）

### 3.1 安装 PostgreSQL

```bash
# Ubuntu/Debian
sudo apt install postgresql-16

# CentOS/RHEL
sudo yum install postgresql16-server

# 创建数据库
sudo -u postgres psql -c "CREATE DATABASE iot_platform;"
sudo -u postgres psql -c "ALTER USER postgres PASSWORD 'Qsh@2026#PvSecure';"

# 初始化表结构
psql -h 127.0.0.1 -U postgres -d iot_platform -f backend/migrations/init.sql
```

### 3.2 编译后端

```bash
cd backend

# 修改 config/config.yaml 中的数据库连接信息
vim config/config.yaml

# 编译
CGO_ENABLED=0 go build -ldflags="-s -w" -o server ./cmd/server

# 运行
./server
# 或使用 systemd 管理
```

### 3.3 构建前端

```bash
cd frontend
npm install
npm run build
# 产物在 dist/ 目录

# 使用 Nginx 托管
cp -r dist/* /var/www/module/
# 配置 Nginx 反代 /api → 后端
```

### 3.4 Nginx 配置

```nginx
server {
    listen 8099;
    server_name _;

    root /var/www/module;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://127.0.0.1:8090;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

---

## 4. 环境变量配置

Docker Compose 支持通过环境变量覆盖 `config.yaml`：

| 环境变量 | 默认值 | 说明 |
|---------|--------|------|
| DB_HOST | postgres | 数据库主机 |
| DB_PORT | 5432 | 数据库端口 |
| DB_USER | postgres | 数据库用户 |
| DB_PASSWORD | Qsh@2026#PvSecure | 数据库密码 |
| DB_NAME | iot_platform | 数据库名 |
| JWT_SECRET | iot-platform-jwt-secret-2026 | JWT 密钥 |
| JWT_EXPIRE | 24 | Token 过期时间（小时） |
| SERVER_PORT | 8090 | 后端端口 |

---

## 5. 与原项目隔离检查清单

部署前请确认以下隔离项：

- [x] **Web 端口**：使用 8099（原项目 8089）
- [x] **数据库**：使用 iot_platform（原项目 pv_platform）
- [x] **PostgreSQL 端口**：使用 5433（原项目 5432）
- [x] **容器名**：module-postgres / module-backend / module-frontend
- [x] **Docker 网络**：module_net
- [x] **数据卷**：module_pg_data
- [x] **部署目录**：/data/module（原项目 /data/zonghe）
- [x] **Nginx 配置**：独立配置文件
- [x] **GitHub 仓库**：zhangjun1990621-arch/module.git

---

## 6. SSL/HTTPS 配置（可选）

```bash
# 使用 Let's Encrypt 申请证书
sudo certbot certonly --standalone -d your-domain.com

# 修改 nginx.conf
```

```nginx
server {
    listen 8099 ssl;
    server_name your-domain.com;

    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;

    # ... 其余配置同上
}
```
