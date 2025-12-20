# 唯一约束添加说明

## 概述

本文档说明如何为 `audit_details` 表的 `device_code` 字段和 `checkpoint_details` 表的 `checkpoint_code` 字段添加唯一约束，以及导入代码如何处理唯一约束错误。

## 数据库表唯一约束

### 1. audit_details 表

为 `device_code` 字段添加唯一约束，确保设备编码在系统中唯一。

### 2. checkpoint_details 表

为 `checkpoint_code` 字段添加唯一约束，确保卡口编号在系统中唯一。

## 使用步骤

### 步骤1：检查当前约束状态

执行 `add_unique_constraints.sql` 文件中的检查语句，查看当前是否已有唯一约束：

```sql
-- 检查 audit_details 表的约束
SELECT 
    CONSTRAINT_NAME,
    CONSTRAINT_TYPE
FROM 
    INFORMATION_SCHEMA.TABLE_CONSTRAINTS
WHERE 
    TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'audit_details'
    AND CONSTRAINT_TYPE = 'UNIQUE'
    AND CONSTRAINT_NAME LIKE '%device_code%';

-- 检查 checkpoint_details 表的约束
SELECT 
    CONSTRAINT_NAME,
    CONSTRAINT_TYPE
FROM 
    INFORMATION_SCHEMA.TABLE_CONSTRAINTS
WHERE 
    TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'checkpoint_details'
    AND CONSTRAINT_TYPE = 'UNIQUE'
    AND CONSTRAINT_NAME LIKE '%checkpoint_code%';
```

### 步骤2：检查重复数据

在添加约束前，需要检查是否存在重复数据：

```sql
-- 检查 audit_details 表中的重复 device_code
SELECT 
    device_code,
    COUNT(*) as duplicate_count
FROM 
    audit_details
WHERE 
    device_code IS NOT NULL 
    AND device_code != ''
GROUP BY 
    device_code
HAVING 
    COUNT(*) > 1;

-- 检查 checkpoint_details 表中的重复 checkpoint_code
SELECT 
    checkpoint_code,
    COUNT(*) as duplicate_count
FROM 
    checkpoint_details
WHERE 
    checkpoint_code IS NOT NULL 
    AND checkpoint_code != ''
GROUP BY 
    checkpoint_code
HAVING 
    COUNT(*) > 1;
```

### 步骤3：处理重复数据（如有）

如果存在重复数据，需要先处理。建议保留最新的记录（id最大的记录），删除旧的记录：

```sql
-- 删除 audit_details 表中的重复数据（保留id最大的记录）
-- ⚠️ 执行前请先备份数据！
DELETE ad1 FROM audit_details ad1
INNER JOIN audit_details ad2 
WHERE ad1.id < ad2.id 
AND ad1.device_code = ad2.device_code
AND ad1.device_code IS NOT NULL 
AND ad1.device_code != '';

-- 删除 checkpoint_details 表中的重复数据（保留id最大的记录）
-- ⚠️ 执行前请先备份数据！
DELETE cd1 FROM checkpoint_details cd1
INNER JOIN checkpoint_details cd2 
WHERE cd1.id < cd2.id 
AND cd1.checkpoint_code = cd2.checkpoint_code
AND cd1.checkpoint_code IS NOT NULL 
AND cd1.checkpoint_code != '';
```

### 步骤4：添加唯一约束

确认没有重复数据后，执行以下ALTER语句添加唯一约束：

```sql
-- 为 audit_details 表的 device_code 字段添加唯一约束
ALTER TABLE audit_details 
ADD UNIQUE INDEX uk_device_code (device_code);

-- 为 checkpoint_details 表的 checkpoint_code 字段添加唯一约束
ALTER TABLE checkpoint_details 
ADD UNIQUE INDEX uk_checkpoint_code (checkpoint_code);
```

## 代码功能说明

### 导入时的唯一约束处理

当导入Excel文件时，如果遇到违反唯一约束的情况：

1. **事务回滚**：整个导入操作会被回滚，确保数据一致性
2. **错误检测**：代码会自动检测MySQL的唯一约束错误（错误代码1062）
3. **错误信息提取**：从错误信息中提取违反约束的字段值
4. **JSON响应**：返回JSON格式的错误信息，包含：
   - `error`: 错误类型（"唯一约束违反"）
   - `message`: 错误消息（包含行号）
   - `field`: 违反约束的字段名（"device_code" 或 "checkpoint_code"）
   - `fieldValue`: 违反约束的字段值
   - `detail`: 详细错误描述

### 前端处理

前端收到唯一约束错误时，应该：
1. 检测响应状态码为 400 (Bad Request)
2. 解析JSON响应
3. 弹窗显示错误信息，包括违反约束的字段值
4. 提示用户检查Excel文件中的重复数据

### 示例错误响应

```json
{
  "error": "唯一约束违反",
  "message": "第 5 行数据违反唯一约束",
  "field": "device_code",
  "fieldValue": "DEV001",
  "detail": "设备编码 'DEV001' 已存在，不能重复导入"
}
```

## 注意事项

1. ⚠️ **备份数据**：在执行删除重复数据的SQL前，请务必备份数据库
2. ⚠️ **NULL值处理**：MySQL的唯一约束允许NULL值存在多个，只对非NULL值进行唯一性检查
3. **空字符串处理**：空字符串会被视为有效值，也会被唯一约束检查
4. **索引命名**：唯一约束的索引名分别为 `uk_device_code` 和 `uk_checkpoint_code`

