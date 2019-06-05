package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/labstack/echo"

	"gopkg.in/redis.v5"
)

type routeFunc func(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext)
type authLevel int

const (
	authLevelNone authLevel = iota
	authLevelLoggedIn
)

var AuthURLBase string
var DB *sql.DB
var RedisClient *redis.Client
var ReverseProxyHeader string

var FeedbackSlackEnabled bool
var FeedbackSlackURL string
var FeedbackSlackHostName string

type StatusResponse struct {
	Status string `json:"status"`
}

type ErrorResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

// A RouteContext contains information relevant to the current route
type RouteContext struct {
}

func route(f routeFunc) func(ec echo.Context) error {
	return func(ec echo.Context) error {
		f(ec.Response(), ec.Request(), ec, RouteContext{})
		return nil
	}
}

func routeStatus(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	ec.String(http.StatusOK, "Alive")
}

// Init will initialize all available API endpoints
func Init(e *echo.Echo) {
	e.GET("/status", route(routeStatus))

	InitAdminAPI(e)
	InitApplicationAPI(e)
	InitAuthAPI(e)
	InitAuth2FAAPI(e)
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
