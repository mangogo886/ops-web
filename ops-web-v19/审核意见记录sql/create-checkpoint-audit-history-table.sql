-- ============================================
-- 卡口审核意见历史记录表（简化版本，不包含外键约束）
-- ============================================
-- 说明：此版本不包含外键约束，如果遇到外键约束问题，可以使用此版本
-- 外键约束可以在表创建后，通过单独的 ALTER 语句添加
--
-- 执行顺序：
-- 1. 执行此脚本创建 checkpoint_audit_history 表
-- 2. 执行 alter-checkpoint-audit-history-table.sql 修改字段类型（如果需要）
--    注意：如果 checkpoint_tasks.id 是 bigint(20) unsigned，需要先执行此脚本
-- 3. 执行 add-checkpoint-audit-history-foreign-key.sql 添加外键约束

CREATE TABLE IF NOT EXISTS `checkpoint_audit_history` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `task_id` int(11) NOT NULL COMMENT '审核任务ID，关联checkpoint_tasks表',
  `audit_comment` text DEFAULT NULL COMMENT '审核意见（历史记录）',
  `audit_status` varchar(50) NOT NULL COMMENT '审核状态（历史记录）',
  `auditor` varchar(50) NOT NULL COMMENT '审核人用户名',
  `audit_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '审核时间',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
  PRIMARY KEY (`id`),
  KEY `task_id` (`task_id`),
  KEY `audit_time` (`audit_time`),
  KEY `auditor` (`auditor`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='卡口审核意见历史记录表';

-- 注意：
-- 1. 如果 checkpoint_tasks.id 字段类型是 bigint(20) unsigned，需要执行 alter-checkpoint-audit-history-table.sql
--    将 task_id 字段类型修改为 bigint(20) unsigned 以匹配 checkpoint_tasks.id
-- 2. 如果需要添加外键约束，请执行 add-checkpoint-audit-history-foreign-key.sql

