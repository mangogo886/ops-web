package main

import (
    "log"
    "net/http"
    "ops-web/internal/auth"
    "ops-web/internal/db"
    "ops-web/internal/filelist"
    "ops-web/internal/operationlog"
    "ops-web/internal/statistics"
    "ops-web/internal/user"
)

func main() {
    // 1. 初始化数据库
    if err := db.InitDB(); err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.DBInstance.Close()

    // 2. 注册路由
    
    // ===== 认证路由（不需要登录） =====
    http.HandleFunc("/login", auth.LoginHandler)
    http.HandleFunc("/logout", auth.LogoutHandler)
    
    // ===== 根路径重定向 =====
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/" {
            if auth.IsAuthenticated(r) {
                http.Redirect(w, r, "/filelist", http.StatusFound)
            } else {
                http.Redirect(w, r, "/login", http.StatusFound)
            }
            return
        }
        http.NotFound(w, r)
    })
    
    // ===== 统计信息路由（需要登录） =====
    // 注意：必须先注册子路由，再注册父路由
    http.HandleFunc("/stats/weekly", auth.RequireAuth(statistics.WeeklyHandler))  // 本周建档统计
    http.HandleFunc("/stats/total", auth.RequireAuth(statistics.TotalHandler))    // 建档全量统计
    http.HandleFunc("/stats", auth.RequireAuth(statistics.Handler))               // 主入口，重定向到本周统计
    
    // ===== 建档明细路由（需要登录） =====
    http.HandleFunc("/filelist", auth.RequireAuth(filelist.Handler))
    http.HandleFunc("/filelist/export", auth.RequireAuth(filelist.ExportHandler))
    http.HandleFunc("/filelist/import", auth.RequireAuth(filelist.ImportHandler))
    http.HandleFunc("/filelist/delete", auth.RequireAuth(filelist.DeleteHandler))
    http.HandleFunc("/filelist/download-template", auth.RequireAuth(filelist.DownloadTemplateHandler))

    // ===== 用户管理路由（需要管理员权限） =====
    http.HandleFunc("/users", auth.RequireAuth(user.Handler))
    http.HandleFunc("/users/add", auth.RequireAdmin(user.AddHandler))
    http.HandleFunc("/users/edit", auth.RequireAdmin(user.EditHandler))
    http.HandleFunc("/users/delete", auth.RequireAdmin(user.DeleteHandler))

    // ===== 操作日志（需要管理员权限） =====
    http.HandleFunc("/logs", auth.RequireAdmin(operationlog.Handler))

    // 3. 启动服务
    log.Println("Server starting on http://127.0.0.1:8080")
    log.Println("登录页面: http://127.0.0.1:8080/login")
    log.Println("统计信息:")
    log.Println("  - 按日期统计: http://127.0.0.1:8080/stats/weekly")
    log.Println("  - 建档全量统计: http://127.0.0.1:8080/stats/total")
    log.Println("建档明细: http://127.0.0.1:8080/filelist")
    log.Println("用户管理: http://127.0.0.1:8080/users")
    
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }
}