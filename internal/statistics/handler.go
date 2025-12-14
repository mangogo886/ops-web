package statistics

import (
	"fmt"
	"html/template"
	"net/http"
	"ops-web/internal/auth"
	"ops-web/internal/db"
	"ops-web/internal/logger"
	"ops-web/internal/operationlog"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
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
	Title       string
	ActiveMenu  string
	SubMenu     string
	Stats       []StatRow // 统计数据
	Summary     StatRow  // 汇总行
	StartDate   string   // 开始日期
	EndDate     string   // 结束日期
	AuditStatus string   // 建档状态查询条件
	Query       string   // 查询参数
}

// 主入口 - 统计信息页面
func Handler(w http.ResponseWriter, r *http.Request) {
	// 获取URL参数
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	auditStatus := r.URL.Query().Get("audit_status") // 建档状态: 0, 1, 2 或空

	// 解析日期（如果提供了日期参数）
	var startDate, endDate time.Time
	var hasDateFilter bool
	
	if startDateStr != "" && endDateStr != "" {
		var err1, err2 error
		startDate, err1 = time.Parse("2006-01-02", startDateStr)
		endDate, err2 = time.Parse("2006-01-02", endDateStr)
		
		// 如果日期解析成功，设置日期过滤
		if err1 == nil && err2 == nil {
			hasDateFilter = true
			// 设置结束时间为当天的 23:59:59
			endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, endDate.Location())
		} else {
			// 日期解析失败，清空日期参数，查询所有数据
			startDateStr = ""
			endDateStr = ""
		}
	}

	stats, summary := getStatisticsByDateRange(startDate, endDate, auditStatus, hasDateFilter)

	// 构建查询参数字符串
	queryParams := []string{}
	if startDateStr != "" {
		queryParams = append(queryParams, "start_date="+startDateStr)
	}
	if endDateStr != "" {
		queryParams = append(queryParams, "end_date="+endDateStr)
	}
	if auditStatus != "" {
		queryParams = append(queryParams, "audit_status="+auditStatus)
	}
	query := strings.Join(queryParams, "&")

	data := StatsPageData{
		Title:       "统计信息",
		ActiveMenu:  "stats",
		SubMenu:     "",
		Stats:       stats,
		Summary:     summary,
		StartDate:   startDateStr,
		EndDate:     endDateStr,
		AuditStatus: auditStatus,
		Query:       query,
	}

	// 记录查询操作日志
	currentUser := auth.GetCurrentUser(r)
	if currentUser != nil && (startDateStr != "" || endDateStr != "" || auditStatus != "") {
		action := fmt.Sprintf("查询统计信息（开始日期：%s，结束日期：%s", startDateStr, endDateStr)
		if auditStatus != "" {
			statusText := map[string]string{"0": "未审核未建档", "1": "已审核未建档", "2": "已建档"}
			action += fmt.Sprintf("，建档状态：%s", statusText[auditStatus])
		}
		action += "）"
		operationlog.Record(r, currentUser.Username, action)
	}

	renderTemplate(w, data)
}


