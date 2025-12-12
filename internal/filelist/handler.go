package filelist

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"ops-web/internal/db"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// --- 模板与数据结构定义 ---

// 导出时的表头定义 (73列)
var ExportHeaders = map[string]string{
	"id": "序号", "device_code": "设备编码（*）", "original_device_code": "原设备编码", "device_name": "设备名称（*）",
	"division_code": "行政区划编码（*）", "monitor_point_type": "监控点位类型（*）", "pickup": "拾音器",
	"parent_device": "父设备", "construction_unit": "建设单位/设备归属（*）", "construction_unit_code": "建设单位/平台归属代码（*）",
	"management_unit": "管理单位（*）", "camera_dept": "摄像机所属部门（警种）（*）", "admin_name": "管理员姓名（*）",
	"admin_contact": "管理员联系电话（*）", "contractor": "承建单位（*）", "maintain_unit": "维护单位（*）",
	"device_vendor": "设备厂商（*）", "device_model": "设备型号", "camera_type": "摄像机类型（*）",
	"access_method": "接入方式", "camera_function_type": "摄像机功能类型（*）", "video_encoding_format": "视频编码格式（*）",
	"image_resolution": "图像分辨率（*）", "camera_light_property": "摄像机补光属性", "backend_structure": "后端结构化",
	"lens_type": "镜头类型", "installation_type": "安装类型", "height_type": "高度类型（*）",
	"jurisdiction_police": "所属辖区公安机关（*）", "installation_address": "安装地址（*）", "surrounding_landmark": "周边标志（*）",
	"longitude": "经度（*）", "latitude": "纬度（*）", "installation_location": "摄像机安装位置室内外（*）",
	"monitoring_direction": "摄像机监控方位（*）", "pole_number": "立杆编号（*）", "scene_picture": "摄像机实景图片",
	"networking_property": "联网属性", "access_network": "接入网络（*）", "ipv4_address": "IPv4地址（*）",
	"ipv6_address": "IPv6地址", "mac_address": "设备MAC地址（*）", "access_port": "访问端口",
	"associated_encoder": "关联编码器", "device_username": "设备用户名", "device_password": "设备口令",
	"channel_number": "通道号", "connection_protocol": "连接协议", "enabled_time": "启用时间（*）",
	"scrapped_time": "报废时间", "device_status": "设备状态（*）", "inspection_status": "巡检状态",
	"video_loss": "视频丢失", "color_distortion": "色彩失真", "video_blur": "视频模糊", "brightness_exception": "亮度异常",
	"video_interference": "视频干扰", "video_lag": "视频卡顿", "video_occlusion": "视频遮挡", "scene_change": "场景变更",
	"online_duration": "在线时长", "offline_duration": "离线时长", "signaling_delay": "信令时延",
	"video_stream_delay": "视频流时延", "key_frame_delay": "关键帧时延", "recording_retention_days": "录像保存天数（*）",
	"storage_device_code": "存储设备编码", "storage_channel_number": "存储通道号", "storage_type": "存储类型",
	"cache_settings": "缓存设置", "notes": "备注", "collection_area_type": "采集区域类型（*）", "update_time": "更新时间",
}

// 模板下载时的表头定义 (73列)
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
	"录像保存天数（*）", "存储设备编码", "存储通道号", "存储类型", "缓存设置", "备注", "采集区域类型（*）", "更新时间",
}

