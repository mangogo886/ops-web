package auditprogress

import (
	"database/sql"
	"fmt"
	"ops-web/internal/db"
	"ops-web/internal/logger"
	"regexp"
	"strings"
	"time"
)

// VideoReminder 录像提醒任务结构体
type VideoReminder struct {
	ID                int
	TaskID            int
	EarliestVideoDate time.Time
	RequiredDays      int
	ActualDays        int
	ReminderDate      time.Time
	Status            string
	CreatedAt         time.Time
	NotifiedAt        sql.NullTime
	CompletedAt       sql.NullTime
	CompletedBy       sql.NullString
	// 关联信息
	FileName     string
	Organization string
}

// ParseVideoDaysIssue 解析审核意见中的录像天数不足信息
// 返回：最早录像日期、要求天数、是否找到
func ParseVideoDaysIssue(comment string) (earliestDate time.Time, requiredDays int, found bool) {
	if comment == "" {
		return time.Time{}, 0, false
	}

	// 定义日期匹配模式（支持多种格式）
	datePatterns := []*regexp.Regexp{
		regexp.MustCompile(`录像最早日期[：:]\s*(\d{4}[-/]\d{1,2}[-/]\d{1,2})`),
		regexp.MustCompile(`最早录像日期[：:]\s*(\d{4}[-/]\d{1,2}[-/]\d{1,2})`),
		regexp.MustCompile(`录像日期[：:]\s*(\d{4}[-/]\d{1,2}[-/]\d{1,2})`),
		regexp.MustCompile(`(\d{4}[-/]\d{1,2}[-/]\d{1,2})`), // 通用日期格式（作为备选）
	}

	// 定义天数匹配模式
	daysPatterns := []*regexp.Regexp{
		regexp.MustCompile(`不足\s*(\d+)\s*天`),
		regexp.MustCompile(`录像天数不足\s*(\d+)\s*天`),
		regexp.MustCompile(`缺少\s*(\d+)\s*天`),
	}

	var dateStr string
	var daysStr string

	// 提取日期
	for _, pattern := range datePatterns {
		matches := pattern.FindStringSubmatch(comment)
		if len(matches) > 1 {
			dateStr = matches[1]
			break
		}
	}

	// 提取天数
	for _, pattern := range daysPatterns {
		matches := pattern.FindStringSubmatch(comment)
		if len(matches) > 1 {
			daysStr = matches[1]
			break
		}
	}

	// 如果没有找到天数，尝试匹配常见的30/90/180天
	if daysStr == "" {
		if strings.Contains(comment, "不足30天") || strings.Contains(comment, "30天") {
			daysStr = "30"
		} else if strings.Contains(comment, "不足90天") || strings.Contains(comment, "90天") {
			daysStr = "90"
		} else if strings.Contains(comment, "不足180天") || strings.Contains(comment, "180天") {
			daysStr = "180"
		}
	}

	// 如果日期和天数都找到了，解析日期
	if dateStr != "" && daysStr != "" {
		// 尝试多种日期格式
		dateFormats := []string{
			"2006-01-02",
			"2006/01/02",
			"2006-1-2",
			"2006/1/2",
		}

		for _, format := range dateFormats {
			if t, err := time.Parse(format, dateStr); err == nil {
				var days int
				fmt.Sscanf(daysStr, "%d", &days)
				// 验证天数是否为30、90或180
				if days == 30 || days == 90 || days == 180 {
					return t, days, true
				}
			}
		}
	}

	return time.Time{}, 0, false
}

// CreateVideoReminder 创建录像提醒任务
func CreateVideoReminder(tx *sql.Tx, taskID int, earliestDate time.Time, requiredDays int, auditDate time.Time) error {
	// 计算实际天数
	actualDays := int(auditDate.Sub(earliestDate).Hours() / 24)
	if actualDays < 0 {
		actualDays = 0
	}

	// 计算提醒日期（最早录像日期 + 要求天数）
	reminderDate := earliestDate.AddDate(0, 0, requiredDays)

	// 检查是否已存在相同的提醒任务（避免重复创建）
	var exists int
	checkSQL := `SELECT COUNT(*) FROM audit_video_reminders 
		WHERE task_id = ? AND earliest_video_date = ? AND required_days = ? AND status != 'completed'`
	err := tx.QueryRow(checkSQL, taskID, earliestDate.Format("2006-01-02"), requiredDays).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		logger.Errorf("检查录像提醒任务是否存在失败: %v, taskID: %d", err, taskID)
		return err
	}

	// 如果已存在，不重复创建
	if exists > 0 {
		logger.Errorf("录像提醒任务已存在，跳过创建: taskID=%d, earliestDate=%s, requiredDays=%d", 
			taskID, earliestDate.Format("2006-01-02"), requiredDays)
		return nil
	}

	// 插入提醒记录
	insertSQL := `INSERT INTO audit_video_reminders 
		(task_id, earliest_video_date, required_days, actual_days, reminder_date, status, created_at) 
		VALUES (?, ?, ?, ?, ?, 'pending', NOW())`
	
	_, err = tx.Exec(insertSQL, taskID, earliestDate.Format("2006-01-02"), requiredDays, actualDays, reminderDate.Format("2006-01-02"))
	if err != nil {
		logger.Errorf("创建录像提醒任务失败: %v, taskID: %d", err, taskID)
		return err
	}

	logger.Errorf("创建录像提醒任务成功: taskID=%d, earliestDate=%s, requiredDays=%d, reminderDate=%s", 
		taskID, earliestDate.Format("2006-01-02"), requiredDays, reminderDate.Format("2006-01-02"))
	
	return nil
}

