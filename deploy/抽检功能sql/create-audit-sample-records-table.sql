-- ============================================
-- 设备审核抽检记录表
-- ============================================
-- 说明：用于记录设备审核任务的抽检历史，支持多次抽检
-- 执行时间：2025-12-19
-- 功能：保存每次抽检的详细信息，包括抽检人员、时间、意见、结果等

-- 检查 audit_tasks 表是否存在（依赖表）
SET @table_exists := (
    SELECT COUNT(*) 
    FROM information_schema.TABLES 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'audit_tasks'
);

-- 如果依赖表不存在，退出
SET @sqlstmt := IF(
    @table_exists = 0,
    'SELECT "错误：audit_tasks 表不存在，请先创建该表" AS message',
    'SELECT "audit_tasks 表存在，开始创建抽检记录表" AS message'
);

PREPARE stmt FROM @sqlstmt;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 检查表是否已存在
SET @sample_table_exists := (
    SELECT COUNT(*) 
    FROM information_schema.TABLES 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'audit_sample_records'
);

-- 如果表不存在，则创建表
-- 注意：CREATE TABLE IF NOT EXISTS 语句不能放在动态 SQL 中，直接执行
CREATE TABLE IF NOT EXISTS `audit_sample_records` (
  `id` BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `task_id` BIGINT(20) UNSIGNED NOT NULL COMMENT '审核任务ID，关联audit_tasks表',
  `sampled_by` VARCHAR(50) NOT NULL COMMENT '抽检人员',
  `sampled_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '抽检时间',
  `sample_comment` TEXT DEFAULT NULL COMMENT '抽检意见',
  `sample_result` VARCHAR(20) DEFAULT NULL COMMENT '抽检结果：通过、不通过、待整改',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_task_id` (`task_id`),
  KEY `idx_sampled_at` (`sampled_at`),
  KEY `idx_sampled_by` (`sampled_by`),
  KEY `idx_sample_result` (`sample_result`),
  CONSTRAINT `fk_audit_sample_task` FOREIGN KEY (`task_id`) REFERENCES `audit_tasks` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='设备审核抽检记录表';

-- 完成提示
SELECT "设备审核抽检记录表创建完成" AS message;

