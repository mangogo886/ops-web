-- 为 audit_tasks 和 checkpoint_tasks 表添加 completed_at 字段
-- 该字段用于记录首次完成建档的时间，固定不变

-- 为设备审核任务表添加 completed_at 字段
ALTER TABLE `audit_tasks` 
ADD COLUMN `completed_at` timestamp NULL COMMENT '建档完成时间（首次完成时间，固定不变）' 
AFTER `updated_at`;

-- 为卡口审核任务表添加 completed_at 字段
ALTER TABLE `checkpoint_tasks` 
ADD COLUMN `completed_at` timestamp NULL COMMENT '建档完成时间（首次完成时间，固定不变）' 
AFTER `updated_at`;

-- 添加索引以提高查询性能
CREATE INDEX `idx_audit_tasks_completed_at` ON `audit_tasks`(`completed_at`);
CREATE INDEX `idx_checkpoint_tasks_completed_at` ON `checkpoint_tasks`(`completed_at`);






