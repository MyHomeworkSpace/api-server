package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/labstack/echo"

	"gopkg.in/redis.v5"
)

var AuthURLBase string
var DB *sql.DB
var RedisClient *redis.Client

var WhitelistEnabled bool
var WhitelistFile string
var WhitelistBlockMsg string

type StatusResponse struct {
	Status string `json:"status"`
}

type ErrorResponse struct {
	Status string `json:"status"`
	Error string `json:"error"`
}

type UserResponse struct {
	Status string `json:"status"`
	ID int `json:"id"`
	Name string `json:"name"`
	Username string `json:"username"`
	Email string `json:"email"`
	Type string `json:"type"`
	Features string `json:"features"`
	ShowMigrateMessage int `json:"showMigrateMessage"`
}

func Init(e *echo.Echo) {
	e.GET("/status", func(c echo.Context) error {
		return c.String(http.StatusOK, "Alive")
	})

	InitApplicationAPI(e)
	InitAuthAPI(e)
	InitCalendarAPI(e)
	InitClassesAPI(e)
	InitFeedbackAPI(e)
	InitHomeworkAPI(e)
	InitPlannerAPI(e)
	InitPrefsAPI(e)

	log.Println("API endpoints ready.")
}
