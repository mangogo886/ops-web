package main

import (
    "fmt"
    "log"
    "net/http"
    "ops-web/internal/auth"
    "ops-web/internal/auditprogress"
    "ops-web/internal/auditstatistics"
    "ops-web/internal/checkpointfilelist"
    "ops-web/internal/checkpointprogress"
    "ops-web/internal/db"
    "ops-web/internal/filelist"
    "ops-web/internal/logger"
    "ops-web/internal/operationlog"
    "ops-web/internal/statistics"
    "ops-web/internal/taskconfig"
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
    
    // 1.2. 初始化卡口审核相关的数据库表
    if err := db.InitCheckpointTables(); err != nil {
        logger.Errorf("初始化卡口审核表失败: %v", err)
        log.Printf("警告: 初始化卡口审核表失败: %v", err)
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
                http.Redirect(w, r, "/device/filelist", http.StatusFound)
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
    // 设备建档明细（原/filelist路由）
    http.HandleFunc("/device/filelist", auth.RequireAuth(filelist.Handler))
    http.HandleFunc("/device/filelist/export", auth.RequireAuth(filelist.ExportHandler))
    // 兼容旧路由，重定向到新路由
    http.HandleFunc("/filelist", auth.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
        http.Redirect(w, r, "/device/filelist", http.StatusFound)
    }))
    http.HandleFunc("/filelist/export", auth.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
        // 重定向到新的导出路由
        newURL := "/device/filelist/export" + r.URL.RawQuery
        if newURL != "" {
            newURL = "?" + newURL
        }
        http.Redirect(w, r, newURL, http.StatusFound)
    }))
    
    // 卡口建档明细
    http.HandleFunc("/checkpoint/filelist", auth.RequireAuth(checkpointfilelist.Handler))
    http.HandleFunc("/checkpoint/filelist/export", auth.RequireAuth(checkpointfilelist.ExportHandler))

    // ===== 审核进度路由（需要登录） =====
    // 注意：必须先注册子路由，再注册父路由
    http.HandleFunc("/audit/progress", auth.RequireAuth(auditprogress.Handler))
    http.HandleFunc("/audit/progress/import", auth.RequireAuth(auditprogress.ImportHandler))
    http.HandleFunc("/audit/progress/detail", auth.RequireAuth(auditprogress.DetailHandler))
    http.HandleFunc("/audit/progress/detail/export", auth.RequireAuth(auditprogress.DetailExportHandler))
    http.HandleFunc("/audit/progress/edit", auth.RequireAuth(auditprogress.EditCommentHandler))
    http.HandleFunc("/audit/progress/delete", auth.RequireAuth(auditprogress.DeleteHandler))
    http.HandleFunc("/audit/progress/download-template", auth.RequireAuth(auditprogress.DownloadTemplateHandler))
    http.HandleFunc("/audit/progress/upload", auth.RequireAuth(auditprogress.UploadHandler))
    http.HandleFunc("/audit/progress/download", auth.RequireAuth(auditprogress.DownloadHandler))
    http.HandleFunc("/audit/statistics", auth.RequireAuth(auditstatistics.Handler))
    http.HandleFunc("/audit/statistics/export", auth.RequireAuth(auditstatistics.ExportHandler))
    http.HandleFunc("/audit", auth.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
        http.Redirect(w, r, "/audit/progress", http.StatusFound)
    }))
    
    // ===== 卡口审核进度路由（需要登录） =====
    http.HandleFunc("/checkpoint/progress", auth.RequireAuth(checkpointprogress.Handler))
    http.HandleFunc("/checkpoint/progress/import", auth.RequireAuth(checkpointprogress.ImportHandler))
    http.HandleFunc("/checkpoint/progress/detail", auth.RequireAuth(checkpointprogress.DetailHandler))
    http.HandleFunc("/checkpoint/progress/detail/export", auth.RequireAuth(checkpointprogress.DetailExportHandler))
    http.HandleFunc("/checkpoint/progress/edit", auth.RequireAuth(checkpointprogress.EditCommentHandler))
    http.HandleFunc("/checkpoint/progress/delete", auth.RequireAuth(checkpointprogress.DeleteHandler))
    http.HandleFunc("/checkpoint/progress/download-template", auth.RequireAuth(checkpointprogress.DownloadTemplateHandler))
    http.HandleFunc("/checkpoint/progress/upload", auth.RequireAuth(checkpointprogress.UploadHandler))
    http.HandleFunc("/checkpoint/progress/download", auth.RequireAuth(checkpointprogress.DownloadHandler))

    // ===== 用户管理路由（需要管理员权限） =====
    http.HandleFunc("/users", auth.RequireAuth(user.Handler))
    http.HandleFunc("/users/add", auth.RequireAdmin(user.AddHandler))
    http.HandleFunc("/users/edit", auth.RequireAdmin(user.EditHandler))
    http.HandleFunc("/users/delete", auth.RequireAdmin(user.DeleteHandler))

    // ===== 操作日志（需要管理员权限） =====
    http.HandleFunc("/logs", auth.RequireAdmin(operationlog.Handler))

    // ===== 任务配置（需要管理员权限） =====
    http.HandleFunc("/taskconfig", auth.RequireAdmin(taskconfig.Handler))
    http.HandleFunc("/taskconfig/save", auth.RequireAdmin(taskconfig.SaveHandler))
    http.HandleFunc("/taskconfig/backup-database", auth.RequireAdmin(taskconfig.BackupDatabaseHandler))
    http.HandleFunc("/taskconfig/backup-files", auth.RequireAdmin(taskconfig.BackupFileHandler))

    // 2.1. 初始化定时任务调度器
    taskconfig.InitScheduler()

    // 3. 启动服务
    serverAddr := ":" + db.AppConfig.ServerPort
    baseURL := fmt.Sprintf("http://%s:%s", db.AppConfig.ServerHost, db.AppConfig.ServerPort)
    
    log.Printf("Server starting on %s", baseURL)
    log.Printf("登录页面: %s/login", baseURL)
    log.Printf("统计信息: %s/stats", baseURL)
    log.Printf("设备建档明细: %s/device/filelist", baseURL)
    log.Printf("卡口建档明细: %s/checkpoint/filelist", baseURL)
    log.Printf("用户管理: %s/users", baseURL)
    
    if err := http.ListenAndServe(serverAddr, nil); err != nil {
        logger.Errorf("HTTP服务启动失败: %v", err)
        log.Fatal(err)
    }
}