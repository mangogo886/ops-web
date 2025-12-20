package checkpointfilelist

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
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// 列表页面专用的轻量级结构体
type CheckpointFileItemList struct {
	ID                    int
	CheckpointCode        string
	CheckpointName        string
	DivisionCode          string
	CheckpointPointType   string
	ManagementUnit        string // 现在显示checkpoint_tasks.organization的值
	CheckpointMaintainUnit string
	UpdateTime            string
	AuditStatus           int // 建档状态：0-未审核未建档，1-已审核未建档，2-已建档
	FileName              string // 所属任务档案（checkpoint_tasks.file_name）
}

// 导出时使用的完整结构体（包含所有字段，除了task_id）
type CheckpointFileItem struct {
	ID                                      int
	CheckpointCode                          sql.NullString
	OriginalCheckpointCode                  sql.NullString
	CheckpointName                          sql.NullString
	CheckpointAddress                       sql.NullString
	RoadName                                sql.NullString
	DirectionType                           sql.NullString
	DirectionDescription                    sql.NullString
	DirectionNotes                          sql.NullString
	DivisionCode                            sql.NullString
	RoadSectionType                         sql.NullString
	RoadCode                                sql.NullString
	KilometerOrIntersectionNumber           sql.NullString
	RoadMeter                               sql.NullString
	PoleNumber                              sql.NullString
	CheckpointPointType                     sql.NullString
	CheckpointLocationType                  sql.NullString
	CheckpointApplicationType               sql.NullString
	HasInterceptionCondition                sql.NullString
	HasSpeedMeasurement                     sql.NullString
	HasRealtimeVideo                        sql.NullString
	HasFaceCapture                          sql.NullString
	HasViolationCapture                     sql.NullString
	HasFrontendSecondaryRecognition        sql.NullString
	IsBoundaryCheckpoint                    sql.NullString
	AdjacentArea                            sql.NullString
	CheckpointLongitude                     sql.NullString
	CheckpointLatitude                      sql.NullString
	CheckpointScenePhotoURL                 sql.NullString
	CheckpointStatus                        sql.NullString
	CaptureTriggerType                      sql.NullString
	CaptureDirectionType                    sql.NullString
	TotalLanes                              sql.NullString
	PanoramicCameraDeviceCode               sql.NullString
	NextCheckpointAlongRoad                 sql.NullString
	NextCheckpointOpposite                  sql.NullString
	NextCheckpointLeftTurn                  sql.NullString
	NextCheckpointRightTurn                 sql.NullString
	NextCheckpointUTurn                     sql.NullString
	ConstructionUnit                        sql.NullString
	ManagementUnit                         sql.NullString
	CheckpointDepartment                    sql.NullString
	AdminName                               sql.NullString
	AdminContact                            sql.NullString
	CheckpointContractor                   sql.NullString
	CheckpointMaintainUnit                 sql.NullString
	AlarmReceivingDepartment                sql.NullString
	AlarmReceivingDepartmentCode           sql.NullString
	AlarmReceivingPhone                     sql.NullString
	InterceptionDepartment                 sql.NullString
	InterceptionDepartmentCode             sql.NullString
	InterceptionDepartmentContact          sql.NullString
	TerminalCode                            sql.NullString
	TerminalIPAddress                       sql.NullString
	TerminalPort                            sql.NullString
	TerminalUsername                        sql.NullString
	TerminalPassword                        sql.NullString
	TerminalVendor                          sql.NullString
	CheckpointEnabledTime                   sql.NullString
	CheckpointRevokedTime                   sql.NullString
	Notes                                   sql.NullString
	CheckpointDeviceType                    sql.NullString
	TotalCaptureCameras                     sql.NullString
	CentralControlCode                      sql.NullString
	CentralControlIPAddress                 sql.NullString
	CentralControlPort                      sql.NullString
	CentralControlUsername                  sql.NullString
	CentralControlPassword                  sql.NullString
	CentralControlVendor                    sql.NullString
	CheckpointScrappedTime                  sql.NullString
	TotalAntennas                           sql.NullString
	TerminalMACAddress                       sql.NullString
	CollectionAreaType                      sql.NullString
	IntegratedCommandPlatformCheckpointCode sql.NullString
	UpdateTime                              string
	AuditStatus                             int
}