// 按日期范围获取统计数据（从audit_details表读取）
func getStatisticsByDateRange(startTime, endTime time.Time, auditStatus string, hasDateFilter bool) ([]StatRow, StatRow) {
	// 构建 SQL 查询
	query := `
		SELECT 
			management_unit,
			monitor_point_type,
			COUNT(*) as count
		FROM audit_details
		WHERE 1=1
	`

	args := []interface{}{}

	// 如果需要过滤时间（按日期范围统计）
	if hasDateFilter && !startTime.IsZero() && !endTime.IsZero() {
		// 使用 DATE() 函数提取日期部分进行比较
		query += " AND DATE(update_time) >= ? AND DATE(update_time) <= ?"
		args = append(args, startTime.Format("2006-01-02"))
		args = append(args, endTime.Format("2006-01-02"))
	}

	// 建档状态查询
	if auditStatus != "" {
		statusInt, err := strconv.Atoi(auditStatus)
		if err == nil && (statusInt == 0 || statusInt == 1 || statusInt == 2) {
			query += " AND audit_status = ?"
			args = append(args, statusInt)
		}
	}

	query += " GROUP BY management_unit, monitor_point_type ORDER BY management_unit, monitor_point_type"

	rows, err := db.DBInstance.Query(query, args...)
	if err != nil {
		logger.Errorf("统计信息-数据库查询失败: %v, SQL: %s, Args: %v", err, query, args)
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
		var pointTypeStr string // 改为字符串接收
		var count int

		if err := rows.Scan(&unit, &pointTypeStr, &count); err != nil {
			continue
		}

		// 解析监控点位类型（varchar类型需要转换为int）
		pointTypeStr = strings.TrimSpace(pointTypeStr)
		pointType, err := strconv.Atoi(pointTypeStr)
		if err != nil {
			continue // 跳过无法转换的类型
		}
		
		// 只统计1-4类，其他类型跳过
		if pointType < 1 || pointType > 4 {
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

// ExportHandler: 导出统计数据到Excel
func ExportHandler(w http.ResponseWriter, r *http.Request) {
	// 获取查询参数（必须与Handler中的查询条件保持一致）
	// 先尝试从 URL Query 获取
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	auditStatus := r.URL.Query().Get("audit_status")

	// 如果从 Query 获取不到，尝试从 RawQuery 手动解析（处理 URL 编码问题）
	if startDateStr == "" && endDateStr == "" && auditStatus == "" && r.URL.RawQuery != "" {
		// 手动解析 RawQuery
		rawQuery := r.URL.RawQuery
		// 处理 URL 编码的 = 号（%3d）
		rawQuery = strings.ReplaceAll(rawQuery, "%3d", "=")
		rawQuery = strings.ReplaceAll(rawQuery, "%3D", "=")
		
		// 解析参数
		parts := strings.Split(rawQuery, "&")
		for _, part := range parts {
			if strings.Contains(part, "=") {
				kv := strings.SplitN(part, "=", 2)
				if len(kv) == 2 {
					key := strings.TrimSpace(kv[0])
					value := strings.TrimSpace(kv[1])
					if key == "start_date" {
						startDateStr = value
					} else if key == "end_date" {
						endDateStr = value
					} else if key == "audit_status" {
						auditStatus = value
					}
				}
			}
		}
	}

	// 解析日期（如果提供了日期参数）
	var startDate, endDate time.Time
	var hasDateFilter bool
	
	if startDateStr != "" && endDateStr != "" {
		var err1, err2 error
		startDate, err1 = time.Parse("2006-01-02", startDateStr)
		endDate, err2 = time.Parse("2006-01-02", endDateStr)
		
		// 如果日期解析成功，设置日期过滤
		if err1 == nil && err2 == nil {
			hasDateFilter = true
			// 设置结束时间为当天的 23:59:59
			endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, endDate.Location())
		}
	}

	// 获取统计数据（应用筛选条件）
	stats, summary := getStatisticsByDateRange(startDate, endDate, auditStatus, hasDateFilter)

	// 创建Excel文件
	f := excelize.NewFile()
	sheetName := "统计信息"
	f.SetSheetName("Sheet1", sheetName)

	// 设置表头
	headers := []interface{}{
		"分局", "一类点", "二类点", "三类点", "四类点", "合计",
	}
	f.SetSheetRow(sheetName, "A1", &headers)

	// 设置表头样式
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#3498db"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	f.SetCellStyle(sheetName, "A1", "F1", headerStyle)

	// 写入数据行
	rowNum := 2
	for _, stat := range stats {
		rowData := []interface{}{
			stat.ManagementUnit,
			stat.Type1Count,
			stat.Type2Count,
			stat.Type3Count,
			stat.Type4Count,
			stat.TotalCount,
		}
		cellName, _ := excelize.CoordinatesToCellName(1, rowNum)
		f.SetSheetRow(sheetName, cellName, &rowData)
		rowNum++
	}

	// 写入汇总行
	summaryRow := []interface{}{
		summary.ManagementUnit,
		summary.Type1Count,
		summary.Type2Count,
		summary.Type3Count,
		summary.Type4Count,
		summary.TotalCount,
	}
	cellName, _ := excelize.CoordinatesToCellName(1, rowNum)
	f.SetSheetRow(sheetName, cellName, &summaryRow)

	// 设置汇总行样式
	summaryStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#ecf0f1"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	lastRow := rowNum
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", lastRow), fmt.Sprintf("F%d", lastRow), summaryStyle)

	// 设置列宽
	f.SetColWidth(sheetName, "A", "A", 15) // 分局列
	f.SetColWidth(sheetName, "B", "F", 12) // 数据列

	// 记录导出操作日志
	currentUser := auth.GetCurrentUser(r)
	if currentUser != nil {
		action := "导出统计信息 Excel"
		if startDateStr != "" || endDateStr != "" || auditStatus != "" {
			action += "（"
			if startDateStr != "" && endDateStr != "" {
				action += fmt.Sprintf("日期：%s 至 %s", startDateStr, endDateStr)
			}
			if auditStatus != "" {
				if startDateStr != "" {
					action += "，"
				}
				statusText := map[string]string{"0": "未审核未建档", "1": "已审核未建档", "2": "已建档"}
				action += fmt.Sprintf("建档状态：%s", statusText[auditStatus])
			}
			action += "）"
		}
		operationlog.Record(r, currentUser.Username, action)
	}

	// 输出文件
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	filename := fmt.Sprintf("attachment; filename=\"统计信息_%s.xlsx\"", time.Now().Format("20060102_150405"))
	w.Header().Set("Content-Disposition", filename)
	f.Write(w)
}

// 渲染模板
func renderTemplate(w http.ResponseWriter, data StatsPageData) {
	tmpl, err := template.ParseFiles("templates/statistics.html")
	if err != nil {
		logger.Errorf("统计信息-模板解析失败: %v", err)
		http.Error(w, "模板解析失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		logger.Errorf("统计信息-模板渲染失败: %v", err)
		http.Error(w, "模板渲染失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
}