package auditprogress

import (
	"database/sql"
	"fmt"
	"ops-web/internal/db"
	"ops-web/internal/logger"
	"time"
)

// QueryOptions 查询选项
type QueryOptions struct {
	SearchName   string // 档案名称搜索
	AuditStatus  string // 审核状态筛选
	ArchiveType  string // 建档类型筛选
	SampleStatus string // 抽检状态筛选
	Tag          string // 标签搜索
	Tab          string // 选项卡：all（全部任务）或 overdue（超时未整改）
	Page         int    // 页码
	PageSize     int    // 每页数量
}

// QueryResult 查询结果
type QueryResult struct {
	Tasks      []AuditTask
	TotalCount int
}

// BuildWhereClause 构建WHERE子句和参数
func BuildWhereClause(options QueryOptions) (string, []interface{}) {
	whereSQL := " WHERE 1=1"
	args := []interface{}{}

	// 选项卡条件处理
	if options.Tab == "overdue" {
		// 超时未整改选项卡：只显示审核状态为"已审核待整改"且超过15天的任务
		whereSQL += " AND audit_status = ?"
		args = append(args, "已审核待整改")
		
		// 最后更新时间到当前时间超过15天
		whereSQL += " AND DATEDIFF(NOW(), updated_at) > 15"
	} else {
		// 全部任务选项卡：排除超时未整改的任务（审核状态为"已审核待整改"且超过15天的任务不显示）
		whereSQL += ` AND NOT (audit_status = '已审核待整改' AND DATEDIFF(NOW(), updated_at) > 15)`
	}

	// 档案名称搜索
	if options.SearchName != "" {
		whereSQL += " AND file_name LIKE ?"
		args = append(args, "%"+options.SearchName+"%")
	}

	// 标签搜索
	if options.Tag != "" {
		whereSQL += " AND tag LIKE ?"
		args = append(args, "%"+options.Tag+"%")
	}

	// 审核状态查询（全部任务选项卡时允许筛选）
	if options.AuditStatus != "" {
		// 验证审核状态值
		validStatuses := map[string]bool{
			"未审核":       true,
			"已审核待整改":  true,
			"已完成":      true,
		}
		if validStatuses[options.AuditStatus] {
			// 如果是超时未整改选项卡，审核状态已经被固定为"已审核待整改"，不需要再添加
			if options.Tab != "overdue" {
				whereSQL += " AND audit_status = ?"
				args = append(args, options.AuditStatus)
			}
		}
	}

	// 建档类型查询
	if options.ArchiveType != "" {
		// 验证建档类型值
		validTypes := map[string]bool{
			"新增":   true,
			"取推":   true,
			"补档案": true,
			"变更":   true,
		}
		if validTypes[options.ArchiveType] {
			whereSQL += " AND archive_type = ?"
			args = append(args, options.ArchiveType)
		}
	}

	// 抽检状态查询（超时未整改选项卡不适用，因为超时未整改是针对审核阶段的"已审核待整改"状态，而抽检状态是针对"已完成"状态的）
	if options.SampleStatus != "" && options.Tab != "overdue" {
		validSampleStatuses := map[string]bool{
			"已抽检": true,
			"待整改": true,
			"待抽检": true,
		}
		if validSampleStatuses[options.SampleStatus] {
			if options.SampleStatus == "已抽检" {
				// 已抽检：审核状态为"已完成"，is_sampled = 1，且最近一次抽检结果不是"待整改"（包括NULL）
				whereSQL += ` AND audit_status = '已完成' 
					AND is_sampled = 1 
					AND COALESCE((SELECT sample_result FROM audit_sample_records 
						WHERE task_id = audit_tasks.id 
						ORDER BY sampled_at DESC LIMIT 1), '') != '待整改'`
			} else if options.SampleStatus == "待整改" {
				// 待整改：审核状态为"已完成"，is_sampled = 1，且最近一次抽检结果为"待整改"
				whereSQL += ` AND audit_status = '已完成' 
					AND is_sampled = 1 
					AND (SELECT sample_result FROM audit_sample_records 
						WHERE task_id = audit_tasks.id 
						ORDER BY sampled_at DESC LIMIT 1) = '待整改'`
			} else if options.SampleStatus == "待抽检" {
				// 待抽检：审核状态为"已完成"，is_sampled = 0 或 NULL
				whereSQL += ` AND audit_status = '已完成' 
					AND (is_sampled = 0 OR is_sampled IS NULL)`
			}
		}
	}

	return whereSQL, args
}

// GetOrderByClause 获取排序子句
func GetOrderByClause(tab string) string {
	if tab == "overdue" {
		// 超时未整改按更新时间倒序排列
		return " ORDER BY updated_at DESC"
	}
	// 全部任务按ID倒序排列（保持原有逻辑）
	return " ORDER BY id DESC"
}

