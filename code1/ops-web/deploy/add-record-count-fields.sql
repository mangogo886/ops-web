-- 为audit_tasks表添加record_count字段
ALTER TABLE `audit_tasks` 
ADD COLUMN `record_count` int(11) NOT NULL DEFAULT 0 COMMENT '导入记录数量' AFTER `audit_status`;

-- 为checkpoint_tasks表添加record_count字段
ALTER TABLE `checkpoint_tasks` 
ADD COLUMN `record_count` int(11) NOT NULL DEFAULT 0 COMMENT '导入记录数量' AFTER `audit_status`;
