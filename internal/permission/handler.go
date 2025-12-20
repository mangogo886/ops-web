package permission

import (
	"html/template"
	"net/http"
	"ops-web/internal/auth"
	"ops-web/internal/db"
	"ops-web/internal/logger"
	"ops-web/internal/operationlog"
)

// PageData 页面数据
type PageData struct {
	Title         string
	ActiveMenu    string
	SubMenu       string
	Message       string
	MessageType   string // success, error
	CurrentUser   *auth.User
	// 权限配置
	AllowDeviceAuditImport    bool
	AllowDeviceAuditDelete    bool
	AllowCheckpointAuditImport bool
	AllowCheckpointAuditDelete bool
}

// Handler 权限设置页面
func Handler(w http.ResponseWriter, r *http.Request) {
	// 获取当前用户
	currentUser := auth.GetCurrentUser(r)
	if currentUser == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// 只允许管理员访问
	if currentUser.RoleCode != 0 {
		http.Error(w, "权限不足", http.StatusForbidden)
		return
	}

	// 获取权限配置
	allowDeviceAuditImport := getSettingBool("allow_device_audit_import")
	allowDeviceAuditDelete := getSettingBool("allow_device_audit_delete")
	allowCheckpointAuditImport := getSettingBool("allow_checkpoint_audit_import")
	allowCheckpointAuditDelete := getSettingBool("allow_checkpoint_audit_delete")

	// 获取消息参数（用于显示保存成功/失败消息）
	message := r.URL.Query().Get("message")
	messageType := r.URL.Query().Get("type")

	data := PageData{
		Title:         "权限设置",
		ActiveMenu:    "settings",
		SubMenu:       "permission",
		Message:       message,
		MessageType:   messageType,
		CurrentUser:   currentUser,
		AllowDeviceAuditImport:    allowDeviceAuditImport,
		AllowDeviceAuditDelete:    allowDeviceAuditDelete,
		AllowCheckpointAuditImport: allowCheckpointAuditImport,
		AllowCheckpointAuditDelete: allowCheckpointAuditDelete,
	}

	// 渲染模板
	tmpl, err := template.ParseFiles("templates/permission.html")
	if err != nil {
		logger.Errorf("权限设置-模板解析失败: %v", err)
		http.Error(w, "模板解析失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		logger.Errorf("权限设置-模板渲染失败: %v", err)
		http.Error(w, "模板渲染失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// SaveHandler 保存权限设置
func SaveHandler(w http.ResponseWriter, r *http.Request) {
	// 获取当前用户
	currentUser := auth.GetCurrentUser(r)
	if currentUser == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// 只允许管理员访问
	if currentUser.RoleCode != 0 {
		http.Error(w, "权限不足", http.StatusForbidden)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	// 获取权限配置
	allowDeviceAuditImport := r.FormValue("allow_device_audit_import") == "on"
	allowDeviceAuditDelete := r.FormValue("allow_device_audit_delete") == "on"
	allowCheckpointAuditImport := r.FormValue("allow_checkpoint_audit_import") == "on"
	allowCheckpointAuditDelete := r.FormValue("allow_checkpoint_audit_delete") == "on"

	// 保存权限配置
	saveSettingBool("allow_device_audit_import", allowDeviceAuditImport)
	saveSettingBool("allow_device_audit_delete", allowDeviceAuditDelete)
	saveSettingBool("allow_checkpoint_audit_import", allowCheckpointAuditImport)
	saveSettingBool("allow_checkpoint_audit_delete", allowCheckpointAuditDelete)

	// 记录操作日志
	action := "保存权限设置"
	operationlog.Record(r, currentUser.Username, action)

	// 重定向到权限设置页面，显示成功消息
	http.Redirect(w, r, "/permission?message=保存成功&type=success", http.StatusFound)
}

// getSettingBool 获取布尔类型参数值
func getSettingBool(key string) bool {
	value := getSetting(key)
	return value == "1" || value == "true" || value == "True" || value == "TRUE"
}

// saveSettingBool 保存布尔类型参数值
func saveSettingBool(key string, value bool) error {
	var strValue string
	if value {
		strValue = "1"
	} else {
		strValue = "0"
	}
	return saveSetting(key, strValue)
}

// getSetting 获取参数值
func getSetting(key string) string {
	var value string
	query := "SELECT param_value FROM system_settings WHERE param_key = ?"
	err := db.DBInstance.QueryRow(query, key).Scan(&value)
	if err != nil {
		// 参数不存在时返回空字符串
		return ""
	}
	return value
}

// saveSetting 保存参数值
func saveSetting(key, value string) error {
	// 使用 INSERT ... ON DUPLICATE KEY UPDATE
	query := `
		INSERT INTO system_settings (param_key, param_value) 
		VALUES (?, ?) 
		ON DUPLICATE KEY UPDATE param_value = ?, update_time = CURRENT_TIMESTAMP
	`
	_, err := db.DBInstance.Exec(query, key, value, value)
	return err
}

// CheckPermission 检查用户是否有权限执行某个操作
// 管理员始终有权限，普通用户需要检查系统设置
func CheckPermission(user *auth.User, permissionKey string) bool {
	// 管理员始终有权限
	if user != nil && user.RoleCode == 0 {
		return true
	}
	// 普通用户检查系统设置
	return getSettingBool(permissionKey)
}

