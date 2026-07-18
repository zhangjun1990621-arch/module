# 运维手册

## 1. 日常运维命令

### 1.1 服务管理

```bash
cd /data/module/deploy

# 查看所有服务状态
docker compose ps

# 启动所有服务
docker compose up -d

# 停止所有服务
docker compose down

# 重启单个服务
docker compose restart backend
docker compose restart frontend
docker compose restart postgres

# 重新构建并启动（代码更新后）
docker compose up -d --build backend frontend
```

### 1.2 日志查看

```bash
# 实时查看后端日志
docker compose logs -f backend

# 实时查看前端日志
docker compose logs -f frontend

# 实时查看数据库日志
docker compose logs -f postgres

# 查看最近 100 行日志
docker compose logs --tail 100 backend

# 查看指定时间后的日志
docker compose logs --since "2026-07-16T00:00:00" backend
```

### 1.3 进入容器

```bash
# 进入后端容器
docker exec -it module-backend sh

# 进入前端容器
docker exec -it module-frontend sh

# 进入数据库容器
docker exec -it module-postgres psql -U postgres -d iot_platform
```

---

## 2. 数据库运维

### 2.1 数据库连接

```bash
# 通过 Docker
docker exec -it module-postgres psql -U postgres -d iot_platform

# 通过宿主机（端口 5433）
psql -h 127.0.0.1 -p 5433 -U postgres -d iot_platform
```

### 2.2 常用查询

```sql
-- 查看所有平台
SELECT id, name, status, sort_order FROM platforms ORDER BY sort_order;

-- 查看所有用户
SELECT id, username, role, platforms, status FROM users;

-- 查看某平台的设备数
SET search_path TO schema_aluminum, public;
SELECT COUNT(*) FROM devices;
SELECT status, COUNT(*) FROM devices GROUP BY status;

-- 查看某平台的活跃告警
SET search_path TO schema_aluminum, public;
SELECT level, COUNT(*) FROM alarms WHERE status = 'active' GROUP BY level;

-- 查看数据库大小
SELECT pg_size_pretty(pg_database_size('iot_platform'));

-- 查看各 Schema 大小
SELECT schema_name, pg_size_pretty(sum(table_size)) as size
FROM (
  SELECT pg_catalog.pg_namespace.nspname as schema_name,
         pg_relation_size(pg_catalog.pg_class.oid) as table_size
  FROM pg_catalog.pg_class
  JOIN pg_catalog.pg_namespace ON pg_class.relnamespace = pg_namespace.oid
  WHERE pg_namespace.nspname LIKE 'schema_%'
) t
GROUP BY schema_name
ORDER BY sum(table_size) DESC;
```

### 2.3 数据备份

```bash
# 创建备份目录
mkdir -p /data/module/backups

# 全库备份
docker exec module-postgres pg_dump -U postgres iot_platform > /data/module/backups/backup_$(date +%Y%m%d_%H%M%S).sql

# 压缩备份
docker exec module-postgres pg_dump -U postgres -Fc iot_platform > /data/module/backups/backup_$(date +%Y%m%d).dump

# 仅备份公共表
docker exec module-postgres pg_dump -U postgres -t platforms -t users iot_platform > /data/module/backups/public_tables.sql
```

### 2.4 数据恢复

```bash
# 恢复全库
cat /data/module/backups/backup_20260716.sql | docker exec -i module-postgres psql -U postgres -d iot_platform

# 恢复压缩备份
docker exec -i module-postgres pg_restore -U postgres -d iot_platform < /data/module/backups/backup_20260716.dump
```

### 2.5 定时备份

```bash
# 创建定时备份脚本
cat > /data/module/backup.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/data/module/backups"
mkdir -p $BACKUP_DIR
docker exec module-postgres pg_dump -U postgres iot_platform | gzip > $BACKUP_DIR/backup_$(date +%Y%m%d_%H%M%S).sql.gz
# 保留最近 30 天的备份
find $BACKUP_DIR -name "backup_*.sql.gz" -mtime +30 -delete
EOF
chmod +x /data/module/backup.sh

# 添加定时任务（每天凌晨 3 点备份）
crontab -e
# 添加: 0 3 * * * /data/module/backup.sh
```

---

## 3. 性能优化

### 3.1 PostgreSQL 优化

```sql
-- 查看慢查询
SELECT query, mean_exec_time, calls
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;

-- 查看索引使用情况
SELECT schemaname, relname, indexrelname, idx_scan
FROM pg_stat_user_indexes
WHERE schemaname LIKE 'schema_%'
ORDER BY idx_scan;

-- 分析表（更新统计信息）
ANALYZE schema_aluminum.devices;
ANALYZE schema_aluminum.alarms;
```

