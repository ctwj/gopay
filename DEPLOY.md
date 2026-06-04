# GoPay 部署指南

本文档指导您在 Linux 服务器上部署 GoPay 聚合支付网关系统。

## 环境要求

### 应用运行环境

GoPay 已编译为独立静态二进制文件，**无需安装 Go 或 Node.js**：

- 最低配置：1 核 CPU，512MB 内存，100MB 磁盘空间
- 需要网络访问（接收支付回调通知）

### 数据库

GoPay 支持 **SQLite** 和 **PostgreSQL** 两种数据库：

| 数据库 | 适用场景 | 最低版本 |
|--------|----------|----------|
| SQLite | 开发测试、小规模部署 | 内置，无需安装 |
| PostgreSQL | 生产环境推荐 | 14+ |

## 快速开始

### 1. 下载

前往 [GitHub Releases](../../releases) 页面，下载对应平台的最新版本：

```bash
# 下载（替换版本号）
wget https://github.com/your-repo/gopay/releases/download/vX.X.X/gopay-X.X.X-linux-amd64

# 赋予执行权限
chmod +x gopay-*-linux-amd64
```

### 2. 创建配置文件

在二进制文件同目录下创建 `gopay.env`：

**SQLite（开箱即用）**：

```bash
cat > gopay.env << 'EOF'
# 数据库类型: sqlite（默认）或 postgres（留空则自动检测）
DB_TYPE=sqlite

# SQLite 数据库文件路径
DB=./data/gopay.db

# 监听地址
HOST=0.0.0.0

# 监听端口
PORT=8080
EOF
```

**PostgreSQL**：

```bash
cat > gopay.env << 'EOF'
# 数据库类型: postgres
DB_TYPE=postgres

# PostgreSQL 连接字符串
DB=host=127.0.0.1 port=5432 user=gopay password=your_secure_password dbname=gopay sslmode=disable

# 监听地址
HOST=0.0.0.0

# 监听端口
PORT=8080
EOF
```

> **提示**：不设置 `DB_TYPE` 时，系统根据 `DB` 值的格式自动判断数据库类型。

### 3. 启动

```bash
./gopay-X.X.X-linux-amd64
```

启动后访问 `http://你的IP:8080` 进入系统。

---

## PostgreSQL 安装与配置

生产环境推荐使用 PostgreSQL。

### 安装 PostgreSQL

```bash
# Ubuntu / Debian
sudo apt update
sudo apt install -y postgresql postgresql-contrib

# CentOS / RHEL
sudo yum install -y postgresql-server postgresql-contrib
```

或使用 Docker：

```bash
docker run -d \
  --name gopay-postgres \
  -e POSTGRES_USER=gopay \
  -e POSTGRES_PASSWORD=your_secure_password \
  -e POSTGRES_DB=gopay \
  -p 5432:5432 \
  -v pgdata:/var/lib/postgresql/data \
  postgres:16
```

### 创建数据库和用户

```bash
# 登录 PostgreSQL
sudo -u postgres psql

# 创建用户和数据库
CREATE USER gopay WITH PASSWORD 'your_secure_password';
CREATE DATABASE gopay OWNER gopay;
GRANT ALL PRIVILEGES ON DATABASE gopay TO gopay;
\q
```

### PostgreSQL 连接参数说明

`DB` 配置项支持 `key=value` 格式的连接字符串：

| 参数 | 说明 | 示例 |
|------|------|------|
| `host` | 数据库主机地址 | `127.0.0.1` |
| `port` | 数据库端口 | `5432` |
| `user` | 数据库用户名 | `gopay` |
| `password` | 数据库密码 | `your_secure_password` |
| `dbname` | 数据库名称 | `gopay` |
| `sslmode` | SSL 模式 | `disable`（本地）/ `require`（远程） |
| `TimeZone` | 数据库时区 | `Asia/Shanghai` |

也支持 URL 格式：`postgres://user:password@host:port/dbname?sslmode=disable`

### PostgreSQL 安全建议

