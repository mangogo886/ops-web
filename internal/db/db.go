package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
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
			err = e
			return
		}
		defer file.Close()

		decoder := json.NewDecoder(file)
		cfg := Config{}
		if e := decoder.Decode(&cfg); e != nil {
			err = e
			return
		}

		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)

		DBInstance, err = sql.Open("mysql", dsn)
		if err != nil {
			return
		}

		DBInstance.SetMaxOpenConns(20)
		DBInstance.SetMaxIdleConns(5)
		DBInstance.SetConnMaxLifetime(time.Minute * 5)

		err = DBInstance.Ping()
	})
	return err
}