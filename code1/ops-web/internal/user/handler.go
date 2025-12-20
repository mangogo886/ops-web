package user

import (
	"html/template"
	"net/http"
	"ops-web/internal/auth"
	"ops-web/internal/db"
	"ops-web/internal/logger"
	"ops-web/internal/operationlog"
	"strconv"
	"strings"
)

// UserInfo 用户信息结构
type UserInfo struct {
	ID       int
	Username string
	RoleID   int
	RoleCode int
	RoleName string
}

// RoleInfo 角色信息结构
type RoleInfo struct {
	ID       int
	RoleName string
	RoleCode int
}

// PageData 页面数据
type PageData struct {
	Title      string
	ActiveMenu string
	SubMenu    string
	Users      []UserInfo
	Roles      []RoleInfo
	Message    string
	MessageType string // success, error
	CurrentUser *auth.User
}

// Handler 用户列表页面
func Handler(w http.ResponseWriter, r *http.Request) {
	// 获取当前用户
	currentUser := auth.GetCurrentUser(r)
	if currentUser == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// 查询所有用户
	query := `
		SELECT u.id, u.username, u.role_id, ur.role_code, ur.role_name
		FROM users u
		LEFT JOIN user_role ur ON u.role_id = ur.id
		ORDER BY u.id DESC
	`
	rows, err := db.DBInstance.Query(query)
	if err != nil {
		logger.Errorf("用户管理-查询用户失败: %v, SQL: %s", err, query)
		http.Error(w, "查询用户失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []UserInfo
	for rows.Next() {
		var user UserInfo
		err := rows.Scan(&user.ID, &user.Username, &user.RoleID, &user.RoleCode, &user.RoleName)
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	// 查询所有角色
	roles, err := getRoles()
	if err != nil {
		logger.Errorf("用户管理-查询角色失败: %v", err)
		http.Error(w, "查询角色失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 获取消息参数
	message := r.URL.Query().Get("message")
	messageType := r.URL.Query().Get("type")

	data := PageData{
		Title:       "用户信息",
		ActiveMenu:  "settings",
		SubMenu:     "users",
		Users:       users,
		Roles:       roles,
		Message:     message,
		MessageType: messageType,
		CurrentUser: currentUser,
	}

	renderTemplate(w, data)
}

// AddHandler 添加用户
func AddHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/users", http.StatusFound)
		return
	}

	currentUser := auth.GetCurrentUser(r)

	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")
	roleIDStr := r.FormValue("role_id")

	if username == "" || password == "" || roleIDStr == "" {
		http.Redirect(w, r, "/users?message=用户名、密码和角色不能为空&type=error", http.StatusFound)
		return
	}

	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil {
		http.Redirect(w, r, "/users?message=角色ID无效&type=error", http.StatusFound)
		return
	}

	// 检查用户名是否已存在
	var count int
	err = db.DBInstance.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&count)
	if err != nil {
		http.Redirect(w, r, "/users?message=数据库查询失败&type=error", http.StatusFound)
		return
	}
	if count > 0 {
		http.Redirect(w, r, "/users?message=用户名已存在&type=error", http.StatusFound)
		return
	}

	// 加密密码
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		http.Redirect(w, r, "/users?message=密码加密失败&type=error", http.StatusFound)
		return
	}

	// 插入用户
	insertSQL := "INSERT INTO users (username, password, role_id) VALUES (?, ?, ?)"
	_, err = db.DBInstance.Exec(insertSQL, username, hashedPassword, roleID)
	if err != nil {
		http.Redirect(w, r, "/users?message=添加用户失败: "+err.Error()+"&type=error", http.StatusFound)
		return
	}

	if currentUser != nil {
		operationlog.Record(r, currentUser.Username, "添加用户:"+username)
	}

	http.Redirect(w, r, "/users?message=用户添加成功&type=success", http.StatusFound)
}

// EditHandler 编辑用户
func EditHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/users", http.StatusFound)
		return
	}

	currentUser := auth.GetCurrentUser(r)

	userIDStr := r.FormValue("user_id")
	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")
	roleIDStr := r.FormValue("role_id")

	if userIDStr == "" || username == "" || roleIDStr == "" {
		http.Redirect(w, r, "/users?message=用户ID、用户名和角色不能为空&type=error", http.StatusFound)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Redirect(w, r, "/users?message=用户ID无效&type=error", http.StatusFound)
		return
	}

	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil {
		http.Redirect(w, r, "/users?message=角色ID无效&type=error", http.StatusFound)
		return
	}

	// 检查用户名是否被其他用户使用
	var count int
	err = db.DBInstance.QueryRow("SELECT COUNT(*) FROM users WHERE username = ? AND id != ?", username, userID).Scan(&count)
	if err != nil {
		http.Redirect(w, r, "/users?message=数据库查询失败&type=error", http.StatusFound)
		return
	}
	if count > 0 {
		http.Redirect(w, r, "/users?message=用户名已被其他用户使用&type=error", http.StatusFound)
		return
	}

	// 如果提供了新密码，则更新密码
	if password != "" {
		hashedPassword, err := auth.HashPassword(password)
		if err != nil {
			http.Redirect(w, r, "/users?message=密码加密失败&type=error", http.StatusFound)
			return
		}
		updateSQL := "UPDATE users SET username = ?, password = ?, role_id = ? WHERE id = ?"
		_, err = db.DBInstance.Exec(updateSQL, username, hashedPassword, roleID, userID)
		if err != nil {
			http.Redirect(w, r, "/users?message=更新用户失败: "+err.Error()+"&type=error", http.StatusFound)
			return
		}
	} else {
		// 不更新密码
		updateSQL := "UPDATE users SET username = ?, role_id = ? WHERE id = ?"
		_, err = db.DBInstance.Exec(updateSQL, username, roleID, userID)
		if err != nil {
			http.Redirect(w, r, "/users?message=更新用户失败: "+err.Error()+"&type=error", http.StatusFound)
			return
		}
	}

	if currentUser != nil {
		operationlog.Record(r, currentUser.Username, "编辑用户:"+username)
	}

	http.Redirect(w, r, "/users?message=用户更新成功&type=success", http.StatusFound)
}

