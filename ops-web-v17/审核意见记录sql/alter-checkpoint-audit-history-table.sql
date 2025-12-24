-- ============================================
-- 修改 checkpoint_audit_history 表以匹配 checkpoint_tasks 表结构
-- ============================================
-- 说明：此脚本用于处理 checkpoint_tasks.id 字段类型不匹配的问题
-- 如果 checkpoint_tasks.id 是 bigint(20) unsigned，需要修改 checkpoint_audit_history.task_id 的类型
-- 执行顺序：
-- 1. 先执行 create-checkpoint-audit-history-table.sql 创建表
-- 2. 再执行此脚本修改字段类型（如果需要）
-- 3. 最后执行 add-checkpoint-audit-history-foreign-key.sql 添加外键约束

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
    'SELECT "checkpoint_tasks 表存在，继续检查字段类型" AS message'
);

PREPARE stmt FROM @sqlstmt;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 检查 checkpoint_tasks.id 的实际类型
SET @id_column_type := (
    SELECT COLUMN_TYPE 
    FROM information_schema.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'checkpoint_tasks' 
    AND COLUMN_NAME = 'id'
);

-- 检查 checkpoint_audit_history 表是否存在
SET @history_table_exists := (
    SELECT COUNT(*) 
    FROM information_schema.TABLES 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'checkpoint_audit_history'
);

-- 如果 checkpoint_tasks.id 是 bigint(20) unsigned，需要修改 checkpoint_audit_history.task_id
SET @sqlstmt2 := IF(
    @history_table_exists > 0 AND (@id_column_type LIKE '%bigint%unsigned%' OR @id_column_type LIKE '%bigint(20)%unsigned%'),
    'ALTER TABLE `checkpoint_audit_history` MODIFY COLUMN `task_id` bigint(20) unsigned NOT NULL COMMENT ''审核任务ID，关联checkpoint_tasks表''',
    IF(
        @history_table_exists = 0,
        'SELECT "警告：checkpoint_audit_history 表不存在，请先执行 create-checkpoint-audit-history-table.sql" AS message',
        'SELECT "checkpoint_tasks.id 类型为 int，无需修改 task_id 类型" AS message'
    )
);

PREPARE stmt2 FROM @sqlstmt2;
EXECUTE stmt2;
DEALLOCATE PREPARE stmt2;

