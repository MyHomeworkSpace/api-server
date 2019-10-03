package main

import (
	"database/sql"

	"github.com/MyHomeworkSpace/api-server/config"

	_ "github.com/go-sql-driver/mysql"
)

// DB is a pointer to the current connection to MySQL
var DB *sql.DB

func initDatabase() {
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

func deinitDatabase() {
	DB.Close()
}
