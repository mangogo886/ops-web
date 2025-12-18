package taskconfig

import (
	"encoding/json"
	"log"
	"ops-web/internal/db"
	"ops-web/internal/logger"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

// 定时任务调度器相关变量
var (
	schedulerStopChan chan bool
	schedulerRunning  bool
	schedulerMutex    sync.Mutex
)

// ReloadScheduler 重新加载定时任务配置
func ReloadScheduler() {
	schedulerMutex.Lock()
	defer schedulerMutex.Unlock()

	// 停止现有调度器
	if schedulerRunning {
		schedulerStopChan <- true
		schedulerRunning = false
		// 等待调度器停止
		time.Sleep(100 * time.Millisecond)
	}

	// 启动新的调度器
	go startScheduler()
}

// startScheduler 启动定时任务调度器
func startScheduler() {
	schedulerMutex.Lock()
	schedulerRunning = true
	schedulerStopChan = make(chan bool)
	schedulerMutex.Unlock()

	ticker := time.NewTicker(1 * time.Hour) // 每小时检查一次
	defer ticker.Stop()

	// 立即检查一次（启动时）
	checkAndRunScheduledTasks()

	for {
		select {
		case <-schedulerStopChan:
			return
		case <-ticker.C:
			checkAndRunScheduledTasks()
		}
	}
}

// checkAndRunScheduledTasks 检查并执行定时任务
func checkAndRunScheduledTasks() {
	now := time.Now()
	currentHour := now.Hour()
	currentWeekday := now.Weekday()

	// 检查数据库备份任务
	dbBackupEnabled := getSetting("db_backup_enabled")
	if dbBackupEnabled == "1" {
		dbBackupFrequency := getSetting("db_backup_frequency")
		dbBackupHourStr := getSetting("db_backup_hour")

		dbBackupHour, err := strconv.Atoi(dbBackupHourStr)
		if err == nil && dbBackupHour >= 1 && dbBackupHour <= 24 {
			shouldRun := false
			if dbBackupFrequency == "daily" {
				// 每天：当前小时匹配
				shouldRun = currentHour == dbBackupHour
			} else if dbBackupFrequency == "weekly" {
				// 每周：当前小时匹配且是周一
				shouldRun = currentHour == dbBackupHour && currentWeekday == time.Monday
			}

			if shouldRun {
				go runDatabaseBackup()
			}
		}
	}

	// 检查文件备份任务
	fileBackupEnabled := getSetting("file_backup_enabled")
	if fileBackupEnabled == "1" {
		fileBackupFrequency := getSetting("file_backup_frequency")
		fileBackupHourStr := getSetting("file_backup_hour")

		fileBackupHour, err := strconv.Atoi(fileBackupHourStr)
		if err == nil && fileBackupHour >= 1 && fileBackupHour <= 24 {
			shouldRun := false
			if fileBackupFrequency == "daily" {
				// 每天：当前小时匹配
				shouldRun = currentHour == fileBackupHour
			} else if fileBackupFrequency == "weekly" {
				// 每周：当前小时匹配且是周一
				shouldRun = currentHour == fileBackupHour && currentWeekday == time.Monday
			}

			if shouldRun {
				go runFileBackup()
			}
		}
	}
}

// runDatabaseBackup 执行数据库备份（定时任务调用）
func runDatabaseBackup() {
	// 获取数据库备份路径
	backupPath := getSetting("database_backup_path")
	if backupPath == "" {
		logger.Errorf("定时任务-数据库备份路径未设置")
		return
	}

	// 读取配置文件获取数据库信息
	configFile, err := os.Open("config/config.json")
	if err != nil {
		logger.Errorf("定时任务-数据库备份-读取配置文件失败: %v", err)
		return
	}
	defer configFile.Close()

	var config db.Config
	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(&config); err != nil {
		logger.Errorf("定时任务-数据库备份-解析配置文件失败: %v", err)
		return
	}

	// 创建当天日期目录
	today := time.Now().Format("2006-01-02")
	backupDir := filepath.Join(backupPath, today)

	// 创建备份目录
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		logger.Errorf("定时任务-数据库备份-创建备份目录失败: %v", err)
		return
	}

	// 获取所有表名
	tables, err := getAllTables(config.DBName)
	if err != nil {
		logger.Errorf("定时任务-数据库备份-获取表列表失败: %v", err)
		return
	}

	// 备份每个表
	backupCount := 0
	for _, table := range tables {
		outputFile := filepath.Join(backupDir, table+".sql")
		if err := backupTable(config, table, outputFile); err != nil {
			logger.Errorf("定时任务-数据库备份-备份表 %s 失败: %v", table, err)
			continue
		}
		backupCount++
	}

	// 清理超过5天的备份
	if err := cleanOldBackups(backupPath, 5); err != nil {
		logger.Errorf("定时任务-数据库备份-清理旧备份失败: %v", err)
	}

	log.Printf("定时任务-数据库备份完成（备份表数：%d/%d）", backupCount, len(tables))
}

// runFileBackup 执行文件备份（定时任务调用）
func runFileBackup() {
	// 获取文件上传路径和备份路径
	uploadPath := getSetting("upload_file_path")
	backupPath := getSetting("backup_file_path")

	if uploadPath == "" || backupPath == "" {
		logger.Errorf("定时任务-文件备份-路径未设置")
		return
	}

	// 检查上传路径是否存在
	if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
		logger.Errorf("定时任务-文件备份-上传路径不存在: %s", uploadPath)
		return
	}

	// 创建备份目录（如果不存在）
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		logger.Errorf("定时任务-文件备份-创建备份目录失败: %v", err)
		return
	}

	// 复制文件
	copiedCount, err := copyDirectory(uploadPath, backupPath)
	if err != nil {
		logger.Errorf("定时任务-文件备份-复制文件失败: %v", err)
		return
	}

	log.Printf("定时任务-文件备份完成（复制文件数：%d）", copiedCount)
}

// InitScheduler 初始化定时任务调度器（在main.go中调用）
func InitScheduler() {
	ReloadScheduler()
}

