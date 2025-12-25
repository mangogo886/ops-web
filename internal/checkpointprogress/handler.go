package checkpointprogress

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"ops-web/internal/auth"
	"ops-web/internal/auditprogress"
	"ops-web/internal/checkpointfilelist"
	"ops-web/internal/db"
	"ops-web/internal/logger"
	"ops-web/internal/operationlog"
	"ops-web/internal/permission"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
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
	// 抽检相关字段
	IsSampled       bool   // 是否已抽检
	LastSampledAt   string // 最后抽检时间
	SampledBy       string // 最后抽检人员
	SampleCount     int    // 抽检次数
	LastSampleResult string // 最近一次抽检结果
	// 操作链接URL（包含筛选条件参数）
	DetailURL string // 查看明细链接
	EditURL   string // 编辑链接
	SampleURL string // 抽检链接
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
	SampleStatus  string // 抽检状态查询条件：全部、已抽检、待整改、待抽检
	CurrentPage   int
	TotalPages    int
	TotalCount    int    // 总记录数
	StartRecord   int    // 当前页起始记录号
	EndRecord     int    // 当前页结束记录号
	HasPrev       bool
	HasNext       bool
	PrevPage      int
	NextPage      int
	FirstPage     int
	LastPage      int
	Query         string
	QueryParams   map[string]string // 筛选条件参数的键值对，用于在链接中单独添加
	ImportMessage string
	ImportCount   int
	// 权限信息
	CanImport bool // 是否可以导入
	CanDelete bool // 是否可以删除
}

