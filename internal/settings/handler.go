package settings

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
	UploadFilePath string
	Message       string
	MessageType   string // success, error
	CurrentUser   *auth.User
}

// Handler 参数设置页面
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

	// 获取上传文件目录参数
	uploadFilePath := getSetting("upload_file_path")

	// 获取消息参数（用于显示保存成功/失败消息）
	message := r.URL.Query().Get("message")
	messageType := r.URL.Query().Get("type")

	data := PageData{
		Title:         "参数设置",
		ActiveMenu:    "settings",
		SubMenu:       "settings",
		UploadFilePath: uploadFilePath,
		Message:       message,
		MessageType:   messageType,
		CurrentUser:   currentUser,
	}

	// 渲染模板
	tmpl, err := template.ParseFiles("templates/settings.html")
	if err != nil {
		logger.Errorf("参数设置-模板解析失败: %v", err)
		http.Error(w, "模板解析失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		logger.Errorf("参数设置-模板渲染失败: %v", err)
		http.Error(w, "模板渲染失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// SaveHandler 保存参数设置
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

	// 获取表单数据
	uploadFilePath := r.FormValue("upload_file_path")

	// 保存参数
	err := saveSetting("upload_file_path", uploadFilePath)
	if err != nil {
		logger.Errorf("参数设置-保存失败: %v", err)
		http.Error(w, "保存失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 记录操作日志
	action := "保存参数设置（上传文件目录：" + uploadFilePath + "）"
	operationlog.Record(r, currentUser.Username, action)

	// 重定向到参数设置页面，显示成功消息
	http.Redirect(w, r, "/settings?message=保存成功&type=success", http.StatusFound)
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





