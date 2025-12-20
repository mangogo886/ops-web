# 系统许可授权使用说明

## ⚠️ 重要提示

**许可文件已加密**，无法手动修改。必须使用 `license-generator.exe` 工具生成许可文件。

详细说明请参考：[许可文件加密说明](LICENSE_ENCRYPTION.md)

## 目录
1. [功能概述](#功能概述)
2. [许可文件说明](#许可文件说明)
3. [许可生成工具使用](#许可生成工具使用)
4. [部署说明](#部署说明)
5. [常见问题](#常见问题)

---

## 功能概述

系统许可授权功能用于：
- **授权期限管理**：限制系统使用期限（如授权到2026-12-30）
- **MAC地址绑定**：许可文件与服务器MAC地址绑定，防止复制到其他服务器
- **登录验证**：用户登录时自动检查许可有效性，如许可无效则拒绝登录

### 验证流程

```
用户登录
  ↓
验证用户名和密码
  ↓
检查许可文件是否存在
  ↓
检查授权日期是否过期
  ↓
获取服务器MAC地址
  ↓
检查MAC地址是否匹配
  ↓
✓ 通过 → 允许登录
✗ 失败 → 提示"系统错误，请联系管理员"
```

---

## 许可文件说明

### 文件位置

许可文件必须放在以下位置：
```
config/license.json
```

**注意**：
- 许可文件必须与 `config/config.json` 在同一目录
- 如果许可文件不存在或格式错误，系统将拒绝所有登录请求

### 文件格式

**注意**：许可文件采用 **AES-256-CBC 加密**存储，文件内容是加密后的Base64编码数据，**无法直接查看或手动编辑**。

文件内容是加密的，类似：
```
base64编码的加密数据（不可读）
```

内部包含的JSON数据结构：
```json
{
    "expireDate": "2026-12-30",
    "macAddress": "00:11:22:33:44:55",
    "issueDate": "2024-01-01",
    "licenseKey": ""
}
```

⚠️ **重要**：不能手动创建或编辑许可文件，必须使用 `license-generator.exe` 工具生成。

### 字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `expireDate` | string | 是 | 授权到期日期，格式：YYYY-MM-DD |
| `macAddress` | string | 是 | 绑定的服务器MAC地址，格式：XX:XX:XX:XX:XX:XX（大写，冒号分隔） |
| `issueDate` | string | 是 | 许可签发日期，格式：YYYY-MM-DD |
| `licenseKey` | string | 否 | 许可密钥（当前版本未使用，可为空） |

### MAC地址格式

- **标准格式**：`XX:XX:XX:XX:XX:XX`（大写字母，冒号分隔）
- **支持格式**：
  - `00:11:22:33:44:55` ✓
  - `00-11-22-33-44-55` ✓（会自动转换为冒号格式）
  - `001122334455` ✗（不支持）
  - `00:11:22:33:44:5` ✗（格式不正确）

---

## 许可生成工具使用

### 编译工具

在项目根目录执行：

```bash
go build -o license-generator.exe tools/license-generator/main.go
```

或者直接在Windows上：

```powershell
go build -o license-generator.exe tools/license-generator/main.go
```

### 使用方法

#### 基本用法

```bash
license-generator.exe -expire 2026-12-30 -mac 00:11:22:33:44:55
```

#### 指定输出文件

```bash
license-generator.exe -expire 2026-12-30 -mac 00:11:22:33:44:55 -output config/license.json
```

#### 指定签发日期

```bash
license-generator.exe -expire 2026-12-30 -mac 00:11:22:33:44:55 -issue 2024-01-01
```

### 命令行参数

| 参数 | 说明 | 必填 | 示例 |
|------|------|------|------|
| `-expire` | 到期日期（格式：YYYY-MM-DD） | 是 | `-expire 2026-12-30` |
| `-mac` | MAC地址（格式：XX:XX:XX:XX:XX:XX） | 是 | `-mac 00:11:22:33:44:55` |
| `-output` | 输出文件路径 | 否 | `-output config/license.json`（默认：license.json） |
| `-issue` | 签发日期（格式：YYYY-MM-DD） | 否 | `-issue 2024-01-01`（默认：当前日期） |

### 使用示例

#### 示例1：生成1年授权许可

```bash
# 假设服务器MAC地址为：00:11:22:33:44:55
# 授权到2025-12-31

license-generator.exe -expire 2025-12-31 -mac 00:11:22:33:44:55 -output config/license.json
```

#### 示例2：生成3年授权许可

```bash
license-generator.exe -expire 2027-12-31 -mac 00:11:22:33:44:55 -output config/license.json
```

#### 示例3：查看帮助

```bash
license-generator.exe -h
```

---

## 部署说明

### 首次部署步骤

1. **获取服务器MAC地址**

   在Windows服务器上执行以下命令查看MAC地址：

   ```cmd
   ipconfig /all
   ```

   找到"物理地址"（Physical Address），格式类似：`00-11-22-33-44-55`

   或者使用PowerShell：

   ```powershell
   Get-NetAdapter | Where-Object {$_.Status -eq "Up"} | Select-Object Name, MacAddress
   ```

2. **生成许可文件**

   ```bash
   license-generator.exe -expire 2026-12-30 -mac 00:11:22:33:44:55 -output config/license.json
   ```

   将生成的 `license.json` 文件放到服务器的 `config` 目录下。

3. **验证许可文件**

   确认文件内容正确：

   ```json
   {
       "expireDate": "2026-12-30",
       "macAddress": "00:11:22:33:44:55",
       "issueDate": "2024-12-20",
       "licenseKey": ""
   }
   ```

4. **启动系统**

   正常启动系统，用户登录时会自动验证许可。

### 许可更新步骤

1. **停止系统服务**（如果正在运行）

2. **生成新的许可文件**

   ```bash
   license-generator.exe -expire 2027-12-31 -mac 00:11:22:33:44:55 -output config/license.json
   ```

3. **替换许可文件**

   将新的 `config/license.json` 文件替换旧文件。

4. **重启系统**

   重启系统后，新许可生效。

### 服务器迁移步骤

如果服务器需要迁移到新机器：

1. **获取新服务器MAC地址**

2. **生成新的许可文件**（使用新MAC地址）

   ```bash
   license-generator.exe -expire 2026-12-30 -mac AA:BB:CC:DD:EE:FF -output config/license.json
   ```

3. **将新许可文件放到新服务器**

4. **迁移系统文件和数据**

5. **在新服务器上启动系统**

---

## 常见问题

### Q1: 登录时提示"系统错误，请联系管理员"，是什么原因？

可能的原因：
1. **许可文件不存在**：检查 `config/license.json` 文件是否存在
2. **许可已过期**：检查 `expireDate` 是否在当前日期之后
3. **MAC地址不匹配**：检查服务器MAC地址是否与许可文件中的MAC地址一致
4. **许可文件格式错误**：检查JSON格式是否正确

**排查步骤**：
1. 查看系统日志文件 `logs/ops-web-YYYY-MM-DD.log`，查找"许可验证"相关错误
2. 确认许可文件位置和内容正确
3. 确认服务器MAC地址与许可文件中的MAC地址一致

### Q2: 如何查看服务器MAC地址？

**Windows系统**：
```cmd
ipconfig /all
```

在输出中找到"物理地址"（Physical Address）。

**PowerShell**：
```powershell
Get-NetAdapter | Where-Object {$_.Status -eq "Up"} | Select-Object Name, MacAddress
```

### Q3: 服务器有多个网卡，应该使用哪个MAC地址？

系统会自动选择第一个**非环回、已启用**的网络接口的MAC地址。

如果服务器有多个网卡，建议：
1. 先启动系统，查看日志中的MAC地址（如果有错误日志）
2. 或者查看代码逻辑：系统会选择第一个有效的非环回接口
3. 生成许可时使用该MAC地址

### Q4: 许可文件可以复制到其他服务器使用吗？

**不可以**。许可文件与服务器MAC地址绑定，只能在MAC地址匹配的服务器上使用。如果将许可文件复制到其他服务器，MAC地址不匹配会导致验证失败。

### Q5: 许可文件格式错误怎么办？

⚠️ **许可文件已加密**，不能手动编辑。

如果出现格式错误：
1. **删除损坏的许可文件**
2. **使用工具重新生成**：
   ```bash
   license-generator.exe -expire 2026-12-30 -mac 00:11:22:33:44:55 -output config/license.json
   ```

**注意**：
- 许可文件是加密的，无法直接查看内容
- 不能手动创建或编辑许可文件
- 必须使用 `license-generator.exe` 工具生成

### Q6: 如何延长授权期限？

1. 停止系统
2. 使用许可生成工具生成新的许可文件（新的到期日期）
3. 替换 `config/license.json` 文件
4. 重启系统

**注意**：MAC地址必须保持不变（同一台服务器）。

### Q7: 可以在不重启系统的情况下更新许可吗？

理论上可以，但建议重启系统以确保许可验证生效。如果系统正在运行：
1. 替换许可文件
2. 用户下次登录时会使用新的许可文件进行验证

### Q8: 系统日志在哪里？

系统日志位于：`logs/ops-web-YYYY-MM-DD.log`

许可验证相关的日志会包含"许可验证"关键字，例如：
```
许可验证-读取许可文件失败: ...
许可验证-授权已过期: 到期日期=2026-12-30
许可验证-MAC地址不匹配: 许可MAC=..., 服务器MAC=...
```

---

## 技术说明

### MAC地址获取逻辑

系统获取MAC地址的流程：
1. 获取所有网络接口
2. 过滤掉环回接口（loopback）
3. 选择第一个已启用（Up）的非环回接口
4. 获取该接口的MAC地址
5. 格式化为标准格式（XX:XX:XX:XX:XX:XX）

### 日期比较逻辑

系统使用日期比较（不考虑时间）：
- 当前日期：2024-12-20
- 到期日期：2026-12-30
- 结果：未过期 ✓

如果当前日期大于或等于到期日期，则视为过期。

### 安全性说明

1. **MAC地址绑定**：防止许可文件被复制到其他服务器使用
2. **日期验证**：限制系统使用期限
3. **错误提示**：统一的错误提示"系统错误，请联系管理员"，不泄露具体原因，增强安全性

---

## 联系支持

如果遇到许可相关问题，请联系系统管理员，并提供：
1. 系统日志文件（`logs/ops-web-YYYY-MM-DD.log`）
2. 服务器MAC地址
3. 许可文件内容（隐藏敏感信息）

---

**文档版本**：1.0  
**最后更新**：2024-12-20

