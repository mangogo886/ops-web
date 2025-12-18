# 档案建档统计系统 - 用户认证与权限管理

## 系统概述

本系统是一个档案建档统计管理系统，支持数据导入、统计查询、审核管理等功能。

## 功能模块

1. **统计信息** - 按分局和视频类型统计建档数据
2. **建档明细** - 查看和管理建档明细数据
3. **档案审核** - 审核进度管理和月度建档数据统计
4. **用户管理** - 用户信息管理（仅管理员）
5. **操作日志** - 查看系统操作日志（仅管理员）

## 快速开始

### 1. 环境要求

- Go 1.16 或更高版本
- MySQL 5.7 或更高版本
- Windows/Linux 操作系统

### 2. 数据库初始化

#### 2.1 创建数据库

```sql
CREATE DATABASE IF NOT EXISTS `ops` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

#### 2.2 执行SQL脚本

按以下顺序执行SQL脚本（位于 `deploy` 目录）：

1. **创建用户和角色表**
   ```bash
   mysql -u root -p ops < deploy/create-user-tables.sql
   ```

2. **创建基础数据表**（可选，如果使用审核功能）
   ```bash
   mysql -u root -p ops < deploy/create-table.sql
   ```

3. **创建审核相关表**（如果使用审核功能）
   ```bash
   mysql -u root -p ops < deploy/create-audit-tables.sql
   ```

4. **添加审核状态字段**（如果表已存在，需要添加字段）
   ```bash
   mysql -u root -p ops < deploy/alter-audit-details-add-status.sql
   ```

#### 2.3 初始化管理员用户

执行 `deploy/create-user-tables.sql` 后，系统会自动创建默认管理员用户：

- **用户名**: `admin`
- **密码**: `admin123`
- **角色**: 管理员（role_code=0）

**⚠️ 重要提示**: 首次登录后请立即修改默认密码！

### 3. 生成自定义密码Hash（可选）

如果需要为管理员设置其他密码，可以使用脚本生成密码Hash：

```bash
go run scripts/init-admin.go <your_password>
```

例如：
```bash
go run scripts/init-admin.go MySecurePassword123
```

输出示例：
```
密码: MySecurePassword123
BCrypt Hash: $2a$10$xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

SQL插入语句（管理员用户，role_id=1对应role_code=0的管理员角色）:
INSERT INTO `users` (`username`, `password`, `role_id`) VALUES
('admin', '$2a$10$xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx', 1);

注意：role_id=1 对应 user_role 表中 role_code=0 的管理员角色
```

然后将输出的SQL语句在数据库中执行，更新管理员密码。

### 4. 配置文件

复制并编辑配置文件：

```bash
cp deploy/config/config.json config/config.json
```

编辑 `config/config.json`，设置数据库连接信息：

```json
{
  "db_host": "127.0.0.1",
  "db_port": "3306",
  "db_user": "root",
  "db_pass": "your_password",
  "db_name": "ops"
}
```

### 5. 编译和运行

#### 开发环境

```bash
go run main.go
```

#### 生产环境

**Windows:**
```bash
go build -o ops-web.exe .
ops-web.exe
```

**Linux:**
```bash
go build -o ops-web .
./ops-web
```

### 6. 访问系统

启动后访问：http://127.0.0.1:8080/login

使用默认管理员账号登录：
- 用户名：`admin`
- 密码：`admin123`

## 用户角色说明

系统支持两种角色：

1. **管理员** (role_code=0, role_id=1)
   - 可以访问所有功能
   - 可以管理用户（添加、编辑、删除）
   - 可以查看操作日志

2. **普通用户** (role_code=1, role_id=2)
   - 可以查看统计信息和建档明细
   - 可以导入和导出数据
   - 可以查看审核进度和月度建档数据
   - 不能管理用户和查看操作日志

## 路由说明

### 认证路由（无需登录）
- `GET /login` - 显示登录页面
- `POST /login` - 处理登录请求
- `GET /logout` - 退出登录

### 受保护路由（需要登录）
- `GET /stats` - 统计信息页面
- `GET /stats/export` - 导出统计信息
- `GET /filelist` - 建档明细列表
- `GET /filelist/export` - 导出建档明细
- `GET /audit/progress` - 审核进度列表
- `POST /audit/progress/import` - 导入审核档案
- `GET /audit/progress/detail` - 查看档案明细
- `GET /audit/progress/edit` - 编辑审核意见
- `POST /audit/progress/edit` - 保存审核意见
- `POST /audit/progress/delete` - 删除审核档案
- `GET /audit/progress/download-template` - 下载导入模板
- `GET /audit/statistics` - 月度建档数据
- `GET /audit/statistics/export` - 导出月度建档数据

### 管理员路由（需要管理员权限）
- `GET /users` - 用户列表
- `POST /users/add` - 添加用户
- `POST /users/edit` - 编辑用户
- `POST /users/delete` - 删除用户
- `GET /logs` - 操作日志列表

## 数据库表结构

### user_role 表（用户角色表）
- `id` - 角色ID（主键）
- `role_name` - 角色名称
- `role_code` - 角色代码（0=管理员，1=普通用户）
- `permissions` - 角色权限（JSON格式，预留）
- `create_time` - 创建时间
- `update_time` - 更新时间

### users 表（用户表）
- `id` - 用户ID（主键）
- `username` - 用户名（唯一）
- `password` - 密码（bcrypt加密）
- `role_id` - 角色ID（外键关联user_role表）
  - role_id=1 对应 role_code=0（管理员）
  - role_id=2 对应 role_code=1（普通用户）
- `create_time` - 创建时间
- `update_time` - 更新时间

### operation_logs 表（操作日志表）
- `id` - 日志ID（主键）
- `username` - 操作用户
- `action` - 操作描述
- `ip` - 来源IP
- `created_at` - 操作时间

### audit_tasks 表（审核任务表）
- `id` - 任务ID（主键）
- `file_name` - 档案名称
- `organization` - 机构/子公司名称
- `import_time` - 导入时间
- `audit_status` - 审核状态（未审核、已审核待整改、已完成）
- `audit_comment` - 审核意见
- `created_at` - 创建时间
- `updated_at` - 更新时间

### audit_details 表（审核明细表）
- `id` - 明细ID（主键）
- `task_id` - 审核任务ID（外键）
- `device_code` - 设备编码
- `device_name` - 设备名称
- `management_unit` - 管理单位
- `monitor_point_type` - 监控点位类型（1-4类）
- `camera_function_type` - 摄像机功能类型
- `audit_status` - 建档状态（0=未审核未建档，1=已审核未建档，2=已建档）
- `update_time` - 更新时间
- ...（其他字段见SQL文件）

## 日志系统

系统会在 `logs` 目录下自动创建日志文件，文件名格式：`ops-web-YYYY-MM-DD.log`

**日志记录规则**：
- 只在异常情况下记录日志（错误、失败等）
- 正常操作（查询、导入成功等）不记录日志
- 日志包含错误详情、SQL语句、参数等信息，便于排查问题

## 生产环境部署

### 1. 目录结构

建议的生产环境目录结构：

```
/opt/ops-web/
├── ops-web              # 可执行文件
├── config/
│   └── config.json      # 配置文件
├── templates/           # HTML模板
├── deploy/              # 部署相关文件
│   ├── *.sql           # SQL脚本
│   └── config/         # 配置示例
└── logs/                # 日志目录（自动创建）
```

### 2. 配置文件

将 `deploy/config/config.json` 复制到 `config/config.json` 并修改数据库连接信息。

### 3. 数据库初始化

在生产数据库上执行所有SQL脚本（按顺序）：
1. `create-user-tables.sql` - 创建用户和角色表
2. `create-table.sql` - 创建基础数据表（如需要）
3. `create-audit-tables.sql` - 创建审核相关表（如需要）
4. `alter-audit-details-add-status.sql` - 添加审核状态字段（如需要）

### 4. 启动服务

**方式一：直接运行**
```bash
./ops-web
```

**方式二：使用systemd（Linux）**

创建服务文件 `/etc/systemd/system/ops-web.service`：

```ini
[Unit]
Description=Ops Web Service
After=network.target mysql.service

