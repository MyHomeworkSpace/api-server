package main

import (
	"database/sql"

	"github.com/MyHomeworkSpace/api-server/config"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func InitDatabase() {
	var err error
	DB, err = sql.Open("mysql", config.GetCurrent().Database.Username+":"+config.GetCurrent().Database.Password+"@/"+config.GetCurrent().Database.Database+"?charset=utf8mb4&collation=utf8mb4_unicode_ci")
	if err != nil {
		panic(err)
	}
	err = DB.Ping()
	if err != nil {
		panic(err)
	}
}

func DeInitDatabase() {
	DB.Close()
}
