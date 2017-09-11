package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func InitDatabase() {
	var err error
	DB, err = sql.Open("mysql", config.Database.Username+":"+config.Database.Password+"@/"+config.Database.Database)
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
