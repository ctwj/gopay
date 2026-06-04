# GoPay 部署指南

本文档指导您在各平台上部署 GoPay 聚合支付网关系统。

## 环境要求

### 应用运行环境

GoPay 已编译为独立静态二进制文件，**无需安装 Go 或 Node.js**：

- 最低配置：1 核 CPU，512MB 内存，100MB 磁盘空间
- 需要网络访问（接收支付回调通知）

### 数据库：PostgreSQL

生产环境**必须使用 PostgreSQL** 作为数据库：

| 要求 | 最低版本 | 推荐 |
|------|----------|------|
| PostgreSQL | 14+ | 16 |

安装 PostgreSQL：

```bash
# Ubuntu / Debian
sudo apt update
sudo apt install -y postgresql postgresql-contrib

# CentOS / RHEL
sudo yum install -y postgresql-server postgresql-contrib

# Docker
docker run -d \
  --name gopay-postgres \
  -e POSTGRES_USER=gopay \
  -e POSTGRES_PASSWORD=your_secure_password \
  -e POSTGRES_DB=gopay \
  -p 5432:5432 \
  -v pgdata:/var/lib/postgresql/data \
  postgres:16
```

创建数据库和用户：

```bash
# 登录 PostgreSQL
sudo -u postgres psql

# 创建用户和数据库
CREATE USER gopay WITH PASSWORD 'your_secure_password';
CREATE DATABASE gopay OWNER gopay;
GRANT ALL PRIVILEGES ON DATABASE gopay TO gopay;
\q
```

## 快速开始

### 1. 准备数据库

确保 PostgreSQL 服务已运行并创建了数据库（见上方说明）。

### 2. 下载

前往 [GitHub Releases](../../releases) 页面，下载对应平台的最新版本：

| 平台 | 文件名 |
|------|--------|
| Linux CLI | `gopay-X.X.X-linux-amd64` |
| Linux GUI（系统托盘）| `gopay-X.X.X-linux-gui-amd64` |
| Windows GUI（系统托盘）| `gopay-X.X.X-windows-amd64.exe` |
| macOS Apple Silicon | `gopay-X.X.X-macos-arm64` |

> 其中 `X.X.X` 为版本号，如 `1.2.0`。

### 3. 启动（连接 PostgreSQL）

```bash
./gopay -db "host=127.0.0.1 port=5432 user=gopay password=your_secure_password dbname=gopay sslmode=disable" -host 0.0.0.0 -port 8080
```

启动后访问 `http://你的IP:8080` 进入系统。

---

## Linux 部署

### CLI 版本（推荐服务器部署）

```bash
# 1. 下载（替换版本号）
wget https://github.com/your-repo/gopay/releases/download/vX.X.X/gopay-X.X.X-linux-amd64

# 2. 赋予执行权限
chmod +x gopay-*-linux-amd64

# 3. 启动（连接 PostgreSQL）
./gopay-X.X.X-linux-amd64 \
  -db "host=127.0.0.1 port=5432 user=gopay password=your_secure_password dbname=gopay sslmode=disable" \
  -host 0.0.0.0 \
  -port 8080
```

### GUI 版本（带系统托盘）

```bash
# 1. 下载
wget https://github.com/your-repo/gopay/releases/download/vX.X.X/gopay-X.X.X-linux-gui-amd64

# 2. 赋予执行权限
chmod +x gopay-*-linux-gui-amd64

# 3. 启动
./gopay-X.X.X-linux-gui-amd64 \
  -db "host=127.0.0.1 port=5432 user=gopay password=your_secure_password dbname=gopay sslmode=disable"
```

> **注意**: GUI 版本需要桌面环境（X11/Wayland）支持。

### 配置为系统服务（推荐）

使用 systemd 管理 GoPay 后台运行：

