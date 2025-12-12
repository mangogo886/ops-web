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

// 本周建档统计（改为按日期范围统计）
func WeeklyHandler(w http.ResponseWriter, r *http.Request) {
	// 获取URL参数中的日期
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	// 如果没有传入日期，默认为本周
	if startDateStr == "" || endDateStr == "" {
		now := time.Now()
		weekday := int(now.Weekday())
		if weekday == 0 { // 周日
			weekday = 7
		}
		weekStart := now.AddDate(0, 0, -(weekday - 1))
		weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())
		
		// 本周日 23:59:59
		weekEnd := weekStart.AddDate(0, 0, 6)
		weekEnd = time.Date(weekEnd.Year(), weekEnd.Month(), weekEnd.Day(), 23, 59, 59, 0, weekEnd.Location())
		
		startDateStr = weekStart.Format("2006-01-02")
		endDateStr = weekEnd.Format("2006-01-02")
	}

	// 解析日期
	startDate, err1 := time.Parse("2006-01-02", startDateStr)
	endDate, err2 := time.Parse("2006-01-02", endDateStr)
	
	// 如果日期解析失败，使用默认值
	if err1 != nil || err2 != nil {
		now := time.Now()
		startDate = now.AddDate(0, 0, -7) // 默认最近7天
		endDate = now
		startDateStr = startDate.Format("2006-01-02")
		endDateStr = endDate.Format("2006-01-02")
	}

	// 设置结束时间为当天的 23:59:59
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, endDate.Location())

	stats, summary := getStatisticsByDateRange(startDate, endDate)

	data := StatsPageData{
		Title:      "按日期统计",
		ActiveMenu: "stats",
		SubMenu:    "weekly",
		Stats:      stats,
		Summary:    summary,
		StartDate:  startDateStr,
		EndDate:    endDateStr,
	}

	renderTemplate(w, data)
}

// 建档全量统计
func TotalHandler(w http.ResponseWriter, r *http.Request) {
	stats, summary := getStatisticsByDateRange(time.Time{}, time.Time{}) // 传入零值时间表示不过滤

	data := StatsPageData{
		Title:      "建档全量统计",
		ActiveMenu: "stats",
		SubMenu:    "total",
		Stats:      stats,
		Summary:    summary,
		StartDate:  "",
		EndDate:    "",
	}

	renderTemplate(w, data)
}

// 按日期范围获取统计数据
func getStatisticsByDateRange(startTime, endTime time.Time) ([]StatRow, StatRow) {
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

	// 如果需要过滤时间（按日期范围统计）
	if !startTime.IsZero() && !endTime.IsZero() {
		// 使用 DATE() 函数提取日期部分进行比较
		query += " AND DATE(update_time) >= ? AND DATE(update_time) <= ?"
		args = append(args, startTime.Format("2006-01-02"))
		args = append(args, endTime.Format("2006-01-02"))
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