// Handler: 卡口审核进度列表页 (GET)
func Handler(w http.ResponseWriter, r *http.Request) {
	// 获取查询参数
	searchName := r.URL.Query().Get("file_name")
	auditStatus := r.URL.Query().Get("audit_status") // 审核状态: 未审核, 已审核待整改, 已完成 或空
	archiveType := r.URL.Query().Get("archive_type") // 建档类型: 新增, 取推, 补档案 或空
	sampleStatus := r.URL.Query().Get("sample_status") // 抽检状态: 全部, 已抽检, 待整改, 待抽检 或空
	pageStr := r.URL.Query().Get("page")
	importMsg := r.URL.Query().Get("message")
	importCountStr := r.URL.Query().Get("count")

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	importCount, _ := strconv.Atoi(importCountStr)

	pageSize := 30
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
			"变更":   true,
		}
		if validTypes[archiveType] {
			whereSQL += " AND archive_type = ?"
			args = append(args, archiveType)
		}
	}

	// 抽检状态查询
	if sampleStatus != "" {
		validSampleStatuses := map[string]bool{
			"已抽检": true,
			"待整改": true,
			"待抽检": true,
		}
		if validSampleStatuses[sampleStatus] {
			if sampleStatus == "已抽检" {
				// 已抽检：审核状态为"已完成"，is_sampled = 1，且最近一次抽检结果不是"待整改"（包括NULL）
				whereSQL += ` AND audit_status = '已完成' 
					AND is_sampled = 1 
					AND COALESCE((SELECT sample_result FROM checkpoint_sample_records 
						WHERE task_id = checkpoint_tasks.id 
						ORDER BY sampled_at DESC LIMIT 1), '') != '待整改'`
			} else if sampleStatus == "待整改" {
				// 待整改：审核状态为"已完成"，is_sampled = 1，且最近一次抽检结果为"待整改"
				whereSQL += ` AND audit_status = '已完成' 
					AND is_sampled = 1 
					AND (SELECT sample_result FROM checkpoint_sample_records 
						WHERE task_id = checkpoint_tasks.id 
						ORDER BY sampled_at DESC LIMIT 1) = '待整改'`
			} else if sampleStatus == "待抽检" {
				// 待抽检：审核状态为"已完成"，is_sampled = 0 或 NULL
				whereSQL += ` AND audit_status = '已完成' 
					AND (is_sampled = 0 OR is_sampled IS NULL)`
			}
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

	// 2. 分页计算
	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}
	if page > totalPages {
		page = totalPages
	}

	// 3. 查询列表数据（包含抽检字段）
	querySQL := fmt.Sprintf("SELECT id, file_name, organization, import_time, audit_status, record_count, audit_comment, updated_at, archive_type, is_sampled, last_sampled_at FROM checkpoint_tasks %s ORDER BY id DESC LIMIT ? OFFSET ?", whereSQL)

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
	var taskIDs []int
	for rows.Next() {
		var task CheckpointTask
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
			&task.ArchiveType,
			&isSampled,
			&lastSampledAtRaw,
		)

		if err != nil {
			logger.Errorf("卡口审核进度-数据库扫描错误: %v, taskID: %d", err, task.ID)
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
		logger.Errorf("卡口审核进度-数据遍历失败: %v", err)
		http.Error(w, "数据遍历失败: "+err.Error(), http.StatusInternalServerError)
		return
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
	queryValues := url.Values{}
	if searchName != "" {
		queryValues.Set("file_name", searchName)
	}
	if auditStatus != "" {
		queryValues.Set("audit_status", auditStatus)
	}
	if archiveType != "" {
		queryValues.Set("archive_type", archiveType)
	}
	if sampleStatus != "" {
		queryValues.Set("sample_status", sampleStatus)
	}
	query := queryValues.Encode()
	
	// 构建每个参数的独立值，用于在模板中单独添加
	queryParams := make(map[string]string)
	if searchName != "" {
		queryParams["file_name"] = searchName
	}
	if auditStatus != "" {
		queryParams["audit_status"] = auditStatus
	}
	if archiveType != "" {
		queryParams["archive_type"] = archiveType
	}
	if sampleStatus != "" {
		queryParams["sample_status"] = sampleStatus
	}
	
	// 为每个任务构建操作链接的URL（包含筛选条件参数）
	baseURL := "/checkpoint/progress"
	for i := range taskList {
		// 构建查看明细链接
		detailURL := fmt.Sprintf("%s/detail?task_id=%d", baseURL, taskList[i].ID)
		if query != "" {
			detailURL += "&" + query
		}
		taskList[i].DetailURL = detailURL
		
		// 构建编辑链接
		editURL := fmt.Sprintf("%s/edit?task_id=%d", baseURL, taskList[i].ID)
		if query != "" {
			editURL += "&" + query
		}
		taskList[i].EditURL = editURL
		
		// 构建抽检链接
		sampleURL := fmt.Sprintf("%s/sample?task_id=%d", baseURL, taskList[i].ID)
		if query != "" {
			sampleURL += "&" + query
		}
		taskList[i].SampleURL = sampleURL
	}

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

	// 计算当前页记录范围
	startRecord := (page-1)*pageSize + 1
	endRecord := page * pageSize
	if endRecord > totalCount {
		endRecord = totalCount
	}
	if totalCount == 0 {
		startRecord = 0
		endRecord = 0
	}

	// 4. 准备数据并渲染模板
	// 检查权限（currentUser已在前面定义）
	canImport := true
	canDelete := true
	if currentUser != nil && currentUser.RoleCode != 0 {
		// 普通用户需要检查权限设置
		canImport = permission.CheckPermission(currentUser, "allow_checkpoint_audit_import")
		canDelete = permission.CheckPermission(currentUser, "allow_checkpoint_audit_delete")
	}

	data := PageData{
		Title:         "卡口审核进度",
		ActiveMenu:    "audit",
		SubMenu:       "checkpoint_progress",
		List:          taskList,
		SearchName:    searchName,
		AuditStatus:   auditStatus,
		ArchiveType:   archiveType,
		SampleStatus:  sampleStatus,
		CurrentPage:   page,
		TotalPages:    totalPages,
		TotalCount:    totalCount,
		StartRecord:   startRecord,
		EndRecord:     endRecord,
		HasPrev:       page > 1,
		HasNext:       page < totalPages,
		PrevPage:      page - 1,
		NextPage:      page + 1,
		FirstPage:     1,
		LastPage:      totalPages,
		Query:         query,
		QueryParams:   queryParams,
		CanImport:     canImport,
		CanDelete:     canDelete,
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

	// 检查权限
	currentUser := auth.GetCurrentUser(r)
	if currentUser == nil {
		http.Error(w, "未登录", http.StatusUnauthorized)
		return
	}
	
	// 如果不是管理员，检查权限设置
	if currentUser.RoleCode != 0 {
		if !permission.CheckPermission(currentUser, "allow_checkpoint_audit_import") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error": "权限不足，请联系管理员开通卡口审核进度档案导入权限"}`))
			return
		}
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
		"变更": true,
	}
	if !validArchiveTypes[archiveType] {
		http.Error(w, "档案类型值无效，必须是：新增、取推、补档案、变更", http.StatusBadRequest)
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
			
			// 检查是否是唯一约束错误（checkpoint_code字段）
			isUniqueErr, fieldName, fieldValue := checkUniqueConstraintError(execErr, "checkpoint_code")
			if isUniqueErr {
				logger.Errorf("卡口审核进度-导入失败，第%d行数据违反唯一约束: checkpoint_code=%s, 文件名: %s", i+1, fieldValue, fileHeader.Filename)
				
				// 返回JSON格式的错误信息，前端可以弹窗显示
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusBadRequest)
				errorResponse := map[string]interface{}{
					"error":      "唯一约束违反",
					"message":    fmt.Sprintf("第 %d 行数据违反唯一约束", i+1),
					"field":      fieldName,
					"fieldValue": fieldValue,
					"detail":     fmt.Sprintf("卡口编号 '%s' 已存在，不能重复导入", fieldValue),
				}
				json.NewEncoder(w).Encode(errorResponse)
				return
			}
			
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
		
		// 广播任务创建事件
		hub := auditprogress.GetEventHub()
		hub.BroadcastTaskCreated(int(taskID), map[string]interface{}{
			"file_name":    fileNameWithoutExt,
			"organization": organization,
			"record_count": importedCount,
		})
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

		// 获取抽检信息，判断是否显示"已整改"复选框
		sampleInfo, err := GetSampleInfo(taskID)
		if err != nil {
			logger.Errorf("查询抽检信息失败: %v, taskID: %d", err, taskID)
			// 不阻断流程，继续执行
			sampleInfo = &SampleInfo{}
		}
		
		// 判断是否显示"已整改"复选框
		// 条件：审核状态为"已完成"且最近一次抽检结果为"待整改"
		showFixedCheckbox := task.AuditStatus == "已完成" && sampleInfo.LastSampleResult == "待整改"

		// 获取筛选条件参数（用于返回链接）
		allParams := r.URL.Query()
		searchName := allParams.Get("file_name")
		auditStatus := allParams.Get("audit_status")
		archiveType := allParams.Get("archive_type")
		sampleStatus := allParams.Get("sample_status")
		
		// 构建查询参数字符串（使用解码后的值重新编码）
		queryValues := url.Values{}
		if searchName != "" {
			queryValues.Set("file_name", searchName)
		}
		if auditStatus != "" {
			queryValues.Set("audit_status", auditStatus)
		}
		if archiveType != "" {
			queryValues.Set("archive_type", archiveType)
		}
		if sampleStatus != "" {
			queryValues.Set("sample_status", sampleStatus)
		}
		query := queryValues.Encode()
		
		// 构建完整的返回URL（避免在模板中拼接导致HTML转义）
		backURL := "/checkpoint/progress"
		if query != "" {
			backURL += "?" + query
		}

		type EditPageData struct {
			Title            string
			ActiveMenu       string
			SubMenu          string
			Task             CheckpointTask
			Query            string            // 筛选条件查询字符串（保留用于兼容）
			QueryParams      map[string]string // 筛选条件参数的键值对
			BackURL          string            // 完整的返回URL
			ShowFixedCheckbox bool             // 是否显示"已整改"复选框
		}

		// 构建QueryParams map
		queryParams := make(map[string]string)
		if searchName != "" {
			queryParams["file_name"] = searchName
		}
		if auditStatus != "" {
			queryParams["audit_status"] = auditStatus
		}
		if archiveType != "" {
			queryParams["archive_type"] = archiveType
		}
		if sampleStatus != "" {
			queryParams["sample_status"] = sampleStatus
		}
		
		data := EditPageData{
			Title:            "编辑审核意见",
			ActiveMenu:       "audit",
			SubMenu:          "checkpoint_progress",
			Task:             task,
			Query:            query,
			QueryParams:      queryParams,
			BackURL:          backURL,
			ShowFixedCheckbox: showFixedCheckbox,
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
		isFixed := r.FormValue("is_fixed") == "1" // 是否勾选了"已整改"

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

		// 如果勾选了"已整改"，需要验证条件：审核状态为"已完成"且最近一次抽检结果为"待整改"
		if isFixed {
			// 查询当前审核状态和最近一次抽检结果
			var currentAuditStatus string
			var lastSampleResult sql.NullString
			checkSQL := `SELECT ct.audit_status, 
				(SELECT sample_result FROM checkpoint_sample_records WHERE task_id = ? ORDER BY sampled_at DESC LIMIT 1) AS last_sample_result
				FROM checkpoint_tasks ct WHERE ct.id = ?`
			err = db.DBInstance.QueryRow(checkSQL, taskID, taskID).Scan(&currentAuditStatus, &lastSampleResult)
			if err != nil {
				http.Error(w, "查询任务信息失败: "+err.Error(), http.StatusInternalServerError)
				return
			}
			
			// 验证条件
			if currentAuditStatus != "已完成" || !lastSampleResult.Valid || lastSampleResult.String != "待整改" {
				http.Error(w, "只有审核状态为'已完成'且最近一次抽检结果为'待整改'时才能标记整改完成", http.StatusBadRequest)
				return
			}
		}

		// 开始事务
		tx, err := db.DBInstance.Begin()
		if err != nil {
			http.Error(w, "事务开始失败", http.StatusInternalServerError)
			return
		}

		// 保存审核意见历史记录（如果内容有变化）
		currentUser := auth.GetCurrentUser(r)
		err = SaveAuditHistory(tx, taskID, auditComment, auditStatus, currentUser)
		if err != nil {
			if err == sql.ErrNoRows {
				tx.Rollback()
				http.Error(w, "档案不存在", http.StatusBadRequest)
				return
			}
			tx.Rollback()
			logger.Errorf("保存卡口审核意见历史失败: %v, taskID: %d", err, taskID)
			http.Error(w, "保存审核意见历史失败: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 更新审核意见和状态
		// 如果状态变为"已完成"，且completed_at为空，则设置completed_at为当前时间（只记录首次完成时间）
		updateSQL := `UPDATE checkpoint_tasks 
		              SET audit_comment = ?, 
		                  audit_status = ?, 
		                  updated_at = NOW(),
		                  completed_at = CASE 
		                      WHEN ? = '已完成' AND completed_at IS NULL 
		                      THEN NOW()
		                      ELSE completed_at
		                  END
		              WHERE id = ?`
		var comment interface{}
		if auditComment == "" {
			comment = nil
		} else {
			comment = auditComment
		}

		_, err = tx.Exec(updateSQL, comment, auditStatus, auditStatus, taskID)
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

		// 如果勾选了"已整改"，重置抽检状态
		if isFixed {
			resetSampleSQL := `UPDATE checkpoint_tasks 
				SET is_sampled = 0, last_sampled_at = NULL 
				WHERE id = ?`
			_, err = tx.Exec(resetSampleSQL, taskID)
			if err != nil {
				tx.Rollback()
				logger.Errorf("重置抽检状态失败: %v, taskID: %d", err, taskID)
				http.Error(w, "重置抽检状态失败: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			http.Error(w, "事务提交失败: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 记录操作日志
		if currentUser != nil {
			action := fmt.Sprintf("编辑卡口审核意见（档案ID：%d，状态：%s）", taskID, auditStatus)
			if isFixed {
				action += "，已标记整改完成"
			}
			operationlog.Record(r, currentUser.Username, action)
			
			// 广播任务更新事件
			hub := auditprogress.GetEventHub()
			updateData := map[string]interface{}{
				"audit_status":  auditStatus,
				"audit_comment": auditComment,
				"updated_by":    currentUser.Username,
			}
			logger.Infof("[CheckpointEditHandler] 广播任务更新事件: taskID=%d, audit_status=%s, audit_comment=%s", 
				taskID, auditStatus, auditComment)
			hub.BroadcastTaskUpdated(int(taskID), updateData)
		}

		// 获取筛选条件参数（用于返回链接）
		searchName := r.FormValue("file_name")
		auditStatusParam := r.FormValue("audit_status")
		archiveType := r.FormValue("archive_type")
		sampleStatus := r.FormValue("sample_status")
		
		// 构建查询参数字符串
		queryValues := url.Values{}
		if searchName != "" {
			queryValues.Set("file_name", searchName)
		}
		if auditStatusParam != "" {
			queryValues.Set("audit_status", auditStatusParam)
		}
		if archiveType != "" {
			queryValues.Set("archive_type", archiveType)
		}
		if sampleStatus != "" {
			queryValues.Set("sample_status", sampleStatus)
		}
		queryValues.Set("message", "EditSuccess")
		redirectURL := "/checkpoint/progress?" + queryValues.Encode()
		
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return
	}

	http.Error(w, "不支持的请求方法", http.StatusMethodNotAllowed)
}

// AuditHistoryHandler: 查看审核意见历史记录（返回JSON格式，用于弹窗显示）
func AuditHistoryHandler(w http.ResponseWriter, r *http.Request) {
	taskIDStr := r.URL.Query().Get("task_id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil || taskID <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success": false, "message": "无效的任务ID"}`))
		return
	}

	// 验证任务是否存在
	var exists int
	err = db.DBInstance.QueryRow("SELECT COUNT(*) FROM checkpoint_tasks WHERE id = ?", taskID).Scan(&exists)
	if err != nil || exists == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"success": false, "message": "档案不存在"}`))
		return
	}

	// 查询历史记录
	historyList, err := GetAuditHistory(taskID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success": false, "message": "查询审核意见历史失败"}`))
		return
	}

	// 转换为JSON格式（处理sql.NullString）
	type HistoryItemJSON struct {
		ID           int    `json:"id"`
		AuditComment string `json:"audit_comment"`
		AuditStatus  string `json:"audit_status"`
		Auditor      string `json:"auditor"`
		AuditTime    string `json:"audit_time"`
	}

	var historyJSON []HistoryItemJSON
	for _, item := range historyList {
		comment := ""
		if item.AuditComment.Valid {
			comment = item.AuditComment.String
		}
		historyJSON = append(historyJSON, HistoryItemJSON{
			ID:           item.ID,
			AuditComment: comment,
			AuditStatus:  item.AuditStatus,
			Auditor:      item.Auditor,
			AuditTime:    item.AuditTime,
		})
	}

	// 返回JSON
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	response := map[string]interface{}{
		"success": true,
		"data":    historyJSON,
	}
	
	jsonData, err := json.Marshal(response)
	if err != nil {
		logger.Errorf("JSON序列化失败: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success": false, "message": "数据序列化失败"}`))
		return
	}

	w.Write(jsonData)
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

// 检测是否是唯一约束错误，并提取违反约束的字段值
// 返回 (isUniqueError, fieldName, fieldValue)
func checkUniqueConstraintError(err error, fieldName string) (bool, string, string) {
	if err == nil {
		return false, "", ""
	}

	// 检查是否是MySQL错误
	mysqlErr, ok := err.(*mysql.MySQLError)
	if !ok {
		return false, "", ""
	}

	// MySQL唯一约束错误代码是1062
	if mysqlErr.Number != 1062 {
		return false, "", ""
	}

	// 从错误信息中提取字段值
	// 错误信息格式: "Duplicate entry 'xxx' for key 'uk_checkpoint_code'"
	// 使用正则表达式提取 'xxx' 部分
	re := regexp.MustCompile(`Duplicate entry '([^']+)' for key`)
	matches := re.FindStringSubmatch(mysqlErr.Message)
	if len(matches) >= 2 {
		return true, fieldName, matches[1]
	}

	return true, fieldName, ""
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

	// 获取筛选条件参数（用于返回链接）
	allParams := r.URL.Query()
	searchName := allParams.Get("file_name")
	auditStatus := allParams.Get("audit_status")
	archiveType := allParams.Get("archive_type")
	sampleStatus := allParams.Get("sample_status")
	
	// 构建查询参数字符串（使用解码后的值重新编码）
	queryValues := url.Values{}
	if searchName != "" {
		queryValues.Set("file_name", searchName)
	}
	if auditStatus != "" {
		queryValues.Set("audit_status", auditStatus)
	}
	if archiveType != "" {
		queryValues.Set("archive_type", archiveType)
	}
	if sampleStatus != "" {
		queryValues.Set("sample_status", sampleStatus)
	}
	query := queryValues.Encode()
	
	// 构建完整的返回URL（避免在模板中拼接导致HTML转义）
	backURL := "/checkpoint/progress"
	if query != "" {
		backURL += "?" + query
	}

	// 准备数据
	type DetailPageData struct {
		Title         string
		ActiveMenu    string
		SubMenu       string
		Task          CheckpointTask
		Details       []DetailItem
		Query         string // 筛选条件查询字符串（保留用于兼容）
		BackURL       string // 完整的返回URL
	}

	data := DetailPageData{
		Title:      "档案明细",
		ActiveMenu: "audit",
		SubMenu:    "checkpoint_progress",
		Task:       task,
		Details:    detailList,
		Query:      query,
		BackURL:    backURL,
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
		terminal_mac_address, collection_area_type, integrated_command_platform_checkpoint_code
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
			&item.IntegratedCommandPlatformCheckpointCode,
		)
		if err != nil {
			continue
		}

		// 构建行数据
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
			item.IntegratedCommandPlatformCheckpointCode.String,
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

	// 检查权限
	currentUser := auth.GetCurrentUser(r)
	if currentUser == nil {
		http.Error(w, "未登录", http.StatusUnauthorized)
		return
	}
	
	// 如果不是管理员，检查权限设置
	if currentUser.RoleCode != 0 {
		if !permission.CheckPermission(currentUser, "allow_checkpoint_audit_delete") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error": "权限不足，请联系管理员开通卡口审核进度档案删除权限"}`))
			return
		}
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

	// 按顺序删除关联记录（先删除子表，再删除父表）
	// 1. 删除审核历史记录
	deleteHistorySQL := "DELETE FROM checkpoint_audit_history WHERE task_id = ?"
	_, err = tx.Exec(deleteHistorySQL, taskID)
	if err != nil {
		tx.Rollback()
		logger.Errorf("删除审核历史记录失败: %v, taskID: %d", err, taskID)
		http.Error(w, "删除审核历史记录失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. 删除抽检记录
	deleteSampleSQL := "DELETE FROM checkpoint_sample_records WHERE task_id = ?"
	_, err = tx.Exec(deleteSampleSQL, taskID)
	if err != nil {
		tx.Rollback()
		logger.Errorf("删除抽检记录失败: %v, taskID: %d", err, taskID)
		http.Error(w, "删除抽检记录失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. 删除关联的明细记录
	deleteDetailsSQL := "DELETE FROM checkpoint_details WHERE task_id = ?"
	_, err = tx.Exec(deleteDetailsSQL, taskID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "删除档案明细失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. 最后删除任务记录
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
	currentUser2 := auth.GetCurrentUser(r)
	if currentUser2 != nil {
		action := fmt.Sprintf("删除卡口审核档案（档案名称：%s，机构：%s，包含 %d 条明细）", task.FileName, task.Organization, detailCount)
		operationlog.Record(r, currentUser2.Username, action)
		
		// 广播任务删除事件
		hub := auditprogress.GetEventHub()
		hub.BroadcastTaskDeleted(int(taskID))
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

// UploadHandler 处理文件上传（支持单个文件和目录上传）
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

	// 设置响应头为JSON格式
	w.Header().Set("Content-Type", "application/json")

	// 获取task_id和档案名称
	taskIDStr := r.FormValue("task_id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := map[string]interface{}{
			"success": false,
			"message": "无效的任务ID",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// 查询档案名称
	var fileName string
	query := "SELECT file_name FROM checkpoint_tasks WHERE id = ?"
	err = db.DBInstance.QueryRow(query, taskID).Scan(&fileName)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			response := map[string]interface{}{
				"success": false,
				"message": "档案不存在",
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		logger.Errorf("卡口审核进度-查询档案名称失败: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		response := map[string]interface{}{
			"success": false,
			"message": "查询档案失败",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// 去除档案名称的前后空格，避免路径问题
	fileName = strings.TrimSpace(fileName)

	// 获取任务配置中的上传路径
	uploadPath := getUploadPath()
	if uploadPath == "" {
		w.WriteHeader(http.StatusInternalServerError)
		response := map[string]interface{}{
			"success": false,
			"message": "未配置上传路径，请先在任务配置中设置",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// 创建以档案名称命名的目录（去除前后空格）
	targetDir := filepath.Join(uploadPath, fileName)
	err = createDirIfNotExists(targetDir)
	if err != nil {
		logger.Errorf("卡口审核进度-创建目录失败: %v, 路径: %s", err, targetDir)
		w.WriteHeader(http.StatusInternalServerError)
		response := map[string]interface{}{
			"success": false,
			"message": "创建目录失败: " + err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// 限制常量
	const maxTotalSize = 200 * 1024 * 1024 // 200MB
	const maxFileCount = 20

	// 解析multipart表单
	err = r.ParseMultipartForm(32 << 20) // 32MB内存缓冲区
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := map[string]interface{}{
			"success": false,
			"message": "解析表单失败: " + err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// 获取所有上传的文件
	files := r.MultipartForm.File["upload_file"]
	if len(files) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		response := map[string]interface{}{
			"success": false,
			"message": "请选择要上传的文件",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// 检查文件数量限制
	if len(files) > maxFileCount {
		w.WriteHeader(http.StatusBadRequest)
		response := map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("文件数量超过限制，最多允许上传 %d 个文件，当前选择了 %d 个", maxFileCount, len(files)),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// 检查总大小限制
	var totalSize int64
	for _, fileHeader := range files {
		totalSize += fileHeader.Size
	}
	if totalSize > maxTotalSize {
		w.WriteHeader(http.StatusBadRequest)
		response := map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("总文件大小超过限制，最大允许 %d MB，当前总大小为 %.2f MB", maxTotalSize/(1024*1024), float64(totalSize)/(1024*1024)),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// 处理每个文件
	var uploadedFiles []string
	var failedFiles []string
	totalFiles := len(files)

	for _, fileHeader := range files {
		// 打开文件
		file, err := fileHeader.Open()
		if err != nil {
			logger.Errorf("卡口审核进度-打开文件失败: %v, 文件名: %s", err, fileHeader.Filename)
			failedFiles = append(failedFiles, fileHeader.Filename)
			continue
		}

		// 处理文件路径（支持目录结构，保留相对路径）
		// 浏览器上传目录时，文件名可能包含相对路径，如 "subdir/file.txt"
		filePath := fileHeader.Filename
		// 清理路径，防止路径遍历攻击
		filePath = filepath.Clean(filePath)
		if strings.HasPrefix(filePath, "..") || strings.HasPrefix(filePath, "/") {
			logger.Errorf("卡口审核进度-非法文件路径: %s", filePath)
			file.Close()
			failedFiles = append(failedFiles, fileHeader.Filename)
			continue
		}

		// 构建目标文件路径
		targetFile := filepath.Join(targetDir, filePath)
		
		// 确保目标目录存在（如果文件在子目录中）
		targetFileDir := filepath.Dir(targetFile)
		err = createDirIfNotExists(targetFileDir)
		if err != nil {
			logger.Errorf("卡口审核进度-创建子目录失败: %v, 路径: %s", err, targetFileDir)
			file.Close()
			failedFiles = append(failedFiles, fileHeader.Filename)
			continue
		}

		// 创建目标文件
		dst, err := os.Create(targetFile)
		if err != nil {
			logger.Errorf("卡口审核进度-创建文件失败: %v, 路径: %s", err, targetFile)
			file.Close()
			failedFiles = append(failedFiles, fileHeader.Filename)
			continue
		}

		// 复制文件内容
		_, err = io.Copy(dst, file)
		dst.Close()
		file.Close()

		if err != nil {
			logger.Errorf("卡口审核进度-写入文件失败: %v, 路径: %s", err, targetFile)
			failedFiles = append(failedFiles, fileHeader.Filename)
			continue
		}

		uploadedFiles = append(uploadedFiles, fileHeader.Filename)
	}

	// 构建操作日志
	var action string
	if len(uploadedFiles) == 1 {
		action = fmt.Sprintf("上传文件（档案：%s，文件名：%s）", fileName, uploadedFiles[0])
	} else {
		action = fmt.Sprintf("上传目录/文件（档案：%s，成功：%d/%d）", fileName, len(uploadedFiles), totalFiles)
		if len(failedFiles) > 0 {
			action += fmt.Sprintf("，失败：%d", len(failedFiles))
		}
	}
	operationlog.Record(r, currentUser.Username, action)

	// 返回响应
	w.WriteHeader(http.StatusOK)
	response := map[string]interface{}{
		"success":       true,
		"message":       fmt.Sprintf("上传完成！成功：%d/%d", len(uploadedFiles), totalFiles),
		"total":         totalFiles,
		"successCount":  len(uploadedFiles),
		"failedCount":   len(failedFiles),
		"uploadedFiles": uploadedFiles,
		"failedFiles":   failedFiles,
	}
	json.NewEncoder(w).Encode(response)
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

// SampleHandler: 抽检操作（POST提交抽检，GET显示抽检表单）
func SampleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// 显示抽检表单页面
		taskIDStr := r.URL.Query().Get("task_id")
		taskID, err := strconv.Atoi(taskIDStr)
		if err != nil || taskID <= 0 {
			http.Error(w, "无效的任务ID", http.StatusBadRequest)
			return
		}

		// 查询任务信息
		var task CheckpointTask
		taskSQL := "SELECT id, file_name, organization, audit_status FROM checkpoint_tasks WHERE id = ?"
		err = db.DBInstance.QueryRow(taskSQL, taskID).Scan(
			&task.ID,
			&task.FileName,
			&task.Organization,
			&task.AuditStatus,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "档案不存在", http.StatusNotFound)
			} else {
				http.Error(w, "查询档案失败: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// 验证审核状态必须是"已完成"
		if task.AuditStatus != "已完成" {
			http.Error(w, "只有审核状态为'已完成'的任务才能进行抽检", http.StatusBadRequest)
			return
		}

		// 获取当前用户
		currentUser := auth.GetCurrentUser(r)
		if currentUser == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// 获取筛选条件参数（用于返回链接）
		allParams := r.URL.Query()
		searchName := allParams.Get("file_name")
		auditStatus := allParams.Get("audit_status")
		archiveType := allParams.Get("archive_type")
		sampleStatus := allParams.Get("sample_status")
		
		// 构建查询参数字符串（使用解码后的值重新编码）
		queryValues := url.Values{}
		if searchName != "" {
			queryValues.Set("file_name", searchName)
		}
		if auditStatus != "" {
			queryValues.Set("audit_status", auditStatus)
		}
		if archiveType != "" {
			queryValues.Set("archive_type", archiveType)
		}
		if sampleStatus != "" {
			queryValues.Set("sample_status", sampleStatus)
		}
		query := queryValues.Encode()
		
		// 构建完整的返回URL（避免在模板中拼接导致HTML转义）
		backURL := "/checkpoint/progress"
		if query != "" {
			backURL += "?" + query
		}

		type SamplePageData struct {
			Title      string
			ActiveMenu string
			SubMenu    string
			Task       CheckpointTask
			SampledBy  string
			Query      string            // 筛选条件查询字符串（保留用于兼容）
			QueryParams map[string]string // 筛选条件参数的键值对
			BackURL    string            // 完整的返回URL
		}

		// 构建QueryParams map
		queryParams := make(map[string]string)
		if searchName != "" {
			queryParams["file_name"] = searchName
		}
		if auditStatus != "" {
			queryParams["audit_status"] = auditStatus
		}
		if archiveType != "" {
			queryParams["archive_type"] = archiveType
		}
		if sampleStatus != "" {
			queryParams["sample_status"] = sampleStatus
		}
		
		data := SamplePageData{
			Title:      "抽检",
			ActiveMenu: "audit",
			SubMenu:    "checkpoint_progress",
			Task:       task,
			SampledBy:  currentUser.Username,
			Query:      query,
			QueryParams: queryParams,
			BackURL:    backURL,
		}

		tmpl, err := template.ParseFiles("templates/checkpointsample.html")
		if err != nil {
			logger.Errorf("抽检-模板解析失败: %v", err)
			http.Error(w, "模板解析失败: "+err.Error(), http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			logger.Errorf("抽检-模板渲染失败: %v", err)
			http.Error(w, "模板渲染失败: "+err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	// POST: 保存抽检记录
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

		// 验证任务是否存在且状态为"已完成"
		var auditStatus string
		checkSQL := "SELECT audit_status FROM checkpoint_tasks WHERE id = ?"
		err = db.DBInstance.QueryRow(checkSQL, taskID).Scan(&auditStatus)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "档案不存在", http.StatusNotFound)
			} else {
				http.Error(w, "查询档案失败: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		if auditStatus != "已完成" {
			http.Error(w, "只有审核状态为'已完成'的任务才能进行抽检", http.StatusBadRequest)
			return
		}

		// 获取当前用户
		currentUser := auth.GetCurrentUser(r)
		if currentUser == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		sampleComment := strings.TrimSpace(r.FormValue("sample_comment"))
		sampleResult := strings.TrimSpace(r.FormValue("sample_result"))

		// 验证抽检结果
		validResults := map[string]bool{
			"通过":   true,
			"待整改": true,
		}
		if sampleResult != "" && !validResults[sampleResult] {
			http.Error(w, "无效的抽检结果", http.StatusBadRequest)
			return
		}

		// 保存抽检记录
		err = SaveSampleRecord(taskID, currentUser.Username, sampleComment, sampleResult)
		if err != nil {
			logger.Errorf("保存抽检记录失败: %v, taskID: %d", err, taskID)
			http.Error(w, "保存抽检记录失败: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 记录操作日志
		action := fmt.Sprintf("抽检卡口审核档案（档案ID：%d，结果：%s）", taskID, sampleResult)
		operationlog.Record(r, currentUser.Username, action)
		
		// 广播任务抽检事件
		hub := auditprogress.GetEventHub()
		hub.BroadcastTaskSampled(int(taskID), map[string]interface{}{
			"sample_result":  sampleResult,
			"sample_comment": sampleComment,
			"sampled_by":     currentUser.Username,
		})

		// 获取筛选条件参数（用于返回链接）
		searchName := r.FormValue("file_name")
		auditStatusParam := r.FormValue("audit_status")
		archiveType := r.FormValue("archive_type")
		sampleStatus := r.FormValue("sample_status")
		
		// 构建查询参数字符串
		queryValues := url.Values{}
		if searchName != "" {
			queryValues.Set("file_name", searchName)
		}
		if auditStatusParam != "" {
			queryValues.Set("audit_status", auditStatusParam)
		}
		if archiveType != "" {
			queryValues.Set("archive_type", archiveType)
		}
		if sampleStatus != "" {
			queryValues.Set("sample_status", sampleStatus)
		}
		queryValues.Set("message", "SampleSuccess")
		redirectURL := "/checkpoint/progress?" + queryValues.Encode()
		
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return
	}

	http.Error(w, "不支持的请求方法", http.StatusMethodNotAllowed)
}

// SampleHistoryHandler: 查看抽检历史记录（返回JSON格式，用于弹窗显示）
func SampleHistoryHandler(w http.ResponseWriter, r *http.Request) {
	taskIDStr := r.URL.Query().Get("task_id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil || taskID <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success": false, "message": "无效的任务ID"}`))
		return
	}

	// 验证任务是否存在
	var exists int
	err = db.DBInstance.QueryRow("SELECT COUNT(*) FROM checkpoint_tasks WHERE id = ?", taskID).Scan(&exists)
	if err != nil || exists == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"success": false, "message": "档案不存在"}`))
		return
	}

	// 查询抽检历史记录
	historyList, err := GetSampleHistory(taskID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success": false, "message": "查询抽检历史失败"}`))
		return
	}

	// 转换为JSON格式（处理sql.NullString）
	type HistoryItemJSON struct {
		ID            int    `json:"id"`
		SampledBy     string `json:"sampled_by"`
		SampledAt     string `json:"sampled_at"`
		SampleComment string `json:"sample_comment"`
		SampleResult  string `json:"sample_result"`
		CreatedAt     string `json:"created_at"`
	}

	var historyJSON []HistoryItemJSON
	for _, item := range historyList {
		comment := ""
		if item.SampleComment.Valid {
			comment = item.SampleComment.String
		}
		result := ""
		if item.SampleResult.Valid {
			result = item.SampleResult.String
		}
		historyJSON = append(historyJSON, HistoryItemJSON{
			ID:            item.ID,
			SampledBy:     item.SampledBy,
			SampledAt:     item.SampledAt,
			SampleComment: comment,
			SampleResult:  result,
			CreatedAt:     item.CreatedAt,
		})
	}

	// 返回JSON
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	response := map[string]interface{}{
		"success": true,
		"data":    historyJSON,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		logger.Errorf("JSON序列化失败: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success": false, "message": "数据序列化失败"}`))
		return
	}

	w.Write(jsonData)
}