```bash
# 1. 移动二进制文件
sudo mv gopay-*-linux-amd64 /usr/local/bin/gopay

# 2. 创建配置文件（存储数据库连接串等敏感信息）
sudo mkdir -p /etc/gopay
sudo tee /etc/gopay/config.env > /dev/null << 'EOF'
GOPAY_DB=host=127.0.0.1 port=5432 user=gopay password=your_secure_password dbname=gopay sslmode=disable
GOPAY_HOST=0.0.0.0
GOPAY_PORT=8080
EOF
sudo chmod 600 /etc/gopay/config.env

# 3. 创建 systemd 服务文件
sudo tee /etc/systemd/system/gopay.service > /dev/null << 'EOF'
[Unit]
Description=GoPay Payment Gateway
After=network.target postgresql.service
Requires=postgresql.service

[Service]
Type=simple
User=gopay
WorkingDirectory=/opt/gopay
EnvironmentFile=/etc/gopay/config.env
ExecStart=/usr/local/bin/gopay -db ${GOPAY_DB} -host ${GOPAY_HOST} -port ${GOPAY_PORT}
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# 4. 创建运行用户和数据目录
sudo useradd -r -s /bin/false gopay
sudo mkdir -p /opt/gopay
sudo chown gopay:gopay /opt/gopay
sudo chown gopay:gopay /etc/gopay/config.env

# 5. 启动服务
sudo systemctl daemon-reload
sudo systemctl enable gopay
sudo systemctl start gopay

# 6. 查看状态
sudo systemctl status gopay
```

### PostgreSQL 连接参数说明

`-db` 参数接受 PostgreSQL 连接字符串，格式为 `key=value` 键值对：

| 参数 | 说明 | 示例 |
|------|------|------|
| `host` | 数据库主机地址 | `127.0.0.1` |
| `port` | 数据库端口 | `5432` |
| `user` | 数据库用户名 | `gopay` |
| `password` | 数据库密码 | `your_secure_password` |
| `dbname` | 数据库名称 | `gopay` |
| `sslmode` | SSL 模式 | `disable`（本地）/ `require`（远程） |
| `TimeZone` | 数据库时区 | `Asia/Shanghai` |

**远程数据库示例**：

```bash
./gopay -db "host=db.example.com port=5432 user=gopay password=your_secure_password dbname=gopay sslmode=require"
```

### PostgreSQL 安全建议

```bash
# 1. 限制数据库仅本地访问（/etc/postgresql/16/main/postgresql.conf）
listen_addresses = 'localhost'

# 2. 设置密码加密（/etc/postgresql/16/main/pg_hba.conf）
# 将 ident 改为 scram-sha-256
local   all   gopay   scram-sha-256
host    all   gopay   127.0.0.1/32   scram-sha-256

# 3. 重启 PostgreSQL
sudo systemctl restart postgresql
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

## Windows 部署

### GUI 版本

1. 从 [Releases](../../releases) 下载 `gopay-X.X.X-windows-amd64.exe`
2. 将文件放到目标目录（如 `C:\GoPay\`）
3. **命令行启动**（连接 PostgreSQL）：

```cmd
gopay-X.X.X-windows-amd64.exe -db "host=127.0.0.1 port=5432 user=gopay password=your_secure_password dbname=gopay sslmode=disable" -host 0.0.0.0 -port 8080
```

启动后系统托盘会出现 GoPay 图标，浏览器自动打开管理页面。

### Windows 防火墙

首次启动时 Windows 可能弹出防火墙提示，请选择 **允许访问**。

如需手动放行：

1. 控制面板 → Windows Defender 防火墙 → 高级设置
2. 入站规则 → 新建规则 → 端口 → TCP 8080 → 允许连接

### 开机自启（可选）

**方式一：任务计划程序**

1. Win+R → 输入 `taskschd.msc` → 回车
2. 创建基本任务 → 名称输入 `GoPay`
3. 触发器选择"计算机启动时"
4. 操作选择"启动程序"，浏览选择 `gopay-X.X.X-windows-amd64.exe`
5. 添加参数：`-db "host=127.0.0.1 port=5432 user=gopay password=your_secure_password dbname=gopay sslmode=disable"`
6. 完成

**方式二：启动文件夹**

按 Win+R → 输入 `shell:startup` → 将 GoPay 的快捷方式放入打开的文件夹。

---

## macOS 部署

### 下载

下载 `gopay-X.X.X-macos-arm64`（Apple Silicon）。

> 不确定？点击左上角  → 关于本机 → 查看"芯片"信息。

### 安装与启动

```bash
# 1. 下载
curl -LO https://github.com/your-repo/gopay/releases/download/vX.X.X/gopay-X.X.X-macos-arm64

