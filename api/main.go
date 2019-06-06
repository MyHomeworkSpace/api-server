package api

import (
	"database/sql"
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

	e.GET("/admin/getAllFeedback", route(routeAdminGetAllFeedback))
	e.GET("/admin/getFeedbackScreenshot/:id", route(routeAdminGetFeedbackScreenshot))
	e.GET("/admin/getUserCount", route(routeAdminGetUserCount))

	e.POST("/application/completeAuth", route(routeApplicationCompleteAuth))
	e.GET("/application/get/:id", route(routeApplicationGet))
	e.GET("/application/getAuthorizations", route(routeApplicationGetAuthorizations))
	e.GET("/application/requestAuth/:id", route(routeApplicationRequestAuth))
	e.POST("/application/revokeAuth", route(routeApplicationRevokeAuth))
	e.POST("/application/revokeSelf", route(routeApplicationRevokeSelf))

	e.POST("/application/manage/create", route(routeApplicationManageCreate))
	e.GET("/application/manage/getAll", route(routeApplicationManageGetAll))
	e.POST("/application/manage/update", route(routeApplicationManageUpdate))
	e.POST("/application/manage/delete", route(routeApplicationManageDelete))

	e.POST("/auth/clearMigrateFlag", route(routeAuthClearMigrateFlag))
	e.GET("/auth/csrf", route(routeAuthCsrf))
	e.POST("/auth/login", route(routeAuthLogin))
	e.GET("/auth/me", route(routeAuthMe))
	e.GET("/auth/logout", route(routeAuthLogout))
	e.GET("/auth/session", route(routeAuthSession))

	e.POST("/auth/2fa/beginEnroll", route(routeAuth2faBeginEnroll))
	e.POST("/auth/2fa/completeEnroll", route(routeAuth2faCompleteEnroll))
	e.GET("/auth/2fa/status", route(routeAuth2faStatus))
	e.POST("/auth/2fa/unenroll", route(routeAuth2faUnenroll))

	e.GET("/calendar/getStatus", route(routeCalendarGetStatus))
	e.GET("/calendar/getView", route(routeCalendarGetView))
	e.POST("/calendar/import", route(routeCalendarImport))
	e.POST("/calendar/resetSchedule", route(routeCalendarResetSchedule))

	e.GET("/calendar/events/getWeek/:monday", route(routeCalendarEventsGetWeek))

	e.POST("/calendar/events/add", route(routeCalendarEventsAdd))
	e.POST("/calendar/events/edit", route(routeCalendarEventsEdit))
	e.POST("/calendar/events/delete", route(routeCalendarEventsDelete))

	e.POST("/calendar/hwEvents/add", route(routeCalendarHWEventsAdd))
	e.POST("/calendar/hwEvents/edit", route(routeCalendarHWEventsEdit))
	e.POST("/calendar/hwEvents/delete", route(routeCalendarHWEventsDelete))

	e.GET("/classes/get", route(routeClassesGet))
	e.GET("/classes/get/:id", route(routeClassesGetID))
	e.GET("/classes/hwInfo/:id", route(routeClassesHWInfo))
	e.POST("/classes/add", route(routeClassesAdd))
	e.POST("/classes/edit", route(routeClassesEdit))
	e.POST("/classes/delete", route(routeClassesDelete))
	e.POST("/classes/swap", route(routeClassesSwap))

	e.POST("/feedback/add", route(routeFeedbackAdd))

	e.GET("/homework/get", route(routeHomeworkGet))
	e.GET("/homework/getForClass/:classId", route(routeHomeworkGetForClass))
	e.GET("/homework/getHWView", route(routeHomeworkGetHWView))
	e.GET("/homework/getHWViewSorted", route(routeHomeworkGetHWViewSorted))
	e.GET("/homework/get/:id", route(routeHomeworkGetID))
	e.GET("/homework/getWeek/:monday", route(routeHomeworkGetWeek))
	e.GET("/homework/getPickerSuggestions", route(routeHomeworkGetPickerSuggestions))
	e.GET("/homework/search", route(routeHomeworkSearch))
	e.POST("/homework/add", route(routeHomeworkAdd))
	e.POST("/homework/edit", route(routeHomeworkEdit))
	e.POST("/homework/delete", route(routeHomeworkDelete))
	e.POST("/homework/markOverdueDone", route(routeHomeworkMarkOverdueDone))

	e.POST("/notifications/add", route(routeNotificationsAdd))
	e.POST("/notifications/delete", route(routeNotificationsDelete))
	e.GET("/notifications/get", route(routeNotificationsGet))

	e.GET("/planner/getWeekInfo/:date", route(routePlannerGetWeekInfo))

	e.GET("/prefixes/getDefaultList", route(routePrefixesGetDefaultList))
	e.GET("/prefixes/getList", route(routePrefixesGetList))
	e.POST("/prefixes/delete", route(routePrefixesDelete))
	e.POST("/prefixes/add", route(routePrefixesAdd))

	e.GET("/prefs/get/:key", route(routePrefsGet))
	e.GET("/prefs/getAll", route(routePrefsGetAll))
	e.POST("/prefs/set", route(routePrefsSet))
}
