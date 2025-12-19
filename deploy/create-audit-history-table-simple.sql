-- ============================================
-- 设备审核意见历史记录表（简化版本，不包含外键约束）
-- ============================================
-- 说明：此版本不包含外键约束，如果遇到外键约束问题，可以使用此版本
-- 外键约束可以在表创建后，通过单独的 ALTER 语句添加

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

-- 如果需要添加外键约束，请确保 audit_tasks 表已存在，然后执行以下语句：
-- ALTER TABLE `audit_audit_history` 
-- ADD CONSTRAINT `audit_audit_history_ibfk_1` 
-- FOREIGN KEY (`task_id`) 
-- REFERENCES `audit_tasks` (`id`) 
-- ON DELETE RESTRICT;

