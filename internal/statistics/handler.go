package statistics

import (
	"html/template"
	"net/http"
	"ops-web/internal/db"
	"time"
)

// 统计数据结构
type StatRow struct {
	ManagementUnit string // 分局名称
	Type1Count     int    // 一类视频数量
	Type2Count     int    // 二类视频数量
	Type3Count     int    // 三类视频数量
	Type4Count     int    // 四类视频数量
	TotalCount     int    // 该分局总数
}

// 页面数据结构
type StatsPageData struct {
	Title      string
	ActiveMenu string
	SubMenu    string    // weekly 或 total
	Stats      []StatRow // 统计数据
	Summary    StatRow   // 汇总行
	StartDate  string    // 开始日期
	EndDate    string    // 结束日期
}

// 主入口 - 重定向到本周统计
func Handler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/stats/weekly", http.StatusFound)
}

// 本周建档统计
func WeeklyHandler(w http.ResponseWriter, r *http.Request) {
	// 计算本周的开始时间（周一 00:00:00）
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 { // 周日
		weekday = 7
	}
	weekStart := now.AddDate(0, 0, -(weekday - 1))
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())

	stats, summary := getStatistics(weekStart)

	data := StatsPageData{
		Title:      "本周建档统计",
		ActiveMenu: "stats",
		SubMenu:    "weekly",
		Stats:      stats,
		Summary:    summary,
	}

	renderTemplate(w, data)
}

// 建档全量统计
func TotalHandler(w http.ResponseWriter, r *http.Request) {
	stats, summary := getStatistics(time.Time{}) // 传入零值时间表示不过滤

	data := StatsPageData{
		Title:      "建档全量统计",
		ActiveMenu: "stats",
		SubMenu:    "total",
		Stats:      stats,
		Summary:    summary,
	}

	renderTemplate(w, data)
}

// 获取统计数据的核心函数
func getStatistics(startTime time.Time) ([]StatRow, StatRow) {
	// 构建 SQL 查询
	query := `
		SELECT 
			management_unit,
			monitor_point_type,
			COUNT(*) as count
		FROM fileList
		WHERE 1=1
	`

	args := []interface{}{}

	// 如果需要过滤时间（本周统计）
	if !startTime.IsZero() {
		query += " AND update_time >= ?"
		args = append(args, startTime.Format("2006-01-02 15:04:05"))
	}

	query += " GROUP BY management_unit, monitor_point_type ORDER BY management_unit, monitor_point_type"

	rows, err := db.DBInstance.Query(query, args...)
	if err != nil {
		// 返回空数据
		return []StatRow{}, StatRow{ManagementUnit: "汇总"}
	}
	defer rows.Close()

	// 使用 map 来组织数据
	statsMap := make(map[string]*StatRow)

	// 汇总数据
	summary := StatRow{
		ManagementUnit: "汇总",
	}

	for rows.Next() {
		var unit string
		var pointType int
		var count int

		if err := rows.Scan(&unit, &pointType, &count); err != nil {
			continue
		}

		// 初始化该分局的数据
		if _, exists := statsMap[unit]; !exists {
			statsMap[unit] = &StatRow{
				ManagementUnit: unit,
			}
		}

		// 根据类型填充数据
		switch pointType {
		case 1:
			statsMap[unit].Type1Count = count
			summary.Type1Count += count
		case 2:
			statsMap[unit].Type2Count = count
			summary.Type2Count += count
		case 3:
			statsMap[unit].Type3Count = count
			summary.Type3Count += count
		case 4:
			statsMap[unit].Type4Count = count
			summary.Type4Count += count
		}

		statsMap[unit].TotalCount += count
		summary.TotalCount += count
	}

	// 将 map 转换为切片
	stats := make([]StatRow, 0, len(statsMap))
	for _, stat := range statsMap {
		stats = append(stats, *stat)
	}

	return stats, summary
}

// 渲染模板
func renderTemplate(w http.ResponseWriter, data StatsPageData) {
	tmpl, err := template.ParseFiles("templates/statistics.html")
	if err != nil {
		http.Error(w, "模板解析失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "模板渲染失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
}