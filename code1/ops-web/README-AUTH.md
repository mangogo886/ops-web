# 档案审核管理系统 - 开发与部署说明

## 系统简介

本系统是一个档案审核管理系统，支持设备档案和卡口档案的导入、审核、统计等功能。系统采用Go语言开发，使用MySQL数据库存储数据。

## 系统要求

### 开发环境
- Go 1.20 或更高版本
- MySQL 5.7+ 或 MariaDB 10.3+
- Windows 7+ / Linux / macOS

### 生产环境
- Windows 7 或更高版本（已编译exe文件）
- MySQL 5.7+ 或 MariaDB 10.3+
- 网络连接（用于访问数据库）

## 快速开始

### 1. 克隆或下载项目

```bash
git clone <repository-url>
cd ops-web
```

### 2. 安装依赖

```bash
go mod download
```

### 3. 配置数据库

编辑 `config/config.json` 文件，设置数据库连接信息：

```json
{
  "db_host": "127.0.0.1",
  "db_port": "3306",
  "db_user": "root",
  "db_pass": "your_password",
  "db_name": "ops",
  "server_host": "127.0.0.1",
  "server_port": "8080"
}
```

### 4. 初始化数据库

#### Windows 系统

```cmd
cd deploy
init-database.bat ops root your_password
```

#### Linux/macOS 系统

```bash
cd deploy
chmod +x init-database.sh
./init-database.sh ops root your_password
```

#### 手动执行SQL

如果自动脚本无法使用，可以手动执行SQL文件：

```bash
# 1. 创建数据库
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS ops DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 2. 创建所有表
mysql -u root -p ops < deploy/create-all-tables.sql

# 3. 初始化管理员账户
mysql -u root -p ops < deploy/init-admin-user.sql
```

### 5. 运行程序

#### 开发模式

```bash
go run main.go
```

#### 生产模式（Windows）

```cmd
cd deploy
ops-web-pro.exe
```

#### 生产模式（Linux）

```bash
./ops-web
```

### 6. 访问系统

打开浏览器访问：http://127.0.0.1:8080/login

## 第一次部署 - 创建管理员账户

### 方法一：使用初始化脚本（推荐）

执行 `deploy/init-admin-user.sql` 脚本，会自动创建默认管理员账户：

- **用户名**: `admin`
- **密码**: `admin123`

**⚠️ 重要：首次登录后请立即修改密码！**

### 方法二：手动创建管理员账户

如果初始化脚本执行失败，可以手动创建管理员账户：

#### 1. 使用Go程序生成密码hash

创建一个临时Go文件 `gen-password.go`：

```go
package main

import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    password := "admin123" // 修改为你想要的密码
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        panic(err)
    }
    fmt.Println(string(hash))
}
```

运行生成hash：

```bash
go run gen-password.go
```

#### 2. 执行SQL插入用户

```sql
-- 确保角色表有数据
INSERT IGNORE INTO `user_role` (`id`, `role_name`, `role_code`) VALUES
(1, '管理员', 0),
(2, '普通用户', 1);

-- 插入管理员账户（将下面的hash替换为你生成的hash）
INSERT INTO `users` (`username`, `password`, `role_id`) VALUES
('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 1);
```

#### 3. 验证账户创建

```sql
SELECT id, username, role_id FROM users WHERE username = 'admin';
```

### 方法三：通过系统界面创建（需要已有管理员账户）

1. 使用已有管理员账户登录系统
2. 进入"系统设置" -> "用户信息"
3. 点击"添加用户"
4. 填写用户名、密码，选择角色为"管理员"
5. 保存

## 数据库表结构

系统包含以下主要数据表：

1. **user_role** - 用户角色表
2. **users** - 用户表
3. **system_settings** - 系统设置表
4. **operation_logs** - 操作日志表
5. **fileList** - 设备建档明细表
6. **checkpoint_details** - 卡口建档明细表
7. **audit_tasks** - 设备审核任务表
8. **audit_details** - 设备审核明细表
9. **checkpoint_tasks** - 卡口审核任务表

详细的表结构定义请参考 `deploy/create-all-tables.sql` 文件。

## 编译说明

### Windows 7+ 编译

```bash
go build -ldflags="-s -w" -o deploy/ops-web-pro.exe main.go
```

编译参数说明：
- `-ldflags="-s -w"`: 去除符号表和调试信息，减小文件大小
- `-o deploy/ops-web-pro.exe`: 输出文件路径

### Linux 编译

```bash
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ops-web main.go
```

### 交叉编译

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ops-web.exe main.go

