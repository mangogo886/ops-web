-- ============================================
-- 数据概览统计查询SQL示例
-- ============================================
-- 说明：用于数据概览页面的统计数据查询
-- 优化建议：使用LEFT JOIN替代子查询，提高查询性能

-- ============================================
-- 1. 设备审核任务统计（一周内）
-- ============================================
SELECT 
    COUNT(*) as total,
    SUM(CASE WHEN at.audit_status = '未审核' THEN 1 ELSE 0 END) as pending,
    SUM(CASE WHEN at.audit_status = '已审核待整改' THEN 1 ELSE 0 END) as need_fix,
    SUM(CASE WHEN at.audit_status = '已完成' THEN 1 ELSE 0 END) as completed,
    SUM(CASE WHEN at.audit_status = '已完成' AND at.is_sampled = 0 THEN 1 ELSE 0 END) as pending_sample,
    SUM(CASE WHEN at.audit_status = '已完成' 
        AND at.is_sampled = 1 
        AND latest_sample.sample_result = '待整改' 
        THEN 1 ELSE 0 END) as sample_need_fix,
    SUM(CASE WHEN at.audit_status = '已完成' 
        AND at.is_sampled = 1 
        AND (latest_sample.sample_result = '通过' OR latest_sample.sample_result IS NULL)
        THEN 1 ELSE 0 END) as sample_passed
FROM audit_tasks at
LEFT JOIN (
    SELECT 
        task_id,
        sample_result,
        ROW_NUMBER() OVER (PARTITION BY task_id ORDER BY sampled_at DESC) as rn
    FROM audit_sample_records
) latest_sample ON at.id = latest_sample.task_id AND latest_sample.rn = 1
WHERE at.import_time >= DATE_SUB(NOW(), INTERVAL 7 DAY);

-- ============================================
-- 2. 设备审核任务统计（一个月内）
-- ============================================
SELECT 
    COUNT(*) as total,
    SUM(CASE WHEN at.audit_status = '未审核' THEN 1 ELSE 0 END) as pending,
    SUM(CASE WHEN at.audit_status = '已审核待整改' THEN 1 ELSE 0 END) as need_fix,
    SUM(CASE WHEN at.audit_status = '已完成' THEN 1 ELSE 0 END) as completed,
    SUM(CASE WHEN at.audit_status = '已完成' AND at.is_sampled = 0 THEN 1 ELSE 0 END) as pending_sample,
    SUM(CASE WHEN at.audit_status = '已完成' 
        AND at.is_sampled = 1 
        AND latest_sample.sample_result = '待整改' 
        THEN 1 ELSE 0 END) as sample_need_fix,
    SUM(CASE WHEN at.audit_status = '已完成' 
        AND at.is_sampled = 1 
        AND (latest_sample.sample_result = '通过' OR latest_sample.sample_result IS NULL)
        THEN 1 ELSE 0 END) as sample_passed
FROM audit_tasks at
LEFT JOIN (
    SELECT 
        task_id,
        sample_result,
        ROW_NUMBER() OVER (PARTITION BY task_id ORDER BY sampled_at DESC) as rn
    FROM audit_sample_records
) latest_sample ON at.id = latest_sample.task_id AND latest_sample.rn = 1
WHERE at.import_time >= DATE_SUB(NOW(), INTERVAL 1 MONTH);

-- ============================================
-- 3. 设备审核任务统计（所有时间段）
-- ============================================
SELECT 
    COUNT(*) as total,
    SUM(CASE WHEN at.audit_status = '未审核' THEN 1 ELSE 0 END) as pending,
    SUM(CASE WHEN at.audit_status = '已审核待整改' THEN 1 ELSE 0 END) as need_fix,
    SUM(CASE WHEN at.audit_status = '已完成' THEN 1 ELSE 0 END) as completed,
    SUM(CASE WHEN at.audit_status = '已完成' AND at.is_sampled = 0 THEN 1 ELSE 0 END) as pending_sample,
    SUM(CASE WHEN at.audit_status = '已完成' 
        AND at.is_sampled = 1 
        AND latest_sample.sample_result = '待整改' 
        THEN 1 ELSE 0 END) as sample_need_fix,
    SUM(CASE WHEN at.audit_status = '已完成' 
        AND at.is_sampled = 1 
        AND (latest_sample.sample_result = '通过' OR latest_sample.sample_result IS NULL)
        THEN 1 ELSE 0 END) as sample_passed
FROM audit_tasks at
LEFT JOIN (
    SELECT 
        task_id,
        sample_result,
        ROW_NUMBER() OVER (PARTITION BY task_id ORDER BY sampled_at DESC) as rn
    FROM audit_sample_records
) latest_sample ON at.id = latest_sample.task_id AND latest_sample.rn = 1;

-- ============================================
-- 4. 卡口审核任务统计（一周内）
-- ============================================
SELECT 
    COUNT(*) as total,
    SUM(CASE WHEN ct.audit_status = '未审核' THEN 1 ELSE 0 END) as pending,
    SUM(CASE WHEN ct.audit_status = '已审核待整改' THEN 1 ELSE 0 END) as need_fix,
    SUM(CASE WHEN ct.audit_status = '已完成' THEN 1 ELSE 0 END) as completed,
    SUM(CASE WHEN ct.audit_status = '已完成' AND ct.is_sampled = 0 THEN 1 ELSE 0 END) as pending_sample,
    SUM(CASE WHEN ct.audit_status = '已完成' 
        AND ct.is_sampled = 1 
        AND latest_sample.sample_result = '待整改' 
        THEN 1 ELSE 0 END) as sample_need_fix,
    SUM(CASE WHEN ct.audit_status = '已完成' 
        AND ct.is_sampled = 1 
        AND (latest_sample.sample_result = '通过' OR latest_sample.sample_result IS NULL)
        THEN 1 ELSE 0 END) as sample_passed
