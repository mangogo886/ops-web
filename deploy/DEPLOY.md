# 生产环境部署指南

## 前置要求

- Linux服务器（推荐Ubuntu 20.04+或CentOS 7+）
- MySQL 5.7+ 或 MariaDB 10.3+
- Go 1.16+（用于编译，生产环境可不需要）
- Nginx（可选，用于反向代理）

## 快速部署

### 方式一：使用安装脚本（推荐）

```bash
# 1. 进入deploy目录
cd deploy

# 2. 运行安装脚本
chmod +x install.sh
./install.sh
```

安装脚本会自动：
- 编译程序
- 创建配置文件和目录
- 提示数据库初始化

### 方式二：手动部署

#### 1. 编译程序

```bash
go build -o ops-web .
```

#### 2. 配置数据库

```bash
# 复制配置文件
cp deploy/config/config.json config/config.json

# 编辑配置文件
vi config/config.json
```

#### 3. 初始化数据库

```bash
# 使用初始化脚本
chmod +x deploy/init-database.sh
./deploy/init-database.sh ops root

# 或手动执行SQL
mysql -u root -p ops < deploy/create-user-tables.sql
mysql -u root -p ops < deploy/create-audit-tables.sql
```

#### 4. 启动服务

```bash
# 使用启动脚本
chmod +x deploy/start.sh
./deploy/start.sh

# 或直接运行
./ops-web
```

## 使用systemd管理服务

### 1. 创建系统用户

```bash
sudo useradd -r -s /bin/false ops
```

### 2. 部署文件

```bash
sudo mkdir -p /opt/ops-web
sudo cp ops-web /opt/ops-web/
sudo cp -r config templates deploy /opt/ops-web/
sudo chown -R ops:ops /opt/ops-web
```

### 3. 配置systemd服务

```bash
# 复制服务文件
sudo cp deploy/ops-web.service /etc/systemd/system/

# 编辑服务文件（如需要）
sudo vi /etc/systemd/system/ops-web.service

# 重新加载systemd
sudo systemctl daemon-reload

# 启用服务（开机自启）
sudo systemctl enable ops-web

# 启动服务
sudo systemctl start ops-web

# 查看状态
sudo systemctl status ops-web

# 查看日志
sudo journalctl -u ops-web -f
```

## 配置Nginx反向代理

### 1. 复制配置文件

```bash
sudo cp deploy/nginx.conf.example /etc/nginx/sites-available/ops-web
sudo vi /etc/nginx/sites-available/ops-web
```

### 2. 修改配置

- 将 `your-domain.com` 替换为实际域名
- 如需HTTPS，取消注释HTTPS配置部分并配置SSL证书

### 3. 启用配置

```bash
sudo ln -s /etc/nginx/sites-available/ops-web /etc/nginx/sites-enabled/
sudo nginx -t  # 测试配置
sudo systemctl reload nginx
```

## 防火墙配置

```bash
# Ubuntu/Debian
sudo ufw allow 8080/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# CentOS/RHEL
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --permanent --add-port=80/tcp
sudo firewall-cmd --permanent --add-port=443/tcp
sudo firewall-cmd --reload
```

## 日志管理

### 查看应用日志

```bash
# 实时查看
tail -f /opt/ops-web/logs/ops-web-$(date +%Y-%m-%d).log

# 查看最近100行
tail -n 100 /opt/ops-web/logs/ops-web-$(date +%Y-%m-%d).log
```

### 日志轮转

创建 `/etc/logrotate.d/ops-web`：

```
/opt/ops-web/logs/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 0644 ops ops
}
```

## 备份

### 数据库备份

```bash
# 创建备份脚本
cat > /opt/ops-web/backup-db.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/opt/backups/ops-web"
mkdir -p $BACKUP_DIR
mysqldump -u root -p ops > $BACKUP_DIR/ops-$(date +%Y%m%d-%H%M%S).sql
# 保留最近30天的备份
find $BACKUP_DIR -name "*.sql" -mtime +30 -delete
EOF

chmod +x /opt/ops-web/backup-db.sh

# 添加到crontab（每天凌晨2点备份）
echo "0 2 * * * /opt/ops-web/backup-db.sh" | crontab -
```

## 监控

### 健康检查

```bash
# 检查服务状态
curl http://127.0.0.1:8080/login

# 检查进程
ps aux | grep ops-web
```

### 资源监控

```bash
# 查看进程资源使用
top -p $(pgrep ops-web)

# 查看端口占用
netstat -tlnp | grep 8080
```

## 故障排查

### 服务无法启动

1. 检查日志：`/opt/ops-web/logs/ops-web-*.log`
2. 检查systemd日志：`sudo journalctl -u ops-web -n 50`
3. 检查配置文件：`cat /opt/ops-web/config/config.json`
4. 检查数据库连接：`mysql -u root -p -h 127.0.0.1 ops`

### 数据库连接失败

1. 检查MySQL服务：`sudo systemctl status mysql`
2. 检查配置文件中的数据库连接信息
3. 测试数据库连接：`mysql -u <user> -p -h <host> <database>`
4. 检查防火墙和网络

### 页面无法访问

1. 检查服务是否运行：`sudo systemctl status ops-web`
2. 检查端口是否监听：`netstat -tlnp | grep 8080`
3. 检查防火墙规则
4. 检查Nginx配置（如使用反向代理）

## 更新部署

```bash
# 1. 停止服务
sudo systemctl stop ops-web

# 2. 备份当前版本
sudo cp /opt/ops-web/ops-web /opt/ops-web/ops-web.backup

# 3. 编译新版本
go build -o ops-web .

# 4. 部署新版本
sudo cp ops-web /opt/ops-web/
sudo chown ops:ops /opt/ops-web/ops-web

# 5. 启动服务
sudo systemctl start ops-web

# 6. 检查状态
sudo systemctl status ops-web
```

## 安全建议

1. **数据库安全**
   - 使用专用数据库用户，不要使用root
   - 为数据库用户设置最小权限
   - 定期更新数据库密码

2. **文件权限**
   - 配置文件权限设置为600：`chmod 600 config/config.json`
   - 日志目录权限设置为755：`chmod 755 logs`

3. **网络安全**
   - 使用HTTPS（配置SSL证书）
   - 限制数据库访问（只允许本地或指定IP）
   - 使用防火墙限制访问

4. **系统安全**
   - 定期更新系统和依赖
   - 使用非root用户运行服务
   - 定期备份数据















