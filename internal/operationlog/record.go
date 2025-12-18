package operationlog

import (
	"net"
	"net/http"
	"ops-web/internal/db"
	"strings"
)

// LogEntry 单条日志
type LogEntry struct {
	ID        int64
	Username  string
	Action    string
	IP        string
	CreatedAt string
}

// Record 写入一条操作日志
func Record(r *http.Request, username, action string) {
	ip := clientIP(r)
	_, _ = db.DBInstance.Exec(
		"INSERT INTO operation_logs (username, action, ip) VALUES (?, ?, ?)",
		username, action, ip,
	)
}

// 获取客户端 IP，优先 X-Forwarded-For
func clientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}






















