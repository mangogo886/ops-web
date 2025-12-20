# 系统许可授权设计方案

## 1. 概述

本文档描述系统许可授权功能的设计方案，包括：
- 授权日期检查
- MAC地址绑定
- 登录时的授权验证
- 许可文件生成工具

## 2. 系统架构设计

### 2.1 许可文件结构

**许可文件位置**：`config/license.json`（位于项目根目录的config目录下）

**许可文件格式（JSON）**：
```json
{
  "expireDate": "2026-12-30",
  "macAddress": "00:11:22:33:44:55",
  "issueDate": "2024-01-01",
  "licenseKey": "encrypted_signature_string"
}
```

**字段说明**：
- `expireDate`: 授权到期日期（格式：YYYY-MM-DD）
- `macAddress`: 绑定的服务器MAC地址（格式：XX:XX:XX:XX:XX:XX）
- `issueDate`: 许可签发日期（格式：YYYY-MM-DD）
- `licenseKey`: 许可密钥（可选，用于验证许可文件的完整性和真实性）

### 2.2 文件结构设计

```
ops-web/
├── config/
│   ├── config.json          # 原有配置文件
│   └── license.json         # 许可文件（新增）
├── internal/
│   └── license/             # 许可模块（新增目录）
│       ├── license.go       # 许可验证逻辑
│       ├── mac.go           # MAC地址获取
│       └── generator.go     # 许可生成工具（可选，用于生成许可）
├── tools/                   # 工具目录（新增）
│   └── license-generator/   # 许可生成工具
│       └── main.go          # 独立的许可生成程序
└── ...
```

## 3. 核心模块设计

### 3.1 许可验证模块 (`internal/license/license.go`)

**主要功能**：
1. 读取许可文件
2. 验证授权日期
3. 验证MAC地址
4. 验证许可完整性（如果使用licenseKey）

**核心接口**：
```go
package license

// LicenseInfo 许可信息结构
type LicenseInfo struct {
    ExpireDate  string `json:"expireDate"`
    MacAddress  string `json:"macAddress"`
    IssueDate   string `json:"issueDate"`
    LicenseKey  string `json:"licenseKey,omitempty"`
}

// ValidateLicense 验证许可是否有效
// 返回 (isValid, errorMessage)
func ValidateLicense() (bool, string)

// GetLicenseInfo 获取许可信息（用于显示）
func GetLicenseInfo() (*LicenseInfo, error)

// IsLicenseExpired 检查许可是否过期
func IsLicenseExpired() bool

// IsMacAddressMatch 检查MAC地址是否匹配
func IsMacAddressMatch() bool
```

**验证逻辑流程**：
```
1. 读取 config/license.json 文件
2. 解析JSON，获取许可信息
3. 检查授权日期是否过期
   - 当前日期 > expireDate → 返回错误："授权已过期"
4. 获取服务器MAC地址
5. 比较MAC地址是否匹配
   - MAC地址不匹配 → 返回错误："授权MAC地址不匹配"
6. （可选）验证licenseKey的签名
7. 返回验证结果
```

### 3.2 MAC地址获取模块 (`internal/license/mac.go`)

**主要功能**：
- 获取服务器第一个非环回网络接口的MAC地址

**核心接口**：
```go
package license

// GetServerMacAddress 获取服务器MAC地址
// 返回格式：XX:XX:XX:XX:XX:XX
func GetServerMacAddress() (string, error)

// NormalizeMacAddress 标准化MAC地址格式
// 统一转换为大写，使用冒号分隔
func NormalizeMacAddress(mac string) string
```

**实现思路**：
1. 使用 `net.Interfaces()` 获取所有网络接口
2. 过滤掉环回接口（loopback）
3. 选择第一个有效的非环回接口
4. 获取其MAC地址
5. 格式化为标准格式（XX:XX:XX:XX:XX:XX）

**注意事项**：
- Windows/Linux/macOS 通用实现
- 处理多网卡情况（选择第一个有效网卡）
- MAC地址格式统一为大写，冒号分隔

### 3.3 许可生成工具 (`tools/license-generator/main.go`)

**功能**：
- 独立命令行工具，用于生成许可文件

**使用方式**：
```bash
# 编译
go build -o license-generator tools/license-generator/main.go

# 使用（指定到期日期和MAC地址）
./license-generator -expire 2026-12-30 -mac 00:11:22:33:44:55

# 输出许可文件到指定路径
./license-generator -expire 2026-12-30 -mac 00:11:22:33:44:55 -output config/license.json
```

**命令行参数**：
- `-expire`: 到期日期（必填，格式：YYYY-MM-DD）
- `-mac`: MAC地址（必填，格式：XX:XX:XX:XX:XX:XX）
- `-output`: 输出文件路径（可选，默认：license.json）
- `-issue`: 签发日期（可选，默认：当前日期）

**生成逻辑**：
```go
1. 解析命令行参数
2. 验证日期格式
3. 验证MAC地址格式
4. 生成许可信息结构
5. 序列化为JSON
6. 写入到指定文件
```

## 4. 登录流程集成

### 4.1 修改登录处理逻辑

**位置**：`internal/auth/auth.go` 的 `LoginHandler` 函数

**修改点**：在验证用户名和密码之后，创建会话之前，添加许可验证