FROM checkpoint_tasks ct
LEFT JOIN (
    SELECT 
        task_id,
        sample_result,
        ROW_NUMBER() OVER (PARTITION BY task_id ORDER BY sampled_at DESC) as rn
    FROM checkpoint_sample_records
) latest_sample ON ct.id = latest_sample.task_id AND latest_sample.rn = 1
WHERE ct.import_time >= DATE_SUB(NOW(), INTERVAL 7 DAY);

-- ============================================
-- 5. 设备和卡口合并统计（一周内）
-- ============================================
SELECT 
    '设备' as type,
    COUNT(*) as total,
    SUM(CASE WHEN at.audit_status = '未审核' THEN 1 ELSE 0 END) as pending,
    SUM(CASE WHEN at.audit_status = '已审核待整改' THEN 1 ELSE 0 END) as need_fix,
    SUM(CASE WHEN at.audit_status = '已完成' THEN 1 ELSE 0 END) as completed,
    SUM(CASE WHEN at.audit_status = '已完成' AND at.is_sampled = 0 THEN 1 ELSE 0 END) as pending_sample,
    SUM(CASE WHEN at.audit_status = '已完成' 
        AND at.is_sampled = 1 
        AND latest_sample.sample_result = '待整改' 
        THEN 1 ELSE 0 END) as sample_need_fix,
    SUM(CASE WHEN at.audit_status = '已完成' 
        AND at.is_sampled = 1 
        AND (latest_sample.sample_result = '通过' OR latest_sample.sample_result IS NULL)
        THEN 1 ELSE 0 END) as sample_passed
FROM audit_tasks at
LEFT JOIN (
    SELECT 
        task_id,
        sample_result,
        ROW_NUMBER() OVER (PARTITION BY task_id ORDER BY sampled_at DESC) as rn
    FROM audit_sample_records
) latest_sample ON at.id = latest_sample.task_id AND latest_sample.rn = 1
WHERE at.import_time >= DATE_SUB(NOW(), INTERVAL 7 DAY)

UNION ALL

SELECT 
    '卡口' as type,
    COUNT(*) as total,
    SUM(CASE WHEN ct.audit_status = '未审核' THEN 1 ELSE 0 END) as pending,
    SUM(CASE WHEN ct.audit_status = '已审核待整改' THEN 1 ELSE 0 END) as need_fix,
    SUM(CASE WHEN ct.audit_status = '已完成' THEN 1 ELSE 0 END) as completed,
    SUM(CASE WHEN ct.audit_status = '已完成' AND ct.is_sampled = 0 THEN 1 ELSE 0 END) as pending_sample,
    SUM(CASE WHEN ct.audit_status = '已完成' 
        AND ct.is_sampled = 1 
        AND latest_sample.sample_result = '待整改' 
        THEN 1 ELSE 0 END) as sample_need_fix,
    SUM(CASE WHEN ct.audit_status = '已完成' 
        AND ct.is_sampled = 1 
        AND (latest_sample.sample_result = '通过' OR latest_sample.sample_result IS NULL)
        THEN 1 ELSE 0 END) as sample_passed
FROM checkpoint_tasks ct
LEFT JOIN (
    SELECT 
        task_id,
        sample_result,
        ROW_NUMBER() OVER (PARTITION BY task_id ORDER BY sampled_at DESC) as rn
    FROM checkpoint_sample_records
) latest_sample ON ct.id = latest_sample.task_id AND latest_sample.rn = 1
WHERE ct.import_time >= DATE_SUB(NOW(), INTERVAL 7 DAY);

-- ============================================
-- 6. 按机构分组统计（用于图表展示）
-- ============================================
SELECT 
    at.organization,
    COUNT(*) as total,
    SUM(CASE WHEN at.audit_status = '未审核' THEN 1 ELSE 0 END) as pending,
    SUM(CASE WHEN at.audit_status = '已审核待整改' THEN 1 ELSE 0 END) as need_fix,
    SUM(CASE WHEN at.audit_status = '已完成' THEN 1 ELSE 0 END) as completed
FROM audit_tasks at
WHERE at.import_time >= DATE_SUB(NOW(), INTERVAL 7 DAY)
GROUP BY at.organization
ORDER BY total DESC;

-- ============================================
-- 7. 时间趋势统计（按天统计，用于折线图）
-- ============================================
SELECT 
    DATE(at.import_time) as date,
    COUNT(*) as count,
    SUM(CASE WHEN at.audit_status = '未审核' THEN 1 ELSE 0 END) as pending,
    SUM(CASE WHEN at.audit_status = '已完成' THEN 1 ELSE 0 END) as completed
FROM audit_tasks at
WHERE at.import_time >= DATE_SUB(NOW(), INTERVAL 30 DAY)
GROUP BY DATE(at.import_time)
ORDER BY date ASC;

-- ============================================
-- 性能优化建议
-- ============================================

-- 如果数据量大，可以考虑添加以下索引：
-- CREATE INDEX idx_audit_status_sampled_time 
-- ON audit_tasks(audit_status, is_sampled, import_time);

-- CREATE INDEX idx_checkpoint_status_sampled_time 
-- ON checkpoint_tasks(audit_status, is_sampled, import_time);

-- CREATE INDEX idx_sample_task_result 
-- ON audit_sample_records(task_id, sampled_at DESC, sample_result);

-- CREATE INDEX idx_checkpoint_sample_task_result 
-- ON checkpoint_sample_records(task_id, sampled_at DESC, sample_result);


