package auditstatistics

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"ops-web/internal/auth"
	"ops-web/internal/db"
	"ops-web/internal/logger"
	"ops-web/internal/operationlog"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// 统计数据行结构
type StatRow struct {
	ManagementUnit string // 分局名称
	// 一类视频
	Type1Video   int // 一类视频-视频
	Type1Face    int // 一类视频-人脸
	Type1Vehicle int // 一类视频-车辆
	Type1Total   int // 一类视频-汇总
	// 二类视频
	Type2Video   int // 二类视频-视频
	Type2Face    int // 二类视频-人脸
	Type2Vehicle int // 二类视频-车辆
	Type2Total   int // 二类视频-汇总
	// 三类视频
	Type3Video   int // 三类视频-视频
	Type3Face    int // 三类视频-人脸
	Type3Vehicle int // 三类视频-车辆
	Type3Total   int // 三类视频-汇总
	// 四类视频
	Type4Video   int // 四类视频-视频
	Type4Face    int // 四类视频-人脸
	Type4Vehicle int // 四类视频-车辆
	Type4Total   int // 四类视频-汇总
	// 单兵设备
	SingleSoldier int // 单兵设备-总数
	// 总计
	TotalVideo   int // 总计-视频
	TotalFace    int // 总计-人脸
	TotalVehicle int // 总计-车辆
	GrandTotal   int // 总计-汇总
}

// 页面数据结构
type PageData struct {
	Title         string
	ActiveMenu    string
	SubMenu       string
	Stats         []StatRow
	Summary       StatRow
	Month         string // 月份查询条件 (格式: 2024-01)
	AuditStatus   string // 建档状态查询条件
	Query         string // 查询参数
}

