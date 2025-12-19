-- ============================================
-- 设备审核意见历史记录表（完整版本，包含外键约束）
-- ============================================
-- 说明：此脚本创建 audit_audit_history 表并添加外键约束
-- 执行前请确保：
-- 1. audit_tasks 表已存在
-- 2. audit_tasks.id 字段类型为 int(11) 或 bigint(20) unsigned
--
-- 注意：如果 audit_tasks.id 是 bigint(20) unsigned，需要先执行此脚本创建表，
--       然后执行 alter-audit-history-table.sql 修改 task_id 字段类型

-- 创建审核意见历史记录表
CREATE TABLE IF NOT EXISTS `audit_audit_history` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `task_id` int(11) NOT NULL COMMENT '审核任务ID，关联audit_tasks表',
  `audit_comment` text DEFAULT NULL COMMENT '审核意见（历史记录）',
  `audit_status` varchar(50) NOT NULL COMMENT '审核状态（历史记录）',
  `auditor` varchar(50) NOT NULL COMMENT '审核人用户名',
  `audit_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '审核时间',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
  PRIMARY KEY (`id`),
  KEY `task_id` (`task_id`),
  KEY `audit_time` (`audit_time`),
  KEY `auditor` (`auditor`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='设备审核意见历史记录表';

-- 检查并删除已存在的外键约束（如果存在）
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
-- 注意：如果 audit_tasks.id 是 bigint(20) unsigned，需要先执行 alter-audit-history-table.sql
--       修改 task_id 字段类型后再执行外键约束添加
ALTER TABLE `audit_audit_history` 
ADD CONSTRAINT `audit_audit_history_ibfk_1` 
FOREIGN KEY (`task_id`) 
REFERENCES `audit_tasks` (`id`) 
ON DELETE RESTRICT;

