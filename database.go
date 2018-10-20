package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func InitDatabase() {
	var err error
	DB, err = sql.Open("mysql", config.Database.Username+":"+config.Database.Password+"@/"+config.Database.Database + "?charset=utf8mb4&collation=utf8mb4_unicode_ci")
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
