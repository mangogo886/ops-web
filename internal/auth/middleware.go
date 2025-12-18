package auth

import (
	"net/http"
)

// renderForbidden 在当前响应中输出一个居中的提示小窗
func renderForbidden(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)
	_, _ = w.Write([]byte(`
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <title>权限不足</title>
  <style>
    body { margin:0; padding:0; font-family:"Microsoft YaHei",sans-serif; background:rgba(0,0,0,0.05); }
    .overlay {
      position: fixed; top:0; left:0; right:0; bottom:0;
      display:flex; align-items:center; justify-content:center;
      background: rgba(0,0,0,0.15);
    }
    .dialog {
      background:white; padding:24px 28px; border-radius:8px;
      box-shadow:0 10px 30px rgba(0,0,0,0.15);
      min-width:260px; max-width:420px; text-align:center;
    }
    .dialog h3 { margin:0 0 10px; color:#c0392b; font-size:20px; }
    .dialog p { margin:0 0 16px; color:#555; font-size:14px; line-height:1.6; }
    .btn { display:inline-block; padding:8px 16px; border:none; border-radius:4px;
           background:#3498db; color:white; cursor:pointer; font-size:14px; }
    .btn:hover { background:#2980b9; }
  </style>
</head>
<body>
  <div class="overlay">
    <div class="dialog">
      <h3>权限不足</h3>
      <p>` + message + `</p>
      <button class="btn" onclick="window.history.length > 1 ? history.back() : window.location.href='/'">返回</button>
    </div>
  </div>
</body>
</html>`))
}

// RequireAuth 要求用户必须登录的中间件
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsAuthenticated(r) {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next(w, r)
	}
}

// RequireAdmin 要求用户必须是管理员的中间件
func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsAuthenticated(r) {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		if !IsAdmin(r) {
			renderForbidden(w, "需要管理员权限才能访问。")
			return
		}
		next(w, r)
	}
}

