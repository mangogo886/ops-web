# Windows 7 部署说明

## 文件说明

`ops-web-pro.exe` 是专门为 Windows 7 编译的可执行文件，已优化并静态链接，无需额外依赖。

## 部署步骤

### 1. 准备文件

将以下文件/目录复制到目标 Windows 7 机器：

```
ops-web-pro.exe          # 主程序
config/                  # 配置目录
  └── config.json        # 数据库配置文件（需要修改）
templates/               # HTML模板目录（所有模板文件）
```

### 2. 配置数据库

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

### 3. 初始化数据库

在 MySQL 中执行以下 SQL 脚本（按顺序）：

1. `create-user-tables.sql` - 创建用户和角色表（必须）
2. `create-table.sql` - 创建基础数据表（如需要）
3. `create-audit-tables.sql` - 创建审核相关表（如需要）

### 4. 运行程序

双击 `ops-web-pro.exe` 或在命令行中运行：

```cmd
ops-web-pro.exe
```

### 5. 访问系统

打开浏览器访问：http://127.0.0.1:8080/login

默认管理员账号：
- 用户名：`admin`
- 密码：`admin123`

**⚠️ 首次登录后请立即修改密码！**

## 系统要求

- Windows 7 或更高版本
- MySQL 5.7+ 或 MariaDB 10.3+
- 网络连接（用于访问数据库）

## 注意事项

1. **防火墙**：如果无法访问，请检查 Windows 防火墙是否允许 8080 端口
2. **数据库**：确保 MySQL 服务正在运行
3. **日志**：程序会在运行目录自动创建 `logs` 目录，日志文件为 `ops-web-YYYY-MM-DD.log`
4. **配置文件**：确保 `config/config.json` 文件存在且配置正确

## 常见问题

### 1. 程序无法启动

- 检查是否有杀毒软件拦截
- 检查是否有其他程序占用 8080 端口
- 查看日志文件 `logs/ops-web-*.log` 获取错误信息

### 2. 无法连接数据库

- 检查 MySQL 服务是否运行
- 检查 `config/config.json` 中的数据库连接信息
- 检查数据库用户权限

### 3. 页面无法访问

- 检查程序是否正常运行
- 检查防火墙设置
- 尝试使用 `http://127.0.0.1:8080` 而不是 `localhost`

## 更新程序

1. 停止当前运行的程序（关闭窗口或按 Ctrl+C）
2. 备份当前版本（可选）
3. 替换 `ops-web-pro.exe` 为新版本
4. 重新运行程序

## 技术支持

如遇到问题，请查看：
1. 日志文件：`logs/ops-web-YYYY-MM-DD.log`
2. 系统操作日志：系统内"操作日志"页面


















