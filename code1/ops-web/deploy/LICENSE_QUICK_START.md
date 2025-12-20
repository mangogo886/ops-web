# 许可授权快速开始指南

## 一、首次部署

### 步骤1：获取服务器MAC地址

在Windows服务器上打开命令提示符或PowerShell，执行：

```cmd
ipconfig /all
```

找到"物理地址"（Physical Address），例如：`00-11-22-33-44-55`

### 步骤2：生成许可文件

使用许可生成工具（需要先编译）：

```bash
# 编译工具（在项目根目录执行）
go build -o license-generator.exe tools/license-generator/main.go

# 生成许可文件（示例：授权到2026-12-30）
license-generator.exe -expire 2026-12-30 -mac 00:11:22:33:44:55 -output config/license.json
```

### 步骤3：部署许可文件

将生成的 `config/license.json` 文件放到服务器 `config` 目录下，与 `config.json` 同一目录。

### 步骤4：启动系统

正常启动系统，用户登录时会自动验证许可。

---

## 二、许可文件示例

`config/license.json` 文件内容示例：

```json
{
    "expireDate": "2026-12-30",
    "macAddress": "00:11:22:33:44:55",
    "issueDate": "2024-12-20",
    "licenseKey": ""
}
```

---

## 三、常见操作

### 查看当前服务器MAC地址

**方法1：命令提示符**
```cmd
ipconfig /all
```

**方法2：PowerShell**
```powershell
Get-NetAdapter | Where-Object {$_.Status -eq "Up"} | Select-Object Name, MacAddress
```

### 生成新的许可文件

```bash
license-generator.exe -expire 2027-12-31 -mac 00:11:22:33:44:55 -output config/license.json
```

### 更新许可（延长授权期限）

1. 停止系统
2. 生成新的许可文件
3. 替换 `config/license.json` 文件
4. 重启系统

---

## 四、故障排查

### 登录失败，提示"系统错误，请联系管理员"

**检查清单**：

1. ✓ 许可文件是否存在：`config/license.json`
2. ✓ 许可文件格式是否正确（JSON格式）
3. ✓ 授权日期是否过期（`expireDate` 必须在当前日期之后）
4. ✓ MAC地址是否匹配（服务器MAC地址必须与许可文件中的 `macAddress` 一致）

**查看日志**：

检查 `logs/ops-web-YYYY-MM-DD.log` 文件，查找"许可验证"相关错误信息。

---

## 五、许可生成工具参数说明

```
用法: license-generator.exe [选项]

选项:
  -expire string
        到期日期（必填，格式：YYYY-MM-DD）
  -mac string
        MAC地址（必填，格式：XX:XX:XX:XX:XX:XX）
  -output string
        输出文件路径（可选，默认：license.json）
  -issue string
        签发日期（可选，默认：当前日期，格式：YYYY-MM-DD）

示例:
  license-generator.exe -expire 2026-12-30 -mac 00:11:22:33:44:55
  license-generator.exe -expire 2026-12-30 -mac 00:11:22:33:44:55 -output config/license.json
```

---

## 六、注意事项

1. ⚠️ **许可文件必须与服务器MAC地址匹配**，否则无法使用
2. ⚠️ **许可文件不可复制到其他服务器**使用（MAC地址绑定）
3. ⚠️ **授权日期不能早于当前日期**，否则系统会拒绝登录
4. ⚠️ **更新许可时建议先停止系统**，避免文件被占用
5. ⚠️ **服务器迁移时**需要为新服务器生成新的许可文件

---

**更多详细信息请参考**：`LICENSE_README.md`

