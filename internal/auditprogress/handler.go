package auditprogress

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"ops-web/internal/auth"
	"ops-web/internal/db"
	"ops-web/internal/filelist"
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
type AuditTask struct {
	ID          int
	FileName    string // 档案名称
	Organization string // 机构名称
	ImportTime  string // 导入时间
	AuditStatus string // 审核状态
	RecordCount int    // 导入记录数量
	AuditComment sql.NullString // 审核意见
	UpdatedAt   string // 审核时间（updated_at）
	Attachments []string // 附件列表
	IsSingleSoldier int // 是否单兵设备：0=否，1=是
	ArchiveType sql.NullString // 档案类型：新增、取推、补档案
	// 抽检相关字段
	IsSampled       bool   // 是否已抽检
	LastSampledAt   string // 最后抽检时间
	SampledBy       string // 最后抽检人员
	SampleCount     int    // 抽检次数
	LastSampleResult string // 最近一次抽检结果
	// 提醒相关字段
	ReminderCount   int    // 待处理提醒数量
}

// 页面数据结构体
type PageData struct {
	Title         string
	ActiveMenu    string
	SubMenu       string
	List          []AuditTask
	SearchName    string
	AuditStatus   string // 审核状态查询条件
	ArchiveType   string // 建档类型查询条件
	SampleStatus  string // 抽检状态查询条件：全部、已抽检、未抽检
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
	ImportMessage string
	ImportCount   int
	HighlightTaskID int // 需要高亮的任务ID（用于从提醒页面跳转过来时定位）
}

