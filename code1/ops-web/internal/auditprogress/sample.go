package auditprogress

import (
	"database/sql"
	"ops-web/internal/db"
	"ops-web/internal/logger"
)

// SampleRecord 抽检记录结构体
type SampleRecord struct {
	ID            int
	TaskID        int
	SampledBy     string
	SampledAt     string
	SampleComment sql.NullString
	SampleResult  sql.NullString
	CreatedAt     string
}

// SampleInfo 抽检信息（用于列表页显示）
type SampleInfo struct {
	IsSampled       bool   // 是否已抽检
	LastSampledAt   string // 最后抽检时间
	SampledBy       string // 最后抽检人员
	SampleCount     int    // 抽检次数
	LastSampleResult string // 最近一次抽检结果
}

// GetSampleInfo 获取任务的抽检信息（用于列表页显示）
func GetSampleInfo(taskID int) (*SampleInfo, error) {
	// 查询任务表的抽检字段
	var isSampled int
	var lastSampledAt sql.NullString
	querySQL := `SELECT is_sampled, last_sampled_at FROM audit_tasks WHERE id = ?`
	err := db.DBInstance.QueryRow(querySQL, taskID).Scan(&isSampled, &lastSampledAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return &SampleInfo{IsSampled: false}, nil
		}
		logger.Errorf("查询任务抽检信息失败: %v, taskID: %d", err, taskID)
		return nil, err
	}

	info := &SampleInfo{
		IsSampled: isSampled == 1,
	}

	// 如果有最后抽检时间，查询最后抽检的人员信息和结果
	if lastSampledAt.Valid && lastSampledAt.String != "" {
		info.LastSampledAt = formatDateTime(lastSampledAt.String)
		
		// 查询最后一条抽检记录获取抽检人员和结果
		var sampledBy, sampleResult sql.NullString
		lastRecordSQL := `SELECT sampled_by, sample_result FROM audit_sample_records 
			WHERE task_id = ? 
			ORDER BY sampled_at DESC 
			LIMIT 1`
		err = db.DBInstance.QueryRow(lastRecordSQL, taskID).Scan(&sampledBy, &sampleResult)
		if err == nil {
			if sampledBy.Valid {
				info.SampledBy = sampledBy.String
			}
			if sampleResult.Valid {
				info.LastSampleResult = sampleResult.String
			}
		}
	}

	// 查询抽检次数
	var count int
	countSQL := `SELECT COUNT(*) FROM audit_sample_records WHERE task_id = ?`
	err = db.DBInstance.QueryRow(countSQL, taskID).Scan(&count)
	if err == nil {
		info.SampleCount = count
	}

	return info, nil
}

// SaveSampleRecord 保存抽检记录
// 同时更新任务表的抽检字段
func SaveSampleRecord(taskID int, sampledBy string, sampleComment string, sampleResult string) error {
	// 开始事务
	tx, err := db.DBInstance.Begin()
	if err != nil {
		logger.Errorf("开始事务失败: %v, taskID: %d", err, taskID)
		return err
	}
	defer tx.Rollback()

	// 1. 插入抽检记录
	insertSQL := `INSERT INTO audit_sample_records 
		(task_id, sampled_by, sampled_at, sample_comment, sample_result) 
		VALUES (?, ?, NOW(), ?, ?)`
	
	var comment interface{}
	if sampleComment != "" {
		comment = sampleComment
	} else {
		comment = nil
	}

	var result interface{}
	if sampleResult != "" {
		result = sampleResult
	} else {
		result = nil
	}

	_, err = tx.Exec(insertSQL, taskID, sampledBy, comment, result)
	if err != nil {
		logger.Errorf("插入抽检记录失败: %v, taskID: %d", err, taskID)
		return err
	}

	// 2. 更新任务表的抽检字段
	updateSQL := `UPDATE audit_tasks 
		SET is_sampled = 1, last_sampled_at = NOW() 
		WHERE id = ?`
	_, err = tx.Exec(updateSQL, taskID)
	if err != nil {
		logger.Errorf("更新任务抽检字段失败: %v, taskID: %d", err, taskID)
		return err
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		logger.Errorf("提交事务失败: %v, taskID: %d", err, taskID)
		return err
	}

	return nil
}