// 导出表头定义（75列，与TemplateHeaders一致）
var ExportHeaders = []interface{}{
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

// 页面数据结构体
type PageData struct {
	Title            string
	ActiveMenu       string
	SubMenu          string
	List             []CheckpointFileItemList
	SearchCode       string
	SearchName       string
	Month            string // 月份查询条件 (格式: 2024-01)
	AuditStatus      string // 建档状态查询条件
	CurrentPage      int
	TotalPages       int
	HasPrev          bool
	HasNext          bool
	PrevPage         int
	NextPage         int
	FirstPage        int
	LastPage         int
	Query            string
	TotalCount       int   // 总记录数
	CurrentPageCount int   // 当前页记录数
}

// Handler: 卡口建档明细列表页 (GET)
func Handler(w http.ResponseWriter, r *http.Request) {
	// 获取查询参数
	searchCode := r.URL.Query().Get("checkpoint_code")
	searchName := r.URL.Query().Get("checkpoint_name")
	month := r.URL.Query().Get("month")        // 月份，格式: 2024-01
	auditStatus := r.URL.Query().Get("audit_status") // 建档状态: 0, 1, 2 或空
	pageStr := r.URL.Query().Get("page")

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	pageSize := 50
	offset := (page - 1) * pageSize

	// 构造查询条件（使用表别名）
	whereSQL := " WHERE 1=1"
	args := []interface{}{}

	if searchCode != "" {
		whereSQL += " AND cd.checkpoint_code LIKE ?"
		args = append(args, "%"+searchCode+"%")
	}
	if searchName != "" {
		whereSQL += " AND cd.checkpoint_name LIKE ?"
		args = append(args, "%"+searchName+"%")
	}

	// 月份查询：根据update_time按月份
	if month != "" {
		if _, err := time.Parse("2006-01", month); err == nil {
			whereSQL += " AND YEAR(cd.update_time) = ? AND MONTH(cd.update_time) = ?"
			parts := strings.Split(month, "-")
			if len(parts) == 2 {
				year, _ := strconv.Atoi(parts[0])
				monthNum, _ := strconv.Atoi(parts[1])
				args = append(args, year, monthNum)
			}
		}
	}

	// 建档状态查询
	if auditStatus != "" {
		statusInt, err := strconv.Atoi(auditStatus)
		if err == nil && (statusInt == 0 || statusInt == 1 || statusInt == 2) {
			whereSQL += " AND cd.audit_status = ?"
			args = append(args, statusInt)
		}
	}

	// 1. 查询总记录数（使用表别名）
	var totalCount int
	countSQL := "SELECT COUNT(*) FROM checkpoint_details cd LEFT JOIN checkpoint_tasks ct ON cd.task_id = ct.id" + whereSQL
	err := db.DBInstance.QueryRow(countSQL, args...).Scan(&totalCount)
	if err != nil && err != sql.ErrNoRows {
		logger.Errorf("卡口建档明细-查询总数失败: %v, SQL: %s, Args: %v", err, countSQL, args)
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

	// 3. 查询列表数据（关联checkpoint_tasks表获取organization和file_name）
	queryFields := "cd.id, cd.checkpoint_code, cd.checkpoint_name, cd.division_code, cd.checkpoint_point_type, COALESCE(ct.organization, cd.management_unit) as management_unit, cd.checkpoint_maintain_unit, cd.update_time, cd.audit_status, COALESCE(ct.file_name, '') as file_name"
	querySQL := fmt.Sprintf("SELECT %s FROM checkpoint_details cd LEFT JOIN checkpoint_tasks ct ON cd.task_id = ct.id %s ORDER BY cd.id DESC LIMIT ? OFFSET ?", queryFields, whereSQL)

	// 准备完整的参数列表
	queryArgs := append(args, pageSize, offset)

	rows, err := db.DBInstance.Query(querySQL, queryArgs...)
	if err != nil {
		logger.Errorf("卡口建档明细-数据库查询失败: %v, SQL: %s, Args: %v", err, querySQL, queryArgs)
		http.Error(w, "数据库查询失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var fileList []CheckpointFileItemList
	for rows.Next() {
		var item CheckpointFileItemList
		var updateTimeRaw sql.NullString

		err = rows.Scan(
			&item.ID,
			&item.CheckpointCode,
			&item.CheckpointName,
			&item.DivisionCode,
			&item.CheckpointPointType,
			&item.ManagementUnit,
			&item.CheckpointMaintainUnit,
			&updateTimeRaw,
			&item.AuditStatus,
			&item.FileName,
		)
		if err != nil {
			logger.Errorf("卡口建档明细-数据扫描失败: %v", err)
			continue
		}
		item.UpdateTime = formatDateTime(updateTimeRaw.String)

		if err != nil {
			fmt.Printf("❌ Database scan error: %v\n", err)
			continue
		}

		fileList = append(fileList, item)
	}

	// 检查遍历过程中的错误
	if err = rows.Err(); err != nil {
		logger.Errorf("卡口建档明细-数据遍历失败: %v", err)
		http.Error(w, "数据遍历失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 构建查询参数字符串（用于表单回显和分页链接）
	queryParams := []string{}
	if searchCode != "" {
		queryParams = append(queryParams, "checkpoint_code="+searchCode)
	}
	if searchName != "" {
		queryParams = append(queryParams, "checkpoint_name="+searchName)
	}
	if month != "" {
		queryParams = append(queryParams, "month="+month)
	}
	if auditStatus != "" {
		queryParams = append(queryParams, "audit_status="+auditStatus)
	}
	query := strings.Join(queryParams, "&")

	// 记录查询操作日志（如果有查询条件）
	currentUser := auth.GetCurrentUser(r)
	if currentUser != nil && (searchCode != "" || searchName != "" || month != "" || auditStatus != "") {
		action := "查询卡口建档明细"
		if searchCode != "" {
			action += fmt.Sprintf("（卡口编码：%s", searchCode)
		}
		if searchName != "" {
			if searchCode != "" {
				action += fmt.Sprintf("，卡口名称：%s", searchName)
			} else {
				action += fmt.Sprintf("（卡口名称：%s", searchName)
			}
		}
		if month != "" {
			if searchCode != "" || searchName != "" {
				action += fmt.Sprintf("，月份：%s", month)
			} else {
				action += fmt.Sprintf("（月份：%s", month)
			}
		}
		if auditStatus != "" {
			if searchCode != "" || searchName != "" || month != "" {
				action += fmt.Sprintf("，建档状态：%s）", auditStatus)
			} else {
				action += fmt.Sprintf("（建档状态：%s）", auditStatus)
			}
		} else if searchCode != "" || searchName != "" || month != "" {
			action += "）"
		}
		operationlog.Record(r, currentUser.Username, action)
	}

	// 准备数据并渲染模板
	data := PageData{
		Title:            "卡口建档明细",
		ActiveMenu:       "filelist",
		SubMenu:          "checkpoint_filelist",
		List:             fileList,
		SearchCode:       searchCode,
		SearchName:       searchName,
		Month:            month,
		AuditStatus:      auditStatus,
		CurrentPage:      page,
		TotalPages:       totalPages,
		HasPrev:          page > 1,
		HasNext:          page < totalPages,
		PrevPage:         page - 1,
		NextPage:         page + 1,
		FirstPage:        1,
		LastPage:         totalPages,
		Query:            query,
		TotalCount:       totalCount,
		CurrentPageCount: len(fileList),
	}

	// 添加状态转换函数到模板
	funcMap := template.FuncMap{
		"getAuditStatusText": getAuditStatusText,
	}

	tmpl, err := template.New("checkpointfilelist.html").Funcs(funcMap).ParseFiles("templates/checkpointfilelist.html")
	if err != nil {
		logger.Errorf("卡口建档明细-模板解析失败: %v", err)
		http.Error(w, "模板解析失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		logger.Errorf("卡口建档明细-模板渲染失败: %v", err)
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

// ExportHandler: 导出卡口建档明细到Excel
func ExportHandler(w http.ResponseWriter, r *http.Request) {
	// 获取查询参数（必须与Handler中的查询条件保持一致）
	// 先尝试从 URL Query 获取
	searchCode := r.URL.Query().Get("checkpoint_code")
	searchName := r.URL.Query().Get("checkpoint_name")
	month := r.URL.Query().Get("month")
	auditStatus := r.URL.Query().Get("audit_status")

	// 如果从 Query 获取不到参数，但 RawQuery 不为空，尝试从 RawQuery 手动解析（处理 URL 编码问题）
	if r.URL.RawQuery != "" && (searchCode == "" || searchName == "" || month == "" || auditStatus == "") {
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
					if key == "checkpoint_code" && searchCode == "" {
						searchCode = value
					} else if key == "checkpoint_name" && searchName == "" {
						searchName = value
					} else if key == "month" && month == "" {
						month = value
					} else if key == "audit_status" && auditStatus == "" {
						auditStatus = value
					}
				}
			}
		}
	}

	// 构造查询条件
	whereSQL := " WHERE 1=1"
	args := []interface{}{}

	if searchCode != "" {
		whereSQL += " AND checkpoint_code LIKE ?"
		args = append(args, "%"+searchCode+"%")
	}
	if searchName != "" {
		whereSQL += " AND checkpoint_name LIKE ?"
		args = append(args, "%"+searchName+"%")
	}

	// 月份查询
	if month != "" {
		if _, err := time.Parse("2006-01", month); err == nil {
			whereSQL += " AND YEAR(update_time) = ? AND MONTH(update_time) = ?"
			parts := strings.Split(month, "-")
			if len(parts) == 2 {
				year, _ := strconv.Atoi(parts[0])
				monthNum, _ := strconv.Atoi(parts[1])
				args = append(args, year, monthNum)
			}
		}
	}

	// 建档状态查询
	if auditStatus != "" {
		statusInt, err := strconv.Atoi(auditStatus)
		if err == nil && (statusInt == 0 || statusInt == 1 || statusInt == 2) {
			whereSQL += " AND audit_status = ?"
			args = append(args, statusInt)
		}
	}

	// 查询所有字段（从checkpoint_details表，不包括task_id）
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
	FROM checkpoint_details` + whereSQL + ` ORDER BY id DESC`

	rows, err := db.DBInstance.Query(querySQL, args...)
	if err != nil {
		logger.Errorf("卡口建档明细-导出查询失败: %v, SQL: %s, Args: %v", err, querySQL, args)
		http.Error(w, "导出查询失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// 创建Excel文件
	f := excelize.NewFile()
	sheetName := "卡口建档明细"
	f.SetSheetName("Sheet1", sheetName)

	// 写入表头
	f.SetSheetRow(sheetName, "A1", &ExportHeaders)

	// 设置表头样式
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#3498db"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	// 计算最后一列的字母（A-Z, AA-ZZ等）
	lastColName, _ := excelize.CoordinatesToCellName(len(ExportHeaders), 1)
	lastColLetter := strings.Split(lastColName, "1")[0]
	f.SetCellStyle(sheetName, "A1", lastColLetter+"1", headerStyle)

	// 写入数据
	rowNum := 2
	for rows.Next() {
		var item CheckpointFileItem
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
		action := "导出卡口建档明细 Excel"
		if searchCode != "" || searchName != "" || month != "" || auditStatus != "" {
			action += "（"
			conditions := []string{}
			if searchCode != "" {
				conditions = append(conditions, fmt.Sprintf("卡口编码：%s", searchCode))
			}
			if searchName != "" {
				conditions = append(conditions, fmt.Sprintf("卡口名称：%s", searchName))
			}
			if month != "" {
				conditions = append(conditions, fmt.Sprintf("月份：%s", month))
			}
			if auditStatus != "" {
				statusText := map[string]string{"0": "未审核未建档", "1": "已审核未建档", "2": "已建档"}
				conditions = append(conditions, fmt.Sprintf("建档状态：%s", statusText[auditStatus]))
			}
			action += strings.Join(conditions, "，") + "）"
		}
		operationlog.Record(r, currentUser.Username, action)
	}

	// 输出文件
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	filename := fmt.Sprintf("attachment; filename=\"卡口建档明细_%s.xlsx\"", time.Now().Format("20060102_150405"))
	w.Header().Set("Content-Disposition", filename)
	f.Write(w)
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




