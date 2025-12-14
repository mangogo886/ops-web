# 部署说明

本目录包含生产环境部署所需的所有文件。

## 文件说明

### SQL脚本
- `create-user-tables.sql` - 创建用户和角色表（**必须执行**）
- `create-table.sql` - 创建基础数据表（如使用建档明细功能）
- `create-audit-tables.sql` - 创建审核相关表（如使用审核功能）
- `alter-audit-details-add-status.sql` - 为审核明细表添加审核状态字段（如表已存在）

### 配置文件
- `config/config.json` - 数据库配置文件示例

### 模板文件
- `templates/` - HTML模板文件（可选，如果模板有更新）

## 部署步骤

### 1. 数据库初始化

按顺序执行SQL脚本：

```bash
# 1. 创建用户和角色表（必须）
mysql -u root -p ops < create-user-tables.sql

# 2. 创建基础数据表（如需要）
mysql -u root -p ops < create-table.sql

# 3. 创建审核相关表（如需要）
mysql -u root -p ops < create-audit-tables.sql

# 4. 添加审核状态字段（如表已存在）
mysql -u root -p ops < alter-audit-details-add-status.sql
```

### 2. 配置文件

复制配置文件到项目根目录：

```bash
cp config/config.json ../../config/config.json
```

然后编辑 `config/config.json`，设置正确的数据库连接信息。

### 3. 编译程序

在项目根目录执行：

```bash
# Windows
go build -o ops-web.exe .

# Linux
go build -o ops-web .
```

### 4. 部署文件

将以下文件/目录复制到生产服务器：

```
ops-web              # 可执行文件
config/              # 配置目录
templates/           # 模板目录
```

### 5. 启动服务

参考主README文档中的"生产环境部署"部分。

## 初始化管理员用户

执行 `create-user-tables.sql` 后，系统会自动创建默认管理员：

- 用户名：`admin`
- 密码：`admin123`
- 角色：管理员

**⚠️ 首次登录后请立即修改密码！**

如需自定义密码，使用以下命令生成密码Hash：

```bash
go run ../../scripts/init-admin.go <your_password>
```

然后将输出的SQL语句在数据库中执行。

## 注意事项

1. 确保数据库用户有足够的权限创建表和插入数据
2. 生产环境请使用专用数据库用户，不要使用root
3. 配置文件包含敏感信息，注意文件权限（建议600）
4. 定期备份数据库



