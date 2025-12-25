-- 创建录像提醒定时任务配置表
CREATE TABLE IF NOT EXISTS audit_video_reminder_schedule_config (
    id INT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    frequency VARCHAR(20) NOT NULL DEFAULT 'daily' COMMENT '执行频率：daily-每天, weekly-每周',
    hour INT NOT NULL DEFAULT 1 COMMENT '执行时间（小时）：1-24',
    day_of_week INT NULL COMMENT '每周执行日期（1-7，1=周一，7=周日），仅当frequency=weekly时有效',
    enabled TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用：0-禁用，1-启用',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    updated_by VARCHAR(100) NULL COMMENT '更新人',
    UNIQUE KEY uk_config (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='设备审核录像提醒定时任务配置表';

-- 插入默认配置（每天凌晨1点）
INSERT INTO audit_video_reminder_schedule_config (frequency, hour, enabled, updated_by) 
VALUES ('daily', 1, 1, 'system')
ON DUPLICATE KEY UPDATE frequency=VALUES(frequency), hour=VALUES(hour);