// DeleteHandler 删除用户
func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/users", http.StatusFound)
		return
	}

	userIDStr := r.FormValue("user_id")
	if userIDStr == "" {
		http.Redirect(w, r, "/users?message=用户ID不能为空&type=error", http.StatusFound)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Redirect(w, r, "/users?message=用户ID无效&type=error", http.StatusFound)
		return
	}

	// 获取当前用户
	currentUser := auth.GetCurrentUser(r)
	if currentUser != nil && currentUser.ID == userID {
		http.Redirect(w, r, "/users?message=不能删除当前登录用户&type=error", http.StatusFound)
		return
	}

	// 删除用户
	_, err = db.DBInstance.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		http.Redirect(w, r, "/users?message=删除用户失败: "+err.Error()+"&type=error", http.StatusFound)
		return
	}

	if currentUser != nil {
		operationlog.Record(r, currentUser.Username, "删除用户ID:"+strconv.Itoa(userID))
	}

	http.Redirect(w, r, "/users?message=用户删除成功&type=success", http.StatusFound)
}

// getRoles 获取所有角色
func getRoles() ([]RoleInfo, error) {
	rows, err := db.DBInstance.Query("SELECT id, role_name, role_code FROM user_role ORDER BY role_code")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []RoleInfo
	for rows.Next() {
		var role RoleInfo
		err := rows.Scan(&role.ID, &role.RoleName, &role.RoleCode)
		if err != nil {
			continue
		}
		roles = append(roles, role)
	}
	return roles, nil
}

// renderTemplate 渲染模板
func renderTemplate(w http.ResponseWriter, data PageData) {
	tmpl, err := template.ParseFiles("templates/users.html")
	if err != nil {
		logger.Errorf("用户管理-模板解析失败: %v", err)
		http.Error(w, "模板解析失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		logger.Errorf("用户管理-模板渲染失败: %v", err)
		http.Error(w, "模板渲染失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

