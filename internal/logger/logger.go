package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	errorLog *log.Logger
	logFile  *os.File
)

// InitLogger 初始化日志系统
func InitLogger() error {
	// 创建logs目录（如果不存在）
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("创建logs目录失败: %v", err)
	}

	// 创建日志文件（按日期命名）
	logFileName := fmt.Sprintf("ops-web-%s.log", time.Now().Format("2006-01-02"))
	logFilePath := filepath.Join(logsDir, logFileName)

	// 打开日志文件（追加模式）
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %v", err)
	}

	logFile = file
	errorLog = log.New(file, "", log.LstdFlags|log.Lshortfile)

	return nil
}

// Error 记录错误日志
func Error(format string, v ...interface{}) {
	if errorLog != nil {
		errorLog.Printf("[ERROR] "+format, v...)
	}
}

// Errorf 记录错误日志（带格式化）
func Errorf(format string, v ...interface{}) {
	Error(format, v...)
}

// Close 关闭日志文件
func Close() {
	if logFile != nil {
		logFile.Close()
	}
}







