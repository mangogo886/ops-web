package checkpointprogress

import (
	"database/sql"
	"ops-web/internal/auth"
	"ops-web/internal/db"
	"ops-web/internal/logger"
)

// HistoryItem 审核意见历史记录项
type HistoryItem struct {
	ID           int
	AuditComment sql.NullString
	AuditStatus  string
	Auditor      string
	AuditTime    string
}

// SaveAuditHistory 保存审核意见历史记录
// 在事务中调用，如果审核意见或状态有变化，则保存历史记录
func SaveAuditHistory(tx *sql.Tx, taskID int, newComment string, newStatus string, currentUser *auth.User) error {
	// 获取当前审核意见和状态
	var currentComment sql.NullString
	var currentStatus string
	selectSQL := `SELECT audit_comment, audit_status FROM checkpoint_tasks WHERE id = ?`
	err := tx.QueryRow(selectSQL, taskID).Scan(&currentComment, &currentStatus)
	if err != nil {
		return err
	}

	// 判断是否需要保存历史记录（审核意见或状态有变化）
	currentCommentStr := ""
	if currentComment.Valid {
		currentCommentStr = currentComment.String
	}

	needSaveHistory := false
	if currentCommentStr != newComment {
		needSaveHistory = true
	}
	if currentStatus != newStatus {
		needSaveHistory = true
	}

	// 如果需要保存历史记录
	if needSaveHistory {
		auditor := "系统"
		if currentUser != nil {
			auditor = currentUser.Username
		}

		// 插入历史记录
		insertHistorySQL := `INSERT INTO checkpoint_audit_history (task_id, audit_comment, audit_status, auditor, audit_time) VALUES (?, ?, ?, ?, NOW())`
		var historyComment interface{}
		if currentComment.Valid && currentCommentStr != "" {
			historyComment = currentCommentStr
		} else {
			historyComment = nil
		}

		_, err = tx.Exec(insertHistorySQL, taskID, historyComment, currentStatus, auditor)
		if err != nil {
			logger.Errorf("保存卡口审核意见历史失败: %v, taskID: %d", err, taskID)
			return err
		}
	}

	return nil
}

// GetAuditHistory 查询审核意见历史记录列表
func GetAuditHistory(taskID int) ([]HistoryItem, error) {
	historySQL := `SELECT id, audit_comment, audit_status, auditor, audit_time 
		FROM checkpoint_audit_history 
		WHERE task_id = ? 
		ORDER BY audit_time DESC`

	rows, err := db.DBInstance.Query(historySQL, taskID)
	if err != nil {
		logger.Errorf("查询卡口审核意见历史失败: %v, taskID: %d", err, taskID)
		return nil, err
	}
	defer rows.Close()

	var historyList []HistoryItem
	for rows.Next() {
		var item HistoryItem
		var auditTimeRaw sql.NullString
		err = rows.Scan(
			&item.ID,
			&item.AuditComment,
			&item.AuditStatus,
			&item.Auditor,
			&auditTimeRaw,
		)
		if err != nil {
			logger.Errorf("扫描卡口审核意见历史记录失败: %v", err)
			continue
		}
		item.AuditTime = formatDateTime(auditTimeRaw.String)
		historyList = append(historyList, item)
	}

	if err = rows.Err(); err != nil {
		logger.Errorf("遍历卡口审核意见历史记录失败: %v", err)
		return nil, err
	}

	return historyList, nil
}