[Service]
Type=simple
User=ops
WorkingDirectory=/opt/ops-web
ExecStart=/opt/ops-web/ops-web
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

启动服务：
```bash
sudo systemctl daemon-reload
sudo systemctl enable ops-web
sudo systemctl start ops-web
```

查看状态：
```bash
sudo systemctl status ops-web
```

查看日志：
```bash
sudo journalctl -u ops-web -f
```

**方式三：使用nohup（简单方式）**
```bash
nohup ./ops-web > /dev/null 2>&1 &
```

### 5. 反向代理（可选）

如果需要使用域名或HTTPS，可以配置Nginx反向代理：

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

## 注意事项

1. **Session 管理**：当前使用内存存储 session，重启服务后 session 会失效。生产环境建议使用 Redis。

2. **密码安全**：
   - 密码使用 bcrypt 加密存储，默认 cost=10
   - 首次登录后请立即修改默认管理员密码
   - 定期更换密码

3. **数据库安全**：
   - 生产环境请使用专用数据库用户，不要使用root
   - 为数据库用户设置最小权限
   - 定期备份数据库

4. **日志管理**：
   - 日志文件按日期自动分割
   - 建议定期清理旧日志文件
   - 日志包含敏感信息，注意权限控制

5. **文件上传**：
   - 导入文件大小限制为32MB
   - 建议定期清理临时文件

6. **性能优化**：
   - 数据库连接池已配置（最大20个连接）
   - 建议为常用查询字段添加索引
   - 大数据量查询建议使用分页

## 故障排查

### 1. 无法连接数据库

检查：
- 配置文件 `config/config.json` 中的数据库连接信息是否正确
- 数据库服务是否运行
- 数据库用户是否有权限访问指定数据库
- 防火墙是否允许连接

查看日志：`logs/ops-web-YYYY-MM-DD.log`

### 2. 页面无法访问

检查：
- 服务是否启动（端口8080是否被占用）
- 防火墙是否开放8080端口
- 查看控制台输出或日志文件

### 3. 导入数据失败

检查：
- Excel文件格式是否正确（73列）
- 必填字段是否都有值
- 数据类型是否正确（经度、纬度、日期等）
- 查看日志文件获取详细错误信息

### 4. 查询无数据

检查：
- 数据库中是否有数据
- 查询条件是否正确
- `monitor_point_type` 字段值是否为1-4（统计信息和月度建档数据只统计1-4类）

## 技术支持

如遇到问题，请查看：
1. 日志文件：`logs/ops-web-YYYY-MM-DD.log`
2. 数据库操作日志：系统内"操作日志"页面
3. 控制台输出

## 更新日志

### v1.0.0
- 初始版本
- 支持用户认证和权限管理
- 支持统计信息、建档明细、审核管理等功能
- 支持数据导入导出
- 支持操作日志记录






