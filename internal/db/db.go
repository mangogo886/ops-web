package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"ops-web/internal/logger"
	"os"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	DBInstance *sql.DB
	once       sync.Once
)

type Config struct {
	DBHost string `json:"db_host"`
	DBPort string `json:"db_port"`
	DBUser string `json:"db_user"`
	DBPass string `json:"db_pass"`
	DBName string `json:"db_name"`
}

func InitDB() error {
	var err error
	once.Do(func() {
		// 读取配置文件
		file, e := os.Open("config/config.json")
		if e != nil {
			logger.Errorf("数据库初始化-读取配置文件失败: %v", e)
			err = e
			return
		}
		defer file.Close()

		decoder := json.NewDecoder(file)
		cfg := Config{}
		if e := decoder.Decode(&cfg); e != nil {
			logger.Errorf("数据库初始化-解析配置文件失败: %v", e)
			err = e
			return
		}

		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)

		DBInstance, err = sql.Open("mysql", dsn)
		if err != nil {
			logger.Errorf("数据库初始化-打开数据库连接失败: %v, DSN: %s", err, dsn)
			return
		}

		DBInstance.SetMaxOpenConns(20)
		DBInstance.SetMaxIdleConns(5)
		DBInstance.SetConnMaxLifetime(time.Minute * 5)

		err = DBInstance.Ping()
		if err != nil {
			logger.Errorf("数据库初始化-连接测试失败: %v, DSN: %s", err, dsn)
		}
	})
	return err
}