// QueryTasks 查询任务列表
func QueryTasks(options QueryOptions) (*QueryResult, error) {
	// 构建WHERE子句
	whereSQL, args := BuildWhereClause(options)

	// 1. 查询总记录数
	var totalCount int
	countSQL := "SELECT COUNT(*) FROM audit_tasks" + whereSQL
	err := db.DBInstance.QueryRow(countSQL, args...).Scan(&totalCount)
	if err != nil && err != sql.ErrNoRows {
		logger.Errorf("审核进度-查询总数失败: %v, SQL: %s, Args: %v", err, countSQL, args)
		return nil, fmt.Errorf("查询总数失败: %w", err)
	}

	// 2. 分页计算
	offset := (options.Page - 1) * options.PageSize

	// 3. 查询列表数据（包含抽检字段）
	orderBy := GetOrderByClause(options.Tab)
	querySQL := fmt.Sprintf(
		"SELECT id, file_name, organization, import_time, audit_status, record_count, audit_comment, updated_at, is_single_soldier, archive_type, tag, is_sampled, last_sampled_at FROM audit_tasks %s%s LIMIT ? OFFSET ?",
		whereSQL,
		orderBy,
	)

	// 准备完整的参数列表
	queryArgs := append(args, options.PageSize, offset)

	rows, err := db.DBInstance.Query(querySQL, queryArgs...)
	if err != nil {
		logger.Errorf("审核进度-数据库查询失败: %v, SQL: %s, Args: %v", err, querySQL, queryArgs)
		return nil, fmt.Errorf("数据库查询失败: %w", err)
	}
	defer rows.Close()

	var taskList []AuditTask
	var taskIDs []int
	for rows.Next() {
		var task AuditTask
		var importTimeRaw, updatedAtRaw, lastSampledAtRaw sql.NullString
		var isSampled int
		err = rows.Scan(
			&task.ID,
			&task.FileName,
			&task.Organization,
			&importTimeRaw,
			&task.AuditStatus,
			&task.RecordCount,
			&task.AuditComment,
			&updatedAtRaw,
			&task.IsSingleSoldier,
			&task.ArchiveType,
			&task.Tag,
			&isSampled,
			&lastSampledAtRaw,
		)

		if err != nil {
			logger.Errorf("设备审核进度-数据库扫描错误: %v, taskID: %d", err, task.ID)
			continue
		}

		// 格式化时间字段为 YYYY-MM-DD HH:mm
		task.ImportTime = formatDateTime(importTimeRaw.String)
		task.UpdatedAt = formatDateTime(updatedAtRaw.String)
		task.IsSampled = isSampled == 1
		if lastSampledAtRaw.Valid {
			task.LastSampledAt = formatDateTime(lastSampledAtRaw.String)
		}

		taskIDs = append(taskIDs, task.ID)
		taskList = append(taskList, task)
	}

	// 检查遍历过程中的错误
	if err = rows.Err(); err != nil {
		logger.Errorf("审核进度-数据遍历失败: %v", err)
		return nil, fmt.Errorf("数据遍历失败: %w", err)
	}

	// 批量获取抽检信息
	sampleInfoMap, err := BatchGetSampleInfo(taskIDs)
	if err != nil {
		logger.Errorf("批量获取抽检信息失败: %v", err)
		// 不阻断流程，继续执行
		sampleInfoMap = make(map[int]*SampleInfo)
	}

	// 填充抽检信息到任务列表
	for i := range taskList {
		if info, ok := sampleInfoMap[taskList[i].ID]; ok {
			taskList[i].IsSampled = info.IsSampled
			taskList[i].LastSampledAt = info.LastSampledAt
			taskList[i].SampledBy = info.SampledBy
			taskList[i].SampleCount = info.SampleCount
			taskList[i].LastSampleResult = info.LastSampleResult
		}
	}

	// 批量获取提醒数量
	reminderCountMap := make(map[int]int)
	for _, taskID := range taskIDs {
		count, err := GetReminderCountByTaskID(taskID)
		if err == nil {
			reminderCountMap[taskID] = count
		}
	}

	// 填充提醒数量到任务列表
	for i := range taskList {
		if count, ok := reminderCountMap[taskList[i].ID]; ok {
			taskList[i].ReminderCount = count
		}
	}

	return &QueryResult{
		Tasks:      taskList,
		TotalCount: totalCount,
	}, nil
}

// GetOverdueDays 计算超时天数（用于显示）
func GetOverdueDays(updatedAt time.Time) int {
	days := int(time.Since(updatedAt).Hours() / 24)
	return days
}

