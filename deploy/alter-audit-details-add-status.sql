-- 为 audit_details 表添加 audit_status 字段
ALTER TABLE `audit_details` 
ADD COLUMN `audit_status` tinyint(4) NOT NULL DEFAULT 0 COMMENT '建档状态：0-未审核未建档，1-已审核未建档，2-已建档' 
AFTER `collection_area_type`;

-- 添加索引以提高查询性能
ALTER TABLE `audit_details` 
ADD INDEX `idx_audit_status` (`audit_status`);