// Handler: 审核进度列表页 (GET)
func Handler(w http.ResponseWriter, r *http.Request) {
	// 获取查询参数
	searchName := r.URL.Query().Get("file_name")
	auditStatus := r.URL.Query().Get("audit_status") // 审核状态: 未审核, 已审核待整改, 已完成 或空
	archiveType := r.URL.Query().Get("archive_type") // 建档类型: 新增, 取推, 补档案 或空
	sampleStatus := r.URL.Query().Get("sample_status") // 抽检状态: 全部, 已抽检, 未抽检 或空
	taskIDStr := r.URL.Query().Get("task_id") // 任务ID，用于定位到特定任务
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
			"未抽检": true,
		}
		if validSampleStatuses[sampleStatus] {
			if sampleStatus == "已抽检" {
				whereSQL += " AND is_sampled = 1"
			} else if sampleStatus == "未抽检" {
				whereSQL += " AND (is_sampled = 0 OR is_sampled IS NULL)"
			}
		}
	}

	// 1. 查询总记录数
	var totalCount int
	countSQL := "SELECT COUNT(*) FROM audit_tasks" + whereSQL
	err := db.DBInstance.QueryRow(countSQL, args...).Scan(&totalCount)
	if err != nil && err != sql.ErrNoRows {
		logger.Errorf("审核进度-查询总数失败: %v, SQL: %s, Args: %v", err, countSQL, args)
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
	querySQL := fmt.Sprintf("SELECT id, file_name, organization, import_time, audit_status, record_count, audit_comment, updated_at, is_single_soldier, archive_type, is_sampled, last_sampled_at FROM audit_tasks %s ORDER BY id DESC LIMIT ? OFFSET ?", whereSQL)

	// 准备完整的参数列表
	queryArgs := append(args, pageSize, offset)

	rows, err := db.DBInstance.Query(querySQL, queryArgs...)
	if err != nil {
		logger.Errorf("审核进度-数据库查询失败: %v, SQL: %s, Args: %v", err, querySQL, queryArgs)
		http.Error(w, "数据库查询失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var taskList []AuditTask
	var taskIDs []int
	for rows.Next() {
		var task AuditTask
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
			&task.IsSingleSoldier,
			&task.ArchiveType,
			&isSampled,
			&lastSampledAtRaw,
		)

		if err != nil {
			fmt.Printf("❌ Database scan error: %v\n", err)
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
		logger.Errorf("审核进度-数据遍历失败: %v", err)
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

	// 批量获取提醒数量
	reminderCountMap := make(map[int]int)
	for _, taskID := range taskIDs {
		count, err := GetReminderCountByTaskID(taskID)
		if err == nil {
			reminderCountMap[taskID] = count
		}
	}

	// 填充提醒数量到任务列表
	for i := range taskList {
		if count, ok := reminderCountMap[taskList[i].ID]; ok {
			taskList[i].ReminderCount = count
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
	if sampleStatus != "" {
		queryParams = append(queryParams, "sample_status="+sampleStatus)
	}
	query := strings.Join(queryParams, "&")

	// 记录查询操作日志（如果有查询条件）
	currentUser := auth.GetCurrentUser(r)
	if currentUser != nil && (searchName != "" || auditStatus != "" || archiveType != "") {
		action := "查询审核进度"
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
	// 如果提供了task_id，需要找到该任务所在的页面
	var highlightTaskID int
	if taskIDStr != "" {
		taskID, err := strconv.Atoi(taskIDStr)
		if err == nil && taskID > 0 {
			// 检查任务是否在当前页面的任务列表中
			found := false
			for _, task := range taskList {
				if task.ID == taskID {
					found = true
					highlightTaskID = taskID
					break
				}
			}
			// 如果不在当前页，需要找到任务所在的页面
			if !found {
				var taskPage int
				var taskExists int
				err = db.DBInstance.QueryRow("SELECT COUNT(*) FROM audit_tasks WHERE id = ?", taskID).Scan(&taskExists)
				if err == nil && taskExists > 0 {
					// 计算任务所在的页面（按ID排序，找到任务的位置）
					var taskPosition int
					err = db.DBInstance.QueryRow(`
						SELECT COUNT(*) + 1 FROM audit_tasks 
						WHERE id < ? AND (`+strings.TrimPrefix(whereSQL, " WHERE 1=1")+`)`,
						append([]interface{}{taskID}, args...)...).Scan(&taskPosition)
					if err == nil {
						taskPage = (taskPosition-1)/pageSize + 1
						if taskPage != page {
							// 重定向到任务所在的页面
							redirectURL := fmt.Sprintf("/audit/progress?page=%d&task_id=%d", taskPage, taskID)
							if searchName != "" {
								redirectURL += "&file_name=" + searchName
							}
							if auditStatus != "" {
								redirectURL += "&audit_status=" + auditStatus
							}
							if archiveType != "" {
								redirectURL += "&archive_type=" + archiveType
							}
							if sampleStatus != "" {
								redirectURL += "&sample_status=" + sampleStatus
							}
							http.Redirect(w, r, redirectURL, http.StatusSeeOther)
							return
						}
					}
				}
			} else {
				highlightTaskID = taskID
			}
		}
	}

	data := PageData{
		Title:         "审核进度",
		ActiveMenu:    "audit",
		SubMenu:       "audit_progress",
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
		ImportMessage: importMsg,
		ImportCount:   importCount,
		HighlightTaskID: highlightTaskID, // 需要高亮的任务ID
	}

	tmpl, err := template.ParseFiles("templates/auditprogress.html")
	if err != nil {
		logger.Errorf("审核进度-模板解析失败: %v", err)
		http.Error(w, "模板解析失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		logger.Errorf("审核进度-模板渲染失败: %v", err)
		http.Error(w, "模板渲染失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// ImportHandler: 导入 XLSX 档案
func ImportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/audit/progress", http.StatusSeeOther)
		return
	}

	// 解析表单数据
	err := r.ParseMultipartForm(32 << 20) // 32MB
	if err != nil {
		logger.Errorf("审核进度-表单解析失败: %v", err)
		http.Error(w, "表单解析失败", http.StatusBadRequest)
		return
	}

	// 获取机构名称（必填字段）
	organization := strings.TrimSpace(r.FormValue("organization"))
	if organization == "" {
		http.Error(w, "机构名称不能为空", http.StatusBadRequest)
		return
	}

	// 获取上传的文件
	file, fileHeader, err := r.FormFile("upload_file")
	if err != nil {
		logger.Errorf("审核进度-文件上传失败: %v", err)
		http.Error(w, "文件上传失败", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 获取文件名（不含路径）
	fileName := fileHeader.Filename
	// 去除扩展名，只保留文件名部分作为档案名称，并去除前后空格
	fileNameWithoutExt := strings.TrimSpace(strings.TrimSuffix(fileName, filepath.Ext(fileName)))

	// 解析Excel文件
	f, err := excelize.OpenReader(file)
	if err != nil {
		logger.Errorf("审核进度-Excel解析失败: %v, 文件名: %s", err, fileHeader.Filename)
		http.Error(w, "Excel解析失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	sheetName := f.GetSheetName(0)

	rows, err := f.GetRows(sheetName)
	if err != nil {
		logger.Errorf("审核进度-数据读取失败: %v, 文件名: %s", err, fileHeader.Filename)
		http.Error(w, "数据读取失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(rows) < 2 {
		http.Error(w, "Excel文件至少需要包含表头和数据行", http.StatusBadRequest)
		return
	}

	// 获取表单数据
	isSingleSoldier := 0
	if r.FormValue("is_single_soldier") == "1" {
		isSingleSoldier = 1
	}
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

	// 开始事务
	tx, err := db.DBInstance.Begin()
	if err != nil {
		http.Error(w, "事务开始失败", http.StatusInternalServerError)
		return
	}

	// 1. 创建审核任务记录
	insertTaskSQL := `INSERT INTO audit_tasks (file_name, organization, import_time, audit_status, is_single_soldier, archive_type) VALUES (?, ?, ?, ?, ?, ?)`
	result, err := tx.Exec(insertTaskSQL, fileNameWithoutExt, organization, time.Now(), "未审核", isSingleSoldier, archiveType)
	if err != nil {
		tx.Rollback()
		logger.Errorf("审核进度-创建审核任务失败: %v, SQL: %s", err, insertTaskSQL)
		http.Error(w, "创建审核任务失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 获取插入的任务ID
	taskID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		logger.Errorf("审核进度-获取任务ID失败: %v", err)
		http.Error(w, "获取任务ID失败", http.StatusInternalServerError)
		return
	}

	// 2. 导入Excel数据到audit_details表
	// 这里需要根据audit_details表结构来插入数据
	// 参考filelist的导入逻辑，但字段映射需要根据audit_details表结构调整
	insertDetailFields := []string{
		"task_id", "device_code", "original_device_code", "device_name", "division_code", "monitor_point_type",
		"pickup", "parent_device", "construction_unit", "construction_unit_code", "management_unit",
		"camera_dept", "admin_name", "admin_contact", "contractor", "maintain_unit", "device_vendor",
		"device_model", "camera_type", "access_method", "camera_function_type", "video_encoding_format",
		"image_resolution", "camera_light_property", "backend_structure", "lens_type", "installation_type",
		"height_type", "jurisdiction_police", "installation_address", "surrounding_landmark", "longitude",
		"latitude", "installation_location", "monitoring_direction", "pole_number", "scene_picture",
		"networking_property", "access_network", "ipv4_address", "ipv6_address", "mac_address",
		"access_port", "associated_encoder", "device_username", "device_password", "channel_number",
		"connection_protocol", "enabled_time", "scrapped_time", "device_status", "inspection_status",
		"video_loss", "color_distortion", "video_blur", "brightness_exception", "video_interference",
		"video_lag", "video_occlusion", "scene_change", "online_duration", "offline_duration",
		"signaling_delay", "video_stream_delay", "key_frame_delay", "recording_retention_days",
		"storage_device_code", "storage_channel_number", "storage_type", "cache_settings", "notes",
		"collection_area_type",
	}

	insertDetailSQL := fmt.Sprintf("INSERT INTO audit_details (%s) VALUES (%s)",
		strings.Join(insertDetailFields, ", "),
		strings.TrimRight(strings.Repeat("?,", len(insertDetailFields)), ","))

	stmt, err := tx.Prepare(insertDetailSQL)
	if err != nil {
		tx.Rollback()
		logger.Errorf("审核进度-SQL Prepare失败: %v, SQL: %s", err, insertDetailSQL)
		http.Error(w, "SQL Prepare失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	importedCount := 0
	const expectedCols = 73 // Excel总共73列

	// 跳过表头，从第2行开始
	for i, row := range rows {
		if i == 0 {
			continue // 跳过表头
		}

		// 强制对齐到73列
		for len(row) < expectedCols {
			row = append(row, "")
		}

		// 数据类型转换（Excel列：A列是序号，B列开始是数据）
		// filelist中row[31]对应longitude，row[32]对应latitude，row[64]对应recording_retention_days
		lon, _ := strconv.ParseFloat(getRowValue(row, 31), 64)  // longitude列 (AF列)
		lat, _ := strconv.ParseFloat(getRowValue(row, 32), 64)  // latitude列 (AG列)
		recDays, _ := strconv.Atoi(getRowValue(row, 64))        // recording_retention_days列 (BS列)

		// 准备参数（task_id + 71个字段）
		params := make([]interface{}, len(insertDetailFields))
		params[0] = taskID // task_id

		// 映射Excel列到数据库字段
		// Excel结构: A列(id序号), B列(device_code), C列(original_device_code)...
		// row[0]=A列, row[1]=B列, row[31]=AF列(longitude), row[32]=AG列(latitude)...
		// params[0]=task_id, params[1]=device_code(B列row[1]), params[2]=original_device_code(C列row[2])...
		// 所以params[j]对应row[j]，j从1开始（因为params[0]是task_id）
		for j := 1; j < len(insertDetailFields); j++ {
			excelIdx := j // Excel列索引（B列=row[1], C列=row[2]...，与filelist一致）
			
			switch excelIdx {
			case 31: // longitude (Excel列32，row[31])
				params[j] = lon
			case 32: // latitude (Excel列33，row[32])
				params[j] = lat
			case 64: // recording_retention_days (Excel列65，row[64])
				params[j] = recDays
			case 48: // enabled_time (Excel列49，row[48])
				timeStr := strings.TrimSpace(getRowValue(row, excelIdx))
				if timeStr == "" {
					params[j] = nil
				} else {
					parsedTime, err := time.Parse(time.RFC3339, timeStr)
					if err == nil {
						params[j] = parsedTime.Format("2006-01-02 15:04:05")
					} else {
						params[j] = timeStr
					}
				}
			default:
				// 必填项列表（对应Excel列号，从B列开始，即row索引）
				requiredCols := []int{1, 3, 4, 5, 8, 9, 10, 11, 12, 13, 14, 15, 16, 18, 20, 21, 22, 27, 28, 29, 30, 33, 34, 35, 38, 39, 41, 50, 71}
				isRequired := false
				for _, col := range requiredCols {
					if excelIdx == col {
						isRequired = true
						break
					}
				}
				params[j] = toDBValue(getRowValue(row, excelIdx), isRequired)
			}
		}

		_, execErr := stmt.Exec(params...)
		if execErr != nil {
			tx.Rollback()
			logger.Errorf("审核进度-导入失败，第%d行数据错误: %v, 文件名: %s", i+1, execErr, fileHeader.Filename)
			errMsg := fmt.Sprintf("导入失败：第 %d 行数据错误。详细信息: %v", i+1, execErr)
			http.Error(w, errMsg, http.StatusInternalServerError)
			return
		}
		importedCount++
	}

	// 3. 更新任务表的记录数量
	updateTaskSQL := `UPDATE audit_tasks SET record_count = ? WHERE id = ?`
	_, err = tx.Exec(updateTaskSQL, importedCount, taskID)
	if err != nil {
		tx.Rollback()
		logger.Errorf("审核进度-更新记录数量失败: %v, SQL: %s", err, updateTaskSQL)
		http.Error(w, "更新记录数量失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		logger.Errorf("审核进度-数据库提交失败: %v, 文件名: %s", err, fileHeader.Filename)
		http.Error(w, "数据库提交失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 记录导入操作日志
	if currentUser := auth.GetCurrentUser(r); currentUser != nil {
		action := fmt.Sprintf("导入审核档案 Excel（档案名称：%s，机构：%s，是否单兵设备：%d，档案类型：%s，共 %d 条数据）", fileNameWithoutExt, organization, isSingleSoldier, archiveType, importedCount)
		operationlog.Record(r, currentUser.Username, action)
	}

	http.Redirect(w, r, "/audit/progress?message=ImportSuccess&count="+strconv.Itoa(importedCount), http.StatusSeeOther)
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

		var task AuditTask
		var importTimeRaw sql.NullString
		taskSQL := "SELECT id, file_name, organization, import_time, audit_status, audit_comment FROM audit_tasks WHERE id = ?"
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
			Task       AuditTask
		}

		data := EditPageData{
			Title:      "编辑审核意见",
			ActiveMenu: "audit",
			SubMenu:    "audit_progress",
			Task:       task,
		}

		tmpl, err := template.ParseFiles("templates/auditedit.html")
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
			logger.Errorf("保存审核意见历史失败: %v, taskID: %d", err, taskID)
			http.Error(w, "保存审核意见历史失败: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 更新审核意见和状态
		updateSQL := `UPDATE audit_tasks SET audit_comment = ?, audit_status = ?, updated_at = NOW() WHERE id = ?`
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
		// 未审核 -> 0, 已审核待整改 -> 1, 已完成 -> 2
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

		updateDetailSQL := `UPDATE audit_details SET audit_status = ? WHERE task_id = ?`
		_, err = tx.Exec(updateDetailSQL, detailStatus, taskID)
		if err != nil {
			tx.Rollback()
			http.Error(w, "更新明细状态失败: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 解析审核意见，检查是否有录像天数不足的情况
		if auditComment != "" {
			earliestDate, requiredDays, found := ParseVideoDaysIssue(auditComment)
			if found {
				logger.Errorf("解析到录像天数不足信息: taskID=%d, earliestDate=%s, requiredDays=%d", 
					taskID, earliestDate.Format("2006-01-02"), requiredDays)
				// 获取当前时间作为审核日期
				auditDate := time.Now()
				err = CreateVideoReminder(tx, taskID, earliestDate, requiredDays, auditDate)
				if err != nil {
					// 记录错误但不阻断流程
					logger.Errorf("创建录像提醒任务失败: %v, taskID: %d", err, taskID)
				}
			} else {
				// 如果审核意见包含相关关键词但未解析成功，记录日志用于调试
				if strings.Contains(auditComment, "录像") && strings.Contains(auditComment, "不足") {
					logger.Errorf("审核意见包含录像相关关键词但解析失败: taskID=%d, comment=%s", taskID, auditComment)
				}
			}
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			http.Error(w, "事务提交失败: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 记录操作日志
		if currentUser := auth.GetCurrentUser(r); currentUser != nil {
			action := fmt.Sprintf("编辑审核意见（档案ID：%d，状态：%s）", taskID, auditStatus)
			operationlog.Record(r, currentUser.Username, action)
		}

		http.Redirect(w, r, "/audit/progress?message=EditSuccess", http.StatusSeeOther)
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
	err = db.DBInstance.QueryRow("SELECT COUNT(*) FROM audit_tasks WHERE id = ?", taskID).Scan(&exists)
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
		return nil // 非必填项传 nil, 数据库存 NULL
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
	var task AuditTask
	taskSQL := "SELECT id, file_name, organization, import_time, audit_status, audit_comment FROM audit_tasks WHERE id = ?"
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

	// 查询明细数据（只查询关键字段用于列表显示）
	type DetailItem struct {
		ID                int
		DeviceCode        string
		DeviceName        string
		DivisionCode      string
		MonitorPointType  string
		ManagementUnit    string
		MaintainUnit      string
		CameraFunctionType string
		AuditStatus       int // 建档状态：0-未审核未建档，1-已审核未建档，2-已建档
	}

	detailSQL := `SELECT id, device_code, device_name, division_code, monitor_point_type, 
		management_unit, maintain_unit, camera_function_type, audit_status
		FROM audit_details WHERE task_id = ? ORDER BY id`
	
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
			&item.DeviceCode,
			&item.DeviceName,
			&item.DivisionCode,
			&item.MonitorPointType,
			&item.ManagementUnit,
			&item.MaintainUnit,
			&item.CameraFunctionType,
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
		Task          AuditTask
		Details       []DetailItem
	}

	data := DetailPageData{
		Title:      "档案明细",
		ActiveMenu: "audit",
		SubMenu:    "audit_progress",
		Task:       task,
		Details:    detailList,
	}

	// 添加状态转换函数到模板
	funcMap := template.FuncMap{
		"getStatusText": getAuditStatusText,
	}

	tmpl, err := template.New("auditdetail.html").Funcs(funcMap).ParseFiles("templates/auditdetail.html")
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

// DetailExportHandler: 导出档案明细到Excel（全量字段）
func DetailExportHandler(w http.ResponseWriter, r *http.Request) {
	taskIDStr := r.URL.Query().Get("task_id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil || taskID <= 0 {
		http.Error(w, "无效的任务ID", http.StatusBadRequest)
		return
	}

	// 查询任务基本信息（获取档案名称）
	var task AuditTask
	taskSQL := "SELECT file_name FROM audit_tasks WHERE id = ?"
	err = db.DBInstance.QueryRow(taskSQL, taskID).Scan(&task.FileName)
	if err != nil {
		http.Error(w, "查询档案失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 查询所有字段（从audit_details表，只查询该任务的数据）
	querySQL := `SELECT 
		id, device_code, original_device_code, device_name, division_code, 
		monitor_point_type, pickup, parent_device, construction_unit, 
		construction_unit_code, management_unit, camera_dept, admin_name, 
		admin_contact, contractor, maintain_unit, device_vendor, device_model, 
		camera_type, access_method, camera_function_type, video_encoding_format, 
		image_resolution, camera_light_property, backend_structure, lens_type, 
		installation_type, height_type, jurisdiction_police, installation_address, 
		surrounding_landmark, longitude, latitude, installation_location, 
		monitoring_direction, pole_number, scene_picture, networking_property, 
		access_network, ipv4_address, ipv6_address, mac_address, access_port, 
		associated_encoder, device_username, device_password, channel_number, 
		connection_protocol, enabled_time, scrapped_time, device_status, 
		inspection_status, video_loss, color_distortion, video_blur, 
		brightness_exception, video_interference, video_lag, video_occlusion, 
		scene_change, online_duration, offline_duration, signaling_delay, 
		video_stream_delay, key_frame_delay, recording_retention_days, 
		storage_device_code, storage_channel_number, storage_type, 
		cache_settings, notes, collection_area_type, update_time
	FROM audit_details WHERE task_id = ? ORDER BY id`

	rows, err := db.DBInstance.Query(querySQL, taskID)
	if err != nil {
		logger.Errorf("设备审核进度-导出查询失败: %v, SQL: %s, taskID: %d", err, querySQL, taskID)
		http.Error(w, "导出查询失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// 创建 Excel 文件
	f := excelize.NewFile()
	sheetName := "档案明细"
	f.SetSheetName("Sheet1", sheetName)

	// 使用filelist的TemplateHeaders（73列）
	TemplateHeaders := []interface{}{
		"序号", "设备编码（*）", "原设备编码", "设备名称（*）", "行政区划编码（*）", "监控点位类型（*）", "拾音器",
		"父设备", "建设单位/设备归属（*）", "建设单位/平台归属代码（*）", "管理单位（*）",
		"摄像机所属部门（警种）（*）", "管理员姓名（*）", "管理员联系电话（*）", "承建单位（*）",
		"维护单位（*）", "设备厂商（*）", "设备型号", "摄像机类型（*）", "接入方式", "摄像机功能类型（*）",
		"视频编码格式（*）", "图像分辨率（*）", "摄像机补光属性", "后端结构化", "镜头类型", "安装类型",
		"高度类型（*）", "所属辖区公安机关（*）", "安装地址（*）", "周边标志（*）", "经度（*）", "纬度（*）",
		"摄像机安装位置室内外（*）", "摄像机监控方位（*）", "立杆编号（*）", "摄像机实景图片", "联网属性",
		"接入网络（*）", "IPv4地址（*）", "IPv6地址", "设备MAC地址（*）", "访问端口", "关联编码器",
		"设备用户名", "设备口令", "通道号", "连接协议", "启用时间（*）", "报废时间", "设备状态（*）",
		"巡检状态", "视频丢失", "色彩失真", "视频模糊", "亮度异常", "视频干扰", "视频卡顿", "视频遮挡",
		"场景变更", "在线时长", "离线时长", "信令时延", "视频流时延", "关键帧时延",
		"录像保存天数（*）", "存储设备编码", "存储通道号", "存储类型", "缓存设置", "备注", "采集区域类型（*）", "更新时间",
	}

	// 写入表头
	f.SetSheetRow(sheetName, "A1", &TemplateHeaders)

	// 写入数据
	rowNum := 2
	for rows.Next() {
		var item filelist.FileItem
		err = rows.Scan(
			&item.ID, &item.DeviceCode, &item.OriginalDeviceCode,
			&item.DeviceName, &item.DivisionCode, &item.MonitorPointType,
			&item.Pickup, &item.ParentDevice, &item.ConstructionUnit,
			&item.ConstructionUnitCode, &item.ManagementUnit, &item.CameraDept,
			&item.AdminName, &item.AdminContact, &item.Contractor,
			&item.MaintainUnit, &item.DeviceVendor, &item.DeviceModel,
			&item.CameraType, &item.AccessMethod, &item.CameraFunctionType,
			&item.VideoEncodingFormat, &item.ImageResolution, &item.CameraLightProperty,
			&item.BackendStructure, &item.LensType, &item.InstallationType,
			&item.HeightType, &item.JurisdictionPolice, &item.InstallationAddress,
			&item.SurroundingLandmark, &item.Longitude, &item.Latitude,
			&item.InstallationLocation, &item.MonitoringDirection, &item.PoleNumber,
			&item.ScenePicture, &item.NetworkingProperty, &item.AccessNetwork,
			&item.IPv4Address, &item.IPv6Address, &item.MACAddress,
			&item.AccessPort, &item.AssociatedEncoder, &item.DeviceUsername,
			&item.DevicePassword, &item.ChannelNumber, &item.ConnectionProtocol,
			&item.EnabledTime, &item.ScrappedTime, &item.DeviceStatus,
			&item.InspectionStatus, &item.VideoLoss, &item.ColorDistortion,
			&item.VideoBlur, &item.BrightnessException, &item.VideoInterference,
			&item.VideoLag, &item.VideoOcclusion, &item.SceneChange,
			&item.OnlineDuration, &item.OfflineDuration, &item.SignalingDelay,
			&item.VideoStreamDelay, &item.KeyFrameDelay, &item.RecordingRetentionDays,
			&item.StorageDeviceCode, &item.StorageChannelNumber, &item.StorageType,
			&item.CacheSettings, &item.Notes, &item.CollectionAreaType,
			&item.UpdateTime,
		)
		if err != nil {
			continue
		}

		// 构建行数据
		rowData := []interface{}{
			item.ID, item.DeviceCode, item.OriginalDeviceCode.String,
			item.DeviceName, item.DivisionCode, item.MonitorPointType,
			item.Pickup.String, item.ParentDevice.String, item.ConstructionUnit,
			item.ConstructionUnitCode, item.ManagementUnit, item.CameraDept,
			item.AdminName, item.AdminContact, item.Contractor,
			item.MaintainUnit, item.DeviceVendor, item.DeviceModel.String,
			item.CameraType, item.AccessMethod.String, item.CameraFunctionType,
			item.VideoEncodingFormat, item.ImageResolution, item.CameraLightProperty.String,
			item.BackendStructure.String, item.LensType.String, item.InstallationType.String,
			item.HeightType, item.JurisdictionPolice, item.InstallationAddress,
			item.SurroundingLandmark, item.Longitude, item.Latitude,
			item.InstallationLocation, item.MonitoringDirection, item.PoleNumber,
			item.ScenePicture.String, item.NetworkingProperty.String, item.AccessNetwork,
			item.IPv4Address, item.IPv6Address.String, item.MACAddress,
			item.AccessPort.String, item.AssociatedEncoder.String, item.DeviceUsername.String,
			item.DevicePassword.String, item.ChannelNumber.String, item.ConnectionProtocol.String,
			item.EnabledTime.String, item.ScrappedTime.String, item.DeviceStatus,
			item.InspectionStatus.String, item.VideoLoss.Int64, item.ColorDistortion.Int64,
			item.VideoBlur.Int64, item.BrightnessException.Int64, item.VideoInterference.Int64,
			item.VideoLag.Int64, item.VideoOcclusion.Int64, item.SceneChange.Int64,
			item.OnlineDuration.Int64, item.OfflineDuration.Int64, item.SignalingDelay.Int64,
			item.VideoStreamDelay.Int64, item.KeyFrameDelay.Int64, item.RecordingRetentionDays,
			item.StorageDeviceCode.String, item.StorageChannelNumber.String, item.StorageType.String,
			item.CacheSettings.String, item.Notes.String, item.CollectionAreaType,
			item.UpdateTime,
		}

		cellName, _ := excelize.CoordinatesToCellName(1, rowNum)
		f.SetSheetRow(sheetName, cellName, &rowData)
		rowNum++
	}

	// 记录导出操作日志
	currentUser := auth.GetCurrentUser(r)
	if currentUser != nil {
		action := fmt.Sprintf("导出设备审核档案明细 Excel（档案名称：%s）", task.FileName)
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

// 模板下载时的表头定义 (73列) - 与filelist保持一致
var TemplateHeaders = []interface{}{
	"序号", "设备编码（*）", "原设备编码", "设备名称（*）", "行政区划编码（*）", "监控点位类型（*）", "拾音器",
	"父设备", "建设单位/设备归属（*）", "建设单位/平台归属代码（*）", "管理单位（*）",
	"摄像机所属部门（警种）（*）", "管理员姓名（*）", "管理员联系电话（*）", "承建单位（*）",
	"维护单位（*）", "设备厂商（*）", "设备型号", "摄像机类型（*）", "接入方式", "摄像机功能类型（*）",
	"视频编码格式（*）", "图像分辨率（*）", "摄像机补光属性", "后端结构化", "镜头类型", "安装类型",
	"高度类型（*）", "所属辖区公安机关（*）", "安装地址（*）", "周边标志（*）", "经度（*）", "纬度（*）",
	"摄像机安装位置室内外（*）", "摄像机监控方位（*）", "立杆编号（*）", "摄像机实景图片", "联网属性",
	"接入网络（*）", "IPv4地址（*）", "IPv6地址", "设备MAC地址（*）", "访问端口", "关联编码器",
	"设备用户名", "设备口令", "通道号", "连接协议", "启用时间（*）", "报废时间", "设备状态（*）",
	"巡检状态", "视频丢失", "色彩失真", "视频模糊", "亮度异常", "视频干扰", "视频卡顿", "视频遮挡",
	"场景变更", "在线时长", "离线时长", "信令时延", "视频流时延", "关键帧时延",
	"录像保存天数（*）", "存储设备编码", "存储通道号", "存储类型", "缓存设置", "备注",
	"采集区域类型（*）", "更新时间",
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
		action := "下载审核档案导入模板"
		operationlog.Record(r, currentUser.Username, action)
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	filename := fmt.Sprintf("attachment; filename=\"审核档案导入模板_%s.xlsx\"", time.Now().Format("20060102"))
	w.Header().Set("Content-Disposition", filename)
	f.Write(w)
}

// DeleteHandler: 删除审核档案（同时删除关联的明细）
func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/audit/progress", http.StatusSeeOther)
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
	var task AuditTask
	taskSQL := "SELECT id, file_name, organization FROM audit_tasks WHERE id = ?"
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
	countSQL := "SELECT COUNT(*) FROM audit_details WHERE task_id = ?"
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
	deleteDetailsSQL := "DELETE FROM audit_details WHERE task_id = ?"
	_, err = tx.Exec(deleteDetailsSQL, taskID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "删除档案明细失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 再删除任务记录
	deleteTaskSQL := "DELETE FROM audit_tasks WHERE id = ?"
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
		action := fmt.Sprintf("删除审核档案（档案名称：%s，机构：%s，包含 %d 条明细）", task.FileName, task.Organization, detailCount)
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
	redirectURL := "/audit/progress?message=DeleteSuccess"
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

	// 查询档案名称和单兵设备信息
	var fileName string
	var isSingleSoldier int
	query := "SELECT file_name, is_single_soldier FROM audit_tasks WHERE id = ?"
	err = db.DBInstance.QueryRow(query, taskID).Scan(&fileName, &isSingleSoldier)
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
		logger.Errorf("设备审核进度-查询档案名称失败: %v", err)
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
		logger.Errorf("设备审核进度-创建目录失败: %v, 路径: %s", err, targetDir)
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
			logger.Errorf("设备审核进度-打开文件失败: %v, 文件名: %s", err, fileHeader.Filename)
			failedFiles = append(failedFiles, fileHeader.Filename)
			continue
		}

		// 处理文件路径（支持目录结构，保留相对路径）
		// 浏览器上传目录时，文件名可能包含相对路径，如 "subdir/file.txt"
		filePath := fileHeader.Filename
		// 清理路径，防止路径遍历攻击
		filePath = filepath.Clean(filePath)
		if strings.HasPrefix(filePath, "..") || strings.HasPrefix(filePath, "/") {
			logger.Errorf("设备审核进度-非法文件路径: %s", filePath)
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
			logger.Errorf("设备审核进度-创建子目录失败: %v, 路径: %s", err, targetFileDir)
			file.Close()
			failedFiles = append(failedFiles, fileHeader.Filename)
			continue
		}

		// 创建目标文件
		dst, err := os.Create(targetFile)
		if err != nil {
			logger.Errorf("设备审核进度-创建文件失败: %v, 路径: %s", err, targetFile)
			file.Close()
			failedFiles = append(failedFiles, fileHeader.Filename)
			continue
		}

		// 复制文件内容
		_, err = io.Copy(dst, file)
		dst.Close()
		file.Close()

		if err != nil {
			logger.Errorf("设备审核进度-写入文件失败: %v, 路径: %s", err, targetFile)
			failedFiles = append(failedFiles, fileHeader.Filename)
			continue
		}

		uploadedFiles = append(uploadedFiles, fileHeader.Filename)
	}

	// 构建操作日志
	var action string
	if len(uploadedFiles) == 1 {
		action = fmt.Sprintf("上传文件（档案：%s，文件名：%s，是否单兵设备：%d）", fileName, uploadedFiles[0], isSingleSoldier)
	} else {
		action = fmt.Sprintf("上传目录/文件（档案：%s，成功：%d/%d，是否单兵设备：%d）", fileName, len(uploadedFiles), totalFiles, isSingleSoldier)
		if len(failedFiles) > 0 {
			action += fmt.Sprintf("，失败：%d", len(failedFiles))
		}
	}
	operationlog.Record(r, currentUser.Username, action)

	// 返回响应
	w.WriteHeader(http.StatusOK)
	response := map[string]interface{}{
		"success":       true,
		"message":        fmt.Sprintf("上传完成！成功：%d/%d", len(uploadedFiles), totalFiles),
		"total":          totalFiles,
		"successCount":  len(uploadedFiles),
		"failedCount":    len(failedFiles),
		"uploadedFiles":  uploadedFiles,
		"failedFiles":    failedFiles,
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
	query := "SELECT file_name FROM audit_tasks WHERE id = ?"
	err = db.DBInstance.QueryRow(query, taskID).Scan(&archiveFileName)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "档案不存在", http.StatusNotFound)
			return
		}
		logger.Errorf("设备审核进度-查询档案名称失败: %v", err)
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
		logger.Errorf("设备审核进度-打开文件失败: %v, 路径: %s", err, filePath)
		http.Error(w, "打开文件失败", http.StatusInternalServerError)
		return
	}
	defer file.Close()
	
	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		logger.Errorf("设备审核进度-获取文件信息失败: %v", err)
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
		logger.Errorf("设备审核进度-下载文件失败: %v", err)
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
		var task AuditTask
		taskSQL := "SELECT id, file_name, organization, audit_status FROM audit_tasks WHERE id = ?"
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

		type SamplePageData struct {
			Title      string
			ActiveMenu string
			SubMenu    string
			Task       AuditTask
			SampledBy  string
		}

		data := SamplePageData{
			Title:      "抽检",
			ActiveMenu: "audit",
			SubMenu:    "audit_progress",
			Task:       task,
			SampledBy:  currentUser.Username,
		}

		tmpl, err := template.ParseFiles("templates/auditsample.html")
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
		checkSQL := "SELECT audit_status FROM audit_tasks WHERE id = ?"
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
		action := fmt.Sprintf("抽检设备审核档案（档案ID：%d，结果：%s）", taskID, sampleResult)
		operationlog.Record(r, currentUser.Username, action)

		http.Redirect(w, r, "/audit/progress?message=SampleSuccess", http.StatusSeeOther)
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
	err = db.DBInstance.QueryRow("SELECT COUNT(*) FROM audit_tasks WHERE id = ?", taskID).Scan(&exists)
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

// VideoReminderHandler: 录像提醒列表页
func VideoReminderHandler(w http.ResponseWriter, r *http.Request) {
	// 获取查询参数
	status := r.URL.Query().Get("status") // 状态筛选：pending, notified, completed 或空
	pageStr := r.URL.Query().Get("page")

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	pageSize := 30

	// 查询提醒列表
	reminders, totalCount, err := GetVideoReminders(status, page, pageSize)
	if err != nil {
		logger.Errorf("查询录像提醒列表失败: %v", err)
		http.Error(w, "查询提醒列表失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 计算分页信息
	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}
	if page > totalPages {
		page = totalPages
	}

	// 计算剩余天数
	type ReminderDisplay struct {
		VideoReminder
		RemainingDays int    // 剩余天数（负数表示已过期）
		StatusText    string // 状态文本
	}

	var displayList []ReminderDisplay
	today := time.Now()
	for _, reminder := range reminders {
		remainingDays := int(reminder.ReminderDate.Sub(today).Hours() / 24)
		statusText := map[string]string{
			"pending":  "待处理",
			"notified": "已通知",
			"completed": "已完成",
		}[reminder.Status]
		if statusText == "" {
			statusText = reminder.Status
		}

		displayList = append(displayList, ReminderDisplay{
			VideoReminder: reminder,
			RemainingDays: remainingDays,
			StatusText:    statusText,
		})
	}

	// 构建查询参数字符串
	queryParams := []string{}
	if status != "" {
		queryParams = append(queryParams, "status="+status)
	}
	query := strings.Join(queryParams, "&")

	// 准备数据
	type ReminderPageData struct {
		Title         string
		ActiveMenu    string
		SubMenu       string
		List          []ReminderDisplay
		Status        string
		CurrentPage   int
		TotalPages    int
		TotalCount    int
		StartRecord   int
		EndRecord     int
		HasPrev       bool
		HasNext       bool
		PrevPage      int
		NextPage      int
		FirstPage     int
		LastPage      int
		Query         string
	}

	startRecord := (page-1)*pageSize + 1
	endRecord := page * pageSize
	if endRecord > totalCount {
		endRecord = totalCount
	}
	if totalCount == 0 {
		startRecord = 0
		endRecord = 0
	}

	data := ReminderPageData{
		Title:       "录像天数不足提醒",
		ActiveMenu:  "audit",
		SubMenu:     "video_reminders",
		List:        displayList,
		Status:      status,
		CurrentPage: page,
		TotalPages:  totalPages,
		TotalCount:  totalCount,
		StartRecord: startRecord,
		EndRecord:   endRecord,
		HasPrev:     page > 1,
		HasNext:     page < totalPages,
		PrevPage:    page - 1,
		NextPage:    page + 1,
		FirstPage:   1,
		LastPage:    totalPages,
		Query:       query,
	}

	tmpl, err := template.ParseFiles("templates/auditvideoreminder.html")
	if err != nil {
		logger.Errorf("录像提醒列表-模板解析失败: %v", err)
		http.Error(w, "模板解析失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		logger.Errorf("录像提醒列表-模板渲染失败: %v", err)
		http.Error(w, "模板渲染失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// CompleteReminderHandler: 标记提醒为已完成
func CompleteReminderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "表单解析失败", http.StatusBadRequest)
		return
	}

	reminderIDStr := r.FormValue("reminder_id")
	reminderID, err := strconv.Atoi(reminderIDStr)
	if err != nil || reminderID <= 0 {
		http.Error(w, "无效的提醒ID", http.StatusBadRequest)
		return
	}

	currentUser := auth.GetCurrentUser(r)
	if currentUser == nil {
		http.Error(w, "未登录", http.StatusUnauthorized)
		return
	}

	err = CompleteVideoReminder(reminderID, currentUser.Username)
	if err != nil {
		logger.Errorf("标记提醒为已完成失败: %v, reminderID: %d", err, reminderID)
		http.Error(w, "标记失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 记录操作日志
	action := fmt.Sprintf("标记录像提醒为已完成（提醒ID：%d）", reminderID)
	operationlog.Record(r, currentUser.Username, action)

	// 重定向回提醒列表
	status := r.FormValue("status")
	redirectURL := "/audit/progress/video-reminders"
	params := []string{}
	if status != "" {
		params = append(params, "status="+status)
	}
	params = append(params, "message=CompleteSuccess")
	if len(params) > 0 {
		redirectURL += "?" + strings.Join(params, "&")
	}
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// DeleteReminderHandler: 删除录像提醒任务（支持单条和批量删除）
func DeleteReminderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	currentUser := auth.GetCurrentUser(r)
	if currentUser == nil {
		http.Error(w, "未登录", http.StatusUnauthorized)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "表单解析失败", http.StatusBadRequest)
		return
	}

	// 获取要删除的提醒ID列表
	reminderIDsStr := r.Form["reminder_id"]
	if len(reminderIDsStr) == 0 {
		http.Error(w, "请选择要删除的记录", http.StatusBadRequest)
		return
	}

	// 转换为整数切片
	var reminderIDs []int
	for _, idStr := range reminderIDsStr {
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			continue
		}
		reminderIDs = append(reminderIDs, id)
	}

	if len(reminderIDs) == 0 {
		http.Error(w, "无效的提醒ID", http.StatusBadRequest)
		return
	}

	// 执行删除
	err = DeleteVideoReminders(reminderIDs)
	if err != nil {
		logger.Errorf("删除录像提醒任务失败: %v, IDs: %v", err, reminderIDs)
		http.Error(w, "删除失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 记录操作日志
	if len(reminderIDs) == 1 {
		action := fmt.Sprintf("删除录像提醒任务（提醒ID：%d）", reminderIDs[0])
		operationlog.Record(r, currentUser.Username, action)
	} else {
		action := fmt.Sprintf("批量删除录像提醒任务（共 %d 条）", len(reminderIDs))
		operationlog.Record(r, currentUser.Username, action)
	}

	// 重定向回提醒列表
	status := r.FormValue("status")
	page := r.FormValue("page")
	redirectURL := "/audit/progress/video-reminders"
	params := []string{}
	if status != "" {
		params = append(params, "status="+status)
	}
	if page != "" {
		params = append(params, "page="+page)
	}
	params = append(params, "message=DeleteSuccess")
	if len(params) > 0 {
		redirectURL += "?" + strings.Join(params, "&")
	}
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// ScheduleConfigHandler: 定时任务配置页面
func ScheduleConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// 显示配置页面
		config, err := GetScheduleConfig()
		if err != nil {
			logger.Errorf("获取定时配置失败: %v", err)
			// 使用默认配置
			config = &ScheduleConfig{
				Frequency: "daily",
				Hour:      1,
				Enabled:   true,
			}
		}

		type ConfigPageData struct {
			Title       string
			ActiveMenu  string
			SubMenu     string
			Config      *ScheduleConfig
		}

		data := ConfigPageData{
			Title:      "录像提醒定时任务配置",
			ActiveMenu: "audit",
			SubMenu:    "video_reminders",
			Config:     config,
		}

		tmpl, err := template.ParseFiles("templates/auditvideoreminderschedule.html")
		if err != nil {
			logger.Errorf("定时配置页面-模板解析失败: %v", err)
			http.Error(w, "模板解析失败: "+err.Error(), http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			logger.Errorf("定时配置页面-模板渲染失败: %v", err)
			http.Error(w, "模板渲染失败: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else if r.Method == http.MethodPost {
		// 保存配置
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "表单解析失败", http.StatusBadRequest)
			return
		}

		frequency := r.FormValue("frequency")
		hourStr := r.FormValue("hour")
		dayOfWeekStr := r.FormValue("day_of_week")
		enabledStr := r.FormValue("enabled")

		// 验证参数
		if frequency != "daily" && frequency != "weekly" {
			http.Error(w, "无效的执行频率", http.StatusBadRequest)
			return
		}

		hour, err := strconv.Atoi(hourStr)
		if err != nil || hour < 1 || hour > 24 {
			http.Error(w, "无效的执行时间", http.StatusBadRequest)
			return
		}

		var dayOfWeek int
		if frequency == "weekly" {
			dayOfWeek, err = strconv.Atoi(dayOfWeekStr)
			if err != nil || dayOfWeek < 1 || dayOfWeek > 7 {
				http.Error(w, "无效的星期几", http.StatusBadRequest)
				return
			}
		}

		enabled := enabledStr == "1"

		currentUser := auth.GetCurrentUser(r)
		if currentUser == nil {
			http.Error(w, "未登录", http.StatusUnauthorized)
			return
		}

		err = SaveScheduleConfig(frequency, hour, dayOfWeek, enabled, currentUser.Username)
		if err != nil {
			logger.Errorf("保存定时配置失败: %v", err)
			http.Error(w, "保存配置失败: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 记录操作日志
		action := fmt.Sprintf("更新录像提醒定时任务配置（频率：%s，时间：%d:00）", frequency, hour)
		operationlog.Record(r, currentUser.Username, action)

		http.Redirect(w, r, "/audit/progress/video-reminders/schedule?message=SaveSuccess", http.StatusSeeOther)
	}
}

