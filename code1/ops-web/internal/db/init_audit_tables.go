package db

import (
	"io/ioutil"
	"log"
	"ops-web/internal/logger"
	"strings"
)

// InitAuditTables 初始化审核相关的数据库表
func InitAuditTables() error {
	// 只执行创建表的SQL，字段修改由用户手动执行SQL
	// 如果SQL文件不存在，不记录错误日志（用户可能手动执行建表语句）
	err := executeSQLFile("deploy/create-audit-tables.sql")
	if err != nil {
		// 只在控制台输出警告，不记录到错误日志
		log.Printf("提示: 未找到审核表SQL文件，将跳过自动建表（如需要请手动执行SQL脚本）")
	}
	
	// 执行添加record_count字段的SQL（如果字段已存在会报错，但可以忽略）
	err = executeSQLFile("deploy/add-record-count-fields.sql")
	if err != nil {
		// 字段可能已存在，不记录错误日志
		log.Printf("提示: 执行添加record_count字段SQL失败（字段可能已存在）")
	}
	
	return nil
}

func executeSQLFile(filename string) error {
	// 读取SQL文件
	sqlBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		// 文件不存在时不记录错误日志（用户可能手动执行建表语句）
		// 只返回错误，由调用者决定是否记录日志
		return err
	}

	sql := string(sqlBytes)
	
	// 移除注释行（以--开头的行）
	lines := strings.Split(sql, "\n")
	var cleanLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// 保留空行用于分隔SQL语句，但移除注释行
		if !strings.HasPrefix(trimmed, "--") {
			cleanLines = append(cleanLines, line)
		}
	}
	sql = strings.Join(cleanLines, "\n")
	
	// 分割SQL语句（以分号分割）
	statements := strings.Split(sql, ";")
	
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		// 跳过空语句
		if stmt == "" {
			continue
		}
		
		// 执行SQL语句（使用 CREATE TABLE IF NOT EXISTS，所以表已存在也不会报错）
		_, err := DBInstance.Exec(stmt)
		if err != nil {
			// 执行SQL失败时记录错误日志（这是真正的错误）
			logger.Errorf("初始化审核表-执行SQL语句失败: %v, SQL: %s", err, stmt)
			log.Printf("执行SQL失败: %v", err)
			// 继续执行其他语句，不中断
		}
	}
	
	return nil
}

// InitCheckpointTables 初始化卡口审核相关的数据库表
func InitCheckpointTables() error {
	// 只执行创建表的SQL，字段修改由用户手动执行SQL
	// 如果SQL文件不存在，不记录错误日志（用户可能手动执行建表语句）
	err := executeSQLFile("deploy/create-checkpoint-tables.sql")
	if err != nil {
		// 只在控制台输出警告，不记录到错误日志
		log.Printf("提示: 未找到卡口审核表SQL文件，将跳过自动建表（如需要请手动执行SQL脚本）")
	}
	
	// 执行添加record_count字段的SQL（如果字段已存在会报错，但可以忽略）
	err = executeSQLFile("deploy/add-record-count-fields.sql")
	if err != nil {
		// 字段可能已存在，不记录错误日志
		log.Printf("提示: 执行添加record_count字段SQL失败（字段可能已存在）")
	}
	
	return nil
}







