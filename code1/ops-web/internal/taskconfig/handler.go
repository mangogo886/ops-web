package taskconfig

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"ops-web/internal/auth"
	"ops-web/internal/db"
	"ops-web/internal/logger"
	"ops-web/internal/operationlog"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// PageData 页面数据
type PageData struct {
	Title                    string
	ActiveMenu               string
	SubMenu                  string
	UploadFilePath           string
	BackupFilePath           string
	DatabaseBackupPath       string
	// 数据库备份定时任务
	DBBackupEnabled          string // 是否启用：1=启用，0=禁用
	DBBackupFrequency        string // 频率：daily=每天，weekly=每周
	DBBackupHour             string // 时间：1-24
	// 文件备份定时任务
	FileBackupEnabled        string // 是否启用：1=启用，0=禁用
	FileBackupFrequency      string // 频率：daily=每天，weekly=每周
	FileBackupHour           string // 时间：1-24
	Message                  string
	MessageType              string // success, error
	CurrentUser              *auth.User
}

// Handler 任务配置页面
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

	// 获取配置参数
	uploadFilePath := getSetting("upload_file_path")
	backupFilePath := getSetting("backup_file_path")
	databaseBackupPath := getSetting("database_backup_path")
	
	// 获取定时任务配置
	dbBackupEnabled := getSetting("db_backup_enabled")
	if dbBackupEnabled == "" {
		dbBackupEnabled = "0"
	}
	dbBackupFrequency := getSetting("db_backup_frequency")
	if dbBackupFrequency == "" {
		dbBackupFrequency = "daily"
	}
	dbBackupHour := getSetting("db_backup_hour")
	if dbBackupHour == "" {
		dbBackupHour = "2"
	}
	
	fileBackupEnabled := getSetting("file_backup_enabled")
	if fileBackupEnabled == "" {
		fileBackupEnabled = "0"
	}
	fileBackupFrequency := getSetting("file_backup_frequency")
	if fileBackupFrequency == "" {
		fileBackupFrequency = "daily"
	}
	fileBackupHour := getSetting("file_backup_hour")
	if fileBackupHour == "" {
		fileBackupHour = "3"
	}

	// 获取消息参数（用于显示保存成功/失败消息）
	message := r.URL.Query().Get("message")
	messageType := r.URL.Query().Get("type")

	data := PageData{
		Title:               "任务配置",
		ActiveMenu:          "settings",
		SubMenu:             "task_config",
		UploadFilePath:      uploadFilePath,
		BackupFilePath:      backupFilePath,
		DatabaseBackupPath:  databaseBackupPath,
		DBBackupEnabled:     dbBackupEnabled,
		DBBackupFrequency:   dbBackupFrequency,
		DBBackupHour:        dbBackupHour,
		FileBackupEnabled:   fileBackupEnabled,
		FileBackupFrequency: fileBackupFrequency,
		FileBackupHour:      fileBackupHour,
		Message:             message,
		MessageType:         messageType,
		CurrentUser:         currentUser,
	}

	// 渲染模板
	tmpl, err := template.ParseFiles("templates/taskconfig.html")
	if err != nil {
		logger.Errorf("任务配置-模板解析失败: %v", err)
		http.Error(w, "模板解析失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		logger.Errorf("任务配置-模板渲染失败: %v", err)
		http.Error(w, "模板渲染失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// SaveHandler 保存任务配置
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
	backupFilePath := r.FormValue("backup_file_path")
	databaseBackupPath := r.FormValue("database_backup_path")
	
	// 获取定时任务配置
	dbBackupEnabled := r.FormValue("db_backup_enabled")
	dbBackupFrequency := r.FormValue("db_backup_frequency")
	dbBackupHour := r.FormValue("db_backup_hour")
	fileBackupEnabled := r.FormValue("file_backup_enabled")
	fileBackupFrequency := r.FormValue("file_backup_frequency")
	fileBackupHour := r.FormValue("file_backup_hour")

	// 保存参数
	err := saveSetting("upload_file_path", uploadFilePath)
	if err != nil {
		logger.Errorf("任务配置-保存上传文件路径失败: %v", err)
		http.Error(w, "保存失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = saveSetting("backup_file_path", backupFilePath)
	if err != nil {
		logger.Errorf("任务配置-保存备份路径失败: %v", err)
		http.Error(w, "保存失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = saveSetting("database_backup_path", databaseBackupPath)
	if err != nil {
		logger.Errorf("任务配置-保存数据库备份路径失败: %v", err)
		http.Error(w, "保存失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// 保存定时任务配置
	err = saveSetting("db_backup_enabled", dbBackupEnabled)
	if err != nil {
		logger.Errorf("任务配置-保存数据库备份启用状态失败: %v", err)
		http.Error(w, "保存失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = saveSetting("db_backup_frequency", dbBackupFrequency)
	if err != nil {
		logger.Errorf("任务配置-保存数据库备份频率失败: %v", err)
		http.Error(w, "保存失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = saveSetting("db_backup_hour", dbBackupHour)
	if err != nil {
		logger.Errorf("任务配置-保存数据库备份时间失败: %v", err)
		http.Error(w, "保存失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	err = saveSetting("file_backup_enabled", fileBackupEnabled)
	if err != nil {
		logger.Errorf("任务配置-保存文件备份启用状态失败: %v", err)
		http.Error(w, "保存失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = saveSetting("file_backup_frequency", fileBackupFrequency)
	if err != nil {
		logger.Errorf("任务配置-保存文件备份频率失败: %v", err)
		http.Error(w, "保存失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = saveSetting("file_backup_hour", fileBackupHour)
	if err != nil {
		logger.Errorf("任务配置-保存文件备份时间失败: %v", err)
		http.Error(w, "保存失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// 重新加载定时任务（通知调度器更新配置）
	ReloadScheduler()

	// 记录操作日志
	action := fmt.Sprintf("保存任务配置（文件上传路径：%s，备份路径：%s，数据库备份路径：%s，数据库备份定时：%s/%s/%s，文件备份定时：%s/%s/%s）",
		uploadFilePath, backupFilePath, databaseBackupPath,
		dbBackupEnabled, dbBackupFrequency, dbBackupHour,
		fileBackupEnabled, fileBackupFrequency, fileBackupHour)
	operationlog.Record(r, currentUser.Username, action)

	// 重定向到任务配置页面，显示成功消息
	http.Redirect(w, r, "/taskconfig?message="+url.QueryEscape("保存成功")+"&type=success", http.StatusFound)
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

// BackupDatabaseHandler 数据库备份处理
func BackupDatabaseHandler(w http.ResponseWriter, r *http.Request) {
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

	// 获取数据库备份路径
	backupPath := getSetting("database_backup_path")
	if backupPath == "" {
		http.Redirect(w, r, "/taskconfig?message="+url.QueryEscape("请先设置数据库备份路径")+"&type=error", http.StatusFound)
		return
	}

	// 读取配置文件获取数据库信息
	configFile, err := os.Open("config/config.json")
	if err != nil {
		logger.Errorf("数据库备份-读取配置文件失败: %v", err)
		http.Redirect(w, r, "/taskconfig?message="+url.QueryEscape("读取配置文件失败")+"&type=error", http.StatusFound)
		return
	}
	defer configFile.Close()

	var config db.Config
	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(&config); err != nil {
		logger.Errorf("数据库备份-解析配置文件失败: %v", err)
		http.Redirect(w, r, "/taskconfig?message="+url.QueryEscape("解析配置文件失败")+"&type=error", http.StatusFound)
		return
	}

	// 创建当天日期目录（YYYY-MM-DD格式）
	today := time.Now().Format("2006-01-02")
	backupDir := filepath.Join(backupPath, today)

	// 创建备份目录
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		logger.Errorf("数据库备份-创建备份目录失败: %v", err)
		http.Redirect(w, r, "/taskconfig?message="+url.QueryEscape("创建备份目录失败")+"&type=error", http.StatusFound)
		return
	}

	// 获取所有表名
	tables, err := getAllTables(config.DBName)
	if err != nil {
		logger.Errorf("数据库备份-获取表列表失败: %v", err)
		http.Redirect(w, r, "/taskconfig?message="+url.QueryEscape("获取表列表失败")+"&type=error", http.StatusFound)
		return
	}

	// 备份每个表
	backupCount := 0
	for _, table := range tables {
		outputFile := filepath.Join(backupDir, table+".sql")
		if err := backupTable(config, table, outputFile); err != nil {
			logger.Errorf("数据库备份-备份表 %s 失败: %v", table, err)
			continue
		}
		backupCount++
	}

	// 清理超过5天的备份
	if err := cleanOldBackups(backupPath, 5); err != nil {
		logger.Errorf("数据库备份-清理旧备份失败: %v", err)
		// 不中断流程，只记录错误
	}

	// 记录操作日志
	action := fmt.Sprintf("数据库备份（备份路径：%s，备份表数：%d/%d）", backupDir, backupCount, len(tables))
	operationlog.Record(r, currentUser.Username, action)

	// 重定向到任务配置页面，显示成功消息
	message := fmt.Sprintf("备份成功！共备份 %d/%d 个表到 %s", backupCount, len(tables), backupDir)
	http.Redirect(w, r, "/taskconfig?message="+url.QueryEscape(message)+"&type=success", http.StatusFound)
}

// getAllTables 获取所有表名
func getAllTables(dbName string) ([]string, error) {
	query := "SELECT table_name FROM information_schema.tables WHERE table_schema = ? AND table_type = 'BASE TABLE'"
	rows, err := db.DBInstance.Query(query, dbName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}
		tables = append(tables, tableName)
	}
	return tables, rows.Err()
}

// backupTable 备份单个表
func backupTable(config db.Config, tableName, outputFile string) error {
	// 构建mysqldump命令
	// mysqldump -h host -P port -u user -ppassword database table > output.sql
	cmd := exec.Command("mysqldump",
		"-h", config.DBHost,
		"-P", config.DBPort,
		"-u", config.DBUser,
		"-p"+config.DBPass,
		config.DBName,
		tableName,
	)

	// 创建输出文件
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %v", err)
	}
	defer outFile.Close()

	// 设置输出
	cmd.Stdout = outFile
	cmd.Stderr = os.Stderr

	// 执行命令
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("执行mysqldump失败: %v", err)
	}

	return nil
}

// cleanOldBackups 清理超过指定天数的备份
func cleanOldBackups(backupPath string, keepDays int) error {
	// 读取备份目录
	entries, err := ioutil.ReadDir(backupPath)
	if err != nil {
		return err
	}

	// 计算截止日期
	cutoffDate := time.Now().AddDate(0, 0, -keepDays)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// 尝试解析目录名为日期
		dirName := entry.Name()
		dirDate, err := time.Parse("2006-01-02", dirName)
		if err != nil {
			// 如果不是日期格式，跳过
			continue
		}

		// 如果日期早于截止日期，删除该目录
		if dirDate.Before(cutoffDate) {
			dirPath := filepath.Join(backupPath, dirName)
			if err := os.RemoveAll(dirPath); err != nil {
				logger.Errorf("数据库备份-删除旧备份目录失败: %v, 路径: %s", err, dirPath)
				continue
			}
			log.Printf("数据库备份-已删除旧备份目录: %s", dirPath)
		}
	}

	return nil
}

// BackupFileHandler 文件备份处理
func BackupFileHandler(w http.ResponseWriter, r *http.Request) {
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

	// 获取文件上传路径和备份路径
	uploadPath := getSetting("upload_file_path")
	backupPath := getSetting("backup_file_path")

	if uploadPath == "" {
		http.Redirect(w, r, "/taskconfig?message="+url.QueryEscape("请先设置文件上传路径")+"&type=error", http.StatusFound)
		return
	}

	if backupPath == "" {
		http.Redirect(w, r, "/taskconfig?message="+url.QueryEscape("请先设置备份路径")+"&type=error", http.StatusFound)
		return
	}

	// 检查上传路径是否存在
	if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
		http.Redirect(w, r, "/taskconfig?message="+url.QueryEscape("文件上传路径不存在")+"&type=error", http.StatusFound)
		return
	}

	// 创建备份目录（如果不存在）
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		logger.Errorf("文件备份-创建备份目录失败: %v", err)
		http.Redirect(w, r, "/taskconfig?message="+url.QueryEscape("创建备份目录失败")+"&type=error", http.StatusFound)
		return
	}

	// 复制文件
	copiedCount, err := copyDirectory(uploadPath, backupPath)
	if err != nil {
		logger.Errorf("文件备份-复制文件失败: %v", err)
		http.Redirect(w, r, "/taskconfig?message="+url.QueryEscape("文件备份失败: "+err.Error())+"&type=error", http.StatusFound)
		return
	}

	// 记录操作日志
	action := fmt.Sprintf("文件备份（上传路径：%s，备份路径：%s，复制文件数：%d）", uploadPath, backupPath, copiedCount)
	operationlog.Record(r, currentUser.Username, action)

	// 重定向到任务配置页面，显示成功消息
	message := fmt.Sprintf("文件备份成功！共复制 %d 个文件到 %s", copiedCount, backupPath)
	http.Redirect(w, r, "/taskconfig?message="+url.QueryEscape(message)+"&type=success", http.StatusFound)
}

// copyDirectory 复制目录下的所有文件（递归）
func copyDirectory(srcDir, dstDir string) (int, error) {
	copiedCount := 0

	// 遍历源目录
	err := filepath.Walk(srcDir, func(srcPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算相对路径
		relPath, err := filepath.Rel(srcDir, srcPath)
		if err != nil {
			return err
		}

		// 目标路径
		dstPath := filepath.Join(dstDir, relPath)

		// 如果是目录，创建目录
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// 如果是文件，复制文件
		srcFile, err := os.Open(srcPath)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		// 创建目标文件的目录（如果不存在）
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}

		// 创建目标文件
		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		// 复制文件内容
		_, err = dstFile.ReadFrom(srcFile)
		if err != nil {
			return err
		}

		copiedCount++
		return nil
	})

	return copiedCount, err
}