# 2. 赋予执行权限
chmod +x gopay-*-macos-*

# 3. 启动（连接 PostgreSQL）
./gopay-X.X.X-macos-arm64 \
  -db "host=127.0.0.1 port=5432 user=gopay password=your_secure_password dbname=gopay sslmode=disable"
```

### macOS 安全提示处理

首次运行可能提示"无法验证开发者"：

1. **方式一**：右键点击文件 → 选择"打开" → 在弹窗中点击"打开"
2. **方式二**：系统设置 → 隐私与安全性 → 在底部找到被阻止的应用 → 点击"仍要打开"
3. **方式三（命令行）**：
   ```bash
   xattr -cr gopay-*-macos-*
   ```

---

## 启动参数说明

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-db` | PostgreSQL 连接字符串 | 无（**必须指定**） |
| `-host` | 监听 IP 地址 | `0.0.0.0`（所有网卡） |
| `-port` | 监听端口 | `8080` |
| `-migrate` | 执行数据库迁移（版本升级时使用） | `false` |

**示例**：

```bash
# 基本 PostgreSQL 连接
./gopay -db "host=127.0.0.1 user=gopay password=secret dbname=gopay sslmode=disable"

# 远程 PostgreSQL + 自定义端口
./gopay -db "host=db.example.com port=5433 user=gopay password=secret dbname=gopay sslmode=require" -port 3000

# 升级时执行数据库迁移
./gopay -db "host=127.0.0.1 user=gopay password=secret dbname=gopay sslmode=disable" -migrate
```

---

## 升级指南

### 标准升级流程

1. **备份数据库**
   ```bash
   # PostgreSQL 全库备份
   sudo -u postgres pg_dump gopay > gopay_backup_$(date +%Y%m%d).sql

   # 或使用自定义格式（推荐，支持并行恢复）
   sudo -u postgres pg_dump -Fc gopay > gopay_backup_$(date +%Y%m%d).dump
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
   source /etc/gopay/config.env
   /usr/local/bin/gopay -db "$GOPAY_DB" -migrate
   ```

6. **启动新版本**
   ```bash
   sudo systemctl start gopay
   ```

7. **验证** — 访问 `http://你的IP:8080` 确认系统正常运行

### 数据库恢复

如升级后出现问题，可恢复备份：

```bash
# 从 SQL 文本恢复
sudo -u postgres psql gopay < gopay_backup_YYYYMMDD.sql

# 从自定义格式恢复
sudo -u postgres pg_restore -d gopay gopay_backup_YYYYMMDD.dump
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

# 4. 检查连接字符串中的密码、用户名是否正确
```

### 端口被占用

**现象**: 启动时提示 `bind: address already in use`

**解决**:
```bash
# 查看占用端口的进程
lsof -i :8080

# 解决方式一：更换端口
./gopay -port 8081

# 解决方式二：结束占用进程
kill <PID>
```

### 权限不足

**现象**: 启动时提示 `Permission denied`

**解决**:
```bash
chmod +x gopay-*
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
1. 确认使用 `-host 0.0.0.0` 启动（而非 `127.0.0.1`）
2. 检查防火墙是否放行了端口
3. 检查云服务器安全组是否放行了端口
4. 确认服务器 IP 地址正确

### Windows 被杀毒软件拦截

**解决**:
1. 将 GoPay 添加到杀毒软件白名单/排除列表
2. 或临时关闭杀毒软件的实时防护

### macOS 提示"已损坏无法打开"

**解决**:
```bash
xattr -cr gopay-*-macos-*
```

### PostgreSQL 性能优化（可选）

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
