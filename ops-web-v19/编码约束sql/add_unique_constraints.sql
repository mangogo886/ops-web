-- 检查并添加唯一约束的SQL脚本
-- 
-- 说明：
-- 1. 为 audit_details 表的 device_code 字段添加唯一约束
-- 2. 为 checkpoint_details 表的 checkpoint_code 字段添加唯一约束
--
-- 使用方法：
-- 1. 先执行检查语句，确认当前是否有唯一约束
-- 2. 如果字段中存在重复数据，需要先清理重复数据
-- 3. 执行添加约束的ALTER语句

-- ============================================
-- 1. 检查 audit_details 表的 device_code 字段是否有唯一约束
-- ============================================
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

-- 如果上面的查询没有返回结果，说明没有唯一约束，需要执行下面的ALTER语句添加约束

-- ============================================
-- 2. 检查 checkpoint_details 表的 checkpoint_code 字段是否有唯一约束
-- ============================================
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

-- 如果上面的查询没有返回结果，说明没有唯一约束，需要执行下面的ALTER语句添加约束

-- ============================================
-- 3. 在添加约束前，检查是否存在重复数据
-- ============================================

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

-- ============================================
-- 4. 如果存在重复数据，需要先处理重复数据（保留最新的记录，删除旧的记录）
-- ============================================

-- 注意：执行删除重复数据前，请先备份数据！
-- 删除 audit_details 表中的重复数据（保留id最大的记录）
-- DELETE ad1 FROM audit_details ad1
-- INNER JOIN audit_details ad2 
-- WHERE ad1.id < ad2.id 
-- AND ad1.device_code = ad2.device_code
-- AND ad1.device_code IS NOT NULL 
-- AND ad1.device_code != '';

-- 删除 checkpoint_details 表中的重复数据（保留id最大的记录）
-- DELETE cd1 FROM checkpoint_details cd1
-- INNER JOIN checkpoint_details cd2 
-- WHERE cd1.id < cd2.id 
-- AND cd1.checkpoint_code = cd2.checkpoint_code
-- AND cd1.checkpoint_code IS NOT NULL 
-- AND cd1.checkpoint_code != '';

-- ============================================
-- 5. 添加唯一约束（如果约束不存在）
-- ============================================

-- 为 audit_details 表的 device_code 字段添加唯一约束
-- 注意：如果字段中存在NULL值，MySQL允许NULL值存在，但非NULL值必须唯一
ALTER TABLE audit_details 
ADD UNIQUE INDEX uk_device_code (device_code);

-- 为 checkpoint_details 表的 checkpoint_code 字段添加唯一约束
-- 注意：如果字段中存在NULL值，MySQL允许NULL值存在，但非NULL值必须唯一
ALTER TABLE checkpoint_details 
ADD UNIQUE INDEX uk_checkpoint_code (checkpoint_code);

