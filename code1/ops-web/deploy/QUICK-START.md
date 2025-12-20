# 快速部署指南

## Windows 7+ 快速部署

### 1. 准备文件

将以下文件/目录复制到目标机器：

```
ops-web-pro.exe          # 主程序（已编译，无需Go环境）
config/                  # 配置目录
  └── config.json        # 数据库配置文件（需要修改）
templates/               # HTML模板目录（所有模板文件）
deploy/                  # 部署脚本目录
  ├── create-all-tables.sql    # 数据库表结构SQL
  ├── init-admin-user.sql      # 初始化管理员账户SQL
  ├── init-database.bat        # Windows初始化脚本
  └── QUICK-START.md           # 本文件
```

### 2. 配置数据库

编辑 `config/config.json`，设置数据库连接信息：

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

### 3. 初始化数据库

#### 方法一：使用初始化脚本（推荐）

```cmd
cd deploy
init-database.bat ops root your_password
```

#### 方法二：手动执行SQL

```cmd
# 1. 创建数据库
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS ops DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 2. 创建所有表
mysql -u root -p ops < deploy\create-all-tables.sql

# 3. 创建管理员账户
mysql -u root -p ops < deploy\init-admin-user.sql
```

### 4. 运行程序

```cmd
ops-web-pro.exe
```

### 5. 访问系统

打开浏览器访问：http://127.0.0.1:8080/login

**默认管理员账户：**
- 用户名：`admin`
- 密码：`admin123`

**⚠️ 首次登录后请立即修改密码！**

## 创建管理员账户

### 方法一：使用初始化脚本（推荐）

执行 `deploy/init-admin-user.sql` 会自动创建默认管理员账户。

### 方法二：手动创建

#### 1. 生成密码hash

创建临时文件 `gen-password.go`：

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

运行：

```bash
go run gen-password.go
```

#### 2. 执行SQL

```sql
-- 确保角色表有数据
INSERT IGNORE INTO `user_role` (`id`, `role_name`, `role_code`) VALUES
(1, '管理员', 0),
(2, '普通用户', 1);

-- 插入管理员账户（将hash替换为你生成的hash）
INSERT INTO `users` (`username`, `password`, `role_id`) VALUES
('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 1);
```

## 重要说明

### DDL依赖已关闭

**程序启动时不会自动创建数据库表**，必须手动执行SQL脚本。

### 初始化步骤（必须按顺序执行）

1. **创建数据库**
2. **执行 `create-all-tables.sql` 创建所有表**
3. **执行 `init-admin-user.sql` 创建管理员账户**
4. **启动程序**

## 常见问题

### 1. 数据库连接失败

- 检查MySQL服务是否运行
- 检查 `config/config.json` 中的连接信息
- 检查数据库用户权限

### 2. 表不存在错误

- 确保已执行 `create-all-tables.sql`
- 检查数据库名称是否正确

### 3. 无法登录

- 检查管理员账户是否已创建
- 默认账户：admin / admin123
- 查看日志文件获取详细错误信息

### 4. 程序无法启动

- 检查是否有杀毒软件拦截
- 检查是否有其他程序占用8080端口
- 查看日志文件 `logs/ops-web-*.log`

## 日志文件

程序运行日志保存在 `logs/ops-web-YYYY-MM-DD.log` 文件中。

## 技术支持

详细文档请参考 `README-AUTH.md`


