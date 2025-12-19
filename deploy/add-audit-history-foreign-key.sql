-- ============================================
-- 为 audit_audit_history 表添加外键约束
-- ============================================
-- 说明：在 audit_audit_history 表创建后，可以使用此脚本添加外键约束
-- 执行前请确保：
-- 1. audit_tasks 表已存在
-- 2. audit_tasks.id 字段类型为 int(11)
-- 3. audit_audit_history 表已创建
-- 4. audit_audit_history.task_id 字段类型为 int(11)

-- 检查并删除已存在的约束（如果存在）
SET @exist := (SELECT COUNT(*) FROM information_schema.table_constraints 
               WHERE constraint_schema = DATABASE() 
               AND table_name = 'audit_audit_history' 
               AND constraint_name = 'audit_audit_history_ibfk_1');

SET @sqlstmt := IF(@exist > 0,
  'ALTER TABLE `audit_audit_history` DROP FOREIGN KEY `audit_audit_history_ibfk_1`',
  'SELECT "约束不存在，跳过删除" AS message');

PREPARE stmt FROM @sqlstmt;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 添加外键约束
ALTER TABLE `audit_audit_history` 
ADD CONSTRAINT `audit_audit_history_ibfk_1` 
FOREIGN KEY (`task_id`) 
REFERENCES `audit_tasks` (`id`) 
ON DELETE RESTRICT;