// Handler: 月度建档数据页面
func Handler(w http.ResponseWriter, r *http.Request) {
	// 获取查询参数
	month := r.URL.Query().Get("month")        // 月份，格式: 2024-01
	auditStatus := r.URL.Query().Get("audit_status") // 建档状态: 0, 1, 2 或空

	// 查询统计数据（直接传递参数，在getStatistics内部构建带表别名的whereSQL）
	stats, summary := getStatistics(month, auditStatus)

	// 构建查询参数字符串（用于表单回显）
	queryParams := []string{}
	if month != "" {
		queryParams = append(queryParams, "month="+month)
	}
	if auditStatus != "" {
		queryParams = append(queryParams, "audit_status="+auditStatus)
	}
	query := strings.Join(queryParams, "&")

	data := PageData{
		Title:       "月度建档数据",
		ActiveMenu:  "audit",
		SubMenu:     "audit_statistics",
		Stats:       stats,
		Summary:     summary,
		Month:       month,
		AuditStatus: auditStatus,
		Query:       query,
	}

	// 记录查询操作日志
	currentUser := auth.GetCurrentUser(r)
	if currentUser != nil && (month != "" || auditStatus != "") {
		action := fmt.Sprintf("查询月度建档数据（月份：%s，建档状态：%s）", month, auditStatus)
		operationlog.Record(r, currentUser.Username, action)
	}

	// 渲染模板
	tmpl, err := template.ParseFiles("templates/auditstatistics.html")
	if err != nil {
		logger.Errorf("月度建档数据-模板解析失败: %v", err)
		http.Error(w, "模板解析失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		logger.Errorf("月度建档数据-模板渲染失败: %v", err)
		http.Error(w, "模板渲染失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// buildWhereSQL: 构建带表别名的WHERE条件
// taskPrefix: 任务表别名（用于completed_at和import_time字段，如"at"或"ct"）
// detailPrefix: 明细表别名（用于audit_status字段，如"ad"或"cd"）
func buildWhereSQL(month, auditStatus string, taskPrefix, detailPrefix string) (string, []interface{}) {
	whereSQL := " WHERE 1=1"
	args := []interface{}{}

	// 月份查询：根据时间字段按月份
	// 如果建档状态为"已建档"（值为"2"），使用completed_at字段（在任务表）；否则使用import_time字段（在任务表）
	if month != "" {
		// 验证月份格式
		if _, err := time.Parse("2006-01", month); err == nil {
			parts := strings.Split(month, "-")
			if len(parts) == 2 {
				year, _ := strconv.Atoi(parts[0])
				monthNum, _ := strconv.Atoi(parts[1])
				// 如果建档状态为"已建档"（值为"2"），使用completed_at字段；否则使用import_time字段
				if auditStatus == "2" {
					whereSQL += " AND YEAR(" + taskPrefix + ".completed_at) = ? AND MONTH(" + taskPrefix + ".completed_at) = ?"
				} else {
					whereSQL += " AND YEAR(" + taskPrefix + ".import_time) = ? AND MONTH(" + taskPrefix + ".import_time) = ?"
				}
				args = append(args, year, monthNum)
			}
		}
	}

	// 建档状态查询
	if auditStatus != "" {
		statusInt, err := strconv.Atoi(auditStatus)
		if err == nil && (statusInt == 0 || statusInt == 1 || statusInt == 2) {
			whereSQL += " AND " + detailPrefix + ".audit_status = ?"
			args = append(args, statusInt)
		}
	}

	return whereSQL, args
}

// getStatistics: 获取统计数据
func getStatistics(month, auditStatus string) ([]StatRow, StatRow) {
	// 使用 map 来组织数据，key 是 organization
	statsMap := make(map[string]*StatRow)

	// 汇总数据
	summary := StatRow{
		ManagementUnit: "汇总",
	}

	// 1. 从 audit_details 表统计视频、人脸和车辆（只统计档案类型为"新增"的）
	// 构建设备审核的WHERE条件：completed_at在任务表（at），update_time和audit_status在明细表（ad）
	auditWhereSQL, auditArgs := buildWhereSQL(month, auditStatus, "at", "ad")
	
	query := `
		SELECT 
			at.organization,
			ad.monitor_point_type,
			ad.camera_function_type,
			at.is_single_soldier
		FROM audit_details ad
		INNER JOIN audit_tasks at ON ad.task_id = at.id
		` + auditWhereSQL + `
		AND at.archive_type = '新增'
		ORDER BY at.organization, ad.monitor_point_type
	`

	rows, err := db.DBInstance.Query(query, auditArgs...)
	if err != nil {
		logger.Errorf("月度建档数据-数据库查询失败: %v, SQL: %s, Args: %v", err, query, auditArgs)
		return []StatRow{}, StatRow{ManagementUnit: "汇总"}
	}

	for rows.Next() {
		var unit string
		var pointTypeStr string
		var functionType sql.NullString
		var isSingleSoldier int

		err := rows.Scan(&unit, &pointTypeStr, &functionType, &isSingleSoldier)
		if err != nil {
			continue
		}

		// 初始化该分局的数据
		if _, exists := statsMap[unit]; !exists {
			statsMap[unit] = &StatRow{
				ManagementUnit: unit,
			}
		}

		// 如果是单兵设备，只统计到单兵设备列，不统计到其他分类
		if isSingleSoldier == 1 {
			statsMap[unit].SingleSoldier++
			summary.SingleSoldier++
			summary.GrandTotal++
			continue
		}

		// 解析监控点位类型（去除空格后转换）
		pointTypeStr = strings.TrimSpace(pointTypeStr)
		pointType, err := strconv.Atoi(pointTypeStr)
		if err != nil {
			continue // 跳过无法转换的类型
		}
		
		// 只统计1-4类，其他类型跳过
		if pointType < 1 || pointType > 4 {
			continue
		}

		// 解析camera_function_type，判断是视频、人脸还是车辆
		var isFace, isVideo, isVehicle bool
		if functionType.Valid && functionType.String != "" {
			// 分割逗号分隔的值
			parts := strings.Split(functionType.String, ",")
			has1 := false
			has2 := false

			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part == "1" {
					has1 = true
				}
				if part == "2" {
					has2 = true
				}
			}

			// 判断逻辑：
			// - 如果包含 "1"（例如 "1" 或 "1,2" 或 "1,3" 等）→ 只统计为车辆
			// - 如果不包含 "1" 但包含 "2"（例如 "2" 或 "2,3" 等）→ 统计为人脸
			// - 其他情况（不包含 "1" 也不包含 "2"）→ 统计为视频
			if has1 {
				// 包含 "1"，只统计为车辆
				isVehicle = true
			} else if has2 {
				// 不包含 "1" 但包含 "2"，统计为人脸
				isFace = true
			} else {
				// 其他情况，统计为视频
				isVideo = true
			}
		} else {
			// 如果camera_function_type为空或NULL，统计为视频
			isVideo = true
		}

		// 根据监控点位类型和功能类型更新统计数据（统计视频、人脸和车辆）
		switch pointType {
		case 1: // 一类点
			if isVideo {
				statsMap[unit].Type1Video++
				summary.Type1Video++
				summary.TotalVideo++
			}
			if isFace {
				statsMap[unit].Type1Face++
				summary.Type1Face++
				summary.TotalFace++
			}
			if isVehicle {
				statsMap[unit].Type1Vehicle++
				summary.Type1Vehicle++
				summary.TotalVehicle++
			}
		case 2: // 二类点
			if isVideo {
				statsMap[unit].Type2Video++
				summary.Type2Video++
				summary.TotalVideo++
			}
			if isFace {
				statsMap[unit].Type2Face++
				summary.Type2Face++
				summary.TotalFace++
			}
			if isVehicle {
				statsMap[unit].Type2Vehicle++
				summary.Type2Vehicle++
				summary.TotalVehicle++
			}
		case 3: // 三类点
			if isVideo {
				statsMap[unit].Type3Video++
				summary.Type3Video++
				summary.TotalVideo++
			}
			if isFace {
				statsMap[unit].Type3Face++
				summary.Type3Face++
				summary.TotalFace++
			}
			if isVehicle {
				statsMap[unit].Type3Vehicle++
				summary.Type3Vehicle++
				summary.TotalVehicle++
			}
		case 4: // 四类点（内部监控）
			if isVideo {
				statsMap[unit].Type4Video++
				summary.Type4Video++
				summary.TotalVideo++
			}
			if isFace {
				statsMap[unit].Type4Face++
				summary.Type4Face++
				summary.TotalFace++
			}
			// 四类点（内部监控）不统计车辆
		}
	}
	rows.Close()

	// 2. 从 checkpoint_details 表统计车辆（只统计档案类型为"新增"的）
	// 构建卡口的WHERE条件：completed_at在任务表（ct），update_time在明细表（cd）
	checkpointWhereSQL, checkpointArgs := buildWhereSQL(month, auditStatus, "ct", "cd")
	
	checkpointQuery := `
		SELECT 
			ct.organization,
			cd.checkpoint_point_type
		FROM checkpoint_details cd
		INNER JOIN checkpoint_tasks ct ON cd.task_id = ct.id
		` + checkpointWhereSQL + `
		AND ct.archive_type = '新增'
		ORDER BY ct.organization, cd.checkpoint_point_type
	`

	checkpointRows, err := db.DBInstance.Query(checkpointQuery, checkpointArgs...)
	if err != nil {
		logger.Errorf("月度建档数据-卡口数据查询失败: %v, SQL: %s, Args: %v", err, checkpointQuery, checkpointArgs)
		// 继续处理，不中断
	} else {
		for checkpointRows.Next() {
			var unit string
			var pointTypeStr sql.NullString

			err := checkpointRows.Scan(&unit, &pointTypeStr)
			if err != nil {
				continue
			}

			// 解析卡口点位类型
			if !pointTypeStr.Valid || pointTypeStr.String == "" {
				continue
			}

			pointTypeStrVal := strings.TrimSpace(pointTypeStr.String)
			pointType, err := strconv.Atoi(pointTypeStrVal)
			if err != nil {
				continue // 跳过无法转换的类型
			}
			
			// 只统计1-4类，其他类型跳过
			if pointType < 1 || pointType > 4 {
				continue
			}

			// 初始化该分局的数据（如果不存在）
			if _, exists := statsMap[unit]; !exists {
				statsMap[unit] = &StatRow{
					ManagementUnit: unit,
				}
			}

			// 根据卡口点位类型更新车辆统计数据（四类点不统计车辆）
			switch pointType {
			case 1: // 一类视频车辆
				statsMap[unit].Type1Vehicle++
				summary.Type1Vehicle++
				summary.TotalVehicle++
			case 2: // 二类视频车辆
				statsMap[unit].Type2Vehicle++
				summary.Type2Vehicle++
				summary.TotalVehicle++
			case 3: // 三类视频车辆
				statsMap[unit].Type3Vehicle++
				summary.Type3Vehicle++
				summary.TotalVehicle++
			case 4: // 内部监控不统计车辆
				// 四类点（内部监控）不统计车辆，跳过
			}
		}
		checkpointRows.Close()
	}

	// 计算每个分局的汇总（四类点不包含车辆）
	for _, stat := range statsMap {
		stat.Type1Total = stat.Type1Video + stat.Type1Face + stat.Type1Vehicle
		stat.Type2Total = stat.Type2Video + stat.Type2Face + stat.Type2Vehicle
		stat.Type3Total = stat.Type3Video + stat.Type3Face + stat.Type3Vehicle
		stat.Type4Total = stat.Type4Video + stat.Type4Face // 内部监控不包含车辆
		stat.TotalVideo = stat.Type1Video + stat.Type2Video + stat.Type3Video + stat.Type4Video
		stat.TotalFace = stat.Type1Face + stat.Type2Face + stat.Type3Face + stat.Type4Face
		stat.TotalVehicle = stat.Type1Vehicle + stat.Type2Vehicle + stat.Type3Vehicle // 内部监控不统计车辆
		stat.GrandTotal = stat.Type1Total + stat.Type2Total + stat.Type3Total + stat.Type4Total + stat.SingleSoldier
	}

	// 计算汇总行的总计（四类点不包含车辆）
	summary.Type1Total = summary.Type1Video + summary.Type1Face + summary.Type1Vehicle
	summary.Type2Total = summary.Type2Video + summary.Type2Face + summary.Type2Vehicle
	summary.Type3Total = summary.Type3Video + summary.Type3Face + summary.Type3Vehicle
	summary.Type4Total = summary.Type4Video + summary.Type4Face // 内部监控不包含车辆
	summary.TotalVideo = summary.Type1Video + summary.Type2Video + summary.Type3Video + summary.Type4Video
	summary.TotalFace = summary.Type1Face + summary.Type2Face + summary.Type3Face + summary.Type4Face
	summary.TotalVehicle = summary.Type1Vehicle + summary.Type2Vehicle + summary.Type3Vehicle // 内部监控不统计车辆
	summary.GrandTotal = summary.Type1Total + summary.Type2Total + summary.Type3Total + summary.Type4Total + summary.SingleSoldier

	// 将 map 转换为切片
	stats := make([]StatRow, 0, len(statsMap))
	for _, stat := range statsMap {
		stats = append(stats, *stat)
	}

	// 按分局名称排序，确保每次查询结果顺序一致
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].ManagementUnit < stats[j].ManagementUnit
	})

	return stats, summary
}

