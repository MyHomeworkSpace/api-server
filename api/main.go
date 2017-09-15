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
	Error  string `json:"error"`
}

func Init(e *echo.Echo) {
	e.GET("/status", func(c echo.Context) error {
		return c.String(http.StatusOK, "Alive")
	})

	InitApplicationAPI(e)
	InitAuthAPI(e)
	InitCalendarAPI(e)
	InitCalendarEventsAPI(e)
	InitClassesAPI(e)
	InitFeedbackAPI(e)
	InitHomeworkAPI(e)
	InitPlannerAPI(e)
	InitPrefixesAPI(e)
	InitPrefsAPI(e)
	InitScheduleAPI(e)

	log.Println("API endpoints ready.")
}