# Linux
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ops-web main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o ops-web main.go
```

## 部署说明

### Windows 部署

1. 将以下文件/目录复制到目标机器：
   - `ops-web-pro.exe`（主程序）
   - `config/` 目录（配置文件）
   - `templates/` 目录（HTML模板）

2. 编辑 `config/config.json`，设置数据库连接信息

3. 初始化数据库（参考"初始化数据库"章节）

4. 运行程序：
   ```cmd
   ops-web-pro.exe
   ```

5. 访问系统：http://127.0.0.1:8080/login

详细说明请参考 `deploy/WIN7-DEPLOY.md`

### Linux 部署

详细说明请参考 `deploy/DEPLOY.md`

## 重要说明

### DDL依赖已关闭

**注意**：从当前版本开始，程序启动时**不会自动创建数据库表**。必须手动执行SQL脚本创建表结构。

原因：
- 避免在生产环境中意外执行DDL操作
- 提供更好的数据库结构控制
- 支持数据库迁移和版本管理

### 初始化步骤

1. **创建数据库**
   ```sql
   CREATE DATABASE IF NOT EXISTS ops DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
   ```

2. **执行建表SQL**
   ```bash
   mysql -u root -p ops < deploy/create-all-tables.sql
   ```

3. **创建管理员账户**
   ```bash
   mysql -u root -p ops < deploy/init-admin-user.sql
   ```

4. **启动程序**
   ```bash
   ./ops-web-pro.exe
   ```

## 配置文件说明

### config/config.json

```json
{
  "db_host": "127.0.0.1",        // 数据库主机地址
  "db_port": "3306",              // 数据库端口
  "db_user": "root",              // 数据库用户名
  "db_pass": "your_password",     // 数据库密码
  "db_name": "ops",               // 数据库名称
  "server_host": "127.0.0.1",     // 服务器监听地址
  "server_port": "8080"           // 服务器监听端口
}
```

## 功能模块

### 1. 用户认证与权限管理
- 用户登录/登出
- 基于角色的权限控制（管理员/普通用户）
- 会话管理

### 2. 设备建档管理
- 设备档案导入（Excel）
- 设备档案查询和筛选
- 设备档案导出（Excel）
- 建档状态管理

### 3. 卡口建档管理
- 卡口档案导入（Excel）
- 卡口档案查询和筛选
- 卡口档案导出（Excel）
- 建档状态管理

### 4. 设备审核管理
- 审核任务创建
- 审核进度跟踪
- 审核意见管理
- 文件上传/下载
- 单兵设备标识
- 档案类型管理（新增/取推/补档案）

### 5. 卡口审核管理
- 审核任务创建
- 审核进度跟踪
- 审核意见管理
- 文件上传/下载
- 档案类型管理（新增/取推/补档案）

### 6. 月度建档数据统计
- 按分局统计
- 按点位类型统计（一类点/二类点/三类点/内部监控）
- 按功能类型统计（视频/人脸/车辆）
- 单兵设备统计
- 数据导出（Excel）

### 7. 系统设置
- 用户管理（管理员）
- 操作日志查看（管理员）
- 任务配置（管理员）
  - 文件上传路径配置
  - 备份路径配置
  - 数据库备份路径配置
  - 定时备份任务配置

### 8. 备份功能
- 数据库备份（手动/定时）
- 文件备份（手动/定时）
- 自动清理旧备份（保留最近5天）

## 开发指南

### 项目结构

```
ops-web/
├── main.go                    # 程序入口
├── config/                     # 配置文件目录
│   └── config.json            # 数据库配置
├── templates/                 # HTML模板目录
├── logs/                      # 日志文件目录（自动创建）
├── deploy/                    # 部署相关文件
│   ├── ops-web-pro.exe        # Windows可执行文件
│   ├── create-all-tables.sql  # 数据库表结构SQL
│   ├── init-admin-user.sql    # 初始化管理员账户SQL
│   ├── init-database.bat      # Windows初始化脚本
│   └── init-database.sh       # Linux初始化脚本
└── internal/                  # 内部包
    ├── auth/                  # 认证模块
    ├── db/                    # 数据库模块
    ├── user/                  # 用户管理
    ├── filelist/              # 设备建档明细
    ├── checkpointfilelist/    # 卡口建档明细
    ├── auditprogress/         # 设备审核进度
    ├── checkpointprogress/    # 卡口审核进度
    ├── auditstatistics/       # 月度建档数据统计
    ├── taskconfig/            # 任务配置
    ├── operationlog/          # 操作日志
    └── logger/                # 日志模块
```

### 添加新功能

1. 在 `internal/` 目录下创建新的包
2. 实现Handler函数
3. 在 `main.go` 中注册路由
4. 创建对应的HTML模板（如需要）

### 数据库操作

- 使用 `db.DBInstance` 进行数据库操作
- 使用事务处理多表操作
- 使用参数化查询防止SQL注入

### 日志记录

- 使用 `logger.Errorf()` 记录错误日志
- 使用 `operationlog.Record()` 记录操作日志

## 常见问题

### 1. 数据库连接失败

- 检查MySQL服务是否运行
- 检查 `config/config.json` 中的连接信息
- 检查数据库用户权限
- 检查防火墙设置

### 2. 表不存在错误

- 确保已执行 `deploy/create-all-tables.sql`
- 检查数据库名称是否正确
- 检查表名是否正确（注意大小写）

### 3. 无法登录

- 检查管理员账户是否已创建
- 检查密码是否正确
- 查看日志文件获取详细错误信息

### 4. 文件上传失败

- 检查"任务配置"中的"文件上传路径"是否设置
- 检查路径是否存在且有写权限
- 检查磁盘空间是否充足

### 5. 备份功能失败

- 检查备份路径是否设置
- 检查路径是否存在且有写权限
- 检查 `mysqldump` 命令是否可用（数据库备份）
- 查看日志文件获取详细错误信息

## 技术支持

如遇到问题，请：

1. 查看日志文件：`logs/ops-web-YYYY-MM-DD.log`
2. 查看系统操作日志：系统内"操作日志"页面
3. 检查数据库连接和表结构
4. 检查配置文件是否正确

## 版本历史

- **v1.0** - 初始版本
  - 基础功能实现
  - 设备/卡口建档管理
  - 审核进度管理
  - 月度统计功能
  - 备份功能
  - 定时任务功能

## 许可证

[根据实际情况填写]