// 数据库记录对应的结构体 (73字段) - 用于导出和导入
type FileItem struct {
	ID                     int
	DeviceCode             string
	OriginalDeviceCode     sql.NullString
	DeviceName             string
	DivisionCode           string
	MonitorPointType       string
	Pickup                 sql.NullString
	ParentDevice           sql.NullString
	ConstructionUnit       string
	ConstructionUnitCode   string
	ManagementUnit         string
	CameraDept             string
	AdminName              string
	AdminContact           string
	Contractor             string
	MaintainUnit           string
	DeviceVendor           string
	DeviceModel            sql.NullString
	CameraType             string
	AccessMethod           sql.NullString
	CameraFunctionType     string
	VideoEncodingFormat    string
	ImageResolution        string
	CameraLightProperty    sql.NullString
	BackendStructure       sql.NullString
	LensType               sql.NullString
	InstallationType       sql.NullString
	HeightType             string
	JurisdictionPolice     string
	InstallationAddress    string
	SurroundingLandmark    string
	Longitude              float64
	Latitude               float64
	InstallationLocation   string
	MonitoringDirection    string
	PoleNumber             string
	ScenePicture           sql.NullString
	NetworkingProperty     sql.NullString
	AccessNetwork          string
	IPv4Address            string
	IPv6Address            sql.NullString
	MACAddress             string
	AccessPort             sql.NullString
	AssociatedEncoder      sql.NullString
	DeviceUsername         sql.NullString
	DevicePassword         sql.NullString
	ChannelNumber          sql.NullString
	ConnectionProtocol     sql.NullString
	EnabledTime            sql.NullString
	ScrappedTime           sql.NullString
	DeviceStatus           string
	InspectionStatus       sql.NullString
	VideoLoss              sql.NullInt64
	ColorDistortion        sql.NullInt64
	VideoBlur              sql.NullInt64
	BrightnessException    sql.NullInt64
	VideoInterference      sql.NullInt64
	VideoLag               sql.NullInt64
	VideoOcclusion         sql.NullInt64
	SceneChange            sql.NullInt64
	OnlineDuration         sql.NullInt64
	OfflineDuration        sql.NullInt64
	SignalingDelay         sql.NullInt64
	VideoStreamDelay       sql.NullInt64
	KeyFrameDelay          sql.NullInt64
	RecordingRetentionDays int
	StorageDeviceCode      sql.NullString
	StorageChannelNumber   sql.NullString
	StorageType            sql.NullString
	CacheSettings          sql.NullString
	Notes                  sql.NullString
	CollectionAreaType     string
	UpdateTime             string
}

// 列表页面专用的轻量级结构体 (只包含7个显示字段)
type FileItemList struct {
	ID             int
	DeviceCode     string
	DeviceName     string
	DivisionCode   string
	Monitor_point_Type  string
	ManagementUnit string
	MaintainUnit   string
	UpdateTime     string
}