```bash
# 1. 限制数据库仅本地访问（/etc/postgresql/16/main/postgresql.conf）
listen_addresses = 'localhost'

# 2. 设置密码加密（/etc/postgresql/16/main/pg_hba.conf）
local   all   gopay   scram-sha-256
host    all   gopay   127.0.0.1/32   scram-sha-256

# 3. 重启 PostgreSQL
sudo systemctl restart postgresql
```

---

## 配置为系统服务

使用 systemd 管理 GoPay 后台运行：

```bash
# 1. 移动二进制文件
sudo mv gopay-*-linux-amd64 /usr/local/bin/gopay

# 2. 创建工作目录
sudo mkdir -p /opt/gopay
sudo chown gopay:gopay /opt/gopay

# 3. 创建配置文件
sudo tee /opt/gopay/gopay.env > /dev/null << 'EOF'
# 数据库类型: postgres（生产环境推荐）
DB_TYPE=postgres

# PostgreSQL 连接字符串
DB=host=127.0.0.1 port=5432 user=gopay password=your_secure_password dbname=gopay sslmode=disable

# 监听地址
HOST=0.0.0.0

# 监听端口
PORT=8080
EOF
sudo chmod 600 /opt/gopay/gopay.env

# 4. 创建运行用户
sudo useradd -r -s /bin/false gopay
sudo chown gopay:gopay /opt/gopay/gopay.env

# 5. 创建 systemd 服务文件
sudo tee /etc/systemd/system/gopay.service > /dev/null << 'EOF'
[Unit]
Description=GoPay Payment Gateway
After=network.target postgresql.service
Requires=postgresql.service

[Service]
Type=simple
User=gopay
WorkingDirectory=/opt/gopay
ExecStart=/usr/local/bin/gopay
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# 6. 启动服务
sudo systemctl daemon-reload
sudo systemctl enable gopay
sudo systemctl start gopay

# 7. 查看状态
sudo systemctl status gopay
```

### 防火墙配置

```bash
# Ubuntu / Debian
sudo ufw allow 8080/tcp

# CentOS / RHEL
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload
```

---

## 配置文件说明

GoPay 通过 `gopay.env` 配置文件管理所有配置，优先级为：

**命令行参数 > 配置文件 > 默认值**

### 配置文件搜索路径

系统按以下顺序查找配置文件：

1. 当前工作目录下的 `gopay.env`
2. 当前工作目录下的 `config.env`
3. 可执行文件同目录下的 `gopay.env`
4. 可执行文件同目录下的 `config.env`
5. Linux 系统级 `/etc/gopay/config.env`

也可通过 `-config` 参数指定配置文件路径。

### 配置项一览

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `DB_TYPE` | 数据库类型：`sqlite` 或 `postgres`，留空自动检测 | 自动检测 |
| `DB` | 数据库连接字符串（PostgreSQL DSN 或 SQLite 文件路径） | 自动检测 |
| `HOST` | 监听 IP 地址 | `0.0.0.0` |
| `PORT` | 监听端口 | `8080` |

### 自动检测规则

当 `DB_TYPE` 未设置时，系统根据 `DB` 值的格式自动判断：

| DB 值格式 | 判定结果 |
|-----------|----------|
| 以 `postgres://` 或 `postgresql://` 开头 | PostgreSQL |
| 包含 `host=` 且包含 `dbname=` | PostgreSQL |
| 其他（文件路径或空） | SQLite |

---

## 启动参数说明

配置文件中的所有配置项均可通过命令行参数覆盖：

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-config` | 指定配置文件路径 | 自动查找 |
| `-db` | 数据库连接字符串 | 读取配置文件 |
| `-db-type` | 数据库类型：`sqlite` 或 `postgres` | 自动检测 |
| `-host` | 监听 IP 地址 | `0.0.0.0` |
| `-port` | 监听端口 | `8080` |
| `-migrate` | 执行数据库迁移（版本升级时使用） | `false` |

**示例**：

```bash
# 使用配置文件启动（推荐）
./gopay

# 命令行覆盖配置文件中的数据库设置
./gopay -db-type postgres -db "host=127.0.0.1 user=gopay password=secret dbname=gopay sslmode=disable"

# 升级时执行数据库迁移
./gopay -migrate

