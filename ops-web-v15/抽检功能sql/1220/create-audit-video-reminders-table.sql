-- 创建设备审核录像天数不足提醒表
CREATE TABLE IF NOT EXISTS audit_video_reminders (
    id INT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    task_id BIGINT(20) UNSIGNED NOT NULL COMMENT '关联的审核任务ID',
    earliest_video_date DATE NOT NULL COMMENT '最早录像日期（从审核意见中提取）',
    required_days INT NOT NULL COMMENT '要求的天数（30/90/180）',
    actual_days INT NOT NULL COMMENT '实际天数（审核时计算：审核日期 - 最早录像日期）',
    reminder_date DATE NOT NULL COMMENT '提醒日期（最早录像日期 + 要求天数）',
    status VARCHAR(20) DEFAULT 'pending' COMMENT '状态：pending-待处理, notified-已通知, completed-已完成',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    notified_at DATETIME NULL COMMENT '通知时间',
    completed_at DATETIME NULL COMMENT '完成时间（标记为已处理的时间）',
    completed_by VARCHAR(100) NULL COMMENT '处理人',
    INDEX idx_reminder_date (reminder_date),
    INDEX idx_status (status),
    INDEX idx_task_id (task_id),
    FOREIGN KEY (task_id) REFERENCES audit_tasks(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='设备审核录像天数不足提醒表';

