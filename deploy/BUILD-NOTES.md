# 编译说明

## 编译信息

- **可执行文件**: `ops-web-pro.exe`
- **编译时间**: 2025-12-23
- **目标平台**: Windows 7 及以上版本（64位）
- **架构**: amd64
- **CGO**: 已禁用（CGO_ENABLED=0，无需C运行时库）
- **文件大小**: 约 10.19 MB

## 编译命令

```powershell
$env:CGO_ENABLED=0
$env:GOOS='windows'
$env:GOARCH='amd64'
go build -ldflags='-s -w' -o deploy\ops-web-pro.exe .
```

## 编译参数说明

- `CGO_ENABLED=0`: 禁用CGO，生成纯Go程序，无需C运行时库，兼容性更好
- `GOOS=windows`: 目标操作系统为Windows
- `GOARCH=amd64`: 目标架构为64位
- `-ldflags='-s -w'`: 
  - `-s`: 去除符号表
  - `-w`: 去除调试信息
  - 减小文件大小，提高运行性能

## DDL依赖

✅ **已去除DDL依赖**

程序已去掉数据库表的自动创建功能，需要手动执行SQL脚本创建数据库表。相关代码已在 `main.go` 中注释：

```go
// 注意：DDL依赖已关闭，请手动执行SQL脚本创建数据库表
// 1.1. 初始化审核相关的数据库表（已禁用，请手动执行SQL）
// if err := db.InitAuditTables(); err != nil {
//     ...
// }

// 1.2. 初始化卡口审核相关的数据库表（已禁用，请手动执行SQL）
// if err := db.InitCheckpointTables(); err != nil {
//     ...
// }
```

## Windows 7 兼容性

✅ **兼容Windows 7及以上版本**

- 使用 `CGO_ENABLED=0` 编译，生成静态链接的可执行文件
- 无需额外的运行时库（如Visual C++ Redistributable）
- 目标架构为amd64，支持64位Windows系统

## 运行要求

1. **操作系统**: Windows 7 及以上版本（64位）
2. **数据库**: MySQL（需要手动创建表结构）
3. **配置文件**: `config/config.json`（数据库连接配置）

## 部署步骤

1. 将 `ops-web-pro.exe` 复制到目标服务器
2. 确保 `config/config.json` 配置文件存在并正确配置
3. 执行数据库初始化SQL脚本（参考 `deploy/` 目录下的SQL文件）
4. 运行 `ops-web-pro.exe`

## 注意事项

- 程序依赖MySQL数据库，确保数据库服务已启动
- 首次运行前需要执行数据库初始化SQL脚本
- 配置文件路径为相对路径 `config/config.json`，确保可执行文件与config目录在同一目录下






