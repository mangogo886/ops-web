package operationlog

import (
	"html/template"
	"net/http"
	"ops-web/internal/db"
	"ops-web/internal/logger"
	"time"
)

// PageData 页面数据
type PageData struct {
	Title      string
	ActiveMenu string
	SubMenu    string
	Logs       []LogEntry
}

// Handler 日志列表
func Handler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DBInstance.Query(`SELECT id, username, action, ip, created_at FROM operation_logs ORDER BY id DESC LIMIT 200`)
	if err != nil {
		logger.Errorf("操作日志-查询失败: %v", err)
		http.Error(w, "查询操作日志失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var logs []LogEntry
	for rows.Next() {
		var item LogEntry
		var createdAtRaw string
		if err := rows.Scan(&item.ID, &item.Username, &item.Action, &item.IP, &createdAtRaw); err != nil {
			continue
		}
		item.CreatedAt = formatDateTime(createdAtRaw)
		logs = append(logs, item)
	}

	data := PageData{
		Title:      "操作日志",
		ActiveMenu: "settings",
		SubMenu:    "logs",
		Logs:       logs,
	}

	tmpl, err := template.ParseFiles("templates/logs.html")
	if err != nil {
		logger.Errorf("操作日志-模板解析失败: %v", err)
		http.Error(w, "模板解析失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_ = tmpl.Execute(w, data)
}

// formatDateTime 格式化时间字符串为 YYYY-MM-DD HH:mm 格式
// 支持多种输入格式：RFC3339 (2006-01-02T15:04:05Z07:00), MySQL datetime (2006-01-02 15:04:05) 等
func formatDateTime(timeStr string) string {
	if timeStr == "" {
		return ""
	}
	
	// 尝试解析多种时间格式
	formats := []string{
		"2006-01-02T15:04:05Z07:00", // RFC3339 with timezone
		"2006-01-02T15:04:05",       // RFC3339 without timezone
		"2006-01-02 15:04:05",       // MySQL datetime
		"2006-01-02 15:04",          // Already formatted
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t.Format("2006-01-02 15:04")
		}
	}
	
	// 如果所有格式都解析失败，尝试截取前16个字符（YYYY-MM-DD HH:mm）
	if len(timeStr) >= 16 {
		return timeStr[:16]
	}
	
	return timeStr
}


