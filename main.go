package main

// ... imports ...
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
    
    http.HandleFunc("/stats", statistics.Handler)
	
    
    // 建档明细
    http.HandleFunc("/filelist", filelist.Handler)
    // ⬇️ 新增路由 ⬇️
    http.HandleFunc("/filelist/export", filelist.ExportHandler)
    http.HandleFunc("/filelist/import", filelist.ImportHandler)
    // ⬆️ 新增路由 ⬆️
	// ⬅️ 新增：下载模板路由
	http.HandleFunc("/filelist/download-template", filelist.DownloadTemplateHandler)

    // 3. 启动服务
    log.Println("Server starting on http://127.0.0.1:8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }
}