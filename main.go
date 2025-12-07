package main

import (
    "log"
    "net/http"
    "ops-web/internal/db"
    "ops-web/internal/filelist"
    "ops-web/internal/statistics"
)

func main() {
    // 1. 初始化数据库
    if err := db.InitDB(); err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.DBInstance.Close()

    // 2. 注册路由
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/" {
            http.Redirect(w, r, "/filelist", http.StatusFound) 
            return
        }
        http.NotFound(w, r)
    })
    
    // ===== 统计信息路由 =====
    // 注意：必须先注册子路由，再注册父路由
    http.HandleFunc("/stats/weekly", statistics.WeeklyHandler)  // 本周建档统计
    http.HandleFunc("/stats/total", statistics.TotalHandler)    // 建档全量统计
    http.HandleFunc("/stats", statistics.Handler)               // 主入口，重定向到本周统计
    
    // ===== 建档明细路由 =====
    http.HandleFunc("/filelist", filelist.Handler)
    http.HandleFunc("/filelist/export", filelist.ExportHandler)
    http.HandleFunc("/filelist/import", filelist.ImportHandler)
    http.HandleFunc("/filelist/download-template", filelist.DownloadTemplateHandler)

    // 3. 启动服务
    log.Println("Server starting on http://127.0.0.1:8080")
    log.Println("统计信息:")
    log.Println("  - 按日期统计: http://127.0.0.1:8080/stats/weekly")
    log.Println("  - 建档全量统计: http://127.0.0.1:8080/stats/total")
    log.Println("建档明细: http://127.0.0.1:8080/filelist")
    
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }
}