// ProcessVideoReminders 处理到期的提醒任务（定时任务调用）
func ProcessVideoReminders() error {
	// 查询到期的提醒任务（reminder_date <= 今天，且状态为pending）
	today := time.Now().Format("2006-01-02")
	querySQL := `SELECT id, task_id FROM audit_video_reminders 
		WHERE reminder_date <= ? AND status = 'pending'`
	
	rows, err := db.DBInstance.Query(querySQL, today)
	if err != nil {
		logger.Errorf("查询到期提醒任务失败: %v", err)
		return err
	}
	defer rows.Close()

	var reminderIDs []int
	var taskIDs []int
	for rows.Next() {
		var id, taskID int
		if err := rows.Scan(&id, &taskID); err != nil {
			logger.Errorf("扫描提醒任务失败: %v", err)
			continue
		}
		reminderIDs = append(reminderIDs, id)
		taskIDs = append(taskIDs, taskID)
	}

	if len(reminderIDs) == 0 {
		return nil // 没有到期的任务
	}

	// 批量更新状态为"已通知"
	updateSQL := `UPDATE audit_video_reminders 
		SET status = 'notified', notified_at = NOW() 
		WHERE id IN (` + strings.Repeat("?,", len(reminderIDs)-1) + "?)"
	
	args := make([]interface{}, len(reminderIDs))
	for i, id := range reminderIDs {
		args[i] = id
	}

	_, err = db.DBInstance.Exec(updateSQL, args...)
	if err != nil {
		logger.Errorf("更新提醒任务状态失败: %v", err)
		return err
	}

	logger.Errorf("处理到期提醒任务成功: 共 %d 条", len(reminderIDs))
	return nil
}

