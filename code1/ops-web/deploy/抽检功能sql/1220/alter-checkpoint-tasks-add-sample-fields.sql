-- ============================================
-- 卡口审核任务表添加抽检功能字段
-- ============================================
-- 说明：为 checkpoint_tasks 表添加抽检相关字段，用于快速查询和显示抽检状态
-- 执行时间：2025-12-20
-- 功能：支持对审核状态为"已完成"的任务进行抽检

-- 检查 checkpoint_tasks 表是否存在
SET @table_exists := (
    SELECT COUNT(*) 
    FROM information_schema.TABLES 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'checkpoint_tasks'
);

-- 如果表不存在，退出
SET @sqlstmt := IF(
    @table_exists = 0,
    'SELECT "错误：checkpoint_tasks 表不存在，请先创建该表" AS message',
    'SELECT "checkpoint_tasks 表存在，开始添加抽检字段" AS message'
);

PREPARE stmt FROM @sqlstmt;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 检查字段是否已存在，如果不存在则添加
SET @is_sampled_exists := (
    SELECT COUNT(*) 
    FROM information_schema.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'checkpoint_tasks' 
    AND COLUMN_NAME = 'is_sampled'
);

-- 添加 is_sampled 字段（是否已抽检）
SET @sqlstmt2 := IF(
    @is_sampled_exists = 0,
    'ALTER TABLE `checkpoint_tasks` ADD COLUMN `is_sampled` TINYINT(1) NOT NULL DEFAULT 0 COMMENT ''是否已抽检：0-未抽检，1-已抽检'' AFTER `audit_status`',
    'SELECT "字段 is_sampled 已存在，跳过添加" AS message'
);

PREPARE stmt2 FROM @sqlstmt2;
EXECUTE stmt2;
DEALLOCATE PREPARE stmt2;

-- 检查 last_sampled_at 字段是否已存在
SET @last_sampled_at_exists := (
    SELECT COUNT(*) 
    FROM information_schema.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'checkpoint_tasks' 
    AND COLUMN_NAME = 'last_sampled_at'
);

-- 添加 last_sampled_at 字段（最后抽检时间）
SET @sqlstmt3 := IF(
    @last_sampled_at_exists = 0,
    'ALTER TABLE `checkpoint_tasks` ADD COLUMN `last_sampled_at` TIMESTAMP NULL DEFAULT NULL COMMENT ''最后抽检时间'' AFTER `is_sampled`',
    'SELECT "字段 last_sampled_at 已存在，跳过添加" AS message'
);

PREPARE stmt3 FROM @sqlstmt3;
EXECUTE stmt3;
DEALLOCATE PREPARE stmt3;

-- 检查索引是否已存在
SET @index_exists := (
    SELECT COUNT(*) 
    FROM information_schema.STATISTICS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'checkpoint_tasks' 
    AND INDEX_NAME = 'idx_is_sampled'
);

-- 添加索引以提高查询性能
SET @sqlstmt4 := IF(
    @index_exists = 0,
    'ALTER TABLE `checkpoint_tasks` ADD INDEX `idx_is_sampled` (`is_sampled`)',
    'SELECT "索引 idx_is_sampled 已存在，跳过添加" AS message'
);

PREPARE stmt4 FROM @sqlstmt4;
EXECUTE stmt4;
DEALLOCATE PREPARE stmt4;

-- 完成提示
SELECT "卡口审核任务表抽检字段添加完成" AS message;
