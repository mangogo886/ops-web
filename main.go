package main

import (
    "log"
    "net/http"
    "ops-web/internal/auth"
    "ops-web/internal/auditprogress"
    "ops-web/internal/auditstatistics"
    "ops-web/internal/db"
    "ops-web/internal/filelist"
    "ops-web/internal/logger"
    "ops-web/internal/operationlog"
    "ops-web/internal/statistics"
    "ops-web/internal/user"
)

func main() {
    // 0. 初始化日志系统
    if err := logger.InitLogger(); err != nil {
        log.Printf("警告: 初始化日志系统失败: %v", err)
    }
    defer logger.Close()
    
    // 1. 初始化数据库
    if err := db.InitDB(); err != nil {
        logger.Errorf("数据库连接失败: %v", err)
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.DBInstance.Close()
    
    // 1.1. 初始化审核相关的数据库表
    if err := db.InitAuditTables(); err != nil {
        logger.Errorf("初始化审核表失败: %v", err)
        log.Printf("警告: 初始化审核表失败: %v", err)
        // 不中断程序运行，允许用户手动创建表
    }

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
    http.HandleFunc("/stats", auth.RequireAuth(statistics.Handler))               // 统计信息页面
    http.HandleFunc("/stats/export", auth.RequireAuth(statistics.ExportHandler))  // 导出统计信息
    
    // ===== 建档明细路由（需要登录） =====
    http.HandleFunc("/filelist", auth.RequireAuth(filelist.Handler))
    http.HandleFunc("/filelist/export", auth.RequireAuth(filelist.ExportHandler))

    // ===== 审核进度路由（需要登录） =====
    // 注意：必须先注册子路由，再注册父路由
    http.HandleFunc("/audit/progress", auth.RequireAuth(auditprogress.Handler))
    http.HandleFunc("/audit/progress/import", auth.RequireAuth(auditprogress.ImportHandler))
    http.HandleFunc("/audit/progress/detail", auth.RequireAuth(auditprogress.DetailHandler))
    http.HandleFunc("/audit/progress/edit", auth.RequireAuth(auditprogress.EditCommentHandler))
    http.HandleFunc("/audit/progress/delete", auth.RequireAuth(auditprogress.DeleteHandler))
    http.HandleFunc("/audit/progress/download-template", auth.RequireAuth(auditprogress.DownloadTemplateHandler))
    http.HandleFunc("/audit/statistics", auth.RequireAuth(auditstatistics.Handler))
    http.HandleFunc("/audit/statistics/export", auth.RequireAuth(auditstatistics.ExportHandler))
    http.HandleFunc("/audit", auth.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
        http.Redirect(w, r, "/audit/progress", http.StatusFound)
    }))

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
    log.Println("统计信息: http://127.0.0.1:8080/stats")
    log.Println("建档明细: http://127.0.0.1:8080/filelist")
    log.Println("用户管理: http://127.0.0.1:8080/users")
    
    if err := http.ListenAndServe(":8080", nil); err != nil {
        logger.Errorf("HTTP服务启动失败: %v", err)
        log.Fatal(err)
    }
}