### 3.2 连接池配置

后端使用 GORM 内置连接池。如需调整，修改 `database/db.go`：

```go
sqlDB, _ := db.DB()
sqlDB.SetMaxIdleConns(10)    // 最大空闲连接
sqlDB.SetMaxOpenConns(100)   // 最大连接数
sqlDB.SetConnMaxLifetime(time.Hour)  // 连接最大生命周期
```

---

## 4. 监控

### 4.1 健康检查

```bash
# 后端健康检查
curl http://localhost:8090/api/health
# 预期: {"status":"ok"}

# 前端可访问性
curl -o /dev/null -s -w "%{http_code}" http://localhost:8099
# 预期: 200

# 数据库连接检查
docker exec module-postgres pg_isready -U postgres -d iot_platform
# 预期: accepting connections
```

### 4.2 资源监控

```bash
# 查看 Docker 容器资源占用
docker stats module-postgres module-backend module-frontend

# 查看磁盘空间
df -h /data

# 查看内存
free -h
```

---

## 5. 故障排查

### 5.1 后端启动失败

```bash
# 查看后端日志
docker compose logs backend

# 常见原因：
# 1. 数据库未就绪 → 检查 postgres 容器状态
# 2. 配置文件错误 → 检查 config.yaml
# 3. 端口被占用 → netstat -tlnp | grep 8090
```

### 5.2 数据库连接失败

```bash
# 检查 PostgreSQL 容器
docker compose ps postgres
docker compose logs postgres

# 手动连接测试
docker exec -it module-postgres psql -U postgres -d iot_platform -c "SELECT 1;"
```

### 5.3 前端页面空白

```bash
# 检查 Nginx 容器
docker compose logs frontend

# 检查前端构建产物
docker exec -it module-frontend ls /usr/share/nginx/html

# 检查 Nginx 配置
docker exec -it module-frontend cat /etc/nginx/conf.d/default.conf
```

### 5.4 API 返回 500

```bash
# 查看后端日志
docker compose logs --tail 50 backend

# 检查数据库 Schema 是否存在
docker exec module-postgres psql -U postgres -d iot_platform -c "\dn"

# 检查平台表数据
docker exec module-postgres psql -U postgres -d iot_platform -c "SELECT id, schema_name, status FROM platforms;"
```

### 5.5 cached plan 错误

如果出现 `cached plan must not change result type`：
- 已通过 `PreferSimpleProtocol: true` 解决
- 如仍出现，检查后端代码是否使用了正确的 GORM 配置

---

## 6. 升级流程

### 6.1 滚动升级

```bash
cd /data/module

# 1. 备份数据库
docker exec module-postgres pg_dump -U postgres iot_platform > /data/module/backups/pre_upgrade_$(date +%Y%m%d).sql

# 2. 拉取最新代码
git pull origin main

# 3. 重新构建后端和前端（不影响数据库）
docker compose -f deploy/docker-compose.yml up -d --build backend frontend

# 4. 验证
curl http://localhost:8090/api/health
curl -o /dev/null -s -w "%{http_code}" http://localhost:8099
```

### 6.2 数据库迁移

如果 `migrations/init.sql` 有变更：

```bash
# 执行迁移（init.sql 使用 IF NOT EXISTS，可安全重复执行）
docker exec -i module-postgres psql -U postgres -d iot_platform < /data/module/backend/migrations/init.sql
```

---

## 7. 安全加固

### 7.1 修改默认密码

```sql
-- 修改 admin 密码（在权限管理页面操作，或直接 SQL）
-- 生成新密码的 bcrypt 哈希:
-- Python: python -c "import bcrypt; print(bcrypt.hashpw(b'newpassword', bcrypt.gensalt()).decode())"
UPDATE users SET password = '$2a$10$新哈希值' WHERE username = 'admin';
```

### 7.2 修改 JWT Secret

```bash
# 修改 docker-compose.yml 中的 JWT_SECRET
# 或修改 backend/config/config.yaml
```

### 7.3 限制数据库端口

生产环境建议不暴露 PostgreSQL 端口：
```yaml
# docker-compose.yml
postgres:
  # ports:
  #   - "5433:5432"  # 注释掉，仅容器内访问
  expose:
    - "5432"
```

### 7.4 Nginx HTTPS

参考 [DEPLOYMENT.md](DEPLOYMENT.md) 中的 SSL 配置部分。