// 页面数据结构体
type PageData struct {
	Title         string
	ActiveMenu    string
	List          []FileItemList // 修改为使用轻量级结构体
	SearchCode    string
	SearchName    string
	CurrentPage   int
	TotalPages    int
	HasPrev       bool
	HasNext       bool
	PrevPage      int
	NextPage      int
	Query         string
	ImportMessage string
	ImportCount   int
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// --- 辅助工具：空字符串转 NULL ---

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

// --- Handler: 建档明细列表页 (GET) ---

func Handler(w http.ResponseWriter, r *http.Request) {
	// 获取查询参数
	searchCode := r.URL.Query().Get("device_code")
	searchName := r.URL.Query().Get("device_name")
	pageStr := r.URL.Query().Get("page")
	importMsg := r.URL.Query().Get("message")
	importCountStr := r.URL.Query().Get("count")

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	importCount, _ := strconv.Atoi(importCountStr)

	pageSize := 20
	offset := (page - 1) * pageSize

	// 构造查询条件
	whereSQL := " WHERE 1=1"
	args := []interface{}{}

	if searchCode != "" {
		whereSQL += " AND device_code LIKE ?"
		args = append(args, "%"+searchCode+"%")
	}
	if searchName != "" {
		whereSQL += " AND device_name LIKE ?"
		args = append(args, "%"+searchName+"%")
	}

	// 1. 查询总记录数
	var totalCount int
	countSQL := "SELECT COUNT(*) FROM fileList" + whereSQL
	err := db.DBInstance.QueryRow(countSQL, args...).Scan(&totalCount)
	if err != nil && err != sql.ErrNoRows {
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

	// 3. 查询列表数据 (只查询 7 个关键字段)
	queryFields := "id, device_code, device_name, division_code, monitor_point_type,management_unit, maintain_unit, update_time"
	querySQL := fmt.Sprintf("SELECT %s FROM fileList %s ORDER BY id DESC LIMIT ? OFFSET ?", queryFields, whereSQL)

	// 准备完整的参数列表
	queryArgs := append(args, pageSize, offset)

	rows, err := db.DBInstance.Query(querySQL, queryArgs...)
	if err != nil {
		http.Error(w, "数据库查询失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var fileList []FileItemList
	for rows.Next() {
		var item FileItemList

		// Scan 参数顺序和数量必须与 SELECT 字段完全匹配
		err = rows.Scan(
			&item.ID,
			&item.DeviceCode,
			&item.DeviceName,
			&item.DivisionCode,
			&item.Monitor_point_Type,
			&item.ManagementUnit,
			&item.MaintainUnit,
			&item.UpdateTime,
		)

		if err != nil {
			fmt.Printf("❌ Database scan error: %v\n", err)
			continue
		}

		fileList = append(fileList, item)
	}

	// 检查遍历过程中的错误
	if err = rows.Err(); err != nil {
		http.Error(w, "数据遍历失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. 准备数据并渲染模板
	data := PageData{
		Title:         "建档明细",
		ActiveMenu:    "filelist",
		List:          fileList,
		SearchCode:    searchCode,
		SearchName:    searchName,
		CurrentPage:   page,
		TotalPages:    totalPages,
		HasPrev:       page > 1,
		HasNext:       page < totalPages,
		PrevPage:      page - 1,
		NextPage:      page + 1,
		Query:         r.URL.RawQuery,
		ImportMessage: importMsg,
		ImportCount:   importCount,
	}

	funcMap := template.FuncMap{
		"contains": contains,
		"split":    strings.Split,
	}

	tmpl, err := template.New("filelist.html").Funcs(funcMap).ParseFiles("templates/filelist.html")
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

// --- ImportHandler: 导入 XLSX ---

func ImportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/filelist", http.StatusSeeOther)
		return
	}

	r.ParseMultipartForm(32 << 20)
	file, _, err := r.FormFile("upload_file")
	if err != nil {
		http.Error(w, "文件上传失败", http.StatusBadRequest)
		return
	}
	defer file.Close()

	f, err := excelize.OpenReader(file)
	if err != nil {
		http.Error(w, "Excel解析失败", http.StatusInternalServerError)
		return
	}
	sheetName := f.GetSheetName(0)

	rows, err := f.GetRows(sheetName)
	if err != nil {
		http.Error(w, "数据读取失败", http.StatusInternalServerError)
		return
	}

	tx, _ := db.DBInstance.Begin()

	// 定义 71 个插入字段 (跳过 id 和 update_time)
	insertFields := []string{
		"device_code", "original_device_code", "device_name", "division_code", "monitor_point_type",
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

	insertSQL := fmt.Sprintf("INSERT INTO fileList (%s) VALUES (%s)",
		strings.Join(insertFields, ", "),
		strings.TrimRight(strings.Repeat("?,", len(insertFields)), ","))

	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		tx.Rollback()
		http.Error(w, "SQL Prepare失败", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	importedCount := 0
	const expectedCols = 73 // A-BU 总计 73 列

	for i, row := range rows {
		if i == 0 {
			continue
		}

		// 强制对齐到 73 列
		for len(row) < expectedCols {
			row = append(row, "")
		}

		// 数据类型转换
		lon, _ := strconv.ParseFloat(row[31], 64)
		lat, _ := strconv.ParseFloat(row[32], 64)
		recDays, _ := strconv.Atoi(row[64])

		// 准备 71 个参数
		params := make([]interface{}, 71)

		for j := 0; j < 71; j++ {
			excelIdx := j + 1 // B列 (device_code) 对应 row[1]

			switch excelIdx {
			case 31:
				params[j] = lon
			case 32:
				params[j] = lat
			case 64:
				params[j] = recDays
    
			// --- 新增：处理启用时间格式 ---
			case 48: // enabled_time
				timeStr := strings.TrimSpace(row[excelIdx])
				if timeStr == "" {
					params[j] = nil // 如果为空存 NULL
				} else {
				// 尝试解析 ISO8601 格式 (2025-11-05T00:00:00+08:00)
					parsedTime, err := time.Parse(time.RFC3339, timeStr)
					if err == nil {
                // 如果解析成功，转为 MySQL 接受的格式: 2025-11-05 00:00:00
						params[j] = parsedTime.Format("2006-01-02 15:04:05")
					} else {
                // 如果不是 ISO 格式，直接按原样存（可能是已经手工改好的格式）
						params[j] = timeStr
            }
        }
    // ----------------------------

			// 必填项列表
			case 1, 3, 4, 5, 8, 9, 10, 11, 12, 13, 14, 15, 16, 18, 20, 21, 22, 27, 28, 29, 30, 33, 34, 35, 38, 39, 41, 50, 71:
			// 注意：我把 48 从这里移除了，因为上面单独处理了
				params[j] = toDBValue(row[excelIdx], true)

			// 可为空项
			default:
				params[j] = toDBValue(row[excelIdx], false)
			}
		}

		_, execErr := stmt.Exec(params...)
		if execErr != nil {
			tx.Rollback()
			errMsg := fmt.Sprintf("导入失败：第 %d 行数据错误。详细信息: %v", i+1, execErr)
			http.Error(w, errMsg, http.StatusInternalServerError)
			return
		}
		importedCount++
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, "数据库提交失败。", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/filelist?message=ImportSuccess&count="+strconv.Itoa(importedCount), http.StatusSeeOther)
}

// --- ExportHandler: 导出 XLSX ---

func ExportHandler(w http.ResponseWriter, r *http.Request) {
	searchCode := r.URL.Query().Get("device_code")
	searchName := r.URL.Query().Get("device_name")

	// 构造查询条件
	whereSQL := " WHERE 1=1"
	args := []interface{}{}

	if searchCode != "" {
		whereSQL += " AND device_code LIKE ?"
		args = append(args, "%"+searchCode+"%")
	}
	if searchName != "" {
		whereSQL += " AND device_name LIKE ?"
		args = append(args, "%"+searchName+"%")
	}

	// 查询所有73个字段
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
	FROM fileList` + whereSQL + ` ORDER BY id DESC`

	rows, err := db.DBInstance.Query(querySQL, args...)
	if err != nil {
		http.Error(w, "数据库查询失败", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// 创建 Excel 文件
	f := excelize.NewFile()
	sheetName := "建档明细"
	f.SetSheetName("Sheet1", sheetName)

	// 写入表头
	f.SetSheetRow(sheetName, "A1", &TemplateHeaders)

	// 写入数据
	rowNum := 2
	for rows.Next() {
		var item FileItem
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

	// 输出文件
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	filename := fmt.Sprintf("attachment; filename=\"建档明细导出_%s.xlsx\"", time.Now().Format("20060102"))
	w.Header().Set("Content-Disposition", filename)
	f.Write(w)
}

// --- DownloadTemplateHandler: 下载模板 ---

func DownloadTemplateHandler(w http.ResponseWriter, r *http.Request) {
	f := excelize.NewFile()
	sheetName := "导入模板"
	f.SetSheetName("Sheet1", sheetName)
	f.SetSheetRow(sheetName, "A1", &TemplateHeaders)

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	filename := fmt.Sprintf("attachment; filename=\"建档数据导入模板_%s.xlsx\"", time.Now().Format("20060102"))
	w.Header().Set("Content-Disposition", filename)
	f.Write(w)
}