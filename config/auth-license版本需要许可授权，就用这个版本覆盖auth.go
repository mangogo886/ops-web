package auth

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
	"ops-web/internal/db"
	"ops-web/internal/license"
	"ops-web/internal/logger"
	"ops-web/internal/operationlog"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	SessionCookieName = "ops_session"
	SessionMaxAge     = 24 * 60 * 60 // 24小时
)

// User 用户信息结构
type User struct {
	ID       int
	Username string
	RoleID   int
	RoleCode int    // 0=管理员，1=普通用户
	RoleName string
}

// Session 会话信息
type Session struct {
	UserID   int
	Username string
	RoleID   int
	RoleCode int
	ExpireAt time.Time
}

// LoginHandler 处理登录请求
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// 如果已登录，重定向到首页
		if IsAuthenticated(r) {
			http.Redirect(w, r, "/filelist", http.StatusFound)
			return
		}
		// 显示登录页面
		renderLoginPage(w, "")
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		renderLoginPage(w, "用户名和密码不能为空")
		return
	}

	// 查询用户
	var user User
	var hashedPassword string
	query := `
		SELECT u.id, u.username, u.password, u.role_id, ur.role_code, ur.role_name
		FROM users u
		LEFT JOIN user_role ur ON u.role_id = ur.id
		WHERE u.username = ?
	`
	err := db.DBInstance.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &hashedPassword, &user.RoleID, &user.RoleCode, &user.RoleName,
	)

	if err == sql.ErrNoRows {
		renderLoginPage(w, "用户名或密码错误")
		return
	}
	if err != nil {
		logger.Errorf("登录-数据库查询失败: %v, 用户名: %s", err, username)
		http.Error(w, "数据库查询失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		renderLoginPage(w, "用户名或密码错误")
		return
	}

	// 验证许可
	isValid, errMsg := license.ValidateLicense()
	if !isValid {
		logger.Errorf("登录-许可验证失败: %s, 用户名: %s", errMsg, username)
		renderLoginPage(w, "系统错误，请联系管理员")
		return
	}

	// 创建会话
	sessionID := createSession(user.ID, user.Username, user.RoleID, user.RoleCode)

	// 设置 Cookie
	cookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    sessionID,
		Path:     "/",
		MaxAge:   SessionMaxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)

	// 记录登录日志
	operationlog.Record(r, user.Username, "登录成功")

	// 重定向到首页
	http.Redirect(w, r, "/filelist", http.StatusFound)
}

// LogoutHandler 处理登出请求
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// 删除 Cookie
	cookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)

	// 记录登出
	if u := GetCurrentUser(r); u != nil {
		operationlog.Record(r, u.Username, "退出登录")
	}

	http.Redirect(w, r, "/login", http.StatusFound)
}

// IsAuthenticated 检查用户是否已登录
func IsAuthenticated(r *http.Request) bool {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return false
	}

	session := getSession(cookie.Value)
	if session == nil {
		return false
	}

	// 检查是否过期
	if time.Now().After(session.ExpireAt) {
		return false
	}

	return true
}

// GetCurrentUser 获取当前登录用户信息
func GetCurrentUser(r *http.Request) *User {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return nil
	}

	session := getSession(cookie.Value)
	if session == nil {
		return nil
	}

	// 检查是否过期
	if time.Now().After(session.ExpireAt) {
		return nil
	}

	// 查询用户详细信息
	var user User
	query := `
		SELECT u.id, u.username, u.role_id, ur.role_code, ur.role_name
		FROM users u
		LEFT JOIN user_role ur ON u.role_id = ur.id
		WHERE u.id = ?
	`
	err = db.DBInstance.QueryRow(query, session.UserID).Scan(
		&user.ID, &user.Username, &user.RoleID, &user.RoleCode, &user.RoleName,
	)
	if err != nil {
		return nil
	}

	return &user
}

// IsAdmin 检查当前用户是否是管理员
func IsAdmin(r *http.Request) bool {
	user := GetCurrentUser(r)
	if user == nil {
		return false
	}
	return user.RoleCode == 0 // 0=管理员
}

// HashPassword 加密密码
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// 简单的内存 session 存储（生产环境建议使用 Redis）
var sessions = make(map[string]*Session)

// createSession 创建会话
func createSession(userID int, username string, roleID int, roleCode int) string {
	sessionID := generateSessionID()
	session := &Session{
		UserID:   userID,
		Username: username,
		RoleID:   roleID,
		RoleCode: roleCode,
		ExpireAt: time.Now().Add(time.Duration(SessionMaxAge) * time.Second),
	}
	sessions[sessionID] = session
	return sessionID
}

// getSession 获取会话
func getSession(sessionID string) *Session {
	return sessions[sessionID]
}

// generateSessionID 生成会话ID
func generateSessionID() string {
	return base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())))
}

// LoginPageData 登录页面数据
type LoginPageData struct {
	ErrorMsg string
}

// renderLoginPage 渲染登录页面
func renderLoginPage(w http.ResponseWriter, errorMsg string) {
	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
		logger.Errorf("登录-模板解析失败: %v", err)
		http.Error(w, "模板解析失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := LoginPageData{
		ErrorMsg: errorMsg,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		logger.Errorf("登录-模板渲染失败: %v", err)
		http.Error(w, "模板渲染失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

