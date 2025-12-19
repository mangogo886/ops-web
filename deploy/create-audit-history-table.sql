-- ============================================
-- 设备审核意见历史记录表
-- ============================================
-- 说明：此版本不包含外键约束，避免外键约束创建失败的问题
-- 外键约束可以在表创建后，通过执行 add-audit-history-foreign-key.sql 来添加
-- 或者直接使用 create-audit-history-table-simple.sql 版本

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

-- 如果需要添加外键约束，请执行：deploy/add-audit-history-foreign-key.sql