// ExportHandler: 导出统计数据到Excel
func ExportHandler(w http.ResponseWriter, r *http.Request) {
	// 获取查询参数（必须与Handler中的查询条件保持一致）
	// 先尝试从 URL Query 获取
	month := r.URL.Query().Get("month")
	auditStatus := r.URL.Query().Get("audit_status")

	// 如果从 Query 获取不到参数，但 RawQuery 不为空，尝试从 RawQuery 手动解析（处理 URL 编码问题）
	if r.URL.RawQuery != "" && (month == "" || auditStatus == "") {
		// 先解码URL编码（包括%26=&, %3d==等）
		decodedQuery, err := url.QueryUnescape(r.URL.RawQuery)
		if err != nil {
			// 如果解码失败，尝试手动处理常见的编码
			decodedQuery = strings.ReplaceAll(r.URL.RawQuery, "%26", "&")
			decodedQuery = strings.ReplaceAll(decodedQuery, "%3d", "=")
			decodedQuery = strings.ReplaceAll(decodedQuery, "%3D", "=")
		}
		
		// 解析参数（只填充空值）
		parts := strings.Split(decodedQuery, "&")
		for _, part := range parts {
			if strings.Contains(part, "=") {
				kv := strings.SplitN(part, "=", 2)
				if len(kv) == 2 {
					key := strings.TrimSpace(kv[0])
					value := strings.TrimSpace(kv[1])
					if key == "month" && month == "" {
						month = value
					} else if key == "audit_status" && auditStatus == "" {
						auditStatus = value
					}
				}
			}
		}
	}

	// 获取统计数据（应用筛选条件，与Handler中的逻辑完全一致）
	stats, summary := getStatistics(month, auditStatus)

	// 创建Excel文件
	f := excelize.NewFile()
	sheetName := "月度建档数据"
	f.SetSheetName("Sheet1", sheetName)

	// 设置表头（第一行：主表头）
	headers1 := []interface{}{
		"分局", "一类点", "", "", "", "二类点", "", "", "",
		"三类点", "", "", "", "内部监控", "", "",
		"单兵设备", "汇总", "", "", "",
	}
	f.SetSheetRow(sheetName, "A1", &headers1)

	// 合并第一行的单元格
	f.MergeCell(sheetName, "A1", "A2") // 分局
	f.MergeCell(sheetName, "B1", "E1") // 一类点
	f.MergeCell(sheetName, "F1", "I1") // 二类点
	f.MergeCell(sheetName, "J1", "M1") // 三类点
	f.MergeCell(sheetName, "N1", "P1") // 内部监控（3列：视频、人脸、小计）
	f.MergeCell(sheetName, "Q1", "Q1") // 单兵设备（1列）
	f.MergeCell(sheetName, "R1", "U1") // 汇总

	// 设置表头（第二行：子表头）
	headers2 := []interface{}{
		"分局", "视频", "人脸", "车辆", "小计",
		"视频", "人脸", "车辆", "小计",
		"视频", "人脸", "车辆", "小计",
		"视频", "人脸", "小计",
		"总数",
		"视频", "人脸", "车辆", "总计",
	}
	f.SetSheetRow(sheetName, "A2", &headers2)

	// 设置表头样式
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#3498db"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	f.SetCellStyle(sheetName, "A1", "U2", headerStyle) // 增加到U（因为添加了单兵设备列）

	// 写入数据行
	rowNum := 3
	for _, stat := range stats {
		rowData := []interface{}{
			stat.ManagementUnit,
			stat.Type1Video, stat.Type1Face, stat.Type1Vehicle, stat.Type1Total,
			stat.Type2Video, stat.Type2Face, stat.Type2Vehicle, stat.Type2Total,
			stat.Type3Video, stat.Type3Face, stat.Type3Vehicle, stat.Type3Total,
			stat.Type4Video, stat.Type4Face, stat.Type4Total, // 内部监控不包含车辆
			stat.SingleSoldier, // 单兵设备
			stat.TotalVideo, stat.TotalFace, stat.TotalVehicle, stat.GrandTotal,
		}
		cellName, _ := excelize.CoordinatesToCellName(1, rowNum)
		f.SetSheetRow(sheetName, cellName, &rowData)
		rowNum++
	}

	// 写入汇总行
	summaryRow := []interface{}{
		summary.ManagementUnit,
		summary.Type1Video, summary.Type1Face, summary.Type1Vehicle, summary.Type1Total,
		summary.Type2Video, summary.Type2Face, summary.Type2Vehicle, summary.Type2Total,
		summary.Type3Video, summary.Type3Face, summary.Type3Vehicle, summary.Type3Total,
		summary.Type4Video, summary.Type4Face, summary.Type4Total, // 内部监控不包含车辆
		summary.SingleSoldier, // 单兵设备
		summary.TotalVideo, summary.TotalFace, summary.TotalVehicle, summary.GrandTotal,
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
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", lastRow), fmt.Sprintf("U%d", lastRow), summaryStyle) // 增加到U

	// 设置列宽
	f.SetColWidth(sheetName, "A", "A", 15) // 分局列
	f.SetColWidth(sheetName, "B", "U", 10) // 数据列（增加到U）

	// 记录导出操作日志
	currentUser := auth.GetCurrentUser(r)
	if currentUser != nil {
		action := "导出月度建档数据 Excel"
		if month != "" || auditStatus != "" {
			action += "（"
			if month != "" {
				action += fmt.Sprintf("月份：%s", month)
			}
			if auditStatus != "" {
				if month != "" {
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
	filename := fmt.Sprintf("attachment; filename=\"月度建档数据_%s.xlsx\"", time.Now().Format("20060102_150405"))
	w.Header().Set("Content-Disposition", filename)
	f.Write(w)
}