// GetVideoReminders 获取提醒任务列表
func GetVideoReminders(status string, page, pageSize int) ([]VideoReminder, int, error) {
	whereSQL := " WHERE 1=1"
	args := []interface{}{}

	if status != "" {
		whereSQL += " AND vr.status = ?"
		args = append(args, status)
	}

	// 查询总数
	countSQL := `SELECT COUNT(*) FROM audit_video_reminders vr` + whereSQL
	var totalCount int
	err := db.DBInstance.QueryRow(countSQL, args...).Scan(&totalCount)
	if err != nil {
		logger.Errorf("查询提醒任务总数失败: %v", err)
		return nil, 0, err
	}

	// 查询列表
	querySQL := `SELECT vr.id, vr.task_id, vr.earliest_video_date, vr.required_days, 
		vr.actual_days, vr.reminder_date, vr.status, vr.created_at, 
		vr.notified_at, vr.completed_at, vr.completed_by,
		at.file_name, at.organization
		FROM audit_video_reminders vr
		LEFT JOIN audit_tasks at ON vr.task_id = at.id
		` + whereSQL + `
		ORDER BY vr.reminder_date ASC, vr.created_at DESC
		LIMIT ? OFFSET ?`
	
	offset := (page - 1) * pageSize
	args = append(args, pageSize, offset)
	rows, err := db.DBInstance.Query(querySQL, args...)
	if err != nil {
		logger.Errorf("查询提醒任务列表失败: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	var reminders []VideoReminder
	for rows.Next() {
		var vr VideoReminder
		var earliestDateStr, reminderDateStr, createdAtStr string
		var notifiedAt, completedAt sql.NullTime
		var completedBy sql.NullString

		err := rows.Scan(
			&vr.ID, &vr.TaskID, &earliestDateStr, &vr.RequiredDays,
			&vr.ActualDays, &reminderDateStr, &vr.Status, &createdAtStr,
			&notifiedAt, &completedAt, &completedBy,
			&vr.FileName, &vr.Organization,
		)
		if err != nil {
			logger.Errorf("扫描提醒任务失败: %v", err)
			continue
		}

		// 解析日期
		vr.EarliestVideoDate, _ = time.Parse("2006-01-02", earliestDateStr)
		vr.ReminderDate, _ = time.Parse("2006-01-02", reminderDateStr)
		vr.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		if notifiedAt.Valid {
			vr.NotifiedAt = notifiedAt
		}
		if completedAt.Valid {
			vr.CompletedAt = completedAt
		}
		if completedBy.Valid {
			vr.CompletedBy = completedBy
		}

		reminders = append(reminders, vr)
	}

	return reminders, totalCount, nil
}

// CompleteVideoReminder 标记提醒任务为已完成
func CompleteVideoReminder(reminderID int, username string) error {
	updateSQL := `UPDATE audit_video_reminders 
		SET status = 'completed', completed_at = NOW(), completed_by = ?
		WHERE id = ?`
	
	_, err := db.DBInstance.Exec(updateSQL, username, reminderID)
	if err != nil {
		logger.Errorf("标记提醒任务为已完成失败: %v, reminderID: %d", err, reminderID)
		return err
	}

	logger.Errorf("标记提醒任务为已完成成功: reminderID=%d, username=%s", reminderID, username)
	return nil
}

// ScheduleConfig 定时任务配置结构体
type ScheduleConfig struct {
	ID         int
	Frequency  string // daily-每天, weekly-每周
	Hour       int    // 执行时间（小时）：1-24
	DayOfWeek  sql.NullInt64 // 每周执行日期（1-7，1=周一，7=周日），仅当frequency=weekly时有效
	Enabled    bool
	UpdatedAt  time.Time
	UpdatedBy  sql.NullString
}

// GetScheduleConfig 获取定时任务配置
func GetScheduleConfig() (*ScheduleConfig, error) {
	querySQL := `SELECT id, frequency, hour, day_of_week, enabled, updated_at, updated_by 
		FROM audit_video_reminder_schedule_config 
		WHERE enabled = 1 
		ORDER BY id DESC 
		LIMIT 1`
	
	var config ScheduleConfig
	var updatedAtRaw string
	var dayOfWeek sql.NullInt64
	var updatedBy sql.NullString
	
	err := db.DBInstance.QueryRow(querySQL).Scan(
		&config.ID,
		&config.Frequency,
		&config.Hour,
		&dayOfWeek,
		&config.Enabled,
		&updatedAtRaw,
		&updatedBy,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			// 如果没有配置，返回默认配置
			return &ScheduleConfig{
				Frequency: "daily",
				Hour:      1,
				Enabled:   true,
			}, nil
		}
		return nil, err
	}
	
	config.DayOfWeek = dayOfWeek
	config.UpdatedBy = updatedBy
	if updatedAtRaw != "" {
		config.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtRaw)
	}
	
	return &config, nil
}

// SaveScheduleConfig 保存定时任务配置
func SaveScheduleConfig(frequency string, hour int, dayOfWeek int, enabled bool, username string) error {
	// 先删除旧配置（只保留一条）
	deleteSQL := `DELETE FROM audit_video_reminder_schedule_config`
	_, err := db.DBInstance.Exec(deleteSQL)
	if err != nil {
		logger.Errorf("删除旧定时配置失败: %v", err)
		// 不阻断流程，继续执行
	}
	
	// 插入新配置
	insertSQL := `INSERT INTO audit_video_reminder_schedule_config 
		(frequency, hour, day_of_week, enabled, updated_by) 
		VALUES (?, ?, ?, ?, ?)`
	
	var dayOfWeekVal interface{}
	if frequency == "weekly" && dayOfWeek > 0 {
		dayOfWeekVal = dayOfWeek
	} else {
		dayOfWeekVal = nil
	}
	
	_, err = db.DBInstance.Exec(insertSQL, frequency, hour, dayOfWeekVal, enabled, username)
	if err != nil {
		logger.Errorf("保存定时配置失败: %v", err)
		return err
	}
	
	return nil
}

// GetReminderCountByTaskID 获取指定任务的提醒数量
func GetReminderCountByTaskID(taskID int) (int, error) {
	var count int
	querySQL := `SELECT COUNT(*) FROM audit_video_reminders 
		WHERE task_id = ? AND status != 'completed'`
	err := db.DBInstance.QueryRow(querySQL, taskID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

