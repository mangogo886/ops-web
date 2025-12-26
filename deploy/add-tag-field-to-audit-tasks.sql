-- 为audit_tasks表添加标签字段
ALTER TABLE audit_tasks 
ADD COLUMN tag VARCHAR(255) NULL COMMENT '标签字段，用于搜索区分' AFTER archive_type;

-- 添加索引以优化搜索性能
CREATE INDEX idx_tag ON audit_tasks(tag);


