package checkpointprogress

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"ops-web/internal/auth"
	"ops-web/internal/checkpointfilelist"
	"ops-web/internal/db"
	"ops-web/internal/logger"
	"ops-web/internal/operationlog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// 审核任务数据结构体
type CheckpointTask struct {
	ID          int
	FileName    string // 档案名称
	Organization string // 机构名称
	ImportTime  string // 导入时间
	AuditStatus string // 审核状态
	RecordCount int    // 导入记录数量
	AuditComment sql.NullString // 审核意见
	UpdatedAt   string // 审核时间（updated_at）
	Attachments []string // 附件列表
	ArchiveType sql.NullString // 档案类型：新增、取推、补档案
}

// 页面数据结构体
type PageData struct {
	Title         string
	ActiveMenu    string
	SubMenu       string
	List          []CheckpointTask
	SearchName    string
	AuditStatus   string // 审核状态查询条件
	ArchiveType   string // 建档类型查询条件
	CurrentPage   int
	TotalPages    int
	HasPrev       bool
	HasNext       bool
	PrevPage      int
	NextPage      int
	FirstPage     int
	LastPage      int
	Query         string
	ImportMessage string
	ImportCount   int
}

// Handler: 卡口审核进度列表页 (GET)
func Handler(w http.ResponseWriter, r *http.Request) {
	// 获取查询参数
	searchName := r.URL.Query().Get("file_name")
	auditStatus := r.URL.Query().Get("audit_status") // 审核状态: 未审核, 已审核待整改, 已完成 或空
	archiveType := r.URL.Query().Get("archive_type") // 建档类型: 新增, 取推, 补档案 或空
	pageStr := r.URL.Query().Get("page")
	importMsg := r.URL.Query().Get("message")
	importCountStr := r.URL.Query().Get("count")

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	importCount, _ := strconv.Atoi(importCountStr)

	pageSize := 50
	offset := (page - 1) * pageSize

	// 构造查询条件
	whereSQL := " WHERE 1=1"
	args := []interface{}{}

	if searchName != "" {
		whereSQL += " AND file_name LIKE ?"
		args = append(args, "%"+searchName+"%")
	}

	// 审核状态查询
	if auditStatus != "" {
		// 验证审核状态值
		validStatuses := map[string]bool{
			"未审核":       true,
			"已审核待整改":  true,
			"已完成":      true,
		}
		if validStatuses[auditStatus] {
			whereSQL += " AND audit_status = ?"
			args = append(args, auditStatus)
		}
	}

	// 建档类型查询
	if archiveType != "" {
		// 验证建档类型值
		validTypes := map[string]bool{
			"新增":   true,
			"取推":   true,
			"补档案": true,
		}
		if validTypes[archiveType] {
			whereSQL += " AND archive_type = ?"
			args = append(args, archiveType)
		}
	}

	// 1. 查询总记录数
	var totalCount int
	countSQL := "SELECT COUNT(*) FROM checkpoint_tasks" + whereSQL
	err := db.DBInstance.QueryRow(countSQL, args...).Scan(&totalCount)
	if err != nil && err != sql.ErrNoRows {
		logger.Errorf("卡口审核进度-查询总数失败: %v, SQL: %s, Args: %v", err, countSQL, args)
		http.Error(w, "查询总数失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if totalCount == 0 {
		totalCount = 1
	}

	// 2. 分页计算
	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}
	if page > totalPages {
		page = totalPages
	}

	// 3. 查询列表数据
	querySQL := fmt.Sprintf("SELECT id, file_name, organization, import_time, audit_status, record_count, audit_comment, updated_at, archive_type FROM checkpoint_tasks %s ORDER BY id DESC LIMIT ? OFFSET ?", whereSQL)

	// 准备完整的参数列表
	queryArgs := append(args, pageSize, offset)

	rows, err := db.DBInstance.Query(querySQL, queryArgs...)
	if err != nil {
		logger.Errorf("卡口审核进度-数据库查询失败: %v, SQL: %s, Args: %v", err, querySQL, queryArgs)
		http.Error(w, "数据库查询失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var taskList []CheckpointTask
	for rows.Next() {
		var task CheckpointTask
		var importTimeRaw, updatedAtRaw sql.NullString
		err = rows.Scan(
			&task.ID,
			&task.FileName,
			&task.Organization,
			&importTimeRaw,
			&task.AuditStatus,
			&task.RecordCount,
			&task.AuditComment,
			&updatedAtRaw,
			&task.ArchiveType,
		)

		if err != nil {
			fmt.Printf("❌ Database scan error: %v\n", err)
			continue
		}

		// 格式化时间字段为 YYYY-MM-DD HH:mm
		task.ImportTime = formatDateTime(importTimeRaw.String)
		task.UpdatedAt = formatDateTime(updatedAtRaw.String)

		taskList = append(taskList, task)
	}

	// 检查遍历过程中的错误
	if err = rows.Err(); err != nil {
		logger.Errorf("卡口审核进度-数据遍历失败: %v", err)
		http.Error(w, "数据遍历失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 为每个任务查询附件列表
	uploadPath := getUploadPath()
	if uploadPath != "" {
		for i := range taskList {
			taskDir := filepath.Join(uploadPath, taskList[i].FileName)
			attachments, err := getAttachments(taskDir)
			if err == nil {
				taskList[i].Attachments = attachments
			}
		}
	}

	// 构建查询参数字符串（用于表单回显和分页链接）
	queryParams := []string{}
	if searchName != "" {
		queryParams = append(queryParams, "file_name="+searchName)
	}
	if auditStatus != "" {
		queryParams = append(queryParams, "audit_status="+auditStatus)
	}
	if archiveType != "" {
		queryParams = append(queryParams, "archive_type="+archiveType)
	}
	query := strings.Join(queryParams, "&")

	// 记录查询操作日志（如果有查询条件）
	currentUser := auth.GetCurrentUser(r)
	if currentUser != nil && (searchName != "" || auditStatus != "" || archiveType != "") {
		action := "查询卡口审核进度"
		hasCondition := false
		if searchName != "" {
			action += fmt.Sprintf("（档案名称：%s", searchName)
			hasCondition = true
		}
		if auditStatus != "" {
			if hasCondition {
				action += fmt.Sprintf("，审核状态：%s", auditStatus)
			} else {
				action += fmt.Sprintf("（审核状态：%s", auditStatus)
				hasCondition = true
			}
		}
		if archiveType != "" {
			if hasCondition {
				action += fmt.Sprintf("，建档类型：%s", archiveType)
			} else {
				action += fmt.Sprintf("（建档类型：%s", archiveType)
				hasCondition = true
			}
		}
		if hasCondition {
			action += "）"
		}
		operationlog.Record(r, currentUser.Username, action)
	}

	// 4. 准备数据并渲染模板
	data := PageData{
		Title:         "卡口审核进度",
		ActiveMenu:    "audit",
		SubMenu:       "checkpoint_progress",
		List:          taskList,
		SearchName:    searchName,
		AuditStatus:   auditStatus,
		ArchiveType:   archiveType,
		CurrentPage:   page,
		TotalPages:    totalPages,
		HasPrev:       page > 1,
		HasNext:       page < totalPages,
		PrevPage:      page - 1,
		NextPage:      page + 1,
		FirstPage:     1,
		LastPage:      totalPages,
		Query:         query,
		ImportMessage: importMsg,
		ImportCount:   importCount,
	}

	tmpl, err := template.ParseFiles("templates/checkpointprogress.html")
	if err != nil {
		logger.Errorf("卡口审核进度-模板解析失败: %v", err)
		http.Error(w, "模板解析失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		logger.Errorf("卡口审核进度-模板渲染失败: %v", err)
		http.Error(w, "模板渲染失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// ImportHandler: 导入 XLSX 档案
func ImportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/checkpoint/progress", http.StatusSeeOther)
		return
	}

	// 解析表单数据
	err := r.ParseMultipartForm(32 << 20) // 32MB
	if err != nil {
		logger.Errorf("卡口审核进度-表单解析失败: %v", err)
		http.Error(w, "表单解析失败", http.StatusBadRequest)
		return
	}

	// 获取机构名称（必填字段）
	organization := strings.TrimSpace(r.FormValue("organization"))
	if organization == "" {
		http.Error(w, "机构名称不能为空", http.StatusBadRequest)
		return
	}

	// 获取档案类型（必填字段）
	archiveType := strings.TrimSpace(r.FormValue("archive_type"))
	if archiveType == "" {
		http.Error(w, "档案类型不能为空", http.StatusBadRequest)
		return
	}
	// 验证档案类型值
	validArchiveTypes := map[string]bool{
		"新增": true,
		"取推": true,
		"补档案": true,
	}
	if !validArchiveTypes[archiveType] {
		http.Error(w, "档案类型值无效，必须是：新增、取推、补档案", http.StatusBadRequest)
		return
	}

	// 获取上传的文件
	file, fileHeader, err := r.FormFile("upload_file")
	if err != nil {
		logger.Errorf("卡口审核进度-文件上传失败: %v", err)
		http.Error(w, "文件上传失败", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 获取文件名（不含路径）
	fileName := fileHeader.Filename
	// 去除扩展名，只保留文件名部分作为档案名称
	fileNameWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))

	// 解析Excel文件
	f, err := excelize.OpenReader(file)
	if err != nil {
		logger.Errorf("卡口审核进度-Excel解析失败: %v, 文件名: %s", err, fileHeader.Filename)
		http.Error(w, "Excel解析失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	sheetName := f.GetSheetName(0)

	rows, err := f.GetRows(sheetName)
	if err != nil {
		logger.Errorf("卡口审核进度-数据读取失败: %v, 文件名: %s", err, fileHeader.Filename)
		http.Error(w, "数据读取失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(rows) < 2 {
		http.Error(w, "Excel文件至少需要包含表头和数据行", http.StatusBadRequest)
		return
	}

	// 开始事务
	tx, err := db.DBInstance.Begin()
	if err != nil {
		http.Error(w, "事务开始失败", http.StatusInternalServerError)
		return
	}

	// 1. 创建审核任务记录
	insertTaskSQL := `INSERT INTO checkpoint_tasks (file_name, organization, import_time, audit_status, archive_type) VALUES (?, ?, ?, ?, ?)`
	result, err := tx.Exec(insertTaskSQL, fileNameWithoutExt, organization, time.Now(), "未审核", archiveType)
	if err != nil {
		tx.Rollback()
		logger.Errorf("卡口审核进度-创建审核任务失败: %v, SQL: %s", err, insertTaskSQL)
		http.Error(w, "创建审核任务失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 获取插入的任务ID
	taskID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		logger.Errorf("卡口审核进度-获取任务ID失败: %v", err)
		http.Error(w, "获取任务ID失败", http.StatusInternalServerError)
		return
	}

	// 2. 导入Excel数据到checkpoint_details表
	// Excel列顺序：序号(跳过), 卡口编号, 原卡口编号, 卡口名称, 卡口地址, 道路名称, 方向类型, 方向描述, 方向备注, 行政区划, 路段类型, 道路代码, 公里数/路口号, 道路米数, 立杆编号, 卡口点位类型, 卡口位置类型, 卡口应用类型, 具备拦截条件, 具备车辆测速功能, 具备实时视频功能, 具备人脸抓拍功能, 具备违章抓拍功能, 具备前端二次识别功能, 是否边界卡口, 邻界地域, 卡口经度, 卡口纬度, 卡口实景照片地址, 卡口状态, 抓拍触发类型, 抓拍方向类型, 车道总数, 全景球机设备编码, 沿线下一卡口编号, 对向下一卡口编号, 左转下一卡口编号, 右转下一卡口编号, 掉头下一卡口编号, 建设单位, 管理单位, 卡口所属部门, 管理员姓名, 管理员联系电话, 卡口承建单位, 卡口维护单位, 接警部门, 接警部门代码, 接警电话, 拦截部门, 拦截部门代码, 拦截部门联系电话, 终端编码, 终端IP地址, 终端端口, 终端用户名, 终端密码, 终端厂商, 卡口启用时间, 卡口撤销时间, 备注, 卡口设备类型, 抓拍摄像机总数, 中控机编码, 中控机IP地址, 中控机端口, 中控机用户名, 中控机密码, 中控机厂商, 卡口报废时间, 天线总数, 终端MAC地址, 采集区域类型, 集成指挥平台卡口编号, 更新时间
	insertDetailFields := []string{
		"task_id", "checkpoint_code", "original_checkpoint_code", "checkpoint_name", "checkpoint_address", "road_name",
		"direction_type", "direction_description", "direction_notes", "division_code", "road_section_type", "road_code",
		"kilometer_or_intersection_number", "road_meter", "pole_number", "checkpoint_point_type", "checkpoint_location_type",
		"checkpoint_application_type", "has_interception_condition", "has_speed_measurement", "has_realtime_video", "has_face_capture",
		"has_violation_capture", "has_frontend_secondary_recognition", "is_boundary_checkpoint", "adjacent_area",
		"checkpoint_longitude", "checkpoint_latitude", "checkpoint_scene_photo_url", "checkpoint_status", "capture_trigger_type",
		"capture_direction_type", "total_lanes", "panoramic_camera_device_code", "next_checkpoint_along_road",
		"next_checkpoint_opposite", "next_checkpoint_left_turn", "next_checkpoint_right_turn", "next_checkpoint_u_turn",
		"construction_unit", "management_unit", "checkpoint_department", "admin_name", "admin_contact",
		"checkpoint_contractor", "checkpoint_maintain_unit", "alarm_receiving_department", "alarm_receiving_department_code",
		"alarm_receiving_phone", "interception_department", "interception_department_code", "interception_department_contact",
		"terminal_code", "terminal_ip_address", "terminal_port", "terminal_username", "terminal_password", "terminal_vendor",
		"checkpoint_enabled_time", "checkpoint_revoked_time", "notes", "checkpoint_device_type", "total_capture_cameras",
		"central_control_code", "central_control_ip_address", "central_control_port", "central_control_username",
		"central_control_password", "central_control_vendor", "checkpoint_scrapped_time", "total_antennas",
		"terminal_mac_address", "collection_area_type", "integrated_command_platform_checkpoint_code",
	}

	insertDetailSQL := fmt.Sprintf("INSERT INTO checkpoint_details (%s) VALUES (%s)",
		strings.Join(insertDetailFields, ", "),
		strings.TrimRight(strings.Repeat("?,", len(insertDetailFields)), ","))

	stmt, err := tx.Prepare(insertDetailSQL)
	if err != nil {
		tx.Rollback()
		logger.Errorf("卡口审核进度-SQL Prepare失败: %v, SQL: %s", err, insertDetailSQL)
		http.Error(w, "SQL Prepare失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	importedCount := 0
	const expectedCols = 75 // Excel总共75列（包括序号列）

	// 跳过表头，从第2行开始
	for i, row := range rows {
		if i == 0 {
			continue // 跳过表头
		}

		// 强制对齐到75列
		for len(row) < expectedCols {
			row = append(row, "")
		}

		// 准备参数（task_id + 74个字段，跳过序号列）
		params := make([]interface{}, len(insertDetailFields))
		params[0] = taskID // task_id

		// 映射Excel列到数据库字段（Excel第1列是序号，从第2列开始是数据）
		// Excel列索引：0=序号(跳过), 1=卡口编号, 2=原卡口编号, ..., 74=更新时间
		// params索引：0=task_id, 1=checkpoint_code(对应Excel列1), 2=original_checkpoint_code(对应Excel列2), ...
		for j := 1; j < len(insertDetailFields); j++ {
			excelIdx := j // Excel列索引（跳过序号列，所以j就是Excel列索引）
			
			// 所有字段都允许为空，如果Excel中没有值，就填充NULL
			params[j] = toDBValue(getRowValue(row, excelIdx), false)
		}

		_, execErr := stmt.Exec(params...)
		if execErr != nil {
			tx.Rollback()
			logger.Errorf("卡口审核进度-导入失败，第%d行数据错误: %v, 文件名: %s", i+1, execErr, fileHeader.Filename)
			errMsg := fmt.Sprintf("导入失败：第 %d 行数据错误。详细信息: %v", i+1, execErr)
			http.Error(w, errMsg, http.StatusInternalServerError)
			return
		}
		importedCount++
	}

	// 3. 更新任务表的记录数量
	updateTaskSQL := `UPDATE checkpoint_tasks SET record_count = ? WHERE id = ?`
	_, err = tx.Exec(updateTaskSQL, importedCount, taskID)
	if err != nil {
		tx.Rollback()
		logger.Errorf("卡口审核进度-更新记录数量失败: %v, SQL: %s", err, updateTaskSQL)
		http.Error(w, "更新记录数量失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		logger.Errorf("卡口审核进度-数据库提交失败: %v, 文件名: %s", err, fileHeader.Filename)
		http.Error(w, "数据库提交失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 记录导入操作日志
	if currentUser := auth.GetCurrentUser(r); currentUser != nil {
		action := fmt.Sprintf("导入卡口审核档案 Excel（档案名称：%s，机构：%s，共 %d 条数据）", fileNameWithoutExt, organization, importedCount)
		operationlog.Record(r, currentUser.Username, action)
	}

	http.Redirect(w, r, "/checkpoint/progress?message=ImportSuccess&count="+strconv.Itoa(importedCount), http.StatusSeeOther)
}

// EditCommentHandler: 编辑审核意见
func EditCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// 显示编辑页面
		taskIDStr := r.URL.Query().Get("task_id")
		taskID, err := strconv.Atoi(taskIDStr)
		if err != nil || taskID <= 0 {
			http.Error(w, "无效的任务ID", http.StatusBadRequest)
			return
		}

		var task CheckpointTask
		var importTimeRaw sql.NullString
		taskSQL := "SELECT id, file_name, organization, import_time, audit_status, audit_comment FROM checkpoint_tasks WHERE id = ?"
		err = db.DBInstance.QueryRow(taskSQL, taskID).Scan(
			&task.ID,
			&task.FileName,
			&task.Organization,
			&importTimeRaw,
			&task.AuditStatus,
			&task.AuditComment,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "档案不存在", http.StatusNotFound)
			} else {
				http.Error(w, "查询档案失败: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}
		task.ImportTime = formatDateTime(importTimeRaw.String)

		type EditPageData struct {
			Title      string
			ActiveMenu string
			SubMenu    string
			Task       CheckpointTask
		}

		data := EditPageData{
			Title:      "编辑审核意见",
			ActiveMenu: "audit",
			SubMenu:    "checkpoint_progress",
			Task:       task,
		}

		tmpl, err := template.ParseFiles("templates/checkpointedit.html")
		if err != nil {
			http.Error(w, "模板解析失败: "+err.Error(), http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, "模板渲染失败: "+err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	// POST: 保存审核意见
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "表单解析失败", http.StatusBadRequest)
			return
		}

		taskIDStr := r.FormValue("task_id")
		taskID, err := strconv.Atoi(taskIDStr)
		if err != nil || taskID <= 0 {
			http.Error(w, "无效的任务ID", http.StatusBadRequest)
			return
		}

		auditComment := strings.TrimSpace(r.FormValue("audit_comment"))
		auditStatus := strings.TrimSpace(r.FormValue("audit_status"))

		// 验证审核状态
		validStatuses := map[string]bool{
			"未审核":       true,
			"已审核待整改":  true,
			"已完成":      true,
		}
		if !validStatuses[auditStatus] {
			http.Error(w, "无效的审核状态", http.StatusBadRequest)
			return
		}

		// 开始事务
		tx, err := db.DBInstance.Begin()
		if err != nil {
			http.Error(w, "事务开始失败", http.StatusInternalServerError)
			return
		}

		// 更新审核意见和状态
		updateSQL := `UPDATE checkpoint_tasks SET audit_comment = ?, audit_status = ?, updated_at = NOW() WHERE id = ?`
		var comment interface{}
		if auditComment == "" {
			comment = nil
		} else {
			comment = auditComment
		}

		_, err = tx.Exec(updateSQL, comment, auditStatus, taskID)
		if err != nil {
			tx.Rollback()
			http.Error(w, "更新审核意见失败: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 根据档案状态更新所有明细的 audit_status
		var detailStatus int
		switch auditStatus {
		case "未审核":
			detailStatus = 0
		case "已审核待整改":
			detailStatus = 1
		case "已完成":
			detailStatus = 2
		default:
			detailStatus = 0
		}

		updateDetailSQL := `UPDATE checkpoint_details SET audit_status = ? WHERE task_id = ?`
		_, err = tx.Exec(updateDetailSQL, detailStatus, taskID)
		if err != nil {
			tx.Rollback()
			http.Error(w, "更新明细状态失败: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			http.Error(w, "事务提交失败: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 记录操作日志
		if currentUser := auth.GetCurrentUser(r); currentUser != nil {
			action := fmt.Sprintf("编辑卡口审核意见（档案ID：%d，状态：%s）", taskID, auditStatus)
			operationlog.Record(r, currentUser.Username, action)
		}

		http.Redirect(w, r, "/checkpoint/progress?message=EditSuccess", http.StatusSeeOther)
		return
	}

	http.Error(w, "不支持的请求方法", http.StatusMethodNotAllowed)
}

// 辅助函数：获取行值（安全地获取，防止索引越界）
func getRowValue(row []string, idx int) string {
	if idx < len(row) {
		return row[idx]
	}
	return ""
}

// 辅助工具：空字符串转 NULL
func toDBValue(s string, required bool) interface{} {
	s = strings.TrimSpace(s)
	if s == "" {
		if required {
			return ""
		}
		return nil
	}
	return s
}

// DetailHandler: 查看档案明细
func DetailHandler(w http.ResponseWriter, r *http.Request) {
	taskIDStr := r.URL.Query().Get("task_id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil || taskID <= 0 {
		http.Error(w, "无效的任务ID", http.StatusBadRequest)
		return
	}

	// 查询任务基本信息
	var task CheckpointTask
	taskSQL := "SELECT id, file_name, organization, import_time, audit_status, audit_comment FROM checkpoint_tasks WHERE id = ?"
	err = db.DBInstance.QueryRow(taskSQL, taskID).Scan(
		&task.ID,
		&task.FileName,
		&task.Organization,
		&task.ImportTime,
		&task.AuditStatus,
		&task.AuditComment,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "档案不存在", http.StatusNotFound)
		} else {
			http.Error(w, "查询档案失败: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// 查询明细数据
	type DetailItem struct {
		ID                    int
		CheckpointCode        string
		CheckpointName        string
		DivisionCode          string
		ManagementUnit        string
		CheckpointPointType   string
		CheckpointMaintainUnit string
		AuditStatus           int
	}

	detailSQL := `SELECT id, checkpoint_code, checkpoint_name, division_code, management_unit, 
		checkpoint_point_type, checkpoint_maintain_unit, audit_status
		FROM checkpoint_details WHERE task_id = ? ORDER BY id`
	
	rows, err := db.DBInstance.Query(detailSQL, taskID)
	if err != nil {
		http.Error(w, "查询明细失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var detailList []DetailItem
	for rows.Next() {
		var item DetailItem
		err = rows.Scan(
			&item.ID,
			&item.CheckpointCode,
			&item.CheckpointName,
			&item.DivisionCode,
			&item.ManagementUnit,
			&item.CheckpointPointType,
			&item.CheckpointMaintainUnit,
			&item.AuditStatus,
		)
		if err != nil {
			continue
		}
		detailList = append(detailList, item)
	}

	// 准备数据
	type DetailPageData struct {
		Title         string
		ActiveMenu    string
		SubMenu       string
		Task          CheckpointTask
		Details       []DetailItem
	}

	data := DetailPageData{
		Title:      "档案明细",
		ActiveMenu: "audit",
		SubMenu:    "checkpoint_progress",
		Task:       task,
		Details:    detailList,
	}

	// 添加状态转换函数到模板
	funcMap := template.FuncMap{
		"getStatusText": getAuditStatusText,
	}

	tmpl, err := template.New("checkpointdetail.html").Funcs(funcMap).ParseFiles("templates/checkpointdetail.html")
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

// getAuditStatusText 将建档状态码转换为文字
func getAuditStatusText(status int) string {
	switch status {
	case 0:
		return "未审核未建档"
	case 1:
		return "已审核未建档"
	case 2:
		return "已建档"
	default:
		return "未知状态"
	}
}

// DetailExportHandler: 导出卡口档案明细到Excel（全量字段）
func DetailExportHandler(w http.ResponseWriter, r *http.Request) {
	taskIDStr := r.URL.Query().Get("task_id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil || taskID <= 0 {
		http.Error(w, "无效的任务ID", http.StatusBadRequest)
		return
	}

	// 查询任务基本信息（获取档案名称）
	var task CheckpointTask
	taskSQL := "SELECT file_name FROM checkpoint_tasks WHERE id = ?"
	err = db.DBInstance.QueryRow(taskSQL, taskID).Scan(&task.FileName)
	if err != nil {
		http.Error(w, "查询档案失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 查询所有字段（从checkpoint_details表，只查询该任务的数据）
	querySQL := `SELECT 
		id, checkpoint_code, original_checkpoint_code, checkpoint_name, checkpoint_address, road_name,
		direction_type, direction_description, direction_notes, division_code, road_section_type, road_code,
		kilometer_or_intersection_number, road_meter, pole_number, checkpoint_point_type, checkpoint_location_type,
		checkpoint_application_type, has_interception_condition, has_speed_measurement, has_realtime_video, has_face_capture,
		has_violation_capture, has_frontend_secondary_recognition, is_boundary_checkpoint, adjacent_area,
		checkpoint_longitude, checkpoint_latitude, checkpoint_scene_photo_url, checkpoint_status, capture_trigger_type,
		capture_direction_type, total_lanes, panoramic_camera_device_code, next_checkpoint_along_road,
		next_checkpoint_opposite, next_checkpoint_left_turn, next_checkpoint_right_turn, next_checkpoint_u_turn,
		construction_unit, management_unit, checkpoint_department, admin_name, admin_contact,
		checkpoint_contractor, checkpoint_maintain_unit, alarm_receiving_department, alarm_receiving_department_code,
		alarm_receiving_phone, interception_department, interception_department_code, interception_department_contact,
		terminal_code, terminal_ip_address, terminal_port, terminal_username, terminal_password, terminal_vendor,
		checkpoint_enabled_time, checkpoint_revoked_time, notes, checkpoint_device_type, total_capture_cameras,
		central_control_code, central_control_ip_address, central_control_port, central_control_username,
		central_control_password, central_control_vendor, checkpoint_scrapped_time, total_antennas,
		terminal_mac_address, collection_area_type, integrated_command_platform_checkpoint_code, update_time, audit_status
	FROM checkpoint_details WHERE task_id = ? ORDER BY id`

	rows, err := db.DBInstance.Query(querySQL, taskID)
	if err != nil {
		logger.Errorf("卡口审核进度-导出查询失败: %v, SQL: %s, taskID: %d", err, querySQL, taskID)
		http.Error(w, "导出查询失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// 创建Excel文件
	f := excelize.NewFile()
	sheetName := "档案明细"
	f.SetSheetName("Sheet1", sheetName)

	// 使用checkpointfilelist的ExportHeaders（75列）
	ExportHeaders := []interface{}{
		"序号", "卡口编号（*）", "原卡口编号", "卡口名称（*）", "卡口地址（*）", "道路名称（*）", "方向类型",
		"方向描述（*）", "方向备注", "行政区划（*）", "路段类型（*）", "道路代码（*）", "公里数/路口号（*）", "道路米数（*）",
		"立杆编号", "卡口点位类型（*）", "卡口位置类型（*）", "卡口应用类型（*）", "具备拦截条件（*）", "具备车辆测速功能",
		"具备实时视频功能", "具备人脸抓拍功能", "具备违章抓拍功能", "具备前端二次识别功能", "是否边界卡口（*）", "邻界地域（*）",
		"卡口经度（*）", "卡口纬度（*）", "卡口实景照片地址", "卡口状态（*）", "抓拍触发类型", "抓拍方向类型（*）",
		"车道总数（*）", "全景球机设备编码（*）", "沿线下一卡口编号", "对向下一卡口编号", "左转下一卡口编号", "右转下一卡口编号",
		"掉头下一卡口编号", "建设单位（*）", "管理单位（*）", "卡口所属部门（*）", "管理员姓名（*）", "管理员联系电话",
		"卡口承建单位", "卡口维护单位", "接警部门（*）", "接警部门代码（*）", "接警电话（*）", "拦截部门（*）",
		"拦截部门代码（*）", "拦截部门联系电话（*）", "终端编码", "终端IP地址（*）", "终端端口", "终端用户名",
		"终端密码", "终端厂商", "卡口启用时间（*）", "卡口撤销时间", "备注", "卡口设备类型（*）",
		"抓拍摄像机总数（*）", "中控机编码", "中控机IP地址", "中控机端口", "中控机用户名", "中控机密码",
		"中控机厂商", "卡口报废时间", "天线总数", "终端MAC地址（*）", "采集区域类型", "集成指挥平台卡口编号（组）",
		"更新时间", "建档状态",
	}

	// 写入表头
	f.SetSheetRow(sheetName, "A1", &ExportHeaders)

	// 设置表头样式
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#3498db"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	lastColName, _ := excelize.CoordinatesToCellName(len(ExportHeaders), 1)
	lastColLetter := strings.Split(lastColName, "1")[0]
	f.SetCellStyle(sheetName, "A1", lastColLetter+"1", headerStyle)

	// 写入数据
	rowNum := 2
	for rows.Next() {
		var item checkpointfilelist.CheckpointFileItem
		err = rows.Scan(
			&item.ID, &item.CheckpointCode, &item.OriginalCheckpointCode,
			&item.CheckpointName, &item.CheckpointAddress, &item.RoadName,
			&item.DirectionType, &item.DirectionDescription, &item.DirectionNotes,
			&item.DivisionCode, &item.RoadSectionType, &item.RoadCode,
			&item.KilometerOrIntersectionNumber, &item.RoadMeter, &item.PoleNumber,
			&item.CheckpointPointType, &item.CheckpointLocationType, &item.CheckpointApplicationType,
			&item.HasInterceptionCondition, &item.HasSpeedMeasurement, &item.HasRealtimeVideo,
			&item.HasFaceCapture, &item.HasViolationCapture, &item.HasFrontendSecondaryRecognition,
			&item.IsBoundaryCheckpoint, &item.AdjacentArea, &item.CheckpointLongitude,
			&item.CheckpointLatitude, &item.CheckpointScenePhotoURL, &item.CheckpointStatus,
			&item.CaptureTriggerType, &item.CaptureDirectionType, &item.TotalLanes,
			&item.PanoramicCameraDeviceCode, &item.NextCheckpointAlongRoad, &item.NextCheckpointOpposite,
			&item.NextCheckpointLeftTurn, &item.NextCheckpointRightTurn, &item.NextCheckpointUTurn,
			&item.ConstructionUnit, &item.ManagementUnit, &item.CheckpointDepartment,
			&item.AdminName, &item.AdminContact, &item.CheckpointContractor,
			&item.CheckpointMaintainUnit, &item.AlarmReceivingDepartment, &item.AlarmReceivingDepartmentCode,
			&item.AlarmReceivingPhone, &item.InterceptionDepartment, &item.InterceptionDepartmentCode,
			&item.InterceptionDepartmentContact, &item.TerminalCode, &item.TerminalIPAddress,
			&item.TerminalPort, &item.TerminalUsername, &item.TerminalPassword, &item.TerminalVendor,
			&item.CheckpointEnabledTime, &item.CheckpointRevokedTime, &item.Notes,
			&item.CheckpointDeviceType, &item.TotalCaptureCameras, &item.CentralControlCode,
			&item.CentralControlIPAddress, &item.CentralControlPort, &item.CentralControlUsername,
			&item.CentralControlPassword, &item.CentralControlVendor, &item.CheckpointScrappedTime,
			&item.TotalAntennas, &item.TerminalMACAddress, &item.CollectionAreaType,
			&item.IntegratedCommandPlatformCheckpointCode, &item.UpdateTime, &item.AuditStatus,
		)
		if err != nil {
			continue
		}

		// 构建行数据
		statusText := getAuditStatusText(item.AuditStatus)
		rowData := []interface{}{
			item.ID,
			item.CheckpointCode.String, item.OriginalCheckpointCode.String, item.CheckpointName.String,
			item.CheckpointAddress.String, item.RoadName.String, item.DirectionType.String,
			item.DirectionDescription.String, item.DirectionNotes.String, item.DivisionCode.String,
			item.RoadSectionType.String, item.RoadCode.String, item.KilometerOrIntersectionNumber.String,
			item.RoadMeter.String, item.PoleNumber.String, item.CheckpointPointType.String,
			item.CheckpointLocationType.String, item.CheckpointApplicationType.String,
			item.HasInterceptionCondition.String, item.HasSpeedMeasurement.String,
			item.HasRealtimeVideo.String, item.HasFaceCapture.String, item.HasViolationCapture.String,
			item.HasFrontendSecondaryRecognition.String, item.IsBoundaryCheckpoint.String,
			item.AdjacentArea.String, item.CheckpointLongitude.String, item.CheckpointLatitude.String,
			item.CheckpointScenePhotoURL.String, item.CheckpointStatus.String, item.CaptureTriggerType.String,
			item.CaptureDirectionType.String, item.TotalLanes.String, item.PanoramicCameraDeviceCode.String,
			item.NextCheckpointAlongRoad.String, item.NextCheckpointOpposite.String,
			item.NextCheckpointLeftTurn.String, item.NextCheckpointRightTurn.String,
			item.NextCheckpointUTurn.String, item.ConstructionUnit.String, item.ManagementUnit.String,
			item.CheckpointDepartment.String, item.AdminName.String, item.AdminContact.String,
			item.CheckpointContractor.String, item.CheckpointMaintainUnit.String,
			item.AlarmReceivingDepartment.String, item.AlarmReceivingDepartmentCode.String,
			item.AlarmReceivingPhone.String, item.InterceptionDepartment.String,
			item.InterceptionDepartmentCode.String, item.InterceptionDepartmentContact.String,
			item.TerminalCode.String, item.TerminalIPAddress.String, item.TerminalPort.String,
			item.TerminalUsername.String, item.TerminalPassword.String, item.TerminalVendor.String,
			item.CheckpointEnabledTime.String, item.CheckpointRevokedTime.String, item.Notes.String,
			item.CheckpointDeviceType.String, item.TotalCaptureCameras.String, item.CentralControlCode.String,
			item.CentralControlIPAddress.String, item.CentralControlPort.String,
			item.CentralControlUsername.String, item.CentralControlPassword.String,
			item.CentralControlVendor.String, item.CheckpointScrappedTime.String, item.TotalAntennas.String,
			item.TerminalMACAddress.String, item.CollectionAreaType.String,
			item.IntegratedCommandPlatformCheckpointCode.String, item.UpdateTime, statusText,
		}

		cellName, _ := excelize.CoordinatesToCellName(1, rowNum)
		f.SetSheetRow(sheetName, cellName, &rowData)
		rowNum++
	}

	// 设置列宽（为所有列设置合适的宽度）
	for i := 0; i < len(ExportHeaders); i++ {
		colName, _ := excelize.CoordinatesToCellName(i+1, 1)
		colLetter := strings.Split(colName, "1")[0]
		f.SetColWidth(sheetName, colLetter, colLetter, 15) // 默认宽度15
	}

	// 记录导出操作日志
	currentUser := auth.GetCurrentUser(r)
	if currentUser != nil {
		action := fmt.Sprintf("导出卡口审核档案明细 Excel（档案名称：%s）", task.FileName)
		operationlog.Record(r, currentUser.Username, action)
	}

	// 输出文件（使用档案名称作为文件名）
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	// 清理文件名中的特殊字符
	fileName := strings.ReplaceAll(task.FileName, "/", "_")
	fileName = strings.ReplaceAll(fileName, "\\", "_")
	fileName = strings.ReplaceAll(fileName, ":", "_")
	fileName = strings.ReplaceAll(fileName, "*", "_")
	fileName = strings.ReplaceAll(fileName, "?", "_")
	fileName = strings.ReplaceAll(fileName, "\"", "_")
	fileName = strings.ReplaceAll(fileName, "<", "_")
	fileName = strings.ReplaceAll(fileName, ">", "_")
	fileName = strings.ReplaceAll(fileName, "|", "_")
	filename := fmt.Sprintf("attachment; filename=\"%s.xlsx\"", fileName)
	w.Header().Set("Content-Disposition", filename)
	f.Write(w)
}

// 模板下载时的表头定义 (75列)
var TemplateHeaders = []interface{}{
	"序号", "卡口编号（*）", "原卡口编号", "卡口名称（*）", "卡口地址（*）", "道路名称（*）", "方向类型",
	"方向描述（*）", "方向备注", "行政区划（*）", "路段类型（*）", "道路代码（*）", "公里数/路口号（*）", "道路米数（*）",
	"立杆编号", "卡口点位类型（*）", "卡口位置类型（*）", "卡口应用类型（*）", "具备拦截条件（*）", "具备车辆测速功能",
	"具备实时视频功能", "具备人脸抓拍功能", "具备违章抓拍功能", "具备前端二次识别功能", "是否边界卡口（*）", "邻界地域（*）",
	"卡口经度（*）", "卡口纬度（*）", "卡口实景照片地址", "卡口状态（*）", "抓拍触发类型", "抓拍方向类型（*）",
	"车道总数（*）", "全景球机设备编码（*）", "沿线下一卡口编号", "对向下一卡口编号", "左转下一卡口编号", "右转下一卡口编号",
	"掉头下一卡口编号", "建设单位（*）", "管理单位（*）", "卡口所属部门（*）", "管理员姓名（*）", "管理员联系电话",
	"卡口承建单位", "卡口维护单位", "接警部门（*）", "接警部门代码（*）", "接警电话（*）", "拦截部门（*）",
	"拦截部门代码（*）", "拦截部门联系电话（*）", "终端编码", "终端IP地址（*）", "终端端口", "终端用户名",
	"终端密码", "终端厂商", "卡口启用时间（*）", "卡口撤销时间", "备注", "卡口设备类型（*）",
	"抓拍摄像机总数（*）", "中控机编码", "中控机IP地址", "中控机端口", "中控机用户名", "中控机密码",
	"中控机厂商", "卡口报废时间", "天线总数", "终端MAC地址（*）", "采集区域类型", "集成指挥平台卡口编号（组）",
	"更新时间",
}

// DownloadTemplateHandler: 下载导入模板
func DownloadTemplateHandler(w http.ResponseWriter, r *http.Request) {
	f := excelize.NewFile()
	sheetName := "导入模板"
	f.SetSheetName("Sheet1", sheetName)
	f.SetSheetRow(sheetName, "A1", &TemplateHeaders)

	// 记录操作日志
	currentUser := auth.GetCurrentUser(r)
	if currentUser != nil {
		action := "下载卡口审核档案导入模板"
		operationlog.Record(r, currentUser.Username, action)
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	filename := fmt.Sprintf("attachment; filename=\"卡口审核档案导入模板_%s.xlsx\"", time.Now().Format("20060102"))
	w.Header().Set("Content-Disposition", filename)
	f.Write(w)
}

// DeleteHandler: 删除审核档案（同时删除关联的明细）
func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/checkpoint/progress", http.StatusSeeOther)
		return
	}

	// 获取要删除的任务ID
	taskIDStr := r.FormValue("task_id")
	if taskIDStr == "" {
		http.Error(w, "缺少任务ID参数", http.StatusBadRequest)
		return
	}

	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil || taskID <= 0 {
		http.Error(w, "无效的任务ID", http.StatusBadRequest)
		return
	}

	// 查询任务信息（用于日志记录）
	var task CheckpointTask
	taskSQL := "SELECT id, file_name, organization FROM checkpoint_tasks WHERE id = ?"
	err = db.DBInstance.QueryRow(taskSQL, taskID).Scan(
		&task.ID,
		&task.FileName,
		&task.Organization,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "档案不存在", http.StatusNotFound)
		} else {
			http.Error(w, "查询档案失败: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// 查询关联的明细数量（用于日志记录）
	var detailCount int
	countSQL := "SELECT COUNT(*) FROM checkpoint_details WHERE task_id = ?"
	err = db.DBInstance.QueryRow(countSQL, taskID).Scan(&detailCount)
	if err != nil {
		detailCount = 0
	}

	// 开始事务
	tx, err := db.DBInstance.Begin()
	if err != nil {
		http.Error(w, "事务开始失败", http.StatusInternalServerError)
		return
	}

	// 先删除关联的明细记录
	deleteDetailsSQL := "DELETE FROM checkpoint_details WHERE task_id = ?"
	_, err = tx.Exec(deleteDetailsSQL, taskID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "删除档案明细失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 再删除任务记录
	deleteTaskSQL := "DELETE FROM checkpoint_tasks WHERE id = ?"
	_, err = tx.Exec(deleteTaskSQL, taskID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "删除档案失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		http.Error(w, "事务提交失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 记录删除操作日志
	currentUser := auth.GetCurrentUser(r)
	if currentUser != nil {
		action := fmt.Sprintf("删除卡口审核档案（档案名称：%s，机构：%s，包含 %d 条明细）", task.FileName, task.Organization, detailCount)
		operationlog.Record(r, currentUser.Username, action)
	}

	// 重定向回列表页（保留查询参数）
	searchName := r.FormValue("file_name")
	if searchName == "" {
		searchName = r.URL.Query().Get("file_name")
	}
	auditStatus := r.FormValue("audit_status")
	if auditStatus == "" {
		auditStatus = r.URL.Query().Get("audit_status")
	}
	redirectURL := "/checkpoint/progress?message=DeleteSuccess"
	if searchName != "" {
		redirectURL += "&file_name=" + searchName
	}
	if auditStatus != "" {
		redirectURL += "&audit_status=" + auditStatus
	}
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// UploadHandler 处理文件上传
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	// 获取当前用户
	currentUser := auth.GetCurrentUser(r)
	if currentUser == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// 获取task_id和档案名称
	taskIDStr := r.FormValue("task_id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		http.Error(w, "无效的任务ID", http.StatusBadRequest)
		return
	}

	// 查询档案名称
	var fileName string
	query := "SELECT file_name FROM checkpoint_tasks WHERE id = ?"
	err = db.DBInstance.QueryRow(query, taskID).Scan(&fileName)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "档案不存在", http.StatusNotFound)
			return
		}
		logger.Errorf("卡口审核进度-查询档案名称失败: %v", err)
		http.Error(w, "查询档案失败", http.StatusInternalServerError)
		return
	}

	// 获取上传文件
	file, header, err := r.FormFile("upload_file")
	if err != nil {
		http.Error(w, "获取上传文件失败: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 获取任务配置中的上传路径
	uploadPath := getUploadPath()
	if uploadPath == "" {
		http.Error(w, "未配置上传路径，请先在任务配置中设置", http.StatusInternalServerError)
		return
	}

	// 创建以档案名称命名的目录
	targetDir := filepath.Join(uploadPath, fileName)
	err = createDirIfNotExists(targetDir)
	if err != nil {
		logger.Errorf("卡口审核进度-创建目录失败: %v, 路径: %s", err, targetDir)
		http.Error(w, "创建目录失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 保存文件
	targetFile := filepath.Join(targetDir, header.Filename)
	dst, err := os.Create(targetFile)
	if err != nil {
		logger.Errorf("卡口审核进度-创建文件失败: %v, 路径: %s", err, targetFile)
		http.Error(w, "保存文件失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		logger.Errorf("卡口审核进度-写入文件失败: %v, 路径: %s", err, targetFile)
		http.Error(w, "写入文件失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 记录操作日志
	action := fmt.Sprintf("上传文件（档案：%s，文件名：%s）", fileName, header.Filename)
	operationlog.Record(r, currentUser.Username, action)

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true, "message": "文件上传成功"}`))
}

// getUploadPath 获取上传路径配置
func getUploadPath() string {
	var value string
	query := "SELECT param_value FROM system_settings WHERE param_key = ?"
	err := db.DBInstance.QueryRow(query, "upload_file_path").Scan(&value)
	if err != nil {
		return ""
	}
	return value
}

// createDirIfNotExists 如果目录不存在则创建
func createDirIfNotExists(dirPath string) error {
	_, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return os.MkdirAll(dirPath, 0755)
	}
	return err
}

// getAttachments 获取指定目录下的所有文件列表
func getAttachments(dirPath string) ([]string, error) {
	var attachments []string
	_, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return attachments, nil // 目录不存在，返回空列表
	}
	
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	
	for _, entry := range entries {
		if !entry.IsDir() {
			attachments = append(attachments, entry.Name())
		}
	}
	
	return attachments, nil
}

// DownloadHandler 处理附件下载
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	// 获取参数
	taskIDStr := r.URL.Query().Get("task_id")
	fileName := r.URL.Query().Get("file_name")
	
	if taskIDStr == "" || fileName == "" {
		http.Error(w, "参数不完整", http.StatusBadRequest)
		return
	}
	
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		http.Error(w, "无效的任务ID", http.StatusBadRequest)
		return
	}
	
	// 查询档案名称
	var archiveFileName string
	query := "SELECT file_name FROM checkpoint_tasks WHERE id = ?"
	err = db.DBInstance.QueryRow(query, taskID).Scan(&archiveFileName)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "档案不存在", http.StatusNotFound)
			return
		}
		logger.Errorf("卡口审核进度-查询档案名称失败: %v", err)
		http.Error(w, "查询档案失败", http.StatusInternalServerError)
		return
	}
	
	// 获取上传路径
	uploadPath := getUploadPath()
	if uploadPath == "" {
		http.Error(w, "未配置上传路径", http.StatusInternalServerError)
		return
	}
	
	// 构建文件路径
	filePath := filepath.Join(uploadPath, archiveFileName, fileName)
	
	// 验证文件是否存在
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		http.Error(w, "文件不存在", http.StatusNotFound)
		return
	}
	
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		logger.Errorf("卡口审核进度-打开文件失败: %v, 路径: %s", err, filePath)
		http.Error(w, "打开文件失败", http.StatusInternalServerError)
		return
	}
	defer file.Close()
	
	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		logger.Errorf("卡口审核进度-获取文件信息失败: %v", err)
		http.Error(w, "获取文件信息失败", http.StatusInternalServerError)
		return
	}
	
	// 设置响应头
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	
	// 复制文件内容到响应
	_, err = io.Copy(w, file)
	if err != nil {
		logger.Errorf("卡口审核进度-下载文件失败: %v", err)
		return
	}
	
	// 记录操作日志
	currentUser := auth.GetCurrentUser(r)
	if currentUser != nil {
		action := fmt.Sprintf("下载附件（档案：%s，文件名：%s）", archiveFileName, fileName)
		operationlog.Record(r, currentUser.Username, action)
	}
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
