-- ============================================
-- 为 checkpoint_audit_history 表添加外键约束
-- ============================================
-- 说明：在 checkpoint_audit_history 表创建后，可以使用此脚本添加外键约束
-- 执行前请确保：
-- 1. checkpoint_tasks 表已存在
-- 2. checkpoint_tasks.id 字段类型为 int(11) 或 bigint(20) unsigned
-- 3. checkpoint_audit_history 表已创建
-- 4. checkpoint_audit_history.task_id 字段类型与 checkpoint_tasks.id 类型匹配

-- 检查并删除已存在的约束（如果存在）
SET @exist := (SELECT COUNT(*) FROM information_schema.table_constraints 
               WHERE constraint_schema = DATABASE() 
               AND table_name = 'checkpoint_audit_history' 
               AND constraint_name = 'checkpoint_audit_history_ibfk_1');

SET @sqlstmt := IF(@exist > 0,
  'ALTER TABLE `checkpoint_audit_history` DROP FOREIGN KEY `checkpoint_audit_history_ibfk_1`',
  'SELECT "约束不存在，跳过删除" AS message');

PREPARE stmt FROM @sqlstmt;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 添加外键约束
ALTER TABLE `checkpoint_audit_history` 
ADD CONSTRAINT `checkpoint_audit_history_ibfk_1` 
FOREIGN KEY (`task_id`) 
REFERENCES `checkpoint_tasks` (`id`) 
ON DELETE RESTRICT;