# 指定配置文件路径
./gopay -config /etc/gopay/config.env
```

---

## 升级指南

### 标准升级流程

1. **备份数据库**

   SQLite：
   ```bash
   cp /opt/gopay/data/gopay.db /opt/gopay/backup/gopay_backup_$(date +%Y%m%d).db
   ```

   PostgreSQL：
   ```bash
   sudo -u postgres pg_dump gopay > /opt/gopay/backup/gopay_backup_$(date +%Y%m%d).sql
   ```

2. **停止服务**
   ```bash
   sudo systemctl stop gopay
   ```

3. **下载新版本** — 从 [Releases](../../releases) 下载最新版本

4. **替换二进制文件**
   ```bash
   mv /usr/local/bin/gopay /usr/local/bin/gopay.bak
   cp gopay-X.X.X-linux-amd64 /usr/local/bin/gopay
   chmod +x /usr/local/bin/gopay
   ```

5. **执行数据库迁移**（如有数据库变更）
   ```bash
   cd /opt/gopay
   /usr/local/bin/gopay -migrate
   ```

6. **启动新版本**
   ```bash
   sudo systemctl start gopay
   ```

7. **验证** — 访问 `http://你的IP:8080` 确认系统正常运行

### 数据库恢复

如升级后出现问题，可恢复备份：

```bash
# SQLite
cp /opt/gopay/backup/gopay_backup_YYYYMMDD.db /opt/gopay/data/gopay.db

# PostgreSQL
sudo -u postgres psql gopay < /opt/gopay/backup/gopay_backup_YYYYMMDD.sql
```

---

## 常见问题排查

### 数据库连接失败

**现象**: 启动时提示 `connection refused` 或 `password authentication failed`

**排查步骤**:
```bash
# 1. 确认 PostgreSQL 服务正在运行
sudo systemctl status postgresql

# 2. 测试连接
psql -h 127.0.0.1 -U gopay -d gopay

# 3. 检查 pg_hba.conf 是否允许密码认证
sudo cat /etc/postgresql/16/main/pg_hba.conf | grep gopay

# 4. 检查 gopay.env 中 DB 配置是否正确
cat /opt/gopay/gopay.env
```

### 端口被占用

**现象**: 启动时提示 `bind: address already in use`

**解决**:
```bash
# 查看占用端口的进程
lsof -i :8080

# 解决方式一：更换端口（修改 gopay.env 中的 PORT）
# 解决方式二：结束占用进程
kill <PID>
```

### 权限不足

**现象**: 启动时提示 `Permission denied`

**解决**:
```bash
chmod +x gopay-*
# 确保配置文件可读
chmod 644 /opt/gopay/gopay.env
```

### 数据库迁移失败

**现象**: `-migrate` 执行后报错

**解决**:
1. 检查数据库用户是否有 CREATE TABLE、ALTER TABLE 权限
2. 查看日志中的具体错误信息
3. 确认 PostgreSQL 版本是否符合要求（14+）
4. 如有备份数据，可恢复后重试

### 前端页面空白

**现象**: 访问页面显示空白

**解决**:
1. 确认启动日志中没有前端资源相关错误
2. 检查浏览器控制台（F12）是否有错误
3. 确认使用的是包含前端构建的完整 Release 版本

### 无法从外网访问

**排查步骤**:
1. 确认 `gopay.env` 中 `HOST=0.0.0.0`（而非 `127.0.0.1`）
2. 检查防火墙是否放行了端口
3. 检查云服务器安全组是否放行了端口
4. 确认服务器 IP 地址正确

---

## PostgreSQL 性能优化（可选）

生产环境建议调整以下 PostgreSQL 参数：

```bash
# /etc/postgresql/16/main/postgresql.conf
shared_buffers = 256MB          # 共享缓冲区（建议为内存的 25%）
effective_cache_size = 768MB    # 有效缓存大小（建议为内存的 75%）
work_mem = 4MB                  # 排序/哈希操作内存
maintenance_work_mem = 64MB     # 维护操作内存
max_connections = 100           # 最大连接数
```

修改后重启 PostgreSQL：

```bash
sudo systemctl restart postgresql
```