**伪代码**：
```go
func LoginHandler(w http.ResponseWriter, r *http.Request) {
    // ... 现有代码：验证用户名和密码 ...
    
    // 验证密码成功后，检查许可
    isValid, errMsg := license.ValidateLicense()
    if !isValid {
        logger.Errorf("登录-许可验证失败: %s, 用户名: %s", errMsg, username)
        renderLoginPage(w, "系统错误，请联系管理员")
        return
    }
    
    // ... 现有代码：创建会话 ...
}
```

### 4.2 修改中间件（可选，用于已登录用户的检查）

**位置**：`internal/auth/middleware.go` 的 `RequireAuth` 函数

**考虑**：可以在中间件中也检查许可，防止已登录用户在许可过期后继续使用

**伪代码**：
```go
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if !IsAuthenticated(r) {
            http.Redirect(w, r, "/login", http.StatusFound)
            return
        }
        
        // 检查许可（可选，如果不需要实时检查，可以只在登录时检查）
        isValid, errMsg := license.ValidateLicense()
        if !isValid {
            // 清除session，重定向到登录页
            // ... 清除cookie逻辑 ...
            http.Redirect(w, r, "/login?error="+url.QueryEscape(errMsg), http.StatusFound)
            return
        }
        
        next(w, r)
    }
}
```

**建议**：为了性能考虑，建议只在登录时检查许可，不 dalam中间件中检查（除非需要实时验证）。

## 5. 错误处理设计

### 5.1 许可文件不存在
- 错误消息："许可文件不存在，请联系管理员"
- 日志记录：记录详细的错误信息

### 5.2 许可文件格式错误
- 错误消息："许可文件格式错误，请联系管理员"
- 日志记录：记录解析错误详情

### 5.3 授权已过期
- 错误消息："系统授权已过期，请联系管理员"
- 日志记录：记录到期日期和当前日期

### 5.4 MAC地址不匹配
- 错误消息："系统授权MAC地址不匹配，请联系管理员"
- 日志记录：记录许可MAC地址和服务器MAC地址

### 5.5 无法获取MAC地址
- 错误消息："系统错误，无法获取服务器信息，请联系管理员"
- 日志记录：记录MAC地址获取失败的原因

## 6. 安全性考虑

### 6.1 许可文件保护

**建议措施**：
1. 许可文件应放在 `config/` 目录下，确保不在版本控制中提交
2. 如果使用 `licenseKey` 字段，可以使用HMAC-SHA256签名验证许可文件完整性
3. 许可文件权限设置（Linux）：`chmod 600 config/license.json`

### 6.2 MAC地址伪造防护

**局限性**：
- MAC地址可以伪造，但在服务器环境中通常不会
- 如果需要更强的保护，可以考虑：
  - 绑定多个硬件特征（CPU ID、硬盘序列号等）
  - 使用硬件指纹组合

**当前方案**：
- 仅绑定MAC地址，适合大多数场景
- 如需增强，可在后续版本中扩展

### 6.3 日期检查

**实现细节**：
- 使用系统当前时间进行日期比较
- 考虑时区问题（使用UTC时间或服务器本地时间）
- 建议使用 `time.Now()` 获取当前时间

## 7. 测试场景

### 7.1 正常场景
- 许可未过期，MAC地址匹配 → 允许登录

### 7.2 过期场景
- 许可已过期 → 拒绝登录，提示"系统错误，请联系管理员"

### 7.3 MAC不匹配场景
- MAC地址不匹配 → 拒绝登录，提示"系统错误，请联系管理员"

### 7.4 文件不存在场景
- 许可文件不存在 → 拒绝登录，提示"系统错误，请联系管理员"

### 7.5 文件格式错误场景
- 许可文件格式错误 → 拒绝登录，提示"系统错误，请联系管理员"

## 8. 部署说明

### 8.1 首次部署
1. 获取服务器MAC地址
2. 使用许可生成工具生成许可文件
3. 将许可文件放到 `config/license.json`
4. 启动系统

### 8.2 许可更新
1. 使用许可生成工具生成新的许可文件
2. 替换 `config/license.json` 文件
3. 重启系统（或用户重新登录）

### 8.3 服务器迁移
1. 获取新服务器的MAC地址
2. 使用许可生成工具生成新的许可文件（使用新MAC地址）
3. 替换许可文件
4. 迁移系统

## 9. 扩展性设计

### 9.1 未来可能的扩展

1. **多MAC地址支持**：允许绑定多个MAC地址（适用于多网卡环境）
2. **硬件指纹组合**：绑定多个硬件特征（CPU ID、主板序列号等）
3. **在线激活**：通过网络激活许可（需要额外的激活服务器）
4. **许可加密**：使用AES加密许可文件内容
5. **许可统计**：记录许可使用情况（登录次数、使用时间等）

### 9.2 扩展接口设计

许可信息结构可以扩展为：
```json
{
  "expireDate": "2026-12-30",
  "macAddresses": ["00:11:22:33:44:55", "AA:BB:CC:DD:EE:FF"],
  "hardwareFingerprint": "...",
  "issueDate": "2024-01-01",
  "licenseKey": "...",
  "features": ["feature1", "feature2"],
  "maxUsers": 100
}
```

## 10. 实施步骤建议

### 阶段1：基础实现
1. 创建 `internal/license` 目录和基础模块
2. 实现MAC地址获取功能
3. 实现许可验证功能
4. 在登录时集成许可检查

### 阶段2：工具开发
1. 创建许可生成工具
2. 测试工具功能

### 阶段3：测试和优化
1. 全面测试各种场景
2. 优化错误处理
3. 优化日志记录

### 阶段4：文档和部署
1. 编写使用文档
2. 准备部署说明
3. 生成初始许可文件

