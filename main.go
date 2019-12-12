package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/MyHomeworkSpace/api-server/schools"

	"github.com/MyHomeworkSpace/api-server/schools/dalton"
	"github.com/MyHomeworkSpace/api-server/schools/mit"

	"github.com/MyHomeworkSpace/api-server/api"
	"github.com/MyHomeworkSpace/api-server/auth"
	"github.com/MyHomeworkSpace/api-server/calendar"
	"github.com/MyHomeworkSpace/api-server/config"
	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/email"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type errorResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type csrfResponse struct {
	Status string `json:"status"`
	Token  string `json:"token"`
}

func main() {
	log.Println("MyHomeworkSpace API Server")

	config.Init()

	initDatabase()
	initRedis()

	email.Init()

	calendar.InitCalendar()

	api.DB = DB
	api.MainRegistry = schools.MainRegistry
	api.RedisClient = RedisClient

	auth.DB = DB
	auth.RedisClient = RedisClient

	data.DB = DB
	data.MainRegistry = schools.MainRegistry
	data.RedisClient = RedisClient

	schools.MainRegistry.Register(dalton.CreateSchool())
	schools.MainRegistry.Register(mit.CreateSchool())

	e := echo.New()
	e.Pre(middleware.RemoveTrailingSlash())
	e.Static("/api_tester", "api_tester")
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "MyHomeworkSpace API Server")
	})

	api.Init(e) // API init delayed because router must be started first

	log.Printf("Listening on port %d", config.GetCurrent().Server.Port)
	err := e.Start(fmt.Sprintf(":%d", config.GetCurrent().Server.Port))
	if err != nil {
		log.Fatalln(err)
	}
}