// GetSampleHistory 查询任务的抽检历史记录
func GetSampleHistory(taskID int) ([]SampleRecord, error) {
	querySQL := `SELECT id, task_id, sampled_by, sampled_at, sample_comment, sample_result, created_at
		FROM audit_sample_records 
		WHERE task_id = ? 
		ORDER BY sampled_at DESC`

	rows, err := db.DBInstance.Query(querySQL, taskID)
	if err != nil {
		logger.Errorf("查询抽检历史失败: %v, taskID: %d", err, taskID)
		return nil, err
	}
	defer rows.Close()

	var records []SampleRecord
	for rows.Next() {
		var record SampleRecord
		var sampledAtRaw, createdAtRaw sql.NullString
		err = rows.Scan(
			&record.ID,
			&record.TaskID,
			&record.SampledBy,
			&sampledAtRaw,
			&record.SampleComment,
			&record.SampleResult,
			&createdAtRaw,
		)
		if err != nil {
			logger.Errorf("扫描抽检记录失败: %v", err)
			continue
		}

		record.SampledAt = formatDateTime(sampledAtRaw.String)
		record.CreatedAt = formatDateTime(createdAtRaw.String)
		records = append(records, record)
	}

	if err = rows.Err(); err != nil {
		logger.Errorf("遍历抽检记录失败: %v", err)
		return nil, err
	}

	return records, nil
}

// BatchGetSampleInfo 批量获取多个任务的抽检信息（用于列表页优化）
func BatchGetSampleInfo(taskIDs []int) (map[int]*SampleInfo, error) {
	if len(taskIDs) == 0 {
		return make(map[int]*SampleInfo), nil
	}

	// 构建 IN 查询的占位符
	placeholders := ""
	args := make([]interface{}, len(taskIDs))
	for i, id := range taskIDs {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args[i] = id
	}

	// 查询任务表的抽检字段
	querySQL := `SELECT id, is_sampled, last_sampled_at 
		FROM audit_tasks 
		WHERE id IN (` + placeholders + `)`
	
	rows, err := db.DBInstance.Query(querySQL, args...)
	if err != nil {
		logger.Errorf("批量查询任务抽检信息失败: %v", err)
		return nil, err
	}
	defer rows.Close()

	result := make(map[int]*SampleInfo)
	for rows.Next() {
		var taskID int
		var isSampled int
		var lastSampledAt sql.NullString
		err = rows.Scan(&taskID, &isSampled, &lastSampledAt)
		if err != nil {
			logger.Errorf("扫描任务抽检信息失败: %v", err)
			continue
		}

		info := &SampleInfo{
			IsSampled: isSampled == 1,
		}

		if lastSampledAt.Valid && lastSampledAt.String != "" {
			info.LastSampledAt = formatDateTime(lastSampledAt.String)
		}

		result[taskID] = info
	}

	// 批量查询抽检记录获取抽检人员、次数和最近一次结果
	if len(result) > 0 {
		// 使用更简单的查询方式
		for taskID := range result {
			var sampledBy, sampleResult sql.NullString
			var count int
			lastSQL := `SELECT sampled_by, sample_result,
				(SELECT COUNT(*) FROM audit_sample_records WHERE task_id = ?) AS sample_count
				FROM audit_sample_records 
				WHERE task_id = ? 
				ORDER BY sampled_at DESC 
				LIMIT 1`
			err = db.DBInstance.QueryRow(lastSQL, taskID, taskID).Scan(&sampledBy, &sampleResult, &count)
			if err == nil {
				if sampledBy.Valid {
					result[taskID].SampledBy = sampledBy.String
				}
				if sampleResult.Valid {
					result[taskID].LastSampleResult = sampleResult.String
				}
				result[taskID].SampleCount = count
			}
		}
	}

	return result, nil
}

// 注意：formatDateTime 函数在 handler.go 中已定义，同一个包内可以直接使用

