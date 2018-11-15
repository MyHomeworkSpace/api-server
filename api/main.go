package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/labstack/echo"

	"gopkg.in/redis.v5"
)

type RouteFunc func(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext)
type AuthLevel int

const (
	AuthLevelNone AuthLevel = iota
	AuthLevelSignedIn
)

var AuthURLBase string
var DB *sql.DB
var RedisClient *redis.Client
var ReverseProxyHeader string

var FeedbackSlackEnabled bool
var FeedbackSlackURL string
var FeedbackSlackHostName string

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

type RouteContext struct {
}

func Route(f RouteFunc) func(ec echo.Context) error {
	return func(ec echo.Context) error {
		f(ec.Response(), ec.Request(), ec, RouteContext{})
		return nil
	}
}

func Route_Status(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	ec.String(http.StatusOK, "Alive")
}

func Init(e *echo.Echo) {
	e.GET("/status", Route(Route_Status))

	InitAdminAPI(e)
	InitApplicationAPI(e)
	InitAuthAPI(e)
	InitCalendarAPI(e)
	InitCalendarEventsAPI(e)
	InitClassesAPI(e)
	InitFeedbackAPI(e)
	InitHomeworkAPI(e)
	InitNotificationsAPI(e)
	InitPlannerAPI(e)
	InitPrefixesAPI(e)
	InitPrefsAPI(e)

	log.Println("API endpoints ready.")
}
